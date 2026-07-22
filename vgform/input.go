package vgform

import "github.com/vugu/vugu"

// Input corresponds to an input HTML element.
// What you provide for the `type` attribute can trigger behavioral differences appropriate
// for specific input types.
//
// type="checkbox"
// type="radio"
// (list them out)
type Input struct {
	Value   StringValuer // get/set the currently selected value
	AttrMap vugu.AttrMap
}

func (c *Input) handleChange(event vugu.DOMEvent) {

	newVal := event.PropString("target", "value")
	// c.curVal = newVal // why not
	c.Value.SetStringValue(newVal)

}
