// Code generated by vugu via vugugen DO NOT EDIT.
// Please regenerate instead of editing or add additional code in a separate file.

package main

import "fmt"
import "reflect"
import "github.com/vugu/vjson"
import "github.com/vugu/vugu"
import js "github.com/vugu/vugu/js"
import "log"

func (c *Nameinput) Build(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {

	vgout = &vugu.BuildOut{}

	var vgiterkey interface{}
	_ = vgiterkey
	var vgn *vugu.VGNode
	vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute{{Namespace: "", Key: "id", Val: "nameinput"}}}
	vgout.Out = append(vgout.Out, vgn)	// root for output
	{
		vgparent := vgn
		_ = vgparent
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "label", Attr: []vugu.VGAttribute{{Namespace: "", Key: "for", Val: "name"}}}
		vgparent.AppendChild(vgn)
		vgn.SetInnerHTML(vugu.HTML("Name"))
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(4), Data: " Note in this case 'c' is the Nameinput component itself. "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(4), Data: " We call Nameinput.Change when the value changes to retrieve the new value "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "input", Attr: []vugu.VGAttribute{{Namespace: "", Key: "id", Val: "name"}, vugu.VGAttribute{Namespace: "", Key: "type", Val: "text"}, vugu.VGAttribute{Namespace: "", Key: "value", Val: "enter your name"}}}
		vgparent.AppendChild(vgn)
		vgn.DOMEventHandlerSpecList = append(vgn.DOMEventHandlerSpecList, vugu.DOMEventHandlerSpec{
			EventType:	"change",
			Func:		func(event vugu.DOMEvent) { c.Change(event) },
			// TODO: implement capture, etc. mostly need to decide syntax
		})
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
