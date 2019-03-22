package vugu

import (
	"bytes"
	"io"

	"golang.org/x/net/html/atom"

	"golang.org/x/net/html"
)

// StaticHTMLEnv is an environment that renders to static HTML.  Can be used to implement "server side rendering"
// as well as during testing.
type StaticHTMLEnv struct {
	// ComponentTypeMap map[string]ComponentType // TODO: probably make this it's own type and have a global instance where things can register
	// rootInst         *ComponentInst
	Out io.Writer
}

// NewStaticHTMLEnv returns a new instance of StaticHTMLEnv with Out set.
func NewStaticHTMLEnv(out io.Writer) *StaticHTMLEnv {
	return &StaticHTMLEnv{
		Out: out,
	}
}

// func (e *StaticHTMLEnv) SetOut(w io.Writer) {
// 	e.Out = w
// }

// func (e *StaticHTMLEnv) SetComponentType(ct ComponentType) {
// 	e.ComponentTypeMap[ct.TagName()] = ct
// }

// Render is equivalent to calling e.RenderTo(e.Out, c)
func (e *StaticHTMLEnv) Render(c *ComponentInst) error {
	return e.RenderTo(e.Out, c)
}

func (e *StaticHTMLEnv) RenderTo(out io.Writer, c *ComponentInst) error {

	vdom, css, err := c.Type.BuildVDOM(c.Data)
	if err != nil {
		return err
	}

	// The basic strategy is to build an equivalent html.Node tree from our vdom, expanding InnerHTML along
	// the way, and then tell the html package to write it out

	// TODO: output css
	_ = css

	ptrMap := make(map[*VGNode]*html.Node)

	var conv func(*VGNode) (*html.Node, error)
	conv = func(vgn *VGNode) (*html.Node, error) {

		if vgn == nil {
			return nil, nil
		}

		// see if it's already in map, if so just return it
		if n := ptrMap[vgn]; n != nil {
			return n, nil
		}

		var err error
		n := &html.Node{}
		// assign this first thing, so that everything below when it recurses will just point to the same instance
		ptrMap[vgn] = n

		// for all node pointers we recursively call conv, which will convert them or just return the pointer if already done
		// Parent
		n.Parent, err = conv(vgn.Parent)
		if err != nil {
			return n, err
		}
		// FirstChild
		n.FirstChild, err = conv(vgn.FirstChild)
		if err != nil {
			return n, err
		}
		// LastChild
		n.LastChild, err = conv(vgn.LastChild)
		if err != nil {
			return n, err
		}
		// PrevSibling
		n.PrevSibling, err = conv(vgn.PrevSibling)
		if err != nil {
			return n, err
		}
		// NextSibling
		n.NextSibling, err = conv(vgn.NextSibling)
		if err != nil {
			return n, err
		}

		// copy the other type and attr info
		n.Type = html.NodeType(vgn.Type)
		n.DataAtom = atom.Atom(vgn.DataAtom)
		n.Data = vgn.Data
		n.Namespace = vgn.Namespace

		for _, vgnAttr := range vgn.Attr {
			n.Attr = append(n.Attr, html.Attribute{Namespace: vgnAttr.Namespace, Key: vgnAttr.Key, Val: vgnAttr.Val})
		}

		// parse and expand InnerHTML if present
		if vgn.InnerHTML != "" {

			innerNs, err := html.ParseFragment(bytes.NewReader([]byte(vgn.InnerHTML)), cruftBody)
			if err != nil {
				return nil, err
			}
			// FIXME: do we just append all of this, what about case where there is already something inside?
			for _, innerN := range innerNs {
				n.AppendChild(innerN)
			}

		}

		return n, nil
	}
	outn, err := conv(vdom)
	if err != nil {
		return err
	}
	// log.Printf("outn: %#v", outn)

	err = html.Render(out, outn)
	if err != nil {
		return err
	}

	return nil
}

// 	data := c.Data
// 	ct := c.Type
// 	rootNode, err := ct.Template()
// 	if err != nil {
// 		return err
// 	}

// 	var w func(vgn *VGNode) (*html.Node, error)
// 	w = func(vgn *VGNode) (*html.Node, error) {
// 		// walker(vgn *VGNode, data interface{}, itemData interface{})

// 		// check for v-if, if not truthy then don't output element

// 		if vgn.VGIf.Key != "" {
// 			s, _, err := tmplRun(`{{if `+vgn.VGIf.Val+`}}#{{end}}`, nil, data, nil)
// 			if err != nil {
// 				return nil, err
// 			}
// 			if s == "" { // if output was empty, it means skip this node, vg-if was false
// 				return nil, nil
// 			}
// 		}

// 		// TODO: check for v-range, render to get refs and then loop over each one and call walk, each with the same node but different element data

// 		// TODO: check for component and if match create instance if not exist, bind attrs get done as refs and turn into props, same with static attrs

// 		// TODO: else if not component, recurse into child nodes and w() them

// 		var n html.Node

// 		// TODO: check for bind attrs and eval each one and set on output node as text

// 		// TODO: what about blocks of {{stuff}}

// 		// for static html we ignore the events

// 		return &n, nil
// 	}

// 	// get root element and run through walker
// 	n, err := w(rootNode)
// 	if err != nil {
// 		return err
// 	}

// 	err = html.Render(out, n)
// 	if err != nil {
// 		return err
// 	}

// 	return nil

// }
