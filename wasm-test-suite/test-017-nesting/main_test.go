package main

import (
	"log"
	"testing"

	"github.com/chromedp/chromedp"

	chromedpHelper "github.com/vugu/vugu/testing/chromedp"
	"github.com/vugu/vugu/testing/pkg"
	"github.com/vugu/vugu/testing/tmpl"
)

func Test017Nesting(t *testing.T) {

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#final1"), // make sure things showed up
		chromedp.WaitVisible("#final2"), // make sure things showed up

		chromedp.Click("#final1 .clicker"),            // click top one
		chromedp.WaitVisible("#final1 .clicked-true"), // should get clicked on #1
		chromedp.WaitNotPresent("#final1 .clicked-false"),
		chromedp.WaitNotPresent("#final2 .clicked-true"), // but not on #2
		chromedp.WaitVisible("#final2 .clicked-false"),

		// now check the reverse

		chromedp.Navigate(url),
		chromedp.WaitVisible("#final1"), // make sure things showed up
		chromedp.WaitVisible("#final2"), // make sure things showed up

		chromedp.Click("#final2 .clicker"),            // click bottom one
		chromedp.WaitVisible("#final2 .clicked-true"), // should get clicked on #2
		chromedp.WaitNotPresent("#final2 .clicked-false"),
		chromedp.WaitNotPresent("#final1 .clicked-true"), // but not on #1
		chromedp.WaitVisible("#final1 .clicked-false"),
	))

}
