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

func Test008ForI(t *testing.T) {

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	expectedText := "01234"
	expectedClicked := "0 clicked!"

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()
	// log.Printf("pathSuffix = %s", pathSuffix)

	var clicked string
	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#content"),
		chromedpHelper.WaitInnerTextTrimEq("#content", expectedText),
		chromedp.Click("#id0"),
		chromedp.WaitVisible("#clicked"),
		chromedp.InnerHTML("#clicked", &clicked),
	))
	assert.Equal(t, expectedClicked, clicked)

}
