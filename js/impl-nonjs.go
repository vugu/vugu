//go:build !js || !wasm

package js

import (
	"errors"
)

var errNotImpl = errors.New("js not implemented")

// ######################
// # syscall/js func.go #
// ######################

// Func is a wrapped Go function to be called by JavaScript.
//
// This is a syscall/js placeholder.
type Func struct {
	Value // the JavaScript function that invokes the Go function
}

// FuncOf returns a function to be used by JavaScript.
//
// The Go function fn is called with the value of JavaScript's "this" keyword and the
// arguments of the invocation. The return value of the invocation is
// the result of the Go function mapped back to JavaScript according to ValueOf.
//
// Invoking the wrapped Go function from JavaScript will
// pause the event loop and spawn a new goroutine.
// Other wrapped functions which are triggered during a call from Go to JavaScript
// get executed on the same goroutine.
//
// As a consequence, if one wrapped function blocks, JavaScript's event loop
// is blocked until that function returns. Hence, calling any async JavaScript
// API, which requires the event loop, like fetch (http.Client), will cause an
// immediate deadlock. Therefore a blocking function should explicitly start a
// new goroutine.
//
// Func.Release must be called to free up resources when the function will not be invoked any more.
//
// This is a syscall/js placeholder.
func FuncOf(fn func(this Value, args []Value) any) Func {
	panic(errNotImpl)
}

// Release frees up resources allocated for the function.
// The function must not be invoked after calling Release.
// It is allowed to call Release while the function is still running.
//
// This is a syscall/js placeholder.
func (c Func) Release() {}

// ####################
// # syscall/js js.go #
// ####################

// Value represents a JavaScript value. The zero value is the JavaScript value "undefined".
// Values can be checked for equality with the Equal method.
//
// This is a syscall/js placeholder.
type Value struct {
	_ [0]func() // uncomparable; to make == not compile
}

// Error wraps a JavaScript error.
//
// This is a syscall/js placeholder.
type Error struct {
	Value
}

// Error implements the error interface.
//
// This is a syscall/js placeholder.
func (e Error) Error() string {
	return errNotImpl.Error()
}

// Equal reports whether v and w are equal according to JavaScript's === operator.
//
// This is a syscall/js placeholder.
func (v Value) Equal(w Value) bool {
	return true
}

// Undefined returns the JavaScript value "undefined".
//
// This is a syscall/js placeholder.
func Undefined() Value {
	return Value{}
}

// IsUndefined reports whether v is the JavaScript value "undefined".
//
// This is a syscall/js placeholder.
func (v Value) IsUndefined() bool {
	return true
}

// Null returns the JavaScript value "null".
//
// This is a syscall/js placeholder.
func Null() Value {
	return Value{}
}

// Null returns the JavaScript value "null".
//
// This is a syscall/js placeholder.
func (v Value) IsNull() bool {
	return false
}

// IsNaN reports whether v is the JavaScript value "NaN".
//
// This is a syscall/js placeholder.
func (v Value) IsNaN() bool {
	return false
}

// Global returns the JavaScript global object, usually "window" or "global".
//
// This is a syscall/js placeholder.
func Global() Value {
	return Value{}
}

// ValueOf returns x as a JavaScript value:
//
//	| Go                     | JavaScript             |
//	| ---------------------- | ---------------------- |
//	| js.Value               | [its value]            |
//	| js.Func                | function               |
//	| nil                    | null                   |
//	| bool                   | boolean                |
//	| integers and floats    | number                 |
//	| string                 | string                 |
//	| []interface{}          | new array              |
//	| map[string]interface{} | new object             |
//
// Panics if x is not one of the expected types.
//
// This is a syscall/js placeholder.
func ValueOf(x any) Value {
	return Value{}
}

// Type represents the JavaScript type of a Value.
//
// This is a syscall/js placeholder.
type Type int

// This is a syscall/js placeholder.
const (
	TypeUndefined = iota
	TypeNull
	TypeBoolean
	TypeNumber
	TypeString
	TypeSymbol
	TypeObject
	TypeFunction
)

// This is a syscall/js placeholder.
func (t Type) String() string {
	return "undefined"
}

// Type returns the JavaScript type of the value v. It is similar to JavaScript's typeof operator,
// except that it returns TypeNull instead of TypeObject for null.
//
// This is a syscall/js placeholder.
func (v Value) Type() Type {
	return TypeUndefined
}

// Get returns the JavaScript property p of value v.
// It panics if v is not a JavaScript object.
//
// This is a syscall/js placeholder.
func (v Value) Get(p string) Value {
	panic(errNotImpl)
}

// Set sets the JavaScript property p of value v to ValueOf(x).
// It panics if v is not a JavaScript object.
//
// This is a syscall/js placeholder.
func (v Value) Set(p string, x any) {
	panic(errNotImpl)
}

// Delete deletes the JavaScript property p of value v.
// It panics if v is not a JavaScript object.
//
// This is a syscall/js placeholder.
func (v Value) Delete(p string) {
	panic(errNotImpl)
}

// Index returns JavaScript index i of value v.
// It panics if v is not a JavaScript object.
//
// This is a syscall/js placeholder.
func (v Value) Index(i int) Value {
	panic(errNotImpl)
}

// SetIndex sets the JavaScript index i of value v to ValueOf(x).
// It panics if v is not a JavaScript object.
//
// This is a syscall/js placeholder.
func (v Value) SetIndex(i int, x any) {
	panic(errNotImpl)
}

// Length returns the JavaScript property "length" of v.
// It panics if v is not a JavaScript object.
//
// This is a syscall/js placeholder.
func (v Value) Length() int {
	panic(errNotImpl)
}

// Call does a JavaScript call to the method m of value v with the given arguments.
// It panics if v has no method m.
// The arguments get mapped to JavaScript values according to the ValueOf function.
//
// This is a syscall/js placeholder.
func (v Value) Call(m string, args ...any) Value {
	panic(errNotImpl)
}

// Invoke does a JavaScript call of the value v with the given arguments.
// It panics if v is not a JavaScript function.
// The arguments get mapped to JavaScript values according to the ValueOf function.
//
// This is a syscall/js placeholder.
func (v Value) Invoke(args ...any) Value {
	panic(errNotImpl)
}

// New uses JavaScript's "new" operator with value v as constructor and the given arguments.
// It panics if v is not a JavaScript function.
// The arguments get mapped to JavaScript values according to the ValueOf function.
//
// This is a syscall/js placeholder.
func (v Value) New(args ...any) Value {
	panic(errNotImpl)
}

// Float returns the value v as a float64.
// It panics if v is not a JavaScript number.
//
// This is a syscall/js placeholder.
func (v Value) Float() float64 {
	panic(errNotImpl)
}

// Int returns the value v truncated to an int.
// It panics if v is not a JavaScript number.
//
// This is a syscall/js placeholder.
func (v Value) Int() int {
	panic(errNotImpl)
}

// Bool returns the value v as a bool.
// It panics if v is not a JavaScript boolean.
//
// This is a syscall/js placeholder.
func (v Value) Bool() bool {
	panic(errNotImpl)
}

// Truthy returns the JavaScript "truthiness" of the value v. In JavaScript,
// false, 0, "", null, undefined, and NaN are "falsy", and everything else is
// "truthy". See https://developer.mozilla.org/en-US/docs/Glossary/Truthy.
//
// This is a syscall/js placeholder.
func (v Value) Truthy() bool {
	return false
}

// String returns the value v as a string.
// String is a special case because of Go's String method convention. Unlike the other getters,
// it does not panic if v's Type is not TypeString. Instead, it returns a string of the form "<T>"
// or "<T: V>" where T is v's type and V is a string representation of v's value.
//
// This is a syscall/js placeholder.
func (v Value) String() string {
	return "<undefined>"
}

// InstanceOf reports whether v is an instance of type t according to JavaScript's instanceof operator.
//
// This is a syscall/js placeholder.
func (v Value) InstanceOf(t Value) bool {
	return false
}

// A ValueError occurs when a Value method is invoked on
// a Value that does not support it. Such cases are documented
// in the description of each method.
//
// This is a syscall/js placeholder.
type ValueError struct {
	Method string
	Type   Type
}

// This is a syscall/js placeholder.
func (e *ValueError) Error() string {
	return errNotImpl.Error()
}

// CopyBytesToGo copies bytes from src to dst.
// It panics if src is not a Uint8Array or Uint8ClampedArray.
// It returns the number of bytes copied, which will be the minimum of the lengths of src and dst.
//
// This is a syscall/js placeholder.
func CopyBytesToGo(dst []byte, src Value) int {
	panic(errNotImpl)
}

// CopyBytesToJS copies bytes from src to dst.
// It panics if dst is not a Uint8Array or Uint8ClampedArray.
// It returns the number of bytes copied, which will be the minimum of the lengths of src and dst.
//
// This is a syscall/js placeholder.
func CopyBytesToJS(dst Value, src []byte) int {
	panic(errNotImpl)
}
