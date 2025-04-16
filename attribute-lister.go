package vugu

// VGAttributeLister implements the required functionality, required by the `:=` attribute setter.
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

// AttrMap implements VGAttributeLister as a map[string]interface{}
type AttrMap map[string]interface{}

// AttributeList returns an attribute list corresponding to this map using the rules from VGNode.AddAttrInterface.
func (m AttrMap) AttributeList() (ret []VGAttribute) {
	var n VGNode
	for k := range m {
		n.AddAttrInterface(k, m[k])
	}
	return n.Attr
}
