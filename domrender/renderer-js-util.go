package domrender

import "github.com/vugu/vugu"

// namespaceToURI resolves the given namespaces to the URI with the specifications
func namespaceToURI(namespace string) string {
	switch namespace {
	case "html":
		return "http://www.w3.org/1999/xhtml"
	case "math":
		return "http://www.w3.org/1998/Math/MathML"
	case "svg":
		return "http://www.w3.org/2000/svg"
	case "xlink":
		return "http://www.w3.org/1999/xlink"
	case "xml":
		return "http://www.w3.org/XML/1998/namespace"
	case "xmlns":
		return "http://www.w3.org/2000/xmlns/"
	default:
		return ""
	}
}

type renderedCtx struct {
	eventEnv vugu.EventEnv
	first    bool
}

// EventEnv implements RenderedCtx by returning the EventEnv.
func (c *renderedCtx) EventEnv() vugu.EventEnv {
	return c.eventEnv
}

// First returns true for the first render and otherwise false.
func (c *renderedCtx) First() bool {
	return c.first
}

type rendered0 interface {
	Rendered()
}
type rendered1 interface {
	Rendered(ctx vugu.RenderedCtx)
}

func invokeRendered(c interface{}, rctx *renderedCtx) {
	if i, ok := c.(rendered0); ok {
		i.Rendered()
	} else if i, ok := c.(rendered1); ok {
		i.Rendered(rctx)
	}
}
