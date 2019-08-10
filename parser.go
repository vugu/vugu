package vugu

import (
	"regexp"
	"sort"
	"strings"

	"github.com/vugu/vugu/internal/htmlx"
	"golang.org/x/net/html"
)

func attrFromHtml(attr html.Attribute) VGAttribute {
	return VGAttribute{
		Namespace: attr.Namespace,
		Key:       attr.Key,
		Val:       attr.Val,
	}
}

func attrFromHtmlx(attr htmlx.Attribute) VGAttribute {
	return VGAttribute{
		Namespace: attr.Namespace,
		Key:       attr.Key,
		Val:       attr.Val,
	}
}

// stuff that is common to both parsers can get moved into here

func staticVGAttr(inAttr []html.Attribute) (ret []VGAttribute) {

	for _, a := range inAttr {
		switch {
		case a.Key == "vg-if":
		case a.Key == "vg-for":
		case a.Key == "vg-html":
		case strings.HasPrefix(a.Key, ":"):
		case strings.HasPrefix(a.Key, "@"):
		default:
			ret = append(ret, attrFromHtml(a))
		}
	}

	return ret
}

func staticVGAttrx(inAttr []htmlx.Attribute) (ret []VGAttribute) {

	for _, a := range inAttr {
		switch {
		case a.Key == "vg-if":
		case a.Key == "vg-for":
		case a.Key == "vg-html":
		case strings.HasPrefix(a.Key, ":"):
		case strings.HasPrefix(a.Key, "@"):
		default:
			ret = append(ret, attrFromHtmlx(a))
		}
	}

	return ret
}

func vgIfExpr(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "vg-if" {
			return a.Val
		}
	}
	return ""
}

func vgIfExprx(n *htmlx.Node) string {
	for _, a := range n.Attr {
		if a.Key == "vg-if" {
			return a.Val
		}
	}
	return ""
}

func vgForExpr(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "vg-for" {

			v := strings.TrimSpace(a.Val)

			if !strings.Contains(v, " ") { // make it so `w` is a shorthand for `key, value := range w`
				v = "key, value := range " + v
			}

			return v
		}
	}
	return ""
}

func vgForExprx(n *htmlx.Node) string {
	for _, a := range n.Attr {
		if a.Key == "vg-for" {

			v := strings.TrimSpace(a.Val)

			if !strings.Contains(v, " ") { // make it so `w` is a shorthand for `key, value := range w`
				v = "key, value := range " + v
			}

			return v
		}
	}
	return ""
}

func vgHTMLExpr(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "vg-html" {
			return a.Val
		}
	}
	return ""
}

func vgHTMLExprx(n *htmlx.Node) string {
	for _, a := range n.Attr {
		if a.Key == "vg-html" {
			return a.Val
		}
	}
	return ""
}

// extract ":prop" stuff from a node
func dynamicVGAttrExpr(n *html.Node) (ret map[string]string) {
	var da []html.Attribute
	// get dynamic attrs first
	for _, a := range n.Attr {
		if strings.HasPrefix(a.Key, ":") {
			da = append(da, a)
		}
	}
	if len(da) == 0 { // don't allocate map if we don't have to
		return
	}
	// make map as small as possible
	ret = make(map[string]string, len(da))
	for _, a := range da {
		ret[strings.TrimPrefix(a.Key, ":")] = a.Val
	}
	return
}

// extract ":prop" stuff from a node
func dynamicVGAttrExprx(n *htmlx.Node) (ret map[string]string, retKeys []string) {
	var da []htmlx.Attribute
	// get dynamic attrs first
	for _, a := range n.Attr {
		if strings.HasPrefix(a.Key, ":") {
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
		k := strings.TrimPrefix(a.Key, ":")
		retKeys[i] = k
		ret[k] = a.Val
	}
	sort.Strings(retKeys)
	return
}

// extract "@event" stuff from a node
func vgDOMEventExprs(n *html.Node) (ret map[string]string) {
	var da []html.Attribute
	// get attrs first
	for _, a := range n.Attr {
		if strings.HasPrefix(a.Key, "@") {
			da = append(da, a)
		}
	}
	if len(da) == 0 { // don't allocate map if we don't have to
		return
	}
	// make map as small as possible
	ret = make(map[string]string, len(da))
	for _, a := range da {
		ret[strings.TrimPrefix(a.Key, "@")] = a.Val
	}
	return
}

var vgDOMParseExprRE = regexp.MustCompile(`^([a-zA-Z0-9_.]+)\((.*)\)$`)

func vgDOMParseExpr(expr string) (receiver string, methodName string, argList string) {
	parts := vgDOMParseExprRE.FindStringSubmatch(expr)
	if len(parts) != 3 {
		return
	}
	argList = parts[2]
	f := parts[1]
	fparts := strings.Split(f, ".")

	receiver, methodName = strings.Join(fparts[:len(fparts)-1], "."), fparts[len(fparts)-1]

	// if len(fparts) == 1 { // just "methodName"
	// 	methodName = f
	// } else if len(fparts) > 2 { // "a.b.MethodName"
	// 	receiver, methodName = strings.Join(fparts[:len(fparts)-1], "."), fparts[len(fparts)-1]
	// } else { // "a.MethodName"
	// 	receiver, methodName = fparts[0], fparts[1]
	// }
	return
}

// ^([a-zA-Z0-9_.]+)\((.*)\)$
