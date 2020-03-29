// +build !js

package js

import "errors"

var errNotImpl = errors.New("js not implemented")

// Error placeholder for syscall/js
type Error struct {
	Value
}

// Error placeholder for syscall/js
func (e Error) Error() string {
	return errNotImpl.Error()
}

// Func placeholder for syscall/js
type Func struct {
	Value
}

// FuncOf placeholder for syscall/js
func FuncOf(fn func(Value, []Value) interface{}) Func {
	// making this panic for now, although I can see some utility in allowing this
	// and then they just never get called, will see...
	panic(errNotImpl)
	// return Func{}
}

// Release placeholder for syscall/js
func (c Func) Release() {
	// nop
}

// Type placeholder for syscall/js
type Type int

// Placeholder for syscall/js
const (
	TypeUndefined Type = iota
	TypeNull
	TypeBoolean
	TypeNumber
	TypeString
	TypeSymbol
	TypeObject
	TypeFunction
)

// String placeholder for syscall/js
func (t Type) String() string {
	return "undefined"
}

// Undefined placeholder for syscall/js
func Undefined() Value {
	return Value{}
}

// Null placeholder for syscall/js
func Null() Value {
	return Value{}
}

// Global placeholder for syscall/js
func Global() Value {
	return Value{}
}

// ValueOf placeholder for syscall/js
func ValueOf(x interface{}) Value {
	return Value{}
}

// CopyBytesToGo placeholder for syscall/js
func CopyBytesToGo(dst []byte, src Value) int {
	return 0
}

// CopyBytesToJS placeholder for syscall/js
func CopyBytesToJS(dst Value, src []byte) int {
	return 0
}

// // TypedArray placeholder for syscall/js
// type TypedArray struct {
// 	Value
// }

// // TypedArrayOf placeholder for syscall/js
// func TypedArrayOf(slice interface{}) TypedArray {
// 	return TypedArray{}
// }

// // Release placeholder for syscall/js
// func (a TypedArray) Release() {
// 	// nop
// }

// ValueError placeholder for syscall/js
type ValueError struct {
	Method string
	Type   Type
}

// Error placeholder for syscall/js
func (e *ValueError) Error() string {
	return errNotImpl.Error()
}

// Wrapper placeholder for syscall/js
type Wrapper interface {
	JSValue() Value
}

// Value placeholder for syscall/js
type Value struct {
	stub uint64
}

// JSValue placeholder for syscall/js
func (v Value) JSValue() Value {
	return v
}

// Type placeholder for syscall/js
func (v Value) Type() Type {
	return TypeUndefined
}

func (v Value) Get(p string) Value {
	return Undefined()
}

func (v Value) Set(p string, x interface{}) {
	panic(errNotImpl)
}

func (v Value) Index(i int) Value {
	return Undefined()
}

func (v Value) SetIndex(i int, x interface{}) {
	panic(errNotImpl)
}

func (v Value) Length() int {
	return 0
}

func (v Value) Call(m string, args ...interface{}) Value {
	panic(errNotImpl)
}

func (v Value) Invoke(args ...interface{}) Value {
	panic(errNotImpl)
}

func (v Value) New(args ...interface{}) Value {
	return Undefined()
}

func (v Value) Float() float64 {
	return 0
}

func (v Value) Int() int {
	return 0
}

func (v Value) Bool() bool {
	return false
}

func (v Value) Truthy() bool {
	return false
}

func (v Value) String() string {
	return ""
}

func (v Value) InstanceOf(t Value) bool {
	return false
}

func (v Value) IsUndefined() bool {
	return true
}

func (v Value) IsNull() bool {
	return false
}
