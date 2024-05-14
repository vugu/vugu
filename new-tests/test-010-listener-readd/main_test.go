package main

import (
	"log"
	"testing"

	"github.com/chromedp/chromedp"

	chromedpHelper "github.com/vugu/vugu/testing/chromedp"
	"github.com/vugu/vugu/testing/pkg"
	"github.com/vugu/vugu/testing/tmpl"
)

func Test010ListenerReadd(t *testing.T) {
	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	// log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		// toggle back and forth a few times and make sure it continues to work
		chromedp.WaitVisible("#view1"),
		chromedp.Click("#switch_btn"),
		chromedp.WaitVisible("#view2"),
		chromedp.Click("#switch_btn"),
		chromedp.WaitVisible("#view1"),
		chromedp.Click("#switch_btn"),
		chromedp.WaitVisible("#view2"),
		chromedp.Click("#switch_btn"),
		chromedp.WaitVisible("#view1"),
	))
}
