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

func Test022EventListener(t *testing.T) {

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	assert := assert.New(t)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()
	// log.Printf("pathSuffix = %s", pathSuffix)

	var text string
	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#top"),
		chromedp.Click("#switch"),
		chromedp.WaitVisible("#noclick"),
		chromedp.Click("#noclick"),
		chromedp.Click("#switch"),
		chromedp.WaitNotPresent("#noclick"),
		chromedp.Click("#click"),
		chromedp.Click("#switch"),
		chromedp.WaitVisible("#text"),
		chromedp.InnerHTML("#text", &text),
	))

	assert.Equal("click", text)

}
