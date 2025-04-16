package vugu

// HTMLer describes something that can return HTML.
type HTMLer interface {
	HTML() string // return raw html (with any needed escaping already done)
}

// HTML implements the HTMLer interface on a string with no transform, just returns
// the string as-is for raw HTML.
type HTML string

// HTML implements the HTMLer interface.
func (h HTML) HTML() string {
	return string(h)
}

// NOTE: I'm bailing on this OptionalHTMLer thing because you can get the same
// functionality with an explicit vg-if.  It's unclear how much benefit
// it is to hide an element when you pass it a nil and if it's worth the effort
// and additional complexity. It is much more important that we handle escaping
// properly to prevent XSS, so we're going to simplify and focus on that.

// // OptionalHTMLer is like HTMLer but can also explicitly express "nothing here"
// // as distinct from an empty string.  This is used, for example, to express
// // the difference between `<div></div>` and no tag at all.
// type OptionalHTMLer interface {
// 	// return html value and true for HTML content, or false to indicate no content.
// 	OptionalHTML() (string, bool)
// }

// // OptionalHTML implements OptionalHTMLer using a string pointer.
// type OptionalHTML struct {
// 	Value *string // nil means return ("", false) from OptionalHTML
// }

// // OptionalHTML implements the OptionalHTMLer interface.
// func (h OptionalHTML) OptionalHTML() (string, bool) {
// 	if h.Value == nil {
// 		return "", false
// 	}
// 	return *h.Value, true
// }
