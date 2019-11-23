package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/vugu/vugu/distutil"
	"github.com/vugu/vugu/gen"
)

func TestBlah(t *testing.T) {

	debugURL := func() string {
		resp, err := http.Get("http://localhost:9222/json/version")
		if err != nil {
			t.Fatal(err)
		}

		var result map[string]interface{}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}
		return result["webSocketDebuggerUrl"].(string)
	}()

	t.Log(debugURL)

	allocCtx, cancel := chromedp.NewRemoteAllocator(context.Background(), debugURL)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx) //, chromedp.WithLogf(log.Printf))
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// MouseClickNode(n *cdp.Node, opts ...MouseOption) MouseAction

	var ss []string
	var text string
	err := chromedp.Run(ctx,
		// chromedp.Navigate("http://127.0.0.1:19944/"),
		chromedp.Navigate("https://www.google.com/"),
		// chromedp.Navigate("https://www.vugu.org/"),
		// chromedp.WaitVisible("#run"),
		chromedp.WaitVisible("[name=q]"),
		// chromedp.Click("#run"),
		// chromedp.WaitVisible("#success"),
		chromedp.InnerHTML("body", &text),
		chromedp.Evaluate(`Object.keys(window);`, &ss),
	)
	if err != nil {
		log.Fatal(err)
	}
	// t.Logf("text: %s", text)
	t.Logf("ss: %#v", ss)

	// MouseClickNode(n *cdp.Node, opts ...MouseOption) MouseAction

	t.Logf("HEY!")
}

func TestBlah2(t *testing.T) {

	dir, err := filepath.Abs("test-001-simple")
	if err != nil {
		panic(err)
	}

	distutil.Must(os.Chdir(dir))

	// TODO: read .gitignore and delete those files (and dirs?)

	pp := gen.NewParserGoPkg(dir, nil)
	err = pp.Run()
	if err != nil {
		panic(err)
	}

	// distutil.MustWasmExecJsPath()
	fmt.Print(distutil.MustEnvExec([]string{"GOOS=js", "GOARCH=wasm"}, "go", "build", "-o", filepath.Join(dir, "main.wasm"), "."))

	uploadPath := mustUploadDir(dir, "http://localhost:8846/upload")
	log.Printf("uploadPath = %q", uploadPath)

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
