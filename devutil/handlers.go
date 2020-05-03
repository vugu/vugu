package devutil

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Compiler is implemented by WasmCompiler and TinygoCompiler.
type Compiler interface {
	Execute() (outpath string, err error)
}

// WasmExecJSer is implemented by WasmCompiler and TinygoCompiler.
type WasmExecJSer interface {
	WasmExecJS() (contents io.Reader, err error)
}

// MainWasmHandler calls WasmCompiler.Build and responds with the resulting .wasm file.
type MainWasmHandler struct {
	wc Compiler
}

// NewMainWasmHandler returns an initialized MainWasmHandler.
func NewMainWasmHandler(wc Compiler) *MainWasmHandler {
	return &MainWasmHandler{
		wc: wc,
	}
}

// ServeHTTP implements http.Handler.
func (h *MainWasmHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	outpath, err := h.wc.Execute()
	if err != nil {
		log.Printf("MainWasmHandler: Execute error:\n%v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, "MainWasmHandler: Execute error:\n"+err.Error(), 500)
		return
	}
	defer os.Remove(outpath)

	w.Header().Set("Content-Type", "application/wasm")

	f, err := os.Open(outpath)
	if err != nil {
		log.Printf("MainWasmHandler: File open error:\n%v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, "MainWasmHandler: File open error:\n"+err.Error(), 500)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		log.Printf("MainWasmHandler: File stat error:\n%v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, "MainWasmHandler: File stat error:\n"+err.Error(), 500)
		return
	}

	http.ServeContent(w, r, r.URL.Path, st.ModTime(), f)

}

// WasmExecJSHandler calls WasmCompiler.WasmExecJS and responds with the resulting .js file.
// WasmCompiler.WasmExecJS will only be called the first time and subsequent times
// will return the same result from memory.  (We're going to assume that you'll restart
// whatever process this is running in when upgrading your Go version.)
type WasmExecJSHandler struct {
	wc WasmExecJSer

	rwmu    sync.RWMutex
	content []byte
	modTime time.Time
}

// NewWasmExecJSHandler returns an initialized WasmExecJSHandler.
func NewWasmExecJSHandler(wc WasmExecJSer) *WasmExecJSHandler {
	return &WasmExecJSHandler{
		wc: wc,
	}
}

// ServeHTTP implements http.Handler.
func (h *WasmExecJSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	h.rwmu.RLock()
	content := h.content
	modTime := h.modTime
	h.rwmu.RUnlock()

	if content == nil {

		h.rwmu.Lock()
		defer h.rwmu.Unlock()

		rd, err := h.wc.WasmExecJS()
		if err != nil {
			log.Printf("error getting wasm_exec.js: %v", err)
			http.Error(w, "error getting wasm_exec.js: "+err.Error(), 500)
			return
		}

		b, err := ioutil.ReadAll(rd)
		if err != nil {
			log.Printf("error reading wasm_exec.js: %v", err)
			http.Error(w, "error reading wasm_exec.js: "+err.Error(), 500)
			return
		}

		h.content = b
		content = h.content
		h.modTime = time.Now()
		modTime = h.modTime

	}

	w.Header().Set("Content-Type", "text/javascript")
	http.ServeContent(w, r, r.URL.Path, modTime, bytes.NewReader(content))
}
