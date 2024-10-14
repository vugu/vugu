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

func Test015AttrList(t *testing.T) {

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)
	assert := assert.New(t)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#testing"), // wait until render
		chromedpHelper.QueryAttributes("#testing", func(attributes map[string]string) {
			assert.Contains(attributes, "class", "attribute is missing")
			assert.Equal("widget", attributes["class"], "attribute value is invalid")
		}),
		chromedpHelper.QueryAttributes("#testing", func(attributes map[string]string) {
			assert.Contains(attributes, "data-test", "attribute is missing")
			assert.Equal("test", attributes["data-test"], "attribute value is invalid")
		}),
		chromedpHelper.QueryAttributes("#functest", func(attributes map[string]string) {
			assert.Contains(attributes, "class", "attribute is missing")
			assert.Equal("funcwidget", attributes["class"], "attribute value is invalid")
		}),
		chromedpHelper.QueryAttributes("#functest", func(attributes map[string]string) {
			assert.Contains(attributes, "data-test", "attribute is missing")
			assert.Equal("functest", attributes["data-test"], "attribute value is invalid")
		}),
		chromedpHelper.QueryAttributes("#functest2", func(attributes map[string]string) { // check with vg-attr syntax
			assert.Contains(attributes, "class", "attribute is missing")
			assert.Equal("funcwidget", attributes["class"], "attribute value is invalid")
			assert.Contains(attributes, "data-test", "attribute is missing")
			assert.Equal("functest", attributes["data-test"], "attribute value is invalid")
		}),
	))
}
