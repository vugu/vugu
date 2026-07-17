package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/vugu/vugu/v2/cmd/vugu/gen"
	"github.com/vugu/vugu/v2/cmd/vugu/initialise"
	"github.com/vugu/vugu/v2/cmd/vugu/version"
)

// The root `vugu` command entry point.
func main() {
	// Calculate the version string.
	// We want this as a separate string variable because we need to assign it to the
	// cli.Command.Version field. Any errors need to be handled here.
	v, err := version.VersionString()
	if err != nil {
		fmt.Println(err)
	}
	// the string has a preable of "vugu version" which we need to chop off.
	// We only want the 3rd part of the returned space separated string
	v = strings.Split(v, " ")[2]

	// The VersionPrinter is called in response to a the "-v" and "--version" flags which
	// the cli package defines. These flags are enabled by default.
	// Changing the VersionPrinter allows us to change the format and method of calculating
	// the version number.
	// In this case we want the behaviour to be the same as the "version" subcommand.
	// To achieve this we set the VersionPrinter to the version.Version function.
	cli.VersionPrinter = func(cmd *cli.Command) {
		// call version.Version
		version.Version(context.Background(), cmd)
	}
	cmd := &cli.Command{
		Version:   v, // the version is printed under the VERSION: section if `vugu` is called without arguments.
		Usage:     "bootstraps and builds a vugu project",
		ArgsUsage: "[GO_MODULE_NAME]",
		Commands: []*cli.Command{
			// The version sub command. See the version.Version
			// Also called in response to the "-v" and "--version" flags.
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Display the version number",
				Action:  version.Version,
			},
			{
				Name:      "gen",
				Aliases:   []string{"g"},
				Usage:     "Generate the Go code from the .vugu files in the directory",
				ArgsUsage: "[OPTIONS] DIRECTORY",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "skip-go-mod",
						Value:       false,
						Usage:       "Do not try to create go.mod as needed",
						Destination: &gen.Opts.SkipGoMod,
					},
					&cli.BoolFlag{
						Name:        "r",
						Value:       false,
						Usage:       "Run recursively on specified path and subdirectories.",
						Destination: &gen.Recursive,
					},
				},
				Action: gen.Gen,
			},
			{
				Name:      "init",
				Aliases:   []string{"i"},
				Usage:     "Bootstrap a new vugu project",
				ArgsUsage: "[OPTIONS] GO_MODULE_NAME",
				//ArgsUsage: "GO_MODULE_NAME",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "dir",
						Value:       ".",
						Usage:       "Creates a project in the specified directory rather than the current one",
						Destination: &initialise.Opts.Dir,
					},
					&cli.StringFlag{
						Name:        "pagetitle",
						Value:       "Vugu Index Page",
						Usage:       "The page title of the generated index.html file",
						Destination: &initialise.Opts.PageTitle,
					},
					&cli.StringFlag{
						Name:        "wasmexecjsdir",
						Value:       "",
						Usage:       "The directory path on the web server (relative to the web document root) of the wasm_exec.js file. Any training slash will be removed. Relative paths are allowed. (default: an empty string)",
						Destination: &initialise.Opts.WasmExecJSDir,
					},
					&cli.StringFlag{
						Name:        "mountpoint",
						Value:       "vugu-mount-point",
						Usage:       "The id of the top level <div> that contains the wasm binary. Must be a valid HTML identifier", // the id can take any form according to the HTML spec https://html.spec.whatwg.org/multipage/dom.html#the-id-attribute so vugu does not check this.
						Destination: &initialise.Opts.MountPoint,
					},
					&cli.StringFlag{
						Name:        "wasmmaindir",
						Value:       "",
						Usage:       "The directory path on the web server (relative to the web document root) of the wasm binary file. Any training slash will be removed. Relative paths are allowed. (default: an empty string)",
						Destination: &initialise.Opts.WasmMainDir,
					},
					&cli.StringFlag{
						Name:        "wasmbinaryname",
						Value:       "main.wasm",
						Usage:       "The name of the wasm binary file.",
						Destination: &initialise.Opts.WasmBinaryName,
					},
					&cli.StringFlag{
						Name:        "wasmgofilename",
						Value:       "main_wasm.go",
						Usage:       "The name of the Go source code file that contains the main function.",
						Destination: &initialise.Opts.WasmGoFilename,
					},

					&cli.StringFlag{
						Name:        "rootstructpkgimportpath",
						Value:       "",
						Usage:       "The import path of the package that contains the root struct. If empty this is assumed to be the main package. (default: an empty string)",
						Destination: &initialise.Opts.RootStructPkgImportPath,
					},
					&cli.StringFlag{
						Name:        "rootstructpkgalias",
						Value:       "",
						Usage:       "The alias for the package that contains the root struct. If empty then the package name from the import path is used. . (default: an empty string)",
						Destination: &initialise.Opts.RootStructPkgAlias,
					},
					&cli.StringFlag{
						Name:        "rootstructtype",
						Value:       "Root",
						Usage:       "The type name of the root struct. The type must exist in the --rootstructpkg if supplied. Otherwise it must exist in the main package.",
						Destination: &initialise.Opts.RootStructType,
					},
					&cli.BoolFlag{
						Name:        "noindex",
						Value:       false,
						Usage:       "Do not generate an index.html file at the root of the module directory.",
						Destination: &initialise.Opts.NoIndex,
					},
					&cli.BoolFlag{
						Name:        "nomain",
						Value:       false,
						Usage:       "Do not generate an main_wasm.go file at the root of the module directory.",
						Destination: &initialise.Opts.NoMain,
					},
				},
				Action: initialise.Initialise, // don't use Init or init so as not to confuse with the package initialisation function "init()"
			},

			// Add other command here e.g. init, possibly with their own sub commands as shown in the comments
			// 	{
			// 		Name:    "init",
			// 		Aliases: []string{"i"},
			// 		Usage:   "Initialise project",
			// 		Commands: []*cli.Command{
			// 			{
			// 				Name:  "add",
			// 				Usage: "add a new template",
			// 				Action: func(ctx context.Context, cmd *cli.Command) error {
			// 					fmt.Println("new task template: ", cmd.Args().First())
			// 					return nil
			// 				},
			// 			},
			// 			{
			// 				Name:  "remove",
			// 				Usage: "remove an existing template",
			// 				Action: func(ctx context.Context, cmd *cli.Command) error {
			// 					fmt.Println("removed task template: ", cmd.Args().First())
			// 					return nil
			// 				},
			// 			},
			// 		},
			// 	},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
