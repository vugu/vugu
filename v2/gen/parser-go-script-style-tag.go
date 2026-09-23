package gen

import (
	"fmt"
	"strings"

	"github.com/vugu/html"
)

// visitScriptOrStyle calls visitJS, visitCSS or visitGo accordingly,
// will error if the node does not correspond to one of those
func (p *ParserGo) visitScriptOrStyle(n *html.Node) error {
	nodeName := strings.ToLower(n.Data)

	// script tag
	if nodeName == "script" {

		var mt string // mime type

		ty := attrWithKey(n, "type")
		if ty == nil {
			// return fmt.Errorf("script tag without type attribute is not valid")
			mt = ""
		} else {
			// tinygo support: just split on semi, don't need to import mime package
			// mt, _, _ = mime.ParseMediaType(ty.Val)
			mt = strings.Split(strings.TrimSpace(ty.Val), ";")[0]
		}

		// go code in script tags is now banned. Any code that was in a <script application/x-go> tag needs to be moved to the
		// corresponding <component>.go file
		if mt == "application/x-go" {
			return fmt.Errorf("%w", ErrGoInScriptTag)
		}

		// component js (type attr omitted okay - means it is JS)
		if mt == "text/javascript" || mt == "application/javascript" || mt == "" {
			err := p.visitJS(n)
			if err != nil {
				return err
			}
			return nil
		}

		return fmt.Errorf("found script tag with invalid mime type %q", mt)

	}

	// component css
	if nodeName == "style" || nodeName == "link" {
		err := p.visitCSS(n)
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("element %q is not a valid script or style - %#v", n.Data, n)
}

func (p *ParserGo) visitJS(n *html.Node) error {
	if n.Type != html.ElementNode {
		return fmt.Errorf("visitJS, not an element node %#v", n)
	}

	nodeName := strings.ToLower(n.Data)

	if nodeName != "script" {
		return fmt.Errorf("visitJS, tag %q not a script", nodeName)
	}

	// see if there's a script inside, or if this is a script include
	if n.FirstChild == nil {
		// script include - we pretty much just let this through, don't care what the attrs are
	} else {
		// if there is a script inside, we do not allow attributes other than "type", to avoid
		// people using features that might not be compatible with the funky stuff we have to do
		// in vugu to make all this work

		for _, a := range n.Attr {
			if a.Key != "type" {
				return fmt.Errorf("attribute %q not allowed on script tag that contains JS code", a.Key)
			}
			if a.Val != "text/javascript" && a.Val != "application/javascript" {
				return fmt.Errorf("script type %q invalid (must be text/javascript)", a.Val)
			}
		}

		// verify that all children are text nodes
		for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
			if childN.Type != html.TextNode {
				return fmt.Errorf("script tag contains non-text child: %#v", childN)
			}
		}

	}

	// allow control stuff, why not

	// vg-for
	if v, _ := vgForExpr(n); v.expr != "" {
		if err := p.emitForExpr(n); err != nil {
			return err
		}
		defer fmt.Fprintf(&p.buildBuf, "}\n")
	}

	// vg-if
	ife := vgIfExpr(n)
	if ife != "" {
		fmt.Fprintf(&p.buildBuf, "if %s {\n", ife)
		defer fmt.Fprintf(&p.buildBuf, "}\n")
	}

	// but then for the actual output, we append to vgout.JS, instead of parentNode
	fmt.Fprintf(&p.buildBuf, "vgn = &vugu.VGNode{VGNodeCommonCore: vugu.VGNodeCommonCore{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}}\n", n.Type, n.Data, staticVGAttr(n.Attr))

	// output any text children
	if n.FirstChild != nil {
		fmt.Fprintf(&p.buildBuf, "{\n")
		for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
			// NOTE: we already verified above that these are just text nodes
			fmt.Fprintf(&p.buildBuf, "vgn.AppendChild(&vugu.VGNode{VGNodeCommonCore: vugu.VGNodeCommonCore{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}})\n", childN.Type, childN.Data, staticVGAttr(childN.Attr))
		}
		fmt.Fprintf(&p.buildBuf, "}\n")
	}

	fmt.Fprintf(&p.buildBuf, "vgout.AppendJS(vgn)\n")

	// dynamic attrs
	p.writeDynamicAttributes(n)

	return nil
}

func (p *ParserGo) visitCSS(n *html.Node) error {
	if n.Type != html.ElementNode {
		return fmt.Errorf("visitCSS, not an element node %#v", n)
	}

	nodeName := strings.ToLower(n.Data)
	switch nodeName {

	case "link":

		// okay as long as nothing is inside this node

		if n.FirstChild != nil {
			return fmt.Errorf("link tag should not have children")
		}

		// and it needs to have an href (url)
		hrefAttr := attrWithKey(n, "href")
		if hrefAttr == nil {
			return fmt.Errorf("link tag must have href attribute but does not: %#v", n)
		}

	case "style":

		// style must have child (will verify it is text below)
		if n.FirstChild == nil {
			return fmt.Errorf("style must have contents but does not: %#v", n)
		}

		// okay as long as only text nodes inside
		for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
			if childN.Type != html.TextNode {
				return fmt.Errorf("style tag contains non-text child: %#v", childN)
			}
		}

	default:
		return fmt.Errorf("visitCSS, unexpected tag name %q", nodeName)
	}

	// allow control stuff, why not

	// vg-for
	if v, _ := vgForExpr(n); v.expr != "" {
		if err := p.emitForExpr(n); err != nil {
			return err
		}
		defer fmt.Fprintf(&p.buildBuf, "}\n")
	}

	// vg-if
	ife := vgIfExpr(n)
	if ife != "" {
		fmt.Fprintf(&p.buildBuf, "if %s {\n", ife)
		defer fmt.Fprintf(&p.buildBuf, "}\n")
	}

	// but then for the actual output, we append to vgout.CSS, instead of parentNode
	fmt.Fprintf(&p.buildBuf, "vgn = &vugu.VGNode{VGNodeCommonCore: vugu.VGNodeCommonCore{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}}\n", n.Type, n.Data, staticVGAttr(n.Attr))

	// output any text children
	if n.FirstChild != nil {
		fmt.Fprintf(&p.buildBuf, "{\n")
		for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
			// NOTE: we already verified above that these are just text nodes
			fmt.Fprintf(&p.buildBuf, "vgn.AppendChild(&vugu.VGNode{VGNodeCommonCore: vugu.VGNodeCommonCore{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}})\n", childN.Type, childN.Data, staticVGAttr(childN.Attr))
		}
		fmt.Fprintf(&p.buildBuf, "}\n")
	}

	fmt.Fprintf(&p.buildBuf, "vgout.AppendCSS(vgn)\n")

	// dynamic attrs
	p.writeDynamicAttributes(n)

	return nil
}
