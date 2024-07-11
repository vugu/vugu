// Code generated by vugu via vugugen DO NOT EDIT.
// Please regenerate instead of editing or add additional code in a separate file.

package main

import "fmt"
import "reflect"
import "github.com/vugu/vjson"
import "github.com/vugu/vugu"
import js "github.com/vugu/vugu/js"
import "log"

func (c *Root) Build(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {

	vgout = &vugu.BuildOut{}

	var vgiterkey interface{}
	_ = vgiterkey
	var vgn *vugu.VGNode
	vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute{{Namespace: "", Key: "class", Val: "demo-comp"}}}
	vgout.Out = append(vgout.Out, vgn)	// root for output
	{
		vgparent := vgn
		_ = vgparent
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		if c.isLoading {
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute(nil)}
			vgparent.AppendChild(vgn)
			{
				vgparent := vgn
				_ = vgparent
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "Loading..."}
				vgparent.AppendChild(vgn)
			}
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		if len(c.bpi.BPI) > 0 {
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute(nil)}
			vgparent.AppendChild(vgn)
			{
				vgparent := vgn
				_ = vgparent
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
				vgparent.AppendChild(vgn)
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute(nil)}
				vgparent.AppendChild(vgn)
				{
					vgparent := vgn
					_ = vgparent
					vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "Updated: "}
					vgparent.AppendChild(vgn)
					vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "span", Attr: []vugu.VGAttribute(nil)}
					vgparent.AppendChild(vgn)
					vgn.SetInnerHTML(c.bpi.Time.Updated)
				}
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
				vgparent.AppendChild(vgn)
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "ul", Attr: []vugu.VGAttribute(nil)}
				vgparent.AppendChild(vgn)
				{
					vgparent := vgn
					_ = vgparent
					vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n            "}
					vgparent.AppendChild(vgn)
					for key, value := range c.bpi.BPI {
						var vgiterkey interface{} = key
						_ = vgiterkey
						key := key
						_ = key
						value := value
						_ = value
						vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "li", Attr: []vugu.VGAttribute(nil)}
						vgparent.AppendChild(vgn)
						{
							vgparent := vgn
							_ = vgparent
							vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n                "}
							vgparent.AppendChild(vgn)
							vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "span", Attr: []vugu.VGAttribute(nil)}
							vgparent.AppendChild(vgn)
							vgn.SetInnerHTML(key)
							vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: " "}
							vgparent.AppendChild(vgn)
							vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "span", Attr: []vugu.VGAttribute(nil)}
							vgparent.AppendChild(vgn)
							vgn.SetInnerHTML(fmt.Sprint(value.Symbol, value.RateFloat))
							vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n            "}
							vgparent.AppendChild(vgn)
						}
					}
					vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
					vgparent.AppendChild(vgn)
				}
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
				vgparent.AppendChild(vgn)
			}
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "button", Attr: []vugu.VGAttribute(nil)}
		vgparent.AppendChild(vgn)
		vgn.DOMEventHandlerSpecList = append(vgn.DOMEventHandlerSpecList, vugu.DOMEventHandlerSpec{
			EventType:	"click",
			Func:		func(event vugu.DOMEvent) { c.HandleClick(event) },
			// TODO: implement capture, etc. mostly need to decide syntax
		})
		{
			vgparent := vgn
			_ = vgparent
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "Fetch Bitcoin Price Index"}
			vgparent.AppendChild(vgn)
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n"}
		vgparent.AppendChild(vgn)
	}
	return vgout
}

// 'fix' unused imports
var _ fmt.Stringer
var _ reflect.Type
var _ vjson.RawMessage
var _ js.Value
var _ log.Logger
