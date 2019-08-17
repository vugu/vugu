package vugu

import "sort"

// RegisteredComponentTypes returns a copy of the map of registered component types.
func RegisteredComponentTypes() ComponentTypeMap {
	// make a copy
	ret := make(ComponentTypeMap, len(globalComponentTypeMap))
	for k, v := range globalComponentTypeMap {
		ret[k] = v
	}
	return ret
}

// RegisterComponentType can be called during init to register a component and make it available by default to the rest of the application.
// A program may retrieved these by calling RegisteredComponentTypes() or it may choose to form its own set.  RegisterComponentType() just
// makes it available in case that's useful.  It is good practice for components to register themselves in an init function.
func RegisterComponentType(tagName string, ct ComponentType) {
	globalComponentTypeMap[tagName] = ct
}

var globalComponentTypeMap = make(ComponentTypeMap)

// ComponentTypeMap is a map of the component tag name to a ComponentType.
type ComponentTypeMap map[string]ComponentType

// Props is a map that corresponds to property names and values.  The name
// can correspond to an HTML attribute, or a property being passed in
// during component instantiation, depending on the context.
type Props map[string]interface{}

// OrderedKeys returns the keys sorted alphabetically.
func (p Props) OrderedKeys() []string {
	if p == nil {
		return nil
	}
	ret := make([]string, len(p))
	for k := range p {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret
}

// Clone makes a copy of the Props map, distinct from the original.
func (p Props) Clone() Props {
	ret := make(Props, len(p))
	for k, v := range p {
		ret[k] = v
	}
	return ret
}

// Merge will copy everything from p2 into p and return p for ease of use.  Does not copy p.
func (p Props) Merge(p2 Props) Props {
	for k, v := range p2 {
		p[k] = v
	}
	return p
}

type BuildIn struct {
	BuildEnv *BuildEnv

	// SlotMap map[string]BuilderFunc
}

type BuildOut struct {
	Out []*VGNode // output element(s) - usually just one but slots can have multiple
	CSS []*VGNode // optional CSS style tag(s)
	JS  []*VGNode // optional JS script tag(s)
}

func (b *BuildOut) AppendCSS(css string) {
	// FIXME: should we be deduplicating here??
	vgn := &VGNode{Type: ElementNode, Data: "style"}
	vgn.AppendChild(&VGNode{Type: TextNode, Data: css})
	b.CSS = append(b.CSS, vgn)
}

func (b *BuildOut) AppendJS(js string) {
	// FIXME: should we be deduplicating here??
	vgn := &VGNode{Type: ElementNode, Data: "script"}
	vgn.AppendChild(&VGNode{Type: TextNode, Data: js})
	b.JS = append(b.JS, vgn)
}

// Builder is the interface that components implement.
type Builder interface {
	Build(in *BuildIn) (out *BuildOut, err error)
}

// BuilderFunc is a Build-like function that implements Builder.
type BuilderFunc func(in *BuildIn) (out *BuildOut, err error)

// Build implements Builder.
func (f BuilderFunc) Build(in *BuildIn) (out *BuildOut, err error) { return f(in) }

// ComponentType is implemented by any type that wants to be a component.
// The BuildVDOM method is called to generate the virtual DOM for a component; and this method
// is usually code generated (by ParserGo) from a .vugu file.
// NewData provides for specific behavior when a component is initialized.
type ComponentType interface {
	BuildVDOM(data interface{}) (vdom *VGNode, css *VGNode, reterr error) // based on the given data, build the VGNode tree
	NewData(props Props) (interface{}, error)                             // initial data when component is instanciated
}

// New instantiates a component based on its ComponentType and a set of Props.
// It essentially just calls NewData and returns a ComponentInst.
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

// ComponentInst corresponds to a ComponentType that has been instantiated.
type ComponentInst struct {
	Type ComponentType // type of component
	Data interface{}   // data as returned by NewData
}
