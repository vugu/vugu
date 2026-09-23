package main

import "github.com/vugu/vugu"

type Root struct {
	EmailValue string
}

func (c *Root) InitEmail(event vugu.DOMEvent) {
	event.PreventDefault()
	c.EmailValue = "default@example.com"
}

func (c *Root) EmailChanged(event vugu.DOMEvent) {
	c.EmailValue = event.PropString("target", "value")
}
