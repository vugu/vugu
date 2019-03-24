package vugu

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"sync"

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

// EventEnv provides locking mechanism to for rendering environment to events so
// data access and rendering can be synchronized and avoid race conditions.
type EventEnv interface {
	Lock()         // acquire write lock
	UnlockOnly()   // release write lock
	UnlockRender() // release write lock and request re-render

	RLock()   // acquire read lock
	RUnlock() // release read lock
}

// eventEnv implements EventEnv
type eventEnv struct {
	rwmu            *sync.RWMutex
	requestRenderCH chan bool
}

// acquire write lock
func (ee *eventEnv) Lock() {
	// if ee.rwmu == nil {
	// 	return
	// }
	ee.rwmu.Lock()
}

// release write lock
func (ee *eventEnv) UnlockOnly() {
	// if ee.rwmu == nil {
	// 	return
	// }
	ee.rwmu.Unlock()
}

// release write lock and request re-render
func (ee *eventEnv) UnlockRender() {
	// if ee.rwmu != nil {
	ee.rwmu.Unlock()
	// }
	if ee.requestRenderCH != nil {
		// send non-blocking
		select {
		case ee.requestRenderCH <- true:
		default:
		}
	}
}

// acquire read lock
func (ee *eventEnv) RLock() {
	// if ee.rwmu == nil {
	// 	return
	// }
	ee.rwmu.RLock()
}

// release read lock
func (ee *eventEnv) RUnlock() {
	// if ee.rwmu == nil {
	// 	return
	// }
	ee.rwmu.RUnlock()
}
