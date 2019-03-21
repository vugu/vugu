package vugu

// // StaticHTMLEnv is an environment that renders to static HTML.  Can be used to implement "server side rendering"
// // as well as during testing.
// type StaticHTMLEnv struct {
// 	ComponentTypeMap map[string]ComponentType // TODO: probably make this it's own type and have a global instance where things can register
// 	// rootInst         *ComponentInst
// 	Out io.Writer
// }

// func (e *StaticHTMLEnv) SetOut(w io.Writer) {
// 	e.Out = w
// }

// func (e *StaticHTMLEnv) SetComponentType(ct ComponentType) {
// 	e.ComponentTypeMap[ct.TagName()] = ct
// }

// // Render is equivalent to calling e.RenderTo(e.Out, c)
// func (e *StaticHTMLEnv) Render(c *ComponentInst) error {
// 	return e.RenderTo(e.Out, c)
// }

// func (e *StaticHTMLEnv) RenderTo(out io.Writer, c *ComponentInst) error {

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
