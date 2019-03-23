package vugu

import (
	"bytes"

	// FIXME: see if we can clean it up programs which only use VGNode at run time (i.e in wasm)
	// will not end up bundling this library for no reason - means the init() below needs to be cleaned up
	// and probably have cruftBody be a function using once.Do - only needed if something calls it.
	"golang.org/x/net/html"
)

// Things to sort out:
// * vg-if, vg-range, etc (JUST COPIED FROM ATTRS, EVALUATED LATER)
// * :whatever syntax (ALSO JUST COPIED FROM ATTRS, EVALUATED LATER)
// * vg-model (IGNORE FOR NOW, WILL BE COMPOSED FROM OTHER FEATURES)
// * @click and other events (JUST A METHOD NAME TO INVOKE ON THE COMPONENT TYPE)
// * components and their properties (TAG NAME IS JUST CHECKED LATER, AND IF MATCH CAUSES THE EVAL STEPS ABOVE TO USE A DIFFERENT PATH, ATTRS GET REFS INSTEAD OF TEXT RENDER ETC)
// actually, we should rework this so they are function pointers... and the
// template compilation can support actual Go code

// VDom states: Just one, after being converted from html.Node

type VGNodeType uint32

const (
	ErrorNode    = VGNodeType(html.ErrorNode)
	TextNode     = VGNodeType(html.TextNode)
	DocumentNode = VGNodeType(html.DocumentNode)
	ElementNode  = VGNodeType(html.ElementNode)
	CommentNode  = VGNodeType(html.CommentNode)
	DoctypeNode  = VGNodeType(html.DoctypeNode)
)

type VGAtom uint32

type VGAttribute struct {
	Namespace, Key, Val string
}

func attrFromHtml(attr html.Attribute) VGAttribute {
	return VGAttribute{
		Namespace: attr.Namespace,
		Key:       attr.Key,
		Val:       attr.Val,
	}
}

// VGNode represents a node from our virtual DOM with the dynamic parts wired up into functions.
type VGNode struct {
	Parent, FirstChild, LastChild, PrevSibling, NextSibling *VGNode

	Type      VGNodeType
	DataAtom  VGAtom
	Data      string
	Namespace string
	Attr      []VGAttribute

	Props Props // dynamic attributes, used as input for components or converted to attributes for regular HTML elements

	InnerHTML string // indicates that children should be ignored and this raw HTML is the children of this tag

	// IfFunc       IfFunc
	// RangeFunc    func() ([]interface{}, error)
	// BindAttrFunc func() (map[string]interface{}, error)
	// TODO: EventFunc...
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
func (vgn *VGNode) Walk(f func(*VGNode) error) error {
	if vgn == nil {
		return nil
	}
	err := f(vgn)
	if err != nil {
		return err
	}
	err = vgn.FirstChild.Walk(f)
	if err != nil {
		return err
	}
	err = vgn.NextSibling.Walk(f)
	if err != nil {
		return err
	}
	return nil
}

var cruftHtml *html.Node
var cruftBody *html.Node

func init() {

	// startTime := time.Now()
	// defer func() { log.Printf("init() took %v", time.Since(startTime)) }() // NOTE: 35us on my laptop

	n, err := html.Parse(bytes.NewReader([]byte(`<html><body></body></html>`)))
	if err != nil {
		panic(err)
	}
	cruftHtml = n

	// find the body tag and store in cruftBody
	var walk func(n *html.Node)
	walk = func(n *html.Node) {

		// log.Printf("walk: %#v", n)

		if n == nil {
			return
		}

		if n.Type == html.ElementNode && n.Data == "body" {
			cruftBody = n
			return
		}

		if n.FirstChild != nil {
			walk(n.FirstChild)
		}
		if n.NextSibling != nil {
			walk(n.NextSibling)
		}
	}
	walk(cruftHtml)

	if cruftBody == nil {
		panic("unable to find <body> tag in html, something went terribly wrong")
	}
}

// // ParseTemplate takes a template file as a Reader and parses it.
// func ParseTemplate(r io.Reader) (*VGNode, error) {

// 	nodeList, err := html.ParseFragment(r, cruftBody)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// should be only on node with type Element and that's what we want
// 	var el *html.Node
// 	for _, n := range nodeList {
// 		if n.Type == html.ElementNode {
// 			if el != nil {
// 				return nil, fmt.Errorf("found more than one element at root of component template")
// 			}
// 			el = n
// 		}
// 	}
// 	if el == nil {
// 		return nil, fmt.Errorf("unable to find an element at root of component template")
// 	}

// 	return htmlToVGNode(el)

// 	// log.Printf("el = %+v", el)

// 	// if len(nodeList) != 1 {
// 	// 	return nil, fmt.Errorf("expected exactly one element but got %d instead", len(nodeList))
// 	// }

// 	// log.Printf("nodeList = %#v", nodeList)
// 	// log.Printf("nodeList[0].Type = %#v", nodeList[0].Type)
// 	// log.Printf("nodeList[0].Data = %#v", nodeList[0].Data)

// 	// return nil, nil
// }

// // htmlToVGNode recursively converts html.Node to VGNode
// func htmlToVGNode(rootN *html.Node) (*VGNode, error) {

// 	ptrMap := make(map[*html.Node]*VGNode)

// 	var conv func(*html.Node) (*VGNode, error)
// 	conv = func(n *html.Node) (*VGNode, error) {

// 		if n == nil {
// 			return nil, nil
// 		}

// 		// see if it's already in map, if so just return it
// 		vgn := ptrMap[n]
// 		if vgn != nil {
// 			return vgn, nil
// 		}

// 		var err error
// 		vgn = &VGNode{}
// 		// assign this first thing, so that everything below when it recurses will just point to the same instance
// 		ptrMap[n] = vgn

// 		// for all node pointers we recursively call conv, which will convert them or just return the pointer if already done
// 		// Parent
// 		vgn.Parent, err = conv(n.Parent)
// 		if err != nil {
// 			return vgn, err
// 		}
// 		// FirstChild
// 		vgn.FirstChild, err = conv(n.FirstChild)
// 		if err != nil {
// 			return vgn, err
// 		}
// 		// LastChild
// 		vgn.LastChild, err = conv(n.LastChild)
// 		if err != nil {
// 			return vgn, err
// 		}
// 		// PrevSibling
// 		vgn.PrevSibling, err = conv(n.PrevSibling)
// 		if err != nil {
// 			return vgn, err
// 		}
// 		// NextSibling
// 		vgn.NextSibling, err = conv(n.NextSibling)
// 		if err != nil {
// 			return vgn, err
// 		}

// 		// copy the other type and attr info
// 		vgn.Type = VGNodeType(n.Type)
// 		vgn.DataAtom = VGAtom(n.DataAtom)
// 		vgn.Data = n.Data
// 		vgn.Namespace = n.Namespace

// 		for _, nAttr := range n.Attr {
// 			switch {
// 			case nAttr.Key == "vg-if":
// 				vgn.VGIf = attrFromHtml(nAttr)
// 			case nAttr.Key == "vg-range":
// 				vgn.VGRange = attrFromHtml(nAttr)
// 			case strings.HasPrefix(nAttr.Key, ":"):
// 				vgn.BindAttr = append(vgn.BindAttr, attrFromHtml(nAttr))
// 			case strings.HasPrefix(nAttr.Key, "@"):
// 				vgn.EventAttr = append(vgn.EventAttr, attrFromHtml(nAttr))
// 			default:
// 				vgn.Attr = append(vgn.Attr, attrFromHtml(nAttr))
// 			}
// 		}

// 		// if len(n.Attr) > 0 {
// 		// 	vgn.Attr = make([]VGAttribute, 0, len(n.Attr))
// 		// 	for _, a := range n.Attr {
// 		// 		vgn.Attr = append(vgn.Attr, VGAttribute{
// 		// 			Namespace: a.Namespace,
// 		// 			Key:       a.Key,
// 		// 			Val:       a.Val,
// 		// 		})
// 		// 	}
// 		// }

// 		// now extract out our VG-specific stuff

// 		// log.Printf("vgn = %#v", vgn)

// 		return vgn, nil
// 	}
// 	return conv(rootN)

// }

/* OLD WORKING VERSION:

package vugu

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// Things to sort out:
// * vg-if, vg-range, etc (JUST COPIED FROM ATTRS, EVALUATED LATER)
// * :whatever syntax (ALSO JUST COPIED FROM ATTRS, EVALUATED LATER)
// * vg-model (IGNORE FOR NOW, WILL BE COMPOSED FROM OTHER FEATURES)
// * @click and other events (JUST A METHOD NAME TO INVOKE ON THE COMPONENT TYPE)
// * components and their properties (TAG NAME IS JUST CHECKED LATER, AND IF MATCH CAUSES THE EVAL STEPS ABOVE TO USE A DIFFERENT PATH, ATTRS GET REFS INSTEAD OF TEXT RENDER ETC)
// actually, we should rework this so they are function pointers... and the
// template compilation can support actual Go code

// VDom states: Just one, after being converted from html.Node

type VGNodeType uint32

const (
	ErrorNode    = VGNodeType(html.ErrorNode)
	TextNode     = VGNodeType(html.TextNode)
	DocumentNode = VGNodeType(html.DocumentNode)
	ElementNode  = VGNodeType(html.ElementNode)
	CommentNode  = VGNodeType(html.CommentNode)
	DoctypeNode  = VGNodeType(html.DoctypeNode)
)

type VGAtom uint32

type VGAttribute struct {
	Namespace, Key, Val string
}

func attrFromHtml(attr html.Attribute) VGAttribute {
	return VGAttribute{
		Namespace: attr.Namespace,
		Key:       attr.Key,
		Val:       attr.Val,
	}
}

// VGNode represents a node from our virtual DOM before any evaluation has been done.
// There is a direct correlation between template source and a tree of VGNodes.
// The golang.org/x/net/html package types are used a model and as input, so we can
// leverage it's parsing and writing capability and easily go between the two structures
// but still have our own fields.
type VGNode struct {
	Parent, FirstChild, LastChild, PrevSibling, NextSibling *VGNode

	Type      VGNodeType
	DataAtom  VGAtom
	Data      string
	Namespace string
	Attr      []VGAttribute

	VGIf      VGAttribute   // vg-if template string
	VGRange   VGAttribute   // vg-range template string
	BindAttr  []VGAttribute // attributes with colon prefix that are evaluated
	EventAttr []VGAttribute // event name -> method name
}

// Walk will walk the tree under a VGNode using the specified callback function.
// If the function provided returns a non-nil error then walking will be stopped
// and this error will be returned.  Only FirstChild and NextSibling are used
// while walking and so with well-formed documents should not loop. (But loops
// created manually by modifying FirstChild or NextSibling pointers could cause
// this function to recurse indefinitely.)
func (vgn *VGNode) Walk(f func(*VGNode) error) error {
	if vgn == nil {
		return nil
	}
	err := f(vgn)
	if err != nil {
		return err
	}
	err = vgn.FirstChild.Walk(f)
	if err != nil {
		return err
	}
	err = vgn.NextSibling.Walk(f)
	if err != nil {
		return err
	}
	return nil
}

var cruftHtml *html.Node
var cruftBody *html.Node

func init() {

	// startTime := time.Now()
	// defer func() { log.Printf("init() took %v", time.Since(startTime)) }() // NOTE: 35us on my laptop

	n, err := html.Parse(bytes.NewReader([]byte(`<html><body></body></html>`)))
	if err != nil {
		panic(err)
	}
	cruftHtml = n

	// find the body tag and store in cruftBody
	var walk func(n *html.Node)
	walk = func(n *html.Node) {

		// log.Printf("walk: %#v", n)

		if n == nil {
			return
		}

		if n.Type == html.ElementNode && n.Data == "body" {
			cruftBody = n
			return
		}

		if n.FirstChild != nil {
			walk(n.FirstChild)
		}
		if n.NextSibling != nil {
			walk(n.NextSibling)
		}
	}
	walk(cruftHtml)

	if cruftBody == nil {
		panic("unable to find <body> tag in html, something went terribly wrong")
	}
}

// ParseTemplate takes a template file as a Reader and parses it.
func ParseTemplate(r io.Reader) (*VGNode, error) {

	nodeList, err := html.ParseFragment(r, cruftBody)
	if err != nil {
		return nil, err
	}

	// should be only on node with type Element and that's what we want
	var el *html.Node
	for _, n := range nodeList {
		if n.Type == html.ElementNode {
			if el != nil {
				return nil, fmt.Errorf("found more than one element at root of component template")
			}
			el = n
		}
	}
	if el == nil {
		return nil, fmt.Errorf("unable to find an element at root of component template")
	}

	return htmlToVGNode(el)

	// log.Printf("el = %+v", el)

	// if len(nodeList) != 1 {
	// 	return nil, fmt.Errorf("expected exactly one element but got %d instead", len(nodeList))
	// }

	// log.Printf("nodeList = %#v", nodeList)
	// log.Printf("nodeList[0].Type = %#v", nodeList[0].Type)
	// log.Printf("nodeList[0].Data = %#v", nodeList[0].Data)

	// return nil, nil
}

// htmlToVGNode recursively converts html.Node to VGNode
func htmlToVGNode(rootN *html.Node) (*VGNode, error) {

	ptrMap := make(map[*html.Node]*VGNode)

	var conv func(*html.Node) (*VGNode, error)
	conv = func(n *html.Node) (*VGNode, error) {

		if n == nil {
			return nil, nil
		}

		// see if it's already in map, if so just return it
		vgn := ptrMap[n]
		if vgn != nil {
			return vgn, nil
		}

		var err error
		vgn = &VGNode{}
		// assign this first thing, so that everything below when it recurses will just point to the same instance
		ptrMap[n] = vgn

		// for all node pointers we recursively call conv, which will convert them or just return the pointer if already done
		// Parent
		vgn.Parent, err = conv(n.Parent)
		if err != nil {
			return vgn, err
		}
		// FirstChild
		vgn.FirstChild, err = conv(n.FirstChild)
		if err != nil {
			return vgn, err
		}
		// LastChild
		vgn.LastChild, err = conv(n.LastChild)
		if err != nil {
			return vgn, err
		}
		// PrevSibling
		vgn.PrevSibling, err = conv(n.PrevSibling)
		if err != nil {
			return vgn, err
		}
		// NextSibling
		vgn.NextSibling, err = conv(n.NextSibling)
		if err != nil {
			return vgn, err
		}

		// copy the other type and attr info
		vgn.Type = VGNodeType(n.Type)
		vgn.DataAtom = VGAtom(n.DataAtom)
		vgn.Data = n.Data
		vgn.Namespace = n.Namespace

		for _, nAttr := range n.Attr {
			switch {
			case nAttr.Key == "vg-if":
				vgn.VGIf = attrFromHtml(nAttr)
			case nAttr.Key == "vg-range":
				vgn.VGRange = attrFromHtml(nAttr)
			case strings.HasPrefix(nAttr.Key, ":"):
				vgn.BindAttr = append(vgn.BindAttr, attrFromHtml(nAttr))
			case strings.HasPrefix(nAttr.Key, "@"):
				vgn.EventAttr = append(vgn.EventAttr, attrFromHtml(nAttr))
			default:
				vgn.Attr = append(vgn.Attr, attrFromHtml(nAttr))
			}
		}

		// if len(n.Attr) > 0 {
		// 	vgn.Attr = make([]VGAttribute, 0, len(n.Attr))
		// 	for _, a := range n.Attr {
		// 		vgn.Attr = append(vgn.Attr, VGAttribute{
		// 			Namespace: a.Namespace,
		// 			Key:       a.Key,
		// 			Val:       a.Val,
		// 		})
		// 	}
		// }

		// now extract out our VG-specific stuff

		// log.Printf("vgn = %#v", vgn)

		return vgn, nil
	}
	return conv(rootN)

}


*/
