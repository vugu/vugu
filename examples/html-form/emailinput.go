package main

import "github.com/vugu/vugu"

// Emailinout component struct holds a reference to the containing form. This is set at construction time in the Samplefom.vugu file
type Emailinput struct {
	Form *Sampleform
}

// Change reads the new valueof the email input control, and sets that value in the controlling form component.
func (c *Emailinput) Change(e vugu.DOMEvent) {
	e.PreventDefault()
	email := e.PropString("target", "value")
	c.Form.SetEmail(email)
}
