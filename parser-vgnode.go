package vugu

// ParserVGNode is a template parser that emits VGNodes directly.
// It only supports Go-template-style dynamic elements, and thus
// can be used without requiring the Go compiler.
type ParserVGNode struct {
	Result *VGNode
}

// func (p *ParserVGNode) Parse(r io.Reader) error {

// 	nodeList, err := html.ParseFragment(r, cruftBody)
// 	if err != nil {
// 		return err
// 	}

// 	// should be only one node with type Element and that's what we want
// 	var el *html.Node
// 	for _, n := range nodeList {
// 		if n.Type == html.ElementNode {
// 			if el != nil {
// 				return fmt.Errorf("found more than one element at root of component template")
// 			}
// 			el = n
// 		}
// 	}
// 	if el == nil {
// 		return fmt.Errorf("unable to find an element at root of component template")
// 	}

// 	n, err := htmlToVGNode(el)
// 	if err != nil {
// 		return err
// 	}

// 	p.Result = n

// 	return nil

// }

// // htmlToVGNode recursively converts html.Node to VGNode
// func htmlToVGNode(rootN *html.Node) (*VGNode, error) {

// 	ptrMap := make(map[*html.Node]*VGNode)

// 	var conv func(*html.Node) (*VGNode, error)
// 	conv = func(n *html.Node) (*VGNode, error) {

// 		if n == nil {
// 			return nil, nil
// 		}

// 		// see if it's already in map, if so just return it
// 		vgn := ptrMap[n]
// 		if vgn != nil {
// 			return vgn, nil
// 		}

// 		var err error
// 		vgn = &VGNode{}
// 		// assign this first thing, so that everything below when it recurses will just point to the same instance
// 		ptrMap[n] = vgn

// 		// for all node pointers we recursively call conv, which will convert them or just return the pointer if already done
// 		// Parent
// 		vgn.Parent, err = conv(n.Parent)
// 		if err != nil {
// 			return vgn, err
// 		}
// 		// FirstChild
// 		vgn.FirstChild, err = conv(n.FirstChild)
// 		if err != nil {
// 			return vgn, err
// 		}
// 		// LastChild
// 		vgn.LastChild, err = conv(n.LastChild)
// 		if err != nil {
// 			return vgn, err
// 		}
// 		// PrevSibling
// 		vgn.PrevSibling, err = conv(n.PrevSibling)
// 		if err != nil {
// 			return vgn, err
// 		}
// 		// NextSibling
// 		vgn.NextSibling, err = conv(n.NextSibling)
// 		if err != nil {
// 			return vgn, err
// 		}

// 		// copy the other type and attr info
// 		vgn.Type = VGNodeType(n.Type)
// 		vgn.DataAtom = VGAtom(n.DataAtom)
// 		vgn.Data = n.Data
// 		vgn.Namespace = n.Namespace

// 		for _, nAttr := range n.Attr {
// 			switch {
// 			case nAttr.Key == "vg-if":
// 				vgn.VGIf = attrFromHtml(nAttr)
// 			case nAttr.Key == "vg-range":
// 				vgn.VGRange = attrFromHtml(nAttr)
// 			case strings.HasPrefix(nAttr.Key, ":"):
// 				vgn.BindAttr = append(vgn.BindAttr, attrFromHtml(nAttr))
// 			case strings.HasPrefix(nAttr.Key, "@"):
// 				vgn.EventAttr = append(vgn.EventAttr, attrFromHtml(nAttr))
// 			default:
// 				vgn.Attr = append(vgn.Attr, attrFromHtml(nAttr))
// 			}
// 		}

// 		// if len(n.Attr) > 0 {
// 		// 	vgn.Attr = make([]VGAttribute, 0, len(n.Attr))
// 		// 	for _, a := range n.Attr {
// 		// 		vgn.Attr = append(vgn.Attr, VGAttribute{
// 		// 			Namespace: a.Namespace,
// 		// 			Key:       a.Key,
// 		// 			Val:       a.Val,
// 		// 		})
// 		// 	}
// 		// }

// 		// now extract out our VG-specific stuff

// 		// log.Printf("vgn = %#v", vgn)

// 		return vgn, nil
// 	}
// 	return conv(rootN)

// }
