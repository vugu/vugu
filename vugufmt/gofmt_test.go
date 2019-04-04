package vugufmt

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGoFmtNoError makes sure that the runGoFmt function
// returns expected output when it deals with go code that
// is perfectly formatted. It uses all the .go files in this
// package to test against.
func TestGoFmtNoError(t *testing.T) {

	fmt := func(f string) {
		// Need to un-relativize the paths
		absPath, err := filepath.Abs(f)

		if filepath.Ext(absPath) != ".go" {
			return
		}

		assert.Nil(t, err, f)
		// get a handle on the file
		testFile, err := ioutil.ReadFile(absPath)
		testFileString := string(testFile)
		assert.Nil(t, err, f)
		// run gofmt on it
		out, err := runGoFmt([]byte(testFileString), false)
		assert.Nil(t, err, f)
		// make sure nothing changed!
		assert.NotNil(t, string(out), f)
		assert.Equal(t, testFileString, string(out), f)
	}

	err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		fmt(path)
		return nil
	})

	assert.NoError(t, err)
}

// TestGoFmtError confirms that gofmt is successfully detecting
// an error, and is reporting it in the expected format.
func TestGoFmtError(t *testing.T) {
	testCode := "package yeah\n\nvar hey := woo\n"
	// run gofmt on it
	_, err := runGoFmt([]byte(testCode), false)
	assert.NotNil(t, err)
	assert.Equal(t, 3, err.Line)
	assert.Equal(t, 9, err.Column)
}
