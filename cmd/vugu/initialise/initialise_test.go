package initialise

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanTemplateData(t *testing.T) {
	// fill the template
	d := newInitOpts()

	// WasmExecJSDir tests
	d.WasmExecJSDir = "/end/in/slash/"
	d.cleanTemplateData()

	if d.WasmExecJSDir != "/end/in/slash" {
		t.Fatal("Failed to remove training slash from \"/end/in/slash/\"")
	}

	d.WasmExecJSDir = ""
	d.cleanTemplateData()

	if d.WasmExecJSDir != "" {
		t.Fatalf("Failed to return an empty string when WasmExecJSDir is empty. Returned %q", d.WasmExecJSDir)
	}

	d.WasmExecJSDir = "."
	d.cleanTemplateData()

	if d.WasmExecJSDir != "" {
		t.Fatalf("Failed to return empty string when WasmExecJSDir was a period. Returned %q", d.WasmExecJSDir)
	}

	d.WasmExecJSDir = ".."
	d.cleanTemplateData()

	if d.WasmExecJSDir != ".." {
		t.Fatalf("Failed to leave unchanged when WasmExecJSDir was a .. Returned %q", d.WasmExecJSDir)
	}

	d.WasmExecJSDir = "/"
	d.cleanTemplateData()

	if d.WasmExecJSDir != "" {
		t.Fatalf("Failed to return empty string when WasmExecJSDir was a /. Returned %q", d.WasmExecJSDir)
	}

	// WasmMainDir tests
	d.WasmMainDir = "/end/in/slash/"
	d.cleanTemplateData()

	if d.WasmMainDir != "/end/in/slash" {
		t.Fatal("Failed to remove training slash from \"/end/in/slash/\"")
	}

	d.WasmMainDir = ""
	d.cleanTemplateData()

	if d.WasmMainDir != "" {
		t.Fatalf("Failed to return an empty string when WasmMainDir is empty. Returned %q", d.WasmMainDir)
	}

	d.WasmMainDir = "."
	d.cleanTemplateData()

	if d.WasmMainDir != "" {
		t.Fatalf("Failed to return empty string when WasmMainDir was a period. Returned %q", d.WasmMainDir)
	}

	d.WasmMainDir = ".."
	d.cleanTemplateData()

	if d.WasmMainDir != ".." {
		t.Fatalf("Failed to leave unchanged when WasmMainDir was a .. Returned %q", d.WasmMainDir)
	}

	d.WasmMainDir = "/"
	d.cleanTemplateData()

	if d.WasmMainDir != "" {
		t.Fatalf("Failed to return empty string when WasmMainDir was a /. Returned %q", d.WasmMainDir)
	}

	// WasmBinaryName tests
	d.WasmBinaryName = "/main.js/"
	d.cleanTemplateData()

	if d.WasmBinaryName != "main.js" {
		t.Fatal("Failed to remove training slash from \"/main.js/\"")
	}

	d.WasmBinaryName = "main.js/"
	d.cleanTemplateData()

	if d.WasmBinaryName != "main.js" {
		t.Fatal("Failed to remove training slash from \"main.js/\"")
	}

	d.WasmBinaryName = "."
	d.cleanTemplateData()

	if d.WasmBinaryName != defaultWasmBinaryName {
		t.Fatalf("Failed to return defaultWasmBinaryName when WasmMainDir was a period. Returned %q", d.WasmBinaryName)
	}

	d.WasmBinaryName = ".."
	d.cleanTemplateData()

	if d.WasmBinaryName != defaultWasmBinaryName {
		t.Fatalf("Failed to return defaultWasmBinaryName WasmMainDir was a .. Returned %q", d.WasmBinaryName)
	}

	d.WasmBinaryName = "/"
	d.cleanTemplateData()

	if d.WasmBinaryName != defaultWasmBinaryName {
		t.Fatalf("Failed to return defaultWasmBinaryName when WasmMainDir was a /. Returned %q", d.WasmBinaryName)
	}

	d.WasmBinaryName = ""
	d.cleanTemplateData()

	if d.WasmBinaryName != defaultWasmBinaryName {
		t.Fatalf("Failed to return defaultWasmBinaryName when WasmMainDir was an empty string. Returned %q", d.WasmBinaryName)
	}

}

func newInitOpts() InitOpts {
	return InitOpts{
		Dir:                     ".",
		PageTitle:               "pagetitle",
		WasmExecJSDir:           "",
		MountPoint:              "mountpoint",
		WasmMainDir:             "",
		WasmBinaryName:          "main.wasm",
		WasmGoFilename:          "main_wasm.go",
		RootStructPkgImportPath: "",
		RootStructPkgAlias:      "",
		RootStructType:          "Root",
		NoIndex:                 false,
		NoMain:                  false,
	}
}

func createTestDir(t *testing.T) string {
	t.Helper()
	d, err := os.MkdirTemp("", "vugu_init_test_") // put the dir under os.TempDir()
	if err != nil {
		t.Fatalf("Failed to create temp directory. %s", err)
	}
	return d
}

func removeTestDir(d string, t *testing.T) {
	t.Helper()
	err := os.RemoveAll(d)
	if err != nil {
		t.Fatalf("Failed to remove temp directory %s. %s", d, err)
	}
}

func runInitialise(opts InitOpts, t *testing.T) {
	t.Helper()
	err := doInitialise(context.Background(), "dummyModule", opts)
	if err != nil {
		t.Fatalf("Failed to initialise. %s", err)
	}
}

type initOptsDummy struct {
	InitOpts
}

func (o *initOptsDummy) checkFilesAreGenerated(path string, d fs.DirEntry, err error) error {
	// the generated file name should be one of
	var rootGoFileName string
	var rootVuguFileName string
	if o.RootStructType != "Root" {
		rootGoFileName = strings.ToLower(o.RootStructType) + ".go"
		rootVuguFileName = strings.ToLower(o.RootStructType) + ".vugu"
	}
	if !d.IsDir() {
		switch d.Name() {
		case "index.html":
			// we need to check the contents of the index.html
			return o.checkIndexHtmlContents(d.Name())
		case o.WasmGoFilename:
			return o.checkWasmGoContents(d.Name())
		case rootGoFileName:
			return o.checkRootGoContents(d.Name())
		case rootVuguFileName:
			return nil
		}
	}
	return nil // should return a dir
}

func (o *initOptsDummy) checkIndexHtmlContents(filename string) error {
	// open the file
	b, err := os.ReadFile(o.Dir + "/" + filename)
	if err != nil {
		return err
	}
	// the file should contain the page title
	title := "<title>" + o.PageTitle + "</title>"
	found := bytes.Contains(b, []byte(title))
	if !found {
		return fmt.Errorf("Could not find PageTitle (%q) in the index.html searched for %q", o.PageTitle, title)
	}

	// check the mount point
	mp := "<div id=" + o.MountPoint + "\">" // note we only check for the div
	found = bytes.Contains(b, []byte(mp))
	if !found {
		return fmt.Errorf("Could not find MontPoint (%q) in the index.html searched for %q", o.MountPoint, mp)
	}

	// check the file contains the main wasm file
	binaryName := "fetch(\"" + o.WasmMainDir + "/" + o.WasmBinaryName + "\""
	found = bytes.Contains(b, []byte(binaryName))
	if !found {
		return fmt.Errorf("Could not find binary name (%q) in the index.html searched for %q", binaryName, binaryName)
	}

	// check the file contains the wasm js dir, where the wasm_exec.js is stored - the wasm-exec.js must come form the Go distribution
	wasmExecJSDir := "<script src=\"" + o.WasmExecJSDir + "/wasm_exec.js\"></script>"
	found = bytes.Contains(b, []byte(wasmExecJSDir))
	if !found {
		return fmt.Errorf("Could not find wasm_exec.js directory (%q) in the index.html searched for %q", o.WasmExecJSDir, wasmExecJSDir)
	}
	return nil
}

func (o *initOptsDummy) checkWasmGoContents(filename string) error {
	// open the file
	b, err := os.ReadFile(o.Dir + "/" + filename)
	if err != nil {
		return err
	}
	// check the mount point
	mp := "mountPoint := \"#\" + " + o.MountPoint // the paces are significant as this is matching a generated go variable declaration
	found := bytes.Contains(b, []byte(mp))
	if !found {
		return fmt.Errorf("Could not find Mount Point (%q) in the %s searched for %q", o.MountPoint, o.WasmGoFilename, mp)
	}

	// check the import line (with and without an alias)
	var imp string
	if o.RootStructPkgAlias != "" {
		imp = o.RootStructPkgAlias + " " + o.RootStructPkgImportPath + "\"" + o.RootStructPkgImportPath + "\""
	}
	if o.RootStructPkgImportPath != "" {
		imp = "\"" + o.RootStructPkgImportPath + "\""
	}
	found = bytes.Contains(b, []byte(imp))
	if !found {
		return fmt.Errorf("Could not find import (%q) in the %s searched for %q", imp, o.WasmGoFilename, imp)
	}

	// check the root builder
	var rb string
	if o.RootStructPkgAlias != "" {
		rb = "&" + o.RootStructPkgImportPath + "." + o.RootStructType + "{}"
	} else {
		rb = "&" + o.RootStructType + "{}"
	}
	found = bytes.Contains(b, []byte(rb))
	if !found {
		return fmt.Errorf("Could not find RootBuilder (%q) in the %s searched for %q", rb, o.WasmGoFilename, rb)
	}
	return nil
}

func (o *initOptsDummy) checkRootGoContents(filename string) error {
	// open the file
	b, err := os.ReadFile(o.Dir + "/" + filename)
	if err != nil {
		return err
	}

	var rootGoFileName string
	if o.RootStructType != "Root" {
		rootGoFileName = strings.ToLower(o.RootStructType) + ".go"
	}
	// check the struct definition
	sd := "type " + o.RootStructType + " struct"
	found := bytes.Contains(b, []byte(sd))
	if !found {
		return fmt.Errorf("Could not find root struct (%q) declaration in the %s searched for %q", o.RootStructPkgAlias, rootGoFileName, sd)
	}
	return nil
}

func TestInitiaiseDefaults(t *testing.T) {

	o := newInitOpts()
	opts := initOptsDummy{o}
	opts.Dir = createTestDir(t)
	t.Cleanup(func() { removeTestDir(opts.Dir, t) })

	runInitialise(opts.InitOpts, t)

	// check we have what we expect.
	filepath.WalkDir(opts.Dir, opts.checkFilesAreGenerated)

}

func TestInitiaiseRootStructType(t *testing.T) {

	o := newInitOpts()
	opts := initOptsDummy{o}
	opts.Dir = createTestDir(t)
	opts.RootStructType = "Banana"
	t.Cleanup(func() { removeTestDir(opts.Dir, t) })

	runInitialise(opts.InitOpts, t)

	// check we have what we expect.
	filepath.WalkDir(opts.Dir, opts.checkFilesAreGenerated)

}

func TestInitiaiseWasmBinaryName(t *testing.T) {

	o := newInitOpts()
	opts := initOptsDummy{o}
	opts.Dir = createTestDir(t)
	opts.RootStructType = "Banana"
	opts.WasmBinaryName = "goldfish.wasm"
	t.Cleanup(func() { removeTestDir(opts.Dir, t) })

	runInitialise(opts.InitOpts, t)

	// check we have what we expect.
	filepath.WalkDir(opts.Dir, opts.checkFilesAreGenerated)

}

func TestInitiaiseWasmGoFilename(t *testing.T) {

	o := newInitOpts()
	opts := initOptsDummy{o}
	opts.Dir = createTestDir(t)
	opts.RootStructType = "Banana"
	opts.WasmBinaryName = "goldfish.wasm"
	opts.WasmGoFilename = "cherry.go"
	t.Cleanup(func() { removeTestDir(opts.Dir, t) })

	runInitialise(opts.InitOpts, t)

	// check we have what we expect.
	filepath.WalkDir(opts.Dir, opts.checkFilesAreGenerated)

}

func TestInitiaiseWasmMainDir(t *testing.T) {

	o := newInitOpts()
	opts := initOptsDummy{o}
	opts.Dir = createTestDir(t)
	opts.RootStructType = "Banana"
	opts.WasmBinaryName = "goldfish.wasm"
	opts.WasmGoFilename = "cherry.go"
	opts.WasmMainDir = "root_dir"
	t.Cleanup(func() { removeTestDir(opts.Dir, t) })

	runInitialise(opts.InitOpts, t)

	// check we have what we expect.
	filepath.WalkDir(opts.Dir, opts.checkFilesAreGenerated)

}

func TestInitiaisePageTitle(t *testing.T) {

	o := newInitOpts()
	opts := initOptsDummy{o}
	opts.Dir = createTestDir(t)
	opts.RootStructType = "Banana"
	opts.WasmBinaryName = "goldfish.wasm"
	opts.WasmGoFilename = "cherry.go"
	opts.WasmMainDir = "root_dir"
	opts.PageTitle = "This pages does nto exist!"

	t.Cleanup(func() { removeTestDir(opts.Dir, t) })

	runInitialise(opts.InitOpts, t)

	// check we have what we expect.
	filepath.WalkDir(opts.Dir, opts.checkFilesAreGenerated)

}

func TestInitiaiseMountPoint(t *testing.T) {

	o := newInitOpts()
	opts := initOptsDummy{o}
	opts.Dir = createTestDir(t)
	opts.RootStructType = "Banana"
	opts.WasmBinaryName = "goldfish.wasm"
	opts.WasmGoFilename = "cherry.go"
	opts.WasmMainDir = "root_dir"
	opts.PageTitle = "This pages does nto exist!"
	opts.MountPoint = "This_is_not_a_Mount_Point"

	t.Cleanup(func() { removeTestDir(opts.Dir, t) })

	runInitialise(opts.InitOpts, t)

	// check we have what we expect.
	filepath.WalkDir(opts.Dir, opts.checkFilesAreGenerated)

}

func TestInitiaiseRootStructImportPath(t *testing.T) {

	o := newInitOpts()
	opts := initOptsDummy{o}
	opts.Dir = createTestDir(t)
	opts.RootStructType = "Banana"
	opts.WasmBinaryName = "goldfish.wasm"
	opts.WasmGoFilename = "cherry.go"
	opts.WasmMainDir = "root_dir"
	opts.PageTitle = "This pages does nto exist!"
	opts.MountPoint = "This_is_not_a_Mount_Point"
	opts.RootStructPkgImportPath = "github.com/junk/thingy"

	t.Cleanup(func() { removeTestDir(opts.Dir, t) })

	runInitialise(opts.InitOpts, t)

	// check we have what we expect.
	filepath.WalkDir(opts.Dir, opts.checkFilesAreGenerated)

}

func TestInitiaiseRootStructImportPathAlias(t *testing.T) {

	o := newInitOpts()
	opts := initOptsDummy{o}
	opts.Dir = createTestDir(t)
	opts.RootStructType = "Banana"
	opts.WasmBinaryName = "goldfish.wasm"
	opts.WasmGoFilename = "cherry.go"
	opts.WasmMainDir = "root_dir"
	opts.PageTitle = "This pages does nto exist!"
	opts.MountPoint = "This_is_not_a_Mount_Point"
	opts.RootStructPkgImportPath = "github.com/junk/thingy"
	opts.RootStructPkgAlias = "silly_me"

	t.Cleanup(func() { removeTestDir(opts.Dir, t) })

	runInitialise(opts.InitOpts, t)

	// check we have what we expect.
	filepath.WalkDir(opts.Dir, opts.checkFilesAreGenerated)

}
