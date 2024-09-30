package main

import (
	"iter"

	"github.com/vugu/vugu"
)

// Root is the root component of the application in the DOM
// The Root contains two strings, to indicate mouse over and click events in the red box
// and a string slice that holds the type of the event (or a user supplied string)
type Root struct {
	mouseOverText string
	clickText     string
	events        []string
}

// Init is a component lifecycle method. It is called when the component is created.
// We use it is to initialise the list and make the list visible initially
// See:
// https://www.vugu.org/doc/components
func (c *Root) Init() {
	c.mouseOverText = "Mouse over and click me!"
	c.events = make([]string, 0, 64)
}

// The event listener for the "onmouseover" event in the red rectangle control
func (c *Root) MouseOver(e vugu.DOMEvent) {
	c.logEvent(e)
	c.mouseOverText = "Mouse entered"
}

// The event listener for the "onmouseout" event in the red rectangle control
func (c *Root) MouseOut(e vugu.DOMEvent) {
	c.mouseOverText = "Mouse exited"
	c.clickText = "" // clear the click text when the mouse leaves the red box
	c.logEvent(e)
}

// The event listener for the "onclick" event in the red rectangle control
// We need a separate method to distinguish from clicks on the button, or any other control
func (c *Root) RedBoxClick(e vugu.DOMEvent) {
	c.clickText = "I've been clicked"
	c.logEventStr("click in red box") // the default would log an event of "click" which doesn't tell us which control was clicked. So we use a unique string.
}

// Set the text in the red rectangle. If the rectangle has been clicked within display teh click text, otherwise display the mouse entry/exit text.
func (c *Root) RedBoxText() string {
	if c.clickText != "" {
		return c.clickText
	}
	return c.mouseOverText
}

// The event listener for the "onkeydown" event in the textarea control
func (c *Root) KeyDown(e vugu.DOMEvent) {
	c.logEvent(e)
}

// The event listener for the "onclick" event in the button control
func (c *Root) ButtonLeftClick(e vugu.DOMEvent) {
	c.logEventStr("left click on button")
}

// The event listener for the "oncontextclick" event in the button control
func (c *Root) ButtonRightClick(e vugu.DOMEvent) {
	c.logEventStr("right click on button") // the default "type" is "contextclick" when we really want right click
}

// The event listener for the "ondblclick" event in the button control
func (c *Root) ButtonDoubleClick(e vugu.DOMEvent) {
	c.logEventStr("double click on button") // the default "type" is "dblclick" when we really want right click
}

// The log the event type. The event type is held in the "type" property of the event.
func (c *Root) logEvent(e vugu.DOMEvent) {
	c.events = append(c.events, e.PropString("type")) // add to head
}

// Log a custom string - this allows the overriding the event type
func (c *Root) logEventStr(e string) {
	c.events = append(c.events, e) // add to head
}

// Return a iterator that loops from newest event to oldest
// Requires go 1.23
func (c *Root) Events() iter.Seq[string] {
	return func(yield func(string) bool) {
		for i := len(c.events) - 1; i >= 0; i-- { // loop in reverse over the events slice so we yield the latest event first
			if !yield(c.events[i]) {
				return
			}
		}
	}
}
