package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
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

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/vugu/vugu/distutil"
	"github.com/vugu/vugu/gen"
	"github.com/vugu/vugu/simplehttp"
)

func Test001Simple(t *testing.T) {

	assert := assert.New(t)

	dir := mustUseDir("test-001-simple")
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()

	var innerHTML string
	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		chromedp.WaitVisible("#testing"),
		chromedp.InnerHTML("#testing", &innerHTML),
	))

	assert.Equal("Testing! A1", strings.TrimSpace(innerHTML))

}

// returns absdir
func mustUseDir(reldir string) string {

	dir, err := filepath.Abs(reldir)
	if err != nil {
		panic(err)
	}

	must(os.Chdir(dir))

	return dir
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

// func TestBlah(t *testing.T) {

// 	debugURL := func() string {
// 		resp, err := http.Get("http://localhost:9222/json/version")
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		var result map[string]interface{}

// 		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 			t.Fatal(err)
// 		}
// 		return result["webSocketDebuggerUrl"].(string)
// 	}()

// 	t.Log(debugURL)

// 	allocCtx, cancel := chromedp.NewRemoteAllocator(context.Background(), debugURL)
// 	defer cancel()

// 	ctx, cancel := chromedp.NewContext(allocCtx) //, chromedp.WithLogf(log.Printf))
// 	defer cancel()
// 	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
// 	defer cancel()

// 	// MouseClickNode(n *cdp.Node, opts ...MouseOption) MouseAction

// 	var ss []string
// 	var text string
// 	err := chromedp.Run(ctx,
// 		// chromedp.Navigate("http://127.0.0.1:19944/"),
// 		chromedp.Navigate("https://www.google.com/"),
// 		// chromedp.Navigate("https://www.vugu.org/"),
// 		// chromedp.WaitVisible("#run"),
// 		chromedp.WaitVisible("[name=q]"),
// 		// chromedp.Click("#run"),
// 		// chromedp.WaitVisible("#success"),
// 		chromedp.InnerHTML("body", &text),
// 		chromedp.Evaluate(`Object.keys(window);`, &ss),
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	// t.Logf("text: %s", text)
// 	t.Logf("ss: %#v", ss)

// 	// MouseClickNode(n *cdp.Node, opts ...MouseOption) MouseAction

// 	t.Logf("HEY!")
// }

// func TestBlah2(t *testing.T) {

// 	dir, err := filepath.Abs("test-001-simple")
// 	if err != nil {
// 		panic(err)
// 	}

// 	distutil.Must(os.Chdir(dir))

// 	// mustRmGitignoreFiles(filepath.Join(dir, ".gitignore"))
// 	mustCleanDir(dir)

// 	pp := gen.NewParserGoPkg(dir, nil)
// 	err = pp.Run()
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Print(distutil.MustEnvExec([]string{"GOOS=js", "GOARCH=wasm"}, "go", "build", "-o", filepath.Join(dir, "main.wasm"), "."))

// 	mustWriteSupportFiles(dir)

// 	uploadPath := mustUploadDir(dir, "http://localhost:8846/upload")
// 	log.Printf("uploadPath = %q", uploadPath)

// }

// func mustBuildAndLoadDir(dir string) string {

// 	dir, err := filepath.Abs("test-001-simple")
// 	if err != nil {
// 		panic(err)
// 	}

// 	distutil.Must(os.Chdir(dir))

// 	mustRmGitignoreFiles(filepath.Join(dir, ".gitignore"))

// 	pp := gen.NewParserGoPkg(dir, nil)
// 	err = pp.Run()
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Print(distutil.MustEnvExec([]string{"GOOS=js", "GOARCH=wasm"}, "go", "build", "-o", filepath.Join(dir, "main.wasm"), "."))

// 	mustWriteSupportFiles(dir)

// 	uploadPath := mustUploadDir(dir, "http://localhost:8846/upload")
// 	log.Printf("uploadPath = %q", uploadPath)

// }

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
