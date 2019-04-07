package vugufmt

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGoImportsNoError makes sure that the runGoImports function
// returns expected output when it deals with go code that
// is perfectly formatted. It uses all the .go files in this
// package to test against.
func TestGoImportsNoError(t *testing.T) {

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
		// run goimports on it
		out, err := runGoImports([]byte(testFileString))
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

// TestGoImportsError confirms that goimports is successfully detecting
// an error, and is reporting it in the expected format.
func TestGoImportsError(t *testing.T) {
	testCode := "package yeah\n\nvar hey := woo\n"
	// run goimports on it
	_, err := runGoImports([]byte(testCode))
	assert.NotNil(t, err)
	assert.Equal(t, 3, err.Line)
	assert.Equal(t, 9, err.Column)
}
