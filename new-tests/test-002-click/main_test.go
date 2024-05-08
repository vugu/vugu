package main

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

type TestPath struct {
	TestDir string
}

func Test002Click(t *testing.T) {

	// The magefile mounts the test's parent directory as the directory into the nginx container
	// So we need to know our package name, which is the last component of the directory path.
	// We then need to append the package name onto the end of the URL passed to chromedp
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	pkgName := filepath.Base(cwd)

	t.Logf("CWD: %q", cwd)
	tp := TestPath{TestDir: pkgName}

	tmpl, err := template.ParseFiles(cwd + "/index.html.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	// remove any existing "index.html" - we don't care if the file does not exist
	err = os.Remove(cwd + "/index.html")
	if err != nil {
		t.Logf("rm error (not fatal) %s", err)
	}
	indexHTML, err := os.Create(cwd + "/index.html")
	if err != nil {
		t.Fatal(err)
	}

	err = tmpl.Execute(indexHTML, tp)
	if err != nil {
		t.Fatal(err)
	}
	indexHTML.Sync()
	err = indexHTML.Close()
	if err != nil {
		t.Fatal(err)
	}

	assert := assert.New(t)

	ctx, cancel := mustChromeCtx()
	defer cancel()
	log.Printf("pkgName = %s", pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)
	var text string
	must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#testdiv"),
		chromedp.WaitNotPresent("#success"),
		chromedp.Click("#run1"),
		chromedp.InnerHTML("#success", &text),
		chromedp.Click("#run1"),
		chromedp.WaitNotPresent("#success"),
	))
	log.Printf("Finished ChromeDp.Run()")
	assert.Equal("success", text)

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

	ctx, _ := chromedp.NewContext(allocCtx) // , chromedp.WithLogf(log.Printf))
	// defer cancel()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	// defer cancel()

	return ctx, cancel
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
