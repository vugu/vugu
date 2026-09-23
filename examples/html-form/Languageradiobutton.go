package main

import (
	"github.com/vugu/vugu"
)

// The Languageradiobutton struct maps to a radio button input type that looks like this:
// <div>
//
//	<input type="radio" id="english" name="language" value="English" checked='c.IsSelectionDefault()' @change='c.Change(event)'></input>
//	<label for="english">English</label><br>
//
// </div>
type Languageradiobutton struct {
	Group *Radiobuttons // the controlling form
	Id    string        // the "id" HTML/JS attribute
	Value string        // the "value" HTML/JS attribute
}

// IsSelectionDefault returns true if the Value of this Langiageradiobutton instance matches the default selection in the controlling radio group
func (c *Languageradiobutton) IsSelectionDefault() bool {
	return c.Group.IsSelectionDefault(c.Value)
}

// Change is called in response ot the "onchange" ever of the radio button. This can only be called when a new value is
// selected. So the Languageradiobutton instance that is called is the one whose value has just changed.
// We don't propgate the event itself to the containing radio group but instead inform the radio group of our value.
// The radio group will then perform any further processing
func (c *Languageradiobutton) Change(e vugu.DOMEvent) {
	e.PreventDefault()
	// bubble up and set the new value in the containing radio group
	c.Group.Change(c.Value)
}
