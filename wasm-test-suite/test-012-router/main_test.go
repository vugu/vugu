package main

import (
	"log"
	"testing"

	"github.com/chromedp/chromedp"
	chromedpHelper "github.com/vugu/vugu/testing/chromedp"
	"github.com/vugu/vugu/testing/pkg"
	"github.com/vugu/vugu/testing/tmpl"
)

func Test012Router(t *testing.T) {

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	// rack the forward/back history stuff and replace option vs not and make sure that all works right
	// with fragment mode and without

	// FIXME This test, based on the original wasm test suite (test-012-router) is currently incomplete.
	// The problem is that the hard coded HTML hrefs in the page1/page2.vugu work, but the links that
	// are generated via the vgrouter and are called in response to the button clicks return a 404 (or timeout).
	// This can be seen by connecting to the nginx container locally with a web browser.
	// See the magefile for details of how to start the nginx docker container locally.
	// The wasm will be served with a path prefix of "/test-012-router". It's unclear if this is
	// related to the vgrouter routing at this stage.
	// The failure mode applies to both the non-fragment and fragment versions.
	//var tmpres []byte

	// regular version
	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#page1"), // make sure page1 showed initially
		chromedp.Click("#page2_link"),  // regular a href link
		chromedp.WaitVisible("#page2"), // make sure it loads
		// FIXME Uncomment the Click(...) and WaitVisible(...) lines to see the failure mode.
		// chromedp.Click("#page1_button"), // button goes to page1 without a reload
		// chromedp.WaitVisible("#page1"),  // make sure it renders
		// chromedp.Click("#page2_button_repl"),                   // go to page2 without adding to history
		// chromedp.WaitVisible("#page2"),                         // make sure it renders
		// chromedp.Evaluate("window.history.back()", &tmpres),    // go back one
		// chromedp.WaitVisible("#page2"),                         // should still be on page2 because of replace
		// chromedp.Evaluate("window.history.back()", &tmpres),    // go back one more
		// chromedp.WaitVisible("#page1"),                         // now should be on page1
		// chromedp.Evaluate("window.history.forward()", &tmpres), // forward one
		// chromedp.WaitVisible("#page2"),
	))

	// fragment version
	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url+"#/"),    // the test has detection code that sees the fragment here and puts it into fragment mode
		chromedp.WaitVisible("#page1"), // make sure page1 showed initially
		// FIXME Uncomment the Evaluate(...) and WaitVisible(...) lines to see the failure mode.
	// 	chromedp.Evaluate("window.location='#page2'", &tmpres), // browse to page2 via fragment
	// 	chromedp.WaitVisible("#page2"),                         // make sure it renders
	// 	chromedp.Click("#page1_button"),                                         // button goes to page1 without a reload
	// 	chromedp.WaitVisible("#page1"),                                          // make sure it renders
	// 	chromedp.Click("#page2_button_repl"),                                    // go to page2 without adding to history
	// 	chromedp.WaitVisible("#page2"),                                          // make sure it renders
	// 	chromedp.Evaluate("window.history.back()", &tmpres),                     // go back one
	// 	chromedp.WaitVisible("#page2"),                                          // should still be on page2 because of replace
	// 	chromedp.Evaluate("window.history.back()", &tmpres),                     // go back one more
	// 	chromedp.WaitVisible("#page1"),                                          // now should be on page1
	// 	chromedp.Evaluate("window.history.forward()", &tmpres),                  // forward one
	// 	chromedp.WaitVisible("#page2"),
	))

}
