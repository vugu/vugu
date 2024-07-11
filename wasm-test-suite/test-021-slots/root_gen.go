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
	vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute(nil)}
	vgout.Out = append(vgout.Out, vgn)	// root for output
	{
		vgparent := vgn
		_ = vgparent
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "p", Attr: []vugu.VGAttribute(nil)}
		vgparent.AppendChild(vgn)
		vgn.SetInnerHTML(vugu.HTML("Some root stuff here"))
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute{{Namespace: "", Key: "id", Val: "tmplparent"}}}
		vgparent.AppendChild(vgn)
		{
			vgparent := vgn
			_ = vgparent
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3)}	// <vg-template>
			vgparent.AppendChild(vgn)
			{
				vgparent := vgn
				_ = vgparent
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "simple template test"}
				vgparent.AppendChild(vgn)
			}
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute{{Namespace: "", Key: "id", Val: "aftertmpl"}}}
		vgparent.AppendChild(vgn)
		vgn.SetInnerHTML(vugu.HTML("after the template"))
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n\n    "}
		vgparent.AppendChild(vgn)
		{
			vgcompKey := vugu.MakeCompKey(0x20F93558E2DE1A3B^vgin.CurrentPositionHash(), vgiterkey)
			// ask BuildEnv for prior instance of this specific component
			vgcomp, _ := vgin.BuildEnv.CachedComponent(vgcompKey).(*MyDataTable)
			if vgcomp == nil {
				// create new one if needed
				vgcomp = new(MyDataTable)
				vgin.BuildEnv.WireComponent(vgcomp)
			}
			vgin.BuildEnv.UseComponent(vgcompKey, vgcomp)	// ensure we can use this in the cache next time around
			var mydt = vgcomp				// vg-var
			_ = mydt
			vgcomp.AttrMap = make(map[string]interface{}, 8)
			vgcomp.AttrMap["id"] = "table1"
			vgcomp.DefaultSlot = vugu.NewBuilderFunc(func(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {
				vgn := &vugu.VGNode{Type: vugu.VGNodeType(3)}
				vgout = &vugu.BuildOut{}
				vgout.Out = append(vgout.Out, vgn)
				vgparent := vgn
				_ = vgparent

				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
				vgparent.AppendChild(vgn)
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute{{Namespace: "", Key: "id", Val: "default1"}}}
				vgparent.AppendChild(vgn)
				{
					vgparent := vgn
					_ = vgparent
					vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "some default slot content here"}
					vgparent.AppendChild(vgn)
				}
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
				vgparent.AppendChild(vgn)
				return
			})
			vgout.Components = append(vgout.Components, vgcomp)
			vgn = &vugu.VGNode{Component: vgcomp}
			vgparent.AppendChild(vgn)
		}
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n\n    "}
		vgparent.AppendChild(vgn)
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute(nil)}
		vgparent.AppendChild(vgn)
		vgn.SetInnerHTML(vugu.HTML("---"))
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n\n    "}
		vgparent.AppendChild(vgn)
		{
			vgcompKey := vugu.MakeCompKey(0xC9595912B0BE3ED0^vgin.CurrentPositionHash(), vgiterkey)
			// ask BuildEnv for prior instance of this specific component
			vgcomp, _ := vgin.BuildEnv.CachedComponent(vgcompKey).(*MyDataTable)
			if vgcomp == nil {
				// create new one if needed
				vgcomp = new(MyDataTable)
				vgin.BuildEnv.WireComponent(vgcomp)
			}
			vgin.BuildEnv.UseComponent(vgcompKey, vgcomp)	// ensure we can use this in the cache next time around
			var mydt2 = vgcomp				// vg-var
			_ = mydt2
			vgcomp.AttrMap = make(map[string]interface{}, 8)
			vgcomp.AttrMap["id"] = "table2"
			if vgcomp.SlotMap == nil {
				vgcomp.SlotMap = make(map[string]vugu.Builder)
			} else {
				for k := range vgcomp.SlotMap {
					delete(vgcomp.SlotMap, k)
				}
			}
			vgcomp.DefaultSlot = vugu.NewBuilderFunc(func(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {
				vgn := &vugu.VGNode{Type: vugu.VGNodeType(3)}
				vgout = &vugu.BuildOut{}
				vgout.Out = append(vgout.Out, vgn)
				vgparent := vgn
				_ = vgparent

				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n            "}
				vgparent.AppendChild(vgn)
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute{{Namespace: "", Key: "id", Val: "default2"}}}
				vgparent.AppendChild(vgn)
				{
					vgparent := vgn
					_ = vgparent
					vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "default slot"}
					vgparent.AppendChild(vgn)
				}
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
				vgparent.AppendChild(vgn)
				return
			})
			vgcomp.AnotherSlot = vugu.NewBuilderFunc(func(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {
				vgn := &vugu.VGNode{Type: vugu.VGNodeType(3)}
				vgout = &vugu.BuildOut{}
				vgout.Out = append(vgout.Out, vgn)
				vgparent := vgn
				_ = vgparent

				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n            "}
				vgparent.AppendChild(vgn)
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute(nil)}
				vgparent.AppendChild(vgn)
				{
					vgparent := vgn
					_ = vgparent
					vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "another slot"}
					vgparent.AppendChild(vgn)
				}
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
				vgparent.AppendChild(vgn)
				return
			})
			vgcomp.SlotMap["mapidx"] = vugu.NewBuilderFunc(func(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {
				vgn := &vugu.VGNode{Type: vugu.VGNodeType(3)}
				vgout = &vugu.BuildOut{}
				vgout.Out = append(vgout.Out, vgn)
				vgparent := vgn
				_ = vgparent

				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n            "}
				vgparent.AppendChild(vgn)
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "div", Attr: []vugu.VGAttribute(nil)}
				vgparent.AppendChild(vgn)
				{
					vgparent := vgn
					_ = vgparent
					vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "mapidx slot"}
					vgparent.AppendChild(vgn)
				}
				vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
				vgparent.AppendChild(vgn)
				return
			})
			vgout.Components = append(vgout.Components, vgcomp)
			vgn = &vugu.VGNode{Component: vgcomp}
			vgparent.AppendChild(vgn)
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
