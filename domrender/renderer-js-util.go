package domrender

// namespaceToURI resolves the given namespaces to the URI with the specifications
func namespaceToURI(namespace string) string {
	switch namespace {
	case "html":
		return "http://www.w3.org/1999/xhtml"
	case "math":
		return "http://www.w3.org/1998/Math/MathML"
	case "svg":
		return "http://www.w3.org/2000/svg"
	case "xlink":
		return "http://www.w3.org/1999/xlink"
	case "xml":
		return "http://www.w3.org/XML/1998/namespace"
	case "xmlns":
		return "http://www.w3.org/2000/xmlns/"
	default:
		return ""
	}
}
