package gen

import (
	"fmt"

	"github.com/vugu/html"
)

func (p *ParserGo) visitHead(n *html.Node) error {
	p.pOutputTag(n)
	// fmt.Fprintf(&p.buildBuf, "vgn = &vugu.VGNode{VGNodeCommonCore: vugu.VGNodeCommonCore{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}}\n", n.Type, n.Data, staticVGAttr(n.Attr))
	// fmt.Fprintf(&p.buildBuf, "vgout.Out = append(vgout.Out, vgn) // root for output\n") // for first element we need to assign as Doc on BuildOut
	// p.outIsSet = true

	// dynamic attrs
	p.writeDynamicAttributes(n)

	fmt.Fprintf(&p.buildBuf, "{\n")
	fmt.Fprintf(&p.buildBuf, "vgparent := vgn; _ = vgparent\n") // vgparent set for this block to vgn

	for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

		if isScriptOrStyle(childN) {
			err := p.visitScriptOrStyle(childN)
			if err != nil {
				return err
			}
			continue
		}

		err := p.visitDefaultByType(childN)
		if err != nil {
			return err
		}

	}

	fmt.Fprintf(&p.buildBuf, "}\n")

	return nil
}
