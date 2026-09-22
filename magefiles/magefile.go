//go:build mage

// This is the magefile used to build and test the 'vugu' project.
// Execute 'mage -l' for a list of the available targets.
// This magefile requires that 'docker' and 'git' are installed.
// For 'docker' install instructions see: https://docs.docker.com/engine/install/
// For 'git' install instructions see: https://git-scm.com/downloads
package main

import (
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// All builds, lints and tests vugu and the vugu examples.
// This includes running both the the 'vugu' tests and the full 'wasm-test-suite' set of test cases.
// The 'wasm-test-suite' will be built with the standard Go compiler.
// All does NOT run the legacy test suite bu default.
// See the AllWithLegacyWasm or TestLegecyWasm targets if you need to run the legacy wasm test suite.
func All() error {
	mg.SerialDeps(
		Build,
		Lint, // we can't run golangci-lint in parallel with the Build as that calls "go generate"
		Test,
		TestVuguInit,
		TestWasm,
		Examples,
		StopLocalNginxForExamples,
	)
	return nil
}

// Like the "All" target but does NOT run the linter, as the Dockerized 'golangci-lint' can take 40+ seconds to run
func AllNoLint() error {
	mg.SerialDeps(
		Build,
		Test,
		TestVuguInit,
		TestWasm,
		Examples,
		StopLocalNginxForExamples,
	)
	return nil
}

// Like 'All' but also runs the legacy wasm test suite
func AllWithLegacyWasm() error {
	mg.SerialDeps(
		Build,
		Lint,
		Test,
		TestVuguInit,
		TestWasm,
		Examples,
		StopLocalNginxForExamples,
		TestLegacyWasm,
	)
	return nil
}

// Like 'AllWithLegacy' but additionally confirm that the generated files that should have been committed are correct. Intended to be used by the GitHub Action
func AllGitHubAction() error {
	mg.SerialDeps(
		BuildWithGeneratedFilesCheck,
		// temporally turn off the linter.
		// golang-lint version 2.0.2 has the errcheck and static check tool enabled.
		// These are now generating a large number of build failure in the
		// cmd/vugufmt
		// devutil
		// distutil
		// domrender
		// gen
		// legacy-wasm-test-suite
		// simplehttp
		// staticrender
		// internal/htmlx
		// vugufmt
		// packages.
		// Lint,
		Test,
		TestVuguInit,
		TestWasmWithGeneratedFilesCheck,
		Examples,
		StopLocalNginxForExamples,
		// With Go 1.24.x the location of the wasm_exec.js file has changed.
		// We do not wan to update the legacy test suite - we want to eventually remove it - so we have not updated the
		// test code to reflect this change.
		// In addition if we use Go 1.23.x then the latest version of tinygo (v0.36+) wil panic with:
		//    --- FAIL: Test012Router/tinygo/Docker (20.85s)
		//		panic: TinygoCompiler: build error: exit status 1; cmd.args: [docker run --rm -v /home/owen/go/pkg/mod:/gomodcache -e GOMODCACHE=/gomodcache -v /home/owen/.cache/tinygo:/tinygocache -e GOCACHE=/tinygocache -v /:/src -v /tmp/:/out -e HOME=/tmp --user=1000 -w /src/home/owen/src/vugu/legacy-wasm-test-suite/test-012-router tinygo/tinygo:latest tinygo build -o /out/WasmCompiler226958135 -target=wasm .], full output:
		//		/usr/local/go/src/crypto/internal/sysrand/rand_js.go:13:22: //go:wasmimport gojs runtime.getRandomData: unsupported parameter type []byte
		//		[recovered]
		//		panic: TinygoCompiler: build error: exit status 1; cmd.args: [docker run --rm -v /home/owen/go/pkg/mod:/gomodcache -e GOMODCACHE=/gomodcache -v /home/owen/.cache/tinygo:/tinygocache -e GOCACHE=/tinygocache -v /:/src -v /tmp/:/out -e HOME=/tmp --user=1000 -w /src/home/owen/src/vugu/legacy-wasm-test-suite/test-012-router tinygo/tinygo:latest tinygo build -o /out/WasmCompiler226958135 -target=wasm .], full output:
		///usr/local/go/src/crypto/internal/sysrand/rand_js.go:13:22: //go:wasmimport gojs runtime.getRandomData: unsupported parameter type []byte
		//
		// Given that we want to remove the legacy test suite this is not worth fixing.
		// However Issues #278 and #279 now urgently need fixed:
		// https://github.com/vugu/vugu/issues/279
		// https://github.com/vugu/vugu/issues/278
		//
		//TestLegacyWasm,
	)
	return nil
}

// Build the 'vugu' package, the 'vugugen' and 'vugufmt' commands
func Build() error {
	mg.SerialDeps(InstallGoImports)
	// install the vugu module by executing
	err := goInstall("github.com/vugu/vugu")
	if err != nil {
		return err
	}
	// install the vugugen command by executing
	err = goInstall("github.com/vugu/vugu/cmd/vugugen")
	if err != nil {
		return err
	}
	// build the vugu cmd, ensuring we set the version number correctly
	version, err := gitDescribe()
	if err != nil {
		return err
	}
	err = goInstallWithSetVariable("github.com/vugu/vugu/cmd/vugu/version.version", version, "github.com/vugu/vugu/cmd/vugu")
	if err != nil {
		return err
	}
	// install the vugufmt command by executing
	err = goInstall("github.com/vugu/vugu/cmd/vugufmt@latest") // vugufmt can use the goimports tool
	if err != nil {
		return err
	}
	// ensure we always build the vgform package - it contains generated code that may nee dto be rebuild
	return buildVgForm("./vgform")
}

// Like Build but additionally confirm that the generated files in the `vgform` package that should have been committed are correct.
func BuildWithGeneratedFilesCheck() error {
	mg.SerialDeps(Clean, Build)
	// sanity check that the generated files have been submitted, and that the go.mod and go.sum are
	// as they should be post calling vugugen
	return generatedFilesCheck("./vgform")
}

// Install the latest vertsion of the goimports tool if a version is not already installed locally.
func InstallGoImports() error {
	mg.SerialDeps(Clean)
	if !checkCmdExists("goimports") {
		log.Printf("No copy of goimports found, installing latest version")
		return goInstall("golang.org/x/tools/cmd/goimports@latest")
	}
	return nil
}

// Pulls the latest version of the 'golangci-lint' docker image. 'The 'vugu' project used 'golangci-lint' as its lint tool. See the Lint target
func PullLatestGolangCiLintDockerImage() error {
	mg.Deps(CheckRequiredCmdsExist)
	return dockerPullImage(GoLangCiLintImageName)
}

// Lint the 'vugu' source code including the tests, and examples and any generated files. The linter used is a dockerized version of the 'golangci-lint' linter.
// See: https://golangci-lint.run/ for more details of the linter
func Lint() error {
	mg.SerialDeps(
		Build,
		PullLatestGolangCiLintDockerImage,
	)
	return runGolangCiLint()
}

// Builds and runs all of the Go unit tests for 'vugu', except the 'legacy-wasm-test-suite' tests using the standard Go compiler.
func Test() error {
	mg.SerialDeps(Build)
	packages, err := goListPackages()
	if err != nil {
		return err
	}
	for _, pkg := range packages {
		if strings.Contains(pkg, LegacyWasmTestSuiteDir) || strings.Contains(pkg, WasmTestSuiteDir) {
			continue // skip the wasm-test-suite package and wasm-test-suite packages
		}
		err = goTest(pkg)
		if err != nil {
			return err
		}
	}
	return nil
}

// Builds and runs all of the 'wasm-test-suite' set of tests (and no other tests). The 'wasm' files are built using the standard Go compiler.
func TestWasm() error {
	return testWasm(false)
}

// Like TestWasm but additionally confirm that the generated files that should have been committed are correct.
func TestWasmWithGeneratedFilesCheck() error {
	return testWasm(true)
}

func testWasm(withGeneratedFilesCheck bool) error {
	mg.SerialDeps(Build, PullLatestNginxImage, PullLatestChromeDpImage)
	// somewhat unusually we add the deferred function first! If createContainers errors we don't care how far it got, as we destoy everything
	// cleanupContainers() is safe - it can't error, if a container isn't running it can't be stopped and there are no consequences of doing this
	// because it cleans up both the containers and the network between them.
	defer cleanupContainers()
	// now if any if the createContainers function errors and we return we will cleanup via the defer
	err := createContainers()
	if err != nil {
		return err
	}

	// find all the modules under dir
	modules, err := modulesUnderDir(WasmTestSuiteDir)
	if err != nil {
		return err
	}

	// loop ove reach module dir
	for _, module := range modules {
		err = buildAndTestModule(module, withGeneratedFilesCheck)
		if err != nil {
			return err
		}
	}
	// cleanup via the deferred functions
	return err // MUST be nil at this point
}

// Builds, and runs a single test from the 'wasm-test-suite'. This is useful if you are trying to debug a particular 'wasm' test case.
// The usage is:
//
//	mage TestSingleWasmTest <test-module-name>
//
// e.g.
//
//	mage TestSingleWasmTest github.com/vugu/vugu/wasm-test-suite/test-012-router
//
// to run the router test case.
// The 'wasm' will be built with the standard Go compiler.
func TestSingleWasmTest(moduleName string) error {
	return testSingleWasmTest(moduleName, false)
}

// Like TestSingleWasm but additionally confirm that the generated files that should have been committed are correct.
func TestSingleWasmTestWithGeneratedFilesCheck(moduleName string) error {
	return testSingleWasmTest(moduleName, true)
}

func testSingleWasmTest(moduleName string, withGeneratedFilesCheck bool) error {
	mg.SerialDeps(Build, PullLatestNginxImage, PullLatestChromeDpImage)
	// somewhat unusually we add the deferred function first! If createContainers errors we don't care how far it got, as we destoy everything
	// cleanupContainers() is safe - it can't error, if a container isn't running it can't be stopped and there are no consequences of doing this
	// because it cleans up both the containers and the network between them.
	defer cleanupContainers()
	// now if any if the createContainers function errors and we return we will cleanup via the defer
	err := createContainers()
	if err != nil {
		return err
	}

	// find all the modules under dir
	allmodules, err := modulesUnderDir(WasmTestSuiteDir)
	if err != nil {
		return err
	}

	// filter the list to find the module we want - there should only be one
	module := moduleData{}
	found := false
	for _, nthModule := range allmodules {
		if nthModule.name == moduleName {
			module = nthModule
			found = true
			break // we have found theone we want
		}
	}
	if !found {
		return fmt.Errorf("Could not find a module named %q", moduleName)
	}
	// no need to loop as there is only one module found be the loop above
	err = buildAndTestModule(module, withGeneratedFilesCheck)
	if err != nil {
		return err
	}

	// cleanup via the deferred functions
	return err // err must be nil at this point
}

// Checks that all required commands have been installed prior to attempting the build.
// The only required command at present is 'docker'.
func CheckRequiredCmdsExist() error {
	err := checkAllCmdsExist(DockerCmdExeName)
	if err != nil {
		return err
	}
	return checkAllCmdsExist(GitCmdExeName) // currently docker is the only required external commands
}

// Removes any existing WASM files by searching for any files whose filename matches "*.wasm" and then deleting them.
func Clean() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	return filepath.WalkDir(cwd, deleteGeneratedWasmFiles)
}

// Removes any existing generated files by searching for any files generated by `vugugen whose filename matches "*_gen.go"and deleting them.
func CleanAutoGeneratedFiles() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	return filepath.WalkDir(cwd, deleteGeneratedFiles)
}

// Removes any existing generated files from the legacy wasm test suite by searching for any files generated by `vugugen whose filename matches "*_gen.go"and deleting them.
// Use this target after calling TestLegacyWasm to cleanup.
// These generated files are intentionally not tracked but git.
func CleanLegacyWasmGeneratedFiles() error {
	return filepath.WalkDir(LegacyWasmTestSuiteDir, deleteGeneratedFiles)
}

// Pulls the latest nginx container image. The 'wasm-test-suite' uses this nginx image to server the wasm tests.
func PullLatestNginxImage() error {
	mg.Deps(CheckRequiredCmdsExist)
	return dockerPullImage(NginxImageName)
}

// Pulls the latest chromedp container image. The 'wasm-test-suite' uses this chromedp image as a headless browser for the wasm tests.
// See: https://github.com/chromedp/chromedp for more information on chromedp.
func PullLatestChromeDpImage() error {
	mg.Deps(CheckRequiredCmdsExist)
	return dockerPullImage(ChromeDpHeadlessShellImageName)
}

// Start a local nginx container to server the 'wasm-test-suite' tests.
// This is useful if you need debug a wasm test case using a browser.
// It is safe to execute this multiple time without calling the "StopLocalNginxForWasmTestSuite" target.
// Any existing nginx container will be stopped before starting a new one.
// To connect to a test the URL is of the form:
//
//	http://localhost:8888/<name-of-test-directory>
//
// e.g.
//
//	http://localhost:8888/test-001-sinple
//
// To see the result of executing the wasm for the test test-001-simple
func StartLocalNginxForWasmTestSuite() error {
	mg.SerialDeps(PullLatestNginxImage)
	stopContainerForWasmTestSuite()
	err := startContainerForWasmTestSuite()
	if err == nil {
		fmt.Printf("Local nginx container started.\nConnect to http://localhost:8888/<wasm-test-directory-name>\ne,g. http://localhost:8888/test-001-simple\n")
		fmt.Printf("To stop the local nginx container please run:\n\tmage StopLocalNginxForWasmTestSuite\n")
	}
	return err
}

// Start a local nginx container to server the 'examples“.
// This is useful if you want to run an example in a browser.
// It is safe to execute this multiple time without calling the "StopLocalNginxForExamples" target.
// Any existing nginx container will be stopped before starting a new one.
// To connect to a test the URL is of the form:
//
//	http://localhost:8889/<example-test-directory-name>
//
// e.g.
//
//	http://localhost:8889/fetch-and-display
//
// To see the result of executing the wasm for the test test-001-simple
func StartLocalNginxForExamples() error {
	mg.SerialDeps(PullLatestNginxImage)
	stopContainerForExamples()
	err := startContainerForExamples()
	if err == nil {
		fmt.Printf("Local nginx container started.\nConnect to http://localhost:8889/<example-test-directory-name>\ne,g. http://localhost:8889/fetch-and-display\n")
		fmt.Printf("To stop the local nginx container please run:\n\tmage StopLocalNginxForExamples\n")
	}
	return err
}

// Stops any locally running nginx container serving from the wasm-test-suite directory.
// See: StartLocalNginxForWasmTestSuite
func StopLocalNginxForWasmTestSuite() error {
	// we intended to ignore the error as the container might not be running
	stopContainerForWasmTestSuite()
	return nil
}

// Stops any locally running nginx container serving from the wasm-test-suite directory.
// See: StartLocalNginxForExamples
func StopLocalNginxForExamples() error {
	// we intended to ignore the error as the container might not be running
	stopContainerForExamples()
	return nil
}

// Pulls the latest version of the 'tinygo' Go compiler docker image. 'tinygo' is an alternative Go compiler used by the 'vugu' to generate the Web Assembly (wasm) files.
// 'vugu' uses it only to ensure tha the 'wasm-test-suite' can be built using 'tinygo' compiler.
// 'tinygo' DOES NOT support the full Go standard library, when compiling to wasm. As such you may not be able to compile your project using 'tinygo'. See: https://tinygo.org/docs/reference/lang-support/stdlib/
func PullLatestTinyGoDockerImage() error {
	mg.Deps(CheckRequiredCmdsExist)
	return dockerPullImage(TinyGoImageName)
}

// Builds and serves all of the examples via a local nginx container.
// The examples will be served at
//
//	http://localhost:8888/<name-of-example-directory>
//
// for example for the fetch and display example
//
//	http://localhost:8888/fetch-and-display
//
// The 'wasm' files are built using the standard Go compiler.
func Examples() error {
	return examples(false)
}

// Like Examples but additionally confirm that the generated files that should have been committed are correct.
func ExamplesWithGeneratedFilesCheck() error {
	return examples(true)
}

func examples(withGeneratedFilesCheck bool) error {
	mg.SerialDeps(Build, PullLatestNginxImage)
	StartLocalNginxForExamples()
	// find all the modules under dir
	modules, err := modulesUnderDir(ExamplesDir)
	if err != nil {
		return err
	}
	// loop ove reach module dir
	for _, module := range modules {
		// build the wasm binary
		err = buildModule(module, withGeneratedFilesCheck)
		if err != nil {
			return err
		}
	}
	// cleanup via the deferred functions
	return err // must be nil at this point
}

// Builds, and serves a single example using a local nginx container.
// The usage is:
//
//	mage SingleExample <example-module-name>
//
// e.g.
//
//	mage SingleExample gtihub.com/vugu/vugu/examples/fetch-and-display
//
// The examples will be served at
//
//	http://localhost:8888/<name-of-example-directory>
//
// for example for the fetch and display example
//
//	http://localhost:8888/fetch-and-display
//
// The 'wasm' files are built using the standard Go compiler.
func SingleExample(moduleName string) error {
	return singleExample(moduleName, false)
}

// Like SingleExample but additionally confirm that the generated files that should have been committed are correct.
func SingleExampleWithGeneratedFilesCheck(moduleName string) error {
	return singleExample(moduleName, true)
}

func singleExample(moduleName string, withGeneratedFilesCheck bool) error {
	mg.SerialDeps(Build, PullLatestNginxImage)
	StartLocalNginxForExamples()
	return buildSingleModule(ExamplesDir, moduleName, withGeneratedFilesCheck)
}

// Builds, and serves a single wasm test using a local nginx container.
// The usage is:
//
//	mage SingleWasmTest <wasm-test-module-name>
//
// e.g.
//
//	mage SingleWasmTest gtihub.com/vugu/vugu/wasm-test-suite/test-002-click
//
// The examples will be served at
//
//	http://localhost:8888/<name-of-wasm-test-directory>
//
// for example for the test-0020-click
//
//	http://localhost:8888/test-0020-click
//
// The 'wasm' files are built using the standard Go compiler.
func SingleWasmTest(moduleName string) error {
	return singleWasmTest(moduleName, false)
}

// Like SingleExample but additionally confirm that the generated files that should have been committed are correct.
func SingleWasmTestWithGeneratedFilesCheck(moduleName string) error {
	return singleWasmTest(moduleName, true)
}

func singleWasmTest(moduleName string, withGeneratedFilesCheck bool) error {
	mg.SerialDeps(Build, PullLatestNginxImage)
	StartLocalNginxForWasmTestSuite()
	return buildSingleModule(WasmTestSuiteDir, moduleName, withGeneratedFilesCheck)
}

func buildSingleModule(dir string, moduleName string, withGeneratedFilesCheck bool) error {
	mg.SerialDeps(Build, PullLatestNginxImage)

	// find all the modules under dir
	allmodules, err := modulesUnderDir(dir)
	if err != nil {
		return err
	}

	// filter the list to find the module we want - there should only be one
	module := moduleData{}
	found := false
	for _, nthModule := range allmodules {
		if nthModule.name == moduleName {
			module = nthModule
			found = true
			break // we have found theone we want
		}
	}
	if !found {
		return fmt.Errorf("Could not find a module named %q", moduleName)
	}

	// build the wasm
	err = buildModule(module, withGeneratedFilesCheck)
	if err != nil {
		return err
	}
	// cleanup via the deferred functions
	return err // err must be nil at this poitnt
}

// Run the 'legacy-wasm-test-suite' building the tests cases using both the standard Go compiler and the 'tinygo' compiler.
// The tests are served via a docker container image that contains a local development web server and a headless chrome browser to
// fetch and execute the wasm.
// This test suite can take 35 minutes to execute.
func TestLegacyWasm() error {
	mg.SerialDeps(Build,
		PullLatestTinyGoDockerImage,
	)
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	err = os.Chdir(LegacyWasmTestSuiteDir + "/docker")
	if err != nil {
		return err
	}
	// build the wasm-test-suite-srv
	err = buildLegacyWasmTestSuiteServer()
	if err != nil {
		return err
	}
	err = buildLegacyWasmTestSuiteContainer()
	if err != nil {
		return err
	}
	// We intend to ignore any errors here, as the container might not be running
	stopLegacyWasmTestSuiteContainer()
	err = startLegacyWasmTestSuiteContainer()
	if err != nil {
		return err
	}
	// stop the running container on function exit
	// We intend to ignore any errors here, as the container might not be running
	defer stopLegacyWasmTestSuiteContainer()
	return testLegacyWasmTestSuite()
}

// Update all of the dependencies by running "go get -u -t <module-name> for every module - the root vugu module, the wasm-test-suite modules and the example modules"
func UpgradeAllDependencies() error {
	mg.SerialDeps(UpgradeRootDependencies, UpgradeWasmTestSuiteDependencies, UpgradeExampleDependencies)
	modules, err := modulesUnderDir("./") // we want to run this in the vugu module root - path MUST tbe relative ofr moduleUnderDir to resolve it correctly
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

// Like UpgradeAllDependencies but only for the root vugu module
func UpgradeRootDependencies() error {
	mg.SerialDeps(Build)
	return upgradeModuleDependencies("./") // we want to run this in the vugu module root - path MUST tbe relative ofr moduleUnderDir to resolve it correctly
}

// Like UpgradeAllDependencies but only for the wasm test suite modules
func UpgradeWasmTestSuiteDependencies() error {
	return upgradeModuleDependencies(WasmTestSuiteDir)
}

// Like UpgradeAllDependencies but only for the examples modules
func UpgradeExampleDependencies() error {
	return upgradeModuleDependencies(ExamplesDir)
}

type vuguGenOpts struct {
	// The name of the wasm binary
	WasmBinaryName string
	// the name of the Go source file that containts the main function
	WasmGoFilename string
	// The name of the base/root type. This struct is the man entry point for the wasm code. This must be a valid, exported type in the RootStructPkg
	RootStructType string
}

// Test 'vugu init' by calling `vugu gen` to ensure the generated Go code is valid
func TestVuguInit() error {
	mg.SerialDeps(Build)
	opts := vuguGenOpts{}
	err := TestVuguInitWithOpts(opts)
	if err != nil {
		return fmt.Errorf("vugu gen with default options failed with %w", err)
	}

	// set a Root struct name
	opts.RootStructType = "Banana"
	err = TestVuguInitWithOpts(opts)
	if err != nil {
		return fmt.Errorf("vugu gen with RootSturctType set to %q failed with %w", opts.RootStructType, err)
	}
	// set a binary name in addition
	opts.WasmBinaryName = "goldfish.wasm"
	err = TestVuguInitWithOpts(opts)
	if err != nil {
		return fmt.Errorf("vugu gen with WasmBinaryName set to %q failed with %w", opts.WasmBinaryName, err)
	}
	//now set the main go filename
	opts.WasmGoFilename = "cherry.go"
	err = TestVuguInitWithOpts(opts)
	if err != nil {
		return fmt.Errorf("vugu gen with WasmGoFileName set to %q failed with %w", opts.WasmGoFilename, err)
	}
	// no errors at this point
	return nil
}

func TestVuguInitWithOpts(opts vuguGenOpts) error {
	mg.SerialDeps(Build)
	pkgName := "example.com/vugu/gentest" // this is fixed for th each tess, as its just the name of the module
	rootFileName := ""
	wasmGoFilenameDefault := "main_wasm.go"
	foundMap := make(map[string]bool, 5)

	// get the cwd
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(cwd) // ignore the error and ensure we go back to the cwd on exit

	// create a temp dir
	d, err := os.MkdirTemp("", "-vugu-gen-test-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(d) // we ignore any error but delete the temp dir and contents
	err = os.Chdir(d)

	// build the args list
	args := make([]string, 0)

	// "init" comes first
	args = append(args, "init")

	// then "--wasmbinaryname" is supplied
	if opts.WasmBinaryName != "" {
		args = append(args, "--wasmbinaryname", opts.WasmBinaryName)
	} else {
		opts.WasmBinaryName = "main.wasm" // match the default of the vugu init command
	}
	foundMap[opts.WasmBinaryName] = false

	// then "--wasmgofilename" if supplied
	if opts.WasmGoFilename != "" {
		args = append(args, "--wasmgofilename", opts.WasmGoFilename)
	} else {
		opts.WasmGoFilename = wasmGoFilenameDefault // match the default of the vugu init command
	}
	foundMap[opts.WasmBinaryName] = false

	// then "--rootstructtype" if supplied
	if opts.RootStructType != "" {
		args = append(args, "--rootstructtype", opts.RootStructType)
	} else {
		opts.RootStructType = "Root"
	}
	rootFileName = opts.RootStructType + ".go"
	foundMap[rootFileName] = false

	// the pkgname (really the module name) is always the last argument
	args = append(args, pkgName)

	// now add the missing files ot the found map - these are both derived from the RootStructType name
	genFileName := opts.RootStructType + "_gen.go"
	rootVuguFileName := opts.RootStructType + ".vugu"
	foundMap[genFileName] = false
	foundMap[rootVuguFileName] = false

	// run vugu init with arguments
	err = sh.RunV("vugu", args...)
	if err != nil {
		return err
	}
	// run vugu gen
	// In v0.Y.Z this is a bit of a mess. Basically, the `vugu init` cmd above will generate a main funciton in a file
	// called either "main_wasm.go" or the filename that was specified in opts.WasmGoFilename
	// But, 'vugu gen` will be default generate a file called main_wasm.go unless the "--skip-main" option
	// is specified. So we need to check for this condition and pass the "--skip-main" as appropiate
	if opts.WasmGoFilename != wasmGoFilenameDefault {
		err = runVugugenInCurrentDirSkipMain()
		if err != nil {
			return err
		}
	} else {
		err = runVugugenInCurrentDir()
		if err != nil {
			return err
		}
	}
	// run go mod tidy
	err = runGoModTidyInCurrentDir()
	if err != nil {
		return err
	}
	// now build
	envs := map[string]string{
		"GOOS":   "js",
		"GOARCH": "wasm",
	}
	err = goBuildWithEnvs(envs, opts.WasmBinaryName, pkgName)
	if err != nil {
		return err
	}
	// as a sanity check list the files
	files, err := os.ReadDir(d)
	if err != nil {
		return err
	}
	for _, file := range files {
		switch file.Name() {
		case opts.WasmBinaryName:
			foundMap[opts.WasmBinaryName] = true
		case genFileName:
			foundMap[genFileName] = true
		case rootFileName:
			foundMap[rootFileName] = true
		case rootVuguFileName:
			foundMap[rootVuguFileName] = true
		case opts.WasmGoFilename:
			foundMap[opts.WasmGoFilename] = true
		}
	}
	// check we found everythig
	for k := range maps.Keys(foundMap) {
		if foundMap[k] == false {
			return fmt.Errorf("vugu init test: failed to find the expected file %s", k)
		}
	}
	return err // will be nil at this point
}
