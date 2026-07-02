package main

// The Parent component struct.
// The relationship between the Parent control and the SimpleForm (and its sub controls) is described in the `parent.vugu` file.
// We DO NOT need to embed the Simpleform stuct (or it's subcomponents) here.
// What we do need to describe in the Parent stuct are the methods that the are called on the Parent component from the `parent.vugu` file.
type Parent struct {
	encodedFormData string // this is the (JSON) encoded version of the form data
	submitted       bool   // indicates that the forms submit button has been pressed.
}

// Sets the encoded version of the form data. The encoding is handled by the form itself.
// This is called when the `Submit` function is called in the form in response to the 'Submit' button being clicked.
func (c *Parent) SetEncodedFormData(encodedFormData string) {
	c.encodedFormData = encodedFormData
}

// Return the the encoded form data. The encoding is handled by the form itself.
func (c *Parent) EncodedFormData() string {
	return c.encodedFormData
}

// Set the form submission indicator. Called when the forms `Submit(e vugu.DOMEvent)` has been called in response to the Submit button being pressed.
func (c *Parent) SetSubmitted() {
	c.submitted = true
}

// Has the "Submit" button ben pressed or not.
func (c *Parent) Submitted() bool {
	return c.submitted
}
