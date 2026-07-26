package gen

import (
	"bytes"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const vuguExt = ".vugu"
const goExt = ".go"
const genExt = "_gen_js_wasm.go"

// A ErrNoComponentGoFile error is returned when the corresponding <component>.go file cannot be found in the same directory
// as the <component>.vugu file.
var ErrNoComponentGoFile = errors.New("no corresponding go file found")

// A ErrCouldNotParseVuguFile is returned when the <component>.vugu file cannot be parsed for any reason
var ErrCouldNotParseVuguFile = errors.New("could not parse .vugu file")

// A ErrCouldNotDeterminePackage is returned when the <component.go> file has been found but does not contain a valid package statement.
var ErrCouldNotDeterminePackage = errors.New("could not determine package from .go file")

// Generate will recursively turn each <component>.vugu file it finds into it's corresponding <component>_gen_js_wasm.go file
// starting from the directory supplied in the pkgPath parameter. Generate expects to find a <component>.go file in the same directory as the
// <component>.vugu file, and will place the <component>_gen_js_wasm.go file in the same directory.
//
// Any errors in this process are returned.
// If no vugu files are found no changes will be made. This is not considered an error condition.
//
// Generate will return any error returned by walkFunc.
// If a <component>.go file cannot be found [ErrNoComponentGoFile] will be returned
// If a <component>.vugu file cannot be parsed successfully [ErrCouldNotParseVuguFile] will be returned
// If the package cannot be determined from the <component>.go ErrCouldNotDeterminePackage will be returned.
// Any other error i.e. [os.PathError] comes from the underlying OS.
func Generate(pkgPath string) error {
	return filepath.WalkDir(pkgPath, walkFunc)
}

// walkFunc is called by [Generate] for each file starting
// As per [fs.WalkDirFunc] path is the full path filename (inc. directory(s) and file extension if any)
func walkFunc(path string, d fs.DirEntry, err error) error {
	// As per [fs.WalkDirFunc] walkFunc is called, with an error it is treated as a fatal error
	if err != nil {
		return fmt.Errorf("could not read directory: %w\n", err)
	}

	// check if we are called with a hidden file or directory
	hidden, herr := isHidden(path)
	if herr != nil {
		return herr
	}
	// skip hidden directories completely
	if d.IsDir() && hidden {
		return fs.SkipDir
	}
	// skip hidden files
	if hidden {
		return nil
	}
	// do we have a *.vugu file
	if !d.IsDir() && filepath.Ext(path) == vuguExt {
		// we do. Now we need to find a *.go file in the same place
		filenameNoExt := strings.TrimSuffix(path, vuguExt)
		goFilename := filenameNoExt + goExt
		// stat it to check it exists
		_, serr := os.Stat(goFilename)
		if serr != nil {
			return fmt.Errorf("%v: %w", path, ErrNoComponentGoFile)
		}
		// at this point we know the component.go file must exist this means ParseFile CANNOT fail because of a missing component.go file
		// generate the *_gen_js_wasm.go" filename by parsing it
		genFileName := filenameNoExt + genExt
		perr := ParseFile(path, goFilename, genFileName)
		if perr != nil {
			return fmt.Errorf("%v: %w\n", path, perr)
		}
	}
	return nil
}

// Run does the work and generates the appropriate _gen_js_wasm.go files from .vugu files.
// Per-file code generation is performed by ParserGo.
func ParseFile(vuguFilename, goFilename, genFilename string) error {
	// read the package name form the corresponding .go file
	// the .go file already exists because [walkFunc] stat'd it
	pkgName, err := findPackage(goFilename)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCouldNotDeterminePackage, err)
	}

	compTypeName := fnameToGoTypeName(filepath.Base(vuguFilename))

	pg := &ParserGo{}

	pg.PackageName = pkgName
	// pg.ComponentType = compTypeName
	pg.StructType = compTypeName
	// pg.DataType = pg.ComponentType + "Data"
	pg.OutFile = genFilename

	// read in source vugu file
	b, err := os.ReadFile(vuguFilename)
	if err != nil {
		fmt.Printf("ParserGoPkg.Run ReadFile error: %s\n", err)
		return err
	}

	// parse the vugu file
	err = pg.Parse(bytes.NewReader(b), vuguFilename)
	if err != nil {
		fmt.Printf("Parse returned: %v\n", err)
		return fmt.Errorf("%w: %w\n", ErrCouldNotParseVuguFile, err)
	}
	return err
}

func fnameToGoTypeName(s string) string {
	// Careful: this is making an assumption that we only take the portion of the filename up to the first period.
	// so file.name.vugu will turn into a Go struct name of "File"
	// This probally needs to be made clear in any docs.
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

func findPackage(goFile string) (string, error) {
	// the *.vugu and the *.go should exist in the same directory. [walkFunc] will have confirmed the later already.
	// The *.go contains the package name in the package statement, so use the Go ast to parse it out.
	fset := token.NewFileSet() // positions are relative to fset

	// Parse src but stop after processing the imports.
	// As per the [parser.ParseFile] docs, we will never see the case where f is nil and err is not nil because
	// we know for sure that the goFile exists already. This means that any error returned by [parser.ParseFile]
	// must relate to a syntax error (and by implication f will be non nil).
	f, err := parser.ParseFile(fset, goFile, nil, parser.PackageClauseOnly|parser.SkipObjectResolution)
	if err != nil {
		return "", err
	}

	return f.Name.Name, nil
}
