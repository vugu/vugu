package vugu

import js "syscall/js"

// NOTE: wasm implementation of events implements conveniences so components can do
// common event handling tasks without importing syscall/js and make it easier to implement
// server-side rendering.

var DOMEventStub = &DOMEvent{} // used as a value to mean "replace this with the actual event that came in"

type DOMEvent struct {
	jsEvent     js.Value
	jsEventThis js.Value
}

func (e *DOMEvent) JSEvent() js.Value {
	return e.jsEvent
}

func (e *DOMEvent) JSEventThis() js.Value {
	return e.jsEventThis
}

func (e *DOMEvent) PreventDefault() {
	e.jsEvent.Call("preventDefault")
}
