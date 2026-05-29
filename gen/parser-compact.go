package gen

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	// "github.com/vugu/vugu/internal/html"
	// "golang.org/x/net/html"
	"github.com/vugu/html"
)

// compactNodeTree operates on a Node tree in-place and find elements with static
// contents and converts them to corresponding vg-html expressions with static output.
// Since vg-html ends up with a call to set innerHTML on an element in the DOM,
// it is much faster for large blocks of HTML than individual syncing DOM nodes.
// Any modern browser's native HTML parser is always going to be a lot faster than
// we can achieve calling back and forth from wasm for each element.
func compactNodeTree(rootN *html.Node) error {

	// do not collapse html, body or head, and nothing inside head

	var visit func(n *html.Node) (canCompact bool, err error)
	visit = func(n *html.Node) (canCompact bool, err error) {

		// certain tags we just refuse to examine at all
		if n.Type == html.ElementNode && (n.Data == "head" ||
			n.Data == "script" ||
			n.Data == "style" ||
			strings.HasPrefix(n.Data, "vg-")) {
			return false, nil
		}

		// text nodes are always compactable (at least in the current implementation)
		if n.FirstChild == nil && n.Type == html.TextNode {
			return true, nil
		}

		var compactableNodes []*html.Node
		allCompactable := true
		// iterate over the immediate children of n
		for n2 := n.FirstChild; n2 != nil; n2 = n2.NextSibling {
			cc, err := visit(n2)
			if err != nil {
				return false, err
			}
			allCompactable = allCompactable && cc // keep track of if they are all compactable
			if cc {
				compactableNodes = append(compactableNodes, n2) // keep track of individual nodes that are compactable
			}
		}

		// if we're in the top level HTML tag, that's it, we visited already above and we're done
		if n.Type == html.ElementNode && n.Data == "html" {
			return false, nil
		}

		// if not everything is compactable or it's the body node, then go through and compact the ones that can be
		if !allCompactable || (n.Type == html.ElementNode && n.Data == "body") {

			for _, cn := range compactableNodes {

				// NOTE: isStaticEl(cn) has already been run, since canCompact returned true above to put it in this list

				if cn.Type != html.ElementNode { // only work on elements
					continue
				}

				var htmlBuf bytes.Buffer
				// walk each immediate child of cn
				for cnChild := cn.FirstChild; cnChild != nil; cnChild = cnChild.NextSibling {
					// render directly into htmlBuf
					err := html.Render(&htmlBuf, cnChild)
					if err != nil {
						return false, err
					}
				}

				// add a vg-html with the static Go string expression of the contents casted to a vugu.HTML
				cn.Attr = append(cn.Attr, html.Attribute{Key: "vg-html", Val: "vugu.HTML(" + htmlGoQuoteString(htmlBuf.String()) + ")"})
				// cn.Attr = append(cn.Attr, html.Attribute{Key: "vg-html", Val: htmlGoQuoteString(htmlBuf.String())})

				// remove children, since vg-html supplants them
				cn.FirstChild = nil
				cn.LastChild = nil

			}

			return false, nil
		}

		// if all of the children are compactable, we need to check if this is an element that contains no dynamic attributes
		if allCompactable {
			return isStaticEl(n), nil
		}

		// default is not compactable
		return false, nil
	}
	_, err := visit(rootN)

	return err
}

func isStaticEl(n *html.Node) bool {

	if n.Type != html.ElementNode { // must be element
		return false
	}

	// component elements cannot be compacted
	if strings.Contains(n.Data, ":") {
		return false
	}
	if n.Data == "vg-comp" {
		return false
	}

	for _, attr := range n.Attr {
		if strings.HasPrefix(attr.Key, "vg-") { // vg- prefix means dynamic stuff
			return false
		}
		if len(attr.Key) == 0 { // avoid panic in this strange case
			continue
		}
		if !unicode.IsLetter(rune(attr.Key[0])) { // anything except a letter as an attr we assume to be dynamic
			return false
		}
	}

	// if it passes above, should be fine to compact
	return true

}

// htmlGoQuoteString is similar to printf'ing with %q but converts common things that require html escaping to
// backslashes instead for improved clarity
func htmlGoQuoteString(s string) string {

	var buf bytes.Buffer

	for _, c := range fmt.Sprintf("%q", s) {
		switch c {
		case '<', '>', '&':
			var qc string
			qc = fmt.Sprintf("\\x%X", uint8(c))
			buf.WriteString(qc)
		default:
			buf.WriteRune(c)
		}

	}

	return buf.String()
}
