package gen

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// missingFixer handles generating various missing types and methods.
// Looks at file structure and scans for `//vugugen:` comments
// See https://github.com/vugu/vugu/issues/128 for more explanation and background.
type missingFixer struct {
	pkgPath   string            // absolute path to package
	pkgName   string            // short name of package from the `package` statement
	vuguComps map[string]string // map of comp.vugu -> comp_vgen.go (all just relative base name of file, no dir)
	outfile   string            // file name of output file (relative), 0_missing_vgen.go by default
}

func newMissingFixer(pkgPath, pkgName string, vuguComps map[string]string) *missingFixer {
	return &missingFixer{
		pkgPath:   pkgPath,
		pkgName:   pkgName,
		vuguComps: vuguComps,
	}
}

// run does work for this one package
func (mf *missingFixer) run() error {

	// remove the output file if it doesn't exist,
	// and then below we re-create it if it turns out
	// we need it
	_ = mf.removeOutfile()

	// parse the package
	var fset token.FileSet
	pkgMap, err := parser.ParseDir(&fset, mf.pkgPath, nil, 0)
	if err != nil {
		return err
	}
	pkg := pkgMap[mf.pkgName]
	if pkg == nil {
		return fmt.Errorf("unable to find package %q after parsing dir %s", mf.pkgName, mf.pkgPath)
	}
	// log.Printf("pkg: %#v", pkg)

	var fout *os.File

	// read each _vgen.go file
	for _, goFile := range mf.vuguComps {

		// var ffset token.FileSet
		// file, err := parser.ParseFile(&ffset, filepath.Join(mf.pkgPath, goFile), nil, 0)
		// if err != nil {
		// 	return fmt.Errorf("error while reading %s: %w", goFile, err)
		// }
		// ast.Print(&ffset, file.Decls)

		file := fileInPackage(pkg, goFile)
		if file == nil {
			return fmt.Errorf("unable to find file %q in package (i.e. parse.ParseDir did not give us this file)", goFile)
		}

		compTypeName := findFileBuildMethodType(file)

		// if we didn't find a build method, we don't need to do anything
		if compTypeName == "" {
			continue
		}

		// log.Printf("found compTypeName=%s", compTypeName)

		// see if the type is already declared somewhere in the package
		compTypeDecl := findTypeDecl(&fset, pkg, compTypeName)

		// the type exists, we don't need to emit a declaration for it
		if compTypeDecl != nil {
			continue
		}

		// open outfile if it doesn't exist
		if fout == nil {
			fout, err = mf.createOutfile()
			if err != nil {
				return err
			}
			defer fout.Close()
		}

		fmt.Fprintf(fout, `// %s is a Vugu component and implements the vugu.Builder interface.
type %s struct {}

`, compTypeName, compTypeName)

		// log.Printf("aaa compTypeName=%s, compTypeDecl=%v", compTypeName, compTypeDecl)

	}

	// scan all .go files for known vugugen comments
	gcomments, err := readVugugenComments(mf.pkgPath)
	if err != nil {
		return err
	}

	// open outputfile if not already done above
	if len(gcomments) > 0 {
		if fout == nil {
			fout, err = mf.createOutfile()
			if err != nil {
				return err
			}
			defer fout.Close()
		}
	}

	gcommentFnames := make([]string, 0, len(gcomments))
	for fname := range gcomments {
		gcommentFnames = append(gcommentFnames, fname)
	}
	sort.Strings(gcommentFnames) // try to get deterministic output

	// for each file with vugugen comments
	for _, fname := range gcommentFnames {

		commentList := gcomments[fname]
		sort.Strings(commentList) // try to get deterministic output

		for _, c := range commentList {

			c := strings.TrimSpace(c)
			c = strings.TrimPrefix(c, "//vugugen:")

			cparts := strings.Fields(c) // split by whitespace

			if len(cparts) == 0 {
				return fmt.Errorf("error parsing %s vugugen comment with no type found %q", fname, c)
			}

			switch cparts[0] {

			case "event":

				args := cparts[1:]

				if len(args) < 1 {
					return fmt.Errorf("error parsing %s vugugen event comment with no args %q", fname, c)
				}

				eventName := args[0]

				if !unicode.IsUpper(rune(eventName[0])) {
					return fmt.Errorf("error parsing %s vugugen event comment, event name must start with a capital letter: %q", fname, c)
				}

				opts := args[1:]
				// isInterface := false

				// try to keep the option parsing very strict, especially for now before we get this
				// all figured out
				/*if len(opts) == 1 && opts[0] == "interface" {
					isInterface = true
				} else */
				if len(opts) == 0 {
					// no opts is fine
				} else {
					return fmt.Errorf("error parsing %s vugugen event comment unexpected options %q", fname, c)
				}

				// check for NameEvent
				decl := findTypeDecl(&fset, pkg, eventName+"Event")

				// emit type if missing as a struct wrapper around a DOMEvent
				if decl == nil {
					fmt.Fprintf(fout, `// %sEvent is a component event.
type %sEvent struct {
	vugu.DOMEvent
}

`, eventName, eventName)
				}

				// check for NameHandler type, emit if missing
				decl = findTypeDecl(&fset, pkg, eventName+"Handler")
				if decl == nil {
					fmt.Fprintf(fout, `// %sHandler is the interface for things that can handle %sEvent.
type %sHandler interface {
	%sHandle(event %sEvent)
}

`, eventName, eventName, eventName, eventName, eventName)
				}

				// check for NameFunc type, emit if missing along with method and type check
				decl = findTypeDecl(&fset, pkg, eventName+"Func")
				if decl == nil {
					fmt.Fprintf(fout, `// %sFunc implements %sHandler as a function.
type %sFunc func(event %sEvent)

// %sHandle implements the %sHandler interface.
func (f %sFunc) %sHandle(event %sEvent) { f(event) }

// assert %sFunc implements %sHandler
var _ %sHandler = %sFunc(nil)

`, eventName, eventName, eventName, eventName, eventName, eventName, eventName, eventName, eventName, eventName, eventName, eventName, eventName)
				}

			default:
				return fmt.Errorf("error parsing %s vugugen comment with unknown type %q", fname, c)
			}

		}

	}

	return nil
}

func (mf *missingFixer) fullOutfilePath() string {
	if mf.outfile == "" {
		return filepath.Join(mf.pkgPath, "0_missing_vgen.go")
	}
	return filepath.Join(mf.pkgPath, mf.outfile)
}

func (mf *missingFixer) removeOutfile() error {
	return os.Remove(mf.fullOutfilePath())
}

func (mf *missingFixer) createOutfile() (*os.File, error) {
	p := mf.fullOutfilePath()
	fout, err := os.Create(p)
	if err != nil {
		return nil, fmt.Errorf("failed to create missingFixer outfile %s: %w", p, err)
	}
	fmt.Fprintf(fout, "package %s\n\nimport \"github.com/vugu/vugu\"\n\nvar _ vugu.DOMEvent // import fixer\n\n", mf.pkgName)
	return fout, nil
}

// readVugugenComments will look in every .go file for a //vugugen: comment
// and return a map with file name keys and a slice of the comments found as the values.
// vugugen comment lines that are exactly identical will be deduplicated (even across files)
// as it will never be correct to generate two of the same thing in one package
func readVugugenComments(pkgPath string) (map[string][]string, error) {
	fis, err := ioutil.ReadDir(pkgPath)
	if err != nil {
		return nil, err
	}

	foundLines := make(map[string]bool)

	ret := make(map[string][]string, len(fis))
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		bname := filepath.Base(fi.Name())
		if !strings.HasSuffix(bname, ".go") {
			continue
		}
		f, err := os.Open(filepath.Join(pkgPath, bname))
		if err != nil {
			return ret, err
		}
		defer f.Close()

		br := bufio.NewReader(f)

		var fc []string

		pfx := []byte("//vugugen:")
		for {
			line, err := br.ReadBytes('\n')
			if err == io.EOF {
				if len(line) == 0 {
					break
				}
			} else if err != nil {
				return ret, fmt.Errorf("missingFixer error while reading %s: %w", bname, err)
			}
			// ignoring whitespace
			line = bytes.TrimSpace(line)
			// line must start with prefix exactly
			if !bytes.HasPrefix(line, pfx) {
				continue
			}
			// and not be a duplicate
			lineStr := string(line)
			if foundLines[lineStr] {
				continue
			}
			foundLines[lineStr] = true
			fc = append(fc, lineStr)
		}

		if fc != nil {
			ret[bname] = fc
		}

	}
	return ret, nil
}

// fileInPackage given pkg and "blah.go" will return the file whose base name is "blah.go"
// (i.e. it ignores the directory part of the map key in pkg.Files)
// Will return nil if not found.
func fileInPackage(pkg *ast.Package, fileName string) *ast.File {
	for fpath, file := range pkg.Files {
		if filepath.Base(fpath) == fileName {
			return file
		}
	}
	return nil
}

// findTypeDecl looks through the package for the given type and returns
// the declaraction or nil if not found
func findTypeDecl(fset *token.FileSet, pkg *ast.Package, typeName string) ast.Decl {
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {

			// ast.Print(fset, decl)

			// looking for genDecl
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			// which is a type declaration
			if genDecl.Tok != token.TYPE {
				continue
			}

			// with one TypeSpec
			if len(genDecl.Specs) != 1 {
				continue
			}
			spec, ok := genDecl.Specs[0].(*ast.TypeSpec)
			if !ok {
				continue
			}

			// with a name
			if spec.Name == nil {
				continue
			}

			// that matches the one we're looking for
			if spec.Name.Name == typeName {
				return genDecl
			}

		}
	}
	return nil
}

// findFileBuildMethodType will return "Comp" given `func (c *Root) Comp` exists in the file.
func findFileBuildMethodType(file *ast.File) string {

	for _, decl := range file.Decls {
		// only care about a function declaration
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		// named Build
		if funcDecl.Name.Name != "Build" {
			continue
		}
		// with exactly one receiver
		if !(funcDecl.Recv != nil && len(funcDecl.Recv.List) == 1) {
			continue
		}
		// which is a pointer
		recv := funcDecl.Recv.List[0]
		starExpr, ok := recv.Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		// to an identifier
		xident, ok := starExpr.X.(*ast.Ident)
		if !ok {
			continue
		}
		// whose name is the component type we're after
		return xident.Name
	}

	return ""
}
