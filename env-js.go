// +build wasm

package vugu

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	js "syscall/js"
)

var _ js.Value

var _ Env = (*JSEnv)(nil) // assert type

var document js.Value
var domEventCB js.Func

func init() {
	document = js.Global().Get("document")
	// we use a single callback function for all of our event handling and dispatch the events from it
	domEventCB = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return jsEnv.handleRawDOMEvent(this, args)
	})
}

var jsEnv *JSEnv

// JSEnv is an environment that renders to DOM in webassembly applications.
type JSEnv struct {
	MountParent string // query selector

	DebugWriter io.Writer // write debug information about render details to this Writer if not nil

	reg      ComponentTypeMap
	rootInst *ComponentInst

	posJSNodeMap           map[uint64]js.Value        // keep track of element position hash value to js element, so we re-use existing nodes
	posElHashMap           map[uint64]uint64          // keep track of which element positions have which exact element hashes, so we don't touch nodes that are the same
	domEventHandlerHashMap map[uint64]DOMEventHandler // DOMEventHandler.hash() -> DOMEventHandler
	lastCSS                string                     // most recent css block value
}

// NewJSEnv returns a new instance of JSEnv.  The mountParent is a query selector of
// where in the DOM the rootInst component will be rendered inside, and components is
// a map of components to be made available.
func NewJSEnv(mountParent string, rootInst *ComponentInst, components ComponentTypeMap) *JSEnv {
	if components == nil {
		components = make(ComponentTypeMap)
	}
	ret := &JSEnv{
		MountParent:            mountParent,
		reg:                    components,
		rootInst:               rootInst,
		posJSNodeMap:           make(map[uint64]js.Value, 1024),
		posElHashMap:           make(map[uint64]uint64, 1024),
		domEventHandlerHashMap: make(map[uint64]DOMEventHandler, 32),
	}
	if jsEnv != nil {
		panic(fmt.Errorf("only one jsEnv allowed per application, for now"))
	}
	jsEnv = ret
	return ret
}

func (e *JSEnv) RegisterComponentType(tagName string, ct ComponentType) {
	e.reg[tagName] = ct
}

var pgmStart = time.Now()

func (e *JSEnv) debugf(s string, args ...interface{}) {
	if e.DebugWriter != nil {
		tel := time.Since(pgmStart)
		telms := tel.Truncate(time.Microsecond)
		// I don't really get what's happening here - it appears that log statements are extremely slow (like 200ms slow),
		// and they don't work at all from other goroutines.  So I'm just dropping in this in here and hoping the debug
		// logging isn't so bad it's unusable.  TODO: maybe just printing one summary at the end after everything is the way to go...
		// println("blah")
		fmt.Fprintf(e.DebugWriter, fmt.Sprintf("JSEnv.debug@%v: ", telms)+s+"\n", args...)
		// println(fmt.Sprintf("JSEnv.debug@%v: ", telms)+s+"\n", args...)
		// println(fmt.Sprintf("JSEnv.debug@"+telms.String()+": "+s, args...))
	}
}

// Render does the DOM syncing.
func (e *JSEnv) Render() (reterr error) {

	// TODO: watch out for concurrency issues with this being called from multiple goroutines, that's not going to work;
	// we probably should just error in this case; needs more consideration.

	// FIXME: We should defer+recover here to catch JS errors, which are translated to panics

	// log.Printf("HERE!!!")
	// ts := time.Now()
	// log.Print("testing1")
	// log.Printf("time: %v", time.Since(ts))
	// log.Print("testing2")

	renderStart := time.Now()
	// log.Print(time.Now())

	e.debugf("Render() starting")

	defer func() {
		// log.Print(time.Now())

		e.debugf("Render() exiting, total time %v (err=%v)", time.Since(renderStart), reterr)
	}()

	c := e.rootInst
	mountParentEl := document.Call("querySelector", e.MountParent)
	if !mountParentEl.Truthy() {
		return fmt.Errorf("failed to find mount parent using query selector %q", e.MountParent)
	}

	vdom, css, err := c.Type.BuildVDOM(c.Data)
	if err != nil {
		return err
	}
	_, _ = vdom, css

	// do basic setup and ensure we have a css style element and a root element, in that order
	mountChild1 := mountParentEl.Get("firstElementChild")
	var mountChild2 js.Value
	if mountChild1.Truthy() {
		mountChild2 = mountChild1.Get("nextElementSibling")
	}
	if !(strings.EqualFold("STYLE", mountChild1.Get("tagName").String()) &&
		strings.EqualFold(vdom.Data, mountChild2.Get("tagName").String())) {

		// something is wrong, just blow everything away and start over
		mountParentEl.Set("innerHTML", fmt.Sprintf(`<style>/* placeholder */</style><%s></%s>`, vdom.Data, vdom.Data))
		mountChild1 = mountParentEl.Get("firstElementChild")
		mountChild2 = mountChild1.Get("nextElementSibling")
		// log.Printf("mountChild1: %#v, mountChild2: %#v", mountChild1, mountChild2)

		// wipe out these too
		e.posJSNodeMap = make(map[uint64]js.Value, 1024)
		e.posElHashMap = make(map[uint64]uint64, 1024)
		e.domEventHandlerHashMap = make(map[uint64]DOMEventHandler, 32)
	}

	styleEl := mountChild1
	rootEl := mountChild2

	// drop in CSS if it's different
	var cssBuf bytes.Buffer
	if css != nil {
		for cssTxt := css.FirstChild; cssTxt != nil && cssTxt.Type == TextNode; cssTxt = cssTxt.NextSibling {
			fmt.Fprint(&cssBuf, cssTxt.Data)
			fmt.Fprint(&cssBuf, "\n")
		}
	}
	newCSS := cssBuf.String()
	if e.lastCSS != newCSS {
		styleEl.Set("textContent", newCSS)
		e.lastCSS = newCSS
	}

	// basic strategy is, starting with root component and doing the same with each
	// nested component:
	// * compute data's hash and use as starting point
	// * render vdom if hash doesn't match
	// * traverse vdom and compute a hash for each element based on (tbd - parent, tag name, and sibling position)
	// * where hash doesn't match, do the dom sync
	// * recurse into other components (create instance where needed, reuse if possible) as we encounter them
	// * prune extra html DOM as we go? (although possibly entire sub-tree contents gets nixed for now, we'll see)
	// * prune discarded component instances when done

	// rootEl := document.Call("createElement", "div")
	// rootEl, err = jsSyncNode(vdom, rootEl)
	// if err != nil {
	// log.Printf("got err 2: %v", err)
	// return err
	// }

	// build a new map of all of the positions we use during rendering
	newPosJSNodeMap := make(map[uint64]js.Value, len(e.posJSNodeMap))
	newPosElHashMap := make(map[uint64]uint64, len(e.posElHashMap))
	newDOMEventHandlerHashMap := make(map[uint64]DOMEventHandler, 32)

	// position hash 0 is always root element
	e.posJSNodeMap[0] = rootEl

	// walk the vdom, handle components along the way,
	// and sync to browser dom

	err = vdom.Walk(func(vgn *VGNode) error {

		// calculate vdom hash - has nothing to do with data, just the position
		// in the tree
		posh := vgn.positionHash()

		// e.debugf("vgn = %#v", vgn)

		{
			// TODO: components

			// check for component, using vdom hash to reuse if present

			// hash component data

			// if component data hash is different (or first time), regenerate its vdom

			// merge vdom into position here
		}

		// check for node with this position hash
		n := e.posJSNodeMap[posh]

		// if not exist, we're creating a new node
		if !n.Truthy() {

			switch vgn.Type {
			case ElementNode:
				n = document.Call("createElement", vgn.Data)
			case TextNode:
				n = document.Call("createTextNode", vgn.Data)
			case CommentNode:
				n = document.Call("createComment", vgn.Data)
			default:
				return fmt.Errorf("unable to handle unknown node type %v", vgn.Type)
			}

			// this should always work - there is always a parent that we can appendChild on for any node that needs to be created
			parentN := e.posJSNodeMap[vgn.Parent.positionHash()]
			parentN.Call("appendChild", n)

			// // check for previous sibling and attach that way
			// if !n.Truthy() && vgn.PrevSibling != nil {
			// 	prevPosH := vgn.PrevSibling.positionHash()
			// 	if prevn, ok := e.posJSNodeMap[prevPosH]; ok {
			// 		// create N from
			// 		document.Call("createElement", vgn.Data)
			// 		prevn.Call("insertAdjacentElement", "afterend")
			// 	}
			// }

		}

		// use position hash to look up element hash and compare to new vdom element hash
		elHash := vgn.elementHash(posh) // hash of position+contents of this vdom element

		// check if element is different than last recorded state
		if elHash != e.posElHashMap[posh] {
			// do a sync
			newEl, err := e.jsSyncNode(vgn, n, newDOMEventHandlerHashMap)
			if err != nil {
				return err
			}
			n = newEl
		}

		// assign node to both new and old, old is used in cases where we grab the parent
		e.posJSNodeMap[posh] = n
		newPosJSNodeMap[posh] = n

		// update in new posEl hash map
		newPosElHashMap[posh] = elHash

		// --

		// see if a node exists for this vdom element hash, if so we're done,
		// otherwise hit the dom and sync

		// 	// element name must match a component
		// 	ct, ok := e.reg[vgn.Data]
		// 	if !ok {
		// 		return nil
		// 	}

		// 	// copy props and merge in static attributes where they don't conflict
		// 	props := vgn.Props.Clone()
		// 	for _, a := range vgn.Attr {
		// 		if _, ok := props[a.Key]; !ok {
		// 			props[a.Key] = a.Val
		// 		}
		// 	}

		// 	// just make a new instance each time - this is static html output
		// 	compInst, err := New(ct, props)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	cdom, ccss, err := ct.BuildVDOM(compInst.Data)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	if ccss != nil && ccss.FirstChild != nil {
		// 		css.AppendChild(ccss.FirstChild)
		// 	}

		// 	// make cdom replace vgn

		// 	// point Parent on each child of cdom to vgn
		// 	for cn := cdom.FirstChild; cn != nil; cn = cn.NextSibling {
		// 		cn.Parent = vgn
		// 	}
		// 	// replace vgn with cdom but preserve vgn.Parent
		// 	*vgn, vgn.Parent = *cdom, vgn.Parent

		return nil
	})

	if err != nil {
		return err
	}

	// to remove elements that are no longer part of the new virtual dom, we look for the elements that are in
	// e.posJSNodeMap but not in newPosJSNodeMap and call remove() on them to remove them from the browser DOM
	for k, v := range e.posJSNodeMap {
		if _, ok := newPosJSNodeMap[k]; !ok {
			v.Call("remove") // remove from DOM
		}
	}

	// call Release on funcs that are no longer being used
	// for k, f := range e.domEventHandlerHashMap {
	// 	if _, ok := newDOMEventHandlerHashMap[k]; !ok {
	// 		f.Release()
	// 	}
	// }

	// replace our maps with the new ones we've just created, which effectively trims any values that are no longer used
	// TODO: is there a better way to do this that doesn't result in so much garbage collection?
	e.posJSNodeMap = newPosJSNodeMap
	e.posElHashMap = newPosElHashMap
	e.domEventHandlerHashMap = newDOMEventHandlerHashMap

	return nil

	// // The basic strategy is to build an equivalent html.Node tree from our vdom, expanding InnerHTML along
	// // the way, and then tell the html package to write it out

	// // output css
	// if css != nil && css.FirstChild != nil {

	// 	cssn := &html.Node{
	// 		Type:     html.ElementNode,
	// 		Data:     "style",
	// 		DataAtom: atom.Style,
	// 	}
	// 	cssn.AppendChild(&html.Node{
	// 		Type: html.TextNode,
	// 		Data: css.FirstChild.Data,
	// 	})

	// 	err = html.Render(out, cssn)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// ptrMap := make(map[*VGNode]*html.Node)

	// var conv func(*VGNode) (*html.Node, error)
	// conv = func(vgn *VGNode) (*html.Node, error) {

	// 	if vgn == nil {
	// 		return nil, nil
	// 	}

	// 	// see if it's already in map, if so just return it
	// 	if n := ptrMap[vgn]; n != nil {
	// 		return n, nil
	// 	}

	// 	var err error
	// 	n := &html.Node{}
	// 	// assign this first thing, so that everything below when it recurses will just point to the same instance
	// 	ptrMap[vgn] = n

	// 	// for all node pointers we recursively call conv, which will convert them or just return the pointer if already done
	// 	// Parent
	// 	n.Parent, err = conv(vgn.Parent)
	// 	if err != nil {
	// 		return n, err
	// 	}
	// 	// FirstChild
	// 	n.FirstChild, err = conv(vgn.FirstChild)
	// 	if err != nil {
	// 		return n, err
	// 	}
	// 	// LastChild
	// 	n.LastChild, err = conv(vgn.LastChild)
	// 	if err != nil {
	// 		return n, err
	// 	}
	// 	// PrevSibling
	// 	n.PrevSibling, err = conv(vgn.PrevSibling)
	// 	if err != nil {
	// 		return n, err
	// 	}
	// 	// NextSibling
	// 	n.NextSibling, err = conv(vgn.NextSibling)
	// 	if err != nil {
	// 		return n, err
	// 	}

	// 	// copy the other type and attr info
	// 	n.Type = html.NodeType(vgn.Type)
	// 	n.DataAtom = atom.Atom(vgn.DataAtom)
	// 	n.Data = vgn.Data
	// 	n.Namespace = vgn.Namespace

	// 	for _, vgnAttr := range vgn.Attr {
	// 		n.Attr = append(n.Attr, html.Attribute{Namespace: vgnAttr.Namespace, Key: vgnAttr.Key, Val: vgnAttr.Val})
	// 	}

	// 	// parse and expand InnerHTML if present
	// 	if vgn.InnerHTML != "" {

	// 		innerNs, err := html.ParseFragment(bytes.NewReader([]byte(vgn.InnerHTML)), cruftBody)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		// FIXME: do we just append all of this, what about case where there is already something inside?
	// 		for _, innerN := range innerNs {
	// 			n.AppendChild(innerN)
	// 		}

	// 	}

	// 	return n, nil
	// }
	// outn, err := conv(vdom)
	// if err != nil {
	// 	return err
	// }
	// // log.Printf("outn: %#v", outn)

	// err = html.Render(out, outn)
	// if err != nil {
	// 	return err
	// }

	// return nil
}

// func canJsSyncNode(vgn *VGNode) bool {
// 	switch vgn.Type {
// 	case ElementNode, TextNode, CommentNode:
// 		return true
// 	}
// 	return false
// }

func (e *JSEnv) handleRawDOMEvent(this js.Value, args []js.Value) interface{} {

	if len(args) < 1 {
		panic(fmt.Errorf("args should be at least 1 element, instead was: %#v", args))
	}

	jsEvent := args[0]

	typeName := jsEvent.Get("type").String()

	key := "vugu_event_" + typeName + "_id"
	funcIDString := this.Get(key).String()
	var funcID uint64
	fmt.Sscanf(funcIDString, "%d", &funcID)

	if funcID == 0 {
		panic(fmt.Errorf("looking for %q on 'this' found %q which parsed into value 0 - cannot find the appropriate function to route to", key, funcIDString))
	}

	handler, ok := e.domEventHandlerHashMap[funcID]
	if !ok {
		panic(fmt.Errorf("nothing found in domEventHandlerHashMap for %d", funcID))
	}

	domE := &DOMEvent{
		jsEvent:     jsEvent,
		jsEventThis: this,
	}

	rvargs := make([]reflect.Value, 0, len(handler.Args))
	for _, a := range handler.Args {
		// anything of type *DOMEvent gets replaced with our DOMEvent instance
		if _, ok := a.(*DOMEvent); ok {
			rvargs = append(rvargs, reflect.ValueOf(domE))
		} else {
			// and everything else just goes as-is
			v := reflect.ValueOf(a)
			rvargs = append(rvargs, v)
		}
	}

	ret := handler.Method.Call(rvargs)

	// if it came back with a single bool value then return that, otherwise return null
	if len(ret) == 1 {
		rv := reflect.ValueOf(ret[0])
		if rv.Kind() == reflect.Bool {
			return rv.Bool()
		}
	}

	return nil
}

// jsSyncNode will take a virtual dom element and update a browser DOM element to match it,
// or if this is not possible the element will be replaced entirely; either way
// as long as no error the correct new element will be returned; emap gets set with
// all of the event handlers we set (or would set if not already)
func (e *JSEnv) jsSyncNode(vgn *VGNode, el js.Value, emap map[uint64]DOMEventHandler) (newEl js.Value, reterr error) {

	// FIXME: Is there a way to merge all this so we only ship one set of data over to the JS side
	// and do the rest from there?  Might be much faster...

	if !el.Truthy() {
		reterr = fmt.Errorf("el is not truthy, cannot sync node")
		return
	}

	newEl = el

	switch vgn.Type {

	case ElementNode:

		// see if it's the same tag name, if not we need to replace the tag and return the new one
		tagName := newEl.Get("tagName").String()
		if !strings.EqualFold(vgn.Data, tagName) {
			newEl = document.Call("createElement", vgn.Data)

			// insert new and remove old - note that the old may not be an element, could be text or comment

			parentNode := el.Get("parentNode")
			parentNode.Call("insertBefore", newEl, el) // insert new one before old

			// move children over from el to newEl
			elChildNodes := el.Get("childNodes")
			elChildNodesLength := elChildNodes.Get("length").Int()
			for i := 0; i < elChildNodesLength; i++ {
				childN := elChildNodes.Call("item", 0) // get first element of el childs
				newEl.Call("appendChild", childN)      // move to end of newEl childs
			}

			el.Call("remove") // remove old el

		}

		// TODO: optimize case where both vgn and newEl have no attributes or events as this is very common

		// TODO: is it faster to just set the attributes and clobber what is there or to check the values first and only
		// set the ones that need changing? needs research

		// TODO: also it might be faster to build the node as a string and replace rather than various attribute calls

		// now that we have the right type of tag, sync the attributes, including rendering dynamic ones to text
		attrNames := make(map[string]bool, len(vgn.Attr)+len(vgn.Props))
		// static attributes
		for _, a := range vgn.Attr {
			attrNames[a.Key] = true
			newEl.Call("setAttribute", a.Key, a.Val)
		}
		// props get converted to attributes
		for k, v := range vgn.Props {
			attrNames[k] = true
			newEl.Call("setAttribute", k, fmt.Sprint(v))
		}

		// look through and prune any left that were not set above
		var rmNames []string
		attributes := newEl.Get("attributes")
		l := attributes.Get("length").Int()
		for i := 0; i < l; i++ {
			name := attributes.Call("item", i).Get("name").String()
			if !attrNames[name] {
				rmNames = append(rmNames, name)
			}
		}
		for _, name := range rmNames {
			newEl.Call("removeAttribute", name)
		}

		// if InnerHTML then set it
		if vgn.InnerHTML != "" {
			newEl.Set("innerHTML", vgn.InnerHTML)
		}

		// now handle event wiring
		for eventName, handler := range vgn.DOMEventHandlers {
			keyName := "vugu_event_" + eventName + "_id"
			hash := handler.hash()
			keyVal := fmt.Sprint(hash)
			emap[hash] = handler
			oldKeyJSVal := newEl.Get(keyName)
			if !oldKeyJSVal.Truthy() {
				// never been added
				newEl.Call("addEventListener", eventName, domEventCB) // global listener handles it all
			}
			newEl.Set(keyName, keyVal) // set key to point it at the right handler when the call comes in
		}

		return

	case TextNode:

		elNodeType := newEl.Get("nodeType").Int()
		if elNodeType == 3 { // 3 means text node
			// already a text node, just set it's contents
			newEl.Set("data", vgn.Data)
			return
		}

		// what's there is not a text node, need to replace
		newEl = document.Call("createTextNode", vgn.Data)
		parentNode := el.Get("parentNode")
		parentNode.Call("insertBefore", newEl, el) // insert new one before old
		el.Call("remove")                          // remove old

		return

	case CommentNode:

		elNodeType := newEl.Get("nodeType").Int()
		if elNodeType == 8 { // 8 means comment node
			// already a comment node, just set it's contents
			newEl.Set("data", vgn.Data)
			return
		}

		// what's there is not a comment node, need to replace
		newEl = document.Call("createComment", vgn.Data)
		parentNode := el.Get("parentNode")
		parentNode.Call("insertBefore", newEl, el) // insert new one before old
		el.Call("remove")                          // remove old

		return
	}

	reterr = fmt.Errorf("cannot sync node of type %v", vgn.Type)
	return
}

func jsRemoveChildren(v js.Value) {
	if !v.Truthy() {
		return
	}
	for firstChild := v.Get("firstChild"); firstChild.Truthy(); firstChild = v.Get("firstChild") {
		v.Call("removeChild", firstChild)
	}
}
