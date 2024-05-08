package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type TestPath struct {
	TestDir string
}

func Test003Prop(t *testing.T) {
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
	log.Printf("pkgName = %s", pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	ctx, cancel := mustChromeCtx()
	defer cancel()

	must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#email"),
		chromedp.SendKeys("#email", "joey@example.com"),
		chromedp.Blur("#email"),
		WaitInnerTextTrimEq("#emailout", "joey@example.com"),
		chromedp.Click("#resetbtn"),
		WaitInnerTextTrimEq("#emailout", "default@example.com"),
	))

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

// WaitInnerTextTrimEq will wait for the innerText of the specified element to match a specific string after whitespace trimming.
func WaitInnerTextTrimEq(sel, innerText string) chromedp.QueryAction {

	return chromedp.Query(sel, func(s *chromedp.Selector) {

		chromedp.WaitFunc(func(ctx context.Context, cur *cdp.Frame, id runtime.ExecutionContextID, ids ...cdp.NodeID) ([]*cdp.Node, error) {

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
				// log.Printf("found text: %s", ret)
				return nodes, errors.New("unexpected value: " + ret)
			}

			// log.Printf("NodeValue: %#v", nodes[0])

			// return nil, errors.New("not ready yet")
			return nodes, nil
		})(s)

	})

}
