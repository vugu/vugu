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

func Test018CompEvents(t *testing.T) {

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)
	assert := assert.New(t)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	var showText string
	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#top"),  // make sure things showed up
		chromedp.Click("#the_button"), // click the button inside the component

		chromedp.WaitVisible("#show_text"), // wait for it to dump event out
		chromedp.InnerHTML("#show_text", &showText),
	))
	//t.Logf("showText=%s", showText)
	assert.Contains(showText, "ClickEvent")

}
