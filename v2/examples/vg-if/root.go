package main

import (
	"fmt"
	"slices"
)

// Root is the root component of the application in the DOM
// This example has only one component that represents a list whose visibility can be controlled.
type Root struct {
	// Show determines the visibility of the list data. Defaults to false.
	Show bool `vugu:"data"` // we need `vugu:"data"` so that the DOM sees any change in state. The properly must also be exported for the same reason.
	// list contains the list of number to be manipulated using stack style push and pop operations
	list []int
}

// Init is a component lifecycle method. It is called when the component is created.
// We use it is to initialise the list and make the list visible initially
// See:
// https://www.vugu.org/doc/components
func (c *Root) Init() {
	c.Show = true
	c.list = []int{1, 2, 3, 4, 5}
}

// Change the visibility of the list
func (c *Root) ToggleShow() bool {
	c.Show = !c.Show
	fmt.Printf("c.Show: %v\n", c.Show) // the output goes to the javascript console in the browser
	return c.Show
}

// Return a copy of the list in its current state
func (c *Root) List() []int {
	return c.list
}

// Return the current list length
func (c *Root) ListLength() int {
	return len(c.list)
}

// Push a new element onto the end of the list
// Note: this implementation uses the new slices package introduces in Go 1.21
func (c *Root) Push() {
	c.list = slices.Insert(c.list, len(c.list), len(c.list)+1)
	fmt.Printf("Push c.List: %v\n", c.list)
}

// Pop the last element from the end of the list
// Note: this implementation uses the new slices package introduces in Go 1.21
func (c *Root) Pop() {
	if len(c.list) > 0 {
		c.list = slices.Delete(c.list, len(c.list)-1, len(c.list))
	}
	fmt.Printf("Pop  c.list: %v\n", c.list)
}

// Revere the list in place
func (c *Root) Reverse() {
	slices.Reverse(c.list)
}
