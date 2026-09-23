package main

import (
	"golang.org/x/text/language"
)

// Radiobuttons component struct it contains a reference to the controlling form. This is set at initialisation time in Sampleform,vugu
type Radiobuttons struct {
	Form             *Sampleform // The form that contains the radio buttons
	SelectionDefault string      // The default value for the radio buttons group
}

// Init is a vugu lifecycle function and initialises the Radiobuttons component. In this case it ensures the default language
// is always set to to the default value defined by the radio button group in the vugu file.
func (c *Radiobuttons) Init() {
	c.Form.SetLanguage(c.SelectionDefault) // ensure we never fail to set a language
}

// IsSelectionDefault reports if the passed value is the default value
// The lower level Languageradiobuttons call this to determine if they should set their "checked" attribute
func (c *Radiobuttons) IsSelectionDefault(value string) bool {
	return value == c.SelectionDefault
}

// Change is called in response to an onChange event on a lower level Languageradiobutton.
// It passes the value of the Languageradiobutton that has received the onChange event.
// The Radiobuttons control then sets this value in the controlling form struct, or the default
// value if no match is found.
func (c *Radiobuttons) Change(v string) {
	switch v {
	case "English":
		c.Form.SetLanguage(language.English.String())
	case "Français":
		c.Form.SetLanguage(language.French.String())
	case "Italiano":
		c.Form.SetLanguage(language.Italian.String())
	case "Deutsch":
		c.Form.SetLanguage(language.German.String())
	default:
		c.Form.SetLanguage(c.SelectionDefault) // ensure we never fail to set a language
	}
}
