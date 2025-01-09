package main

import (
	"github.com/vugu/vugu"
	"golang.org/x/text/language"
)

// Radiobuttons component struct it contains a reference to the controlling form. This is set at initialisation time in Sampleform,vugu
type Radiobuttons struct {
	Form *Sampleform
}

// Init is a vugu lifecycle function and initialises the Radiobuttons component. In this case it ensures the default language
// is always set to English. This matches the "checked" value in the HTML in the Radiobuttons.vugu file.
func (c *Radiobuttons) Init() {
	c.Form.SetLanguage(language.English.String()) // ensure we never fail to set a language
}

// Change is the event function that is called each time a radio button is selected.
// It takes the value of the radio and converts it to the string form of a Go langage.Tag.
// If we don;t find a matching value we default to English
// Note: we switch on the "value" property and do not use the "value" property directly.
// This avoids using a untrusted input form the web page.
func (c *Radiobuttons) Change(e vugu.DOMEvent) {
	e.PreventDefault()
	switch e.PropString("target", "value") {
	case "English":
		c.Form.SetLanguage(language.English.String())
	case "Français":
		c.Form.SetLanguage(language.French.String())
	case "Italiano":
		c.Form.SetLanguage(language.Italian.String())
	default:
		c.Form.SetLanguage(language.English.String()) // ensure we never fail to set a language
	}
}
