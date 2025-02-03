package domrender

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/vugu/vjson"

	"github.com/vugu/vugu"
)

//go:generate go run renderer-js-script-maker.go

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
