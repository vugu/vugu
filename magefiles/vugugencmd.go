package main

import "github.com/magefile/mage/sh"

func runVugugenInTestDirs() error {
	f := func() error {
		return sh.RunV("vugugen") // run in src dir
	}
	return runFuncInDir(WasmTestSuiteDir, f)

}

func runVugugenForTest(moduleName string) error {
	f := func() error {
		// cd into the directory and run vugen to generate the glue source code
		return sh.RunV("vugugen") // run in src dir
	}
	return runFuncInModuleDir(moduleName, WasmTestSuiteDir, f)

}

func runVugugenInExampleDirs() error {
	f := func() error {
		return sh.RunV("vugugen") // run in src dir
	}
	return runFuncInDir(ExamplesDir, f)

}

func runVugugenForExample(moduleName string) error {
	f := func() error {
		// cd into the directory and run go generate followed by go mod tidy
		return sh.RunV("vugugen") // run in src dir
	}
	return runFuncInModuleDir(moduleName, ExamplesDir, f)

}
