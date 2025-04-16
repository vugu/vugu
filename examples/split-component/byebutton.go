package main

import (
	"github.com/vugu/vugu"
)

// The Byebutton component
type Byebutton struct {
	// Parent is a reference to a Parent struct. This is the component that the Byebutton is embedded within in the HTML
	Parent *Parent
}

// Called when the Byebutton receives an onclick DOM event. It calls Parent.SetMsg(...) internally to update the message in the parent control
func (c *Byebutton) Click(event vugu.DOMEvent) {
	c.Parent.SetMsg("... Goodbye cruel world, goodbye.")
}
