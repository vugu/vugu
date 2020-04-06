package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"path/filepath"
)

// mergeGoFiles combines go source files into one.
// dir is the package path, out and in are file names (no slashes, same directory).
func mergeGoFiles(dir, out string, in ...string) error {

	pkgName := goGuessPkgName(dir)

	fset := token.NewFileSet()
	files := make(map[string]*ast.File)

	// parse all the files
	for _, name := range in {

		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("error reading file %q: %w", name, err)
		}
		files[name] = f
	}

	pkg := &ast.Package{Name: pkgName, Files: files}
	fout := ast.MergePackageFiles(pkg,
		ast.FilterImportDuplicates, // this doesn't seem to be doing anything... sigh
	)

	dedupAstFileImports(fout)

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, fout)

	return ioutil.WriteFile(filepath.Join(dir, out), buf.Bytes(), 0644)
}
