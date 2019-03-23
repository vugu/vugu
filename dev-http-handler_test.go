package vugu

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDevHTTPHandler(t *testing.T) {

	assert := assert.New(t)

	tmpDir, err := ioutil.TempDir("", "TestDevHTTPHandler")
	assert.NoError(err)
	// log.Printf("tmpDir = %q", tmpDir)
	defer os.RemoveAll(tmpDir)

	wd, _ := os.Getwd()

	// write a go.mod that points vugu module to use our local path instead of pulling from github
	assert.NoError(ioutil.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(`
module main
replace github.com/vugu/vugu => `+wd+`
	`), 0644))

	assert.NoError(ioutil.WriteFile(filepath.Join(tmpDir, "root.vugu"), []byte(`
<div>I Am Root</div>
`), 0644))

	assert.NoError(ioutil.WriteFile(filepath.Join(tmpDir, "test.js"), []byte(`
// test.js here
`), 0644))

	h := NewDevHTTPHandler(tmpDir, http.Dir(tmpDir))
	srv := httptest.NewServer(h)
	defer srv.Close()

	assert.Contains(mustGetPage(srv.URL+"/"), "<body")                      // index page
	assert.Contains(mustGetPage(srv.URL+"/other-page"), "<body")            // other HTML page
	assert.Contains(mustGetPage(srv.URL+"/test.js"), "// test.js here")     // static file
	assert.Contains(mustGetPage(srv.URL+"/wasm_exec.js"), "global.Go")      // Go WASM support js file
	assert.Contains(mustGetPage(srv.URL+"/does-not-exist.js"), "not found") // other misc not found file
	assert.Contains(mustGetPage(srv.URL+"/main.wasm"), "\x00asm")           // WASM binary should have marker

}

func mustGetPage(u string) string {
	res, err := http.Get(u)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	return string(b)
}
