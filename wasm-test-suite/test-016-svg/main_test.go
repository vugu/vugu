package main

import (
	"context"
	"log"
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"

	chromedpHelper "github.com/vugu/vugu/testing/chromedp"
	"github.com/vugu/vugu/testing/pkg"
	"github.com/vugu/vugu/testing/tmpl"
)

func Test016SVG(t *testing.T) {

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	assert := assert.New(t)
	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#icon"),          // wait for the icon to show up
		chromedp.WaitVisible("#icon polyline"), // make sure that the svg element is complete
		chromedp.QueryAfter("#icon", func(ctx context.Context, r runtime.ExecutionContextID, node ...*cdp.Node) error {
			// checking if the element is recognized as SVG by chrome should be enough
			assert.True(node[0].IsSVG)
			return nil
		}),
	))

}
