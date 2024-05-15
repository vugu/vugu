package main

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"

	chromedpHelper "github.com/vugu/vugu/testing/chromedp"
	"github.com/vugu/vugu/testing/pkg"
	"github.com/vugu/vugu/testing/tmpl"
)

func Test024EventBufferSize(t *testing.T) {

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	ctx, ccancel := chromedpHelper.MustChromeCtx()
	defer ccancel()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	var val string
	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#top"),
		// trigger the change event
		chromedp.SendKeys("select", kb.ArrowDown+kb.ArrowDown),
		chromedp.Value(`select`, &val),
	))
	log.Println(val)
}
