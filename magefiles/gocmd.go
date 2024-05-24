package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
)

func goInstall(pkgName string) error {
	return goCmdV("install", pkgName)
}

func goGenerate(pkgName string) error {
	return goCmdV("generate", pkgName)
}

func goTest(pkgName string) error {
	return goCmdV("test", "-v", pkgName)
}

func goListOutput(pkgName string) (string, error) {
	return goCmdCaptureOutput("list", pkgName)
}

func goListPackages() ([]string, error) {
	output, err := goCmdCaptureOutput("list", "./...")
	if err != nil {
		return nil, err
	}
	packages := strings.Split(output, "\n")
	return packages, nil
}

func goBuildForLinuxAMD64(binaryName, pkgName string) error {
	envs := map[string]string{
		"CGO_ENABLED": "0",
		"GOOS":        "linux",
		"GOARCH":      "amd64",
	}
	return goBuildWithEnvs(envs, binaryName, pkgName)
}

func goBuildWithEnvs(envs map[string]string, binaryName, pkgName string) error {
	return goCmdWithV(envs, "build", "-o", binaryName, pkgName)

}

func buildWasmTestSuiteServer() error {
	envs := map[string]string{
		"CGO_ENABLED": "0",
		"GOOS":        "linux",
		"GOARCH":      "amd64",
	}
	return goCmdWithV(envs, "build", "-o", "./wasm-test-suite-srv", "github.com/vugu/vugu/wasm-test-suite/docker")

}

func testLegacyWasmtestSuite() error {
	return goCmdV("test", "-v", "-timeout", "35m", "github.com/vugu/vugu/wasm-test-suite")
}

func runGoGenerateInTestDirs() error {
	f := func() error {
		return goCmdV("generate") // run in src dir
	}
	return runFuncInDir(WasmTestSuiteDir, f)
}

func runGoModTidyInTestDirs() error {
	f := func() error {
		return goCmdV("mod", "tidy") // run in src dir
	}
	return runFuncInDir(WasmTestSuiteDir, f)
}

func runGoBuildInTestDirs() error {
	f := func() error {
		envs := map[string]string{
			"GOOS":   "js",
			"GOARCH": "wasm",
		}

		// the test is defined as a module, so we need to use go list -m to find the module name
		// go list will read and parse the local go.mod for us
		module, err := goCmdCaptureOutput("list", "-m") // -m will print the module path
		if err != nil {
			return err
		}
		return goCmdWithV(envs, "build", "-o", "./main.wasm", module)
	}
	return runFuncInDir(WasmTestSuiteDir, f)
}

func runGoTestInTestDirs() error {
	f := func() error {
		// finally run the actual test - this uses the standard Go compiler

		// the test is defined as a module, so we need to use go list -m to find the module name
		// go list will read and parse the local go.mod for us
		module, err := goCmdCaptureOutput("list", "-m") // -m will print the module path
		if err != nil {
			return err
		}
		return goCmdV("test", "-v", module)
	}
	return runFuncInDir(WasmTestSuiteDir, f)
}

func runGoGenerateForTest(moduleName string) error {
	f := func() error {
		// cd into the directory and run go generate followed by go mod tidy
		return goCmdV("generate") // run in src dir
	}
	return runFuncInModuleDir(moduleName, WasmTestSuiteDir, f)
}

func runGoModTidyForTest(moduleName string) error {
	f := func() error {
		// cd into the directory and run go generate followed by go mod tidy
		return goCmdV("mod", "tidy") // run in src dir
	}
	return runFuncInModuleDir(moduleName, WasmTestSuiteDir, f)
}

func runGoBuildForTest(moduleName string) error {
	f := func() error {
		envs := map[string]string{
			"GOOS":   "js",
			"GOARCH": "wasm",
		}

		// the test is defined as a module, so we need to use go list -m to find the module name
		// go list will read and parse the local go.mod for us
		module, err := goCmdCaptureOutput("list", "-m") // -m will print the module path
		if err != nil {
			return err
		}
		return goCmdWithV(envs, "build", "-o", "./main.wasm", module)
	}
	return runFuncInModuleDir(moduleName, WasmTestSuiteDir, f)
}

func runGoTestForTest(moduleName string) error {
	f := func() error {
		// finally run the actual test - this uses the standard Go compiler

		// the test is defined as a module, so we need to use go list -m to find the module name
		// go list will read and parse the local go.mod for us
		module, err := goCmdCaptureOutput("list", "-m") // -m will print the module path
		if err != nil {
			return err
		}
		return goCmdV("test", "-v", module)
	}
	return runFuncInModuleDir(moduleName, WasmTestSuiteDir, f)
}

func runFuncInDir(dir string, f func() error) error {
	cwd, err := os.Getwd() //  cwd should be the module root
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()

	// resolve the dir to an absolute, not symlink path. On MacOS we need to resolve the symlinks....
	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		return err
	}
	// Change dir into the new test directory - currently called "new-tests"
	err = os.Chdir(dir)
	if err != nil {
		return err
	}

	// find all the directories under "new-tests"
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		// skip any files
		if !e.IsDir() {
			continue
		}
		cwd, err := os.Getwd()
		if err != nil {
			return nil
		}
		err = os.Chdir(e.Name())
		if err != nil {
			_ = os.Chdir(cwd) // can't defer this as defer is function scope when we want loop scope
			return nil
		}
		newcwd, _ := os.Getwd()
		log.Printf("Running command in: %s", newcwd)
		err = f()
		if err != nil {
			_ = os.Chdir(cwd)
			return err
		}
		_ = os.Chdir(cwd)
	}
	return nil
}

func runFuncInModuleDir(moduleName, dir string, f func() error) error {
	cwd, err := os.Getwd() //  cwd should be the module root
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()

	// resolve the dir to an absolute, not symlink path. On MacOS we need to resolve the symlinks....
	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		return err
	}
	// Change dir into the new test directory - currently called "new-tests"
	err = os.Chdir(dir)
	if err != nil {
		return err
	}

	// find all the directories under "new-tests"
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	moduleFound := 0
	for _, e := range entries {
		// skip any files
		if !e.IsDir() {
			continue
		}
		cwd, err := os.Getwd()
		if err != nil {
			return nil
		}
		err = os.Chdir(e.Name())
		if err != nil {
			_ = os.Chdir(cwd) // can't defer this as defer is function scope when we want loop scope
			return nil
		}
		newcwd, _ := os.Getwd()
		log.Printf("Running command in: %s", newcwd)

		// is this the module we want?
		output, err := goCmdCaptureOutput("list", "-m")
		if err != nil {
			return err
		}
		modules := strings.Split(output, "\n")

		if len(modules) != 1 {
			return fmt.Errorf("There should be exactly one module in the directory %s but have found %d.", e.Name(), len(modules))
		}
		// is this the module we want?
		if moduleName != modules[0] {
			log.Printf("Skipping module %q", modules[0])
			//  cd out of the directory back to the
			_ = os.Chdir(cwd) // can't defer this as defer is function scope when we want loop scope
			continue
		} else {
			log.Printf("Found module %q", modules[0])
		}
		moduleFound++
		// we have the correct module do do the work in the modules directory.
		err = f()
		if err != nil {
			fmt.Println("5")
			_ = os.Chdir(cwd)
			return err
		}
		_ = os.Chdir(cwd)
	}
	if moduleFound != 1 {
		return fmt.Errorf("Could nto find module %q", moduleName)
	}
	return nil
}

// func goBuildForWasm(outputWasmFileName string, pkgName string) error {
// 	return
// }

func goCmdV(args ...string) error {
	return sh.RunV("go", args...)
}

func goCmdCaptureOutput(args ...string) (string, error) {
	return sh.Output("go", args...)
}

func goCmdWithV(env map[string]string, args ...string) error {
	return sh.RunWithV(env, "go", args...)
}

func logGoCmd(args ...string) {
	var sb strings.Builder
	sb.WriteString("go ")
	for _, s := range args {
		sb.WriteString(s)
		sb.WriteString(" ")
	}
	log.Printf("Executing: %q", strings.TrimSpace(sb.String()))
}
