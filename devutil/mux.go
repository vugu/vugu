package devutil

import (
	"fmt"
	"net/http"
)

/*
mux := http.NewServeMux()
mux.Handle("/main.wasm", comp)
mux.Handle("/wasm_exec.js", comp)
mux.Handle("/", fs)
- blarg, except it will serve the index page for /whatever/ - this is dumb behavior, we should not encourage it
  hm, it doesn't matter for most cases, since "/" will be the same handler (FileServer)
  actually it does if we want to make an index page handler that doesn't answer for other URLs...
*/
// we still are going to want a one-ish-liner for the above I think

// RequestMatcher describes something that can say yes/no if a request matches.
// We use it here to mean "should this path be followed in order to answer this request".
type RequestMatcher interface {
	RequestMatch(r *http.Request) bool
}

// RequestMatcherFunc implements RequestMatcher as a function.
type RequestMatcherFunc func(r *http.Request) bool

// RequestMatch implements RequestMatcher.
func (f RequestMatcherFunc) RequestMatch(r *http.Request) bool { return f(r) }

type Mux struct {
}

func NewMux() *Mux {
	panic(fmt.Errorf("not yet implemented"))
}

func (m *Mux) Exact(name string, h http.Handler) *Mux {
	panic(fmt.Errorf("not yet implemented"))
}

func (m *Mux) Func(fn func(*http.Request) bool, h http.Handler) *Mux {
	panic(fmt.Errorf("not yet implemented"))
}

func (m *Mux) Default(h http.Handler) *Mux {
	panic(fmt.Errorf("not yet implemented"))
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
