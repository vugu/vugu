//go:build mage

package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
)

func buildAndTestModule(module moduleData, withGeneratedFileCheck bool) error {
	return doBuildAndTestModule(module, true, withGeneratedFileCheck)
}

func buildModule(module moduleData, withGeneratedFileCheck bool) error {
	return doBuildAndTestModule(module, false, withGeneratedFileCheck)
}

func upgradeModuleDependencies(dir string) error {
	modules, err := modulesUnderDir(dir)
	if err != nil {
		return err
	}
	// loop ove reach module dir
	for _, module := range modules {
		// build the wasm binary
		err = updateModuleDependencies(module)
		if err != nil {
			return err
		}
	}
	// cleanup via the deferred functions
	return err // must be nil at this point
}

func doBuildAndTestModule(module moduleData, withTests bool, withGeneratedFileCheck bool) error {
	err := runVugugenInModuleDir(module)
	if err != nil {
		return err
	}

	err = runGoModTidyInModuleDir(module)
	if err != nil {
		return err
	}

	if withGeneratedFileCheck {
		// sanity check that we have what we expect
		err = runGitDiffFilesInModuleDir(module)
		if err != nil {
			return err
		}
	}

	err = runGoBuildInModuleDir(module)
	if err != nil {
		return err
	}

	if withTests {
		err = runGoTestInModuleDir(module)
		if err != nil {
			return err
		}
	}
	return err
}

func runVugugenInModuleDir(module moduleData) error {
	f := func() error {
		return sh.RunV("vugugen") // run in src dir
	}
	return runFuncIn(module, f)
}
func runGoModTidyInModuleDir(module moduleData) error {
	f := func() error {
		return goCmdV("mod", "tidy") // run in current dir
	}

	return runFuncIn(module, f)
}

func runGoBuildInModuleDir(module moduleData) error {
	f := func() error {
		envs := map[string]string{
			"GOOS":   "js",
			"GOARCH": "wasm",
		}
		return goCmdWithV(envs, "build", "-o", "./main.wasm", module.name)
	}
	return runFuncIn(module, f)
}

func runGoTestInModuleDir(module moduleData) error {
	f := func() error {
		return goCmdV("test", "-v", module.name)
	}
	return runFuncIn(module, f)
}

func runGitDiffFilesInModuleDir(module moduleData) error {
	return runFuncIn(module, gitDiffFiles)
}

func updateModuleDependencies(module moduleData) error {
	f := func() error {
		return goCmdV("get", "-u", "-t", module.name)
	}
	return runFuncIn(module, f)
}

func runFuncIn(module moduleData, f func() error) error {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	// remove symlinks - if any - necessary on MacOs. The incoming dir is absolute
	dir, err := filepath.EvalSymlinks(module.dir)
	if err != nil {
		return err
	}
	err = os.Chdir(dir) // does dir have symlinks?
	if err != nil {
		return err
	}
	return f()

}

func resolveDirToAbs(rootDir, dir string) (abdDir string, err error) {
	// resolve the dir to an absolute, not symlink path. On MacOS we need to resolve the symlinks....
	dir, err = filepath.Abs(filepath.Join(rootDir, dir))
	if err != nil {
		return "", err
	}
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		return "", err
	}
	return dir, err
}

type moduleData struct {
	name string
	dir  string
}

func modulesUnderDir(dir string) ([]moduleData, error) {
	moduleDataMap := make(map[string]string)
	moduledata := make([]moduleData, 0, 27) // 27 is the current number of wasm tests
	vuguModuleRoot, err := os.Getwd()       //  cwd should be the module root
	if err != nil {
		return nil, err
	}
	defer func() {
		// ensure we finish in the moduleRoot no matter what.
		_ = os.Chdir(vuguModuleRoot)
	}()

	// resolve to abd without symlinks - we need the later to work safely on MacOS
	if dir, err = resolveDirToAbs(vuguModuleRoot, dir); err != nil {
		return nil, err
	}
	// Change dir into the requested directory. We expect this directory to contian only sub-dirs which contian the code we want to work with
	// if the dir is not at the module root then this will fail indicating that the requested directory does nto exist.
	err = os.Chdir(dir)
	if err != nil {
		return nil, err
	}

	// find all the directories under dir - dir is absolute at this point
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		// skip any files
		if !e.IsDir() {
			continue
		}
		if strings.Contains(e.Name(), MagefilesDir) {
			continue // ignore the magefiles directory
		}
		subdir, err := resolveDirToAbs(dir, e.Name())
		if err != nil {
			return nil, err
		}
		// CDW into the absolute subdir
		err = os.Chdir(subdir)
		if err != nil {
			return nil, nil
		}
		// seems we need to call go list -m twice - once to get the module name (i.e. Path) and once to get the source code dir
		moduleName, err := goCmdCaptureOutput("list", "-m", "-f", "{{.Path}}") // we only want to module path i.e. no replacement info required. This returns a single line
		if err != nil {
			return nil, err
		}
		moduleDir, err := goCmdCaptureOutput("list", "-m", "-f", "{{.Dir}}") // we only want to module directory where the source code is located. This returns a single line
		if err != nil {
			return nil, err
		}

		moduleName = strings.TrimSpace(moduleName)
		moduleDir = strings.TrimSpace(moduleDir)
		// not every subdir of dir contains a module - i.e. we can have just package directories or hidden directories
		// so we will have duplicate moduleName and moduleDir for these entries - as they will always refer to the
		// module that contains the subdir.
		// We use a map to deduplicate the module entries
		moduleDataMap[moduleName] = moduleDir
	}
	// now we hae a map, convert it to a slice of moduleData, as that's what the rest of the code expects
	for k, v := range moduleDataMap {
		moduledata = append(moduledata, moduleData{name: k, dir: v})
		if err != nil {
			return nil, err
		}
	}
	return moduledata, nil
}
