package main

import (
	"fmt"
	"log"

	"github.com/magefile/mage/sh"
)

func runGitDiffFilesInTestDirs() error {
	return runFuncInDir(WasmTestSuiteDir, gitDiffFiles)
}

func runGitDiffFilesForTest(moduleName string) error {
	return runFuncInModuleDir(moduleName, WasmTestSuiteDir, gitDiffFiles)
}

func runGitDiffFilesInExampleDirs() error {
	return runFuncInDir(ExamplesDir, gitDiffFiles)
}

func runGitDiffFilesForExample(moduleName string) error {
	return runFuncInModuleDir(moduleName, ExamplesDir, gitDiffFiles)
}

func gitDiffFiles() error {
	out, err := gitCmdCaptureOutput("diff", "--quiet")
	if err != nil {
		return fmt.Errorf("git diff --quiet fails with %q\n%q\n", err, out)
	} else {
		log.Printf("As expected err is nil. No differences found\n")
	}
	return err
}

func gitCmdV(args ...string) error {
	return sh.RunV("git", args...)
}

func gitCmdCaptureOutput(args ...string) (string, error) {
	return sh.Output("git", args...)
}
