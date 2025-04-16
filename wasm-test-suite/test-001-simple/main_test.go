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

func Test001Simple(t *testing.T) {

	t.Logf("Refactored test-001-simple running in networked docker containers!\n")
	ctx, cancel := chromedpHelper.MustChromeCtx()
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

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)
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
