// Code generated by vugu via vugugen DO NOT EDIT.
// Please regenerate instead of editing or add additional code in a separate file.

package main

import "fmt"
import "reflect"
import "github.com/vugu/vjson"
import "github.com/vugu/vugu"
import js "syscall/js"
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
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "body", Attr: []vugu.VGAttribute(nil)}
		vgparent.AppendChild(vgn)
		{
			vgparent := vgn
			_ = vgparent
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute{{Namespace: "", Key: "id", Val: "content"}}}
			vgparent.AppendChild(vgn)
			{
				vgparent := vgn
				_ = vgparent
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n      "}
				vgparent.AppendChild(vgn)
				for i := 0; i < 5; i++ {
					var vgiterkey interface{} = i
					_ = vgiterkey
					i := i
					_ = i
					vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "span", Attr: []vugu.VGAttribute(nil)}
					vgparent.AppendChild(vgn)
					vgn.AddAttrInterface("id", fmt.Sprintf("id%d", i))
					vgn.SetInnerHTML(i)
					vgn.DOMEventHandlerSpecList = append(vgn.DOMEventHandlerSpecList, vugu.DOMEventHandlerSpec{
						EventType:	"click",
						Func:		func(event vugu.DOMEvent) { c.Clicked = fmt.Sprint(i) },
						// TODO: implement capture, etc. mostly need to decide syntax
					})
					{
						vgparent := vgn
						_ = vgparent
						vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n      "}
						vgparent.AppendChild(vgn)
					}
				}
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n      "}
				vgparent.AppendChild(vgn)
				if c.Clicked != "" {
					vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "p", Attr: []vugu.VGAttribute{{Namespace: "", Key: "id", Val: "clicked"}}}
					vgparent.AppendChild(vgn)
					vgn.SetInnerHTML(c.Clicked + " clicked!")
				}
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
				vgparent.AppendChild(vgn)
			}
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
