package main

import "github.com/vugu/vugu"

// The Hellobutton component
type Hellobutton struct {
	// Parent is a reference to a Parent struct. This is the component that the Hellobutton is embedded within in the HTML
	Parent *Parent
}

// Called when the Hellobutton receives an onclick DOM event. It calls Parent.SetMsg(...) internally to update the message in the parent control
func (c *Hellobutton) Click(event vugu.DOMEvent) {
	c.Parent.SetMsg("Hello world....")
}
