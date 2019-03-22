package vugu

type ComponentType interface {
	TagName() string                                                      // HTML-compatible tag name, e.g. "demo-button"
	BuildVDOM(data interface{}) (vdom *VGNode, css *VGNode, reterr error) // based on the given data, build the VGNode tree
	InitData() (interface{}, error)                                       // initial data when component is instanciated
	// Invoke(name string, args ...interface{}) error

	// NOTES:
	// props???
	// template
	// data
	// methods
	// computed!? - probably best implemented purely in the data binding code, rather than as part of components
	// lifecycle hooks (created, mounted, etc.)
}

func New(ct ComponentType) (*ComponentInst, error) {
	data, err := ct.InitData()
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
