package devutil

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMux(t *testing.T) {

	tmpFile, err := ioutil.TempFile("", "TestMux")
	must(err)
	tmpFile.Write([]byte("<html><body>contents of temp file</body></html>"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	m := NewMux().
		Exact("/blah", DefaultIndex).
		Exact("/tmpfile", StaticFilePath(tmpFile.Name())).
		Match(NoFileExt, StaticContent(`<html><body>NoFileExt test</body></html>`))

	// exact route with StaticContent
	wr := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/blah", nil)
	m.ServeHTTP(wr, r)
	checkStatus(t, r, wr.Result(), 200)
	checkBody(t, r, wr.Result(), "<script")
	checkHeader(t, r, wr.Result(), "Content-Type", "text/html; charset=utf-8")

	// exact route with StaticFilePath
	wr = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/tmpfile", nil)
	m.ServeHTTP(wr, r)
	checkStatus(t, r, wr.Result(), 200)
	checkBody(t, r, wr.Result(), "contents of temp file")
	checkHeader(t, r, wr.Result(), "Content-Type", "text/html; charset=utf-8")

	// NoFileExt
	wr = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/otherfile", nil)
	m.ServeHTTP(wr, r)
	checkStatus(t, r, wr.Result(), 200)
	checkBody(t, r, wr.Result(), "NoFileExt test")
	checkHeader(t, r, wr.Result(), "Content-Type", "text/html; charset=utf-8")

	// no default, 404
	wr = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/aintthere.css", nil)
	m.ServeHTTP(wr, r)
	checkStatus(t, r, wr.Result(), 404)
	checkBody(t, r, wr.Result(), "404 page not found")

	// default
	m.Default(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body>default overridden</body></body>"))
	}))
	wr = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/aintthere.css", nil)
	m.ServeHTTP(wr, r)
	checkBody(t, r, wr.Result(), "default overridden")

}
