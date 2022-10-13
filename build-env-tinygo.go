// +build tinygo

package vugu

import (
	"reflect"
	"unsafe"
)

type buildCacheKey struct {
	typ, ptr uintptr
}

func makeBuildCacheKey(v interface{}) buildCacheKey {
	vv := reflect.ValueOf(v)
	var ret buildCacheKey
	// idata := vv.InterfaceData()
	// ret.typ, ret.ptr = idata[0], idata[1]
	tp := vv.Type()
	ret.typ = uintptr(unsafe.Pointer(&tp)) // HACK: this is an implementation-specific thing for TinyGo as it stands right now, will need be refactored as reflect is built out
	ret.ptr = vv.Pointer()
	return ret
}
