package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/vugu/vugu/distutil"
	"github.com/vugu/vugu/gen"
	"github.com/vugu/vugu/simplehttp"
)

// TO ADD A TEST:
// - make a folder of the same pattern test-NNN-description
// - copy .gitignore, go.mod and create a root.vugu, plus whatever else
// - write a TestNNNDescription method to drive it
// - to manually view the page from a test log the URL passed to chromedp.Navigate and view it in your browser
//   (if you suspect you are getting console errors that you can't see, this is a simple way to check)

func Test001Simple(t *testing.T) {

	assert := assert.New(t)

	dir, origDir := mustUseDir("test-001-simple")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()

	var t1, t2 string
	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		// chromedp.WaitVisible("#testing"),
		chromedp.InnerHTML("#t1", &t1), // NOTE: InnerHTML will wait until the element exists before returning
		chromedp.InnerHTML("#t2", &t2),
	))

	assert.Equal("t1text", strings.TrimSpace(t1))
	assert.Equal("t2text", strings.TrimSpace(t2))

}

func Test002Click(t *testing.T) {

	assert := assert.New(t)

	dir, origDir := mustUseDir("test-002-click")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()
	// log.Printf("pathSuffix = %s", pathSuffix)

	var text string
	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		chromedp.WaitVisible("#testdiv"),
		chromedp.WaitNotPresent("#success"),
		chromedp.Click("#run1"),
		chromedp.InnerHTML("#success", &text),
		chromedp.Click("#run1"),
		chromedp.WaitNotPresent("#success"),
	))

	assert.Equal("success", text)

}

func Test003Prop(t *testing.T) {

	assert := assert.New(t)

	dir, origDir := mustUseDir("test-003-prop")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()
	// log.Printf("pathSuffix = %s", pathSuffix)

	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		chromedp.WaitVisible("#email"),
		chromedp.SendKeys("#email", "joey@example.com"),
		chromedp.Blur("#email"),
		WaitInnerTextTrimEq("#emailout", "joey@example.com"),
		chromedp.Click("#resetbtn"),
		WaitInnerTextTrimEq("#emailout", "default@example.com"),
	))

	_ = assert
	// assert.Equal("success", text)

}

func Test004Component(t *testing.T) {

	assert := assert.New(t)

	dir, origDir := mustUseDir("test-004-component")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()
	// log.Printf("pathSuffix = %s", pathSuffix)

	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		chromedp.WaitVisible("#testdiv"),
		WaitInnerTextTrimEq("ul", "0 a line is here\n1 a line is here\n2 a line is here"),
		chromedp.Click("#addbtn"),
		WaitInnerTextTrimEq("ul", "0 a line is here\n1 a line is here\n2 a line is here\n3 a line is here"),
	))

	_ = assert

}

func Test005Issue80(t *testing.T) {

	assert := assert.New(t)

	dir, origDir := mustUseDir("test-005-issue-80")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()
	// log.Printf("pathSuffix = %s", pathSuffix)

	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		chromedp.WaitVisible("#items"),
		WaitInnerTextTrimEq("#items", "abcd"),
	))

	_ = assert

}

// WaitInnerTextTrimEq will wait for the innerText of the specified element to match a specific string after whitespace trimming.
func WaitInnerTextTrimEq(sel, innerText string) chromedp.QueryAction {

	return chromedp.Query(sel, func(s *chromedp.Selector) {

		chromedp.WaitFunc(func(ctx context.Context, cur *cdp.Frame, ids ...cdp.NodeID) ([]*cdp.Node, error) {

			nodes := make([]*cdp.Node, len(ids))
			cur.RLock()
			for i, id := range ids {
				nodes[i] = cur.Nodes[id]
				if nodes[i] == nil {
					cur.RUnlock()
					// not yet ready
					return nil, nil
				}
			}
			cur.RUnlock()

			var ret string
			err := chromedp.EvaluateAsDevTools("document.querySelector('"+sel+"').innerText", &ret).Do(ctx)
			if err != nil {
				return nodes, err
			}
			if strings.TrimSpace(ret) != innerText {
				//log.Printf("found text: %s", ret)
				return nodes, errors.New("unexpected value: " + ret)
			}

			// log.Printf("NodeValue: %#v", nodes[0])

			// return nil, errors.New("not ready yet")
			return nodes, nil
		})(s)

	})

}

// returns absdir
func mustUseDir(reldir string) (newdir, olddir string) {

	odir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	olddir = odir

	dir, err := filepath.Abs(reldir)
	if err != nil {
		panic(err)
	}

	must(os.Chdir(dir))

	newdir = dir

	return
}

func mustGen(absdir string) {

	pp := gen.NewParserGoPkg(absdir, nil)
	err := pp.Run()
	if err != nil {
		panic(err)
	}

}

// returns path suffix
func mustBuildAndLoad(absdir string) string {

	fmt.Print(distutil.MustEnvExec([]string{"GOOS=js", "GOARCH=wasm"}, "go", "build", "-o", filepath.Join(absdir, "main.wasm"), "."))

	mustWriteSupportFiles(absdir)

	uploadPath := mustUploadDir(absdir, "http://localhost:8846/upload")
	// log.Printf("uploadPath = %q", uploadPath)

	return uploadPath
}

func mustChromeCtx() (context.Context, context.CancelFunc) {

	debugURL := func() string {
		resp, err := http.Get("http://localhost:9222/json/version")
		if err != nil {
			panic(err)
		}

		var result map[string]interface{}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			panic(err)
		}
		return result["webSocketDebuggerUrl"].(string)
	}()

	// t.Log(debugURL)

	allocCtx, _ := chromedp.NewRemoteAllocator(context.Background(), debugURL)
	// defer cancel()

	ctx, _ := chromedp.NewContext(allocCtx) //, chromedp.WithLogf(log.Printf))
	// defer cancel()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	// defer cancel()

	return ctx, cancel
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustCleanDir(dir string) {
	must(os.Chdir(dir))
	b, err := ioutil.ReadFile(".gitignore")
	if err != nil {
		panic(err)
	}
	ss := strings.Split(string(b), "\n")
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if s == "" || s == "." || s == ".." || strings.HasPrefix(s, "#") {
			continue
		}
		// log.Printf("removing: %s", s)
		os.Remove(s)
	}

}

// mustWriteSupportFiles will write index.html and wasm_exec.js to a directory
func mustWriteSupportFiles(dir string) {
	distutil.MustCopyFile(distutil.MustWasmExecJsPath(), filepath.Join(dir, "wasm_exec.js"))
	// distutil.MustCopyFile(distutil.(), filepath.Join(dir, "wasm_exec.js"))
	// log.Println(simplehttp.DefaultPageTemplateSource)

	// "/wasm_exec.js"

	var buf bytes.Buffer

	req, _ := http.NewRequest("GET", "/index.html", nil)
	outf, err := os.OpenFile(filepath.Join(dir, "index.html"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	distutil.Must(err)
	defer outf.Close()
	template.Must(template.New("_page_").Parse(simplehttp.DefaultPageTemplateSource)).Execute(&buf, map[string]interface{}{"Request": req})
	// HACK: fix wasm_exec.js path, unti we can come up with a better way to do this
	outf.Write(
		bytes.Replace(
			bytes.Replace(buf.Bytes(), []byte(`"/wasm_exec.js"`), []byte(`"wasm_exec.js"`), 1),
			[]byte("/main.wasm"), []byte("main.wasm"), 1,
		),
	)

}

// mustUploadDir tar+gz's the given directory and posts that file to the specified endpoint,
// returning the path of where to access the files
func mustUploadDir(dir, endpoint string) string {

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	absDir, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}

	// var hdr tar.Header
	err = filepath.Walk(dir, filepath.WalkFunc(func(fpath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		absPath, err := filepath.Abs(fpath)
		if err != nil {
			panic(err)
		}

		relPath := path.Clean("/" + strings.TrimPrefix(absPath, absDir))

		// log.Printf("path = %q, fi.Name = %q", path, fi.Name())
		hdr, err := tar.FileInfoHeader(fi, "")
		hdr.Name = relPath
		// hdr = tar.Header{
		// 	Name: fi.Name(),
		// 	Mode: 0644,
		// 	Size: fi.Size(),
		// }
		err = tw.WriteHeader(hdr)
		if err != nil {
			return err
		}

		// no body to write for directories
		if fi.IsDir() {
			return nil
		}

		inf, err := os.Open(fpath)
		if err != nil {
			return err
		}
		defer inf.Close()
		_, err = io.Copy(tw, inf)

		if err != nil {
			return err
		}

		return nil
	}))
	if err != nil {
		panic(err)
	}

	tw.Close()
	gw.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("archive", filepath.Base(dir)+".tar.gz")
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(part, &buf)
	if err != nil {
		panic(err)
	}
	err = writer.Close()
	if err != nil {
		panic(err)
	}

	// for key, val := range params {
	// 	_ = writer.WriteField(key, val)
	// }
	// err = writer.Close()
	// if err != nil {
	// 	return nil, err
	// }

	req, err := http.NewRequest("POST", endpoint, &body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	var retData struct {
		Path string `json:"path"`
	}
	err = json.NewDecoder(res.Body).Decode(&retData)
	if err != nil {
		panic(err)
	}
	return retData.Path
}
