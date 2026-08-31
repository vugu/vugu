package gen

import (
	"fmt"

	"github.com/vugu/html"
)

func (p *ParserGo) visitHead(state *parseGoState, n *html.Node) error {
	pOutputTag(state, n)
	// fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{VGNodeCommonCore: vugu.VGNodeCommonCore{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}}\n", n.Type, n.Data, staticVGAttr(n.Attr))
	// fmt.Fprintf(&state.buildBuf, "vgout.Out = append(vgout.Out, vgn) // root for output\n") // for first element we need to assign as Doc on BuildOut
	// state.outIsSet = true

	// dynamic attrs
	writeDynamicAttributes(state, n)

	fmt.Fprintf(&state.buildBuf, "{\n")
	fmt.Fprintf(&state.buildBuf, "vgparent := vgn; _ = vgparent\n") // vgparent set for this block to vgn

	for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

		if isScriptOrStyle(childN) {
			err := p.visitScriptOrStyle(state, childN)
			if err != nil {
				return err
			}
			continue
		}

		err := p.visitDefaultByType(state, childN)
		if err != nil {
			return err
		}

	}

	fmt.Fprintf(&state.buildBuf, "}\n")

	return nil
}
