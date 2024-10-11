package main

// The Parent component struct.
type Parent struct {
	msg string // this is the message that will be set in response to the button clicks of the button components
}

// Init is a component lifecycle method. It is called when the component is created.
// We use it is to initialise the array and the map
// See:
// https://www.vugu.org/doc/components
func (c *Parent) Init() {
	c.SetMsg("Click on one of the buttons below to change this text!")
}

// Sets the message we want to display.
// This is called by the button components in response to a click DOM event
func (c *Parent) SetMsg(msg string) {
	c.msg = msg
}

// Return the current message ready to be displayed
func (c *Parent) Msg() string {
	return c.msg
}
