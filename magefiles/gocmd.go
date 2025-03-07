//go:build mage

package main

import (
	"strings"

	"github.com/magefile/mage/sh"
)

func goInstall(pkgName string) error {
	return goCmdV("install", pkgName)
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

// Get the GOROOT for the standard go compiler via go env
// we need a tinygo version of this!
func goGetGoRoot() (string, error) {
	// TODO - ref Issue #344 - https://github.com/vugu/vugu/issues/344
	// What does tinygo do???? The GOROOT has to point to a different location so we need to take this into account.
	return goCmdCaptureOutput("env", GoRoot)
}

func buildLegacyWasmTestSuiteServer() error {
	envs := map[string]string{
		"CGO_ENABLED": "0",
		"GOOS":        "linux",
		"GOARCH":      "amd64",
	}
	return goCmdWithV(envs, "build", "-o", "./wasm-test-suite-srv", "github.com/vugu/vugu/legacy-wasm-test-suite/docker")

}

func testLegacyWasmTestSuite() error {
	return goCmdV("test", "-v", "-timeout", "35m", "github.com/vugu/vugu/legacy-wasm-test-suite")
}

func runGoModTidyInCurrentDir() error {
	return goCmdV("mod", "tidy") // run in current dir
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
