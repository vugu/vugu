package tmpl

import (
	"html/template"
	"os"
	"testing"
)

func CreateIndexHtml(t *testing.T, pkgName string) {
	type TestPath struct {
		TestDir string
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("CWD: %q", cwd)
	tp := TestPath{TestDir: pkgName}

	tmpl, err := template.ParseFiles(cwd + "/index.html.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	// remove any existing "index.html" - we don't care if the file does not exist
	err = os.Remove(cwd + "/index.html")
	if err != nil {
		t.Logf("rm error (not fatal) %s", err)
	}
	indexHTML, err := os.Create(cwd + "/index.html")
	if err != nil {
		t.Fatal(err)
	}

	err = tmpl.Execute(indexHTML, tp)
	if err != nil {
		t.Fatal(err)
	}
	indexHTML.Sync() // ensure we flush to disk
	err = indexHTML.Close()
	if err != nil {
		t.Fatal(err)
	}

}
