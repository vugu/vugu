package vugufmt

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptsCustom(t *testing.T) {
	jsFormat := func(f *Formatter) {
		f.ScriptFormatters["js"] = func(input []byte) ([]byte, *FmtError) {
			return nil, nil
		}
	}
	formatter := NewFormatter(jsFormat)
	assert.NotNil(t, formatter.ScriptFormatters["js"])
}

func TestOptsGoFmt(t *testing.T) {
	gofmt := UseGoFmt(false)
	formatter := NewFormatter(gofmt)
	assert.NotNil(t, formatter.ScriptFormatters["application/x-go"])
}

func TestOptsGoFmtSimple(t *testing.T) {
	gofmt := UseGoFmt(true)
	formatter := NewFormatter(gofmt)
	assert.NotNil(t, formatter.ScriptFormatters["application/x-go"])
}

func TestOptsGoImports(t *testing.T) {
	goimports := UseGoImports
	formatter := NewFormatter(goimports)
	assert.NotNil(t, goimports, formatter.ScriptFormatters["application/x-go"])
}

func TestVuguFmtNoError(t *testing.T) {
	formatter := NewFormatter(UseGoFmt(false))
	fmtr := func(f string) {
		// Need to un-relativize the paths
		absPath, err := filepath.Abs(f)

		if filepath.Ext(absPath) != ".vugu" {
			return
		}

		assert.NoError(t, err, f)
		// get a handle on the file
		testFile, err := ioutil.ReadFile(absPath)
		testFileString := string(testFile)
		assert.NoError(t, err, f)
		// run gofmt on it
		var buf bytes.Buffer
		err = formatter.FormatHTML(absPath, strings.NewReader(testFileString), &buf)
		assert.Nil(t, err, f)
		prettyVersion := buf.String()

		// make sure nothing changed!
		assert.NotNil(t, buf.String(), f)
		assert.Equal(t, testFileString, prettyVersion, f)

		//ioutil.WriteFile(absPath+".html", []byte(prettyVersion), 0644)
	}

	err := filepath.Walk("./testdata/ok/", func(path string, info os.FileInfo, err error) error {
		fmtr(path)
		return nil
	})

	assert.NoError(t, err)
}

func TestUncompilableGo(t *testing.T) {
	formatter := NewFormatter(UseGoFmt(false))
	fmtr := func(f string) {
		// Need to un-relativize the paths
		absPath, err := filepath.Abs(f)

		if filepath.Ext(absPath) != ".vugu" {
			return
		}

		assert.NoError(t, err, f)
		// get a handle on the file
		testFile, err := ioutil.ReadFile(absPath)
		testFileString := string(testFile)
		assert.NoError(t, err, f)
		// run gofmt on it
		var buf bytes.Buffer
		ferr := formatter.FormatHTML("oknow", strings.NewReader(testFileString), &buf)
		assert.NotNil(t, ferr, f)
		// confirm the offset is correct!
		assert.Equal(t, 46, ferr.Line, f)
		assert.Equal(t, 22, ferr.Column, f)
	}

	fmtr("./testdata/bad/badgo.vugu")

}

func TestEscaping(t *testing.T) {
	// I'd like the > to not be escaped into &gt;
	testCode := "<div vg-if='len(data.bpi.BPI) > 0'></div>"
	formatter := NewFormatter(UseGoFmt(false))
	// run gofmt on it
	var buf bytes.Buffer
	assert.Nil(t, formatter.FormatHTML("", strings.NewReader(testCode), &buf), testCode)
	prettyVersion := buf.String()
	assert.Equal(t, testCode, prettyVersion)
}

func TestBadHTML(t *testing.T) {
	// I'd like the > to not be escaped into &gt;
	testCode := "<html><head></head><body><oh no></body></html>"
	formatter := NewFormatter(UseGoFmt(false))
	// run gofmt on it
	var buf bytes.Buffer
	err := formatter.FormatHTML("", strings.NewReader(testCode), &buf)
	assert.Error(t, err, testCode)
	prettyVersion := buf.String()
	assert.NotEqual(t, testCode, prettyVersion)
}
