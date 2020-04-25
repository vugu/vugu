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
	Value StringValuer // get/set the currently selected value
	// TODO: should we also just make a ValuePtr *string - which would let people
	// do :ValuePtr="&c.SomeRegularString" - seems like some people will want the convenience
	// TODO: multiple (will need a new field and a new type, StringSliceValuer?)

	Options Options // provide KeyLister and TextMapper in one

	AttrMap vugu.AttrMap // regular HTML attributes like id and class
	// TODO: might make sense to refactor this AttrMap thing to work well
	// with SetAttributeInterface and AttributeLister/vg-attr - perhaps a slice of attributes
	// is better and for vg-attr we just append to the slice and for the
	// others we us SetAttributeInterace (or some variation that uses the same
	// logic but works on an list of attributes.)  A type AttributeList []VGAttribute
	// might be in order.  Refactor to separate vgnode package might be appropriate...

	// el     js.Value
	keys   []string
	curVal string
}

func (c *Select) buildKeys() []string {

	// if c.el.IsUndefined() {
	// 	panic(errors.New("Select should have c.el set"))
	// }
	if c.Value == nil {
		panic(errors.New("Select.Value must not be nil"))
	}
	if c.Options == nil {
		panic(errors.New("Select.Options must not be nil (TODO: when slots are supported that will be allowed instead of Options)"))
	}

	c.keys = c.Options.KeyList()
	c.curVal = c.Value.StringValue()

	return c.keys
}

func (c *Select) isOptSelected(k string) bool {
	return c.curVal == k
}

func (c *Select) optText(k string) string {
	return c.Options.TextMap(k)
}

func (c *Select) handlePopulate() {

	// var buf bytes.Buffer
	// buf.Grow(len(`<option value=""></option>`) * len(c.keys) * 2)
	// for _, k := range c.keys {
	// 	buf.WriteString(`<option value="`)
	// 	buf.WriteString(html.EscapeString(k))
	// 	buf.WriteString(`">`)
	// 	buf.WriteString(html.EscapeString(c.Options.TextMap(k)))
	// 	buf.WriteString(`</option>`)
	// }
	// c.el.Set("innerHTML", buf.String())

	// see if we need this...
	// v := c.curVal
	// idx := -1
	// for i, k := range c.keys {
	// 	if k == v {
	// 		idx = i
	// 	}
	// }

	// c.el.Set("selectedIndex", idx)

}

func (c *Select) handleChange(event vugu.DOMEvent) {

	newVal := event.PropString("target", "value")
	c.curVal = newVal // why not
	c.Value.SetStringValue(newVal)

}
