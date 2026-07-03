package initialise

import (
	"bufio"
	"context"
	"embed"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
	"github.com/vugu/vugu/v2/cmd/vugu/sh"
)

type InitOpts struct {
	Dir string
	// The title of the html page
	PageTitle string
	// The directory where the wasm_exec.js file is located on the web server.
	// Any training slash will be stripped.
	WasmExecJSDir string
	// The name of the div id i.e. <div id="..."> which contains the wasm binary
	MountPoint string
	// The directory where the wasm binary is located on the web server.
	// Any training slash will be stripped.
	WasmMainDir string
	// The name of the wasm binary
	WasmBinaryName string
	// the name of the Go source file that containts the main function
	WasmGoFilename string
	// The full import path of the package that contains the base/root type. This will be used by a go import so must be valid package import path
	RootStructPkgImportPath string
	// The name of package that containts the the base/root type. The package name will be used by a go import so must be valid package name
	RootStructPkgAlias string
	// The name of the base/root type. This struct is the man entry point for the wasm code. This must be a valid, exported type in the RootStructPkg
	RootStructType string
	// Do not generate an index.html at the root of the module directory if true
	NoIndex bool
	//Do not generate a main_wasm.go at teh root if the module directory is true
	NoMain bool
}

func (d *InitOpts) cleanTemplateData() {
	// clean the directories - we need to remove any trailing slash
	d.WasmExecJSDir = filepath.Clean(d.WasmExecJSDir)
	// clean can result in a empty path which Clean returns as "." or if a root a "/", but we really want the empty path, so flip this back
	if d.WasmExecJSDir == "." || d.WasmExecJSDir == string(os.PathSeparator) {
		d.WasmExecJSDir = ""
	}
	d.WasmMainDir = filepath.Clean(d.WasmMainDir)
	if d.WasmMainDir == "." || d.WasmMainDir == string(os.PathSeparator) {
		d.WasmMainDir = ""
	}
	d.WasmBinaryName = filepath.Base(d.WasmBinaryName)
	// base can return ".", ".." or "/" i.e. os.PathSeparator. If we get any of these we want the default name
	if d.WasmBinaryName == "." || d.WasmBinaryName == ".." || d.WasmBinaryName == string(os.PathSeparator) {
		d.WasmBinaryName = defaultWasmBinaryName
	}
}

const templateDirName = "templates"
const indexHTMLTmplName = templateDirName + "/" + "index.html.tmpl"
const mainWasmTmplName = templateDirName + "/" + "main_wasm.go.tmpl"
const rootDotGoTmplName = templateDirName + "/" + "root.go.tmpl"
const rootDotVuguName = templateDirName + "/" + "root.vugu"

const defaultWasmBinaryName = "main.wasm"

var (
	Opts InitOpts
	//go:embed templates
	content embed.FS
)

func Initialise(ctx context.Context, cmd *cli.Command) error {
	// we need to get the arguments from the command as a slice.
	// The only argument would be the module name to create.
	args := cmd.Args().Slice()

	if len(args) != 1 {
		return fmt.Errorf("init: too many arguments. Expected one but found %d.", len(args))
	}

	// if the RootStructAlias is set then the RootStructPkgImportPath must also be set
	if Opts.RootStructPkgAlias != "" && Opts.RootStructPkgImportPath == "" {
		return fmt.Errorf("The --rootstructpkgalis flag was set to %q but the --rootstructpkgimportpath was not set. Both flags are required is a package is to be aliased.", Opts.RootStructPkgAlias)
	}

	moduleName := args[0]
	fmt.Printf("ModuleName: %q\n", args[0])
	return doInitialise(ctx, moduleName, Opts)
}

func doInitialise(ctx context.Context, moduleName string, opts InitOpts) error {
	// Opts is a package level variable
	fmt.Printf("Dir Option %q\n", opts.Dir)

	// CD to the dir specified
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Printf("cwd: %s\n", cwd)
	defer os.Chdir(cwd) // ignore an error

	err = os.Chdir(opts.Dir)
	if err != nil {
		return err
	}

	opts.cleanTemplateData()

	// create the index.html
	if opts.NoIndex == false { // use ==false test rather than !NoIndex to make the indent clear. The default is false.
		err = createIndexHtml(content, opts)
		if err != nil {
			return err
		}
	}

	// create the main_wasm.go in the package
	if opts.NoMain == false { // use ==false test rather than !NoIndex to make the indent clear. The default is false.
		err = createMainWasmDotGo(content, opts)
		if err != nil {
			return err
		}
	}

	// create the root.vugu file
	// if a user has supplied an import path for the root component we assume they know what they are doing and that the root
	// components source (both .go and .vugu) are in the package
	if opts.RootStructPkgImportPath == "" { // No import prth set so we can generate a root.vugu and root.go files
		err = createRootDotVugu(content, opts)
		if err != nil {
			return err
		}
		// create the root.go file
		err = createRootDotGo(content, opts)
		if err != nil {
			return err
		}
	}
	// run "go mod init moduleName
	err = sh.RunV("go", "mod", "init", moduleName)

	return err
}

func createIndexHtml(fs fs.FS, opts InitOpts) error {
	tmpl, err := template.ParseFS(fs, indexHTMLTmplName)
	if err != nil {
		return err
	}
	indexHTML, err := os.Create(opts.Dir + "/index.html")
	if err != nil {
		return err
	}
	defer indexHTML.Close()
	err = tmpl.Execute(indexHTML, opts)
	if err != nil {
		return err
	}
	err = indexHTML.Sync() // ensure we flush to disk
	return err
}

func createMainWasmDotGo(fs fs.FS, opts InitOpts) error {
	tmpl, err := template.ParseFS(fs, mainWasmTmplName)
	if err != nil {
		return err
	}
	mainWasm, err := os.Create(opts.Dir + "/" + opts.WasmGoFilename) // how do we ensure this is a .go filename?
	if err != nil {
		return err
	}
	defer mainWasm.Close()

	// we must strip any training slash from the RootStructPkg name.
	// if its not an empty string we can use filepath.Clean to do this for us.

	if opts.RootStructPkgImportPath != "" {
		// if its an empty string the root struct is in the main package
		opts.RootStructPkgImportPath = filepath.Clean(opts.RootStructPkgImportPath)
	}
	// now we need the just the package name from the full import path.
	// we can use filepath.Base for this
	fmt.Printf("Before %q\n", opts.RootStructPkgAlias)

	if opts.RootStructPkgAlias == "" {
		// the package has not been aliased so we need to get the package name from the import path
		opts.RootStructPkgAlias = filepath.Base(opts.RootStructPkgImportPath)
		// make sure we are not left with "." as Base does if opts.RootStructPkgImportPath was an empty string
		if opts.RootStructPkgAlias == "." {
			opts.RootStructPkgAlias = ""
		}
	}
	fmt.Printf("After %q\n", opts.RootStructPkgAlias)

	err = tmpl.Execute(mainWasm, opts)
	if err != nil {
		return err
	}
	err = mainWasm.Sync() // ensure we flush to disk
	return err
}

func createRootDotVugu(fs fs.FS, opts InitOpts) error {
	rootDotVuguSrc, err := fs.Open(rootDotVuguName)
	if err != nil {
		return err
	}
	defer rootDotVuguSrc.Close()
	r := bufio.NewReader(rootDotVuguSrc)

	rootDotVuguDest, err := os.Create(opts.Dir + "/" + opts.RootStructType + ".vugu")
	if err != nil {
		return err
	}
	defer rootDotVuguDest.Close()
	w := bufio.NewWriter(rootDotVuguDest)
	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}
	err = rootDotVuguDest.Sync() // ensure we flush to disk
	return err
}

func createRootDotGo(fs fs.FS, opts InitOpts) error {
	tmpl, err := template.ParseFS(fs, rootDotGoTmplName)
	if err != nil {
		return err
	}
	rootDotGo, err := os.Create(opts.Dir + "/" + opts.RootStructType + ".go")
	if err != nil {
		return err
	}
	defer rootDotGo.Close()

	err = tmpl.Execute(rootDotGo, opts)
	if err != nil {
		return err
	}
	err = rootDotGo.Sync() // ensure we flush to disk
	return err
}
