package vugu

// AttributeLister implements the required functionality, required by the `:=` attribute setter.
// Types may implement this interface, to
type VGAttributeLister interface {
	AttributeList() []VGAttribute
}

// The VGAttributeListerFunc type is an adapter to allow the use of functions as VGAttributeLister
type VGAttributeListerFunc func() []VGAttribute

// AttributeList calls the underlying function
func (f VGAttributeListerFunc) AttributeList() []VGAttribute {
	return f()
}
