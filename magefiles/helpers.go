//go:build mage

package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"regexp"

	"github.com/magefile/mage/sh"
)

//------- utils ------------------------

func checkAllCmdsExist(cmds ...string) error {
	missingCmds := make([]string, 0)
	for _, cmd := range cmds {
		if !checkCmdExists(cmd) {
			missingCmds = append(missingCmds, cmd)
		}
	}
	if len(missingCmds) != 0 {
		return fmt.Errorf("failed to find commands: %q", missingCmds)
	}
	return nil
}

func checkCmdExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func matchFilename(fn string) (bool, error) {
	match, err := regexp.MatchString("^.*_gen[.]go$", fn)
	// did we fail
	if err != nil {
		return match, err // error in the regex - so match is false
	}
	// did we match a file called "*_gen.go"?
	// if we did we bail early
	if match {
		return match, err
	}
	// nope we didn't match "*_gen.go" do we match "*.wasm"
	// return if we matched or not and any error
	return regexp.MatchString("^.*[.]wasm$", fn)
}

func deleteGeneratedFiles(path string, d fs.DirEntry, err error) error {
	// skip all directories
	if !d.IsDir() {
		match, err := matchFilename(d.Name())
		if err != nil {
			return err
		}
		if match { // name contains only the base name i.e. the last part of the path inc the extension
			// is this a dry run?
			dryrun := os.Getenv("VUGU_MAGE_DRY_RUN")
			if dryrun != "" {
				// dry run set DO NOT DELETE
				log.Printf("VUGU_MAGE_DRY_RUN set: Would have removed: %s\n", path)
				return nil
			}
			// else remove the file - the full name is in path
			return os.Remove(path)
		}
	}
	return nil
}

func goCmdV(args ...string) error {
	return sh.RunV("go", args...)
}

func goCmdCaptureOutput(args ...string) (string, error) {
	return sh.Output("go", args...)
}

func goCmdWithV(env map[string]string, args ...string) error {
	return sh.RunWithV(env, "go", args...)
}

func dockerCmdV(args ...string) error {
	return sh.RunV("docker", args...)
}
