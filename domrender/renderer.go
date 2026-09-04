package domrender

import (
	"syscall/js"
)

type JSRenderer struct {
	eventFunc js.Func
}

func (r *JSRenderer) eventHandler(this js.Value, args []js.Value) interface{} {
	go func() {
		r.handleEvent(this, args)
	}()
	return nil
}

func (r *JSRenderer) handleEvent(this js.Value, args []js.Value) {
}
