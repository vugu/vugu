package main

import "github.com/vugu/vugu"

// Nameinput component struct contains a reference to the controlling form. This is set at when the component is initialised in the Sampleform.vugu file
type Nameinput struct {
	Form *Sampleform
}

// Change reads the new value from the component and assigns the value to the name field in the controlling form,
func (c *Nameinput) Change(e vugu.DOMEvent) {
	e.PreventDefault()
	name := e.PropString("target", "value")
	c.Form.SetName(name)
}
