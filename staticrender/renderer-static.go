package staticrender

import (
	"fmt"
	"io"
	"strings"

	// "golang.org/x/net/html"
	// "golang.org/x/net/html/atom"
	"github.com/vugu/html"
	"github.com/vugu/html/atom"
	"github.com/vugu/vugu"
)

// caller should be able to just specify a directory,
// or provide custom logic for how to give an exact io.Writer to write to for a given output path,
// should be callable a bunch of times for different URLs

func NewStaticRenderer(w io.Writer) *StaticRenderer {
	return &StaticRenderer{
		w: w,
	}
}

type StaticRenderer struct {
	w io.Writer
}

func (r *StaticRenderer) SetWriter(w io.Writer) {
	r.w = w
}

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

	var visit func(vgn *vugu.VGNode) (*html.Node, error)
	visit = func(vgn *vugu.VGNode) (*html.Node, error) {

		// if component then look up BuildOut for it and call renderOne again and return
		if vgn.Component != nil {
			cbo := br.ResultFor(vgn.Component)
			return r.renderOne(br, cbo)
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

			for _, nc := range nparts {
				n.AppendChild(nc)
			}

			// InnerHTML precludes other children
			return n, nil
		}

		// handle children
		for vgchild := vgn.FirstChild; vgchild != nil; vgchild = vgchild.NextSibling {
			nchild, err := visit(vgchild)
			if err != nil {
				return nil, err
			}
			n.AppendChild(nchild)
		}

		return n, nil
	}

	return visit(vgn)

}
