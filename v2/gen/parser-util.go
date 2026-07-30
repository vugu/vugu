package gen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/vugu/html"

	"github.com/vugu/vugu/v2"
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
	expr string
}

func vgForExpr(n *html.Node) (vgForAttr, error) {
	for _, a := range n.Attr {
		if strings.HasPrefix(a.Key, "vg-for") {
			v := vgForAttr{expr: strings.TrimSpace(a.Val)}
			opts := strings.Split(a.Key, ".")
			if len(opts) > 1 {
				for _, opt := range opts[1:] {
					switch opt {
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
