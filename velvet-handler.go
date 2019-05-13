package fff

//
// This handler makes a bunch of assumptions about your project layout.
//
// 1) Everything ends up in assetPath (usually /assets in URL space
// and an immediate subdirectory of the server working dir called "assets")
// 2) Content pages (html) are addressed by logical name and served
// from the /assets/html directory.  So "/home") => "/assets/html/index.html"
// 3) Your content files share common header, nav, and footer sections,
// so the actual content portion is the only content in something like
// index.html.
// For #2 and #3 see helpers.go and the map logicalNameToLink.
// 4) Content files can have zero or one interactive portion handled
// by .vugu files.  Content files are converted to go source then to
// compiled wasm and placed in /assets/wasm/foo.wasm.
// 5) Content pages dictate the names of the "interactive" portion
// of their page.  For a content page called "foo" the resulting
// wasm file will be /assets/wasm/foo.wasm and the vugu files must
// be in ui/foo/*.vugu The parent directory is the same name as the
// content page's logical name.
// 6) The "root" component (the one that gets placed in the page)
// must have the same name as the logical name of the page,
// **EXCEPT** the first letter is capitalized.
// 7) You can change the helpers that are made visible in the templates
// by editing the function addMyHelpers.
// 8) You can change assetPath but *inside* assetPath there must exactly
// one level of directories, each named by the type of asset. E.g.
// img, js, html, etc.

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gobuffalo/velvet"
	"github.com/logrusorgru/aurora"
	"github.com/vugu/vugu/simplehttp"
)

// where the UI files are... these are the vugu files for
// more sophisticated interaction
const UIPath = "ui"

const vuguExt = ".vugu"
const wasmExt = ".wasm"

const mainFile = "main_wasm.go" // we generate this

const assetPath = "/assets"

//
// VelvetHandler is basically the vugu.SimpleHandler with a few tweaks
// to make it convenient to use velvet as a template language and
// have "active" parts of the page use vugu.
//
// Assets always are exact paths, like /assets/img/mypic.jpg
// But pages only use logical names, like /home
//

type VelvetHandler struct {
	root         string //path to the place where we store templates
	delegate     *simplehttp.SimpleHandler
	detailedLogs bool
}

func NewVelvetHandler(root string, wantDetails bool) *VelvetHandler { //the velvet glove treatment...
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("unable to get current working directory (%v), something is seriously broken", err)
	}
	simple := simplehttp.New(wd, true)
	result := &VelvetHandler{
		root:         root,
		delegate:     simple,
		detailedLogs: wantDetails,
	}
	result.addMyHelpers()
	result.delegate.IsPage = result.isPage
	result.delegate.PageHandler = result
	result.delegate.EnableGenerate = false
	return result
}

func (v *VelvetHandler) isPage(req *http.Request) bool {
	p := path.Clean("/" + req.URL.Path)
	ext := path.Ext(p)
	return ext == ""
}

func (v *VelvetHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	cleaned := path.Clean("/" + req.URL.Path)
	//special case for /
	if cleaned == "/" || cleaned == "/index" || cleaned == "/index.html" {
		resp.Header().Set("Location", "/home")
		resp.WriteHeader(http.StatusMovedPermanently)
		return
	}
	//we use extensions to know what is going on
	ext := path.Ext(cleaned)
	if ext == wasmExt { //WASM is special because in dev mode we run tools
		log.Printf("%s", aurora.Cyan(fmt.Sprintf("ServeHTTP (wasm) %s", cleaned)))
		v.serveWasm(resp, req, cleaned)
		return
	}
	if ext != "" {
		if v.detailedLogs {
			log.Printf("%s", aurora.Magenta(fmt.Sprintf("<<< Delegating %s to simplehttp", cleaned[1:])))
		}
		v.delegate.ServeHTTP(resp, req)
		return
	}
	candidate := v.logicalLink(cleaned)
	if candidate == "#" {
		http.NotFound(resp, req)
		return
	}
	log.Printf("%s", aurora.Cyan(fmt.Sprintf("ServeHTTP (page) %s -> %s", cleaned, candidate)))

	// only need [1:] when talking to the disk
	if _, err := os.Stat(candidate[1:]); err != nil {
		if os.IsNotExist(err) {
			if v.detailedLogs {
				log.Printf("checked %s: does not exist\n", candidate)
			}
		}
		log.Printf("unable to open %s: %v\n", candidate, err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	if v.detailedLogs {
		log.Printf("found logical name to path: %s -> %s", cleaned, candidate)
	}
	//note: only to do [1:] when touching the DISK
	components := []string{
		v.logicalLink("header")[1:],
		v.logicalLink("nav")[1:],
		candidate[1:],
		v.logicalLink("footer")[1:],
	}
	for _, c := range components {
		if err := v.loadVelvetAndSend(cleaned, c, resp); err != nil {
			log.Printf("failed to render %s: %v", c, err)
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	return
}

func (v *VelvetHandler) loadVelvetAndSend(page string, path string, resp http.ResponseWriter) error {
	fp, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = fp.Close() //nothing useful to do
	}()
	buffer, err := ioutil.ReadAll(fp)
	if err != nil {
		return err
	}
	if v.detailedLogs {
		log.Printf("loadVelvetAndSend %s ---> (%d bytes)\n",
			path, len(buffer))
	}
	out, err := v.renderVelvet(string(buffer), v.defaultData(page), nil)
	if err != nil {
		return err
	}
	_, err = io.Copy(resp, strings.NewReader(out))
	if err != nil {
		return err
	}
	return nil
}

func (v *VelvetHandler) renderVelvet(input string, data map[string]interface{}, helpers map[string]interface{}) (string, error) {
	t, err := velvet.Parse(input)
	if err != nil {
		return "", err
	}
	if v.detailedLogs {
		log.Printf("renderVelvet parsed %d bytes\n", len(input))
	}

	if helpers != nil {
		for k, v := range helpers {
			data[k] = v
		}
	}
	s, err := t.Exec(velvet.NewContextWith(data))
	if err != nil {
		return "", err
	}
	if v.detailedLogs {
		log.Printf("renderVelvet executed and produced %d bytes\n", len(s))
	}
	return s, nil
}

func (v *VelvetHandler) defaultData(page string) map[string]interface{} {
	if v.detailedLogs {
		log.Printf("defaultData: page='%s'\n", page)
	}
	result := make(map[string]interface{})
	result["page"] = page
	result["title"] = logicalNameToTitle[page]
	result["isHomePage"] = (page == "/home")
	return result
}
