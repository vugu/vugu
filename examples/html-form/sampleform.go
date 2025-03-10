package main

import (
	"encoding/json"

	"github.com/vugu/vugu"
)

// A simple struct to trepresent the forms data. The Go struct tags are the key names in the JSON encoded version of the data
type formData struct {
	Email        string `json:"email"`
	Language     string `json:"lang"`
	Name         string `json:"name"`
	Car          string `json:"car"`
	RandomNumber int    `json:"randomnumber"`
}

// The form component struct. We don't embed the other form components here. The component hierarchy is described in the Sampleform.vugu file
// We do however need to define the methods on the struct that the sub components can call.
type Sampleform struct {
	Parent *Parent
	data   formData
}

// Submit is called when the Submit button is pressed.
// The method encodes the form data into JSON, and sets the encoded version in the Parent control. It also sets the parents submit flag to indicated the submit
// button has been pressed.
func (c *Sampleform) Submit(e vugu.DOMEvent) {
	e.PreventDefault()
	// for the purposes of the example we are ignoring the error.  We are also called via an event handler which can't return an error directly.
	b, _ := json.Marshal(c.data)
	c.Parent.SetEncodedFormData(string(b))
	c.Parent.SetSubmitted()
}

// Set the email field in the form data
func (c *Sampleform) SetEmail(email string) {
	c.data.Email = email
}

// Set the langage field in the form data
func (c *Sampleform) SetLanguage(language string) {
	c.data.Language = language
}

// Set the name filed in the form data
func (c *Sampleform) SetName(name string) {
	c.data.Name = name
}

// Set the car field in the form data
func (c *Sampleform) SetCar(car string) {
	c.data.Car = car
}

// Set the random number field in the form data
func (c *Sampleform) SetRandomNumber(n int) {
	c.data.RandomNumber = n

}
