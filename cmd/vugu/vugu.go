package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/vugu/vugu/cmd/vugu/version"
)

// The root `vugu` command entry point.
func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			// The version sub command. See the version.Version
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
