package main

import (
	"log"
	"testing"

	"github.com/chromedp/chromedp"
	chromedpHelper "github.com/vugu/vugu/testing/chromedp"
	"github.com/vugu/vugu/testing/pkg"
	"github.com/vugu/vugu/testing/tmpl"
)

func Test013Issue117(t *testing.T) {
	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#create_button"), // make sure page1 showed initially
		chromedp.Click("#create_button"),       // regular a href link
		chromedp.WaitVisible("#myform"),        // make sure it loads
		chromedp.WaitNotPresent("#mytable"),
	))

}
