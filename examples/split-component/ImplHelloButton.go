package main

import "github.com/vugu/vugu"

// Called when the Hellobutton receives an onclick DOM event. It calls Parent.SetMsg(...) internally to update the message in the parent control
func (c *Hellobutton) Click(event vugu.DOMEvent) {
	c.Parent.SetMsg("Hello world....")
}
