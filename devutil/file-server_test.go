package devutil

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"path/filepath"
	"testing"
)

func TestFileServer(t *testing.T) {

	tmpDir, err := ioutil.TempDir("", "TestFileServer")
	must(err)
	defer os.RemoveAll(tmpDir)
	t.Logf("Using temporary dir: %s", tmpDir)

	fs := NewFileServer().SetDir(tmpDir)

	// redirect /dir to /dir/
	os.Mkdir(filepath.Join(tmpDir, "dir"), 0755)
	wr := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/dir", nil)
	fs.ServeHTTP(wr, r)
	checkStatus(t, r, wr.Result(), 301)
	checkHeader(t, r, wr.Result(), "Location", "dir/")

	// should error in a sane way on /dir/ if no listings
	wr = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/dir/", nil)
	fs.ServeHTTP(wr, r)
	checkStatus(t, r, wr.Result(), 404)

	// serve index.html from /dir/
	must(ioutil.WriteFile(filepath.Join(tmpDir, "dir/index.html"), []byte(`<html><body>index page here</body></html>`), 0644))
	wr = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/dir/", nil)
	fs.ServeHTTP(wr, r)
	checkStatus(t, r, wr.Result(), 200)
	checkBody(t, r, wr.Result(), "index page here")

	// listing for /dir/
	os.Remove(filepath.Join(tmpDir, "dir/index.html"))
	must(ioutil.WriteFile(filepath.Join(tmpDir, "dir/blerg.html"), []byte(`<html><body>blerg page here</body></html>`), 0644))
	fs.SetListings(true)
	wr = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/dir/", nil)
	fs.ServeHTTP(wr, r)
	checkStatus(t, r, wr.Result(), 200)
	checkBody(t, r, wr.Result(), "blerg.html")
	checkHeader(t, r, wr.Result(), "Content-Type", "text/html; charset=utf-8")

	fs.SetListings(false)

	// /a.html should serve a.html
	must(ioutil.WriteFile(filepath.Join(tmpDir, "a.html"), []byte(`<html><body>a page here</body></html>`), 0644))
	wr = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/a.html", nil)
	fs.ServeHTTP(wr, r)
	checkStatus(t, r, wr.Result(), 200)
	checkBody(t, r, wr.Result(), "a page here")
	checkHeader(t, r, wr.Result(), "Content-Type", "text/html; charset=utf-8")

	// /a should also serve a.html
	wr = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/a", nil)
	fs.ServeHTTP(wr, r)
	checkStatus(t, r, wr.Result(), 200)
	checkBody(t, r, wr.Result(), "a page here")
	checkHeader(t, r, wr.Result(), "Content-Type", "text/html; charset=utf-8")

	// not found should serve 404.html if present
	must(ioutil.WriteFile(filepath.Join(tmpDir, "404.html"), []byte(`<html><body>custom not found page here</body></html>`), 0644))
	wr = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/ainthere", nil)
	fs.ServeHTTP(wr, r)
	checkStatus(t, r, wr.Result(), 404)
	checkBody(t, r, wr.Result(), "custom not found page here")
	checkHeader(t, r, wr.Result(), "Content-Type", "text/html; charset=utf-8")

	// default not found
	os.Remove(filepath.Join(tmpDir, "404.html"))
	wr = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/ainthere", nil)
	fs.ServeHTTP(wr, r)
	checkStatus(t, r, wr.Result(), 404)
	checkBody(t, r, wr.Result(), "404 page not found")

	// custom not found
	fs.SetNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte("some other response here"))
	}))
	wr = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/ainthere", nil)
	fs.ServeHTTP(wr, r)
	checkStatus(t, r, wr.Result(), 403)
	checkBody(t, r, wr.Result(), "some other response here")

}

func checkBody(t *testing.T, req *http.Request, res *http.Response, text string) {
	b, err := httputil.DumpResponse(res, true)
	if err != nil {
		t.Logf("response dump failed: %v", err)
		return
	}
	if !bytes.Contains(b, []byte(text)) {
		t.Errorf("for %q expected response body to contain %q but it did not, full body: %s", req.URL.Path, text, b)
	}

}

func checkStatus(t *testing.T, req *http.Request, res *http.Response, status int) {
	st := res.StatusCode
	if st != status {
		t.Errorf("for %q expected status to be %v but got %v", req.URL.Path, status, st)
	}
}

func checkHeader(t *testing.T, req *http.Request, res *http.Response, key, val string) {
	hval := res.Header.Get(key)
	if hval != val {
		t.Errorf("for %q expected header %q to be %q but got %q", req.URL.Path, key, val, hval)
	}
}
