package main

import (
	"log"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"

	chromedpHelper "github.com/vugu/vugu/testing/chromedp"
	"github.com/vugu/vugu/testing/pkg"
	"github.com/vugu/vugu/testing/tmpl"
)

// TO ADD A TEST:
// - make a folder of the same pattern test-NNN-description
// - copy .gitignore, go.mod and create a root.vugu, plus whatever else
// - write a TestNNNDescription method to drive it
// - to manually view the page from a test log the URL passed to chromedp.Navigate and view it in your browser
//   (if you suspect you are getting console errors that you can't see, this is a simple way to check)

// Adapted from test-001-simple to show Issue #328 vugu.js.ValueOf causing a panic.
func Test025ValueOf(t *testing.T) {

	t.Logf("Refactored test-025-valueof running in networked docker containers!\n")
	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	// just a single test case in this case.
	cases := []struct {
		id       string
		expected string
	}{
		{"date_object", "Ok"},
	}

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)
	// connect teh headless browser to the URL in the private docker network
	actions := []chromedp.Action{chromedp.Navigate(url)}

	tout := make([]string, len(cases))
	for i, c := range cases {
		actions = append(actions, chromedp.InnerHTML("#"+c.id, &tout[i]))
	}

	chromedpHelper.Must(chromedp.Run(ctx, actions...))

	for i, c := range cases {
		i, c := i, c
		t.Run(c.id, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(c.expected, tout[i])
		})
	}

}
