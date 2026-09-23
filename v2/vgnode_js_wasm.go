//go:build js && wasm

package vugu

import (
	"fmt"
	"syscall/js"
)

// VGNode represents a node from our virtual DOM with the dynamic parts wired up into functions.
//
// For the static parts, an instance of VGNode corresponds directly to the DOM representation of
// an HTML node.  The pointers to other VGNode instances (Parent, FirstChild, etc.) are used to manage the tree.
// Type, Data, Namespace and Attr have the usual meanings for nodes.
//
// The Component field, if not-nil indicates that rendering should be delegated to the specified component,
// all other fields are ignored.
//
// Another special case is when Type is ElementNode and Data is an empty string (and Component is nil) then this node is a "template"
// (i.e. <vg-template>) and its children will be "flattened" into the DOM in position of this element
// and attributes, events, etc. ignored.
//
// Prop contains JavaScript property values to be assigned during render. InnerHTML provides alternate
// HTML content instead of children.  DOMEventHandlerSpecList specifies DOM handlers to register.
// And the JS...Handler fields are used to register callbacks to obtain information at JS render-time.
//
// TODO: This and its related parts should probably move into a sub-package (vgnode?) and
// the "VG" prefixes removed.
type VGNode struct {
	VGNodeCommonCore

	DOMEventHandlerSpecList []DOMEventHandlerSpec // describes invocations when DOM events happen

	// if not-nil, called when element is created (but before examining child nodes)
	JSCreateHandler JSValueHandler
	// if not-nil, called after children have been visited
	JSPopulateHandler JSValueHandler
}

// JSValueHandler does something with a js.Value
type JSValueHandler interface {
	JSValueHandle(js.Value)
}

// JSValueFunc implements JSValueHandler as a function.
type JSValueFunc func(js.Value)

// JSValueHandle implements the JSValueHandler interface.
func (f JSValueFunc) JSValueHandle(v js.Value) { f(v) }

func (vg *VGNode) Dummy() {
	t := vg.Type
	fmt.Printf("%q", t)
}
