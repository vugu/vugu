package vugu

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ParserGoPkg knows how to perform source file generation in relation to a package folder.
// Whereas ParserGo handles converting a single template, ParserGoPkg is a higher level interface
// and provides the functionality of the vugugen command line tool.  It will scan a package
// folder for .vugu files and convert them to .go, with the appropriate defaults and logic.
type ParserGoPkg struct {
	pkgPath string
	opts    ParserGoPkgOpts
}

type ParserGoPkgOpts struct {
	SkipRegisterComponentTypes bool // indicates func init() { vugu.RegisterComponentType(...) } code should not be emitted in each file
	SkipGoMod                  bool // do not try and create go.mod if it doesn't exist
	SkipMainGo                 bool // do not try and create main.go if it doesn't exist in a main package
}

// NewParserGoPkg returns a new ParserGoPkg with the specified options or default if nil.  The pkgPath is required and must be an absolute path.
func NewParserGoPkg(pkgPath string, opts *ParserGoPkgOpts) *ParserGoPkg {
	ret := &ParserGoPkg{
		pkgPath: pkgPath,
	}
	if opts != nil {
		ret.opts = *opts
	}
	return ret
}

// Run does the work.
func (p *ParserGoPkg) Run() error {

	// vugugen path/to/package

	// comp-name.vugu
	// comp-name.go
	// tag is "comp-name"
	// component type is CompName
	// component data type is CompNameData
	// register unless disabled
	// create CompName if it doesn't exist in the package
	// create CompNameData if it doesn't exist in the package
	// create CompName.NewData with defaults if it doesn't exist in the package

	// how about a default main.go if one doesn't exist in the package? would be really useful!
	// also go.mod

	// flags:
	// * component registration
	// * skip generating go.mod

	// --

	pkgF, err := os.Open(p.pkgPath)
	if err != nil {
		return err
	}
	defer pkgF.Close()
	allFileNames, err := pkgF.Readdirnames(-1)
	if err != nil {
		return err
	}

	var vuguFileNames []string
	for _, fn := range allFileNames {
		if filepath.Ext(fn) == ".vugu" {
			vuguFileNames = append(vuguFileNames, fn)
		}
	}

	if len(vuguFileNames) == 0 {
		return fmt.Errorf("no .vugu files found, please create one and try again")
	}

	pkgName := goGuessPkgName(p.pkgPath)

	namesToCheck := []string{"main"}

	// run ParserGo on each file to generate the .go files
	for _, fn := range vuguFileNames {

		baseFileName := strings.TrimSuffix(fn, ".vugu")
		goFileName := baseFileName + ".go"
		compTypeName := fnameToGoTypeName(goFileName)

		pg := &ParserGo{}

		pg.PackageName = pkgName
		pg.ComponentType = compTypeName
		pg.DataType = pg.ComponentType + "Data"
		pg.OutDir = p.pkgPath
		pg.OutFile = goFileName

		// add to our list of names to check after
		namesToCheck = append(namesToCheck, pg.ComponentType)
		namesToCheck = append(namesToCheck, pg.ComponentType+".NewData")
		namesToCheck = append(namesToCheck, pg.DataType)

		// read in source
		b, err := ioutil.ReadFile(filepath.Join(p.pkgPath, fn))
		if err != nil {
			return err
		}

		// parse it
		err = pg.Parse(bytes.NewReader(b))
		if err != nil {
			return fmt.Errorf("error parsing %q: %v", fn, err)
		}

	}

	// after the code generation is done, check the package for the various names in question to see
	// what we need to generate
	namesFound, err := goPkgCheckNames(p.pkgPath, namesToCheck)
	if err != nil {
		return err
	}

	// if main package, generate main.go with default stuff if no main func in the package and no main.go
	if (!p.opts.SkipMainGo) && pkgName == "main" {

		mainGoPath := filepath.Join(p.pkgPath, "main.go")
		if _, ok := namesFound["main"]; (!ok) && !fileExists(mainGoPath) {

			err := ioutil.WriteFile(mainGoPath, []byte(`package `+pkgName+`
	
import (
	"fmt"
)

func main() {
	// TODO: some cool main stuff
	fmt.Printf("HERE1111\n")
}

`), 0644)
			if err != nil {
				return err
			}

		}

	}

	// write go.mod if it doesn't exist and not disabled - actually this really only makes sense for main,
	// otherwise we really don't know what the right module name is
	goModPath := filepath.Join(p.pkgPath, "go.mod")
	if pkgName == "main" && !p.opts.SkipGoMod && !fileExists(goModPath) {
		err := ioutil.WriteFile(goModPath, []byte(`module `+pkgName+"\n"), 0644)
		if err != nil {
			return err
		}
	}

	for _, fn := range vuguFileNames {

		goFileName := strings.TrimSuffix(fn, ".vugu") + ".go"
		goFilePath := filepath.Join(p.pkgPath, goFileName)

		err := func() error {
			// get ready to append to file
			f, err := os.OpenFile(goFilePath, os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return err
			}
			defer f.Close()

			compTypeName := fnameToGoTypeName(goFileName)

			// create CompName struct if it doesn't exist in the package
			if _, ok := namesFound[compTypeName]; !ok {
				fmt.Fprintf(f, "\ntype %s struct {}\n", compTypeName)
			}

			// create CompNameData struct if it doesn't exist in the package
			if _, ok := namesFound[compTypeName+"Data"]; !ok {
				fmt.Fprintf(f, "\ntype %s struct {}\n", compTypeName+"Data")
			}

			// create CompName.NewData with defaults if it doesn't exist in the package
			if _, ok := namesFound[compTypeName+".NewData"]; !ok {
				fmt.Fprintf(f, "\nfunc (ct *%s) NewData(props vugu.Props) (interface{}, error) { return &%s{}, nil }\n",
					compTypeName, compTypeName+"Data")
			}

			// register component unless disabled
			if !p.opts.SkipRegisterComponentTypes && !fileHasInitFunc(goFilePath) {
				fmt.Fprintf(f, "\nfunc init() { vugu.RegisterComponentType(%q, &%s{}) }\n", strings.TrimSuffix(goFileName, ".go"), compTypeName)
			}

			return nil
		}()
		if err != nil {
			return err
		}

	}

	return nil

	// // code to generate if a name check fails
	// type genItem struct {
	// 	CheckName string
	// 	OutSource string
	// }

	// log.Printf("namesFound: %#v", namesFound)

	// fset := token.NewFileSet()
	// pkgs, err := parser.ParseDir(fset, p.pkgPath, nil, 0)
	// if err != nil {
	// 	return err
	// }

	// if len(pkgs) != 1 {
	// 	return fmt.Errorf("unexpected package count after parsing, expected 1 and got this: %#v", pkgs)
	// }

	// var pkg *ast.Package
	// for _, pkg1 := range pkgs {
	// 	pkg = pkg1
	// }

	// for _, file := range pkg.Files {
	// 	log.Printf("file: %#v", file)
	// 	log.Printf("file.Scope.Objects: %#v", file.Scope.Objects)
	// 	log.Printf("next: %#v", file.Scope.Objects["Example1"])
	// 	// e1 := file.Scope.Objects["Example1"]
	// 	// if e1.Kind == ast.Typ {
	// 	// e1.Decl
	// 	// }
	// 	for _, d := range file.Decls {
	// 		if fd, ok := d.(*ast.FuncDecl); ok {
	// 			if fd.Recv != nil {
	// 				for _, f := range fd.Recv.List {
	// 					log.Printf("f.Type: %#v", f.Type)
	// 				}
	// 			}
	// 			log.Printf("fd.Name: %#v", fd.Name)
	// 		}
	// 	}
	// }
	// // log.Printf("Objects: %#v", pkg.Scope.Objects)

	// return nil
}

func fileHasInitFunc(p string) bool {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return false
	}
	// hacky but workable for now
	return regexp.MustCompile(`^func init\(`).Match(b)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return !os.IsNotExist(err)
}

func fnameToGoTypeName(s string) string {
	s = strings.Split(s, ".")[0] // remove file extension if present
	parts := strings.Split(s, "-")
	for i := range parts {
		p := parts[i]
		if len(p) > 0 {
			p = strings.ToUpper(p[:1]) + p[1:]
		}
		parts[i] = p
	}
	return strings.Join(parts, "")
}

func goGuessPkgName(pkgPath string) string {

	// see if the package already has a name and use it if so
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkgPath, nil, parser.PackageClauseOnly) // just get the package name
	if err != nil {
		goto checkMore
	}
	if len(pkgs) != 1 {
		goto checkMore
	}
	{
		var pkg *ast.Package
		for _, pkg1 := range pkgs {
			pkg = pkg1
		}
		return pkg.Name
	}

checkMore:

	// check for a root.vugu file, in which case we assume "main"
	_, err = os.Stat(filepath.Join(pkgPath, "root.vugu"))
	if err == nil {
		return "main"
	}

	// otherwise we use the name of the folder...
	dirBase := filepath.Base(pkgPath)
	if regexp.MustCompile(`^[a-z0-9]+$`).MatchString(dirBase) {
		return dirBase
	}

	// ...unless it makes no sense in which case we use "main"

	return "main"

}

// goPkgCheckNames parses a package dir and looks for names, returning a map of what was
// found.  Names like "A.B" mean a method of name "B" with receiver of type "*A"
// (so we can check for existence of a "NewData" method and whatever else)
func goPkgCheckNames(pkgPath string, names []string) (map[string]interface{}, error) {

	ret := make(map[string]interface{})

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkgPath, nil, 0)
	if err != nil {
		return ret, err
	}

	if len(pkgs) != 1 {
		return ret, fmt.Errorf("unexpected package count after parsing, expected 1 and got this: %#v", pkgs)
	}

	var pkg *ast.Package
	for _, pkg1 := range pkgs {
		pkg = pkg1
	}

	for _, file := range pkg.Files {

		if file.Scope != nil {
			for _, n := range names {
				if v, ok := file.Scope.Objects[n]; ok {
					ret[n] = v
				}
			}
		}

		// log.Printf("file: %#v", file)
		// log.Printf("file.Scope.Objects: %#v", file.Scope.Objects)
		// log.Printf("next: %#v", file.Scope.Objects["Example1"])
		// e1 := file.Scope.Objects["Example1"]
		// if e1.Kind == ast.Typ {
		// e1.Decl
		// }
		for _, d := range file.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {

				var drecv, dmethod string
				if fd.Recv != nil {
					for _, f := range fd.Recv.List {
						// log.Printf("f.Type: %#v", f.Type)
						if tstar, ok := f.Type.(*ast.StarExpr); ok {
							// log.Printf("tstar.X: %#v", tstar.X)
							if tstarXi, ok := tstar.X.(*ast.Ident); ok && tstarXi != nil {
								// log.Printf("namenamenamename: %#v", tstarXi.Name)
								drecv = tstarXi.Name
							}
						}
						// log.Printf("f.Names: %#v", f.Names)
						// for _, fn := range f.Names {
						// 	if fn != nil {
						// 		log.Printf("NAMENAME: %#v", fn.Name)
						// 		if fni, ok := fn.Name.(*ast.Ident); ok && fni != nil {
						// 		}
						// 	}
						// }

					}
				} else {
					continue // don't care methods with no receiver - found them already above as single (no period) names
				}

				// log.Printf("fd.Name: %#v", fd.Name)
				if fd.Name != nil {
					dmethod = fd.Name.Name
				}

				for _, n := range names {
					recv, method := nameParts(n)
					if drecv == recv && dmethod == method {
						ret[n] = d
					}
				}
			}
		}
	}
	// log.Printf("Objects: %#v", pkg.Scope.Objects)

	return ret, nil
}

func nameParts(n string) (recv, method string) {

	ret := strings.SplitN(n, ".", 2)
	if len(ret) < 2 {
		method = n
		return
	}
	recv = ret[0]
	method = ret[1]
	return
}
