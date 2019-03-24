package vugu

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DevHTTPHandler is a handler that makes developing web apps with vugu easier by handling
// the build process for you and providing some other sensible default behavior.
type DevHTTPHandler struct {
	BuildDir         string          // go package dir to build and run as wasm
	DisableCache     bool            // if true BuildDir is not checked for file changes before rebuilding
	PathPrefix       string          // prefix to answer requests for, defaults to root
	StaticFileSystem http.FileSystem // check for and serve static files from here, e.g. with http.Dir
	IndexOnly        bool            // if true then only "/" and "/index.html" will answer as HTML pages, whereas by default anything not otherwise satisfied returns the index page
	IndexTemplate    string          // the Go template to output for the HTML page, uses a sensible default
	ParserGoPkgOpts  ParserGoPkgOpts // so the ParserGoPkg behavior can be modified if needed
	// TODO: SkipGoGenerate - for speed or if not need for specific pjt workflow
	// TODO: SkipParserGoPkg - if go generate is already doing this work via vugugen

	// TODO: "hot module reloading"

	mu                 sync.RWMutex // so startup stuff can be exclusive
	errMsg             string       // if startup fails, server just returns this error
	wasmExecJSPath     string       // path to wasm_exec.js
	lastBuildTime      time.Time    // time of last successful build
	lastBuildContentGZ []byte       // last successful build gzipped
}

/*
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/css-spinning-spinners/1.1.1/load8.css" />
<div class="loader">Loading...</div>
<!-- span style="font-family:sans-serif">Loading...</style -->
*/

// NewDevHTTPHandler returns a DevHTTPHandler as configured and otherwise with sensible defaults.
func NewDevHTTPHandler(buildDir string, staticFileSystem http.FileSystem) *DevHTTPHandler {
	return &DevHTTPHandler{
		BuildDir:         buildDir,
		StaticFileSystem: staticFileSystem,
		IndexTemplate: `<!doctype html>
<html>
<head>
<title>Vugu Dev</title>
<meta charset="utf-8">
<script src="/wasm_exec.js"></script>
</head>
<body>
<div id="root_mount_parent">
<img style="position: absolute; top: 50%; left: 50%;" src="https://cdnjs.cloudflare.com/ajax/libs/galleriffic/2.0.1/css/loader.gif">
</div>
<script>
// FIXME: need to handle unloading properly and making sure the app exits and doesn't keep eating memory
// FIXME: check for wasm support and show an error message if not
const go = new Go();
WebAssembly.instantiateStreaming(fetch("/main.wasm"), go.importObject).then((result) => {
	go.run(result.instance);
});
</script>
</body>
</html>`,
		errMsg: "setup required",
	}
}

// ServeHTTP implements http.Handler.
func (h *DevHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		// already set-up, we're done
		if h.errMsg == "" {
			return
		}

		h.errMsg = ""
		// otherwise we just try again...

		b, err := exec.Command("go", "env", "GOROOT").CombinedOutput()
		if err != nil {
			h.errMsg = "failed to run `go env GOROOT`: " + err.Error()
			return
		}
		h.wasmExecJSPath = filepath.Join(strings.TrimSpace(string(b)), "misc/wasm/wasm_exec.js")
		// log.Printf("h.wasmExecJSPath = %q", h.wasmExecJSPath)

	}()

	// tell the browser to check with us every time, although it should be able to do 304s where appropriate,
	// should be be both fast and will keep things up to date upon page refresh
	w.Header().Set("Cache-control", "max-age=0, no-cache")

	// if errMsg is set then something failed at startup and we just show this error
	if h.errMsg != "" {
		http.Error(w, h.errMsg, 500)
		return
	}

	p := path.Clean("/" + r.URL.Path)

	// not found if prefix doesn't match
	pprefix := path.Clean("/" + h.PathPrefix)
	if len(pprefix) > 1 {

		// if path does not match prefix exactly or doesn't start with prefix+"/", 404
		if !(p == pprefix || strings.HasPrefix(p, pprefix+"/")) {
			http.Error(w, "404 not found (path doesn't match prefix)", 404)
			return
		}

		// found prefix, trim it
		p = strings.TrimPrefix(p, pprefix)
	}

	if p == "" {
		p = "/"
	}

	// if file extension, check static and serve if exist
	ext := path.Ext(p)
	if ext != "" && h.StaticFileSystem != nil {
		f, err := h.StaticFileSystem.Open(p)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("StaticFileSystem.Open(%q) returned error: %v", p, err)
			}
			goto notStatic
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			log.Printf("StaticFileSystem.Open(%q).Stat() returned error: %v", p, err)
			goto notStatic
		}
		http.ServeContent(w, r, p, fi.ModTime(), f)

		return
	}
notStatic:

	// wasm_exec.js
	if p == "/wasm_exec.js" {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, h.wasmExecJSPath)
		return
	}

	// main.wasm and build process, first check if it's needed
	if p == "/main.wasm" {

		h.mu.RLock()
		lastBuildTime := h.lastBuildTime
		lastBuildContentGZ := h.lastBuildContentGZ
		h.mu.RUnlock()

		buildDirTs, err := dirTimestamp(h.BuildDir)
		// log.Printf("1, buildDirTs = %v", buildDirTs)
		if err != nil {
			log.Printf("error in dirTimestamp(%q): %v", h.BuildDir, err)
			goto doBuild
		}

		if len(lastBuildContentGZ) == 0 {
			// log.Printf("2")
			goto doBuild
		}

		if h.DisableCache {
			goto doBuild
		}

		// skip build process if timestamp from build dir is equal or older than our last build
		if !buildDirTs.After(lastBuildTime) {
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
			pg := NewParserGoPkg(h.BuildDir, &h.ParserGoPkgOpts)
			err := pg.Run()
			if err != nil {
				msg := fmt.Sprintf("Error from ParserGoPkg: %v", err)
				log.Print(msg)
				http.Error(w, msg, 500)
				return
			}

			f, err := ioutil.TempFile("", "main_wasm_")
			if err != nil {
				panic(err)
			}
			fpath := f.Name()
			f.Close()
			os.Remove(f.Name())
			defer os.Remove(f.Name())

			startTime := time.Now()
			cmd := exec.Command("go", "generate", ".")
			cmd.Dir = h.BuildDir
			cmd.Env = append(cmd.Env, os.Environ()...)
			b, err := cmd.CombinedOutput()
			w.Header().Set("X-Go-Generate-Duration", time.Since(startTime).String())
			if err != nil {
				msg := fmt.Sprintf("Error from generate: %v; Output:\n%s", err, b)
				log.Print(msg)
				http.Error(w, msg, 500)
				return
			}

			// GOOS=js GOARCH=wasm go build -o main.wasm .
			startTime = time.Now()
			cmd = exec.Command("go", "build", "-o", fpath, ".")
			cmd.Dir = h.BuildDir
			cmd.Env = append(cmd.Env, os.Environ()...)
			cmd.Env = append(cmd.Env, "GOOS=js", "GOARCH=wasm")
			b, err = cmd.CombinedOutput()
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

			lastBuildTime = buildDirTs
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
			http.ServeContent(w, r, p, lastBuildTime, bytes.NewReader(lastBuildContentGZ))
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

	// if no file extension or if index only and is home, serve html
	if (h.IndexOnly && (p == "/" || p == "/index" || p == "/index.html")) || ext == "" {
		w.Header().Set("Content-Type", "text/html") // FIXME: cache headers?
		fmt.Fprint(w, h.IndexTemplate)              // FIXME: just printing for now but making this a html/template is probably better
		return
	}

	// otherwise 404
	http.NotFound(w, r)
}

// // returns true if content should be re-served, or false if 304 should be returned
// func checkIfModifiedSince(r *http.Request, modtime time.Time) bool {
// 	if r.Method != "GET" && r.Method != "HEAD" {
// 		return true
// 	}
// 	ims := r.Header.Get("If-Modified-Since")
// 	if ims == "" || modtime.IsZero() {
// 		return true
// 	}
// 	t, err := http.ParseTime(ims)
// 	if err != nil {
// 		return true
// 	}
// 	// The Date-Modified header truncates sub-second precision, so
// 	// use mtime < t+1s instead of mtime <= t to check for unmodified.
// 	if modtime.Before(t.Add(1 * time.Second)) {
// 		return false
// 	}
// 	return true
// }

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
