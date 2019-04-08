package vugu

import (
	// js "syscall/js"
	js "github.com/vugu/vugu/js"
)

// NOTE: wasm implementation of events implements conveniences so components can do
// common event handling tasks without importing syscall/js and make it easier to implement
// server-side rendering.

var DOMEventStub = &DOMEvent{} // used as a value to mean "replace this with the actual event that came in"

type DOMEvent struct {
	jsEvent     js.Value
	jsEventThis js.Value
	eventEnv    *eventEnv
}

func (e *DOMEvent) JSEvent() js.Value {
	return e.jsEvent
}

func (e *DOMEvent) JSEventThis() js.Value {
	return e.jsEventThis
}

func (e *DOMEvent) EventEnv() EventEnv {
	return e.eventEnv
}

// // RequestRender tells the environment that it should re-render.
// // Unlike most other methods, this can be called from any goroutine
// // at any time, e.g. after HTTP requests have completed in a separate
// // goroutine.
// func (e *DOMEvent) RequestRender() {
// 	// TODO: EventEnv
// 	// TODO: probably we should just do render as a read lock and event handlers as a write lock,
// 	// and expose both read and write locking via EventEnv.  Doc is goin gto need to be really good
// 	// at showing good exampels of how to use this, because JS people migrating to Go and figuring it out on their own
// 	// is going to mean a lot of deadlocks...
// 	panic("HM< HOW DO WE KNOW A RENDER ISNT ALREDY IN PROGRESS?? progrmw need a way to lock before modifying data from background goroutine")
// }

// PreventDefault calls preventDefault() on the underlying DOM event.
// May only be used within event handler in same goroutine.
func (e *DOMEvent) PreventDefault() {
	e.jsEvent.Call("preventDefault")
}
