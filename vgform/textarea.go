package vgform

import "github.com/vugu/vugu"

// Textarea corresponds to a textarea HTML element.
type Textarea struct {
	Value   StringValuer // get/set the currently selected value
	AttrMap vugu.AttrMap
}

func (c *Textarea) handleChange(event vugu.DOMEvent) {

	newVal := event.PropString("target", "value")
	// c.curVal = newVal // why not
	c.Value.SetStringValue(newVal)

}
