package vugu

// RegisteredComponentTypes returns a copy of the map of registered component types.
func RegisteredComponentTypes() ComponentTypeMap {
	// make a copy
	ret := make(ComponentTypeMap, len(globalComponentTypeMap))
	for k, v := range globalComponentTypeMap {
		ret[k] = v
	}
	return ret
}

func RegisterComponentType(tagName string, ct ComponentType) {
	globalComponentTypeMap[tagName] = ct
}

var globalComponentTypeMap = make(ComponentTypeMap)

// ComponentTypeMap is a map of the component tag name to a ComponentType.
type ComponentTypeMap map[string]ComponentType

type Props map[string]interface{}

// Clone makes a copy of the Props map, distinct from the original.
func (p Props) Clone() Props {
	ret := make(Props, len(p))
	for k, v := range p {
		ret[k] = v
	}
	return ret
}

// Merge will copy everything from p2 into p and return p for ease of use.
func (p Props) Merge(p2 Props) Props {
	for k, v := range p2 {
		p[k] = v
	}
	return p
}

type ComponentType interface {
	// TagName() string                                                      // HTML-compatible tag name, e.g. "demo-button"
	BuildVDOM(data interface{}) (vdom *VGNode, css *VGNode, reterr error) // based on the given data, build the VGNode tree
	NewData(props Props) (interface{}, error)                             // initial data when component is instanciated
	// Invoke(name string, args ...interface{}) error

	// NOTES:
	// props???
	// template
	// data
	// methods
	// computed!? - probably best implemented purely in the data binding code, rather than as part of components
	// lifecycle hooks (created, mounted, etc.)
}

func New(ct ComponentType, props Props) (*ComponentInst, error) {
	data, err := ct.NewData(props)
	if err != nil {
		return nil, err
	}
	return &ComponentInst{
		Type: ct,
		Data: data,
	}, nil
}

type ComponentInst struct {
	Type ComponentType
	Data interface{}
	// ChildComponents ComponentInstPtrList
}

// type ComponentInstPtrList []*ComponentInst
