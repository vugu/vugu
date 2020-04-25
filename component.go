package vugu

import (
	"bytes"
)

// BuildIn is the input to a Build call.
type BuildIn struct {

	// the overall build environment
	BuildEnv *BuildEnv

	// a stack of position hashes, the last one can be used by a component to get a unique hash for overall position
	PositionHashList []uint64
}

// CurrentPositionHash returns the hash value that can be used by a component to
// mix into it's own hash values to achieve uniqueness based on overall position
// in the output tree.  Basically you should XOR this with whatever unique reference
// ID within the component itself used during Build.  The purpose is to ensure
// that we use the same ID for the same component in the same position within
// the overall tree but a different one for the same component in a different position.
func (bi *BuildIn) CurrentPositionHash() uint64 {
	if len(bi.PositionHashList) == 0 {
		return 0
	}
	return bi.PositionHashList[len(bi.PositionHashList)-1]
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

// builderFunc is a Build-like function that implements Builder.
type builderFunc func(in *BuildIn) (out *BuildOut)

// Build implements Builder.
func (f builderFunc) Build(in *BuildIn) (out *BuildOut) { return f(in) }

// NewBuilderFunc returns a Builder from a function that is called for Build.
func NewBuilderFunc(f func(in *BuildIn) (out *BuildOut)) Builder {
	// NOTE: for now, all components have to be struct pointers, so we wrap
	// this function in a struct pointer.  Would be nice to fix this at some point.
	return &struct {
		Builder
	}{
		Builder: builderFunc(f),
	}
}

// BeforeBuilder can be implemented by components that need a chance to compute internal values
// after properties have been set but before Build is called.
type BeforeBuilder interface {
	BeforeBuild()
}
