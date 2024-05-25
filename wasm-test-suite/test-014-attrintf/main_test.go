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

func Test014AttrIntf(t *testing.T) {

	pkgName := pkg.PkgName(t)
	tmpl.CreateIndexHtml(t, pkgName)

	url := "http://vugu-nginx/" + pkgName
	log.Printf("URL: %s", url)

	assert := assert.New(t)

	ctx, cancel := chromedpHelper.MustChromeCtx()
	defer cancel()

	attributeEq := func(ref, val string) chromedp.QueryAction {
		return chromedpHelper.QueryAttributes(ref, func(attributes map[string]string) {
			assert.Contains(attributes, "attr", "attribute on '%s' is missing", ref)
			assert.Equal(val, attributes["attr"], "attribute value on '%s' is invalid", ref)
		})
	}

	noAttribute := func(ref string) chromedp.QueryAction {
		return chromedpHelper.QueryAttributes(ref, func(attributes map[string]string) {
			assert.NotContains(attributes, "attr", "attribute on '%s' exists, but shouldn't", ref)
		})
	}

	chromedpHelper.Must(chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#testing"), // wait until render
		attributeEq("#plain_string", "string"),
		attributeEq("#string_var", "aString"),
		attributeEq("#string_ptr", "aString"),
		attributeEq("#int_var", "42"),
		attributeEq("#int_ptr", "42"),
		attributeEq("#true_var", "attr"),
		noAttribute("#false_var"),
		attributeEq("#true_ptr", "attr"),
		noAttribute("#false_ptr"),
		noAttribute("#string_nil_ptr"),
		attributeEq("#stringer", "myString"),
		noAttribute("#stringer_nil_ptr"),
	))
}
