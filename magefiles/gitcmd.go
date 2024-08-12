//go:build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/sh"
)

func gitDiffFiles() error {
	err := gitCmdV("diff", "--quiet") // --quiet surpresses all output - see "git hulp diff" so we don't need gitCmdCombinedOutput here
	if err != nil {
		// the diff failed so return the output of "git status" to indicate which files have changed.
		// The contributor can then determin the actual change
		out, statusErr := gitCmdCaptureOutput("status")
		if statusErr != nil {
			return fmt.Errorf("git diff --quiet failed with: %s\ngit status failed with: %s\nOutput follows:\n%s\n", err, statusErr, out)
		}
		return fmt.Errorf("The git repo is dirty. Where these files committed to the repository after `vugugen` and `go mod tidy` had been run?\ngit status output:\n%s\n", out)
	}
	return err
}

func gitCmdV(args ...string) error {
	return sh.RunV("git", args...)
}

func gitCmdCaptureOutput(args ...string) (string, error) {
	return sh.Output("git", args...)
}
