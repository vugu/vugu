package vugu

import (
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/cespare/xxhash"
)

// NOTE: common event code which is the same for wasm and not can go here

type DOMEventHandler struct {
	Method reflect.Value // method to be called, with receiver baked into it if needed (see reflect/Value.MethodByName)
	Args   []interface{} // arguments to be passed in when calling (special case for eventStub)
}

// func (d DOMEventHandler) hashString() string {
// 	return fmt.Sprintf("%x", d.hash())
// }

func (d DOMEventHandler) hash() uint64 {
	if !d.Method.IsValid() && len(d.Args) == 0 {
		return 0
	}

	b8 := make([]byte, 8)
	h := xxhash.New()

	// TODO: this may end up being an issue but I need to move through this for now -
	// this is only unique for the method itself, not the instance it's tied to, although
	// it's probably rare to have to distinguish between multiple component types in use in an application
	fmt.Fprintf(h, "%#v", d.Method)

	// add the args
	for _, a := range d.Args {
		binary.BigEndian.PutUint64(b8, ComputeHash(a))
		h.Write(b8)
	}
	return h.Sum64()
}
