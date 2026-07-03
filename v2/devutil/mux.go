package devutil

import (
	"net/http"
	"path"
)

/*

Example use:
wc := devutil.NewWasmCompiler().SetDir(".")
mux := devutil.NewMux()
//mux.Exact("/", devutil.DefaultIndex)
mux.Match(devutil.NoFileExt, devutil.DefaultIndex)
//mux.Match(devutil.NoFileExt, devutil.StaticFilePath("index.html"))
mux.Exact("/main.wasm", devutil.MainWasmHandler(wc))
mux.Exact("/wasm_exec.js", devutil.WasmExecJSHandler(wc))
mux.Default(devutil.NewFileServer().SetDir("."))

*/

// RequestMatcher describes something that can say yes/no if a request matches.
// We use it here to mean "should this path be followed in order to answer this request".
type RequestMatcher interface {
	RequestMatch(r *http.Request) bool
}

// RequestMatcherFunc implements RequestMatcher as a function.
type RequestMatcherFunc func(r *http.Request) bool

// NoFileExt is a RequestMatcher that will return true for all paths which do not have a file extension.
var NoFileExt = RequestMatcherFunc(func(r *http.Request) bool {
	return path.Ext(path.Clean("/"+r.URL.Path)) == ""
})

// RequestMatch implements RequestMatcher.
func (f RequestMatcherFunc) RequestMatch(r *http.Request) bool { return f(r) }

// Mux is simple HTTP request multiplexer that has more generally useful behavior for Vugu development than http.ServeMux.
// Routes are considered in the order they are added.
type Mux struct {
	routeList      []muxRoute
	defaultHandler http.Handler
}

type muxRoute struct {
	rm RequestMatcher
	h  http.Handler
}

// NewMux returns a new Mux.
func NewMux() *Mux {
	return &Mux{}
}

// Exact adds an exact route match.  If path.Clean("/"+r.URL.Path)==name then h is called to handle the request.
func (m *Mux) Exact(name string, h http.Handler) *Mux {
	m.Match(RequestMatcherFunc(func(r *http.Request) bool {
		return path.Clean("/"+r.URL.Path) == name
	}), h)
	return m
}

// Match adds a route that is used if the provided RequestMatcher returns true.
func (m *Mux) Match(rm RequestMatcher, h http.Handler) *Mux {
	m.routeList = append(m.routeList, muxRoute{rm: rm, h: h})
	return m
}

// Default sets the defualt handler to be called if no other matches were found.
func (m *Mux) Default(h http.Handler) *Mux {
	m.defaultHandler = h
	return m
}

// ServeHTTP implements http.Handler.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	for _, rt := range m.routeList {
		if rt.rm.RequestMatch(r) {
			rt.h.ServeHTTP(w, r)
			return
		}
	}

	if m.defaultHandler != nil {
		m.defaultHandler.ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}
