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
	//var tmpres []byte

	// regular version
	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#page1"), // make sure page1 showed initially
		chromedp.Click("#page2_link"),  // regular a href link
		chromedp.WaitVisible("#page2"), // make sure it loads
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
