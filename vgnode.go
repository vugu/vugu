package vugu

import (
	"fmt"
	"html"
	"reflect"
	"strconv"

	"github.com/vugu/vugu/js"
)

// VGNodeType is one of the valid node types (error, text, document, element, comment, doctype).
// Note that only text, element and comment are currently used.
type VGNodeType uint32

// Available VGNodeTypes.
const (
	ErrorNode    = VGNodeType(0 /*html.ErrorNode*/)
	TextNode     = VGNodeType(1 /*html.TextNode*/)
	DocumentNode = VGNodeType(2 /*html.DocumentNode*/)
	ElementNode  = VGNodeType(3 /*html.ElementNode*/)
	CommentNode  = VGNodeType(4 /*html.CommentNode*/)
	DoctypeNode  = VGNodeType(5 /*html.DoctypeNode*/)
)

// VGAtom is an integer corresponding to golang.org/x/net/html/atom.Atom.
// Note that this may be removed for simplicity and to remove the dependency
// on the package above.  Suggest you don't use it.
// type VGAtom uint32

// VGAttribute is the attribute on an HTML tag.
type VGAttribute struct {
	Namespace, Key, Val string
}

// VGProperty is a JS property to be set on a DOM element.
type VGProperty struct {
	Key     string
	JSONVal []byte // value as JSON expression
}

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
	Parent, FirstChild, LastChild, PrevSibling, NextSibling *VGNode

	Type VGNodeType
	// DataAtom  VGAtom // this needed to come out, we're not using it and without well-defined behavior it just becomes confusing and problematic
	Data      string
	Namespace string
	Attr      []VGAttribute

	// JS properties to e set during render
	Prop []VGProperty

	// Props Props // dynamic attributes, used as input for components or converted to attributes for regular HTML elements

	InnerHTML *string // indicates that children should be ignored and this raw HTML is the children of this tag; nil means not set, empty string means explicitly set to empty string

	DOMEventHandlerSpecList []DOMEventHandlerSpec // describes invocations when DOM events happen

	// indicates this node's output should be delegated to the specified component
	Component interface{}

	// if not-nil, called when element is created (but before examining child nodes)
	JSCreateHandler JSValueHandler
	// if not-nil, called after children have been visited
	JSPopulateHandler JSValueHandler
}

// IsComponent returns true if this is a component (Component != nil).
// Components have rendering delegated to them instead of processing this node.
func (n *VGNode) IsComponent() bool {
	return n.Component != nil
}

// IsTemplate returns true if this is a template (Type is ElementNode and Data is an empty string and not a Component).
// Templates have their children flattened into the output DOM instead of being processed directly.
func (n *VGNode) IsTemplate() bool {
	if n.Type == ElementNode && n.Data == "" && n.Component == nil {
		return true
	}
	return false
}

// InsertBefore inserts newChild as a child of n, immediately before oldChild
// in the sequence of n's children. oldChild may be nil, in which case newChild
// is appended to the end of n's children.
//
// It will panic if newChild already has a parent or siblings.
func (n *VGNode) InsertBefore(newChild, oldChild *VGNode) {
	if newChild.Parent != nil || newChild.PrevSibling != nil || newChild.NextSibling != nil {
		panic("html: InsertBefore called for an attached child Node")
	}
	var prev, next *VGNode
	if oldChild != nil {
		prev, next = oldChild.PrevSibling, oldChild
	} else {
		prev = n.LastChild
	}
	if prev != nil {
		prev.NextSibling = newChild
	} else {
		n.FirstChild = newChild
	}
	if next != nil {
		next.PrevSibling = newChild
	} else {
		n.LastChild = newChild
	}
	newChild.Parent = n
	newChild.PrevSibling = prev
	newChild.NextSibling = next
}

// AppendChild adds a node c as a child of n.
//
// It will panic if c already has a parent or siblings.
func (n *VGNode) AppendChild(c *VGNode) {
	if c.Parent != nil || c.PrevSibling != nil || c.NextSibling != nil {
		panic("html: AppendChild called for an attached child Node")
	}
	last := n.LastChild
	if last != nil {
		last.NextSibling = c
	} else {
		n.FirstChild = c
	}
	n.LastChild = c
	c.Parent = n
	c.PrevSibling = last
}

// RemoveChild removes a node c that is a child of n. Afterwards, c will have
// no parent and no siblings.
//
// It will panic if c's parent is not n.
func (n *VGNode) RemoveChild(c *VGNode) {
	if c.Parent != n {
		panic("html: RemoveChild called for a non-child Node")
	}
	if n.FirstChild == c {
		n.FirstChild = c.NextSibling
	}
	if c.NextSibling != nil {
		c.NextSibling.PrevSibling = c.PrevSibling
	}
	if n.LastChild == c {
		n.LastChild = c.PrevSibling
	}
	if c.PrevSibling != nil {
		c.PrevSibling.NextSibling = c.NextSibling
	}
	c.Parent = nil
	c.PrevSibling = nil
	c.NextSibling = nil
}

// type IfFunc func(data interface{}) (bool, error)

// Walk will walk the tree under a VGNode using the specified callback function.
// If the function provided returns a non-nil error then walking will be stopped
// and this error will be returned.  Only FirstChild and NextSibling are used
// while walking and so with well-formed documents should not loop. (But loops
// created manually by modifying FirstChild or NextSibling pointers could cause
// this function to recurse indefinitely.)  Note that f may modify nodes as it
// visits them with predictable results as long as it does not modify elements
// higher on the tree (up, toward the parent); it is safe to modify self and children.
func (n *VGNode) Walk(f func(*VGNode) error) error {
	if n == nil {
		return nil
	}
	err := f(n)
	if err != nil {
		return err
	}
	err = n.FirstChild.Walk(f)
	if err != nil {
		return err
	}
	err = n.NextSibling.Walk(f)
	if err != nil {
		return err
	}
	return nil
}

// AddAttrInterface sets an attribute based on the given interface. The followings types are supported
// - string - value is used as attr value as it is
// - int,float,... - the value is converted to string with strconv and used as attr value
// - bool - treat the attribute as a flag. If false, the attribute will be ignored, if true outputs the attribute without a value
// - fmt.Stringer - if the value implements fmt.Stringer, the returned string of StringVar() is used
// - ptr - If the ptr is nil, the attribute will be ignored. Else, the rules above apply
// any other type is handled via fmt.Sprintf()
func (n *VGNode) AddAttrInterface(key string, val interface{}) {
	// ignore nil attributes
	if val == nil {
		return
	}

	nattr := VGAttribute{
		Key: key,
	}

	switch v := val.(type) {
	case string:
		nattr.Val = v
	case int:
		nattr.Val = strconv.Itoa(v)
	case int8:
		nattr.Val = strconv.Itoa(int(v))
	case int16:
		nattr.Val = strconv.Itoa(int(v))
	case int32:
		nattr.Val = strconv.Itoa(int(v))
	case int64:
		nattr.Val = strconv.FormatInt(v, 10)
	case uint:
		nattr.Val = strconv.FormatUint(uint64(v), 10)
	case uint8:
		nattr.Val = strconv.FormatUint(uint64(v), 10)
	case uint16:
		nattr.Val = strconv.FormatUint(uint64(v), 10)
	case uint32:
		nattr.Val = strconv.FormatUint(uint64(v), 10)
	case uint64:
		nattr.Val = strconv.FormatUint(v, 10)
	case float32:
		nattr.Val = strconv.FormatFloat(float64(v), 'f', 6, 32)
	case float64:
		nattr.Val = strconv.FormatFloat(v, 'f', 6, 64)
	case bool:
		if !v {
			return
		}
		// FIXME: according to the HTML spec this is valid for flags, but not pretty
		// To change this, we have to adopt the changes inside the domrender and staticrender too.
		nattr.Val = key
	case fmt.Stringer:
		// we have to check that the given interface does not hide a nil ptr
		if p := reflect.ValueOf(val); p.Kind() == reflect.Ptr && p.IsNil() {
			return
		}
		nattr.Val = v.String()
	default:
		// check if this is a ptr
		rv := reflect.ValueOf(val)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return
			}
			// indirect the ptr and recurse
			if !rvIsZero(rv) {
				n.AddAttrInterface(key, reflect.Indirect(rv).Interface())
			}
			return
		}
		// fall back to fmt
		nattr.Val = fmt.Sprintf("%v", val)
	}

	n.Attr = append(n.Attr, nattr)
}

// AddAttrList takes a VGAttributeLister and sets the returned attributes to the node
func (n *VGNode) AddAttrList(lister VGAttributeLister) {
	n.Attr = append(n.Attr, lister.AttributeList()...)
}

// SetInnerHTML assigns the InnerHTML field with useful logic based on the type of input.
// Values of string type are escaped using html.EscapeString().  Values in the
// int or float type families or bool are converted to a string using the strconv package
// (no escaping is required).  Nil values set InnerHTML to nil.  Non-nil pointers
// are followed and the same rules applied.  Values implementing HTMLer will have their
// HTML() method called and the result put into InnerHTML without escaping.
// Values implementing fmt.Stringer have thier String() method called and the escaped result
// used as in string above.
//
// All other values have undefined behavior but are currently handled by setting InnerHTML
// to the result of: `html.EscapeString(fmt.Sprintf("%v", val))`
func (n *VGNode) SetInnerHTML(val interface{}) {

	var s string

	if val == nil {
		n.InnerHTML = nil
		return
	}

	switch v := val.(type) {
	case string:
		s = html.EscapeString(v)
	case int:
		s = strconv.Itoa(v)
	case int8:
		s = strconv.Itoa(int(v))
	case int16:
		s = strconv.Itoa(int(v))
	case int32:
		s = strconv.Itoa(int(v))
	case int64:
		s = strconv.FormatInt(v, 10)
	case uint:
		s = strconv.FormatUint(uint64(v), 10)
	case uint8:
		s = strconv.FormatUint(uint64(v), 10)
	case uint16:
		s = strconv.FormatUint(uint64(v), 10)
	case uint32:
		s = strconv.FormatUint(uint64(v), 10)
	case uint64:
		s = strconv.FormatUint(v, 10)
	case float32:
		s = strconv.FormatFloat(float64(v), 'f', 6, 32)
	case float64:
		s = strconv.FormatFloat(v, 'f', 6, 64)
	case bool:
		// I don't see any use in making false result in no output in this case.
		// I'm guessing if someone puts a bool in vg-content/vg-html
		// they are expecting it to print out the text "true" or "false".
		// It's different from AddAttrInterface but I think appropriate here.
		s = strconv.FormatBool(v)

	case HTMLer:
		s = v.HTML()

	case fmt.Stringer:
		// we have to check that the given interface does not hide a nil ptr
		if p := reflect.ValueOf(val); p.Kind() == reflect.Ptr && p.IsNil() {
			n.InnerHTML = nil
			return
		}
		s = html.EscapeString(v.String())
	default:
		// check if this is a ptr
		rv := reflect.ValueOf(val)

		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return
			}
			if !rvIsZero(rv) {
				indirected := reflect.Indirect(rv)
				if !indirected.IsNil() {
					n.SetInnerHTML(indirected.Interface())
				}
			}
			return
		}

		// fall back to fmt
		s = html.EscapeString(fmt.Sprintf("%v", val))
	}

	n.InnerHTML = &s
}

// JSValueHandler does something with a js.Value
type JSValueHandler interface {
	JSValueHandle(js.Value)
}

// JSValueFunc implements JSValueHandler as a function.
type JSValueFunc func(js.Value)

// JSValueHandle implements the JSValueHandler interface.
func (f JSValueFunc) JSValueHandle(v js.Value) { f(v) }
