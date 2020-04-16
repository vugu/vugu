/*
Package simplehttp provides an http.Handler that makes it easy to serve Vugu applications.
Useful for development and production.

The idea is that the common behaviors needed to serve a Vugu site are readily available
in one place.   If you require more functionality than simplehttp provides, nearly everything
it does is available in the github.com/vugu/vugu package and you can construct what you
need from its parts.  That said, simplehttp should make it easy to start:


	// dev flag enables most common development features
	// including rebuild your .wasm upon page reload
	dev := true
	h := simplehttp.New(dir, dev)

After creation, some flags are available for tuning, e.g.:

	h.EnableGenerate = true // upon page reload run "go generate ."
	h.DisableBuildCache = true // do not try to cache build results during development, just rebuild every time
	h.ParserGoPkgOpts.SkipRegisterComponentTypes = true // do not generate component registration init() stuff

Since it's just a regular http.Handler, starting a webserver is as simple as:

	log.Fatal(http.ListenAndServe("127.0.0.1:5678", h))

*/
package simplehttp

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/vugu/vugu/gen"
)

// SimpleHandler provides common web serving functionality useful for building Vugu sites.
type SimpleHandler struct {
	Dir string // project directory

	EnableBuildAndServe          bool                 // enables the build-and-serve sequence for your wasm binary - useful for dev, should be off in production
	EnableGenerate               bool                 // if true calls `go generate` (requires EnableBuildAndServe)
	ParserGoPkgOpts              *gen.ParserGoPkgOpts // if set enables running ParserGoPkg with these options (requires EnableBuildAndServe)
	DisableBuildCache            bool                 // if true then rebuild every time instead of trying to cache (requires EnableBuildAndServe)
	DisableTimestampPreservation bool                 // if true don't try to keep timestamps the same for files that are byte for byte identical (requires EnableBuildAndServe)
	MainWasmPath                 string               // path to serve main wasm file from, in dev mod defaults to "/main.wasm" (requires EnableBuildAndServe)
	WasmExecJsPath               string               // path to serve wasm_exec.js from after finding in the local Go installation, in dev mode defaults to "/wasm_exec.js"

	IsPage      func(r *http.Request) bool // func that returns true if PageHandler should serve the request
	PageHandler http.Handler               // returns the HTML page

	StaticHandler http.Handler // returns static assets from Dir with appropriate filtering or appropriate error

	wasmExecJsOnce    sync.Once
	wasmExecJsContent []byte
	wasmExecJsTs      time.Time

	lastBuildTime      time.Time // time of last successful build
	lastBuildContentGZ []byte    // last successful build gzipped

	mu sync.RWMutex
}

// New returns an SimpleHandler ready to serve using the specified directory.
// The dev flag indicates if development functionality is enabled.
// Settings on SimpleHandler may be tuned more specifically after creation, this function just
// returns sensible defaults for development or production according to if dev is true or false.
func New(dir string, dev bool) *SimpleHandler {

	if !filepath.IsAbs(dir) {
		panic(fmt.Errorf("dir %q is not an absolute path", dir))
	}

	ret := &SimpleHandler{
		Dir: dir,
	}

	ret.IsPage = DefaultIsPageFunc
	ret.PageHandler = &PageHandler{
		Template:         template.Must(template.New("_page_").Parse(DefaultPageTemplateSource)),
		TemplateDataFunc: DefaultTemplateDataFunc,
	}

	ret.StaticHandler = FilteredFileServer(
		regexp.MustCompile(`[.](css|js|html|map|jpg|jpeg|png|gif|svg|eot|ttf|otf|woff|woff2|wasm)$`),
		http.Dir(dir))

	if dev {
		ret.EnableBuildAndServe = true
		ret.ParserGoPkgOpts = &gen.ParserGoPkgOpts{}
		ret.MainWasmPath = "/main.wasm"
		ret.WasmExecJsPath = "/wasm_exec.js"
	}

	return ret
}

// ServeHTTP implements http.Handler.
func (h *SimpleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// by default we tell browsers to always check back with us for content, even in production;
	// we allow disabling by the caller just setting another value first; otherwise too much
	// headache caused by pages that won't reload and we still reduce a lot of bandwidth usage with
	// 304 responses, seems like a sensible trade off for now
	if w.Header().Get("Cache-Control") == "" {
		w.Header().Set("Cache-Control", "max-age=0, no-cache")
	}

	p := path.Clean("/" + r.URL.Path)

	if h.EnableBuildAndServe && h.MainWasmPath == p {
		h.buildAndServe(w, r)
		return
	}

	if h.WasmExecJsPath == p {
		h.serveGoEnvWasmExecJs(w, r)
		return
	}

	if h.IsPage(r) {
		h.PageHandler.ServeHTTP(w, r)
		return
	}

	h.StaticHandler.ServeHTTP(w, r)
}

func (h *SimpleHandler) buildAndServe(w http.ResponseWriter, r *http.Request) {

	// EnableGenerate      bool                  // if true calls `go generate` (requires EnableBuildAndServe)

	// main.wasm and build process, first check if it's needed

	h.mu.RLock()
	lastBuildTime := h.lastBuildTime
	lastBuildContentGZ := h.lastBuildContentGZ
	h.mu.RUnlock()

	var buildDirTs time.Time
	var err error

	if !h.DisableTimestampPreservation {
		buildDirTs, err = dirTimestamp(h.Dir)
		if err != nil {
			log.Printf("error in dirTimestamp(%q): %v", h.Dir, err)
			goto doBuild
		}
	}

	if len(lastBuildContentGZ) == 0 {
		// log.Printf("2")
		goto doBuild
	}

	if h.DisableBuildCache {
		goto doBuild
	}

	// skip build process if timestamp from build dir exists and is equal or older than our last build
	if !buildDirTs.IsZero() && !buildDirTs.After(lastBuildTime) {
		// log.Printf("3")
		goto serveBuiltFile
	}

	// // a false return value means we should send a 304
	// if !checkIfModifiedSince(r, buildDirTs) {
	// 	w.WriteHeader(http.StatusNotModified)
	// 	return
	// }

	// FIXME: might be useful to make it so only one thread rebuilds at a time and they both use the result

doBuild:

	// log.Printf("GOT HERE")

	{

		if h.ParserGoPkgOpts != nil {
			pg := gen.NewParserGoPkg(h.Dir, h.ParserGoPkgOpts)
			err := pg.Run()
			if err != nil {
				msg := fmt.Sprintf("Error from ParserGoPkg: %v", err)
				log.Print(msg)
				http.Error(w, msg, 500)
				return
			}
		}

		f, err := ioutil.TempFile("", "main_wasm_")
		if err != nil {
			panic(err)
		}
		fpath := f.Name()
		f.Close()
		os.Remove(f.Name())
		defer os.Remove(f.Name())

		var cmd *exec.Cmd

		startTime := time.Now()
		if h.EnableGenerate {
			cmd := exec.Command("go", "generate", ".")
			cmd.Dir = h.Dir
			cmd.Env = append(cmd.Env, os.Environ()...)
			b, err := cmd.CombinedOutput()
			w.Header().Set("X-Go-Generate-Duration", time.Since(startTime).String())
			if err != nil {
				msg := fmt.Sprintf("Error from generate: %v; Output:\n%s", err, b)
				log.Print(msg)
				http.Error(w, msg, 500)
				return
			}
		}

		// GOOS=js GOARCH=wasm go build -o main.wasm .
		startTime = time.Now()
		cmd = exec.Command("go", "build", "-o", fpath, ".")
		cmd.Dir = h.Dir
		cmd.Env = append(cmd.Env, os.Environ()...)
		cmd.Env = append(cmd.Env, "GOOS=js", "GOARCH=wasm")
		b, err := cmd.CombinedOutput()
		w.Header().Set("X-Go-Build-Duration", time.Since(startTime).String())
		if err != nil {
			msg := fmt.Sprintf("Error from compile: %v (out path=%q); Output:\n%s", err, fpath, b)
			log.Print(msg)
			http.Error(w, msg, 500)
			return
		}

		f, err = os.Open(fpath)
		if err != nil {
			msg := fmt.Sprintf("Error opening file after build: %v", err)
			log.Print(msg)
			http.Error(w, msg, 500)
			return
		}

		// gzip with max compression
		var buf bytes.Buffer
		gzw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
		n, err := io.Copy(gzw, f)
		if err != nil {
			msg := fmt.Sprintf("Error reading and compressing binary: %v", err)
			log.Print(msg)
			http.Error(w, msg, 500)
			return
		}
		gzw.Close()

		w.Header().Set("X-Gunzipped-Size", fmt.Sprint(n))

		// update cache

		if buildDirTs.IsZero() {
			lastBuildTime = time.Now()
		} else {
			lastBuildTime = buildDirTs
		}
		lastBuildContentGZ = buf.Bytes()

		// log.Printf("GOT TO UPDATE")
		h.mu.Lock()
		h.lastBuildTime = lastBuildTime
		h.lastBuildContentGZ = lastBuildContentGZ
		h.mu.Unlock()

	}

serveBuiltFile:

	w.Header().Set("Content-Type", "application/wasm")
	// w.Header().Set("Last-Modified", lastBuildTime.Format(http.TimeFormat)) // handled by http.ServeContent

	// if client supports gzip response (the usual case), we just set the gzip header and send back
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("X-Gzipped-Size", fmt.Sprint(len(lastBuildContentGZ)))
		http.ServeContent(w, r, h.MainWasmPath, lastBuildTime, bytes.NewReader(lastBuildContentGZ))
		return
	}

	// no gzip, we decompress internally and send it back
	gzr, _ := gzip.NewReader(bytes.NewReader(lastBuildContentGZ))
	_, err = io.Copy(w, gzr)
	if err != nil {
		log.Print(err)
	}
	return

}

func (h *SimpleHandler) serveGoEnvWasmExecJs(w http.ResponseWriter, r *http.Request) {

	b, err := exec.Command("go", "env", "GOROOT").CombinedOutput()
	if err != nil {
		http.Error(w, "failed to run `go env GOROOT`: "+err.Error(), 500)
		return
	}

	h.wasmExecJsOnce.Do(func() {
		h.wasmExecJsContent, err = ioutil.ReadFile(filepath.Join(strings.TrimSpace(string(b)), "misc/wasm/wasm_exec.js"))
		if err != nil {
			http.Error(w, "failed to run `go env GOROOT`: "+err.Error(), 500)
			return
		}
		h.wasmExecJsTs = time.Now() // hack but whatever for now
	})

	if len(h.wasmExecJsContent) == 0 {
		http.Error(w, "failed to read wasm_exec.js from local Go environment", 500)
		return
	}

	w.Header().Set("Content-Type", "text/javascript")
	http.ServeContent(w, r, "/wasm_exec.js", h.wasmExecJsTs, bytes.NewReader(h.wasmExecJsContent))
}

// FilteredFileServer is similar to the standard librarie's http.FileServer
// but the handler it returns will refuse to serve any files which don't
// match the specified regexp pattern after running through path.Clean().
// The idea is to make it easy to serve only specific kinds of
// static files from a directory.  If pattern does not match a 404 will be returned.
// Be sure to include a trailing "$" if you are checking for file extensions, so it
// only matches the end of the path, e.g. "[.](css|js)$"
func FilteredFileServer(pattern *regexp.Regexp, fs http.FileSystem) http.Handler {

	if pattern == nil {
		panic(fmt.Errorf("pattern is nil"))
	}

	if fs == nil {
		panic(fmt.Errorf("fs is nil"))
	}

	fserver := http.FileServer(fs)

	ret := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		p := path.Clean("/" + r.URL.Path)

		if !strings.HasPrefix(p, "/") { // should never happen after Clean above, but just being extra cautious
			http.NotFound(w, r)
			return
		}

		if !pattern.MatchString(p) {
			http.NotFound(w, r)
			return
		}

		// delegate to the regular file-serving behavior
		fserver.ServeHTTP(w, r)

	})

	return ret
}

// DefaultIsPageFunc will return true for any request to a path with no file extension.
var DefaultIsPageFunc = func(r *http.Request) bool {
	// anything without a file extension is a page
	return path.Ext(path.Clean("/"+r.URL.Path)) == ""
}

// DefaultPageTemplateSource a useful default HTML template for serving pages.
var DefaultPageTemplateSource = `<!doctype html>
<html>
<head>
{{if .Title}}
<title>{{.Title}}</title>
{{else}}
<title>Vugu Dev - {{.Request.URL.Path}}</title>
{{end}}
<meta charset="utf-8"/>
{{if .MetaTags}}{{range $k, $v := .MetaTags}}
<meta name="{{$k}}" content="{{$v}}"/>
{{end}}{{end}}
{{if .CSSFiles}}{{range $f := .CSSFiles}}
<link rel="stylesheet" href="{{$f}}" />
{{end}}{{end}}
<script src="https://cdn.jsdelivr.net/npm/text-encoding@0.7.0/lib/encoding.min.js"></script> <!-- MS Edge polyfill -->
<script src="/wasm_exec.js"></script>
</head>
<body>
<div id="vugu_mount_point">
{{if .ServerRenderedOutput}}{{.ServerRenderedOutput}}{{else}}
<img style="position: absolute; top: 50%; left: 50%;" src="https://cdnjs.cloudflare.com/ajax/libs/galleriffic/2.0.1/css/loader.gif">
{{end}}
</div>
<script>
var wasmSupported = (typeof WebAssembly === "object");
if (wasmSupported) {
	if (!WebAssembly.instantiateStreaming) { // polyfill
		WebAssembly.instantiateStreaming = async (resp, importObject) => {
			const source = await (await resp).arrayBuffer();
			return await WebAssembly.instantiate(source, importObject);
		};
	}
	const go = new Go();
	WebAssembly.instantiateStreaming(fetch("/main.wasm"), go.importObject).then((result) => {
		go.run(result.instance);
	});
} else {
	document.getElementById("vugu_mount_point").innerHTML = 'This application requires WebAssembly support.  Please upgrade your browser.';
}
</script>
</body>
</html>
`

// PageHandler executes a Go template and responsds with the page.
type PageHandler struct {
	Template         *template.Template
	TemplateDataFunc func(r *http.Request) interface{}
}

// DefaultStaticData is a map of static things added to the return value of DefaultTemplateDataFunc.
// Provides a quick and dirty way to do things like add CSS files to every page.
var DefaultStaticData = make(map[string]interface{}, 4)

// DefaultTemplateDataFunc is the default behavior for making template data.  It
// returns a map with "Request" set to r and all elements of DefaultStaticData added to it.
var DefaultTemplateDataFunc = func(r *http.Request) interface{} {
	ret := map[string]interface{}{
		"Request": r,
	}
	for k, v := range DefaultStaticData {
		ret[k] = v
	}
	return ret
}

// ServeHTTP implements http.Handler
func (h *PageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	tmplData := h.TemplateDataFunc(r)
	if tmplData == nil {
		http.NotFound(w, r)
		return
	}

	err := h.Template.Execute(w, tmplData)
	if err != nil {
		log.Printf("Error during simplehttp.PageHandler.Template.Execute: %v", err)
	}

}

// dirTimestamp finds the most recent time stamp associated with files in a folder
// TODO: we should look into file watcher stuff, better performance for large trees
func dirTimestamp(dir string) (ts time.Time, reterr error) {

	dirf, err := os.Open(dir)
	if err != nil {
		return ts, err
	}
	defer dirf.Close()

	fis, err := dirf.Readdir(-1)
	if err != nil {
		return ts, err
	}

	for _, fi := range fis {

		if fi.Name() == "." || fi.Name() == ".." {
			continue
		}

		// for directories we recurse
		if fi.IsDir() {
			dirTs, err := dirTimestamp(filepath.Join(dir, fi.Name()))
			if err != nil {
				return ts, err
			}
			if dirTs.After(ts) {
				ts = dirTs
			}
			continue
		}

		// for files check timestamp
		mt := fi.ModTime()
		if mt.After(ts) {
			ts = mt
		}
	}

	return
}
