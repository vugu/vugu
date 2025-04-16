package main

import (
	"github.com/vugu/vugu"
)

// Staticselect component struct contains a reference to the conrolling form. This is set at initialisation time in Sampleform.vugu
type Staticselect struct {
	Form *Sampleform
}

// Init is a vugu lifecycle method it initialises the default value.
func (c *Staticselect) Init() {
	c.Form.SetCar("Audi") // ensure we never fail to set a value this matched the "selected" value in Staticselect.vugu
}

// Change is the event function that is called each time a new option is selected.
// It reads the value attribute of the option and NOT the value of the option itself.
// This approach avoids any possibility of using untrusted user supplied input to the Option tag.
func (c *Staticselect) Change(e vugu.DOMEvent) {
	e.PreventDefault()
	selection := e.PropString("target", "value")
	c.Form.SetCar(selection)
}
