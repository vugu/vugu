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

func Test002Click(t *testing.T) {

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	assert := assert.New(t)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()
	log.Printf("pkgName = %s", pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)
	var text string
	chromedpHelper.Must(chromedp.Run(ctx,
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
