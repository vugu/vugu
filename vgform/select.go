package vgform

import (
	"errors"

	"github.com/vugu/vugu"
)

// Select wraps an HTML select element.
//
// For clarity in naming: The word "value" refers to the value
// of the currently selected option.  The word "key" refers to
// the text inside the value attribute of an option tag, and the
// word "text" refers to the HTML content of an option tag, i.e.
// `<option value="some-key">some text</option>`.
//
// The Value field is a StringValuer which is used to get the current
// value upon render, and update it when changed.
// The StringPtr type provides a simple adapter for *string.
//
// NOTE: When slots are supported this will be updated to support
// providing your own option tags so you can do things like
// option groups.
//
// For the common case of simple lists of texts, set the Options
// as appropriate.  See SliceOptions and MapOptions
// for convenient adapters for []string and map[string]string.
type Select struct {
	Value   StringValuer // get/set the currently selected value
	Options Options      // provide KeyLister and TextMapper in one
	// AttrMap map[string]interface{} // regular HTML attributes like id and class
	AttrMap vugu.AttrMap // regular HTML attributes like id and class

	// el     js.Value
	keys   []string
	curVal string
}

func (s *Select) buildKeys() []string {

	// if s.el.IsUndefined() {
	// 	panic(errors.New("Select should have s.el set"))
	// }
	if s.Value == nil {
		panic(errors.New("Select.Value must not be nil"))
	}
	if s.Options == nil {
		panic(errors.New("Select.Options must not be nil (TODO: when slots are supported that will be allowed instead of Options)"))
	}

	s.keys = s.Options.KeyList()
	s.curVal = s.Value.StringValue()

	return s.keys
}

func (s *Select) isOptSelected(k string) bool {
	return s.curVal == k
}

func (s *Select) optText(k string) string {
	return s.Options.TextMap(k)
}

func (s *Select) handlePopulate() {

	// var buf bytes.Buffer
	// buf.Grow(len(`<option value=""></option>`) * len(s.keys) * 2)
	// for _, k := range s.keys {
	// 	buf.WriteString(`<option value="`)
	// 	buf.WriteString(html.EscapeString(k))
	// 	buf.WriteString(`">`)
	// 	buf.WriteString(html.EscapeString(s.Options.TextMap(k)))
	// 	buf.WriteString(`</option>`)
	// }
	// s.el.Set("innerHTML", buf.String())

	// see if we need this...
	// v := s.curVal
	// idx := -1
	// for i, k := range s.keys {
	// 	if k == v {
	// 		idx = i
	// 	}
	// }

	// s.el.Set("selectedIndex", idx)

}

func (s *Select) handleChange(event *vugu.DOMEvent) {

	newVal := event.PropString("target", "value")
	s.curVal = newVal // why not
	s.Value.SetStringValue(newVal)

}
