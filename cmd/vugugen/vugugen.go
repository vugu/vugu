// vugugen is a command line tool to convert .vugu files into Go source code.
package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/vugu/vugu/gen"
)

// we basically just wrap ParserGoPKg
func main() {

	// vugugen path/to/package

	var opts gen.ParserGoPkgOpts
	flag.BoolVar(&opts.SkipGoMod, "skip-go-mod", false, "Do not try to create go.mod as needed")
	flag.BoolVar(&opts.SkipMainGo, "skip-main", false, "Do not try to create main.go as needed")
	flag.BoolVar(&opts.TinyGo, "tinygo", false, "Generate code intended for compilation under Tinygo")
	flag.BoolVar(&opts.MergeSingle, "s", false, "Merge generated code for a package into a single file.")
	recursive := flag.Bool("r", false, "Run recursively on specified path and subdirectories.")
	flag.Parse()

	args := flag.Args()

	// default to current directory
	if len(args) == 0 {
		args = []string{"."}
	}

	for _, arg := range args {

		pkgPath := arg
		var err error
		pkgPath, err = filepath.Abs(pkgPath)
		if err != nil {
			log.Fatal(err)
		}

		if *recursive {
			err = gen.RunRecursive(pkgPath, &opts)
		} else {
			err = gen.Run(pkgPath, &opts)
		}

		if err != nil {
			log.Fatal(err)
		}

	}

}
