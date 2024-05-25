package main

import (
	"log"
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"

	chromedpHelper "github.com/vugu/vugu/testing/chromedp"
	"github.com/vugu/vugu/testing/pkg"
	"github.com/vugu/vugu/testing/tmpl"
)

func Test019JSCreatePopulate(t *testing.T) {

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	assert := assert.New(t)
	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	var logText string
	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#the_button"), // make sure things showed up
		chromedp.Click("#the_button"),       // click the button
		chromedp.WaitVisible("#log"),        // wait for the log to show up
		chromedp.Click("#the_button"),       // click the button again
		chromedp.WaitVisible("#log_7"),      // wait for all of the log entries to show
		chromedp.Text(`#log`, &logText, chromedp.NodeVisible, chromedp.ByID),
	))
	// t.Logf("logText=%s", logText)
	logLines := strings.Split(logText, "\n")

	// vg-js-create first pass
	assert.Equal("vg-js-create className thing", strings.TrimSpace(logLines[0]))
	assert.Equal("vg-js-create firstElementChild <null>", strings.TrimSpace(logLines[1]))

	// vg-js-populate first pass
	assert.Equal("vg-js-populate className thing", strings.TrimSpace(logLines[2]))
	assert.Equal("vg-js-populate firstElementChild <object>", strings.TrimSpace(logLines[3]))

	// vg-js-create second pass
	assert.Equal("vg-js-create className thing", strings.TrimSpace(logLines[4]))
	assert.Equal("vg-js-create firstElementChild <object>", strings.TrimSpace(logLines[5]))

	// vg-js-populate second pass
	assert.Equal("vg-js-populate className thing", strings.TrimSpace(logLines[6]))
	assert.Equal("vg-js-populate firstElementChild <object>", strings.TrimSpace(logLines[7]))

}
