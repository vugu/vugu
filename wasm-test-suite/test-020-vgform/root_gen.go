// Code generated by vugu via vugugen DO NOT EDIT.
// Please regenerate instead of editing or add additional code in a separate file.

package main

import "fmt"
import "reflect"
import "github.com/vugu/vjson"
import "github.com/vugu/vugu"
import js "github.com/vugu/vugu/js"
import "log"

import "github.com/vugu/vugu/vgform"

func (c *Root) Build(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {

	vgout = &vugu.BuildOut{}

	var vgiterkey interface{}
	_ = vgiterkey
	var vgn *vugu.VGNode
	vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute{{Namespace: "", Key: "id", Val: "top"}}}
	vgout.Out = append(vgout.Out, vgn)	// root for output
	{
		vgparent := vgn
		_ = vgparent
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "form", Attr: []vugu.VGAttribute(nil)}
		vgparent.AppendChild(vgn)
		{
			vgparent := vgn
			_ = vgparent
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
			vgparent.AppendChild(vgn)
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute{{Namespace: "", Key: "class", Val: "form-group"}}}
			vgparent.AppendChild(vgn)
			{
				vgparent := vgn
				_ = vgparent
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n            "}
				vgparent.AppendChild(vgn)
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "label", Attr: []vugu.VGAttribute{{Namespace: "", Key: "for", Val: "food_group"}}}
				vgparent.AppendChild(vgn)
				vgn.SetInnerHTML(vugu.HTML("Select a Food Group"))
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n            "}
				vgparent.AppendChild(vgn)
				{
					vgcompKey := vugu.MakeCompKey(0x5B2857874B96E67C^vgin.CurrentPositionHash(), vgiterkey)
					// ask BuildEnv for prior instance of this specific component
					vgcomp, _ := vgin.BuildEnv.CachedComponent(vgcompKey).(*vgform.Select)
					if vgcomp == nil {
						// create new one if needed
						vgcomp = new(vgform.Select)
						vgin.BuildEnv.WireComponent(vgcomp)
					}
					vgin.BuildEnv.UseComponent(vgcompKey, vgcomp)	// ensure we can use this in the cache next time around
					vgcomp.Options = c.SetSliceOptions().Title()
					vgcomp.Value = c.SetStringPtrDefault(&c.FoodGroup, "jungle_group")
					vgcomp.AttrMap = make(map[string]interface{}, 8)
					vgcomp.AttrMap["id"] = "food_group"
					vgcomp.AttrMap["class"] = "form-control"
					vgout.Components = append(vgout.Components, vgcomp)
					vgn = &vugu.VGNode{Component: vgcomp}
					vgparent.AppendChild(vgn)
				}
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
				vgparent.AppendChild(vgn)
			}
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
			vgparent.AppendChild(vgn)
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute{{Namespace: "", Key: "class", Val: "form-group"}}}
			vgparent.AppendChild(vgn)
			{
				vgparent := vgn
				_ = vgparent
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n            "}
				vgparent.AppendChild(vgn)
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "label", Attr: []vugu.VGAttribute{{Namespace: "", Key: "for", Val: "textarea1"}}}
				vgparent.AppendChild(vgn)
				vgn.SetInnerHTML(vugu.HTML("Enter a bunch of text"))
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n            "}
				vgparent.AppendChild(vgn)
				{
					vgcompKey := vugu.MakeCompKey(0xD0F9E94D59F698F5^vgin.CurrentPositionHash(), vgiterkey)
					// ask BuildEnv for prior instance of this specific component
					vgcomp, _ := vgin.BuildEnv.CachedComponent(vgcompKey).(*vgform.Textarea)
					if vgcomp == nil {
						// create new one if needed
						vgcomp = new(vgform.Textarea)
						vgin.BuildEnv.WireComponent(vgcomp)
					}
					vgin.BuildEnv.UseComponent(vgcompKey, vgcomp)	// ensure we can use this in the cache next time around
					vgcomp.Value = c.SetStringPtrDefault(&c.Textarea1Value, "testing")
					vgcomp.AttrMap = make(map[string]interface{}, 8)
					vgcomp.AttrMap["id"] = "textarea1"
					vgcomp.AttrMap["class"] = "form-control"
					vgcomp.AttrMap["rows"] = "10"
					vgout.Components = append(vgout.Components, vgcomp)
					vgn = &vugu.VGNode{Component: vgcomp}
					vgparent.AppendChild(vgn)
				}
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
				vgparent.AppendChild(vgn)
			}
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
			vgparent.AppendChild(vgn)
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute{{Namespace: "", Key: "class", Val: "form-group"}}}
			vgparent.AppendChild(vgn)
			{
				vgparent := vgn
				_ = vgparent
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n            "}
				vgparent.AppendChild(vgn)
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "label", Attr: []vugu.VGAttribute{{Namespace: "", Key: "for", Val: "inputtext1"}}}
				vgparent.AppendChild(vgn)
				vgn.SetInnerHTML(vugu.HTML("Enter a line of text"))
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n            "}
				vgparent.AppendChild(vgn)
				{
					vgcompKey := vugu.MakeCompKey(0x2EE7D65B9CDEAFAF^vgin.CurrentPositionHash(), vgiterkey)
					// ask BuildEnv for prior instance of this specific component
					vgcomp, _ := vgin.BuildEnv.CachedComponent(vgcompKey).(*vgform.Input)
					if vgcomp == nil {
						// create new one if needed
						vgcomp = new(vgform.Input)
						vgin.BuildEnv.WireComponent(vgcomp)
					}
					vgin.BuildEnv.UseComponent(vgcompKey, vgcomp)	// ensure we can use this in the cache next time around
					vgcomp.Value = c.SetStringPtrDefault(&c.Inputtext1Value, "joe@example.com")
					vgcomp.AttrMap = make(map[string]interface{}, 8)
					vgcomp.AttrMap["type"] = "email"
					vgcomp.AttrMap["id"] = "inputtext1"
					vgcomp.AttrMap["class"] = "form-control"
					vgout.Components = append(vgout.Components, vgcomp)
					vgn = &vugu.VGNode{Component: vgcomp}
					vgparent.AppendChild(vgn)
				}
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
				vgparent.AppendChild(vgn)
			}
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
			vgparent.AppendChild(vgn)
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute(nil)}
		vgparent.AppendChild(vgn)
		{
			vgparent := vgn
			_ = vgparent
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "Your select: "}
			vgparent.AppendChild(vgn)
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "span", Attr: []vugu.VGAttribute{{Namespace: "", Key: "id", Val: "food_group_value"}}}
			vgparent.AppendChild(vgn)
			vgn.SetInnerHTML(c.FoodGroup)
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute(nil)}
		vgparent.AppendChild(vgn)
		{
			vgparent := vgn
			_ = vgparent
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "Your textarea: "}
			vgparent.AppendChild(vgn)
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "pre", Attr: []vugu.VGAttribute{{Namespace: "", Key: "id", Val: "textarea1_value"}}}
			vgparent.AppendChild(vgn)
			vgn.SetInnerHTML(c.Textarea1Value)
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute(nil)}
		vgparent.AppendChild(vgn)
		{
			vgparent := vgn
			_ = vgparent
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "Your inputtext: "}
			vgparent.AppendChild(vgn)
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "pre", Attr: []vugu.VGAttribute{{Namespace: "", Key: "id", Val: "inputtext1_value"}}}
			vgparent.AppendChild(vgn)
			vgn.SetInnerHTML(c.Inputtext1Value)
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n\n"}
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
