package gen

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/vugu/xxhash"
)

// ParserGoPkg knows how to perform source file generation in relation to a package folder.
// Whereas ParserGo handles converting a single template, ParserGoPkg is a higher level interface
// and provides the functionality of the vugugen command line tool.  It will scan a package
// folder for .vugu files and convert them to .go, with the appropriate defaults and logic.
type ParserGoPkg struct {
	pkgPath string
	opts    ParserGoPkgOpts
}

// ParserGoPkgOpts is the options for ParserGoPkg.
type ParserGoPkgOpts struct {
	SkipGoMod        bool    // do not try and create go.mod if it doesn't exist
	SkipMainGo       bool    // do not try and create main_wasm.go if it doesn't exist in a main package
	TinyGo           bool    // emit code intended for TinyGo compilation
	GoFileNameAppend *string // suffix to append to file names, after base name plus .go, if nil then "_vgen" is used
	MergeSingle      bool    // merge all output files into a single one
	MergeSingleName  string  // name of merged output file, only used if MergeSingle is true, defaults to "0_components_vgen.go"
}

// TODO: CallVuguSetup bool // always call vuguSetup instead of trying to auto-detect it's existence

var errNoVuguFile = errors.New("no .vugu file(s) found")

// RunRecursive will create a new ParserGoPkg and call Run on it recursively for each
// directory under pkgPath.  The opts will be modified for subfolders to disable go.mod and main.go
// logic.  If pkgPath does not contain a .vugu file this function will return an error.
func RunRecursive(pkgPath string, opts *ParserGoPkgOpts) error {

	if opts == nil {
		opts = &ParserGoPkgOpts{}
	}

	dirf, err := os.Open(pkgPath)
	if err != nil {
		return err
	}

	fis, err := dirf.Readdir(-1)
	if err != nil {
		return err
	}
	hasVugu := false
	var subDirList []string
	for _, fi := range fis {
		if fi.IsDir() && !strings.HasPrefix(fi.Name(), ".") {
			subDirList = append(subDirList, fi.Name())
			continue
		}
		if filepath.Ext(fi.Name()) == ".vugu" {
			hasVugu = true
		}
	}
	if !hasVugu {
		return errNoVuguFile
	}

	p := NewParserGoPkg(pkgPath, opts)
	err = p.Run()
	if err != nil {
		return err
	}

	for _, subDir := range subDirList {
		subPath := filepath.Join(pkgPath, subDir)
		opts2 := *opts
		// sub folders should never get these behaviors
		opts2.SkipGoMod = true
		opts2.SkipMainGo = true
		err := RunRecursive(subPath, &opts2)
		if err == errNoVuguFile {
			continue
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// Run will create a new ParserGoPkg and call Run on it.
func Run(pkgPath string, opts *ParserGoPkgOpts) error {
	p := NewParserGoPkg(pkgPath, opts)
	return p.Run()
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

// Opts returns the options.
func (p *ParserGoPkg) Opts() ParserGoPkgOpts {
	return p.opts
}

// Run does the work and generates the appropriate .go files from .vugu files.
// It will also create a go.mod file if not present and not SkipGoMod.  Same for main.go and SkipMainGo (will also skip
// if package already has file with package name something other than main).
// Per-file code generation is performed by ParserGo.
func (p *ParserGoPkg) Run() error {

	// record the times of existing files, so we can restore after if the same
	hashTimes, err := fileHashTimes(p.pkgPath)
	if err != nil {
		return err
	}

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

	goFnameAppend := "_vgen"
	if p.opts.GoFileNameAppend != nil {
		goFnameAppend = *p.opts.GoFileNameAppend
	}

	var mergeFiles []string

	mergeSingleName := "0_components_vgen.go"
	if p.opts.MergeSingleName != "" {
		mergeSingleName = p.opts.MergeSingleName
	}

	missingFmap := make(map[string]string, len(vuguFileNames))

	// run ParserGo on each file to generate the .go files
	for _, fn := range vuguFileNames {

		baseFileName := strings.TrimSuffix(fn, ".vugu")
		goFileName := baseFileName + goFnameAppend + ".go"
		compTypeName := fnameToGoTypeName(baseFileName)

		// keep track of which files to scan for missing structs
		missingFmap[fn] = goFileName

		mergeFiles = append(mergeFiles, goFileName)

		pg := &ParserGo{}

		pg.PackageName = pkgName
		// pg.ComponentType = compTypeName
		pg.StructType = compTypeName
		// pg.DataType = pg.ComponentType + "Data"
		pg.OutDir = p.pkgPath
		pg.OutFile = goFileName
		pg.TinyGo = p.opts.TinyGo

		// add to our list of names to check after
		namesToCheck = append(namesToCheck, pg.StructType)
		// namesToCheck = append(namesToCheck, pg.ComponentType+".NewData")
		// namesToCheck = append(namesToCheck, pg.DataType)
		namesToCheck = append(namesToCheck, "vuguSetup")

		// read in source
		b, err := ioutil.ReadFile(filepath.Join(p.pkgPath, fn))
		if err != nil {
			return err
		}

		// parse it
		err = pg.Parse(bytes.NewReader(b), fn)
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

	// if main package, generate main_wasm.go with default stuff if no main func in the package and no main_wasm.go
	if (!p.opts.SkipMainGo) && pkgName == "main" {

		mainGoPath := filepath.Join(p.pkgPath, "main_wasm.go")
		// log.Printf("namesFound: %#v", namesFound)
		// log.Printf("maingo found: %v", fileExists(mainGoPath))
		// if _, ok := namesFound["main"]; (!ok) && !fileExists(mainGoPath) {

		// NOTE: For now we're disabling the "main" symbol name check, because in single-dir cases
		// it's picking up the main_wasm.go in server.go (even though it's excluded via build tag).  This
		// needs some more thought but for now this will work for the common cases.
		if !fileExists(mainGoPath) {

			// log.Printf("WRITING TO main_wasm.go STUFF")
			var buf bytes.Buffer
			t, err := template.New("_main_").Parse(`// +build wasm
{{$opts := .Parser.Opts}}
package main

import (
	"fmt"
{{if not $opts.TinyGo}}
	"flag"
{{end}}

	"github.com/vugu/vugu"
	"github.com/vugu/vugu/domrender"
)

func main() {

{{if $opts.TinyGo}}
	var mountPoint *string
	{
		mp := "#vugu_mount_point"
		mountPoint = &mp
	}
{{else}}
	mountPoint := flag.String("mount-point", "#vugu_mount_point", "The query selector for the mount point for the root component, if it is not a full HTML component")
	flag.Parse()
{{end}}

	fmt.Printf("Entering main(), -mount-point=%q\n", *mountPoint)
	{{if not $opts.TinyGo}}defer fmt.Printf("Exiting main()\n")
{{end}}

	renderer, err := domrender.New(*mountPoint)
	if err != nil {
		panic(err)
	}
	{{if not $opts.TinyGo}}defer renderer.Release()
{{end}}

	buildEnv, err := vugu.NewBuildEnv(renderer.EventEnv())
	if err != nil {
		panic(err)
	}

{{if (index .NamesFound "vuguSetup")}}
	rootBuilder := vuguSetup(buildEnv, renderer.EventEnv())
{{else}}
	rootBuilder := &Root{}
{{end}}


	for ok := true; ok; ok = renderer.EventWait() {

		buildResults := buildEnv.RunBuild(rootBuilder)
		
		err = renderer.Render(buildResults)
		if err != nil {
			panic(err)
		}
	}
	
}
`)
			if err != nil {
				return err
			}
			err = t.Execute(&buf, map[string]interface{}{
				"Parser":     p,
				"NamesFound": namesFound,
			})
			if err != nil {
				return err
			}

			bufstr := buf.String()
			bufstr, err = gofmt(bufstr)
			if err != nil {
				log.Printf("WARNING: gofmt on main_wasm.go failed: %v", err)
			}

			err = ioutil.WriteFile(mainGoPath, []byte(bufstr), 0644)
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

	// remove the merged file so it doesn't mess with detection
	if p.opts.MergeSingle {
		os.Remove(filepath.Join(p.pkgPath, mergeSingleName))
	}

	// for _, fn := range vuguFileNames {

	// 	goFileName := strings.TrimSuffix(fn, ".vugu") + goFnameAppend + ".go"
	// 	goFilePath := filepath.Join(p.pkgPath, goFileName)

	// 	err := func() error {
	// 		// get ready to append to file
	// 		f, err := os.OpenFile(goFilePath, os.O_WRONLY|os.O_APPEND, 0644)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		defer f.Close()

	// 		// TODO: would be nice to clean this up and get a better grip on how we do this filename -> struct name mapping, but this works for now
	// 		compTypeName := fnameToGoTypeName(strings.TrimSuffix(goFileName, goFnameAppend+".go"))

	// 		// create CompName struct if it doesn't exist in the package
	// 		if _, ok := namesFound[compTypeName]; !ok {
	// 			fmt.Fprintf(f, "\ntype %s struct {}\n", compTypeName)
	// 		}

	// 		// // create CompNameData struct if it doesn't exist in the package
	// 		// if _, ok := namesFound[compTypeName+"Data"]; !ok {
	// 		// 	fmt.Fprintf(f, "\ntype %s struct {}\n", compTypeName+"Data")
	// 		// }

	// 		// create CompName.NewData with defaults if it doesn't exist in the package
	// 		// if _, ok := namesFound[compTypeName+".NewData"]; !ok {
	// 		// 	fmt.Fprintf(f, "\nfunc (ct *%s) NewData(props vugu.Props) (interface{}, error) { return &%s{}, nil }\n",
	// 		// 		compTypeName, compTypeName+"Data")
	// 		// }

	// 		// // register component unless disabled - nope, no more component registry
	// 		// if !p.opts.SkipRegisterComponentTypes && !fileHasInitFunc(goFilePath) {
	// 		// 	fmt.Fprintf(f, "\nfunc init() { vugu.RegisterComponentType(%q, &%s{}) }\n", strings.TrimSuffix(goFileName, ".go"), compTypeName)
	// 		// }

	// 		return nil
	// 	}()
	// 	if err != nil {
	// 		return err
	// 	}

	// }

	// generate anything missing and process vugugen comments
	mf := newMissingFixer(p.pkgPath, pkgName, missingFmap)
	err = mf.run()
	if err != nil {
		return fmt.Errorf("missing fixer error: %w", err)
	}

	// if requested, do merge
	if p.opts.MergeSingle {

		// if a missing fix file was produced include it in the list to be merged
		_, err := os.Stat(filepath.Join(p.pkgPath, "0_missing_vgen.go"))
		if err == nil {
			mergeFiles = append(mergeFiles, "0_missing_vgen.go")
		}

		err = mergeGoFiles(p.pkgPath, mergeSingleName, mergeFiles...)
		if err != nil {
			return err
		}
		// remove files if merge worked
		for _, mf := range mergeFiles {
			err := os.Remove(filepath.Join(p.pkgPath, mf))
			if err != nil {
				return err
			}
		}

	}

	err = restoreFileHashTimes(p.pkgPath, hashTimes)
	if err != nil {
		return err
	}

	return nil

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

func goGuessPkgName(pkgPath string) (ret string) {

	// defer func() { log.Printf("goGuessPkgName returning %q", ret) }()

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

// fileHashTimes will scan a directory and return a map of hashes and corresponding mod times
func fileHashTimes(dir string) (map[uint64]time.Time, error) {

	ret := make(map[uint64]time.Time)

	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fis, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		h := xxhash.New()
		fmt.Fprint(h, fi.Name()) // hash the name too so we don't confuse different files with the same contents
		b, err := ioutil.ReadFile(filepath.Join(dir, fi.Name()))
		if err != nil {
			return nil, err
		}
		h.Write(b)
		ret[h.Sum64()] = fi.ModTime()
	}

	return ret, nil
}

// restoreFileHashTimes takes the map returned by fileHashTimes and for any files where the hash
// matches we restore the mod time - this way we can clobber files during code generation but
// then if the resulting output is byte for byte the same we can just change the mod time back and
// things that look at timestamps will see the file as unchanged; somewhat hacky, but simple and
// workable for now - it's important for the developer experince we don't do unnecessary builds
// in cases where things don't change
func restoreFileHashTimes(dir string, hashTimes map[uint64]time.Time) error {

	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()

	fis, err := f.Readdir(-1)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		fiPath := filepath.Join(dir, fi.Name())
		h := xxhash.New()
		fmt.Fprint(h, fi.Name()) // hash the name too so we don't confuse different files with the same contents
		b, err := ioutil.ReadFile(fiPath)
		if err != nil {
			return err
		}
		h.Write(b)
		if t, ok := hashTimes[h.Sum64()]; ok {
			err := os.Chtimes(fiPath, time.Now(), t)
			if err != nil {
				log.Printf("Error in os.Chtimes(%q, now, %q): %v", fiPath, t, err)
			}
		}
	}

	return nil
}
