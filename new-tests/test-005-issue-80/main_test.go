package main

import (
	"log"
	"testing"

	"github.com/chromedp/chromedp"
	chromedpHelper "github.com/vugu/vugu/testing/chromedp"
	"github.com/vugu/vugu/testing/pkg"
	"github.com/vugu/vugu/testing/tmpl"
)

func Test005Issue80(t *testing.T) {
	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()
	// log.Printf("pathSuffix = %s", pathSuffix)

	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#items"),
		chromedpHelper.WaitInnerTextTrimEq("#items", "abcd")))
}
