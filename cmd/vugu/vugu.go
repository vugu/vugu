package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/vugu/vugu/cmd/vugu/version"
)

// The root `vugu` command entry point.
func main() {
	// Calculate the version string.
	// We want this as a separate string variable becase we need to assign it to the
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
		Version: v, // the version is printed under the VERSION: section if `vugu` is called without arguments.
		Commands: []*cli.Command{
			// The version sub command. See the version.Version
			// Also called in response to the "-v" and "--version" flags.
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Display the version number",
				Action:  version.Version,
			},
			// the future gen sub command, that should do the same as `vugugen` does currently
			// {
			// 	Name:    "gen",
			// 	Aliases: []string{"g"},
			// 	Usage:   "Generate code",
			// 	Action:  version.Version,
			// },
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
