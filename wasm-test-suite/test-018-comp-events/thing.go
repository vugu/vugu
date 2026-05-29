package main

import "github.com/vugu/vugu"

type Thing struct {
	Click ClickHandler
}

func (c *Thing) HandleButtonClick(event vugu.DOMEvent) {
	if c.Click != nil {
		c.Click.ClickHandle(ClickEvent{DOMEvent: event})
	}
}

//vugugen:event Click

type ClickEvent struct {
	vugu.DOMEvent
}

func (e ClickEvent) String() string {
	return "This is the ClickEvent"
}
