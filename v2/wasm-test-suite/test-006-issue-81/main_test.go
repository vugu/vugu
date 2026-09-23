package main

import (
	"log"
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	chromedpHelper "github.com/vugu/vugu/v2/testing/chromedp"
	"github.com/vugu/vugu/v2/testing/pkg"
	"github.com/vugu/vugu/v2/testing/tmpl"
)

func Test006Issue81(t *testing.T) {

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	assert := assert.New(t)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#content"),
		// chromedpHelper.QueryNode("html", func(n *cdp.Node) {
		// 	assert.Equal(
		// 		[]string{"class", "html-class", "lang", "en"},
		// 		n.Attributes,
		// 		"wrong html attributes",
		// 	)
		// }),
		// chromedpHelper.QueryNode("head", func(n *cdp.Node) {
		// 	assert.Equal(
		// 		[]string{"class", "head-class"},
		// 		n.Attributes,
		// 		"wrong head attributes",
		// 	)
		// }),
		chromedpHelper.QueryNode("div", func(n *cdp.Node) {
			assert.Equal(
				[]string{"class", "body-class"},
				n.Attributes,
				"wrong body attributes",
			)
		}),
	))

}

// func Test006Issue81DumpDocument(t *testing.T) {
// 	pkgName := pkg.PkgName(t)
// 	tmpl.CreateIndexHtml(t, pkgName)

// 	url := "http://vugu-nginx/" + pkgName
// 	log.Printf("URL: %s", url)

// 	ctx, cancel := chromedpHelper.MustChromeCtx()
// 	defer cancel()

// 	var nodes []*cdp.Node
// 	if err := chromedp.Run(ctx,
// 		chromedp.Navigate(url),
// 		chromedp.Nodes(`document`, &nodes,
// 			chromedp.ByJSPath, chromedp.Populate(-1, true)),
// 	); err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Println("Document tree:")
// 	fmt.Println(nodes[0].Dump("  ", "  ", false))
// 	t.Fatalf("This is the html and the <head> block from %q and not from %q. Conclusion. The <head> block is not replaced, so full html mode does not work.", "index.html", "root.vugu")
// }
