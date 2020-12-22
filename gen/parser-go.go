package gen

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	// "github.com/vugu/vugu/internal/htmlx"
	// "github.com/vugu/vugu/internal/htmlx/atom"
	// "golang.org/x/net/html"
	// "golang.org/x/net/html/atom"
	"github.com/vugu/html"
	"github.com/vugu/html/atom"
	"github.com/vugu/vugu"
)

// ParserGo is a template parser that emits Go source code that will construct the appropriately wired VGNodes.
type ParserGo struct {
	PackageName string // name of package to use at top of files
	StructType  string // just the struct name, no "*" (replaces ComponentType and DataType)
	// ComponentType string // just the struct name, no "*"
	// DataType      string // just the struct name, no "*"
	OutDir  string // output dir
	OutFile string // output file name with ".go" suffix

	NoOptimizeStatic bool // set to true to disable optimization of static blocks of HTML into vg-html expressions
	TinyGo           bool // set to true to enable TinyGo compatability changes to the generated code
}

func gofmt(pgm string) (string, error) {

	// build up command to run
	cmd := exec.Command("gofmt")

	// I need to capture output
	var fmtOutput bytes.Buffer
	cmd.Stderr = &fmtOutput
	cmd.Stdout = &fmtOutput

	// also set up input pipe
	read, write := io.Pipe()
	defer write.Close() // make sure this always gets closed, it is safe to call more than once
	cmd.Stdin = read

	// copy down environment variables
	cmd.Env = os.Environ()
	// force wasm,js target
	cmd.Env = append(cmd.Env, "GOOS=js")
	cmd.Env = append(cmd.Env, "GOARCH=wasm")

	// start gofmt
	if err := cmd.Start(); err != nil {
		return pgm, fmt.Errorf("can't run gofmt: %v", err)
	}

	// stream in the raw source
	if _, err := write.Write([]byte(pgm)); err != nil && err != io.ErrClosedPipe {
		return pgm, fmt.Errorf("gofmt failed: %v", err)
	}

	write.Close()

	// wait until gofmt is done
	if err := cmd.Wait(); err != nil {
		return pgm, fmt.Errorf("go fmt error %v; full output: %s", err, string(fmtOutput.Bytes()))
	}

	return string(fmtOutput.Bytes()), nil
}

// Parse is an experiment...
// r is the actual input, fname is only used to emit line directives
func (p *ParserGo) Parse(r io.Reader, fname string) error {

	state := &parseGoState{}

	inRaw, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	// use a tokenizer to peek at the first element and see if it's an HTML tag
	state.isFullHTML = false
	tmpZ := html.NewTokenizer(bytes.NewReader(inRaw))
	for {
		tt := tmpZ.Next()
		if tt == html.ErrorToken {
			return tmpZ.Err()
		}
		if tt != html.StartTagToken { // skip over non-tags
			continue
		}
		t := tmpZ.Token()
		if t.Data == "html" {
			state.isFullHTML = true
			break
		}
		break
	}

	// log.Printf("isFullHTML: %v", state.isFullHTML)

	if state.isFullHTML {

		n, err := html.Parse(bytes.NewReader(inRaw))
		if err != nil {
			return err
		}
		state.docNodeList = append(state.docNodeList, n) // docNodeList is just this one item

	} else {

		nlist, err := html.ParseFragment(bytes.NewReader(inRaw), &html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Div,
			Data:     "div",
		})
		if err != nil {
			return err
		}

		// only add elements
		for _, n := range nlist {
			if n.Type != html.ElementNode {
				continue
			}
			// log.Printf("FRAGMENT: %#v", n)
			state.docNodeList = append(state.docNodeList, n)
		}

	}

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

	// log.Printf("parsed document looks like so upon start of parsing:")
	// for i, n := range state.docNodeList {
	// 	var buf bytes.Buffer
	// 	err := html.Render(&buf, n)
	// 	if err != nil {
	// 		return fmt.Errorf("error during debug render: %v", err)
	// 	}
	// 	log.Printf("state.docNodeList[%d]:\n%s", i, buf.Bytes())
	// }

	err = p.visitOverall(state)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	// log.Printf("goBuf.Len == %v", goBuf.Len())
	buf.Write(state.goBuf.Bytes())
	buf.Write(state.buildBuf.Bytes())
	buf.Write(state.goBufBottom.Bytes())

	outPath := filepath.Join(p.OutDir, p.OutFile)

	fo, err := gofmt(buf.String())
	if err != nil {

		// if the gofmt errors, we still attempt to write out the non-fmt'ed output to the file, to assist in debugging
		ioutil.WriteFile(outPath, buf.Bytes(), 0644)

		return err
	}

	// run the import deduplicator
	var dedupedBuf bytes.Buffer
	err = dedupImports(bytes.NewReader([]byte(fo)), &dedupedBuf, p.OutFile)
	if err != nil {
		return err
	}

	// write to final output file
	err = ioutil.WriteFile(outPath, dedupedBuf.Bytes(), 0644)
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
	isFullHTML  bool         // is the first node an <html> tag
	docNodeList []*html.Node // top level nodes parsed out of source file
	goBuf       bytes.Buffer // additional Go code (at top)
	buildBuf    bytes.Buffer // Build() method Go code (below)
	goBufBottom bytes.Buffer // additional Go code that is put as the very last thing
	// cssChunkList []codeChunk
	// jsChunkList  []codeChunk
	outIsSet bool // set to true when vgout.Out has been set for to the level node
}

func (p *ParserGo) visitOverall(state *parseGoState) error {

	fmt.Fprintf(&state.goBuf, "package %s\n\n", p.PackageName)
	fmt.Fprintf(&state.goBuf, "// Code generated by vugu via vugugen. Please regenerate instead of editing or add additional code in a separate file. DO NOT EDIT.\n\n")
	fmt.Fprintf(&state.goBuf, "import %q\n", "fmt")
	fmt.Fprintf(&state.goBuf, "import %q\n", "reflect")
	fmt.Fprintf(&state.goBuf, "import %q\n", "github.com/vugu/vjson")
	fmt.Fprintf(&state.goBuf, "import %q\n", "github.com/vugu/vugu")
	fmt.Fprintf(&state.goBuf, "import js %q\n", "github.com/vugu/vugu/js")
	fmt.Fprintf(&state.goBuf, "\n")

	// TODO: we use a prefix like "vg" as our namespace; should document that user code should not use that prefix to avoid conflicts
	fmt.Fprintf(&state.buildBuf, "func (c *%s) Build(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {\n", p.StructType)
	fmt.Fprintf(&state.buildBuf, "    \n")
	fmt.Fprintf(&state.buildBuf, "    vgout = &vugu.BuildOut{}\n")
	fmt.Fprintf(&state.buildBuf, "    \n")
	fmt.Fprintf(&state.buildBuf, "    var vgiterkey interface{}\n")
	fmt.Fprintf(&state.buildBuf, "    _ = vgiterkey\n")
	fmt.Fprintf(&state.buildBuf, "    var vgn *vugu.VGNode\n")
	// fmt.Fprintf(&buildBuf, "    var vgparent *vugu.VGNode\n")

	// NOTE: Use things that are lightweight here - e.g. don't do var _ = fmt.Sprintf because that brings in all of the
	// (possibly quite large) formatting code, which might otherwise be avoided.
	fmt.Fprintf(&state.goBufBottom, "// 'fix' unused imports\n")
	fmt.Fprintf(&state.goBufBottom, "var _ fmt.Stringer\n")
	fmt.Fprintf(&state.goBufBottom, "var _ reflect.Type\n")
	fmt.Fprintf(&state.goBufBottom, "var _ vjson.RawMessage\n")
	fmt.Fprintf(&state.goBufBottom, "var _ js.Value\n")
	fmt.Fprintf(&state.goBufBottom, "\n")

	// remove document node if present
	if len(state.docNodeList) == 1 && state.docNodeList[0].Type == html.DocumentNode {
		state.docNodeList = []*html.Node{state.docNodeList[0].FirstChild}
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

		gotTopNode := false

		for _, n := range state.docNodeList {

			// ignore comments
			if n.Type == html.CommentNode {
				continue
			}

			if n.Type == html.TextNode {

				// ignore whitespace text
				if strings.TrimSpace(n.Data) == "" {
					continue
				}

				// error on non-whitespace text
				return fmt.Errorf("unexpected text outside any element: %q", n.Data)

			}

			// must be an element at this point
			if n.Type != html.ElementNode {
				return fmt.Errorf("unexpected node type %v; node=%#v", n.Type, n)
			}

			if isScriptOrStyle(n) {

				err := p.visitScriptOrStyle(state, n)
				if err != nil {
					return err
				}
				continue
			}

			if gotTopNode {
				return fmt.Errorf("Found more than one top level element: %s", n.Data)
			}
			gotTopNode = true

			// handle top node

			// check for forbidden top level tags
			nodeName := strings.ToLower(n.Data)
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

	// for _, chunk := range state.cssChunkList {
	// 	// fmt.Fprintf(&buildBuf, "    out.AppendCSS(/*line %s:%d*/%q)\n\n", fname, chunk.Line, chunk.Code)
	// 	// fmt.Fprintf(&state.buildBuf, "    out.AppendCSS(%q)\n\n", chunk.Code)
	// 	_ = chunk
	// 	panic("need to append whole node, not AppendCSS")
	// }

	// for _, chunk := range state.jsChunkList {
	// 	// fmt.Fprintf(&buildBuf, "    out.AppendJS(/*line %s:%d*/%q)\n\n", fname, chunk.Line, chunk.Code)
	// 	// fmt.Fprintf(&state.buildBuf, "    out.AppendJS(%q)\n\n", chunk.Code)
	// 	_ = chunk
	// 	panic("need to append whole node, not AppendJS")
	// }

	fmt.Fprintf(&state.buildBuf, "    return vgout\n")
	fmt.Fprintf(&state.buildBuf, "}\n\n")

	return nil
}

func (p *ParserGo) visitHTML(state *parseGoState, n *html.Node) error {

	pOutputTag(state, n)
	// fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}\n", n.Type, n.Data, staticVGAttr(n.Attr))
	// fmt.Fprintf(&state.buildBuf, "vgout.Out = append(vgout.Out, vgn) // root for output\n") // for first element we need to assign as Doc on BuildOut
	// state.outIsSet = true

	// dynamic attrs
	writeDynamicAttributes(state, n)

	fmt.Fprintf(&state.buildBuf, "{\n")
	fmt.Fprintf(&state.buildBuf, "vgparent := vgn; _ = vgparent\n") // vgparent set for this block to vgn

	for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

		if childN.Type != html.ElementNode {
			continue
		}

		var err error
		if strings.ToLower(childN.Data) == "head" {
			err = p.visitHead(state, childN)
		} else if strings.ToLower(childN.Data) == "body" {
			err = p.visitBody(state, childN)
		} else {
			return fmt.Errorf("unknown tag inside html %q", childN.Data)

		}

		if err != nil {
			return err
		}

	}

	fmt.Fprintf(&state.buildBuf, "}\n")

	return nil
}

func (p *ParserGo) visitHead(state *parseGoState, n *html.Node) error {

	pOutputTag(state, n)
	// fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}\n", n.Type, n.Data, staticVGAttr(n.Attr))
	// fmt.Fprintf(&state.buildBuf, "vgout.Out = append(vgout.Out, vgn) // root for output\n") // for first element we need to assign as Doc on BuildOut
	// state.outIsSet = true

	// dynamic attrs
	writeDynamicAttributes(state, n)

	fmt.Fprintf(&state.buildBuf, "{\n")
	fmt.Fprintf(&state.buildBuf, "vgparent := vgn; _ = vgparent\n") // vgparent set for this block to vgn

	for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

		if isScriptOrStyle(childN) {
			err := p.visitScriptOrStyle(state, childN)
			if err != nil {
				return err
			}
			continue
		}

		err := p.visitDefaultByType(state, childN)
		if err != nil {
			return err
		}

	}

	fmt.Fprintf(&state.buildBuf, "}\n")

	return nil

}

func (p *ParserGo) visitBody(state *parseGoState, n *html.Node) error {

	pOutputTag(state, n)
	// fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}\n", n.Type, n.Data, staticVGAttr(n.Attr))
	// fmt.Fprintf(&state.buildBuf, "vgout.Out = append(vgout.Out, vgn) // root for output\n") // for first element we need to assign as Doc on BuildOut
	// state.outIsSet = true

	// dynamic attrs
	writeDynamicAttributes(state, n)

	fmt.Fprintf(&state.buildBuf, "{\n")
	fmt.Fprintf(&state.buildBuf, "vgparent := vgn; _ = vgparent\n") // vgparent set for this block to vgn

	foundMountEl := false

	for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

		// ignore whitespace and comments directly in body
		if childN.Type != html.ElementNode {
			continue
		}

		if isScriptOrStyle(childN) {
			err := p.visitScriptOrStyle(state, childN)
			if err != nil {
				return err
			}
			continue
		}

		if foundMountEl {
			return fmt.Errorf("element %q found after we already have a mount element", childN.Data)
		}
		foundMountEl = true

		err := p.visitDefaultByType(state, childN)
		if err != nil {
			return err
		}

	}

	fmt.Fprintf(&state.buildBuf, "}\n")

	return nil

}

// visitScriptOrStyle calls visitJS, visitCSS or visitGo accordingly,
// will error if the node does not correspond to one of those
func (p *ParserGo) visitScriptOrStyle(state *parseGoState, n *html.Node) error {

	nodeName := strings.ToLower(n.Data)

	// script tag
	if nodeName == "script" {

		var mt string // mime type

		ty := attrWithKey(n, "type")
		if ty == nil {
			// return fmt.Errorf("script tag without type attribute is not valid")
			mt = ""
		} else {
			// tinygo support: just split on semi, don't need to import mime package
			// mt, _, _ = mime.ParseMediaType(ty.Val)
			mt = strings.Split(strings.TrimSpace(ty.Val), ";")[0]
		}

		// go code
		if mt == "application/x-go" {
			err := p.visitGo(state, n)
			if err != nil {
				return err
			}
			return nil
		}

		// component js (type attr omitted okay - means it is JS)
		if mt == "text/javascript" || mt == "application/javascript" || mt == "" {
			err := p.visitJS(state, n)
			if err != nil {
				return err
			}
			return nil
		}

		return fmt.Errorf("found script tag with invalid mime type %q", mt)

	}

	// component css
	if nodeName == "style" || nodeName == "link" {
		err := p.visitCSS(state, n)
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("element %q is not a valid script or style - %#v", n.Data, n)
}

func (p *ParserGo) visitJS(state *parseGoState, n *html.Node) error {

	if n.Type != html.ElementNode {
		return fmt.Errorf("visitJS, not an element node %#v", n)
	}

	nodeName := strings.ToLower(n.Data)

	if nodeName != "script" {
		return fmt.Errorf("visitJS, tag %q not a script", nodeName)
	}

	// see if there's a script inside, or if this is a script include
	if n.FirstChild == nil {
		// script include - we pretty much just let this through, don't care what the attrs are
	} else {
		// if there is a script inside, we do not allow attributes other than "type", to avoid
		// people using features that might not be compatible with the funky stuff we have to do
		// in vugu to make all this work

		for _, a := range n.Attr {
			if a.Key != "type" {
				return fmt.Errorf("attribute %q not allowed on script tag that contains JS code", a.Key)
			}
			if a.Val != "text/javascript" && a.Val != "application/javascript" {
				return fmt.Errorf("script type %q invalid (must be text/javascript)", a.Val)
			}
		}

		// verify that all children are text nodes
		for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
			if childN.Type != html.TextNode {
				return fmt.Errorf("script tag contains non-text child: %#v", childN)
			}
		}

	}

	// allow control stuff, why not

	// vg-for
	if v, _ := vgForExpr(n); v.expr != "" {
		if err := p.emitForExpr(state, n); err != nil {
			return err
		}
		defer fmt.Fprintf(&state.buildBuf, "}\n")
	}

	// vg-if
	ife := vgIfExpr(n)
	if ife != "" {
		fmt.Fprintf(&state.buildBuf, "if %s {\n", ife)
		defer fmt.Fprintf(&state.buildBuf, "}\n")
	}

	// but then for the actual output, we append to vgout.JS, instead of parentNode
	fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}\n", n.Type, n.Data, staticVGAttr(n.Attr))

	// output any text children
	if n.FirstChild != nil {
		fmt.Fprintf(&state.buildBuf, "{\n")
		for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
			// NOTE: we already verified above that these are just text nodes
			fmt.Fprintf(&state.buildBuf, "vgn.AppendChild(&vugu.VGNode{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v})\n", childN.Type, childN.Data, staticVGAttr(childN.Attr))
		}
		fmt.Fprintf(&state.buildBuf, "}\n")
	}

	fmt.Fprintf(&state.buildBuf, "vgout.AppendJS(vgn)\n")

	// dynamic attrs
	writeDynamicAttributes(state, n)

	return nil
}

func (p *ParserGo) visitCSS(state *parseGoState, n *html.Node) error {

	if n.Type != html.ElementNode {
		return fmt.Errorf("visitCSS, not an element node %#v", n)
	}

	nodeName := strings.ToLower(n.Data)
	switch nodeName {

	case "link":

		// okay as long as nothing is inside this node

		if n.FirstChild != nil {
			return fmt.Errorf("link tag should not have children")
		}

		// and it needs to have an href (url)
		hrefAttr := attrWithKey(n, "href")
		if hrefAttr == nil {
			return fmt.Errorf("link tag must have href attribute but does not: %#v", n)
		}

	case "style":

		// style must have child (will verify it is text below)
		if n.FirstChild == nil {
			return fmt.Errorf("style must have contents but does not: %#v", n)
		}

		// okay as long as only text nodes inside
		for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
			if childN.Type != html.TextNode {
				return fmt.Errorf("style tag contains non-text child: %#v", childN)
			}
		}

	default:
		return fmt.Errorf("visitCSS, unexpected tag name %q", nodeName)
	}

	// allow control stuff, why not

	// vg-for
	if v, _ := vgForExpr(n); v.expr != "" {
		if err := p.emitForExpr(state, n); err != nil {
			return err
		}
		defer fmt.Fprintf(&state.buildBuf, "}\n")
	}

	// vg-if
	ife := vgIfExpr(n)
	if ife != "" {
		fmt.Fprintf(&state.buildBuf, "if %s {\n", ife)
		defer fmt.Fprintf(&state.buildBuf, "}\n")
	}

	// but then for the actual output, we append to vgout.CSS, instead of parentNode
	fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}\n", n.Type, n.Data, staticVGAttr(n.Attr))

	// output any text children
	if n.FirstChild != nil {
		fmt.Fprintf(&state.buildBuf, "{\n")
		for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
			// NOTE: we already verified above that these are just text nodes
			fmt.Fprintf(&state.buildBuf, "vgn.AppendChild(&vugu.VGNode{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v})\n", childN.Type, childN.Data, staticVGAttr(childN.Attr))
		}
		fmt.Fprintf(&state.buildBuf, "}\n")
	}

	fmt.Fprintf(&state.buildBuf, "vgout.AppendCSS(vgn)\n")

	// dynamic attrs
	writeDynamicAttributes(state, n)

	return nil
}

func (p *ParserGo) visitGo(state *parseGoState, n *html.Node) error {

	for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
		if childN.Type != html.TextNode {
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
func (p *ParserGo) visitTopNode(state *parseGoState, n *html.Node) error {

	// handle the top element other than <html>

	err := p.visitNodeJustElement(state, n)
	if err != nil {
		return err
	}

	return nil
}

// visitNodeElementAndCtrl handles an element that supports vg-if, vg-for etc
func (p *ParserGo) visitNodeElementAndCtrl(state *parseGoState, n *html.Node) error {

	// vg-for
	if v, _ := vgForExpr(n); v.expr != "" {
		if err := p.emitForExpr(state, n); err != nil {
			return err
		}
		defer fmt.Fprintf(&state.buildBuf, "}\n")
	}

	// vg-if
	ife := vgIfExpr(n)
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
func (p *ParserGo) visitNodeJustElement(state *parseGoState, n *html.Node) error {

	// regular element

	// if n.Line > 0 {
	// 	fmt.Fprintf(&buildBuf, "//line %s:%d\n", fname, n.Line)
	// }

	pOutputTag(state, n)
	// fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{Type:vugu.VGNodeType(%d),Data:%q,Attr:%#v}\n", n.Type, n.Data, staticVGAttr(n.Attr))
	// if state.outIsSet {
	// 	fmt.Fprintf(&state.buildBuf, "vgparent.AppendChild(vgn)\n") // if not root, make AppendChild call
	// } else {
	// 	fmt.Fprintf(&state.buildBuf, "vgout.Out = append(vgout.Out, vgn) // root for output\n") // for first element we need to assign as Doc on BuildOut
	// 	state.outIsSet = true
	// }

	// dynamic attrs
	writeDynamicAttributes(state, n)

	// vg-js-*
	writeJSCallbackAttributes(state, n)

	// js properties
	propExprMap, propExprMapKeys := propVGAttrExpr(n)
	for _, k := range propExprMapKeys {
		valExpr := propExprMap[k]
		fmt.Fprintf(&state.buildBuf, "{b, err := vjson.Marshal(%s); if err != nil { panic(err) }; vgn.Prop = append(vgn.Prop, vugu.VGProperty{Key:%q,JSONVal:vjson.RawMessage(b)})}\n", valExpr, k)
	}

	// vg-html
	htmlExpr := vgHTMLExpr(n)
	if htmlExpr != "" {
		fmt.Fprintf(&state.buildBuf, "vgn.SetInnerHTML(%s)\n", htmlExpr)
	}

	// DOM events
	eventMap, eventKeys := vgDOMEventExprs(n)
	for _, k := range eventKeys {
		expr := eventMap[k]
		fmt.Fprintf(&state.buildBuf, "vgn.DOMEventHandlerSpecList = append(vgn.DOMEventHandlerSpecList, vugu.DOMEventHandlerSpec{\n")
		fmt.Fprintf(&state.buildBuf, "EventType: %q,\n", k)
		fmt.Fprintf(&state.buildBuf, "Func: func(event vugu.DOMEvent) { %s },\n", expr)
		fmt.Fprintf(&state.buildBuf, "// TODO: implement capture, etc. mostly need to decide syntax\n")
		fmt.Fprintf(&state.buildBuf, "})\n")
	}

	if n.FirstChild != nil {

		fmt.Fprintf(&state.buildBuf, "{\n")
		fmt.Fprintf(&state.buildBuf, "vgparent := vgn; _ = vgparent\n") // vgparent set for this block to vgn

		// iterate over children
		for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

			err := p.visitDefaultByType(state, childN)
			if err != nil {
				return err
			}
		}

		fmt.Fprintf(&state.buildBuf, "}\n")

	}

	return nil
}

func (p *ParserGo) visitDefaultByType(state *parseGoState, n *html.Node) error {

	// handle child according to type
	var err error
	switch {
	case n.Type == html.CommentNode:
		err = p.visitNodeComment(state, n)
	case n.Type == html.TextNode:
		err = p.visitNodeText(state, n)
	case n.Type == html.ElementNode:
		if strings.Contains(n.Data, ":") {
			// NOTE: this should check for a capital letter after the colon - this would distinguish
			// svg:svg (valid regular HTML) from svg:Svg (a component reference)
			err = p.visitNodeComponentElement(state, n)
		} else if n.Data == "vg-comp" {
			err = p.visitVGCompTag(state, n)
		} else if n.Data == "vg-template" {
			err = p.visitVGTemplateTag(state, n)
		} else {
			err = p.visitNodeElementAndCtrl(state, n)
		}
	default:
		return fmt.Errorf("child node of unknown type %v: %#v", n.Type, n)
	}

	if err != nil {
		return err
	}

	return nil
}

func (p *ParserGo) visitNodeText(state *parseGoState, n *html.Node) error {

	fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{Type:vugu.VGNodeType(%d),Data:%q}\n", n.Type, n.Data)
	fmt.Fprintf(&state.buildBuf, "vgparent.AppendChild(vgn)\n")

	return nil
}

func (p *ParserGo) visitNodeComment(state *parseGoState, n *html.Node) error {

	fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{Type:vugu.VGNodeType(%d),Data:%q}\n", n.Type, n.Data)
	fmt.Fprintf(&state.buildBuf, "vgparent.AppendChild(vgn)\n")

	return nil
}

// visitVGCompTag handles a vg-comp
func (p *ParserGo) visitVGCompTag(state *parseGoState, n *html.Node) error {

	// vg-for not allowed here

	// vg-if is supported
	ife := vgIfExpr(n)
	if ife != "" {
		fmt.Fprintf(&state.buildBuf, "if %s {\n", ife)
		defer fmt.Fprintf(&state.buildBuf, "}\n")
	}

	// for now, nothing else supported

	// must have a "expr" which gives the Go expression which will result in a component
	expr := vgCompExpr(n)
	if expr == "" {
		return fmt.Errorf("vg-comp must have an `expr` attribute with a Go expression in it")
	}
	fmt.Fprintf(&state.buildBuf, "{\n")
	defer fmt.Fprintf(&state.buildBuf, "}\n")

	fmt.Fprintf(&state.buildBuf, "var vgcomp vugu.Builder = %s\n", expr)
	fmt.Fprintf(&state.buildBuf, "if vgcomp != nil {\n")
	fmt.Fprintf(&state.buildBuf, "    vgin.BuildEnv.WireComponent(vgcomp)\n")
	fmt.Fprintf(&state.buildBuf, "    vgout.Components = append(vgout.Components, vgcomp)\n")
	fmt.Fprintf(&state.buildBuf, "    vgn = &vugu.VGNode{Component:vgcomp}\n")
	fmt.Fprintf(&state.buildBuf, "    vgparent.AppendChild(vgn)\n")
	fmt.Fprintf(&state.buildBuf, "}\n")

	return nil
}

// visitVGTemplateTag handles vg-template
func (p *ParserGo) visitVGTemplateTag(state *parseGoState, n *html.Node) error {

	// vg-for
	if v, _ := vgForExpr(n); v.expr != "" {
		if err := p.emitForExpr(state, n); err != nil {
			return err
		}
		defer fmt.Fprintf(&state.buildBuf, "}\n")
	}

	// vg-if
	ife := vgIfExpr(n)
	if ife != "" {
		fmt.Fprintf(&state.buildBuf, "if %s {\n", ife)
		defer fmt.Fprintf(&state.buildBuf, "}\n")
	}

	// output a node with type Element but empty data
	fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{Type:vugu.VGNodeType(%d)} // <vg-template>\n", vugu.ElementNode)
	fmt.Fprintf(&state.buildBuf, "vgparent.AppendChild(vgn)\n")

	// and then only process children
	if n.FirstChild != nil {

		fmt.Fprintf(&state.buildBuf, "{\n")
		fmt.Fprintf(&state.buildBuf, "vgparent := vgn; _ = vgparent\n") // vgparent set for this block to vgn

		// iterate over children
		for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

			err := p.visitDefaultByType(state, childN)
			if err != nil {
				return err
			}
		}

		fmt.Fprintf(&state.buildBuf, "}\n")

	}

	return nil
}

// visitNodeComponentElement handles an element that is a call to a component
func (p *ParserGo) visitNodeComponentElement(state *parseGoState, n *html.Node) error {

	// components are just different so we handle all of our own vg-for vg-if and everything else

	// vg-for
	if v, _ := vgForExpr(n); v.expr != "" {
		if err := p.emitForExpr(state, n); err != nil {
			return err
		}
		defer fmt.Fprintf(&state.buildBuf, "}\n")
	}

	// vg-if
	ife := vgIfExpr(n)
	if ife != "" {
		fmt.Fprintf(&state.buildBuf, "if %s {\n", ife)
		defer fmt.Fprintf(&state.buildBuf, "}\n")
	}

	nodeName := n.OrigData // use original case of element
	nodeNameParts := strings.Split(nodeName, ":")
	if len(nodeNameParts) != 2 {
		return fmt.Errorf("invalid component tag name %q must contain exactly one colon", nodeName)
	}

	// x.Y or just Y depending on if in same package
	typeExpr := strings.Join(nodeNameParts, ".")
	pkgPrefix := nodeNameParts[0] + "." // needed so we can calc pkg name for pkg.WhateverEvent
	if nodeNameParts[0] == p.PackageName {
		typeExpr = nodeNameParts[1]
		pkgPrefix = ""
	}

	compKeyID := compHashCounted(p.StructType + "." + n.OrigData)

	fmt.Fprintf(&state.buildBuf, "{\n")
	defer fmt.Fprintf(&state.buildBuf, "}\n")

	keyExpr := vgKeyExpr(n)
	if keyExpr != "" {
		fmt.Fprintf(&state.buildBuf, "vgcompKey := vugu.MakeCompKey(0x%X^vgin.CurrentPositionHash(), %s)\n", compKeyID, keyExpr)
	} else {
		fmt.Fprintf(&state.buildBuf, "vgcompKey := vugu.MakeCompKey(0x%X^vgin.CurrentPositionHash(), vgiterkey)\n", compKeyID)
	}
	fmt.Fprintf(&state.buildBuf, "// ask BuildEnv for prior instance of this specific component\n")
	fmt.Fprintf(&state.buildBuf, "vgcomp, _ := vgin.BuildEnv.CachedComponent(vgcompKey).(*%s)\n", typeExpr)
	fmt.Fprintf(&state.buildBuf, "if vgcomp == nil {\n")
	fmt.Fprintf(&state.buildBuf, "// create new one if needed\n")
	fmt.Fprintf(&state.buildBuf, "vgcomp = new(%s)\n", typeExpr)
	fmt.Fprintf(&state.buildBuf, "vgin.BuildEnv.WireComponent(vgcomp)\n")
	fmt.Fprintf(&state.buildBuf, "}\n")
	fmt.Fprintf(&state.buildBuf, "vgin.BuildEnv.UseComponent(vgcompKey, vgcomp) // ensure we can use this in the cache next time around\n")

	// now that we have vgcomp with the right type and a correct value, we can declare the vg-var if specified
	if vgv := vgVarExpr(n); vgv != "" {
		fmt.Fprintf(&state.buildBuf, "var %s = vgcomp // vg-var\n", vgv)

		// NOTE: It's a bit too much to have "unused variable" errors coming from a Vugu code-generated file,
		// too far off the beaten path of making "type-safe HTML templates with Go".  It makes sense with
		// hand-written Go code but I don't think so here.
		fmt.Fprintf(&state.buildBuf, "_ = %s\n", vgv) // avoid unused var error
	}

	didAttrMap := false

	// dynamic attrs
	dynExprMap, dynExprMapKeys := dynamicVGAttrExpr(n)
	for _, k := range dynExprMapKeys {
		// if k == "" {
		// 	return fmt.Errorf("invalid empty dynamic attribute name on component %#v", n)
		// }

		valExpr := dynExprMap[k]

		// if starts with upper case, it's a field name
		if hasUpperFirst(k) {
			fmt.Fprintf(&state.buildBuf, "vgcomp.%s = %s\n", k, valExpr)
		} else {
			// otherwise we use an "AttrMap"
			if !didAttrMap {
				didAttrMap = true
				fmt.Fprintf(&state.buildBuf, "vgcomp.AttrMap = make(map[string]interface{}, 8)\n")
			}
			fmt.Fprintf(&state.buildBuf, "vgcomp.AttrMap[%q] = %s\n", k, valExpr)
		}

	}

	// static attrs
	vgAttrs := staticVGAttr(n.Attr)
	for _, a := range vgAttrs {
		// if starts with upper case, it's a field name
		if hasUpperFirst(a.Key) {
			fmt.Fprintf(&state.buildBuf, "vgcomp.%s = %q\n", a.Key, a.Val)
		} else {
			// otherwise we use an "AttrMap"
			if !didAttrMap {
				didAttrMap = true
				fmt.Fprintf(&state.buildBuf, "vgcomp.AttrMap = make(map[string]interface{}, 8)\n")
			}
			fmt.Fprintf(&state.buildBuf, "vgcomp.AttrMap[%q] = %q\n", a.Key, a.Val)
		}
	}

	// component events
	// NOTE: We keep component events really simple and the @ is just a thin wrapper around a field assignment:
	//     <pkg:Comp @Something="log.Println(event)"></pkg:Comp>
	// is shorthand for:
	//     <pkg:Comp :Something='func(event pkg.SomethingEvent) { log.Println(event) }'></pkg:Comp>
	//
	// I considered using the handler interface function approach for this, but it would mean
	// SomethingHandlerFunc would have to exist as a type, with a SomethingHandle method, which
	// implements a SomethingHandler interface, so the type of Comp.Something could be SomethingHandler,
	// and the emitted code could be vgcomp.Something = pkg.SomethingHandlerFunc(func...)
	// But that's two additional types and a method for every event.  I'm very concerned that it will
	// make component events feel crufty and arduous to implement (unless we could find a good way
	// to automatically generate those when missing - that's a possibility - actually I think
	// I'm going to try this, see https://github.com/vugu/vugu/issues/128).
	// But this this way with a func you can just do
	// type SomethingEvent struct { /* whatever relevant data */ } and then define your field on
	// your component as Something func(SomethingEvent) - still type-safe but very straightforward.
	// So far it seems like the best approach.

	eventMap, eventKeys := vgEventExprs(n)
	for _, k := range eventKeys {
		expr := eventMap[k]
		// fmt.Fprintf(&state.buildBuf, "vgcomp.%s = func(event %s%sEvent){%s}\n", k, pkgPrefix, k, expr)
		// switched to using interfaces
		fmt.Fprintf(&state.buildBuf, "vgcomp.%s = %s%sFunc(func(event %s%sEvent){%s})\n", k, pkgPrefix, k, pkgPrefix, k, expr)
	}

	// NOTE: vugugen:slot might come in really handy, have to work out the types involved - update: as it stands, this won't be needed.

	// slots:

	// scan children and see if it's default slot mode or vg-slot tags
	foundTagSlot, foundDefSlot := false, false
	var foundTagSlotNames []string
	for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

		// non-ws text means default slot
		if childN.Type == html.TextNode {
			if strings.TrimSpace(childN.Data) != "" {
				foundDefSlot = true
			}
			continue
		}

		// ignore comments
		if childN.Type == html.CommentNode {
			continue
		}

		// should only be element at this point
		if childN.Type != html.ElementNode {
			return fmt.Errorf("in tag %q unexpected node found where slot expected: %#v", n.Data, childN)
		}

		if childN.Data == "vg-slot" {
			foundTagSlot = true
			name := strings.TrimSpace(vgSlotName(childN))
			if name != "" {
				foundTagSlotNames = append(foundTagSlotNames, name)
			}
		} else {
			foundDefSlot = true
		}
	}

	// now process slot(s) appropriately according to format
	switch {

	case foundTagSlot && foundDefSlot:
		return fmt.Errorf("in tag %q found both vg-slot and other markup, only one or the other is allowed", n.Data)

	case foundTagSlot:

		// NOTE:
		// <vg-slot name="X"> will assign to vgcomp.X
		// <vg-slot name='X[Y]'> will assume X is of type map[string]Builder and create the map and then assign with X[Y] =

		// find any names with map expressions and clear the maps
		sort.Strings(foundTagSlotNames)
		slotMapInited := make(map[string]bool)
		for _, slotName := range foundTagSlotNames {
			slotNameParts := strings.Split(slotName, "[") // check for map expr
			if len(slotNameParts) > 1 {                   // if map
				if slotMapInited[slotNameParts[0]] { // if not already initialized
					continue
				}
				slotMapInited[slotNameParts[0]] = true

				// if nil create map, otherwise reuse
				fmt.Fprintf(&state.buildBuf, "if vgcomp.%s == nil {\n", slotNameParts[0])
				fmt.Fprintf(&state.buildBuf, "    vgcomp.%s = make(map[string]vugu.Builder)\n", slotNameParts[0])
				fmt.Fprintf(&state.buildBuf, "} else {\n")
				fmt.Fprintf(&state.buildBuf, "    for k := range vgcomp.%s { delete(vgcomp.%s, k) }\n", slotNameParts[0], slotNameParts[0])
				fmt.Fprintf(&state.buildBuf, "}\n")
			}
		}

		// iterate over children
		for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

			// ignore white space and coments
			if childN.Type == html.CommentNode ||
				(childN.Type == html.TextNode && strings.TrimSpace(childN.Data) == "") {
				continue
			}

			if childN.Type != html.ElementNode { // should be impossible from foundTagSlot check above, just making sure
				panic(fmt.Errorf("unexpected non-element found where vg-slot should be: %#v", childN))
			}

			if childN.Data != "vg-slot" { // should also be imposible
				panic(fmt.Errorf("unexpected element found where vg-slot should be: %#v", childN))
			}

			slotName := strings.TrimSpace(vgSlotName(childN))
			if slotName == "" {
				return fmt.Errorf("found vg-slot tag without a 'name' attribute, the name is required")
			}

			fmt.Fprintf(&state.buildBuf, "vgcomp.%s = vugu.NewBuilderFunc(func(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {\n", slotName)
			fmt.Fprintf(&state.buildBuf, "vgn := &vugu.VGNode{Type:vugu.VGNodeType(%d)}\n", vugu.ElementNode)
			fmt.Fprintf(&state.buildBuf, "vgout = &vugu.BuildOut{}\n")
			fmt.Fprintf(&state.buildBuf, "vgout.Out = append(vgout.Out, vgn)\n")
			fmt.Fprintf(&state.buildBuf, "vgparent := vgn; _ = vgparent\n")
			fmt.Fprintf(&state.buildBuf, "\n")

			// iterate over children and do the usual with each one
			for innerChildN := childN.FirstChild; innerChildN != nil; innerChildN = innerChildN.NextSibling {
				err := p.visitDefaultByType(state, innerChildN)
				if err != nil {
					return err
				}
			}

			fmt.Fprintf(&state.buildBuf, "return\n")
			fmt.Fprintf(&state.buildBuf, "})\n")

		}

	case foundDefSlot:
		fmt.Fprintf(&state.buildBuf, "vgcomp.DefaultSlot = vugu.NewBuilderFunc(func(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {\n")
		// vgn is the equivalent of a vg-template tag and becomes the contents of vgout.Out and the vgparent
		fmt.Fprintf(&state.buildBuf, "vgn := &vugu.VGNode{Type:vugu.VGNodeType(%d)}\n", vugu.ElementNode)
		fmt.Fprintf(&state.buildBuf, "vgout = &vugu.BuildOut{}\n")
		fmt.Fprintf(&state.buildBuf, "vgout.Out = append(vgout.Out, vgn)\n")
		fmt.Fprintf(&state.buildBuf, "vgparent := vgn; _ = vgparent\n")
		fmt.Fprintf(&state.buildBuf, "\n")

		// iterate over children and do the usual with each one
		for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
			err := p.visitDefaultByType(state, childN)
			if err != nil {
				return err
			}
		}

		fmt.Fprintf(&state.buildBuf, "return\n")
		fmt.Fprintf(&state.buildBuf, "})\n")

	default:
		// nothing meaningful inside this component tag
	}

	// // keep track of contents for default slot
	// var defSlotNodes []*html.Node
	// defSlotMode := false // start off not in default slot mode and look for <vg-slot> tags

	// // loop over all component children
	// for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

	// 	if !defSlotMode {

	// 		// anything not an element just add to the list for default
	// 		if childN.Type != html.ElementNode {
	// 			defSlotNodes = append(defSlotNodes, childN)
	// 			continue
	// 		}

	// 		if childN.Data == "vg-slot" {

	// 		}

	// 	}

	// }

	// ignore whitespace
	// first non-slot, non-ws child, assume "DefaultSlot" (or whatever name) and consume rest of children
	// if vg-slot, then consume with specified name
	// <vg-slot name="SomeSlot"> <!-- field name syntax
	// <vg-slot name='SomeDynaSlot' index='"Row.FirstName"'> <!-- expression syntax, HM, NO
	// <vg-slot index='SomeDynaSlot["Row.FirstName"]'> <!-- maybe this - still annoying that we have to limit it to a map expression, but whatever
	// emit vgcomp.SlotName = vugu.NewBuilderFunc(func(vgin *vugu.BuildIn) (vgout *BuildOut, vgerr error) { ... })
	// and descend into children

	fmt.Fprintf(&state.buildBuf, "vgout.Components = append(vgout.Components, vgcomp)\n")
	fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{Component:vgcomp}\n")
	fmt.Fprintf(&state.buildBuf, "vgparent.AppendChild(vgn)\n")

	return nil
	// return fmt.Errorf("component tag not yet supported (%q)", nodeName)
}

// NOTE: caller is responsible for emitting the closing curly bracket
func (p *ParserGo) emitForExpr(state *parseGoState, n *html.Node) error {

	forattr, err := vgForExpr(n)
	if err != nil {
		return err
	}
	forx := forattr.expr
	if forx == "" {
		return errors.New("no for expression, code should not be calling emitForExpr when no vg-for is present")
	}

	// cases to select vgiterkey:
	// * check for vg-key attribute
	// * _, v := // replace _ with vgiterkey
	// * key, value := // unused vars, use 'key' as iter val
	// * k, v := // detect `k` and use as iterval

	vgiterkeyx := vgKeyExpr(n)

	// determine iteration variables
	var iterkey, iterval string
	if !strings.Contains(forx, ":=") {
		// make it so `w` is a shorthand for `key, value := range w`
		iterkey, iterval = "key", "value"
		forx = "key, value := range " + forx
	} else {
		// extract iteration variables
		var (
			itervars [2]string
			iteridx  int
		)
		for _, c := range forx {
			if c == ':' {
				break
			}
			if c == ',' {
				iteridx++
				continue
			}
			if unicode.IsSpace(c) {
				continue
			}
			itervars[iteridx] += string(c)
		}

		iterkey = itervars[0]
		iterval = itervars[1]
	}

	// detect "_, k := " form combined with no vg-key specified and replace
	if vgiterkeyx == "" && iterkey == "_" {
		iterkey = "vgiterkeyt"
		forx = "vgiterkeyt " + forx[1:]
	}

	// if still no vgiterkeyx use the first identifier
	if vgiterkeyx == "" {
		vgiterkeyx = iterkey
	}

	fmt.Fprintf(&state.buildBuf, "for %s {\n", forx)
	fmt.Fprintf(&state.buildBuf, "var vgiterkey interface{} = %s\n", vgiterkeyx)
	fmt.Fprintf(&state.buildBuf, "_ = vgiterkey\n")
	if !forattr.noshadow {
		if iterkey != "_" && iterkey != "vgiterkeyt" {
			fmt.Fprintf(&state.buildBuf, "%[1]s := %[1]s\n", iterkey)
			fmt.Fprintf(&state.buildBuf, "_ = %s\n", iterkey)
		}
		if iterval != "_" && iterval != "" {
			fmt.Fprintf(&state.buildBuf, "%[1]s := %[1]s\n", iterval)
			fmt.Fprintf(&state.buildBuf, "_ = %s\n", iterval)
		}
	}

	return nil
}

func hasUpperFirst(s string) bool {
	for _, c := range s {
		return unicode.IsUpper(c)
	}
	return false
}

// isScriptOrStyle returns true if this is a "script", "style" or "link" tag
func isScriptOrStyle(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	switch strings.ToLower(n.Data) {
	case "script", "style", "link":
		return true
	}
	return false
}

func pOutputTag(state *parseGoState, n *html.Node) {
	fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{Type:vugu.VGNodeType(%d),Namespace:%q,Data:%q,Attr:%#v}\n", n.Type, n.Namespace, n.Data, staticVGAttr(n.Attr))
	if state.outIsSet {
		fmt.Fprintf(&state.buildBuf, "vgparent.AppendChild(vgn)\n") // if not root, make AppendChild call
	} else {
		fmt.Fprintf(&state.buildBuf, "vgout.Out = append(vgout.Out, vgn) // root for output\n") // for first element we need to assign as Doc on BuildOut
		state.outIsSet = true
	}

}

func attrWithKey(n *html.Node, key string) *html.Attribute {
	for i := range n.Attr {
		if n.Attr[i].Key == key {
			return &n.Attr[i]
		}
	}
	return nil
}

func writeDynamicAttributes(state *parseGoState, n *html.Node) {
	dynExprMap, dynExprMapKeys := dynamicVGAttrExpr(n)
	for _, k := range dynExprMapKeys {
		valExpr := dynExprMap[k]
		if k == "" || k == "vg-attr" {
			fmt.Fprintf(&state.buildBuf, "vgn.AddAttrList(%s)\n", valExpr)
		} else {
			fmt.Fprintf(&state.buildBuf, "vgn.AddAttrInterface(%q,%s)\n", k, valExpr)
		}
	}
}

// writeJSCallbackAttributes handles vg-js-create and vg-js-populate
func writeJSCallbackAttributes(state *parseGoState, n *html.Node) {
	m := jsCallbackVGAttrExpr(n)
	createStmt := m["vg-js-create"]
	if createStmt != "" {
		fmt.Fprintf(&state.buildBuf, "vgn.JSCreateHandler = vugu.JSValueFunc(func(value js.Value) { %s })\n", createStmt)
	}
	populateStmt := m["vg-js-populate"]
	if populateStmt != "" {
		fmt.Fprintf(&state.buildBuf, "vgn.JSPopulateHandler = vugu.JSValueFunc(func(value js.Value) { %s })\n", populateStmt)
	}
}
