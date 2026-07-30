package staticrender

import (
	"fmt"
	"io"
	"strings"
	"sync"

	// "golang.org/x/net/html"
	// "golang.org/x/net/html/atom"
	"github.com/vugu/html"
	"github.com/vugu/html/atom"
	"github.com/vugu/vugu"
)

// caller should be able to just specify a directory,
// or provide custom logic for how to give an exact io.Writer to write to for a given output path,
// should be callable a bunch of times for different URLs

// New returns a new instance.  w may be nil as long as SetWriter is called with a valid value before rendering.
func New(w io.Writer) *StaticRenderer {
	return &StaticRenderer{
		w: w,
	}
}

// StaticRenderer provides rendering as static HTML to an io.Writer.
type StaticRenderer struct {
	w io.Writer
}

// SetWriter assigns the Writer to be used for subsequent calls to Render.
// A Writer must be assigned before rendering (either with this method or in New).
func (r *StaticRenderer) SetWriter(w io.Writer) {
	r.w = w
}

// Render will perform a static render of the given BuildResults and write it to the writer assigned.
func (r *StaticRenderer) Render(buildResults *vugu.BuildResults) error {

	n, err := r.renderOne(buildResults, buildResults.Out)
	if err != nil {
		return err
	}

	err = html.Render(r.w, n)
	if err != nil {
		return err
	}

	return nil

}

func (r *StaticRenderer) renderOne(br *vugu.BuildResults, bo *vugu.BuildOut) (*html.Node, error) {

	if len(bo.Out) != 1 {
		return nil, fmt.Errorf("BuildOut must contain exactly one element in Out")
	}

	vgn := bo.Out[0]

	var visit func(vgn *vugu.VGNode) ([]*html.Node, error)
	visit = func(vgn *vugu.VGNode) ([]*html.Node, error) {

		// log.Printf("vgn: %#v", vgn)

		// if component then look up BuildOut for it and call renderOne again and return
		if vgn.Component != nil {
			cbo := br.ResultFor(vgn.Component)
			retn, err := r.renderOne(br, cbo)
			if err != nil {
				return nil, err
			}
			return []*html.Node{retn}, nil
			// if len(retn) != 1 {
			// 	return nil, fmt.Errorf("StaticRenderer.renderOne component renderOne returned unexpected %d nodes", len(retn))
			// }
			// return retn[0], nil
		}

		// if template then just traverse the children directly and return them in a series, omitting vgn
		if vgn.IsTemplate() {
			var retn []*html.Node
			for vgchild := vgn.FirstChild; vgchild != nil; vgchild = vgchild.NextSibling {
				nchildren, err := visit(vgchild)
				if err != nil {
					return nil, err
				}
				retn = append(retn, nchildren...)
			}
			return retn, nil
		}

		// no component set, continue with building this node
		n := &html.Node{}
		n.Type = html.NodeType(vgn.Type)           // type numbers are the same
		n.Data = vgn.Data                          // copy Data over
		n.DataAtom = atom.Lookup([]byte(vgn.Data)) // lookup atom
		for _, vgattr := range vgn.Attr {
			n.Attr = append(n.Attr, html.Attribute{Key: vgattr.Key, Val: vgattr.Val})
		}

		// handle InnerHTML

		if vgn.InnerHTML != nil {

			nparts, err := html.ParseFragment(strings.NewReader(*vgn.InnerHTML), n)
			if err != nil {
				return nil, err
			}

			// for _, nc := range nparts {
			// 	n.AppendChild(nc)
			// }
			appendChildren(n, nparts)

			// InnerHTML precludes other children
			return []*html.Node{n}, nil
		}

		// handle children
		for vgchild := vgn.FirstChild; vgchild != nil; vgchild = vgchild.NextSibling {
			nchildren, err := visit(vgchild)
			if err != nil {
				return nil, err
			}
			// n.AppendChild(nchildren)
			appendChildren(n, nchildren)
		}

		// special case for <head>, we need to emit the CSS here as that is separate
		// (Vugu build output does not always have a head tag and multiple components
		// can each emit it, so we have to keep things like CSS separate)
		if n.Type == html.ElementNode && n.Data == "head" {
			for _, css := range bo.CSS {

				// convert each one
				nchildren, err := visit(css)
				if err != nil {
					return nil, err
				}
				// n.AppendChild(n2)
				appendChildren(n, nchildren)
			}
		}

		// special case to append JS to end of <body>
		if n.Type == html.ElementNode && n.Data == "body" {
			for _, js := range bo.JS {

				// convert each one
				nchildren, err := visit(js)
				if err != nil {
					return nil, err
				}
				// n.AppendChild(n2)
				appendChildren(n, nchildren)
			}
		}

		return []*html.Node{n}, nil
	}

	nret, err := visit(vgn)
	if err != nil {
		return nil, err
	}
	if len(nret) != 1 {
		return nil, fmt.Errorf("StaticRenderer.renderOne visit returned unexpected %d nodes", len(nret))
	}
	return nret[0], nil
}

func appendChildren(parent *html.Node, children []*html.Node) {
	for _, c := range children {
		parent.AppendChild(c)
	}
}

// EventEnv returns a simple EventEnv implementation suitable for use with the static renderer.
func (r *StaticRenderer) EventEnv() *RWMutexEventEnv {
	return &RWMutexEventEnv{}
}

// RWMutexEventEnv implements EventEnv interface with a mutex but no render behavior.
// (Because static rendering is driven by the caller, there is no way for dynamic logic
// to re-render a static page, so UnlockRender is the same as UnlockOnly.)
type RWMutexEventEnv struct {
	sync.RWMutex
}

// UnlockOnly will release the write lock
func (ee *RWMutexEventEnv) UnlockOnly() { ee.Unlock() }

// UnlockRender is an alias for UnlockOnly.
func (ee *RWMutexEventEnv) UnlockRender() { ee.Unlock() }
