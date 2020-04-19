package gen

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// I tried to do this the "right" way using go/parser but ran into various strange behavior;
// I probably just am missing something with how to use it properly.  Regardless, doing this
// the hacky way should serve us just as well for now.
func mergeGoFiles(dir, out string, in ...string) error {

	var pkgClause string
	var importBlocks []string
	var otherBlocks []string

	sort.Strings(in) // try to get deterministic output

	// read and split each go file
	for _, fname := range in {
		fpath := filepath.Join(dir, fname)
		pkgPart, importPart, rest, err := readAndSplitGoFile(fpath)
		if err != nil {
			return fmt.Errorf("error trying to read and split Go file: %w", err)
		}

		if pkgClause == "" {
			pkgClause = pkgPart
		}

		importBlocks = append(importBlocks, importPart)
		otherBlocks = append(otherBlocks, rest)
	}

	var newPgm bytes.Buffer

	// use the package part from the first one
	newPgm.WriteString(pkgClause)
	newPgm.WriteString("\n\n")

	// concat the imports
	for _, bl := range importBlocks {
		newPgm.WriteString(bl)
		newPgm.WriteString("\n\n")
	}

	// concat the rest
	for _, bl := range otherBlocks {
		newPgm.WriteString(bl)
		newPgm.WriteString("\n\n")
	}

	// now read it back in using the parser and see if it will help us clean up the imports
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, out, newPgm.String(), parser.ParseComments)
	if err != nil {
		log.Printf("DEBUG: full merged file contents:\n%s", newPgm.String())
		return fmt.Errorf("error trying to parse merged file: %w", err)
	}
	ast.SortImports(fset, f)

	dedupAstFileImports(f)

	fileout, err := os.Create(filepath.Join(dir, out))
	if err != nil {
		return fmt.Errorf("error trying to open output file: %w", err)
	}
	defer fileout.Close()
	err = printer.Fprint(fileout, fset, f)
	if err != nil {
		return err
	}
	return nil

}

func readAndSplitGoFile(fpath string) (pkgPart, importPart, rest string, reterr error) {

	// NOTE: this is not perfect, it's only meant to be good enough to correctly parse the files
	// we generate, not any general .go file
	// (it does not understand multi-line comments, for example)

	var fullInput bytes.Buffer
	// defer func() {
	// log.Printf("readAndSplitGoFile(%q) full input:\n%s\n\nPKG:\n%s\n\nIMPORT:\n%s\n\nREST:\n%s\n\nErr:%v",
	// 	fpath,
	// 	fullInput.Bytes(),
	// 	pkgPart,
	// 	importPart,
	// 	rest,
	// 	reterr)
	// }()

	var pkgBuf, importBuf, restBuf bytes.Buffer
	var commentBuf bytes.Buffer

	const (
		inPkg = iota
		inImport
		inRest
	)
	state := inPkg

	f, err := os.Open(fpath)
	if err != nil {
		reterr = err
		return
	}
	defer f.Close()
	br := bufio.NewReader(f)
	i := 0
loop:
	for {
		i++
		line, err := br.ReadString('\n')
		if err == io.EOF {
			if len(line) == 0 {
				break
			}
		} else if err != nil {
			reterr = err
			return
		}
		fullInput.WriteString(line)

		lineFields := strings.Fields(line)
		var first string
		if len(lineFields) > 0 {
			first = lineFields[0]
		}

		_ = i
		// log.Printf("%s: iteration %d; lineFields=%#v", fpath, i, lineFields)

		switch state {

		case inPkg: // in package block, haven't see the package line yet
			pkgBuf.WriteString(line)
			if first == "package" {
				state = inImport
			}
			continue loop

		case inImport: // after package and are still getting what look like imports

			// hack to move line comments below the import area into the rest section - since
			// while we're going through there we can't know if there will be more imports or not
			if strings.HasPrefix(first, "//") {
				commentBuf.WriteString(line)
				continue loop
			}

			switch first {
			case "type", "func", "var":
				state = inRest

				restBuf.Write(commentBuf.Bytes())
				commentBuf.Reset()

				restBuf.WriteString(line)
				continue loop
			}

			importBuf.Write(commentBuf.Bytes())
			commentBuf.Reset()

			importBuf.WriteString(line)
			continue loop

			// // things we assume are part of the import block:
			// switch {
			// case strings.TrimSpace(first) == "": // blank line
			// case strings.HasPrefix(first, "//"): // line comment
			// case strings.HasPrefix(first, "import"): // import statement
			// case strings.HasPrefix(first, `"`): // should be a multi-line import package name
			// }

		case inRest:
			restBuf.WriteString(line)
			continue loop

		default:
		}

		panic("unreachable")

	}

	pkgPart = pkgBuf.String()
	importPart = importBuf.String()
	rest = restBuf.String()
	return
}

// // mergeGoFiles combines go source files into one.
// // dir is the package path, out and in are file names (no slashes, same directory).
// func mergeGoFiles(dir, out string, in ...string) error {

// 	pkgName := goGuessPkgName(dir)

// 	fset := token.NewFileSet()
// 	files := make(map[string]*ast.File)

// 	// parse all the files
// 	for _, name := range in {

// 		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
// 		if err != nil {
// 			return fmt.Errorf("error reading file %q: %w", name, err)
// 		}
// 		files[name] = f
// 	}

// 	pkg := &ast.Package{Name: pkgName, Files: files}
// 	fout := ast.MergePackageFiles(pkg,
// 		ast.FilterImportDuplicates, // this doesn't seem to be doing anything... sigh
// 	)

// 	// ast.SortImports(fset, fout)
// 	// ast.Print(fset, fout.Decls)
// 	moveImportsToTop(fout)

// 	dedupAstFileImports(fout)

// 	var buf bytes.Buffer
// 	printer.Fprint(&buf, fset, fout)

// 	return ioutil.WriteFile(filepath.Join(dir, out), buf.Bytes(), 0644)
// }

// func moveImportsToTop(f *ast.File) {

// 	var idecl []ast.Decl // import decls
// 	var odecl []ast.Decl // other decls

// 	// go through every declaration and move any imports into a separate list
// 	for _, decl := range f.Decls {

// 		{
// 			// import must be genDecl
// 			genDecl, ok := decl.(*ast.GenDecl)
// 			if !ok {
// 				goto notImport
// 			}

// 			// with token "import"
// 			if genDecl.Tok != token.IMPORT {
// 				goto notImport
// 			}

// 			idecl = append(idecl, decl)
// 			continue
// 		}

// 	notImport:
// 		odecl = append(odecl, decl)
// 		continue
// 	}

// 	// new decl list imports plus everything else
// 	f.Decls = append(idecl, odecl...)
// }
