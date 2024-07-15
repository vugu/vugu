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
		vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "ul", Attr: []vugu.VGAttribute(nil)}
		vgparent.AppendChild(vgn)
		{
			vgparent := vgn
			_ = vgparent
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
			vgparent.AppendChild(vgn)
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "li", Attr: []vugu.VGAttribute(nil)}
			vgparent.AppendChild(vgn)
			{
				vgparent := vgn
				_ = vgparent
				{
					vgcompKey := vugu.MakeCompKey(0xEAA2321F813543CB^vgin.CurrentPositionHash(), vgiterkey)
					// ask BuildEnv for prior instance of this specific component
					vgcomp, _ := vgin.BuildEnv.CachedComponent(vgcompKey).(*DemoComp1)
					if vgcomp == nil {
						// create new one if needed
						vgcomp = new(DemoComp1)
						vgin.BuildEnv.WireComponent(vgcomp)
					}
					vgin.BuildEnv.UseComponent(vgcompKey, vgcomp)	// ensure we can use this in the cache next time around
					vgout.Components = append(vgout.Components, vgcomp)
					vgn = &vugu.VGNode{Component: vgcomp}
					vgparent.AppendChild(vgn)
				}
			}
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n        "}
			vgparent.AppendChild(vgn)
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(3), Namespace: "", Data: "li", Attr: []vugu.VGAttribute(nil)}
			vgparent.AppendChild(vgn)
			{
				vgparent := vgn
				_ = vgparent
				{
					vgcompKey := vugu.MakeCompKey(0xB36A247F1D4AEB99^vgin.CurrentPositionHash(), vgiterkey)
					// ask BuildEnv for prior instance of this specific component
					vgcomp, _ := vgin.BuildEnv.CachedComponent(vgcompKey).(*DemoComp2)
					if vgcomp == nil {
						// create new one if needed
						vgcomp = new(DemoComp2)
						vgin.BuildEnv.WireComponent(vgcomp)
					}
					vgin.BuildEnv.UseComponent(vgcompKey, vgcomp)	// ensure we can use this in the cache next time around
					vgout.Components = append(vgout.Components, vgcomp)
					vgn = &vugu.VGNode{Component: vgcomp}
					vgparent.AppendChild(vgn)
				}
			}
			vgn = &vugu.VGNode{Type: vugu.VGNodeType(1), Data: "\n    "}
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