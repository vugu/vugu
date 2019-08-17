package vugu

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"path/filepath"
	"strings"

	"github.com/vugu/vugu/internal/htmlx"
	"github.com/vugu/vugu/internal/htmlx/atom"
)

func attrWithKey(n *htmlx.Node, key string) *htmlx.Attribute {
	for i := range n.Attr {
		if n.Attr[i].Key == key {
			return &n.Attr[i]
		}
	}
	return nil
}

// Parse2 is an experiment...
// r is the actual input, fname is only used to emit line directives
func (p *ParserGo) Parse(r io.Reader, fname string) error {

	state := &parseGoState{}

	inRaw, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	// use a tokenizer to peek at the first element and see if it's an HTML tag
	state.isFullHTML = false
	tmpZ := htmlx.NewTokenizer(bytes.NewReader(inRaw))
	for {
		tt := tmpZ.Next()
		if tt == htmlx.ErrorToken {
			return tmpZ.Err()
		}
		if tt != htmlx.StartTagToken { // skip over non-tags
			continue
		}
		t := tmpZ.Token()
		if t.Data == "html" {
			state.isFullHTML = true
			break
		}
		break
	}

	log.Printf("isFullHTML: %v", state.isFullHTML)

	if state.isFullHTML {

		n, err := htmlx.Parse(bytes.NewReader(inRaw))
		if err != nil {
			return err
		}
		state.docNodeList = append(state.docNodeList, n) // docNodeList is just this one item

	} else {

		nlist, err := htmlx.ParseFragment(bytes.NewReader(inRaw), &htmlx.Node{
			Type:     htmlx.ElementNode,
			DataAtom: atom.Div,
			Data:     "div",
		})
		if err != nil {
			return err
		}

		// only add elements
		for _, n := range nlist {
			if n.Type != htmlx.ElementNode {
				continue
			}
			state.docNodeList = append(state.docNodeList, n)
		}

		// // log.Printf("nlist = %#v", nlist)
		// for _, nl := range nlist {
		// 	log.Printf("nl.Data = %q", nl.Data)
		// }

		// if len(nlist) != 1 {
		// 	return fmt.Errorf("found %d fragment(s) instead of exactly 1", len(nlist))
		// }

		// n = nlist[0]

	}

	// err = htmlx.Render(os.Stdout, n)
	// if err != nil {
	// 	panic(err)
	// }

	// run n through the optimizer and convert large chunks of static elements into
	// vg-html attributes, this should provide a significiant performance boost for static HTML
	if !p.NoOptimizeStatic {
		for _, n := range state.docNodeList {
			err = compactNodeTree(n)
			if err != nil {
				return err
			}
		}
	}

	err = p.visitOverall(state)
	if err != nil {
		return err
	}

	// didFirstNode := false

	// var visit func(n *htmlx.Node) error
	// visit = func(n *htmlx.Node) error {

	// 	// log.Printf("n.Type = %v", n.Type)

	// 	if n.Type == htmlx.DocumentNode {
	// 		// nop
	// 	} else if n.Type == htmlx.ElementNode && n.Data == "script" {

	// 		// // script tag, determine if it's JS or Go
	// 		// typeAttr := attrWithKey(n, "type")
	// 		// if typeAttr == nil || typeAttr.Val == "application/javascript" {
	// 		// 	// for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
	// 		// 	// 	if childN.Type != htmlx.TextNode {
	// 		// 	// 		return fmt.Errorf("unexpected node type %v inside of script tag", childN.Type)
	// 		// 	// 	}
	// 		// 	// 	jsChunkList = append(jsChunkList, codeChunk{Line: childN.Line, Code: childN.Data})
	// 		// 	// }
	// 		// } else if typeAttr.Val == "application/x-go" {
	// 		// 	// for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
	// 		// 	// 	if childN.Type != htmlx.TextNode {
	// 		// 	// 		return fmt.Errorf("unexpected node type %v inside of script tag", childN.Type)
	// 		// 	// 	}
	// 		// 	// 	// if childN.Line > 0 {
	// 		// 	// 	// 	fmt.Fprintf(&goBuf, "//line %s:%d\n", fname, childN.Line)
	// 		// 	// 	// }
	// 		// 	// 	goBuf.WriteString(childN.Data)
	// 		// 	// }
	// 		// } else {
	// 		// 	return fmt.Errorf("unknown script type %q", typeAttr.Val)
	// 		// }

	// 		// // for script, this is it we're done with the visit
	// 		// return nil

	// 	} else if n.Type == htmlx.ElementNode && n.Data == "style" {

	// 		// // CSS
	// 		// for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
	// 		// 	if childN.Type != htmlx.TextNode {
	// 		// 		return fmt.Errorf("unexpected node type %v inside of style tag", childN.Type)
	// 		// 	}
	// 		// 	cssChunkList = append(cssChunkList, codeChunk{Line: childN.Line, Code: childN.Data})
	// 		// }

	// 		// // for style, this is it we're done with the visit
	// 		// return nil

	// 	} else {

	// 		// group the processing of this one node into a func so the defer's are called before moving onto the next sibling
	// 		err := func() error {

	// 			// // vg-for
	// 			// if forx := vgForExprx(n); forx != "" {
	// 			// 	// fmt.Fprintf(&buildBuf, "for /*line %s:%d*/%s {\n", fname, n.Line, forx)
	// 			// 	fmt.Fprintf(&buildBuf, "for %s {\n", forx)
	// 			// 	defer fmt.Fprintf(&buildBuf, "}\n")
	// 			// }

	// 			// // vg-if
	// 			// ife := vgIfExprx(n)
	// 			// if ife != "" {
	// 			// 	fmt.Fprintf(&buildBuf, "if %s {\n", ife)
	// 			// 	defer fmt.Fprintf(&buildBuf, "}\n")
	// 			// }

	// 			// if n.Type == htmlx.ElementNode && strings.Contains(n.Data, ":") {

	// 			// 	// component

	// 			// 	// dynamic attrs

	// 			// 	// component events

	// 			// } else {

	// 			// 	// regular element

	// 			// 	// if n.Line > 0 {
	// 			// 	// 	fmt.Fprintf(&buildBuf, "//line %s:%d\n", fname, n.Line)
	// 			// 	// }
	// 			// 	fmt.Fprintf(&buildBuf, "vgn = &vugu.VGNode{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}\n", n.Type, n.Data, staticVGAttrx(n.Attr))
	// 			// 	if didFirstNode {
	// 			// 		fmt.Fprintf(&buildBuf, "vgparent.AppendChild(vgn)\n") // if not root, make AppendChild call
	// 			// 	} else {
	// 			// 		fmt.Fprintf(&buildBuf, "vgout.Doc = vgn // Doc root for output\n") // for first element we need to assign as Doc on BuildOut
	// 			// 	}

	// 			// 	// dynamic attrs
	// 			// 	dynExprMap, dynExprMapKeys := dynamicVGAttrExprx(n)
	// 			// 	for _, k := range dynExprMapKeys {
	// 			// 		valExpr := dynExprMap[k]
	// 			// 		fmt.Fprintf(&buildBuf, "vgn.Attr = append(vgn.Attr, vugu.VGAttribute{Key:%q,Val:fmt.Sprint(%s)})\n", k, valExpr)
	// 			// 	}

	// 			// 	// vg-html
	// 			// 	htmlExpr := vgHTMLExprx(n)
	// 			// 	if htmlExpr != "" {
	// 			// 		fmt.Fprintf(&buildBuf, "{\nvghtml := %s; \nvgn.InnerHTML = &vghtml\n}\n", htmlExpr)
	// 			// 	}

	// 			// 	// DOM events
	// 			// 	eventMap, eventKeys := vgDOMEventExprsx(n)
	// 			// 	for _, k := range eventKeys {
	// 			// 		expr := eventMap[k]
	// 			// 		fmt.Fprintf(&buildBuf, "vgn.DOMEventHandlerSpecList = append(vgn.DOMEventHandlerSpecList, vugu.DOMEventHandlerSpec{\n")
	// 			// 		fmt.Fprintf(&buildBuf, "EventType: %q,\n", k)
	// 			// 		fmt.Fprintf(&buildBuf, "Func: func(event *vugu.DOMEvent) { %s },\n", expr)
	// 			// 		fmt.Fprintf(&buildBuf, "// TODO: implement capture, etc.\n")
	// 			// 		fmt.Fprintf(&buildBuf, "})\n")
	// 			// 	}

	// 			// }

	// 			// didFirstNode = true

	// 			return nil
	// 		}()
	// 		if err != nil {
	// 			return err
	// 		}

	// 	}

	// 	// if descending into a child we need to set the parent appropriately
	// 	if n.FirstChild != nil {
	// 		fmt.Fprintf(&buildBuf, "{\n")
	// 		fmt.Fprintf(&buildBuf, "vgparent := vgn; _ = vgparent\n") // vgparent set for this block to vgn
	// 		err := visit(n.FirstChild)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		fmt.Fprintf(&buildBuf, "}\n")
	// 	}

	// 	// siblings don't need special handling, they can just add to the same parent
	// 	if n.NextSibling != nil {
	// 		err := visit(n.NextSibling)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}

	// 	return nil
	// }

	// for _, n := range docNodeList {
	// 	err = visit(n)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// for _, chunk := range state.cssChunkList {
	// 	// fmt.Fprintf(&buildBuf, "    out.AppendCSS(/*line %s:%d*/%q)\n\n", fname, chunk.Line, chunk.Code)
	// 	fmt.Fprintf(&state.buildBuf, "    out.AppendCSS(%q)\n\n", chunk.Code)
	// }

	// for _, chunk := range state.jsChunkList {
	// 	// fmt.Fprintf(&buildBuf, "    out.AppendJS(/*line %s:%d*/%q)\n\n", fname, chunk.Line, chunk.Code)
	// 	fmt.Fprintf(&state.buildBuf, "    out.AppendJS(%q)\n\n", chunk.Code)
	// }

	// fmt.Fprintf(&state.buildBuf, "    return vgout, nil\n")
	// fmt.Fprintf(&state.buildBuf, "}\n\n")

	var buf bytes.Buffer
	// log.Printf("goBuf.Len == %v", goBuf.Len())
	buf.Write(state.goBuf.Bytes())
	buf.Write(state.buildBuf.Bytes())
	buf.Write(state.goBufBottom.Bytes())

	outPath := filepath.Join(p.OutDir, p.OutFile)
	// err = ioutil.WriteFile(outPath, buf.Bytes(), 0644)
	// if err != nil {
	// 	return err
	// }

	fo, err := p.gofmt(buf.String())
	if err != nil {

		// if the gofmt errors, we still attempt to write out the non-fmt'ed output to the file, to assist in debugging
		ioutil.WriteFile(outPath, buf.Bytes(), 0644)

		return err
	}

	err = ioutil.WriteFile(outPath, []byte(fo), 0644)
	if err != nil {
		return err
	}

	return nil
}

type codeChunk struct {
	Line   int
	Column int
	Code   string
}

type parseGoState struct {
	isFullHTML   bool
	docNodeList  []*htmlx.Node
	goBuf        bytes.Buffer // additional Go code (at top)
	buildBuf     bytes.Buffer // Build() method Go code (below)
	goBufBottom  bytes.Buffer // additional Go code that is put as the very last thing
	cssChunkList []codeChunk
	jsChunkList  []codeChunk
	outIsSet     bool // set to true when vgout.Out has been set for to the level node
}

// cases:
// - html
// - js
// - css
// - go code
// - top node
// - node

func (p *ParserGo) visitOverall(state *parseGoState) error {

	fmt.Fprintf(&state.goBuf, "package %s\n\n", p.PackageName)
	fmt.Fprintf(&state.goBuf, "// DO NOT EDIT: This file was generated by vugu. Please regenerate instead of editing or add additional code in a separate file.\n\n")
	fmt.Fprintf(&state.goBuf, "import %q\n", "fmt")
	fmt.Fprintf(&state.goBuf, "import %q\n", "reflect")
	fmt.Fprintf(&state.goBuf, "import %q\n", "github.com/vugu/vugu")
	fmt.Fprintf(&state.goBuf, "import js %q\n", "github.com/vugu/vugu/js")
	fmt.Fprintf(&state.goBuf, "\n")

	// TODO: we use a prefix like "vg" as our namespace; should document that user code should not use that prefix to avoid conflicts
	fmt.Fprintf(&state.buildBuf, "func (c *%s) Build(vgin *vugu.BuildIn) (vgout *vugu.BuildOut, vgreterr error) {\n", p.StructType)
	fmt.Fprintf(&state.buildBuf, "    \n")
	fmt.Fprintf(&state.buildBuf, "    vgout = &vugu.BuildOut{}\n")
	fmt.Fprintf(&state.buildBuf, "    \n")
	fmt.Fprintf(&state.buildBuf, "    var vgn *vugu.VGNode\n")
	// fmt.Fprintf(&buildBuf, "    var vgparent *vugu.VGNode\n")

	fmt.Fprintf(&state.goBufBottom, "// 'fix' unused imports\n")
	fmt.Fprintf(&state.goBufBottom, "var _ = fmt.Sprintf\n")
	fmt.Fprintf(&state.goBufBottom, "var _ = reflect.New\n")
	fmt.Fprintf(&state.goBufBottom, "var _ = js.ValueOf\n")
	fmt.Fprintf(&state.goBufBottom, "\n")

	// remove document node if present
	if len(state.docNodeList) == 1 && state.docNodeList[0].Type == htmlx.DocumentNode {
		state.docNodeList = []*htmlx.Node{state.docNodeList[0].FirstChild}
	}

	if state.isFullHTML {

		if len(state.docNodeList) != 1 {
			return fmt.Errorf("full HTML mode but not exactly 1 node found (found %d)", len(state.docNodeList))
		}
		err := p.visitHTML(state, state.docNodeList[0])
		if err != nil {
			return err
		}

	} else {

		for _, n := range state.docNodeList {

			// ignore comments
			if n.Type == htmlx.CommentNode {
				continue
			}

			if n.Type == htmlx.TextNode {

				// ignore whitespace text
				if strings.TrimSpace(n.Data) == "" {
					continue
				}

				// error on non-whitespace text
				return fmt.Errorf("unexpected text outside any element: %q", n.Data)

			}

			// must be an element at this point
			if n.Type != htmlx.ElementNode {
				return fmt.Errorf("unexpected node type %v; node=%#v", n.Type, n)
			}

			nodeName := strings.ToLower(n.Data)

			// script tag
			if nodeName == "script" {

				ty := attrWithKey(n, "type")
				if ty == nil {
					return fmt.Errorf("script tag without type attribute is not valid")
				}

				mt, _, _ := mime.ParseMediaType(ty.Val)

				// go code
				if mt == "application/x-go" {
					err := p.visitGo(state, n)
					if err != nil {
						return err
					}
					continue
				}

				// component js
				if mt == "application/javascript" {
					err := p.visitBuildJS(state, n)
					if err != nil {
						return err
					}
					continue
				}

				return fmt.Errorf("found script tag with invalid mime type %q", mt)

			}

			// component css
			if nodeName == "style" {
				err := p.visitBuildCSS(state, n)
				if err != nil {
					return err
				}
				continue
			}

			// top node

			// check for forbidden top level tags
			if nodeName == "head" ||
				nodeName == "body" {
				return fmt.Errorf("component cannot use %q as top level tag", nodeName)
			}

			err := p.visitTopNode(state, n)
			if err != nil {
				return err
			}
			continue

		}

	}

	for _, chunk := range state.cssChunkList {
		// fmt.Fprintf(&buildBuf, "    out.AppendCSS(/*line %s:%d*/%q)\n\n", fname, chunk.Line, chunk.Code)
		fmt.Fprintf(&state.buildBuf, "    out.AppendCSS(%q)\n\n", chunk.Code)
	}

	for _, chunk := range state.jsChunkList {
		// fmt.Fprintf(&buildBuf, "    out.AppendJS(/*line %s:%d*/%q)\n\n", fname, chunk.Line, chunk.Code)
		fmt.Fprintf(&state.buildBuf, "    out.AppendJS(%q)\n\n", chunk.Code)
	}

	fmt.Fprintf(&state.buildBuf, "    return vgout, nil\n")
	fmt.Fprintf(&state.buildBuf, "}\n\n")

	return nil
}

func (p *ParserGo) visitHTML(state *parseGoState, n *htmlx.Node) error {
	return fmt.Errorf("html tag not yet supported: %#v", n)
}

func (p *ParserGo) visitBuildJS(state *parseGoState, n *htmlx.Node) error {

	for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
		if childN.Type != htmlx.TextNode {
			return fmt.Errorf("unexpected node type %v inside of script tag", childN.Type)
		}
		state.jsChunkList = append(state.jsChunkList, codeChunk{Line: childN.Line, Code: childN.Data})
	}

	return nil
}

func (p *ParserGo) visitBuildCSS(state *parseGoState, n *htmlx.Node) error {

	// CSS
	for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
		if childN.Type != htmlx.TextNode {
			return fmt.Errorf("unexpected node type %v inside of style tag", childN.Type)
		}
		state.cssChunkList = append(state.cssChunkList, codeChunk{Line: childN.Line, Code: childN.Data})
	}

	return nil
}

func (p *ParserGo) visitGo(state *parseGoState, n *htmlx.Node) error {

	for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
		if childN.Type != htmlx.TextNode {
			return fmt.Errorf("unexpected node type %v inside of script tag", childN.Type)
		}
		// if childN.Line > 0 {
		// 	fmt.Fprintf(&goBuf, "//line %s:%d\n", fname, childN.Line)
		// }
		state.goBuf.WriteString(childN.Data)
	}

	return nil
}

// visitTopNode handles the "mount point"
func (p *ParserGo) visitTopNode(state *parseGoState, n *htmlx.Node) error {

	// handle the top element other than <html>

	err := p.visitNodeJustElement(state, n)
	if err != nil {
		return err
	}

	return nil
}

// visitNodeElementAndCtrl handles an element that supports vg-if, vg-for etc
func (p *ParserGo) visitNodeElementAndCtrl(state *parseGoState, n *htmlx.Node) error {

	// vg-for
	if forx := vgForExprx(n); forx != "" {
		// fmt.Fprintf(&buildBuf, "for /*line %s:%d*/%s {\n", fname, n.Line, forx)
		fmt.Fprintf(&state.buildBuf, "for %s {\n", forx)
		defer fmt.Fprintf(&state.buildBuf, "}\n")
	}

	// vg-if
	ife := vgIfExprx(n)
	if ife != "" {
		fmt.Fprintf(&state.buildBuf, "if %s {\n", ife)
		defer fmt.Fprintf(&state.buildBuf, "}\n")
	}

	err := p.visitNodeJustElement(state, n)
	if err != nil {
		return err
	}

	return nil
}

// visitNodeJustElement handles an element, ignoring any vg-if, vg-for (but it does handle vg-html - since that is not really "control" just a shorthand for it's contents)
func (p *ParserGo) visitNodeJustElement(state *parseGoState, n *htmlx.Node) error {

	// regular element

	// if n.Line > 0 {
	// 	fmt.Fprintf(&buildBuf, "//line %s:%d\n", fname, n.Line)
	// }

	fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}\n", n.Type, n.Data, staticVGAttrx(n.Attr))
	if state.outIsSet {
		fmt.Fprintf(&state.buildBuf, "vgparent.AppendChild(vgn)\n") // if not root, make AppendChild call
	} else {
		fmt.Fprintf(&state.buildBuf, "vgout.Out = append(vgout.Out, vgn) // root for output\n") // for first element we need to assign as Doc on BuildOut
		state.outIsSet = true
	}

	// dynamic attrs
	dynExprMap, dynExprMapKeys := dynamicVGAttrExprx(n)
	for _, k := range dynExprMapKeys {
		valExpr := dynExprMap[k]
		fmt.Fprintf(&state.buildBuf, "vgn.Attr = append(vgn.Attr, vugu.VGAttribute{Key:%q,Val:fmt.Sprint(%s)})\n", k, valExpr)
	}

	// vg-html
	htmlExpr := vgHTMLExprx(n)
	if htmlExpr != "" {
		fmt.Fprintf(&state.buildBuf, "{\nvghtml := %s; \nvgn.InnerHTML = &vghtml\n}\n", htmlExpr)
	}

	// DOM events
	eventMap, eventKeys := vgDOMEventExprsx(n)
	for _, k := range eventKeys {
		expr := eventMap[k]
		fmt.Fprintf(&state.buildBuf, "vgn.DOMEventHandlerSpecList = append(vgn.DOMEventHandlerSpecList, vugu.DOMEventHandlerSpec{\n")
		fmt.Fprintf(&state.buildBuf, "EventType: %q,\n", k)
		fmt.Fprintf(&state.buildBuf, "Func: func(event *vugu.DOMEvent) { %s },\n", expr)
		fmt.Fprintf(&state.buildBuf, "// TODO: implement capture, etc. mostly need to decide syntax\n")
		fmt.Fprintf(&state.buildBuf, "})\n")
	}

	if n.FirstChild != nil {

		fmt.Fprintf(&state.buildBuf, "{\n")
		fmt.Fprintf(&state.buildBuf, "vgparent := vgn; _ = vgparent\n") // vgparent set for this block to vgn

		// iterate over children
		for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

			// handle child according to type
			var err error
			switch {
			case childN.Type == htmlx.CommentNode:
				err = p.visitNodeComment(state, childN)
			case childN.Type == htmlx.TextNode:
				err = p.visitNodeText(state, childN)
			case childN.Type == htmlx.ElementNode:
				if strings.Contains(childN.Data, ":") {
					err = p.visitNodeComponentElement(state, childN)
				} else {
					err = p.visitNodeElementAndCtrl(state, childN)
				}
			default:
				return fmt.Errorf("child node of unknown type %v: %#v", childN.Type, childN)
			}

			if err != nil {
				return err
			}
		}

		fmt.Fprintf(&state.buildBuf, "}\n")

	}

	return nil
}

func (p *ParserGo) visitNodeText(state *parseGoState, n *htmlx.Node) error {

	fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{Type:vugu.VGNodeType(%d),Data:%q}\n", n.Type, n.Data)
	fmt.Fprintf(&state.buildBuf, "vgparent.AppendChild(vgn)\n")

	return nil
}

func (p *ParserGo) visitNodeComment(state *parseGoState, n *htmlx.Node) error {

	fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{Type:vugu.VGNodeType(%d),Data:%q}\n", n.Type, n.Data)
	fmt.Fprintf(&state.buildBuf, "vgparent.AppendChild(vgn)\n")

	return nil
}

// visitNodeComponentElement handles an element that is a call to a component
func (p *ParserGo) visitNodeComponentElement(state *parseGoState, n *htmlx.Node) error {

	nodeName := n.Data
	nodeNameParts := strings.Split(nodeName, ":")
	if len(nodeNameParts) != 2 {
		return fmt.Errorf("invalid component tag name %q must contain exactly one colon", nodeName)
	}

	// dynamic attrs

	// component events

	// slots

	return fmt.Errorf("component tag not yet supported (%q)", nodeName)
}
