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
	vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "html", Attr: []vugu.VGAttribute(nil)}
	vgout.Out = append(vgout.Out, vgn)	// root for output
	{
		vgparent := vgn
		_ = vgparent
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "head", Attr: []vugu.VGAttribute(nil)}
		vgparent.AppendChild(vgn)
		{
			vgparent := vgn
			_ = vgparent
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
			vgparent.AppendChild(vgn)
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "title", Attr: []vugu.VGAttribute(nil)}
			vgparent.AppendChild(vgn)
			{
				vgparent := vgn
				_ = vgparent
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "Test page"}
				vgparent.AppendChild(vgn)
			}
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
			vgparent.AppendChild(vgn)
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Data: "link", Attr: []vugu.VGAttribute{{Namespace: "", Key: "rel", Val: "stylesheet"}, vugu.VGAttribute{Namespace: "", Key: "href", Val: "https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css"}, vugu.VGAttribute{Namespace: "", Key: "integrity", Val: "sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T"}, vugu.VGAttribute{Namespace: "", Key: "crossorigin", Val: "anonymous"}}}
			vgout.AppendCSS(vgn)
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
			vgparent.AppendChild(vgn)
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "body", Attr: []vugu.VGAttribute(nil)}
		vgparent.AppendChild(vgn)
		{
			vgparent := vgn
			_ = vgparent
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute{{Namespace: "", Key: "class", Val: "test-div"}, vugu.VGAttribute{Namespace: "", Key: "id", Val: "testdiv"}}}
			vgparent.AppendChild(vgn)
			{
				vgparent := vgn
				_ = vgparent
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n\n            "}
				vgparent.AppendChild(vgn)
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "ul", Attr: []vugu.VGAttribute(nil)}
				vgparent.AppendChild(vgn)
				{
					vgparent := vgn
					_ = vgparent
					vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n                "}
					vgparent.AppendChild(vgn)
					for i := 0; i < c.ItemCount; i++ {
						{
							vgcompKey := vugu.MakeCompKey(0x4A265D7939989913^vgin.CurrentPositionHash(), i)
							// ask BuildEnv for prior instance of this specific component
							vgcomp, _ := vgin.BuildEnv.CachedComponent(vgcompKey).(*DemoLine)
							if vgcomp == nil {
								// create new one if needed
								vgcomp = new(DemoLine)
								vgin.BuildEnv.WireComponent(vgcomp)
							}
							vgin.BuildEnv.UseComponent(vgcompKey, vgcomp)	// ensure we can use this in the cache next time around
							vgcomp.Num = i
							vgout.Components = append(vgout.Components, vgcomp)
							vgn = &vugu.VGNode{Component: vgcomp}
							vgparent.AppendChild(vgn)
						}
					}
					vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n            "}
					vgparent.AppendChild(vgn)
				}
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n\n            "}
				vgparent.AppendChild(vgn)
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "button", Attr: []vugu.VGAttribute{{Namespace: "", Key: "id", Val: "addbtn"}}}
				vgparent.AppendChild(vgn)
				vgn.DOMEventHandlerSpecList = append(vgn.DOMEventHandlerSpecList, vugu.DOMEventHandlerSpec{
					EventType:	"click",
					Func:		func(event vugu.DOMEvent) { c.OnAdd() },
					// TODO: implement capture, etc. mostly need to decide syntax
				})
				{
					vgparent := vgn
					_ = vgparent
					vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "Add"}
					vgparent.AppendChild(vgn)
				}
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n\n        "}
				vgparent.AppendChild(vgn)
			}
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Data: "style", Attr: []vugu.VGAttribute(nil)}
			{
				vgn.AppendChild(&vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n#test_div_id {\n    background: #ddd;\n}\n", Attr: []vugu.VGAttribute(nil)})
			}
			vgout.AppendCSS(vgn)
		}
	}
	return vgout
}

// 'fix' unused imports
var _ fmt.Stringer
var _ reflect.Type
var _ vjson.RawMessage
var _ js.Value
var _ log.Logger
