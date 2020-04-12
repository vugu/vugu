package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

// TO ADD A TEST:
// - make a folder of the same pattern test-NNN-description
// - copy .gitignore, go.mod and create a create.vugu, plus whatever else
// - write a TestNNNDescription method to drive it
// - to manually view the page from a test log the URL passed to chromedp.Navigate and view it in your browser
//   (if you suspect you are getting console errors that you can't see, this is a simple way to check)

func Test001Simple(t *testing.T) {

	assert := assert.New(t)

	dir, origDir := mustUseDir("test-001-simple")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()

	var t1, t2 string
	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		// chromedp.WaitVisible("#testing"),
		chromedp.InnerHTML("#t1", &t1), // NOTE: InnerHTML will wait until the element exists before returning
		chromedp.InnerHTML("#t2", &t2),
	))

	assert.Equal("t1text", strings.TrimSpace(t1))
	assert.Equal("t2text", strings.TrimSpace(t2))

}

func Test002Click(t *testing.T) {

	assert := assert.New(t)

	dir, origDir := mustUseDir("test-002-click")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()
	// log.Printf("pathSuffix = %s", pathSuffix)

	var text string
	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		chromedp.WaitVisible("#testdiv"),
		chromedp.WaitNotPresent("#success"),
		chromedp.Click("#run1"),
		chromedp.InnerHTML("#success", &text),
		chromedp.Click("#run1"),
		chromedp.WaitNotPresent("#success"),
	))

	assert.Equal("success", text)

}

func Test003Prop(t *testing.T) {

	assert := assert.New(t)

	dir, origDir := mustUseDir("test-003-prop")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()
	// log.Printf("pathSuffix = %s", pathSuffix)

	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		chromedp.WaitVisible("#email"),
		chromedp.SendKeys("#email", "joey@example.com"),
		chromedp.Blur("#email"),
		WaitInnerTextTrimEq("#emailout", "joey@example.com"),
		chromedp.Click("#resetbtn"),
		WaitInnerTextTrimEq("#emailout", "default@example.com"),
	))

	_ = assert
	// assert.Equal("success", text)

}

func Test004Component(t *testing.T) {

	assert := assert.New(t)

	dir, origDir := mustUseDir("test-004-component")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()
	// log.Printf("pathSuffix = %s", pathSuffix)

	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		chromedp.WaitVisible("#testdiv"),
		WaitInnerTextTrimEq("ul", "0 a line is here\n1 a line is here\n2 a line is here"),
		chromedp.Click("#addbtn"),
		WaitInnerTextTrimEq("ul", "0 a line is here\n1 a line is here\n2 a line is here\n3 a line is here"),
	))

	_ = assert

}

func Test005Issue80(t *testing.T) {

	assert := assert.New(t)

	dir, origDir := mustUseDir("test-005-issue-80")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()
	// log.Printf("pathSuffix = %s", pathSuffix)

	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		chromedp.WaitVisible("#items"),
		WaitInnerTextTrimEq("#items", "abcd"),
	))

	_ = assert

}

// TODO Rename it to Test006HtmlAttr ?
func Test006Issue81(t *testing.T) {

	assert := assert.New(t)

	dir, origDir := mustUseDir("test-006-issue-81")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()
	// log.Printf("pathSuffix = %s", pathSuffix)

	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		chromedp.WaitVisible("#content"),
		queryNode("html", func(n *cdp.Node) {
			assert.Equal(
				[]string{"class", "html-class", "lang", "en"},
				n.Attributes,
				"wrong html attributes",
			)
		}),
		queryNode("head", func(n *cdp.Node) {
			assert.Equal(
				[]string{"class", "head-class"},
				n.Attributes,
				"wrong head attributes",
			)
		}),
		queryNode("body", func(n *cdp.Node) {
			assert.Equal(
				[]string{"class", "body-class"},
				n.Attributes,
				"wrong body attributes",
			)
		}),
	))
}

func Test007Issue85(t *testing.T) {
	dir, origDir := mustUseDir("test-007-issue-85")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()

	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		chromedp.WaitVisible("#content"),
	))
}

func Test008For(t *testing.T) {
	tests := []struct {
		name            string
		dir             string
		expectedText    string
		expectedClicked string
	}{
		{
			name:            "for i",
			dir:             "test-008-for-i",
			expectedText:    "01234",
			expectedClicked: "0 clicked!",
		},
		{
			name:            "for no iteration vars",
			dir:             "test-008-for-keyvalue",
			expectedText:    "0-a1-b2-c3-d4-e",
			expectedClicked: "0-a clicked!",
		},
		{
			name:            "for with iteration vars",
			dir:             "test-008-for-kv",
			expectedText:    "0-a1-b2-c3-d4-e",
			expectedClicked: "0-a clicked!",
		},
		{
			name:            "for no iteration vars noshadow",
			dir:             "test-008-for-keyvalue-noshadow",
			expectedText:    "0-a1-b2-c3-d4-e",
			expectedClicked: "4-e clicked!",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, origDir := mustUseDir(tt.dir)
			defer os.Chdir(origDir)
			mustGen(dir)
			pathSuffix := mustBuildAndLoad(dir)
			ctx, cancel := mustChromeCtx()
			defer cancel()
			// log.Printf("pathSuffix = %s", pathSuffix)

			var clicked string
			must(chromedp.Run(ctx,
				chromedp.Navigate("http://localhost:8846"+pathSuffix),
				chromedp.WaitVisible("#content"),
				WaitInnerTextTrimEq("#content", tt.expectedText),
				chromedp.Click("#id0"),
				chromedp.WaitVisible("#clicked"),
				chromedp.InnerHTML("#clicked", &clicked),
			))
			assert.Equal(t, tt.expectedClicked, clicked)
		})
	}
}

func Test009TrimUnused(t *testing.T) {
	dir, origDir := mustUseDir("test-009-trim-unused")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()

	// log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		chromedp.WaitVisible("#content"),
		chromedp.Click("#make2"),
		chromedp.WaitVisible("#n2of2"),
		chromedp.Click("#make6"),
		chromedp.WaitVisible("#n2of6"),
		chromedp.WaitVisible("#n6of6"),
		chromedp.Click("#make2"),
		chromedp.WaitNotPresent("#n6of6"),
		chromedp.WaitVisible("#n2of2"),
	))
}

func Test010ListenerReadd(t *testing.T) {
	dir, origDir := mustUseDir("test-010-listener-readd")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()

	// log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		// toggle back and forth a few times and make sure it continues to work
		chromedp.WaitVisible("#view1"),
		chromedp.Click("#switch_btn"),
		chromedp.WaitVisible("#view2"),
		chromedp.Click("#switch_btn"),
		chromedp.WaitVisible("#view1"),
		chromedp.Click("#switch_btn"),
		chromedp.WaitVisible("#view2"),
		chromedp.Click("#switch_btn"),
		chromedp.WaitVisible("#view1"),
	))
}

func Test011Wire(t *testing.T) {

	dir, origDir := mustUseDir("test-011-wire")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()

	log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		WaitInnerTextTrimEq(".demo-comp1-c", "1"),
		WaitInnerTextTrimEq(".demo-comp2-c", "2"),
	))

}

func Test012Router(t *testing.T) {

	dir, origDir := mustUseDir("test-012-router")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()

	log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

	// rack the forward/back history stuff and replace option vs not and make sure that all works right
	// with fragment mode and without
	var tmpres []byte

	// regular version
	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		chromedp.WaitVisible("#page1"),                         // make sure page1 showed initially
		chromedp.Click("#page2_link"),                          // regular a href link
		chromedp.WaitVisible("#page2"),                         // make sure it loads
		chromedp.Click("#page1_button"),                        // button goes to page1 without a reload
		chromedp.WaitVisible("#page1"),                         // make sure it renders
		chromedp.Click("#page2_button_repl"),                   // go to page2 without adding to history
		chromedp.WaitVisible("#page2"),                         // make sure it renders
		chromedp.Evaluate("window.history.back()", &tmpres),    // go back one
		chromedp.WaitVisible("#page2"),                         // should still be on page2 because of replace
		chromedp.Evaluate("window.history.back()", &tmpres),    // go back one more
		chromedp.WaitVisible("#page1"),                         // now should be on page1
		chromedp.Evaluate("window.history.forward()", &tmpres), // forward one
		chromedp.WaitVisible("#page2"),
	))

	// fragment version
	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix+"#/"), // the test has detection code that sees the fragment here and puts it into fragment mode
		chromedp.WaitVisible("#page1"),                             // make sure page1 showed initially
		chromedp.Evaluate("window.location='#/page2'", &tmpres),    // browse to page2 via fragment
		chromedp.WaitVisible("#page2"),                             // make sure it renders
		chromedp.Click("#page1_button"),                            // button goes to page1 without a reload
		chromedp.WaitVisible("#page1"),                             // make sure it renders
		chromedp.Click("#page2_button_repl"),                       // go to page2 without adding to history
		chromedp.WaitVisible("#page2"),                             // make sure it renders
		chromedp.Evaluate("window.history.back()", &tmpres),        // go back one
		chromedp.WaitVisible("#page2"),                             // should still be on page2 because of replace
		chromedp.Evaluate("window.history.back()", &tmpres),        // go back one more
		chromedp.WaitVisible("#page1"),                             // now should be on page1
		chromedp.Evaluate("window.history.forward()", &tmpres),     // forward one
		chromedp.WaitVisible("#page2"),
	))

}

func Test013Issue117(t *testing.T) {

	dir, origDir := mustUseDir("test-013-issue-117")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()

	log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		chromedp.WaitVisible("#create_button"), // make sure page1 showed initially
		chromedp.Click("#create_button"),       // regular a href link
		chromedp.WaitVisible("#myform"),        // make sure it loads
		chromedp.WaitNotPresent("#mytable"),
	))

}

func Test014AttrIntf(t *testing.T) {

	assert := assert.New(t)
	dir, origDir := mustUseDir("test-014-attrintf")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()

	log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

	attributeEq := func(ref, val string) chromedp.QueryAction {
		return queryAttributes(ref, func(attributes map[string]string) {
			assert.Contains(attributes, "attr", "attribute on '%s' is missing", ref)
			assert.Equal(val, attributes["attr"], "attribute value on '%s' is invalid", ref)
		})
	}

	noAttribute := func(ref string) chromedp.QueryAction {
		return queryAttributes(ref, func(attributes map[string]string) {
			assert.NotContains(attributes, "attr", "attribute on '%s' exists, but shouldn't", ref)
		})
	}

	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
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

func Test015AttrList(t *testing.T) {

	assert := assert.New(t)
	dir, origDir := mustUseDir("test-015-attribute-lister")
	defer os.Chdir(origDir)
	mustGen(dir)
	pathSuffix := mustBuildAndLoad(dir)
	ctx, cancel := mustChromeCtx()
	defer cancel()

	log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		chromedp.WaitVisible("#testing"), // wait until render
		queryAttributes("#testing", func(attributes map[string]string) {
			assert.Contains(attributes, "class", "attribute is missing")
			assert.Equal("widget", attributes["class"], "attribute value is invalid")
		}),
		queryAttributes("#testing", func(attributes map[string]string) {
			assert.Contains(attributes, "data-test", "attribute is missing")
			assert.Equal("test", attributes["data-test"], "attribute value is invalid")
		}),
		queryAttributes("#functest", func(attributes map[string]string) {
			assert.Contains(attributes, "class", "attribute is missing")
			assert.Equal("funcwidget", attributes["class"], "attribute value is invalid")
		}),
		queryAttributes("#functest", func(attributes map[string]string) {
			assert.Contains(attributes, "data-test", "attribute is missing")
			assert.Equal("functest", attributes["data-test"], "attribute value is invalid")
		}),
		queryAttributes("#functest2", func(attributes map[string]string) { // check with vg-attr syntax
			assert.Contains(attributes, "class", "attribute is missing")
			assert.Equal("funcwidget", attributes["class"], "attribute value is invalid")
			assert.Contains(attributes, "data-test", "attribute is missing")
			assert.Equal("functest", attributes["data-test"], "attribute value is invalid")
		}),
	))
}

func Test100TinygoSimple(t *testing.T) {

	// TODO: This is work in progress - it does actually compile but needs some more work to
	// get files into the right places, pull in the correct wasm_exec.js and then we need
	// to actually test the execution.
	t.SkipNow()

	assert := assert.New(t)

	dir, origDir := mustUseDir("test-100-tinygo-simple")
	defer os.Chdir(origDir)

	buildGopath := mustTGTempGopathSetup(dir, "src/tgtestpgm")
	log.Printf("buildGopath: %s", buildGopath)
	mustTGGoGet(buildGopath, "github.com/vugu/xxhash", "github.com/vugu/vjson")
	mustTGGen(filepath.Join(buildGopath, "src/tgtestpgm"))
	// pathSuffix := mustTGBuildAndLoad(filepath.Join(dir, "main.wasm"), buildGopath)
	pathSuffix := mustTGBuildAndLoad(dir, buildGopath)

	ctx, cancel := mustChromeCtx()
	defer cancel()
	log.Printf("pathSuffix = %s", pathSuffix)

	must(chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8846"+pathSuffix),
		// chromedp.WaitVisible("#items"),
		// WaitInnerTextTrimEq("#items", "abcd"),
	))

	_ = assert

	// if it passes then remove the temp dir
	os.RemoveAll(buildGopath)

}
