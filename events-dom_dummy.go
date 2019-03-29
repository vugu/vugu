// +build !wasm

package vugu

// NOTE: Dummy implementation allows code that does JS things to compile in non-wasm,
// even though that code will never get called - needed so we can do server side rendering of components.

// DOMEventStub is used as a value to mean "replace this with the actual event that came in".
// In .vugu files, "event" points to this.
var DOMEventStub = &DOMEvent{}

// DOMEvent is an event originated in the browser.  It wraps the JS event that comes in.
// It is meant to be used in WebAssembly but some methods exist here so code can compile
// server-side as well (although DOMEvents should not ever be generated server-side).
type DOMEvent struct {
}

// JSEvent this returns a js.Value in wasm that corresponds to the event object.
// Non-wasm implementation returns nil.
func (e *DOMEvent) JSEvent() interface{} {
	return nil
}

// JSEventThis returns a js.Value in wasm that corresponds to the "this" variable from the event callback.
// It is the DOM element the event was attached to.
// Non-wasm implementation returns nil.
func (e *DOMEvent) JSEventThis() interface{} {
	return nil
}

// RequestRender tells the environment it should be re-rendered as soon as possible (as soon as it can obtain a read lock).
// Non-wasm implementation is nop.
func (e *DOMEvent) RequestRender() {
}

// PreventDefault calls preventDefault() on the JS event object.
// Non-wasm implementation is nop.
func (e *DOMEvent) PreventDefault() {
}

// EventEnv returns the EventEnv for the current environment and allows locking and unlocking around modifications.
// See EventEnv struct.
// Non-wasm implementation returns nil.
func (e *DOMEvent) EventEnv() EventEnv {
	return nil // do we need to do anything with this? in static html rendering the events are never called... right?
}
