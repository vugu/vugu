//go:build !js || !wasm

package vugu

import (
	"github.com/vugu/vugu/js"
)

// NewDOMEvent returns a new initialized DOMEvent.
func NewDOMEvent(eventEnv EventEnv, eventSummary map[string]interface{}) DOMEvent {
	return &domEvent{
		eventSummary: eventSummary,
		eventEnv:     eventEnv,
		window:       js.Global().Get("window"),
	}
}

// DOMEvent is an event originated in the browser.  It wraps the JS event that comes in.
// It is meant to be used in WebAssembly but some methods exist here so code can compile
// server-side as well (although DOMEvents should not ever be generated server-side).
type DOMEvent interface {

	// Prop returns a value from the EventSummary using the keys you specify.
	// The keys is a list of map keys to be looked up. For example:
	// e.Prop("target", "name") will return the same value as e.EventSummary()["target"]["name"],
	// except that Prop helps with some edge cases and if a value is missing
	// of the wrong type, nil will be returned, instead of panicing.
	Prop(keys ...string) interface{}

	// PropString is like Prop but returns it's value as a string.
	// No type conversion is done, if the requested value is not
	// already a string then an empty string will be returned.
	PropString(keys ...string) string

	// PropFloat64 is like Prop but returns it's value as a float64.
	// No type conversion is done, if the requested value is not
	// already a float64 then float64(0) will be returned.
	PropFloat64(keys ...string) float64

	// PropBool is like Prop but returns it's value as a bool.
	// No type conversion is done, if the requested value is not
	// already a bool then false will be returned.
	PropBool(keys ...string) bool

	// EventSummary returns a map with simple properties (primitive types) from the event.
	// Accessing values returns by EventSummary incurs no additional performance or memory
	// penalty, whereas calls to JSEvent, JSEventTarget, etc. require a call into the browser
	// JS engine and the attendant resource usage.  So if you can get the information you
	// need from the EventSummary, that's better.
	EventSummary() map[string]interface{}

	// JSEvent returns a js.Value in wasm that corresponds to the event object.
	// Non-wasm implementation returns nil.
	JSEvent() js.Value

	// JSEventTarget returns the value of the "target" property of the event, the element
	// that the event was originally fired/registered on.
	JSEventTarget() js.Value

	// JSEventCurrentTarget returns the value of the "currentTarget" property of the event, the element
	// that is currently processing the event.
	JSEventCurrentTarget() js.Value

	// EventEnv returns the EventEnv for the current environment and allows locking and unlocking around modifications.
	// See EventEnv struct.
	EventEnv() EventEnv

	// PreventDefault calls preventDefault() on the underlying DOM event.
	// May only be used within event handler in same goroutine.
	PreventDefault()

	// StopPropagation calls stopPropagation() on the underlying DOM event.
	// May only be used within event handler in same goroutine.
	StopPropagation()
}

// domEvent implements the DOMEvent interface.
// This way DOMEvent can be passed around without a pointer this helps make
// DOM events and component events similar and consistent.
// It might actually makes sense to move this into the domrender package at some point.
type domEvent struct {
	eventSummary map[string]interface{}

	eventEnv EventEnv // from the renderer

	// TODO: yeah but why? this does not need to be here
	window js.Value // sure, why not
}

// JSEvent this returns a js.Value in wasm that corresponds to the event object.
// Non-wasm implementation returns nil.
func (e *domEvent) JSEvent() js.Value {
	return e.window.Call("vuguGetActiveEvent")
}

// JSEventTarget returns the value of the "target" property of the event, the element
// that the event was originally fired/registered on.
func (e *domEvent) JSEventTarget() js.Value {
	return e.window.Call("vuguGetActiveEventTarget")
}

// JSEventCurrentTarget returns the value of the "currentTarget" property of the event, the element
// that is currently processing the event.
func (e *domEvent) JSEventCurrentTarget() js.Value {
	return e.window.Call("vuguGetActiveEventCurrentTarget")
}
