package vugu

import (
	"fmt"
	"log"
	"time"

	js "github.com/vugu/vugu/js"
)

//go:generate go run renderer-js-script-maker.go

// NewJSRenderer will create a new JSRenderer with the speicifc mount parent selector.
// If an empty string is passed then the root component should include a top level <html> tag
// and the entire page will be rendered.
func NewJSRenderer(mountParentSelector string) (*JSRenderer, error) {

	ret := &JSRenderer{
		MountParentSelector: mountParentSelector,
	}

	ret.domEventCB = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return jsEnv.handleRawDOMEvent(this, args)
	})

	ret.instructionBuffer = make([]byte, 4096)
	ret.instructionBuffer[0] = 3
	ret.instructionTypedArray = js.TypedArrayOf(ret.instructionBuffer)

	ret.window = js.Global().Get("window")
	// log.Printf("ret.window: %#v", ret.window)
	ret.window.Call("eval", jsHelperScript)
	// log.Printf("eval: %#v", ret.window.Get("eval"))

	return ret, nil
}

// JSRenderer implements Renderer against the browser's DOM.
type JSRenderer struct {
	MountParentSelector string

	domEventCB js.Func // the callback function for DOM events

	instructionBuffer     []byte
	instructionTypedArray js.TypedArray

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

	startTime := time.Now()
	il := newInstructionList(r.instructionBuffer)
	// for i := 0; i < 10; i++ {
	il.writeClearRefmap()
	il.writeSetHTMLRef(99)
	il.writeSelectRef(99)
	// il.writeSetAttrStr("lang", "en-gb")
	il.writeSetAttrStr("whatever", "yes it is")
	// }
	il.writeEnd()
	log.Printf("instruction write time: %v", time.Since(startTime))

	r.window.Call("vuguRender", r.instructionTypedArray)
	// r.window.Call("vuguRender")

	// log.Printf("at pos 6: %#v", r.instructionBuffer[6])

	panic(fmt.Errorf("not yet implemented"))
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
