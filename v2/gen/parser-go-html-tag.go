package gen

import (
	"fmt"
	"strings"

	"github.com/vugu/html"
)

func (p *ParserGo) visitHTML(n *html.Node) error {
	p.pOutputTag(n)
	// fmt.Fprintf(&p.buildBuf, "vgn = &vugu.VGNode{VGNodeCommonCore: vugu.VGNodeCommonCore{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}}\n", n.Type, n.Data, staticVGAttr(n.Attr))
	// fmt.Fprintf(&p.buildBuf, "vgout.Out = append(vgout.Out, vgn) // root for output\n") // for first element we need to assign as Doc on BuildOut
	// p.outIsSet = true

	// dynamic attrs
	p.writeDynamicAttributes(n)

	fmt.Fprintf(&p.buildBuf, "{\n")
	fmt.Fprintf(&p.buildBuf, "vgparent := vgn; _ = vgparent\n") // vgparent set for this block to vgn

	for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

		if childN.Type != html.ElementNode {
			continue
		}

		var err error
		if strings.ToLower(childN.Data) == "head" {
			err = p.visitHead(childN)
		} else if strings.ToLower(childN.Data) == "body" {
			err = p.visitBody(childN)
		} else {
			return fmt.Errorf("unknown tag inside html %q", childN.Data)
		}

		if err != nil {
			return err
		}

	}

	fmt.Fprintf(&p.buildBuf, "}\n")

	return nil
}
