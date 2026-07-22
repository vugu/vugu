package gen

import (
	"context"
	"path/filepath"

	"github.com/urfave/cli/v3"
	"github.com/vugu/vugu/v2/gen"
)

var (
	Recursive bool
)

// Generates the code from the "*.vugu" and "*.htnml" files.
// This function is a direct copy of cmd/vugugen/vugugen.go, no
// functionality has been changed.
func Gen(ctx context.Context, cmd *cli.Command) error {
	// we need to get the arguments from the command as a slice.
	// The only argument would be the directory to run in.
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
			err = gen.RunRecursive(pkgPath)
		} else {
			err = gen.Run(pkgPath)
		}
		if err != nil {
			return err
		}

	}
	return nil
}
