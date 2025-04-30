package vugu

import (
	"syscall/js"
)

// JSValueHandler does something with a js.Value
type JSValueHandler interface {
	JSValueHandle(js.Value)
}

// JSValueFunc implements JSValueHandler as a function.
type JSValueFunc func(js.Value)

// JSValueHandle implements the JSValueHandler interface.
func (f JSValueFunc) JSValueHandle(v js.Value) { f(v) }
