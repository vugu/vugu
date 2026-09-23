package main

import (
	"log"
	"testing"

	"github.com/chromedp/chromedp"

	chromedpHelper "github.com/vugu/vugu/testing/chromedp"
	"github.com/vugu/vugu/testing/pkg"
	"github.com/vugu/vugu/testing/tmpl"
)

func Test020VGForm(t *testing.T) {

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#food_group_value"), // make sure things showed up
		chromedp.SendKeys(`#food_group`, "Butterfinger Group"),
		chromedpHelper.WaitInnerTextTrimEq("#food_group_value", "butterfinger_group"),
	))

}
