package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

// TO ADD A TEST:
// - make a folder of the same pattern test-NNN-description
// - copy .gitignore, go.mod and create a root.vugu, plus whatever else
// - write a TestNNNDescription method to drive it
// - to manually view the page from a test log the URL passed to chromedp.Navigate and view it in your browser
//   (if you suspect you are getting console errors that you can't see, this is a simple way to check)

func Test001Simple(t *testing.T) {

	t.Logf("Refactored test-001-simple running in networked docker containers!\n")
	ctx, cancel := mustChromeCtx()
	defer cancel()

	cases := []struct {
		id       string
		expected string
	}{
		{"t0", "t0text"},
		{"t1", "t1text"},
		{"t2", "t2text"},
		{"t3", "&amp;amp;"},
		{"t4", "&amp;"},
		{"t5", "false"},
		{"t6", "10"},
		{"t7", "20.000000"},
		{"t8", ""},
		{"t9", "S-HERE:blah"},
	}

	tout := make([]string, len(cases))

	//log.Printf("URL: http://localhost:8888")
	actions := []chromedp.Action{chromedp.Navigate("http://vugu-nginx")}
	for i, c := range cases {
		actions = append(actions, chromedp.InnerHTML("#"+c.id, &tout[i]))
	}

	must(chromedp.Run(ctx, actions...))

	for i, c := range cases {
		i, c := i, c
		t.Run(c.id, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(c.expected, tout[i])
		})
	}

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
