package devutil

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWasmCompiler(t *testing.T) {

	tmpDir, err := ioutil.TempDir("", "TestWasmCompiler")
	must(err)
	defer os.RemoveAll(tmpDir)
	t.Logf("Using temporary dir: %s", tmpDir)

	wc := NewWasmCompiler().SetBuildDir(tmpDir)

	must(ioutil.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(`module TestWasmCompiler
`), 0644))

	// just build
	must(ioutil.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main
func main() {}`), 0644))
	outpath, err := wc.Execute()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outpath)
	_, err = os.Stat(outpath)
	if err != nil {
		t.Fatal(err)
	}

	// build with error
	ioutil.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main
func main() { not valid go code }`), 0644)
	outpath, err = wc.Execute()
	if err == nil {
		t.Fatal("should have gotten error here but didn't")
	}
	t.Logf("we correctly got a build error here: %v", err)

	// with generate
	must(ioutil.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main
import "fmt"
//go:generate go run gen.go
func main() { fmt.Println(other) }`), 0644))
	must(ioutil.WriteFile(filepath.Join(tmpDir, "gen.go"), []byte(`// +build ignore

package main
import "io/ioutil"
func main() { ioutil.WriteFile("./other.go", []byte("package main\nvar other = 123\n"), 0644) }`), 0644))
	wc.SetGenerateDir(tmpDir)
	outpath, err = wc.Execute()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outpath)
	_, err = os.Stat(outpath)
	if err != nil {
		t.Fatal(err)
	}

	// new temp dir
	tmpDir, err = ioutil.TempDir("", "TestWasmCompiler")
	must(err)
	defer os.RemoveAll(tmpDir)
	t.Logf("Using temporary dir: %s", tmpDir)
	wc = NewWasmCompiler().SetBuildDir(tmpDir)
	var h http.Handler

	must(ioutil.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(`module TestWasmCompiler
`), 0644))

	// main wasm handler without error
	must(ioutil.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main
func main() {}`), 0644))
	h = NewMainWasmHandler(wc)
	req, err := http.NewRequest("GET", "/main.wasm", nil)
	must(err)
	wr := httptest.NewRecorder()
	h.ServeHTTP(wr, req)
	res := wr.Result()
	defer res.Body.Close()
	// resb, _ := httputil.DumpResponse(res, true)
	// t.Logf("RESPONSE: %s", resb)
	if res.StatusCode != 200 {
		t.Errorf("unexpected http status code: %d", res.StatusCode)
	}
	ct := res.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/wasm") {
		t.Errorf("unexpected value for Content-Type header: %s", ct)
	}
	b, err := ioutil.ReadAll(res.Body)
	must(err)
	if bytes.Compare(b[:4], []byte("\x00asm")) != 0 {
		t.Errorf("got back bytes that do not look like a wasm file: %X (len=%d, cap=%d)", b[:4], len(b), cap(b))
	}

	// main wasm handler with error
	must(ioutil.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main
func main() { not valid go code }`), 0644))
	h = NewMainWasmHandler(wc)
	req, err = http.NewRequest("GET", "/main.wasm", nil)
	must(err)
	wr = httptest.NewRecorder()
	h.ServeHTTP(wr, req)
	res = wr.Result()
	defer res.Body.Close()
	// resb, _ := httputil.DumpResponse(res, true)
	// t.Logf("RESPONSE: %s", resb)
	if res.StatusCode != 500 {
		t.Errorf("unexpected http status code: %d", res.StatusCode)
	}
	ct = res.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("unexpected value for Content-Type header: %s", ct)
	}
	b, err = ioutil.ReadAll(res.Body)
	must(err)
	if !bytes.Contains(b, []byte("build error")) {
		t.Errorf("unexpected error result: %s", b)
	}

	// wasm exec js handler
	h = NewWasmExecJSHandler(wc)
	req, err = http.NewRequest("GET", "/wasm_exec.js", nil)
	must(err)
	wr = httptest.NewRecorder()
	h.ServeHTTP(wr, req)
	res = wr.Result()
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Errorf("unexpected http status code: %d", res.StatusCode)
	}
	ct = res.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/javascript") {
		t.Errorf("unexpected value for Content-Type header: %s", ct)
	}
	b, err = ioutil.ReadAll(res.Body)
	must(err)
	if !bytes.Contains(b, []byte("The Go Authors")) {
		t.Errorf("unexpected js result: %s", b)
	}

}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
