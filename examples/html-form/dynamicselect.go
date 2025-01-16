package main

import (
	"log"
	"math/rand/v2"
	"strconv"

	"github.com/vugu/vugu"
)

// Dyanmicselect struct component, holds a reference to the controlling form, the number of random numbers and their value.
type Dynamicselect struct {
	Form            *Sampleform
	numberOfOptions int // strictly we don't need this - as it's just len(randomNumbers) + 1
	randomNumbers   []int
}

// Init is a vugu component lifecycle method called when the component is first created.
// we use it to initialise a slice of integers. The length of the slice is a random number between
// [1..10] (inclusive). Each random number is between ]0..99] (inclusive)
func (c *Dynamicselect) Init() {
	// generate a randon number - this is the number of options we will display
	c.numberOfOptions = rand.IntN(9) + 1 // max of 10 options [1..10] inclusive
	c.randomNumbers = make([]int, c.numberOfOptions)
	// generate some random numbers - between 0 and 99
	for i := range c.randomNumbers {
		c.randomNumbers[i] = rand.IntN(100)
	}
	// set the default to the first number in the slice
	c.Form.SetRandomNumber(c.randomNumbers[0])
	log.Printf("Random Numbers: %v", c.randomNumbers) // this will log the list of numbers chosen to the browser console
}

// NumberOfOptions returns the (random) number of possible options
func (c *Dynamicselect) NumberOfOptions() int {
	return c.numberOfOptions
}

// Random returns the ith randon number from the list
func (c *Dynamicselect) Random(i int) int {
	return c.randomNumbers[i]
}

// Change is the event function that is called each time a new option is selected.
// It reads the value attribute of the option and NOT the value of the option itself.
// This "value" attribute of the option tag is set dynamically in the Dynamicselect.vugu file.
// The value is the index position in the list (slice) of random numbers, and not the value of the random number itself.
// We then lookup and return the ith random number.
// This approach avoids any possibility of using untrusted user supplied input to the Option tag.
func (c *Dynamicselect) Change(e vugu.DOMEvent) {
	e.PreventDefault()
	selection := e.PropString("target", "value") // we look up the value as a string (in principle it could be an int, but that's not hwo we defined it in the corresponding vugu file)
	log.Printf("selection: %T %v", selection, selection)
	n, _ := strconv.Atoi(selection) // ignore the error
	c.Form.SetRandomNumber(c.randomNumbers[n])
}
