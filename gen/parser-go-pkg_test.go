package gen

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGoPkgRun(t *testing.T) {

	assert := assert.New(t)

	tmpDir, err := ioutil.TempDir("", "TestParseGoPkgRun")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 	assert.NoError(ioutil.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(`
	// module main
	// `), 0644))

	assert.NoError(ioutil.WriteFile(filepath.Join(tmpDir, "root.vugu"), []byte(`
<div id="root_comp">
	<h1>Hello!</h1>
</div>
`), 0644))

	p := NewParserGoPkg(tmpDir, nil)

	assert.NoError(p.Run())

	b, err := ioutil.ReadFile(filepath.Join(tmpDir, "root.go"))
	assert.NoError(err)
	log.Printf("OUT FILE root.go: %s", b)

}
