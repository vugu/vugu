package vugu

import (
	"fmt"
	"log"
	"strings"
	"time"

	js "github.com/vugu/vugu/js"
)

//go:generate go run renderer-js-script-maker.go

// NewJSRenderer will create a new JSRenderer with the speicifc mount point selector.
// If an empty string is passed then the root component should include a top level <html> tag
// and the entire page will be rendered.
func NewJSRenderer(mountPointSelector string) (*JSRenderer, error) {

	ret := &JSRenderer{
		MountPointSelector: mountPointSelector,
	}

	ret.domEventCB = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return jsEnv.handleRawDOMEvent(this, args)
	})

	ret.instructionBuffer = make([]byte, 4096)
	ret.instructionTypedArray = js.TypedArrayOf(ret.instructionBuffer)

	ret.window = js.Global().Get("window")

	ret.window.Call("eval", jsHelperScript)

	ret.instructionList = newInstructionList(ret.instructionBuffer, func(il *instructionList) error {

		// call vuguRender to have the instructions processed in JS
		ret.window.Call("vuguRender", ret.instructionTypedArray)

		return nil
	})

	// log.Printf("ret.window: %#v", ret.window)
	// log.Printf("eval: %#v", ret.window.Get("eval"))

	return ret, nil
}

// JSRenderer implements Renderer against the browser's DOM.
type JSRenderer struct {
	MountPointSelector string

	domEventCB js.Func // the callback function for DOM events

	instructionBuffer     []byte
	instructionTypedArray js.TypedArray
	instructionList       *instructionList

	window js.Value
}

// Release calls release on any resources that this renderer allocated.
func (r *JSRenderer) Release() {
	r.instructionTypedArray.Release()
}

// Render implements Renderer.
func (r *JSRenderer) Render(bo *BuildOut) error {
	if !js.Global().Truthy() {
		return fmt.Errorf("js environment not available")
	}

	if bo == nil {
		return fmt.Errorf("BuildOut is nil")
	}

	if bo.Doc == nil {
		return fmt.Errorf("BuildOut.Doc is nil")
	}

	if bo.Doc.Type != ElementNode {
		return fmt.Errorf("BuildOut.Doc.Type is (%v), not ElementNode", bo.Doc.Type)
	}

	// log.Printf("BuildOut: %#v", b)

	// NOTE:
	// We need two different strategies for rendering elements in <html> or <head>
	// as opposed to the stuff once we're inside <body>.  If we have a specific element
	// we're rendering into we have complete control and can just replace it out entirely.
	// But for <head> we want things like existing meta tags, title tags, and script includes not be removed
	// when we sync.
	// Interestingly enough, we also want to override the title tag (not add a second one).
	// Meta tags likewise it would make the most sense to selectively replace them by their name attribute.
	// So what we're headed for is something like we record what the <head> section looked like at startup,
	// and then based on the tag type we have different handlings - like meta tags we set by name, title
	// tag we always update if present, script tags we just avoid duplication if the tag is already there
	// exactly the same or with the same src, and so on.  But then you if you just drop in some random
	// crap it will just be added.
	// This is fine for <head> - it's easy to identify. But what about script tags added toward the bottom
	// of the page just inside the body tag.  Potentially we could get away with just nuking the script tags,
	// but what if someone puts in in-line styles in a style tag - replacing it would definitely remove
	// the styling - unexpected for sure.  There needs to be some well-defined rules about this.
	// Perhaps we have separately logic for the <head>, and then otherwise we still have a designated
	// element within body which is what we target.  It should be possible to just make this the body
	// tag if nobody care, but if they need to be able to do other custom stuff outside of head, it should
	// be possible - while still controlling title and meta tags etc from the Vugu app.

	// const (
	// 	modeHTML          int = iota // in html tag
	// 	modeHead                     // in head tag
	// 	modeBodyUnmounted            // in body tag but not yet mounted
	// 	modeMounted                  // in the mounted tag (could be body or something else)
	// )

	// var mode int
	// switch strings.ToLower(bo.Doc.Data) {
	// case "html":
	// 	mode = modeHTML
	// case "head":
	// 	return fmt.Errorf("BuildOut.Doc is a head element, use html instead")
	// case "body":
	// 	mode = modeBodyUnmounted
	// default:
	// 	mode = modeBodyUnmounted
	// }

	// _ = mode

	// TODO: we need to make sure this call path back and forth to the outside for DOM syncing is
	// efficient - it determines Vugu's overall performance characteristics to a large degree.
	// Do some tests and see what the performance difference is if pass data directly using Call()
	// a number of times vs shipping JSON over in an ArrayBuffer and parse and process it in JS.
	// NOTE: doing Call() all over the place and getting/passing various element referances around
	// causes a memory leak, since the Go code has no way of gargage collecting the various references.
	// And while the situation should improve at some point it may be a while and it might be prudent.
	// to use a solution that does not leak (as much).  Will give a more stable feel until full DOM
	// access is figured out in WASM.

	// Test results:

	// 25us per call
	// g.Call("eval", "function testf1(a) { return a + 'f1'; }")

	// 42us per call
	// g.Call("eval", "function testf1(a) { var e = document.createElement('div'); e.innerHTML = a + 'f1'; document.body.appendChild(e); }")

	// Conclusion: it takes twice as long to call from Go into JS as it does to create and attach an HTML element.
	// Ergo, we need to minimize the number of Call()s we do.

	// Approach: used a typed array buffer to produce a sort of simple set of instructions that describe how to synchronize
	// the DOM tree.  The Go code can optimize away nodes that don't need to be changed, but if they do then it writes
	// an instruction into the queue - once this queue/buffer is filled up with an appropriate number of instructions, a
	// single call is done over to JS, which reads and executes these instructions.
	// It's a bit of work, but it will force us to break down and think through the synchronization and the result
	// should be much much faster.
	// (While we're at it, we should also see if we can optimize the callback path for events - so those use a preallocated
	// buffer as well and avoid accumulating references for each event)

	// r.instructionBuffer[0] = 7
	// r.instructionBuffer[1] = 9

	log.Printf("BuildOut: %#v", bo)

	el := bo.Doc
	log.Printf("el: %#v", el)

	// NOTE: Mount rules:
	// <body>, <head> forbidden as top level component tag
	// * if component tag is not <html>, then whatever it is gets mounted at mount point
	// * if component tag is <html>, then html attrs are sync, head elements are synced, body attrs are synced,
	//   and first element inside <body> is mounted at mount point

	// how do we do this mountpoint thing, it's pretty important...

	// start cases:
	// * starts with html tag
	// * starts with something else
	// loop cases:
	// * in html, waiting for head or body
	// * in head, needs careful replacement
	// * in body, waiting for mount point
	// * inside mounted aread, main dom sync logic

	err := r.visitFirst(bo, bo.Doc)
	if err != nil {
		return err
	}

	err = r.instructionList.flush()
	if err != nil {
		return err
	}

	return nil

	// il := r.instructionList

	// const (
	// 	modeStart int = iota // first element
	// 	// modeHTML                // in html tag
	// 	modePreMount // found the tag that needs to be mounted
	// 	modeHead     // in head tag
	// 	// modeBodyUnmounted            // in body tag but not yet mounted
	// 	// modeMounted                  // in the mounted tag (could be body or something else)
	// )

	// // TODO: replace these strings.ToLower(n.Data) == "tagname" with Atoms

	// var visit func(n *VGNode, mode int) error
	// visit = func(n *VGNode, mode int) error {

	// 	switch mode {
	// 	case modeStart:

	// 		if n.Type != ElementNode {
	// 			return fmt.Errorf("root of component must be element")
	// 		}

	// 		// first tag is html
	// 		if strings.ToLower(n.Data) == "html" {

	// 			// TODO: sync html tag attributes

	// 			for nchild := n.FirstChild; nchild != nil; nchild = nchild.NextSibling {

	// 				if strings.ToLower(nchild.Data) == "head" {
	// 					err := visit(nchild, modeHead)
	// 					if err != nil {
	// 						return err
	// 					}
	// 					continue
	// 				} else if strings.ToLower(nchild.Data) == "body" {

	// 					continue
	// 				}

	// 				return fmt.Errorf("unexpected tag inside html %q (VGNode=%#v)", nchild.Data, nchild)

	// 			}

	// 			return nil
	// 		}

	// 		// else, first tag is anything else - set mode to pre mount and try again
	// 		mode = modePreMount
	// 		return visit(n)

	// 	default:
	// 		return fmt.Errorf("unknown mode %v", mode)
	// 	}

	// 	return nil
	// }

	// err := visit(bo.Doc)
	// if err != nil {
	// 	return err
	// }

	// // startTime := time.Now()

	// // for i := 0; i < 10; i++ {
	// // il.writeClearRefmap()
	// // il.writeSetHTMLRef(99)
	// // il.writeSelectRef(99)
	// // // il.writeSetAttrStr("lang", "en-gb")
	// // il.writeSetAttrStr("whatever", "yes it is")
	// // // }
	// // il.writeEnd()
	// // log.Printf("instruction write time: %v", time.Since(startTime))

	// err = il.flush()
	// if err != nil {
	// 	return err
	// }

	// // r.window.Call("vuguRender", r.instructionTypedArray)
	// // r.window.Call("vuguRender")

	// // log.Printf("at pos 6: %#v", r.instructionBuffer[6])

	// // panic(fmt.Errorf("not yet implemented"))
	// return nil
}

// EventWait blocks until an event has occurred which causes a re-render.
// It returns true if the render loop should continue or false if it should exit.
func (r *JSRenderer) EventWait() bool {

	// make sure the JS environment is still available, returning false otherwise
	if !js.Global().Truthy() {
		return false
	}

	// TODO: implement event loop
	time.Sleep(10 * time.Second)

	return true
}

func (r *JSRenderer) handleRawDOMEvent(this js.Value, args []js.Value) interface{} {
	panic(fmt.Errorf("not yet implemented"))
	return nil
}

// var window js.Value

// func init() {
// 	window = js.Global().Get("window")
// 	if window.Truthy() {
// 		js.Global().Call("eval", jsHelperScript)
// 	}
// }

func (r *JSRenderer) visitFirst(bo *BuildOut, n *VGNode) error {

	log.Printf("TODO: We need to go through and optimize away unneeded calls to create elements, set attributes, set event handlers, etc. for cases where they are the same per hash")

	log.Printf("JSRenderer.visitFirst")

	if n.Type != ElementNode {
		return fmt.Errorf("root of component must be element")
	}

	err := r.instructionList.writeClearEl()
	if err != nil {
		return err
	}

	// first tag is html
	if strings.ToLower(n.Data) == "html" {

		// TODO: sync html tag attributes

		for nchild := n.FirstChild; nchild != nil; nchild = nchild.NextSibling {

			if strings.ToLower(nchild.Data) == "head" {

				err := r.visitHead(bo, nchild)
				if err != nil {
					return err
				}

			} else if strings.ToLower(nchild.Data) == "body" {

				err := r.visitBody(bo, nchild)
				if err != nil {
					return err
				}

			} else {
				return fmt.Errorf("unexpected tag inside html %q (VGNode=%#v)", nchild.Data, nchild)
			}

		}

		return nil
	}

	// else, first tag is anything else - try again as the element to be mounted
	return r.visitMount(bo, n)

}

func (r *JSRenderer) visitHead(bo *BuildOut, n *VGNode) error {
	log.Printf("TODO: visitHead")
	return nil
}

func (r *JSRenderer) visitBody(bo *BuildOut, n *VGNode) error {
	log.Printf("TODO: visitBody")
	return nil
}

func (r *JSRenderer) visitMount(bo *BuildOut, n *VGNode) error {

	log.Printf("visitMount got here")

	err := r.instructionList.writeSelectMountPoint(r.MountPointSelector, n.Data)
	if err != nil {
		return err
	}

	return r.visitSyncElementEtc(bo, n)

	// err = r.writeAllStaticAttrs(n)
	// if err != nil {
	// 	return err
	// }

	// if n.FirstChild != nil {

	// 	err = r.instructionList.writeMoveToFirstChild()
	// 	if err != nil {
	// 		return err
	// 	}

	// 	// err = r.instructionList.writePicardFirstChild(uint8(n.FirstChild.Type), n.FirstChild.Data)
	// 	// if err != nil {
	// 	// 	return err
	// 	// }

	// 	// err := r.writePicardFirstChildNode(n)
	// 	// if err != nil {
	// 	// 	return err
	// 	// }

	// 	for nchild := n.FirstChild; nchild != nil; nchild = nchild.NextSibling {
	// 		err = r.visitSyncNode(bo, nchild)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		err = r.instructionList.writeMoveToNextSibling()
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}

	// 	err = r.instructionList.writeMoveToParent()
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (r *JSRenderer) visitSyncNode(bo *BuildOut, n *VGNode) error {

	log.Printf("visitSyncNode")

	var err error

	switch n.Type {
	case ElementNode:
		err = r.instructionList.writeSetElement(n.Data)
		if err != nil {
			return err
		}
	case TextNode:
		return r.instructionList.writeSetText(n.Data) // no children possible, just return
	case CommentNode:
		return r.instructionList.writeSetComment(n.Data) // no children possible, just return
	default:
		return fmt.Errorf("unknown node type %v", n.Type)
	}

	// only elements have attributes, child or events
	return r.visitSyncElementEtc(bo, n)

}

// visitSyncElementEtc syncs the rest of the stuff that only applies to elements
func (r *JSRenderer) visitSyncElementEtc(bo *BuildOut, n *VGNode) error {

	err := r.writeAllStaticAttrs(n)
	if err != nil {
		return err
	}

	err = r.instructionList.writeRemoveOtherAttrs()
	if err != nil {
		return err
	}

	if n.FirstChild != nil {

		err = r.instructionList.writeMoveToFirstChild()
		if err != nil {
			return err
		}

		for nchild := n.FirstChild; nchild != nil; nchild = nchild.NextSibling {
			err = r.visitSyncNode(bo, nchild)
			if err != nil {
				return err
			}
			err = r.instructionList.writeMoveToNextSibling()
			if err != nil {
				return err
			}
		}

		err = r.instructionList.writeMoveToParent()
		if err != nil {
			return err
		}
	}

	return nil
}

// writeAllStaticAttrs is a helper to write all the static attrs from a VGNode
func (r *JSRenderer) writeAllStaticAttrs(n *VGNode) error {
	for _, a := range n.Attr {
		err := r.instructionList.writeSetAttrStr(a.Key, a.Val)
		if err != nil {
			return err
		}
	}
	return nil
}

// // writePicardFirstChildNode calls writePicardFirstChildElement or other variation based on node type
// func (r *JSRenderer) writePicardFirstChildNode(n *VGNode) error {

// 	return r.instructionList.writePicardFirstChild(htmlx.NodeType(n.Type),n.Data)

// 	switch n.Type {
// 	case ElementNode:
// 		return r.instructionList.writePicardFirstChild(n.Data)
// 	case TextNode:
// 		return r.instructionList.writePicardFirstChildText(n.Data)
// 	case CommentNode:
// 		return r.instructionList.writePicardFirstChildComment(n.Data)
// 	}

// 	return fmt.Errorf("writePicardFirstChildNode unknown node type %v", n.Type)

// }
