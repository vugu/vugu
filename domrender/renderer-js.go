package domrender

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/vugu/vjson"

	"github.com/vugu/vugu"

	js "github.com/vugu/vugu/js"
)

//go:generate go run renderer-js-script-maker.go

// NewJSRenderer is an alias for New.
//
// Deprecated: Use New instead.
func NewJSRenderer(mountPointSelector string) (*JSRenderer, error) {
	return New(mountPointSelector)
}

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

type jsRenderState struct {
	// stores positionID to slice of DOMEventHandlerSpec
	domHandlerMap map[string][]vugu.DOMEventHandlerSpec

	// callback stuff is handled by callbackManager
	callbackManager callbackManager
}

func newJsRenderState() *jsRenderState {
	return &jsRenderState{
		domHandlerMap: make(map[string][]vugu.DOMEventHandlerSpec, 8),
	}
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

type lifecycleState struct {
	passNum uint8
}

// EventEnv returns an EventEnv that can be used for synchronizing updates.
func (r *JSRenderer) EventEnv() vugu.EventEnv {
	return r.eventEnv
}

// Release calls release on any resources that this renderer allocated.
func (r *JSRenderer) Release() {
	// NOTE: seems sensible to leave this here in case we do need something to be released, better than
	// omitting it and people getting used to no release being needed and then requiring it later.
	// r.instructionTypedArray.Release()
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

// var window js.Value

// func init() {
// 	window = js.Global().Get("window")
// 	if window.Truthy() {
// 		js.Global().Call("eval", jsHelperScript)
// 	}
// }

func (r *JSRenderer) visitFirst(state *jsRenderState, bo *vugu.BuildOut, br *vugu.BuildResults, n *vugu.VGNode, positionID []byte) error {

	// log.Printf("TODO: We need to go through and optimize away unneeded calls to create elements, set attributes, set event handlers, etc. for cases where they are the same per hash")

	// log.Printf("JSRenderer.visitFirst")

	if n.Type != vugu.ElementNode {
		return errors.New("root of component must be element")
	}

	err := r.instructionList.writeClearEl()
	if err != nil {
		return err
	}

	// first tag is html
	if strings.ToLower(n.Data) == "html" {

		err := r.syncHtml(state, n, []byte("html"))
		if err != nil {
			return err
		}

		for nchild := n.FirstChild; nchild != nil; nchild = nchild.NextSibling {

			if strings.ToLower(nchild.Data) == "head" {

				err := r.visitHead(state, bo, br, nchild, []byte("head"))
				if err != nil {
					return err
				}

			} else if strings.ToLower(nchild.Data) == "body" {

				err := r.visitBody(state, bo, br, nchild, []byte("body"))
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
	return r.visitMount(state, bo, br, n, positionID)

}

func (r *JSRenderer) syncHtml(state *jsRenderState, n *vugu.VGNode, positionID []byte) error {
	err := r.instructionList.writeSelectQuery("html")
	if err != nil {
		return err
	}
	return r.syncElement(state, n, positionID)
}

func (r *JSRenderer) visitHead(state *jsRenderState, bo *vugu.BuildOut, br *vugu.BuildResults, n *vugu.VGNode, positionID []byte) error {

	err := r.instructionList.writeSelectQuery("head")
	if err != nil {
		return err
	}
	err = r.syncElement(state, n, positionID)
	if err != nil {
		return err
	}

	return nil
}

func (r *JSRenderer) visitBody(state *jsRenderState, bo *vugu.BuildOut, br *vugu.BuildResults, n *vugu.VGNode, positionID []byte) error {

	err := r.instructionList.writeSelectQuery("body")
	if err != nil {
		return err
	}
	err = r.syncElement(state, n, positionID)
	if err != nil {
		return err
	}

	if !(n.FirstChild != nil && n.FirstChild.NextSibling == nil) {
		return errors.New("body tag must contain exactly one element child")
	}

	return r.visitMount(state, bo, br, n.FirstChild, positionID)
}

func (r *JSRenderer) visitMount(state *jsRenderState, bo *vugu.BuildOut, br *vugu.BuildResults, n *vugu.VGNode, positionID []byte) error {

	// log.Printf("visitMount got here")

	err := r.instructionList.writeSelectMountPoint(r.MountPointSelector, n.Data)
	if err != nil {
		return err
	}

	return r.visitSyncElementEtc(state, bo, br, n, positionID)

}

func (r *JSRenderer) visitSyncNode(state *jsRenderState, bo *vugu.BuildOut, br *vugu.BuildResults, n *vugu.VGNode, positionID []byte) error {

	// log.Printf("visitSyncNode")

	var err error

	// check for Component, in which case we descend into it instead of processing like a regular node
	if n.Component != nil {
		compBuildOut := br.ResultFor(n.Component)
		if len(compBuildOut.Out) != 1 {
			return fmt.Errorf("component %#v expected exactly one Out element but got %d instead",
				n.Component, len(compBuildOut.Out))
		}
		return r.visitSyncNode(state, compBuildOut, br, compBuildOut.Out[0], positionID)
	}

	// check for template (used by vg-template and vg-slot) in which case we process the children directly and ignore n
	if n.IsTemplate() {

		childIndex := 1
		for nchild := n.FirstChild; nchild != nil; nchild = nchild.NextSibling {

			// use a different character here for the position to ensure it's unique
			childPositionID := append(positionID, []byte(fmt.Sprintf("_t_%d", childIndex))...)

			err = r.visitSyncNode(state, bo, br, nchild, childPositionID)
			if err != nil {
				return err
			}

			// if there are more children, advance to the next
			if nchild.NextSibling != nil {
				err = r.instructionList.writeMoveToNextSibling()
				if err != nil {
					return err
				}

			}

			childIndex++
		}

		// element is fully handled
		return nil
	}

	switch n.Type {
	case vugu.ElementNode:
		// check if this element has a namespace set
		if ns := namespaceToURI(n.Namespace); ns != "" {
			err = r.instructionList.writeSetElementNS(n.Data, ns)
		} else {
			err = r.instructionList.writeSetElement(n.Data)
		}
		if err != nil {
			return err
		}
	case vugu.TextNode:
		return r.instructionList.writeSetText(n.Data) // no children possible, just return
	case vugu.CommentNode:
		return r.instructionList.writeSetComment(n.Data) // no children possible, just return
	default:
		return errors.New("unknown node type: " + strconv.Itoa(int(n.Type)))
	}

	// only elements have attributes, child or events
	return r.visitSyncElementEtc(state, bo, br, n, positionID)

}

// visitSyncElementEtc syncs the rest of the stuff that only applies to elements
func (r *JSRenderer) visitSyncElementEtc(state *jsRenderState, bo *vugu.BuildOut, br *vugu.BuildResults, n *vugu.VGNode, positionID []byte) error {

	err := r.syncElement(state, n, positionID)
	if err != nil {
		return err
	}

	if n.InnerHTML != nil {
		return r.instructionList.writeSetInnerHTML(*n.InnerHTML)
	}

	// tell callbackManager about the create and populate functions
	// (if present, otherwise this is a nop and will return 0,0)
	cid, pid := state.callbackManager.addCreateAndPopulateHandlers(n.JSCreateHandler, n.JSPopulateHandler)

	// for vg-js-create, send an instruction to call us back when this element is created
	// (handled by callbackManager)
	if cid != 0 {
		err := r.instructionList.writeCallbackLastElement(cid)
		if err != nil {
			return err
		}
	}

	if n.FirstChild != nil {

		err = r.instructionList.writeMoveToFirstChild()
		if err != nil {
			return err
		}

		childIndex := 1
		for nchild := n.FirstChild; nchild != nil; nchild = nchild.NextSibling {

			childPositionID := append(positionID, []byte(fmt.Sprintf("_%d", childIndex))...)

			err = r.visitSyncNode(state, bo, br, nchild, childPositionID)
			if err != nil {
				return err
			}
			// log.Printf("GOT HERE X: %#v", n)
			err = r.instructionList.writeMoveToNextSibling()
			if err != nil {
				return err
			}
			childIndex++
		}

		err = r.instructionList.writeMoveToParent()
		if err != nil {
			return err
		}
	}

	// for vg-js-populate, send an instruction to call us back again with the populate flag for this same one
	// (handled by callbackManager)
	if pid != 0 {
		err := r.instructionList.writeCallback(pid)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *JSRenderer) syncElement(state *jsRenderState, n *vugu.VGNode, positionID []byte) error {
	if namespaceToURI(n.Namespace) != "" {
		for _, a := range n.Attr {
			ns := namespaceToURI(a.Namespace)
			// FIXME: we skip Namespace="" && Key = "xmlns" here, because this WILL cause an js exception
			// the correct way would be, to parse the xmlns attribute in the generator, set the namespace of the holding element
			// and then forget about this attribute
			if ns == "" && a.Key == "xmlns" {
				continue
			}
			err := r.instructionList.writeSetAttrNSStr(ns, a.Key, a.Val)
			if err != nil {
				return err
			}
		}
	} else {
		for _, a := range n.Attr {
			err := r.instructionList.writeSetAttrStr(a.Key, a.Val)
			if err != nil {
				return err
			}
		}
	}

	err := r.instructionList.writeRemoveOtherAttrs()
	if err != nil {
		return err
	}

	// do any JS properties
	for _, p := range n.Prop {
		err := r.instructionList.writeSetProperty(p.Key, []byte(p.JSONVal))
		if err != nil {
			return err
		}
	}

	if len(n.DOMEventHandlerSpecList) > 0 {

		// store in domHandlerMap
		state.domHandlerMap[string(positionID)] = n.DOMEventHandlerSpecList

		for _, hs := range n.DOMEventHandlerSpecList {
			err := r.instructionList.writeSetEventListener(positionID, hs.EventType, hs.Capture, hs.Passive)
			if err != nil {
				return err
			}
		}
	}
	// always write the remove for event listeners so any previous ones are taken away
	return r.instructionList.writeRemoveOtherEventListeners(positionID)
}

// // writeAllStaticAttrs is a helper to write all the static attrs from a VGNode
// func (r *JSRenderer) writeAllStaticAttrs(n *vugu.VGNode) error {
// 	for _, a := range n.Attr {
// 		err := r.instructionList.writeSetAttrStr(a.Key, a.Val)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func (r *JSRenderer) handleCallback(this js.Value, args []js.Value) interface{} {
	return r.jsRenderState.callbackManager.callback(this, args)
}

func (r *JSRenderer) handleDOMEvent() {

	strlen := binary.BigEndian.Uint32(r.eventHandlerBuffer[:4])
	b := r.eventHandlerBuffer[4 : strlen+4]
	// log.Printf("handleDOMEvent JSON from event buffer: %q", b)

	// var ee eventEnv
	// rwmu            *sync.RWMutex
	// requestRenderCH chan bool

	var eventDetail struct {
		PositionID string // `json:"position_id"`
		EventType  string // `json:"event_type"`
		Capture    bool   // `json:"capture"`
		Passive    bool   // `json:"passive"`

		// the event object data as extracted above
		EventSummary map[string]interface{} // `json:"event_summary"`
	}

	edm := make(map[string]interface{}, 6)
	// err := json.Unmarshal(b, &eventDetail)
	err := vjson.Unmarshal(b, &edm)
	if err != nil {
		panic(err)
	}

	// manually extract fields
	eventDetail.PositionID, _ = edm["position_id"].(string)
	eventDetail.EventType, _ = edm["event_type"].(string)
	eventDetail.Capture, _ = edm["capture"].(bool)
	eventDetail.Passive, _ = edm["passive"].(bool)
	eventDetail.EventSummary, _ = edm["event_summary"].(map[string]interface{})

	domEvent := vugu.NewDOMEvent(r.eventEnv, eventDetail.EventSummary)

	// log.Printf("eventDetail: %#v", eventDetail)

	// it is important that we lock around accessing anything that might change (domHandlerMap)
	// and around the invokation of the handler call itself

	r.eventRWMU.Lock()
	handlers := r.jsRenderState.domHandlerMap[eventDetail.PositionID]
	var f func(vugu.DOMEvent)
	for _, h := range handlers {
		if h.EventType == eventDetail.EventType && h.Capture == eventDetail.Capture {
			f = h.Func
			break
		}
	}

	// make sure we found something, panic if not
	if f == nil {
		r.eventRWMU.Unlock()
		panic(fmt.Errorf("Unable to find event handler for positionID=%q, eventType=%q, capture=%v",
			eventDetail.PositionID, eventDetail.EventType, eventDetail.Capture))
	}

	// NOTE: For tinygo support we are not using defer here for now - it would probably be better to do so since
	// the handler can panic.  However, Vugu program behavior after panicing from an event is currently
	// undefined so whatever for now.  We'll have to make a decision later about whether or not Vugu
	// programs should keep running after an event handler panics.  They do in JS after an exception,
	// but... this is not JS.  Needs more thought.

	// invoke handler
	f(domEvent)

	r.eventRWMU.Unlock()

	// TODO: Also give this more thought: For now we just do a non-blocking push to the
	// eventWaitCh, telling the render loop that a render is required, but if a bunch
	// of them stack up we don't wait
	r.sendEventWaitCh()

}
