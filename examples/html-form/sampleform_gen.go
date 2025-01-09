// Code generated by vugu via vugugen DO NOT EDIT.
// Please regenerate instead of editing or add additional code in a separate file.

package main

import "fmt"
import "reflect"
import "github.com/vugu/vjson"
import "github.com/vugu/vugu"
import js "github.com/vugu/vugu/js"
import "log"

func (c *Sampleform) Build(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {

	vgout = &vugu.BuildOut{}

	var vgiterkey interface{}
	_ = vgiterkey
	var vgn *vugu.VGNode
	vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "form", Attr: []vugu.VGAttribute(nil)}
	vgout.Out = append(vgout.Out, vgn)	// root for output
	vgn.DOMEventHandlerSpecList = append(vgn.DOMEventHandlerSpecList, vugu.DOMEventHandlerSpec{
		EventType:	"submit",
		Func:		func(event vugu.DOMEvent) { c.Submit(event) },
		// TODO: implement capture, etc. mostly need to decide syntax
	})
	{
		vgparent := vgn
		_ = vgparent
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(4), Data: " Add an text input box, we do this via a sub component "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(4), Data: " Note the inversion onf control in each of the sub components "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(4), Data: " in this case 'c' is the Sampleform component and NOT the Parent component "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		{
			vgcompKey := vugu.MakeCompKey(0xEF6EC669734DD6FF^vgin.CurrentPositionHash(), vgiterkey)
			// ask BuildEnv for prior instance of this specific component
			vgcomp, _ := vgin.BuildEnv.CachedComponent(vgcompKey).(*Nameinput)
			if vgcomp == nil {
				// create new one if needed
				vgcomp = new(Nameinput)
				vgin.BuildEnv.WireComponent(vgcomp)
			}
			vgin.BuildEnv.UseComponent(vgcompKey, vgcomp)	// ensure we can use this in the cache next time around
			vgcomp.Form = c
			vgout.Components = append(vgout.Components, vgcomp)
			vgn = &vugu.VGNode{Component: vgcomp}
			vgparent.AppendChild(vgn)
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(4), Data: " Add an email input box "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		{
			vgcompKey := vugu.MakeCompKey(0x5C76A6F1331E84A6^vgin.CurrentPositionHash(), vgiterkey)
			// ask BuildEnv for prior instance of this specific component
			vgcomp, _ := vgin.BuildEnv.CachedComponent(vgcompKey).(*Emailinput)
			if vgcomp == nil {
				// create new one if needed
				vgcomp = new(Emailinput)
				vgin.BuildEnv.WireComponent(vgcomp)
			}
			vgin.BuildEnv.UseComponent(vgcompKey, vgcomp)	// ensure we can use this in the cache next time around
			vgcomp.Form = c
			vgout.Components = append(vgout.Components, vgcomp)
			vgn = &vugu.VGNode{Component: vgcomp}
			vgparent.AppendChild(vgn)
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(4), Data: " Add some radio buttons "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		{
			vgcompKey := vugu.MakeCompKey(0x2F83E693BC19BD7^vgin.CurrentPositionHash(), vgiterkey)
			// ask BuildEnv for prior instance of this specific component
			vgcomp, _ := vgin.BuildEnv.CachedComponent(vgcompKey).(*Radiobuttons)
			if vgcomp == nil {
				// create new one if needed
				vgcomp = new(Radiobuttons)
				vgin.BuildEnv.WireComponent(vgcomp)
			}
			vgin.BuildEnv.UseComponent(vgcompKey, vgcomp)	// ensure we can use this in the cache next time around
			vgcomp.Form = c
			vgout.Components = append(vgout.Components, vgcomp)
			vgn = &vugu.VGNode{Component: vgcomp}
			vgparent.AppendChild(vgn)
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(4), Data: " Add a select from a static list "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		{
			vgcompKey := vugu.MakeCompKey(0x5DECC76AC9929BD1^vgin.CurrentPositionHash(), vgiterkey)
			// ask BuildEnv for prior instance of this specific component
			vgcomp, _ := vgin.BuildEnv.CachedComponent(vgcompKey).(*Staticselect)
			if vgcomp == nil {
				// create new one if needed
				vgcomp = new(Staticselect)
				vgin.BuildEnv.WireComponent(vgcomp)
			}
			vgin.BuildEnv.UseComponent(vgcompKey, vgcomp)	// ensure we can use this in the cache next time around
			vgcomp.Form = c
			vgout.Components = append(vgout.Components, vgcomp)
			vgn = &vugu.VGNode{Component: vgcomp}
			vgparent.AppendChild(vgn)
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(4), Data: " Add a select from a dynamically generated list "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		{
			vgcompKey := vugu.MakeCompKey(0x59ECC269DC5DCB9D^vgin.CurrentPositionHash(), vgiterkey)
			// ask BuildEnv for prior instance of this specific component
			vgcomp, _ := vgin.BuildEnv.CachedComponent(vgcompKey).(*Dynamicselect)
			if vgcomp == nil {
				// create new one if needed
				vgcomp = new(Dynamicselect)
				vgin.BuildEnv.WireComponent(vgcomp)
			}
			vgin.BuildEnv.UseComponent(vgcompKey, vgcomp)	// ensure we can use this in the cache next time around
			vgcomp.Form = c
			vgout.Components = append(vgout.Components, vgcomp)
			vgn = &vugu.VGNode{Component: vgcomp}
			vgparent.AppendChild(vgn)
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(4), Data: " Add the \"Submit\" button to the form "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "input", Attr: []vugu.VGAttribute{{Namespace: "", Key: "type", Val: "submit"}, vugu.VGAttribute{Namespace: "", Key: "value", Val: "Submit Form"}}}
		vgparent.AppendChild(vgn)
		vgn.SetInnerHTML(vugu.HTML(""))
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
