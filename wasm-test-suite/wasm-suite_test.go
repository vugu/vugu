package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/stretchr/testify/assert"
)

// TO ADD A TEST:
// - make a folder of the same pattern test-NNN-description
// - copy .gitignore, go.mod and create a root.vugu, plus whatever else
// - write a TestNNNDescription method to drive it
// - to manually view the page from a test log the URL passed to chromedp.Navigate and view it in your browser
//   (if you suspect you are getting console errors that you can't see, this is a simple way to check)

func Test001Simple(t *testing.T) {

	dir, origDir := mustUseDir("test-001-simple")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

		ctx, cancel := mustChromeCtx()
		defer cancel()

		cases := []struct {
			id       string
			expected string
		}{
			{"t0", "t0text"},
			{"t1", "t1text"},
			{"t2", "t2text"},
			{"t3", "&amp;amp;"},
			{"t4", "&amp;"},
			{"t5", "false"},
			{"t6", "10"},
			{"t7", "20.000000"},
			{"t8", ""},
			{"t9", "S-HERE:blah"},
		}

		tout := make([]string, len(cases))

		log.Printf("URL: http://localhost:8846%s", pathSuffix)
		actions := []chromedp.Action{chromedp.Navigate("http://localhost:8846" + pathSuffix)}
		for i, c := range cases {
			actions = append(actions, chromedp.InnerHTML("#"+c.id, &tout[i]))
		}

		must(chromedp.Run(ctx, actions...))

		for i, c := range cases {
			i, c := i, c
			t.Run(c.id, func(t *testing.T) {
				assert := assert.New(t)
				assert.Equal(c.expected, tout[i])
			})
		}

	}

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })

}

func Test002Click(t *testing.T) {

	dir, origDir := mustUseDir("test-002-click")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

		assert := assert.New(t)

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

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })

}

func Test003Prop(t *testing.T) {

	dir, origDir := mustUseDir("test-003-prop")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

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

	}

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })
}

func Test004Component(t *testing.T) {

	dir, origDir := mustUseDir("test-004-component")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

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

	}

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })
}

func Test005Issue80(t *testing.T) {

	dir, origDir := mustUseDir("test-005-issue-80")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

		ctx, cancel := mustChromeCtx()
		defer cancel()
		// log.Printf("pathSuffix = %s", pathSuffix)

		must(chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:8846"+pathSuffix),
			chromedp.WaitVisible("#items"),
			WaitInnerTextTrimEq("#items", "abcd"),
		))

	}

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })
}

// TODO Rename it to Test006HtmlAttr ?
func Test006Issue81(t *testing.T) {

	dir, origDir := mustUseDir("test-006-issue-81")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

		assert := assert.New(t)

		ctx, cancel := mustChromeCtx()
		defer cancel()
		log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

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

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })
}

func Test007Issue85(t *testing.T) {
	dir, origDir := mustUseDir("test-007-issue-85")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

		ctx, cancel := mustChromeCtx()
		defer cancel()

		must(chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:8846"+pathSuffix),
			chromedp.WaitVisible("#content"),
		))

	}

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })
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

			tf := func(t *testing.T, pathSuffix string) {

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

			}

			t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
			t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
			t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })
		})
	}
}

func Test009TrimUnused(t *testing.T) {
	dir, origDir := mustUseDir("test-009-trim-unused")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

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

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })
}

func Test010ListenerReadd(t *testing.T) {
	dir, origDir := mustUseDir("test-010-listener-readd")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

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

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })

}

func Test011Wire(t *testing.T) {

	dir, origDir := mustUseDir("test-011-wire")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

		ctx, cancel := mustChromeCtx()
		defer cancel()

		log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

		must(chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:8846"+pathSuffix),
			WaitInnerTextTrimEq(".demo-comp1-c", "1"),
			WaitInnerTextTrimEq(".demo-comp2-c", "2"),
		))

	}

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })
}

func Test012Router(t *testing.T) {

	dir, origDir := mustUseDir("test-012-router")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

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

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })
}

func Test013Issue117(t *testing.T) {

	dir, origDir := mustUseDir("test-013-issue-117")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

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

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })
}

func Test014AttrIntf(t *testing.T) {

	dir, origDir := mustUseDir("test-014-attrintf")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

		assert := assert.New(t)

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

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })

}

func Test015AttrList(t *testing.T) {

	dir, origDir := mustUseDir("test-015-attribute-lister")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {
		assert := assert.New(t)

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

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })

}

func Test016SVG(t *testing.T) {

	dir, origDir := mustUseDir("test-016-svg")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {
		assert := assert.New(t)
		ctx, cancel := mustChromeCtx()
		defer cancel()

		log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

		must(chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:8846"+pathSuffix),
			chromedp.WaitVisible("#icon"),          // wait for the icon to show up
			chromedp.WaitVisible("#icon polyline"), // make sure that the svg element is complete
			chromedp.QueryAfter("#icon", func(ctx context.Context, node ...*cdp.Node) error {
				// checking if the element is recognized as SVG by chrome should be enough
				assert.True(node[0].IsSVG)
				return nil
			}),
		))

	}

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })

}

func Test017Nesting(t *testing.T) {

	dir, origDir := mustUseDir("test-017-nesting")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

		ctx, cancel := mustChromeCtx()
		defer cancel()

		log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

		must(chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:8846"+pathSuffix),
			chromedp.WaitVisible("#final1"), // make sure things showed up
			chromedp.WaitVisible("#final2"), // make sure things showed up

			chromedp.Click("#final1 .clicker"),            // click top one
			chromedp.WaitVisible("#final1 .clicked-true"), // should get clicked on #1
			chromedp.WaitNotPresent("#final1 .clicked-false"),
			chromedp.WaitNotPresent("#final2 .clicked-true"), // but not on #2
			chromedp.WaitVisible("#final2 .clicked-false"),

			// now check the reverse

			chromedp.Navigate("http://localhost:8846"+pathSuffix),
			chromedp.WaitVisible("#final1"), // make sure things showed up
			chromedp.WaitVisible("#final2"), // make sure things showed up

			chromedp.Click("#final2 .clicker"),            // click bottom one
			chromedp.WaitVisible("#final2 .clicked-true"), // should get clicked on #2
			chromedp.WaitNotPresent("#final2 .clicked-false"),
			chromedp.WaitNotPresent("#final1 .clicked-true"), // but not on #1
			chromedp.WaitVisible("#final1 .clicked-false"),
		))

	}

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })

}

func Test018CompEvents(t *testing.T) {

	dir, origDir := mustUseDir("test-018-comp-events")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

		assert := assert.New(t)

		ctx, cancel := mustChromeCtx()
		defer cancel()

		log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

		var showText string
		must(chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:8846"+pathSuffix),
			chromedp.WaitVisible("#top"),  // make sure things showed up
			chromedp.Click("#the_button"), // click the button inside the component

			chromedp.WaitVisible("#show_text"), // wait for it to dump event out
			chromedp.InnerHTML("#show_text", &showText),
		))
		//t.Logf("showText=%s", showText)
		assert.Contains(showText, "ClickEvent")

	}

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })

}

func Test019JSCreatePopulate(t *testing.T) {

	dir, origDir := mustUseDir("test-019-js-create-populate")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

		assert := assert.New(t)
		ctx, cancel := mustChromeCtx()
		defer cancel()

		log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

		var logText string
		must(chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:8846"+pathSuffix),
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

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	// FIXME: this fails with tinygo 0.22.0 in the UI with: syscall/js.finalizeRef not implemented; panic: JavaScript error: unreachable;
	// t.Run("tinygo", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, nil)) })

}

func Test020VGForm(t *testing.T) {

	dir, origDir := mustUseDir("test-020-vgform")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

		ctx, cancel := mustChromeCtx()
		defer cancel()

		log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

		must(chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:8846"+pathSuffix),
			chromedp.WaitVisible("#food_group_value"), // make sure things showed up
			chromedp.SendKeys(`#food_group`, "Butterfinger Group"),
			WaitInnerTextTrimEq("#food_group_value", "butterfinger_group"),
		))

	}

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })
}

func Test021Slots(t *testing.T) {

	dir, origDir := mustUseDir("test-021-slots")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

		ctx, cancel := mustChromeCtx()
		defer cancel()

		log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

		var tmplparentInnerHTML string
		must(chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:8846"+pathSuffix),
			chromedp.WaitVisible("#tmplparent"), // make sure things showed up
			chromedp.InnerHTML("#tmplparent", &tmplparentInnerHTML),
			WaitInnerTextTrimEq("#table2 #another_slot", "another slot"),
			WaitInnerTextTrimEq("#table2 #mapidx_slot", "mapidx slot"),
		))

		if tmplparentInnerHTML != "simple template test" {
			t.Errorf("tmplparent did not have expected innerHTML, instead got: %s", tmplparentInnerHTML)
		}

		rootvgengo, err := ioutil.ReadFile(filepath.Join(dir, "root_vgen.go"))
		must(err)
		if !regexp.MustCompile(`var mydt `).Match(rootvgengo) {
			t.Errorf("missing vg-var reference in root_vgen.go")
		}

	}

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })

}

func Test022EventListener(t *testing.T) {

	dir, origDir := mustUseDir("test-022-event-listener")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

		assert := assert.New(t)

		ctx, cancel := mustChromeCtx()
		defer cancel()
		// log.Printf("pathSuffix = %s", pathSuffix)

		var text string
		must(chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:8846"+pathSuffix),
			chromedp.WaitVisible("#top"),
			chromedp.Click("#switch"),
			chromedp.WaitVisible("#noclick"),
			chromedp.Click("#noclick"),
			chromedp.Click("#switch"),
			chromedp.WaitNotPresent("#noclick"),
			chromedp.Click("#click"),
			chromedp.Click("#switch"),
			chromedp.WaitVisible("#text"),
			chromedp.InnerHTML("#text", &text),
		))

		assert.Equal("click", text)

	}

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })
}

func Test023LifecycleCallbacks(t *testing.T) {

	dir, origDir := mustUseDir("test-023-lifecycle-callbacks")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

		assert := assert.New(t)
		_ = assert

		ctx, cancel := mustChromeCtx()
		defer cancel()
		log.Printf("URL: %s", "http://localhost:8846"+pathSuffix)

		var logTextContent string
		must(chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:8846"+pathSuffix),
			chromedp.WaitVisible("#top"),
			chromedp.Click("#togglec1"),
			chromedp.WaitNotPresent("#c1"),
			chromedp.Click("#togglec2"),
			chromedp.WaitNotPresent("#c2"),
			chromedp.Click("#refresh"),
			chromedp.Click("#togglec1"),
			chromedp.WaitVisible("#c1"),
			chromedp.Click("#togglec2"),
			chromedp.WaitVisible("#c2"),
			chromedp.Evaluate("window.logTextContent", &logTextContent),
		))

		assert.Equal(logTextContent, `got C1.Init(ctx)
got C1.Compute(ctx)
got C2.Init()
got C2.Compute()
got C1.Rendered(ctx)[first=true]
got C2.Rendered()
got C2.Compute()
got C1.Destroy(ctx)
got C2.Rendered()
got C2.Destroy()
got C1.Init(ctx)
got C1.Compute(ctx)
got C1.Rendered(ctx)[first=true]
got C1.Compute(ctx)
got C2.Init()
got C2.Compute()
got C1.Rendered(ctx)[first=false]
got C2.Rendered()
`)
		// log.Printf("logTextContext: %s", logTextContent)

	}

	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("tinygo/NoDocker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
	t.Run("tinygo/Docker", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })
}

func Test024EventBufferSize(t *testing.T) {

	dir, origDir := mustUseDir("test-024-event-buffer-size")
	defer os.Chdir(origDir)

	tf := func(t *testing.T, pathSuffix string) {

		ctx, cancel := mustChromeCtx()
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		var val string
		must(chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:8846"+pathSuffix),
			chromedp.WaitVisible("#top"),
			// trigger the change event
			chromedp.SendKeys("select", kb.ArrowDown+kb.ArrowDown),
			chromedp.Value(`select`, &val),
		))
		log.Println(val)
	}
	t.Run("go", func(t *testing.T) { tf(t, mustGenBuildAndLoad(dir)) })
	t.Run("go", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, true)) })
	t.Run("go", func(t *testing.T) { tf(t, mustTGGenBuildAndLoad(dir, false)) })
}
