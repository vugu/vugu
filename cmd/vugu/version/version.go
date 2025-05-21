package version

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/urfave/cli/v3"
)

const (
	prefix      = "vugu version "
	dirtySuffix = "-dirty"
)

var FailedToReadBuildInfoError = errors.New("version: failed ot read the binaries build information")

var version string // this value should be set by Mage during the build process. If not it is a local build.

// Print the vugu version information to STDOUT. This is the implementation of the `vugu version` sub command.
// The version number written to STDOUT has one of two forms.
//
// If the `vugu“ command was built
// via the mage build script then the format is the same as the output of a [git describe] comamnd,
// when passed the "--dirty" option and prefixed with "vugu version". For example:
//
// vugu version v0.4.0-139-g3fe0108-dirty
//
// Where the format is <last-tag>-<number-of-commits-beyond-tag>-<short-hash-of-last-commit>[-dirty]
//
// So in the above case the tag is "v0.4.0", the last commit is 130 commits after that tag, and
// the last commit was "g3fe0108". The "-dirty" indicates that the repo has chnages yet
// to be commited.
//
// If the `vugu` command was build directly via a "go install" or a "go build" outside of the
// magefile then the version format only prints the hash of the last commit
// and the optional "-dirty" if the repo had uncommited changes. For example:
//
// vugu version 3fe01084fba260c70b6cf7862100471fbb01e834-dirty
//
// Where fe01084fba260c70b6cf7862100471fbb01e834 is the hash of the last commit.
//
// [git describe]: https://git-scm.com/docs/git-describe
func Version(ctx context.Context, cmd *cli.Command) error {
	if versionNotSet() {
		// version was not set at build time so pull the rcs inform rom the binary to get the commit hash
		var (
			rev      = "unknown"
			modified bool
		)
		// read the build info becuase the compile time variable has not been set
		// so we assume this binary was built outside of the magefile
		buildInfo, ok := debug.ReadBuildInfo()
		if !ok {
			return FailedToReadBuildInfoError
		}
		for _, v := range buildInfo.Settings {
			if v.Key == "vcs.revision" {
				rev = v.Value // we will use this as the version number if the it was not set at build time
			}
			if v.Key == "vcs.modified" {
				if v.Value == "true" {
					modified = true
				}
			}
		}
		if modified {
			rev = rev + dirtySuffix
		}
		fmt.Println(prefix + rev)
		// we have a version number set at built tiem but was the repo dirty?
	} else {
		// the version has been set at build tiem by mage so just print it
		fmt.Println(prefix + version)
	}
	return nil
}

func versionNotSet() bool {
	return version == ""
}
