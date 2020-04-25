package gen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"sort"
	"strings"

	// "github.com/vugu/vugu/internal/htmlx"

	// "golang.org/x/net/html"
	"github.com/vugu/html"

	"github.com/vugu/vugu"
)

func attrFromHtml(attr html.Attribute) vugu.VGAttribute {
	return vugu.VGAttribute{
		Namespace: attr.Namespace,
		Key:       attr.OrigKey,
		Val:       attr.Val,
	}
}

// func attrFromHtmlx(attr htmlx.Attribute) vugu.VGAttribute {
// 	return vugu.VGAttribute{
// 		Namespace: attr.Namespace,
// 		Key:       attr.Key,
// 		Val:       attr.Val,
// 	}
// }

// stuff that is common to both parsers can get moved into here

func staticVGAttr(inAttr []html.Attribute) (ret []vugu.VGAttribute) {

	for _, a := range inAttr {
		switch {
		// case a.Key == "vg-if":
		// case a.Key == "vg-for":
		// case a.Key == "vg-key":
		// case a.Key == "vg-html":
		case strings.HasPrefix(a.Key, "vg-"):
		case strings.HasPrefix(a.Key, "."):
		case strings.HasPrefix(a.Key, ":"):
		case strings.HasPrefix(a.Key, "@"):
		default:
			ret = append(ret, attrFromHtml(a))
		}
	}

	return ret
}

func vgSlotName(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "name" {
			return a.Val
		}
	}
	return ""
}

func vgVarExpr(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "vg-var" {
			return a.Val
		}
	}
	return ""
}

func vgIfExpr(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "vg-if" {
			return a.Val
		}
	}
	return ""
}

func vgKeyExpr(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "vg-key" {
			return a.Val
		}
	}
	return ""
}

func vgCompExpr(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "expr" {
			return a.Val
		}
	}
	return ""
}

// func vgIfExprx(n *htmlx.Node) string {
// 	for _, a := range n.Attr {
// 		if a.Key == "vg-if" {
// 			return a.Val
// 		}
// 	}
// 	return ""
// }

type vgForAttr struct {
	expr     string
	noshadow bool
}

func vgForExpr(n *html.Node) (vgForAttr, error) {
	for _, a := range n.Attr {
		if strings.HasPrefix(a.Key, "vg-for") {
			v := vgForAttr{expr: strings.TrimSpace(a.Val)}
			opts := strings.Split(a.Key, ".")
			if len(opts) > 1 {
				for _, opt := range opts[1:] {
					switch opt {
					case "noshadow":
						v.noshadow = true
					default:
						return vgForAttr{}, fmt.Errorf("option %q unknown", opt)
					}
				}
			}
			return v, nil
		}
	}
	return vgForAttr{}, nil
}

func vgHTMLExpr(n *html.Node) string {
	for _, a := range n.Attr {
		// vg-html and vg-content are the same thing,
		// the name vg-content was introduced to call out
		// the difference between Vue's v-html attribute
		// which does not perform escaping.
		if a.Key == "vg-html" {
			return a.Val
		}
		if a.Key == "vg-content" {
			return a.Val
		}
	}
	return ""
}

// extract ":attr" stuff from a node
func dynamicVGAttrExpr(n *html.Node) (ret map[string]string, retKeys []string) {
	var da []html.Attribute
	// get dynamic attrs first
	for _, a := range n.Attr {
		// ":" and "vg-attr" are the AttributeLister case
		if strings.HasPrefix(a.OrigKey, ":") || a.OrigKey == "vg-attr" {
			da = append(da, a)
		}
	}
	if len(da) == 0 { // don't allocate map if we don't have to
		return
	}
	// make map as small as possible
	ret = make(map[string]string, len(da))
	retKeys = make([]string, len(da))
	for i, a := range da {
		k := strings.TrimPrefix(a.OrigKey, ":")
		retKeys[i] = k
		ret[k] = a.Val
	}
	sort.Strings(retKeys)
	return
}

// extract ".prop" stuff from a node
func propVGAttrExpr(n *html.Node) (ret map[string]string, retKeys []string) {
	var da []html.Attribute
	// get prop attrs first
	for _, a := range n.Attr {
		if strings.HasPrefix(a.OrigKey, ".") {
			da = append(da, a)
		}
	}
	if len(da) == 0 { // don't allocate map if we don't have to
		return
	}
	// make map as small as possible
	ret = make(map[string]string, len(da))
	retKeys = make([]string, len(da))
	for i, a := range da {
		k := strings.TrimPrefix(a.OrigKey, ".")
		retKeys[i] = k
		ret[k] = a.Val
	}
	sort.Strings(retKeys)
	return
}

// returns vg-js-create and vg-js-populate
func jsCallbackVGAttrExpr(n *html.Node) (ret map[string]string) {
	for _, attr := range n.Attr {
		if strings.HasPrefix(attr.OrigKey, "vg-js-") {
			if ret == nil {
				ret = make(map[string]string, 2)
			}
			ret[attr.OrigKey] = attr.Val
		}
	}
	return ret
}

func vgDOMEventExprs(n *html.Node) (ret map[string]string, retKeys []string) {
	return vgEventExprs(n)
}

// extract "@event" stuff from a node
func vgEventExprs(n *html.Node) (ret map[string]string, retKeys []string) {
	var da []html.Attribute
	// get attrs first
	for _, a := range n.Attr {
		if strings.HasPrefix(a.OrigKey, "@") {
			da = append(da, a)
		}
	}
	if len(da) == 0 { // don't allocate map if we don't have to
		return
	}
	// make map as small as possible
	ret = make(map[string]string, len(da))
	for _, a := range da {
		k := strings.TrimPrefix(a.OrigKey, "@")
		retKeys = append(retKeys, k)
		ret[k] = a.Val
	}
	return
}

// var vgDOMParseExprRE = regexp.MustCompile(`^([a-zA-Z0-9_.]+)\((.*)\)$`)

// func vgDOMParseExpr(expr string) (receiver string, methodName string, argList string) {
// 	parts := vgDOMParseExprRE.FindStringSubmatch(expr)
// 	if len(parts) != 3 {
// 		return
// 	}
// 	argList = parts[2]
// 	f := parts[1]
// 	fparts := strings.Split(f, ".")

// 	receiver, methodName = strings.Join(fparts[:len(fparts)-1], "."), fparts[len(fparts)-1]

// 	// if len(fparts) == 1 { // just "methodName"
// 	// 	methodName = f
// 	// } else if len(fparts) > 2 { // "a.b.MethodName"
// 	// 	receiver, methodName = strings.Join(fparts[:len(fparts)-1], "."), fparts[len(fparts)-1]
// 	// } else { // "a.MethodName"
// 	// 	receiver, methodName = fparts[0], fparts[1]
// 	// }
// 	return
// }

// ^([a-zA-Z0-9_.]+)\((.*)\)$

// dedupImports reads Go source and removes duplicate import statements.
func dedupImports(r io.Reader, w io.Writer, fname string) error {

	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, fname, r, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return err
	}

	// ast.Print(fset, f)
	// ast.Print(fset, f)

	// ast.Print(fset, f.Decls)
	// f.Decls = f.Decls[1:]
	dedupAstFileImports(f)
	ast.SortImports(fset, f)

	err = printer.Fprint(w, fset, f)
	if err != nil {
		return err
	}

	return nil
}

func dedupAstFileImports(f *ast.File) {

	imap := make(map[string]bool, len(f.Imports)+10)

	outdecls := make([]ast.Decl, 0, len(f.Decls))
	for _, decl := range f.Decls {

		// check for import declaration
		genDecl, _ := decl.(*ast.GenDecl)
		// not an import declaration, just copy and continue
		if genDecl == nil || genDecl.Tok != token.IMPORT {
			outdecls = append(outdecls, decl)
			continue
		}

		// for imports, we loop over each ImportSpec (each package, regardless of which form of import statement)
		outspecs := make([]ast.Spec, 0, len(genDecl.Specs))
		for _, spec := range genDecl.Specs {
			ispec := spec.(*ast.ImportSpec)
			// always use path
			key := ispec.Path.Value
			// if name is present, prepend
			if ispec.Name != nil {
				key = ispec.Name.Name + " " + key
			}

			// if we've seen this import before, then just move to the next
			if imap[key] {
				continue
			}
			imap[key] = true // mark this import as having been seen

			// keep the import
			outspecs = append(outspecs, ispec)
		}

		// use outspecs for this import decl, unless it's empty in which case we remove/skip the whole import decl
		if len(outspecs) == 0 {
			continue
		}
		genDecl.Specs = outspecs
		outdecls = append(outdecls, genDecl)

	}
	f.Decls = outdecls

}
