package vugu

import (
	"sync"

	"github.com/vugu/vugu/js"
)

// DOMEvent is an event originated in the browser.  It wraps the JS event that comes in.
// It is meant to be used in WebAssembly but some methods exist here so code can compile
// server-side as well (although DOMEvents should not ever be generated server-side).
type DOMEvent struct {
	// jsEvent     js.Value
	// jsEventThis js.Value

	eventSummary map[string]interface{}

	eventEnv *eventEnv // eventEnv from the renderer

	window js.Value // sure, why not
}

// EventSummary returns a map with simple properties (primitive types) from the event.
// Accessing values returns by EventSummary incurs no additional performance or memory
// penalty, whereas calls to JSEvent, JSEventTarget, etc. require a call into the browser
// JS engine and the attendant resource usage.  So if you can get the information you
// need from the EventSummary, that's better.
func (e *DOMEvent) EventSummary() map[string]interface{} {
	return e.eventSummary
}

// JSEvent this returns a js.Value in wasm that corresponds to the event object.
// Non-wasm implementation returns nil.
func (e *DOMEvent) JSEvent() js.Value {
	return e.window.Call("vuguGetActiveEvent")
}

// JSEventTarget returns the value of the "target" property of the event, the element
// that the event was originally fired/registered on.
func (e *DOMEvent) JSEventTarget() js.Value {
	return e.window.Call("vuguGetActiveEventTarget")
}

// JSEventCurrentTarget returns the value of the "currentTarget" property of the event, the element
// that is currently processing the event.
func (e *DOMEvent) JSEventCurrentTarget() js.Value {
	return e.window.Call("vuguGetActiveEventCurrentTarget")
}

// EventEnv returns the EventEnv for the current environment and allows locking and unlocking around modifications.
// See EventEnv struct.
func (e *DOMEvent) EventEnv() EventEnv {
	return e.eventEnv
}

// PreventDefault calls preventDefault() on the underlying DOM event.
// May only be used within event handler in same goroutine.
func (e *DOMEvent) PreventDefault() {
	e.window.Call("vuguActiveEventPreventDefault")
}

// StopPropagation calls stopPropagation() on the underlying DOM event.
// May only be used within event handler in same goroutine.
func (e *DOMEvent) StopPropagation() {
	e.window.Call("vuguActiveEventStopPropagation")
}

type DOMEventHandlerSpec struct {
	EventType string // "click", "mouseover", etc.
	Func      func(*DOMEvent)
	Capture   bool
	Passive   bool
}

// // DOMEventHandler is created in BuildVDOM to represent a method call that is performed to handle an event.
// type DOMEventHandler struct {
// 	ReceiverAndMethodHash uint64        // hash value corresponding to the method and receiver, so we get a unique value for each combination of method and receiver
// 	Method                reflect.Value // method to be called, with receiver baked into it if needed (see reflect/Value.MethodByName)
// 	Args                  []interface{} // arguments to be passed in when calling (special case for eventStub)
// }

// func (d DOMEventHandler) hashString() string {
// 	return fmt.Sprintf("%x", d.hash())
// }

// func (d DOMEventHandler) hash() (ret uint64) {

// 	// defer func() {
// 	// 	log.Printf("DOMEventHandler.hash for (receiver_and_method_hash=%v, method=%#v, args=%#v) is returning %v", d.ReceiverAndMethodHash, d.Method, d.Args, ret)
// 	// }()

// 	if !d.Method.IsValid() && len(d.Args) == 0 {
// 		return 0
// 	}

// 	b8 := make([]byte, 8)
// 	h := xxhash.New()

// 	// this may end up being an issue but I need to move through this for now -
// 	// this is only unique for the method itself, not the instance it's tied to, although
// 	// it's probably rare to have to distinguish between multiple component types in use in an application
// 	// fmt.Fprintf(h, "%#v", d.Method)

// 	// NOTE: added ReceiverAndMethodHash to deal with the note above, let's see if it solves the problem
// 	binary.BigEndian.PutUint64(b8, d.ReceiverAndMethodHash)
// 	h.Write(b8)

// 	// add the args
// 	for _, a := range d.Args {
// 		binary.BigEndian.PutUint64(b8, ComputeHash(a))
// 		h.Write(b8)
// 	}
// 	return h.Sum64()
// }

// EventEnv provides locking mechanism to for rendering environment to events so
// data access and rendering can be synchronized and avoid race conditions.
type EventEnv interface {
	Lock()         // acquire write lock
	UnlockOnly()   // release write lock
	UnlockRender() // release write lock and request re-render

	RLock()   // acquire read lock
	RUnlock() // release read lock
}

// eventEnv implements EventEnv
type eventEnv struct {
	rwmu            *sync.RWMutex
	requestRenderCH chan bool
}

// Lock will acquire write lock
func (ee *eventEnv) Lock() {
	// if ee.rwmu == nil {
	// 	return
	// }
	ee.rwmu.Lock()
}

// UnlockOnly will release the write lock
func (ee *eventEnv) UnlockOnly() {
	// if ee.rwmu == nil {
	// 	return
	// }
	ee.rwmu.Unlock()
}

// UnlockRender will release write lock and request re-render
func (ee *eventEnv) UnlockRender() {
	// if ee.rwmu != nil {
	ee.rwmu.Unlock()
	// }
	if ee.requestRenderCH != nil {
		// send non-blocking
		select {
		case ee.requestRenderCH <- true:
		default:
		}
	}
}

// RLock will acquire a read lock
func (ee *eventEnv) RLock() {
	// if ee.rwmu == nil {
	// 	return
	// }
	ee.rwmu.RLock()
}

// RUnlock will release the read lock
func (ee *eventEnv) RUnlock() {
	// if ee.rwmu == nil {
	// 	return
	// }
	ee.rwmu.RUnlock()
}
