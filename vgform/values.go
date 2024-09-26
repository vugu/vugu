package vgform

import "errors"

// StringValuer is a string that can be gotten and set.
type StringValuer interface {
	StringValue() string
	SetStringValue(string)
}

// TODO: hm, would this not be better as:
// type StringValue string
// func (s *StringValue) StringValue() string {
// This would allow people to either cast a *string to *StringValue
// or just use vgform.StringValue directly as a struct member and
// pass it's address right into the Value property, i.e.
// < ... :Value="&c.SomeStringValue">
// Maybe a StrPtr(&c.RegularString) would be good.

// NOTE: It is useful to do `StringPtr{something}` so give careful
// thought before adding another field to StringPtr.
// StringPtr must be a struct because you cannot add methods to
// types declared as `type x *string`.

// StringPtr implements StringValuer on a string pointer.
type StringPtr struct {
	Value *string
}

// StringPtrDefault returns a StringPtr and sets the underlying string to def if it empty.
func StringPtrDefault(p *string, def string) StringPtr {
	if p == nil {
		panic(errors.New("StringPtr must not have a nil pointer"))
	}
	if *p == "" {
		*p = def
	}
	return StringPtr{p}
}

// StringValue implements StringValuer
func (s StringPtr) StringValue() string {
	// I can't see any benefit to hiding the nil ptr by returning "",
	// especially because when the value comes back with SetStringValue
	// we would be forced to throw it away.
	// Since this is probably a mistake we panic and let someone know.
	if s.Value == nil {
		panic(errors.New("StringPtr must not have a nil pointer"))
	}
	return *s.Value
}

// SetStringValue implements StringValuer
func (s StringPtr) SetStringValue(v string) {
	if s.Value == nil {
		panic(errors.New("StringPtr must not have a nil pointer"))
	}
	*s.Value = v
}
