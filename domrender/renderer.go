package domrender

import (
	"github.com/vugu/vugu"
)

// JSRenderer applies Vugu build results to the browser DOM.
type JSRenderer struct {
}

// Render applies the build results to the DOM and executes lifecycle methods.
func (r *JSRenderer) Render(results *vugu.BuildResults) error {
	// ... existing DOM synchronization logic ...

	// Call Rendered() on all components that were part of the build and implement the interface.
	for _, comp := range results.Components {
		if rendered, ok := comp.(vugu.Rendered); ok {
			rendered.Rendered()
		}
	}

	return nil
}
