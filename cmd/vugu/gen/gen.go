package gen

import (
	"context"
	"path/filepath"

	"github.com/urfave/cli/v3"
	"github.com/vugu/vugu/gen"
)

var (
	Opts      gen.ParserGoPkgOpts
	Recursive bool
)

// Generates the code from the "*.vugu" and "*.htnml" files.
// This function is a direct copy of cmd/vugugen/vugugen.go, no
// functionality has been changed.
func Gen(ctx context.Context, cmd *cli.Command) error {
	// we need to get the argumets from the command as a slice.
	// The only argumnt would be the directory to run in.
	args := cmd.Args().Slice()

	// default to current directory
	if len(args) == 0 {
		args = []string{"."}
	}

	for _, arg := range args {

		pkgPath := arg
		var err error
		pkgPath, err = filepath.Abs(pkgPath)
		if err != nil {
			return err
		}

		if Recursive {
			err = gen.RunRecursive(pkgPath, &Opts)
		} else {
			err = gen.Run(pkgPath, &Opts)
		}
		if err != nil {
			return err
		}

	}
	return nil
}
