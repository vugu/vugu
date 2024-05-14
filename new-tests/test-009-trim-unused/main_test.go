package main

import (
	"log"
	"testing"

	"github.com/chromedp/chromedp"

	chromedpHelper "github.com/vugu/vugu/testing/chromedp"
	"github.com/vugu/vugu/testing/pkg"
	"github.com/vugu/vugu/testing/tmpl"
)

func Test009TrimUnused(t *testing.T) {
	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	// log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#content"),
		chromedp.Click("#make2"),
		chromedp.WaitVisible("#n2of2"),
		chromedp.Click("#make6"),
		chromedp.WaitVisible("#n2of6"),
		chromedp.WaitVisible("#n6of6"),
		chromedp.Click("#make2"),
		chromedp.WaitNotPresent("#n6of6"),
		chromedp.WaitVisible("#n2of2"),
	))

}
