// +build !wasm

package vugu

import (
	"errors"
	"io"
)

var _ Env = (*JSEnv)(nil) // assert type

// JSEnv is responsible for rendering (efficiently synchronizing) HTML in the browser in a WebAssembly application.
// Browser calls are performed using syscall/js.  Various internal hashing and other optimizations are performed
// to attempt to make rendering as effience as possible.  If used outside of a wasm program all methods will panic.
type JSEnv struct {
	MountParent string    // query selector
	DebugWriter io.Writer // write debug information about render details to this Writer if not nil
}

// NewJSEnv returns a new instance with the specified mountParent (CSS selector of where in the HTML the ouptut will go),
// the root component instance, and any other component types available.
func NewJSEnv(mountParent string, rootInst *ComponentInst, components ComponentTypeMap) *JSEnv {
	panic(errors.New("JSEnv is only available in wasm"))
}

// RegisterComponentType assigns a component type to a tag name.
func (e *JSEnv) RegisterComponentType(tagName string, ct ComponentType) {
	panic(errors.New("JSEnv is only available in wasm"))
}

// EventWait blocks until an event occurs and re-rendering is needed.
// Returns true unless the browser has gone away and the program should exit.
func (e *JSEnv) EventWait() bool {
	panic(errors.New("JSEnv is only available in wasm"))
}

// Render is where the magic happens and your root component instance has its virtual DOM rendered and synchronized to the browser DOM.
// Render attempts to perform as few operations as possible and uses hashing to avoid unnecessary work.
// Component instances of children of the root component (referenced with HTML tags that have the components tag name),
// are managed as rendering occurs.  Child components are created according to a hash of their properties and are recreated
// if their position in the DOM changes OR if their properties (HTML attributes on the referencing tag) change.
// Render is normally called in a loop with EventWait() being used to block until re-rendering is needed.
func (e *JSEnv) Render() error {
	panic(errors.New("JSEnv is only available in wasm"))
}
