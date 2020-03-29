// +build js

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

import sjs "syscall/js"

// import "unsafe"

// Error alias to syscall/js
type Error struct {
	// Value is the underlying JavaScript error value.
	Value
}

// Error alias to syscall/js
func (e Error) Error() string {
	return sjs.Error{Value: sjs.Value(e.Value)}.Error()
	// return (*(*sjs.Error)(unsafe.Pointer(&e))).Error()
}

// Func alias to syscall/js
type Func struct {
	Value
	f sjs.Func // proxy for this func from syscall/js
}

// FuncOf alias to syscall/js
func FuncOf(fn func(Value, []Value) interface{}) Func {

	fn2 := func(this sjs.Value, args []sjs.Value) interface{} {
		args2 := make([]Value, len(args))
		for i := range args {
			args2[i] = Value(args[i])
		}
		return fn(Value(this), args2)
	}

	f := sjs.FuncOf(fn2)

	ret := Func{
		Value: Value(f.Value),
		f:     f,
	}

	return ret
}

// Release alias to syscall/js
func (c Func) Release() {
	// return (*(*sjs.Func)(unsafe.Pointer(&c))).Release()
	c.f.Release()
}

// Type alias to syscall/js
type Type int

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

// String alias to syscall/js
func (t Type) String() string {
	return sjs.Type(t).String()
}

// Undefined alias to syscall/js
func Undefined() Value {
	return Value(sjs.Undefined())
}

// Null alias to syscall/js
func Null() Value {
	return Value(sjs.Null())
}

// Global alias to syscall/js
func Global() Value {
	return Value(sjs.Global())
}

// ValueOf alias to syscall/js
func ValueOf(x interface{}) Value {
	return Value(sjs.ValueOf(x))
}

// CopyBytesToGo alias to syscall/js
func CopyBytesToGo(dst []byte, src Value) int {
	return sjs.CopyBytesToGo(dst, sjs.Value(src))
}

// CopyBytesToJS alias to syscall/js
func CopyBytesToJS(dst Value, src []byte) int {
	return sjs.CopyBytesToJS(sjs.Value(dst), src)
}

// // TypedArray alias to syscall/js
// type TypedArray struct {
// 	Value
// }

// // TypedArrayOf alias to syscall/js
// func TypedArrayOf(slice interface{}) TypedArray {
// 	return TypedArray{Value: Value(sjs.TypedArrayOf(slice).Value)}
// }

// // Release alias to syscall/js
// func (a TypedArray) Release() {
// 	sjs.TypedArray{Value: sjs.Value(a.Value)}.Release()
// }

// ValueError alias to syscall/js
type ValueError struct {
	Method string
	Type   Type
}

// Error alias to syscall/js
func (e *ValueError) Error() string {
	e2 := sjs.ValueError{Method: e.Method, Type: sjs.Type(e.Type)}
	return e2.Error()
}

// Wrapper alias to syscall/js
type Wrapper interface {
	JSValue() Value
}

// Value alias to syscall/js
type Value sjs.Value

// JSValue alias to syscall/js
func (v Value) JSValue() Value {
	return v
}

// Type alias to syscall/js
func (v Value) Type() Type {
	return Type(sjs.Value(v).Type())
}

func (v Value) Get(p string) Value {
	return Value(sjs.Value(v).Get(p))
}

func (v Value) Set(p string, x interface{}) {
	sjs.Value(v).Set(p, x)
}

func (v Value) Index(i int) Value {
	return Value(sjs.Value(v).Index(i))
}

func (v Value) SetIndex(i int, x interface{}) {
	sjs.Value(v).SetIndex(i, x)
}

func (v Value) Length() int {
	return sjs.Value(v).Length()
}

func (v Value) Call(m string, args ...interface{}) Value {
	return Value(sjs.Value(v).Call(m, fixArgsToSjs(args)...))
}

func (v Value) Invoke(args ...interface{}) Value {
	return Value(sjs.Value(v).Invoke(fixArgsToSjs(args)...))
}

func (v Value) New(args ...interface{}) Value {
	return Value(sjs.Value(v).New(fixArgsToSjs(args)...))
}

func (v Value) Float() float64 {
	return sjs.Value(v).Float()
}

func (v Value) Int() int {
	return sjs.Value(v).Int()
}

func (v Value) Bool() bool {
	return sjs.Value(v).Bool()
}

func (v Value) Truthy() bool {
	return sjs.Value(v).Truthy()
}

func (v Value) String() string {
	return sjs.Value(v).String()
}

func (v Value) InstanceOf(t Value) bool {
	return sjs.Value(v).InstanceOf(sjs.Value(t))
}

func (v Value) IsUndefined() bool {
	return sjs.Value(v).IsUndefined()
}

func (v Value) IsNull() bool {
	return sjs.Value(v).IsNull()
}

func fixArgsToSjs(args []interface{}) []interface{} {
	for i := 0; i < len(args); i++ {
		v := args[i]
		if val, ok := v.(Value); ok {
			args[i] = sjs.Value(val) // convert to sjs.Value
		}
		if f, ok := v.(Func); ok {
			args[i] = f.f
		}
		// if ta, ok := v.(TypedArray); ok {
		// 	args[i] = sjs.TypedArray{Value: sjs.Value(ta.Value)}
		// }
	}
	return args
}
