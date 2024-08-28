//go:build mage

package main

import (
	"fmt"
	"log"
	"os"

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
		cwd, _ := os.Getwd()
		diff, _ := gitCmdCaptureOutput("diff")
		return fmt.Errorf("The git repo is dirty. Where these files committed to the repository after `vugugen` and `go mod tidy` had been run?\ngit status output from dir: %s:\n%s\ndiff:\n%s\n", cwd, out, diff)
	}
	return err
}

func gitDiffFile(f string) error {
	err := gitCmdV("diff", "--quiet", f) // --quiet surpresses all output - see "git hulp diff" so we don't need gitCmdCombinedOutput here
	if err != nil {
		// the diff failed so run it again but capture the diff - in this case the return is always nil
		out, diffErr := gitCmdCaptureOutput("diff", f)
		if diffErr != nil {
			return fmt.Errorf("git diff %s failed with: %s\n", f, err)
		}
		// We have a difference between the GOROOT copy of wasm_exec.js and the local copy in teh repo.
		// Ee have no idea if the local file is compatabel with the version of the Go compile rin use - and have no way to determine this
		// until the wasm_exec.js is contains a version string that can be matched against the compiler.
		// In order to ensure a successful run we MUST take the version that matched thee compile rin use i.e. the GOROOT copy will overright the repo copy (if any)
		// In tis case we will emit a warning to indicates that the wasm_exec.js does not match the compiler verison in use.
		log.Printf("WARNING There is a difference between the %s copy of %s and the local copy at %s", GoRoot, WasmExecJS, f)
		log.Printf("WARNING The build process cannot tell which version of Go that the local file relates to because the Go team do not embed a version into %s\n", WasmExecJS)
		log.Printf("WARNING Copying the %S version of %s as this file must match the compiler version in use.\n", GoRoot, WasmExecJS)
		log.Printf("WARNING The file difference was:\n%s\n", out)
		return nil
	}
	return err
}

func gitCmdV(args ...string) error {
	return sh.RunV("git", args...)
}

func gitCmdCaptureOutput(args ...string) (string, error) {
	return sh.Output("git", args...)
}
