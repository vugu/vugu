package vugu

import (
	"bytes"
)

// BuildIn is the input to a Build call.
type BuildIn struct {
	BuildEnv *BuildEnv
}

// BuildOut is the output from a Build call.  It includes Out as the DOM elements
// produced, plus slices for CSS and JS elements.
type BuildOut struct {

	// output element(s) - usually just one that is parent to the rest but slots can have multiple
	Out []*VGNode

	// components that need to be built next - corresponding to each VGNode in Out with a Component value set,
	// we make the Builder populate this so BuildEnv doesn't have to traverse Out to gather up each node with a component
	// FIXME: should this be called ChildComponents, NextComponents? something to make it clear that these are pointers to more work to
	// do rather than something already done
	Components []Builder

	// optional CSS style or link tag(s)
	CSS []*VGNode

	// optional JS script tag(s)
	JS []*VGNode
}

// AppendCSS will append a unique node to CSS (nodes match exactly will not be added again).
func (b *BuildOut) AppendCSS(nlist ...*VGNode) {
	// FIXME: we really should consider doing this dedeplication with a map or something, but for now this works
nlistloop:
	for _, n := range nlist {
		for _, css := range b.CSS {
			if !ssNodeEq(n, css) {
				continue
			}
			if ssText(n) != ssText(css) {
				continue
			}
			// these are the same, skip duplicate add
			continue nlistloop
		}
		// not a duplicate, add it
		b.CSS = append(b.CSS, n)
	}
}

// AppendJS will append a unique node to JS (nodes match exactly will not be added again).
func (b *BuildOut) AppendJS(nlist ...*VGNode) {
nlistloop:
	for _, n := range nlist {
		for _, js := range b.JS {
			if !ssNodeEq(n, js) {
				continue
			}
			if ssText(n) != ssText(js) {
				continue
			}
			// these are the same, skip duplicate add
			continue nlistloop
		}
		// not a duplicate, add it
		b.JS = append(b.JS, n)
	}
}

// returns text content of child of script/style tag (assumes only one level of children and all text)
func ssText(n *VGNode) string {
	// special case 0 children
	if n.FirstChild == nil {
		return ""
	}
	// special case 1 child
	if n.FirstChild.NextSibling == nil {
		return n.FirstChild.Data
	}
	// more than one child
	var buf bytes.Buffer
	for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
		buf.WriteString(childN.Data)
	}
	return buf.String()
}

// script/style node compare - only care about Type, Data and Attributes, does not examine children
func ssNodeEq(n1, n2 *VGNode) bool {
	if n1.Type != n2.Type {
		return false
	}
	if n1.Data != n2.Data {
		return false
	}
	if len(n1.Attr) != len(n2.Attr) {
		return false
	}
	for i := 0; i < len(n1.Attr); i++ {
		if n1.Attr[i] != n2.Attr[i] {
			return false
		}
	}
	return true
}

// FIXME: CSS and JS are now full element nodes not just blocks of text (style, link, or script tags),
// and we should have some sort of AppendCSS and AppendJS methods that understand that and also perform
// deduplication.  There is no reason to wait and deduplicate later, we should be doing it here during
// the Build.

// func (b *BuildOut) AppendCSS(css string) {
// 	// FIXME: should we be deduplicating here??
// 	vgn := &VGNode{Type: ElementNode, Data: "style"}
// 	vgn.AppendChild(&VGNode{Type: TextNode, Data: css})
// 	b.CSS = append(b.CSS, vgn)
// }

// func (b *BuildOut) AppendJS(js string) {
// 	// FIXME: should we be deduplicating here??
// 	vgn := &VGNode{Type: ElementNode, Data: "script"}
// 	vgn.AppendChild(&VGNode{Type: TextNode, Data: js})
// 	b.JS = append(b.JS, vgn)
// }

// Builder is the interface that components implement.
type Builder interface {
	// Build is called to construct a tree of VGNodes corresponding to the DOM of this component.
	// Note that Build does not return an error because there is nothing the caller can do about
	// it except for stop rendering.  This way we force errors during Build to either be handled
	// internally or to explicitly result in a panic.
	Build(in *BuildIn) (out *BuildOut)
}

// BuilderFunc is a Build-like function that implements Builder.
type BuilderFunc func(in *BuildIn) (out *BuildOut)

// Build implements Builder.
func (f BuilderFunc) Build(in *BuildIn) (out *BuildOut) { return f(in) }

// BeforeBuilder can be implemented by components that need a chance to compute internal values
// after properties have been set but before Build is called.
type BeforeBuilder interface {
	BeforeBuild()
}

// // RegisteredComponentTypes returns a copy of the map of registered component types.
// func RegisteredComponentTypes() ComponentTypeMap {
// 	// make a copy
// 	ret := make(ComponentTypeMap, len(globalComponentTypeMap))
// 	for k, v := range globalComponentTypeMap {
// 		ret[k] = v
// 	}
// 	return ret
// }

// // RegisterComponentType can be called during init to register a component and make it available by default to the rest of the application.
// // A program may retrieved these by calling RegisteredComponentTypes() or it may choose to form its own set.  RegisterComponentType() just
// // makes it available in case that's useful.  It is good practice for components to register themselves in an init function.
// func RegisterComponentType(tagName string, ct ComponentType) {
// 	globalComponentTypeMap[tagName] = ct
// }

// var globalComponentTypeMap = make(ComponentTypeMap)

// // ComponentTypeMap is a map of the component tag name to a ComponentType.
// type ComponentTypeMap map[string]ComponentType

// // Props is a map that corresponds to property names and values.  The name
// // can correspond to an HTML attribute, or a property being passed in
// // during component instantiation, depending on the context.
// type Props map[string]interface{}

// // OrderedKeys returns the keys sorted alphabetically.
// func (p Props) OrderedKeys() []string {
// 	if p == nil {
// 		return nil
// 	}
// 	ret := make([]string, len(p))
// 	for k := range p {
// 		ret = append(ret, k)
// 	}
// 	sort.Strings(ret)
// 	return ret
// }

// // Clone makes a copy of the Props map, distinct from the original.
// func (p Props) Clone() Props {
// 	ret := make(Props, len(p))
// 	for k, v := range p {
// 		ret[k] = v
// 	}
// 	return ret
// }

// // Merge will copy everything from p2 into p and return p for ease of use.  Does not copy p.
// func (p Props) Merge(p2 Props) Props {
// 	for k, v := range p2 {
// 		p[k] = v
// 	}
// 	return p
// }

// // ComponentType is implemented by any type that wants to be a component.
// // The BuildVDOM method is called to generate the virtual DOM for a component; and this method
// // is usually code generated (by ParserGo) from a .vugu file.
// // NewData provides for specific behavior when a component is initialized.
// type ComponentType interface {
// 	BuildVDOM(data interface{}) (vdom *VGNode, css *VGNode, reterr error) // based on the given data, build the VGNode tree
// 	NewData(props Props) (interface{}, error)                             // initial data when component is instanciated
// }

// // New instantiates a component based on its ComponentType and a set of Props.
// // It essentially just calls NewData and returns a ComponentInst.
// func New(ct ComponentType, props Props) (*ComponentInst, error) {
// 	data, err := ct.NewData(props)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &ComponentInst{
// 		Type: ct,
// 		Data: data,
// 	}, nil
// }

// // ComponentInst corresponds to a ComponentType that has been instantiated.
// type ComponentInst struct {
// 	Type ComponentType // type of component
// 	Data interface{}   // data as returned by NewData
// }
