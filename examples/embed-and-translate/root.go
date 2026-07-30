package main

import (
	"embed"
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/vugu/vugu"
	"golang.org/x/text/language"
)

// We use the embed package to embed each of the language files.
// Although we build a WASM binary, the embed package works in the same was as any other other target.
//
//go:embed en.json
//go:embed fr.json
//go:embed it.json
var fs embed.FS // fs is the embedded file system interface

type Root struct {
	// The translated message we want to display to the user
	msg string
	// The BCP 47 encoded language value e.e.g "en-GB" or just "en"
	language string
	// The i18n translation bundle which stores the messages and their translations.
	bundle *i18n.Bundle
}

// Init the root component via the vugu Init lifecycle method. We use it to initialise the i18n bundle.
func (c *Root) Init() {
	c.initI18n()
}

// initI18n initializes the i18n bundle ready for use, but loading each of the translation files.
// As there is typically only one bundle per application, this initialisation code would be better placed
// in the main() of teh application or an Init function of the main package.
// The i18n bundle would then be passed into the Root component via a new Root constructor.
// We haven't taken this approach, so as to maintain the same style as the other examples, but suggest this in a production setting.
func (c *Root) initI18n() {
	c.language = language.English.String()
	c.bundle = i18n.NewBundle(language.English)
	c.bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	// Load the 3 transaltion files from the embedded FS
	// We are ignoring the return values inc. the error for simplicity. Don't do this in production.
	// We intend only to show that the embedded FS can be used to read files in the WASM binary.
	// If these load functions errored the only reasonable choice here would be to panic, especially as we can't return an error from either
	// the components Init(), the main modules Init() or the main() functions.
	c.bundle.LoadMessageFileFS(fs, "en.json")
	c.bundle.LoadMessageFileFS(fs, "fr.json")
	c.bundle.LoadMessageFileFS(fs, "it.json")
}

// Return a localised copy of the message.
// The message is translated according to the value of c.language which is set via the Change method
// The message itself is defined in each of the language file - the *.json files that are embedded within the binary.
func (c *Root) Msg() string {
	localizer := i18n.NewLocalizer(c.bundle, c.language)
	// As this is an example we are ignoring any possible error from Localize. Don't do this in production.
	c.msg, _ = localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "hello_in_different_language", // this is the ID of the message we want to translate.
	})
	return c.msg // return the translated message
}

// Change is the event function that is called each time a radio button is selected.
// It takes the value of the radio and converts it to the string form of a Go langage.Tag.
// If we don;t find a matching value we default to English so that we always have a valid (translated) message
// Note: we switch on the "value" property and do not use the "value" property directly.
// This avoids using a untrusted input form the web page.
func (c *Root) Change(e vugu.DOMEvent) {
	switch e.PropString("target", "value") {
	case "English":
		c.language = language.English.String()
	case "Français":
		c.language = language.French.String()
	case "Italiano":
		c.language = language.Italian.String()
	default:
		c.language = language.English.String() // ensure we never fail to set a language
	}
}

// Returns the BCP 47 encoded language string - which might be the short form e.e. just "en" and not "en-GB"
func (c *Root) SelectedLanguage() string {
	return c.language
}
