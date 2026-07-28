package main

import (
	"fmt"
	"log"
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"

	chromedpHelper "github.com/vugu/vugu/testing/chromedp"
	"github.com/vugu/vugu/testing/pkg"
	"github.com/vugu/vugu/testing/tmpl"
)

func Test003Prop(t *testing.T) {
	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#email"),
		chromedp.SendKeys("#email", "joey@example.com"),
		chromedp.Blur("#email"),
		chromedpHelper.WaitInnerTextTrimEq("#emailout", "joey@example.com"),
		chromedp.Click("#resetbtn"),
		chromedpHelper.WaitInnerTextTrimEq("#emailout", "default@example.com"),
	))

}

func Test003PropTitleChange(t *testing.T) {
	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	title := ""
	expectedTitle := "Test page"
	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Title(&title),
	))
	if expectedTitle != title {
		t.Fatalf("Expect the page title to change to %q but it was %q", expectedTitle, title)
	}
}

func Test003PropDumpDocument(t *testing.T) {
	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	var nodes []*cdp.Node
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Nodes(`document`, &nodes,
			chromedp.ByJSPath, chromedp.Populate(-1, true)),
	); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Document tree:")
	fmt.Println(nodes[0].Dump("  ", "  ", false))
	t.Fatalf("This is the html and the <head> block from %q and not from %q. Conclusion. The <head> block is not replaced, so full html mode does not work.", "index.html", "root.vugu")
}
