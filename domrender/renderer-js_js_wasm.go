package domrender

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/vugu/vugu"

	"syscall/js"
)

// New will create a new JSRenderer with the speicifc mount point selector.
// If an empty string is passed then the root component should include a top level <html> tag
// and the entire page will be rendered.
func New(mountPointSelector string) (*JSRenderer, error) {

	ret := &JSRenderer{
		MountPointSelector: mountPointSelector,
	}

	ret.instructionBuffer = make([]byte, 16384)
	// ret.instructionTypedArray = js.TypedArrayOf(ret.instructionBuffer)

	ret.window = js.Global().Get("window")

	ret.window.Call("eval", jsHelperScript)

	ret.instructionBufferJS = ret.window.Call("vuguGetRenderArray")

	ret.instructionList = newInstructionList(ret.instructionBuffer, func(il *instructionList) error {

		// call vuguRender to have the instructions processed in JS
		ret.instructionBuffer[il.pos] = 0 // ensure zero terminator

		// copy the data over
		js.CopyBytesToJS(ret.instructionBufferJS, ret.instructionBuffer)

		// then call vuguRender
		ret.window.Call("vuguRender" /*, ret.instructionBufferJS*/)

		return nil
	})

	// enable debug logging
	// ret.instructionList.logWriter = os.Stdout

	ret.eventHandlerBuffer = make([]byte, 16384)
	// ret.eventHandlerTypedArray = js.TypedArrayOf(ret.eventHandlerBuffer)

	ret.eventHandlerFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			panic(fmt.Errorf("eventHandlerFunc got arg slice not exactly 1 element in length: %#v", args))
		}
		bufferLength := args[0].Length()
		if cap(ret.eventHandlerBuffer) < bufferLength+1 {
			ret.eventHandlerBuffer = make([]byte, bufferLength+1)
		}
		//log.Println(cap(ret.eventHandlerBuffer))
		n := js.CopyBytesToGo(ret.eventHandlerBuffer, args[0])
		if n >= len(ret.eventHandlerBuffer) {
			panic(errors.New("event data is too large, cannot continue, len: " + strconv.Itoa(n)))
		}
		ret.handleDOMEvent() // discard this and args, all data should be in eventHandlerBuffer; avoid using js.Value
		return nil
		// return jsEnv.handleRawDOMEvent(this, args)
	})

	// wire up the event handler func and the array that we used to communicate with instead of js.Value
	// ret.window.Call("vuguSetEventHandlerAndBuffer", ret.eventHandlerFunc, ret.eventHandlerTypedArray)

	// wire up callback handler
	ret.window.Call("vuguSetCallbackHandler", js.FuncOf(ret.handleCallback))

	// wire up the event handler func
	ret.window.Call("vuguSetEventHandler", ret.eventHandlerFunc)

	// log.Printf("ret.window: %#v", ret.window)
	// log.Printf("eval: %#v", ret.window.Get("eval"))

	ret.eventWaitCh = make(chan bool, 64)

	ret.eventEnv = vugu.NewEventEnvImpl(
		&ret.eventRWMU,
		ret.eventWaitCh,
	)

	return ret, nil
}

// JSRenderer implements Renderer against the browser's DOM.
type JSRenderer struct {
	MountPointSelector string

	eventWaitCh chan bool          // events send to this and EventWait receives from it
	eventRWMU   sync.RWMutex       // make sure Render and event handling are not attempted at the same time (not totally sure if this is necessary in terms of the wasm threading model but enforce it with a rwmutex all the same)
	eventEnv    *vugu.EventEnvImpl // our EventEnv implementation that exposes eventRWMU and eventWaitCh to events in a clean way

	eventHandlerFunc   js.Func // the callback function for DOM events
	eventHandlerBuffer []byte
	// eventHandlerTypedArray js.TypedArray

	instructionBuffer   []byte   // our local instruction buffer
	instructionBufferJS js.Value // a Uint8Array on the JS side that we copy into
	instructionList     *instructionList

	window js.Value

	jsRenderState *jsRenderState

	// manages the Rendered lifecycle callback stuff
	lifecycleStateMap map[interface{}]lifecycleState
	lifecyclePassNum  uint8
}

func (r *JSRenderer) render(buildResults *vugu.BuildResults) error {

	bo := buildResults.Out

	if !js.Global().Truthy() {
		return errors.New("js environment not available")
	}

	if bo == nil {
		return errors.New("BuildOut is nil")
	}

	if len(bo.Out) != 1 {
		return errors.New("BuildOut.Out has bad len " + strconv.Itoa(len(bo.Out)))
	}

	if bo.Out[0].Type != vugu.ElementNode {
		return errors.New("BuildOut.Out[0].Type is not vugu.ElementNode: " + strconv.Itoa(int(bo.Out[0].Type)))
	}

	// always make sure we have at least a non-nil render state
	if r.jsRenderState == nil {
		r.jsRenderState = newJsRenderState()
	}

	state := r.jsRenderState

	state.callbackManager.startRender()
	defer state.callbackManager.doneRender()

	// TODO: move this next chunk out to it's own func at least

	visitCSSList := func(cssList []*vugu.VGNode) error {
		// CSS stuff first
		for _, cssEl := range cssList {

			// some basic sanity checking
			if cssEl.Type != vugu.ElementNode || !(cssEl.Data == "style" || cssEl.Data == "link") {
				return errors.New("CSS output must be link or style tag")
			}

			var textBuf bytes.Buffer
			for childN := cssEl.FirstChild; childN != nil; childN = childN.NextSibling {
				if childN.Type != vugu.TextNode {
					return fmt.Errorf("CSS tag must contain only text children, found %v instead: %#v", childN.Type, childN)
				}
				textBuf.WriteString(childN.Data)
			}

			var attrPairs []string
			if len(cssEl.Attr) > 0 {
				attrPairs = make([]string, 0, len(cssEl.Attr)*2)
				for _, attr := range cssEl.Attr {
					attrPairs = append(attrPairs, attr.Key, attr.Val)
				}
			}

			err := r.instructionList.writeSetCSSTag(cssEl.Data, textBuf.Bytes(), attrPairs)
			if err != nil {
				return err
			}
		}

		return nil
	}

	var walkCSSBuildOut func(buildOut *vugu.BuildOut) error
	walkCSSBuildOut = func(buildOut *vugu.BuildOut) error {
		err := visitCSSList(buildOut.CSS)
		if err != nil {
			return err
		}
		for _, c := range buildOut.Components {
			// nextBuildOut := buildResults.AllOut[c]
			nextBuildOut := buildResults.ResultFor(c)
			if nextBuildOut == nil {
				panic(fmt.Errorf("walkCSSBuildOut nextBuildOut was nil for %#v", c))
			}
			err := walkCSSBuildOut(nextBuildOut)
			if err != nil {
				return err
			}
		}
		return nil
	}
	err := walkCSSBuildOut(bo)
	if err != nil {
		return err
	}

	err = r.instructionList.writeRemoveOtherCSSTags()
	if err != nil {
		return err
	}

	// main output
	err = r.visitFirst(state, bo, buildResults, bo.Out[0], []byte("0"))
	if err != nil {
		return err
	}

	// // JS stuff last
	// // log.Printf("TODO: handle JS")

	err = r.instructionList.flush()
	if err != nil {
		return err
	}

	// handle Rendered lifecycle callback
	if r.lifecycleStateMap == nil {
		r.lifecycleStateMap = make(map[interface{}]lifecycleState, len(bo.Components))
	}
	r.lifecyclePassNum++

	var rctx renderedCtx

	for _, c := range bo.Components {

		rctx = renderedCtx{eventEnv: r.eventEnv}

		st, ok := r.lifecycleStateMap[c]
		rctx.first = !ok
		st.passNum = r.lifecyclePassNum

		invokeRendered(c, &rctx)

		r.lifecycleStateMap[c] = st

	}

	// now purge from lifecycleStateMap anything not touched in this pass
	for k, st := range r.lifecycleStateMap {
		if st.passNum != r.lifecyclePassNum {
			delete(r.lifecycleStateMap, k)
		}
	}

	return nil

}

// EventWait blocks until an event has occurred which causes a re-render.
// It returns true if the render loop should continue or false if it should exit.
func (r *JSRenderer) EventWait() (ok bool) {

	// make sure the JS environment is still available, returning false otherwise
	if !js.Global().Truthy() {
		return false
	}

	// FIXME: this should probably have some sort of "debouncing" on it to handle the case of
	// several events in rapid succession causing multiple renders - maybe we read from eventWaitCH
	// continuously until it's empty, with a max of like 20ms pause between each or something, and then
	// only return after we don't see anything for that time frame.

	ok = <-r.eventWaitCh
	return

}

func (r *JSRenderer) handleCallback(this js.Value, args []js.Value) interface{} {
	return r.jsRenderState.callbackManager.callback(this, args)
}
