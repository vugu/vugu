//go:build mage

// This is the magefile used to build and test the 'vugu' project.
// Execute 'mage -l' for a list of the available targets.
// This magefile requires that 'docker' is installed.
// For 'docker' install instructions see:https://docs.docker.com/engine/install/
package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
)

// All builds, lints and tests vugu.
// This includes running both the the 'vugu' tests and the full 'wasm-test-suite' set of test cases.
// The 'wasm-test-suite' will be built with the standard Go compiler.
// All does NOT run the legacy test suite bu default.
// See the AllWithLegacyWasm or TestLegecyWasm targets if you need to run the legacy wasm test suite.
func All() error {
	mg.SerialDeps(
		Build,
		Lint, // we can't run golangci-lint in parallel with the Build as that calls "go generate"
		Test,
		TestWasm,
	)
	return nil
}

// Like the "All" target but does NOT run the linter, as the Dockerized 'golangci-lint' can take 40+ seconds to run
func AllNoLint() error {
	mg.SerialDeps(
		Build,
		Test,
		TestWasm,
	)
	return nil
}

// Like 'All' but also runs the legacy wasm test suite
func AllWithLegacyWasm() error {
	mg.SerialDeps(
		Build,
		Lint,
		Test,
		TestWasm,
		TestLegacyWasm,
	)
	return nil
}

// Build the 'vugu' package, the 'vugugen' and 'vugufmt' commands
func Build() error {
	mg.SerialDeps(Clean)
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
	// install the vugufmt command by executing
	return goInstall("github.com/vugu/vugu/cmd/vugufmt")
}

// Pulls the latest version of the 'golangci-lint' docker image. 'The 'vugu' project used 'golangci-lint' as its lint tool. See the Lint target
func PullLatestGolangCiLintDockerImage() error {
	mg.Deps(CheckRequiredCmdsExist)
	return dockerPullImage(GoLangCiLintImageName)
}

// Lints the 'vugu' source code including the tests and any generated files. The linter used is a dockerized version of the 'golangci-lint' linter.
// See: https://golangci-lint.run/ for more details of the linter
func Lint() error {
	mg.SerialDeps(
		Build,
		PullLatestGolangCiLintDockerImage,
	)
	return runGolangCiLint()
}

// Builds and runs all of the tests for 'vugu', except the 'legacy-wasm-test-suite' tests using the standard Go compiler.
func Test() error {
	mg.SerialDeps(Build)
	packages, err := goListPackages()
	if err != nil {
		return err
	}
	for _, pkg := range packages {
		if strings.Contains(pkg, LegacyWasmTestSuiteDir) {
			continue // skip the wasm-test-suite packages
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
	mg.SerialDeps(Build, PullLatestNginxImage, PullLatestChromeDpImage)
	cleanupContainers()
	// create the container network named "vugu-net"
	err := createContainerNetwork(VuguContainerNetworkName)
	if err != nil {
		return err
	}
	defer func() {
		_ = cleanupContainerNetwork(VuguContainerNetworkName)
	}()

	// start the nginx container and attach it to the new vugu-net network
	err = startNginxContainer(WasmTestSuiteDir)
	if err != nil {
		return err
	}
	defer func() {
		_ = cleanupContainer(VuguNginxContainerName)
	}()
	// add the vugu-ngix container to the defautl bridgenetwork as well (so the container has outbound internet access)
	err = connectContainerToHostNetwork(VuguNginxContainerName)
	if err != nil {
		return err
	}
	// start the chromedp container
	err = startChromeDpContainer()
	if err != nil {
		return err
	}
	defer func() {
		_ = cleanupContainer(VuguChromeDpContainerName)
	}()
	// add the vugu-chromdp container to the default bridgenetwork as well (so the container has outbound internet access)
	err = connectContainerToHostNetwork(VuguChromeDpContainerName)
	if err != nil {
		return err
	}

	// build the wasm binary
	err = runVugugenInTestDirs()
	if err != nil {
		return err
	}
	err = runGoModTidyInTestDirs()
	if err != nil {
		return err
	}

	err = runGoBuildInTestDirs()
	if err != nil {
		return err
	}

	return runGoTestInTestDirs()
	// cleanup via the deferred functions
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
	mg.SerialDeps(Build, PullLatestNginxImage, PullLatestChromeDpImage)
	cleanupContainers()
	// create the container network named "vugu-net"
	err := createContainerNetwork(VuguContainerNetworkName)
	if err != nil {
		return err
	}
	defer func() {
		_ = cleanupContainerNetwork(VuguContainerNetworkName)
	}()

	// start the nginx container and attach it to the new vugu-net network
	err = startNginxContainer(WasmTestSuiteDir)
	if err != nil {
		return err
	}
	defer func() {
		_ = cleanupContainer(VuguNginxContainerName)
	}()
	// add the vugu-ngix container to the defautl bridgenetwork as well (so the container has outbound internet access)
	err = connectContainerToHostNetwork(VuguNginxContainerName)
	if err != nil {
		return err
	}
	// start the chromedp container
	err = startChromeDpContainer()
	if err != nil {
		return err
	}
	defer func() {
		_ = cleanupContainer(VuguChromeDpContainerName)
	}()
	// add the vugu-chromdp container to the default bridgenetwork as well (so the container has outbound internet access)
	err = connectContainerToHostNetwork(VuguChromeDpContainerName)
	if err != nil {
		return err
	}

	// build the wasm binary
	err = runVugugenForTest(moduleName)
	if err != nil {
		return err
	}
	err = runGoModTidyForTest(moduleName)
	if err != nil {
		return err
	}

	err = runGoBuildForTest(moduleName)
	if err != nil {
		return err
	}

	return runGoTestForTest(moduleName)
	// cleanup via the deferred functions
}

// Checks that all required commands have been installed prior to attempting the build.
// The only required command at present is 'docker'.
func CheckRequiredCmdsExist() error {
	return checkAllCmdsExist("docker") // currently docker is the only required external commands
}

// Searches for any files whose filename matches "*.wasm" and deletes them.
func Clean() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	return filepath.WalkDir(cwd, deleteGeneratedWasmFiles)

}

// Searches for any files generated by `vugugen whose filename matches "*_gen.go"and deletes them.
func CleanAutoGeneratedFiles() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	return filepath.WalkDir(cwd, deleteGeneratedFiles)

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
// It is safe to execute this multiple time without calling the "StopLocalNginx" target.
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
func StartLocalNginx() error {
	mg.SerialDeps(StopLocalNginx,
		PullLatestNginxImage,
	)
	err := startLocalNginxContainer(WasmTestSuiteDir)
	if err == nil {
		log.Printf("Local nginx container started.\nConnect to http://localhost:8888/<wasm-test-directory-name>\ne,g. http://localhost:8888/test-001-simple")
		log.Printf("To stop the local nginx container please run:\n\tmage StopLocalNginx")
	}
	return err
}

func StartLocalNginxForExamples() error {
	mg.SerialDeps(StopLocalNginx,
		PullLatestNginxImage,
	)

	err := startLocalNginxContainer(ExamplesDir)
	if err == nil {
		log.Printf("Local nginx container started.\nConnect to http://localhost:8888/<example-test-directory-name>\ne,g. http://localhost:8888/fetch-and-display")
		log.Printf("To stop the local nginx container please run:\n\tmage StopLocalNginx")
	}
	return err
}

// Stops any locally running nginx container.
// See: StartLocalNginx
func StopLocalNginx() error {
	// we intended to ignore the error as the container might not be running
	_ = cleanupContainer(VuguNginxContainerName)
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
	mg.SerialDeps(Build, PullLatestNginxImage, StartLocalNginxForExamples)
	log.Printf("After StartLocalNginxForExasmples")
	//cleanupContainers()
	log.Printf("After cleanupContainers")
	// build the wasm binary
	err := runVugugenInExampleDirs()
	if err != nil {
		return err
	}
	err = runGoModTidyInExampleDirs()
	if err != nil {
		return err
	}

	return runGoBuildInExampleDirs()
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
	mg.SerialDeps(Build, PullLatestNginxImage, StartLocalNginxForExamples)

	// build the wasm binary
	err := runVugugenForExample(moduleName)
	if err != nil {
		return err
	}
	err = runGoModTidyForExample(moduleName)
	if err != nil {
		return err
	}

	return runGoBuildForExample(moduleName)
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
