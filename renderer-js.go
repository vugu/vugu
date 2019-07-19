package vugu

import (
	"fmt"

	js "github.com/vugu/vugu/js"
)

type JSRenderer struct {
}

func (r *JSRenderer) Render(b *BuildOut) error {
	if !js.Global().Truthy() {
		return fmt.Errorf("js environment not available")
	}
	panic(fmt.Errorf("not yet implemented"))
}
