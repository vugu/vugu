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

func matchWasmFilename(fn string) (bool, error) {
	return regexp.MatchString("^.*[.]wasm$", fn)
}

func matchGeneratedFilename(fn string) (bool, error) {
	return regexp.MatchString("^.*_gen[.]go$", fn)
}

func deleteGeneratedWasmFiles(path string, d fs.DirEntry, err error) error {
	// skip all directories
	if !d.IsDir() {
		match, err := matchWasmFilename(d.Name())
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

func deleteGeneratedFiles(path string, d fs.DirEntry, err error) error {
	// skip all directories
	if !d.IsDir() {
		match, err := matchGeneratedFilename(d.Name())
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

func generatedFilesCheck(dir string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	err = os.Chdir(dir)
	if err != nil {
		return err
	}
	// run vugugen to regenerated the generated files. This will overwrite any existing files
	err = sh.RunV("vugugen")
	if err != nil {
		return err
	}
	// now run go mod tidy
	err = runGoModTidyInCurrentDir()
	if err != nil {
		return err
	}
	// now check the generated files are identical
	err = gitDiffFiles()
	if err != nil {
		return err
	}

	return err
}
