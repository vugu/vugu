//go:build !js || !wasm

package main

import (
	"fmt"
	"time"

	js "github.com/vugu/vugu/js"
)

// Make a vugu.js.ValueOf call on an object derived from a JS global object - in this case the global "Date" object.
// Returns the string "Ok" or a panic message
// See Issue #328 https://github.com/vugu/vugu/issues/328
func (c *Root) ValueOf() (s string) {
	// We have to be a little clever here
	// We need to recover from the panic but change the value we return.
	// The simplest way to achieve this is with a named return value.
	// By setting the named return value ot the panic string we will overwrite any previous
	// value (in this case "Ok") that we would otherwise return.
	defer func() {
		if r := recover(); r != nil {
			s = fmt.Sprintf("%s", r) // set the names return value to the panic string
		}
	}()
	c.panicingFunc()
	return "Ok"
}

// panicingFunc makes the vugu.js.ValueOf call that rill result in a panic
func (c *Root) panicingFunc() {
	date := js.Global().Get("Date") // NOTE: use of double quotes not single quotes as in Issue #328 https://github.com/vugu/vugu/issues/328
	timeEndValue := date.New(time.Now().UnixMilli())
	// This line with panic with:
	//
	// ValueOf: invalid value
	//
	// The underlying implementation of js.Value() calls syscall.js.Value.
	// The underlying type of the interface{} that is passed to syscall.js.Value is a vugu.js.Value
	// because a vugu.js.Value is the return type of the js.Global.Get(...) call above.
	// But because a vugu.js.Value is not a known type inside syscall.js.ValueOf the result is a panic from syscall.js.ValueOf.
	// The internal implementation of syscall.js.ValueOf consists of a type switch statement i.e. `switch x := x.(type)`
	// and since a vugu.js.Value is not one of the listed types the syscall.js.ValueOf we reach the default case which panics.
	js.ValueOf(timeEndValue)
}
