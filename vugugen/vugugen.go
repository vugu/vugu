// vugugen is a command line tool to convert .vugu files into Go source code.
package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/vugu/vugu"
)

// we basically just wrap ParserGoPKg
func main() {

	// vugugen path/to/package

	var opts vugu.ParserGoPkgOpts
	flag.BoolVar(&opts.SkipRegisterComponentTypes, "skip-register", false, "Do not attempt to register component in init() method")
	flag.BoolVar(&opts.SkipGoMod, "skip-go-mod", false, "Do not try to create go.mod as needed")
	flag.BoolVar(&opts.SkipMainGo, "skip-main", false, "Do not try to create main.go as needed")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Printf("expected exactly one argument of package path but got %d args instead", len(args))
	}

	pkgPath := args[0]
	var err error
	pkgPath, err = filepath.Abs(pkgPath)
	if err != nil {
		log.Fatal(err)
	}

	p := vugu.NewParserGoPkg(pkgPath, &opts)

	err = p.Run()
	if err != nil {
		log.Fatal(err)
	}

}
