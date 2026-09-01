package gen

import (
	"fmt"

	"github.com/vugu/html"
)

func (p *ParserGo) visitBody(n *html.Node) error {
	p.pOutputTag(n)
	// fmt.Fprintf(&p.buildBuf, "vgn = &vugu.VGNode{VGNodeCommonCore: vugu.VGNodeCommonCore{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}}\n", n.Type, n.Data, staticVGAttr(n.Attr))
	// fmt.Fprintf(&p.buildBuf, "vgout.Out = append(vgout.Out, vgn) // root for output\n") // for first element we need to assign as Doc on BuildOut
	// p.outIsSet = true

	// dynamic attrs
	p.writeDynamicAttributes(n)

	fmt.Fprintf(&p.buildBuf, "{\n")
	fmt.Fprintf(&p.buildBuf, "vgparent := vgn; _ = vgparent\n") // vgparent set for this block to vgn

	foundMountEl := false

	for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

		// ignore whitespace and comments directly in body
		if childN.Type != html.ElementNode {
			continue
		}

		if isScriptOrStyle(childN) {
			err := p.visitScriptOrStyle(childN)
			if err != nil {
				return err
			}
			continue
		}

		if foundMountEl {
			return fmt.Errorf("element %q found after we already have a mount element, you might have to wrap all your body content into a div", childN.Data)
		}
		foundMountEl = true

		err := p.visitDefaultByType(childN)
		if err != nil {
			return err
		}

	}

	fmt.Fprintf(&p.buildBuf, "}\n")

	return nil
}
