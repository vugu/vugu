package vugu

import "fmt"

// // Env specifies the common methods for environment implementations.
// // See JSEnv and StaticHtmlEnv for implementations.
// type Env interface {
// 	RegisterComponentType(tagName string, ct ComponentType)
// 	Render() error
// }

func NewBuildEnv(root Builder) (*BuildEnv, error) {
	return &BuildEnv{root: root}, nil
}

// BuildEnv is the environment used when building virtual DOM.
type BuildEnv struct {
	root Builder // FIXME: does this even belong here?  BuildEnv keeps track of components created but why does it need a reference to the root component?
	// Maybe BuildRoot is just a naken func?
}

func (e *BuildEnv) BuildRoot() (*BuildOut, error) {
	panic(fmt.Errorf("not yet implemented"))
}

func (e *BuildEnv) ComponentFor(n *VGNode) (Builder, error) {
	panic(fmt.Errorf("not yet implemented"))
}

func (e *BuildEnv) SetComponentFor(n *VGNode, c Builder) error {
	panic(fmt.Errorf("not yet implemented"))
}
