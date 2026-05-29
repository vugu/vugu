package main

import (
	"log"
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	chromedpHelper "github.com/vugu/vugu/testing/chromedp"
	"github.com/vugu/vugu/testing/pkg"
	"github.com/vugu/vugu/testing/tmpl"
)

func Test006Issue81(t *testing.T) {

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	assert := assert.New(t)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#content"),
		chromedpHelper.QueryNode("html", func(n *cdp.Node) {
			assert.Equal(
				[]string{"class", "html-class", "lang", "en"},
				n.Attributes,
				"wrong html attributes",
			)
		}),
		chromedpHelper.QueryNode("head", func(n *cdp.Node) {
			assert.Equal(
				[]string{"class", "head-class"},
				n.Attributes,
				"wrong head attributes",
			)
		}),
		chromedpHelper.QueryNode("body", func(n *cdp.Node) {
			assert.Equal(
				[]string{"class", "body-class"},
				n.Attributes,
				"wrong body attributes",
			)
		}),
	))

}
