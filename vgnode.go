package vugu

// Things to sort out:
// * vg-if, vg-range, etc (JUST COPIED FROM ATTRS, EVALUATED LATER)
// * :whatever syntax (ALSO JUST COPIED FROM ATTRS, EVALUATED LATER)
// * vg-model (IGNORE FOR NOW, WILL BE COMPOSED FROM OTHER FEATURES)
// * @click and other events (JUST A METHOD NAME TO INVOKE ON THE COMPONENT TYPE)
// * components and their properties (TAG NAME IS JUST CHECKED LATER, AND IF MATCH CAUSES THE EVAL STEPS ABOVE TO USE A DIFFERENT PATH, ATTRS GET REFS INSTEAD OF TEXT RENDER ETC)
// actually, we should rework this so they are function pointers... and the
// template compilation can support actual Go code

// VDom states: Just one, after being converted from html.Node

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

	// DOMEventHandlers map[string]DOMEventHandler // describes invocations when DOM events happen

	DOMEventHandlerSpecList []DOMEventHandlerSpec // replacing DOMEventHandlers

	// indicates this node's output should be delegated to the specified component
	Component interface{}

	// default of 0 means it will be calculated when positionHash() is called
	// positionHashVal uint64
}

// // SetDOMEventHandler will assign a named event to DOMEventHandlers (will allocate the map if nil).
// // Used during VDOM construction and during render to determine browser events to hook up.
// func (n *VGNode) SetDOMEventHandler(name string, h DOMEventHandler) {
// 	if n.DOMEventHandlers == nil {
// 		n.DOMEventHandlers = map[string]DOMEventHandler{name: h}
// 		return
// 	}
// 	n.DOMEventHandlers[name] = h
// }

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

// // positionHash calculates a hash based on position in the tree only (specifically it does not consider element attributes),
// // stores the result in positionHashVal and returns it.
// // The same element position will have the same hash, regardless of element contents.  Each unique position in the vdom should have
// // a unique positionHash value, allowing for the usual hash collision caveat.
// // (subsequent calls return positionHashVal).  Special case: root nodes (Parent==nil) always return 0.
// // This is used to efficiently answer the question "is this vdom element structurally in the same position as ...".
// func (vgn *VGNode) positionHash() uint64 {

// 	if vgn.positionHashVal != 0 {
// 		return vgn.positionHashVal
// 	}

// 	// root element always returns 0
// 	if vgn.Parent == nil {
// 		return 0
// 	}

// 	b8 := make([]byte, 8)

// 	h := xxhash.New()

// 	// hash in parent's positionHash
// 	binary.BigEndian.PutUint64(b8, vgn.Parent.positionHash())
// 	h.Write(b8)

// 	// calculate our sibling depth
// 	var n uint64
// 	for prev := vgn.PrevSibling; prev != nil; prev = prev.PrevSibling {
// 		n++
// 	}

// 	// hash in sibling depth
// 	binary.BigEndian.PutUint64(b8, n)
// 	h.Write(b8)

// 	// that's it
// 	vgn.positionHashVal = h.Sum64()
// 	return vgn.positionHashVal
// }

// // elementHash calculates and returns a hash of just the contents of this element - it's type, Data, Attrs and Props and InnerHTML,
// // it specifically does not include element position data like any of the Parent, NextSibling, etc pointers.
// // The same element at different positions in the tree, assuming the same start param, will result in the same hash.
// // FIXME: will need to add events here.
// // The return value is not cached, it is calculated newly each time.
// // A starting point to be hashed in can be provided also if needed, otherwise pass 0.
// func (vgn *VGNode) elementHash(start uint64) uint64 {

// 	b8 := make([]byte, 8)

// 	h := xxhash.New()

// 	if start != 0 {
// 		binary.BigEndian.PutUint64(b8, start)
// 		h.Write(b8)
// 	}

// 	// type (element, text, comment)
// 	binary.BigEndian.PutUint64(b8, uint64(vgn.Type))
// 	h.Write(b8)

// 	// name of the element or text content
// 	fmt.Fprint(h, vgn.Data)

// 	// hash static attrs
// 	for _, a := range vgn.Attr {
// 		fmt.Fprint(h, a.Key)
// 		fmt.Fprint(h, a.Val)
// 	}

// 	// // hash props (dynamic attrs)
// 	// pks := vgn.Props.OrderedKeys() // stable sequence
// 	// for _, pk := range pks {
// 	// 	fmt.Fprint(h, pk) // key goes in as string

// 	// 	// use ComputeHash on the value because we don't know it's type
// 	// 	vh := ComputeHash(vgn.Props[pk])
// 	// 	binary.BigEndian.PutUint64(b8, vh)
// 	// 	h.Write(b8)
// 	// }

// 	// hash events
// 	if len(vgn.DOMEventHandlers) > 0 {
// 		keys := make([]string, len(vgn.DOMEventHandlers))
// 		for k := range vgn.DOMEventHandlers {
// 			keys = append(keys, k)
// 		}
// 		sort.Strings(keys)
// 		for _, k := range keys {
// 			vh := vgn.DOMEventHandlers[k].hash()
// 			binary.BigEndian.PutUint64(b8, vh)
// 			h.Write(b8)
// 		}
// 	}

// 	// innerHTML if any
// 	if vgn.InnerHTML != nil {
// 		fmt.Fprint(h, *vgn.InnerHTML)
// 	}

// 	return h.Sum64()
// }

// var cruftHtml *html.Node
// var cruftBody *html.Node

// func init() {

// 	// startTime := time.Now()
// 	// defer func() { log.Printf("init() took %v", time.Since(startTime)) }() // NOTE: 35us on my laptop

// 	n, err := html.Parse(bytes.NewReader([]byte(`<html><body></body></html>`)))
// 	if err != nil {
// 		panic(err)
// 	}
// 	cruftHtml = n

// 	// find the body tag and store in cruftBody
// 	var walk func(n *html.Node)
// 	walk = func(n *html.Node) {

// 		// log.Printf("walk: %#v", n)

// 		if n == nil {
// 			return
// 		}

// 		if n.Type == html.ElementNode && n.Data == "body" {
// 			cruftBody = n
// 			return
// 		}

// 		if n.FirstChild != nil {
// 			walk(n.FirstChild)
// 		}
// 		if n.NextSibling != nil {
// 			walk(n.NextSibling)
// 		}
// 	}
// 	walk(cruftHtml)

// 	if cruftBody == nil {
// 		panic("unable to find <body> tag in html, something went terribly wrong")
// 	}
// }
