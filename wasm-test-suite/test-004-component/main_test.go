package main

import (
	"log"
	"testing"

	"github.com/chromedp/chromedp"
	chromedpHelper "github.com/vugu/vugu/testing/chromedp"
	"github.com/vugu/vugu/testing/pkg"
	"github.com/vugu/vugu/testing/tmpl"
)

func Test004Component(t *testing.T) {
	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)
	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()
	// log.Printf("pathSuffix = %s", pathSuffix)

	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#testdiv"),
		chromedpHelper.WaitInnerTextTrimEq("ul", "0 a line is here\n1 a line is here\n2 a line is here"),
		chromedp.Click("#addbtn"),
		chromedpHelper.WaitInnerTextTrimEq("ul", "0 a line is here\n1 a line is here\n2 a line is here\n3 a line is here"),
	))

}
