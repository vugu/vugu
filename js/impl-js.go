//go:build js && wasm

package js

// IMPLEMENTATION NOTE: (7 Apr 2019, bgp)
// Right now (as of Go 1.12), syscall/js essentially interacts with a big map of values
// on the JS side of things in the browser.  "Value" is alias for "ref" which is basically
// a map key with some bit swizzling to stuff it into a float64, presumably so it's both
// easy to send back and forth between wasm and browser and also to encode info about the
// type in there so things like Truthy() and Type(), etc. can be answered without
// having to reach back out to the browser.  As such, Values and refs and everything else
// don't appear to have any internal pointers or memory concerns at least on the Go side.
// So the approach here is we just alias syscall/js.Value in each appropriate place and then
// delegate all the calls using either casting or unsafe pointers to treat our js.Value exactly like
// a syscall/js.Value.  (FuncOf and friends handled slightly differently.)
// This works with the current implementation.  However if significant
// changes are made to syscall/js (probable), it could very well break this package.
// Since the future of Go-Wasm integration is unclear at this point, I'm just rolling
// forward with the approach above and if 1.13 or later breaks things, I'll just have to
// deal with it then.  If Wasm gains standardized support for GC and DOM references,
// (see https://github.com/WebAssembly/proposals/issues/16) the odds are this package would
// end up just being completely rewritten to use that API+instruction set.
// It's also possible a WASI interface could address this functionality and/or make this entire package
// go away, can't really tell; my crystal ball is broken.
// See: https://hacks.mozilla.org/2019/03/standardizing-wasi-a-webassembly-system-interface/

import (
	sjs "syscall/js"
)

// ######################
// # syscall/js func.go #
// ######################

// Func is a wrapped Go function to be called by JavaScript.
//
// This is a syscall/js alias.
type Func struct {
	Value          // A copy of the syscall/js Value inside of f, casted into Value.
	f     sjs.Func // The original syscall/js Func object.
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
// This is a syscall/js alias.
func FuncOf(fn func(this Value, args []Value) any) Func {
	fnConv := func(this sjs.Value, args []sjs.Value) any {
		argsConv := make([]Value, len(args))
		for i, arg := range args {
			argsConv[i] = Value(arg)
		}
		return convertAnyToSJS(fn(Value(this), argsConv))
	}

	f := sjs.FuncOf(fnConv)

	return Func{
		Value: Value(f.Value),
		f:     f,
	}
}

// Release frees up resources allocated for the function.
// The function must not be invoked after calling Release.
// It is allowed to call Release while the function is still running.
//
// This is a syscall/js alias.
func (c Func) Release() {
	c.f.Release()
}

// ####################
// # syscall/js js.go #
// ####################

// Value represents a JavaScript value. The zero value is the JavaScript value "undefined".
// Values can be checked for equality with the Equal method.
//
// This is a syscall/js alias.
type Value sjs.Value

// Error wraps a JavaScript error.
//
// This is a syscall/js alias.
type Error = sjs.Error

// Error implements the error interface.
//
// This is a syscall/js alias.
// func (e Error) Error() string

// Equal reports whether v and w are equal according to JavaScript's === operator.
//
// This is a syscall/js alias.
func (v Value) Equal(w Value) bool {
	return sjs.Value(v).Equal(sjs.Value(w))
}

// Undefined returns the JavaScript value "undefined".
//
// This is a syscall/js alias.
func Undefined() Value {
	return Value(sjs.Undefined())
}

// IsUndefined reports whether v is the JavaScript value "undefined".
//
// This is a syscall/js alias.
func (v Value) IsUndefined() bool {
	return sjs.Value(v).IsUndefined()
}

// Null returns the JavaScript value "null".
//
// This is a syscall/js alias.
func Null() Value {
	return Value(sjs.Null())
}

// Null returns the JavaScript value "null".
//
// This is a syscall/js alias.
func (v Value) IsNull() bool {
	return sjs.Value(v).IsNull()
}

// IsNaN reports whether v is the JavaScript value "NaN".
//
// This is a syscall/js alias.
func (v Value) IsNaN() bool {
	return sjs.Value(v).IsNaN()
}

// Global returns the JavaScript global object, usually "window" or "global".
//
// This is a syscall/js alias.
func Global() Value {
	return Value(sjs.Global())
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
// This is a syscall/js alias.
func ValueOf(x any) Value {
	switch x := x.(type) {
	case []any:
		return Value(sjs.ValueOf(convertSliceToSJS(x)))

	case map[string]any:
		return Value(sjs.ValueOf(convertMapToSJS(x)))

	default:
		return Value(sjs.ValueOf(convertAnyToSJS(x)))

	}
}

// Type represents the JavaScript type of a Value.
//
// This is a syscall/js alias.
type Type = sjs.Type

// This is a syscall/js alias.
const (
	TypeUndefined = Type(sjs.TypeUndefined)
	TypeNull      = Type(sjs.TypeNull)
	TypeBoolean   = Type(sjs.TypeBoolean)
	TypeNumber    = Type(sjs.TypeNumber)
	TypeString    = Type(sjs.TypeString)
	TypeSymbol    = Type(sjs.TypeSymbol)
	TypeObject    = Type(sjs.TypeObject)
	TypeFunction  = Type(sjs.TypeFunction)
)

// This is a syscall/js alias.
// func (t Type) String() string

// Type returns the JavaScript type of the value v. It is similar to JavaScript's typeof operator,
// except that it returns TypeNull instead of TypeObject for null.
//
// This is a syscall/js alias.
func (v Value) Type() Type {
	return Type(sjs.Value(v).Type())
}

// Get returns the JavaScript property p of value v.
// It panics if v is not a JavaScript object.
//
// This is a syscall/js alias.
func (v Value) Get(p string) Value {
	return Value(sjs.Value(v).Get(p))
}

// Set sets the JavaScript property p of value v to ValueOf(x).
// It panics if v is not a JavaScript object.
//
// This is a syscall/js alias.
func (v Value) Set(p string, x any) {
	sjs.Value(v).Set(p, convertAnyToSJS(x))
}

// Delete deletes the JavaScript property p of value v.
// It panics if v is not a JavaScript object.
//
// This is a syscall/js alias.
func (v Value) Delete(p string) {
	sjs.Value(v).Delete(p)
}

// Index returns JavaScript index i of value v.
// It panics if v is not a JavaScript object.
//
// This is a syscall/js alias.
func (v Value) Index(i int) Value {
	return Value(sjs.Value(v).Index(i))
}

// SetIndex sets the JavaScript index i of value v to ValueOf(x).
// It panics if v is not a JavaScript object.
//
// This is a syscall/js alias.
func (v Value) SetIndex(i int, x any) {
	sjs.Value(v).SetIndex(i, convertAnyToSJS(x))
}

// Length returns the JavaScript property "length" of v.
// It panics if v is not a JavaScript object.
//
// This is a syscall/js alias.
func (v Value) Length() int {
	return sjs.Value(v).Length()
}

// Call does a JavaScript call to the method m of value v with the given arguments.
// It panics if v has no method m.
// The arguments get mapped to JavaScript values according to the ValueOf function.
//
// This is a syscall/js alias.
func (v Value) Call(m string, args ...any) Value {
	return Value(sjs.Value(v).Call(m, convertSliceToSJS(args)...))
}

// Invoke does a JavaScript call of the value v with the given arguments.
// It panics if v is not a JavaScript function.
// The arguments get mapped to JavaScript values according to the ValueOf function.
//
// This is a syscall/js alias.
func (v Value) Invoke(args ...any) Value {
	return Value(sjs.Value(v).Invoke(convertSliceToSJS(args)...))
}

// New uses JavaScript's "new" operator with value v as constructor and the given arguments.
// It panics if v is not a JavaScript function.
// The arguments get mapped to JavaScript values according to the ValueOf function.
//
// This is a syscall/js alias.
func (v Value) New(args ...any) Value {
	return Value(sjs.Value(v).New(convertSliceToSJS(args)...))
}

// Float returns the value v as a float64.
// It panics if v is not a JavaScript number.
//
// This is a syscall/js alias.
func (v Value) Float() float64 {
	return sjs.Value(v).Float()
}

// Int returns the value v truncated to an int.
// It panics if v is not a JavaScript number.
//
// This is a syscall/js alias.
func (v Value) Int() int {
	return sjs.Value(v).Int()
}

// Bool returns the value v as a bool.
// It panics if v is not a JavaScript boolean.
//
// This is a syscall/js alias.
func (v Value) Bool() bool {
	return sjs.Value(v).Bool()
}

// Truthy returns the JavaScript "truthiness" of the value v. In JavaScript,
// false, 0, "", null, undefined, and NaN are "falsy", and everything else is
// "truthy". See https://developer.mozilla.org/en-US/docs/Glossary/Truthy.
//
// This is a syscall/js alias.
func (v Value) Truthy() bool {
	return sjs.Value(v).Truthy()
}

// String returns the value v as a string.
// String is a special case because of Go's String method convention. Unlike the other getters,
// it does not panic if v's Type is not TypeString. Instead, it returns a string of the form "<T>"
// or "<T: V>" where T is v's type and V is a string representation of v's value.
//
// This is a syscall/js alias.
func (v Value) String() string {
	return sjs.Value(v).String()
}

// InstanceOf reports whether v is an instance of type t according to JavaScript's instanceof operator.
//
// This is a syscall/js alias.
func (v Value) InstanceOf(t Value) bool {
	return sjs.Value(v).InstanceOf(sjs.Value(t))
}

// A ValueError occurs when a Value method is invoked on
// a Value that does not support it. Such cases are documented
// in the description of each method.
//
// This is a syscall/js alias.
type ValueError = sjs.ValueError

// This is a syscall/js alias.
// func (e *ValueError) Error() string

// CopyBytesToGo copies bytes from src to dst.
// It panics if src is not a Uint8Array or Uint8ClampedArray.
// It returns the number of bytes copied, which will be the minimum of the lengths of src and dst.
//
// This is a syscall/js alias.
func CopyBytesToGo(dst []byte, src Value) int {
	return sjs.CopyBytesToGo(dst, sjs.Value(src))
}

// CopyBytesToJS copies bytes from src to dst.
// It panics if dst is not a Uint8Array or Uint8ClampedArray.
// It returns the number of bytes copied, which will be the minimum of the lengths of src and dst.
//
// This is a syscall/js alias.
func CopyBytesToJS(dst Value, src []byte) int {
	return sjs.CopyBytesToJS(sjs.Value(dst), src)
}
