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

var skipStr = `
Error Trace:	/home/owen/src/vugu/new-tests/test-008-for-keyvalue-noshadow/main_test.go:40
Error:      	Not equal: 
				expected: "4-e clicked!"
				actual  : "0-a clicked!"
				
				Diff:
				--- Expected
				+++ Actual
				@@ -1 +1 @@
				-4-e clicked!
				+0-a clicked!
Test:       	Test008ForKeyValueNoShadow
`

func Test008ForKeyValueNoShadow(t *testing.T) {
	t.Skipf("Skip - is this due to Go's for loop changes - currently failing with %s", skipStr)
	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	expectedText := "0-a1-b2-c3-d4-e"
	expectedClicked := "4-e clicked!"

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()
	// log.Printf("pathSuffix = %s", pathSuffix)

	var clicked string
	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#content"),
		chromedpHelper.WaitInnerTextTrimEq("#content", expectedText),
		chromedp.Click("#id0"),
		chromedp.WaitVisible("#clicked"),
		chromedp.InnerHTML("#clicked", &clicked),
	))
	assert.Equal(t, expectedClicked, clicked)

}
