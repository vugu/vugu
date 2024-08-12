//go:build mage

package main

import (
	"os"

	"github.com/magefile/mage/sh"
)

func createContainers() error {
	// create the container network named "vugu-net"
	err := createContainerNetwork(VuguContainerNetworkName)
	if err != nil {
		return err
	}
	// start the nginx container and attach it to the new vugu-net network
	err = startContainerForWasmTestSuiteWithNetwork()
	if err != nil {
		return err
	}
	// add the vugu-ngix container to the default bridge network as well (so the container has outbound internet access)
	err = connectContainerToHostNetwork(VuguNginxWasmTestsContainerName)
	if err != nil {
		return err
	}
	// start the chromedp container
	err = startChromeDpContainer()
	if err != nil {
		return err
	}
	// add the vugu-chromdp container to the default bridge network as well (so the container has outbound internet access)
	err = connectContainerToHostNetwork(VuguChromeDpContainerName)
	if err != nil {
		return err
	}
	return err
}

func cleanupContainers() {
	// remove any bridge network that we may have created previously - before removing any containers that my be left over
	// this must be in the reverse order of createContainers above
	_ = cleanupContainer(VuguChromeDpContainerName)
	_ = cleanupContainer(VuguNginxWasmTestsContainerName)
	_ = cleanupContainerNetwork(VuguContainerNetworkName)

}

func cleanupContainer(containerName string) error {
	// We do not care about errors here. The commands can fail if the named container has not started.
	// This can be the case during container cleanup when the first container started byt the 2nd errored and the
	// 3rd never started. We want this function to clean up all of the running containers and ignore any that never started
	// stop the container
	_ = dockerCmdV("stop", containerName)
	// remove the container
	_ = dockerCmdV("rm", containerName)
	return nil
}

func cleanupContainerNetwork(networkName string) error {
	// remove the vugu-net network - again ignore the error
	_ = dockerCmdV("network", "rm", networkName)
	return nil
}

func runGolangCiLint() error { // we do mean run and not start in the function name as this container runs once and then exits.
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	dashVArg := cwd + ":/app"
	return dockerCmdV("run", "--rm", "-v", dashVArg, "-w", "/app", GoLangCiLintImageName, "golangci-lint", "run", "-v")
}

// we INTEND to ignore any errors here, as the containers might not be running.
func stopContainerForWasmTestSuite() {
	_ = cleanupContainer(VuguNginxWasmTestsContainerName)
}

func stopContainerForExamples() {
	_ = cleanupContainer(VuguNginxExamplesContainerName)
}

func stopLegacyWasmTestSuiteContainer() {
	_ = cleanupContainer(LegacyWasmTestSuiteContainerName)
}

func startLegacyWasmTestSuiteContainer() error {
	return dockerCmdV("run", "-d", "-t", "-p", "9222:9222", "-p", "8846:8846", "--name", LegacyWasmTestSuiteContainerName, LegacyWasmTestSuiteImageName)
}

func buildLegacyWasmTestSuiteContainer() error {
	return dockerCmdV("build", "-t", LegacyWasmTestSuiteImageName, ".")
}

func createContainerNetwork(networkName string) error {
	return dockerCmdV("network", "create", "--driver", "bridge", networkName)
}

func startContainerForWasmTestSuiteWithNetwork() error {
	return startNginxContainerWithNetwork(VuguNginxWasmTestsContainerName, VuguNginxWasmTestsContainerPort, WasmTestSuiteDir, VuguContainerNetworkName)
}

func startContainerForWasmTestSuite() error {
	return startNginxContainer(VuguNginxWasmTestsContainerName, VuguNginxWasmTestsContainerPort, WasmTestSuiteDir)
}

func startContainerForExamples() error {
	return startNginxContainer(VuguNginxExamplesContainerName, VuguNginxExamplesContainerPort, ExamplesDir)
}

func startNginxContainer(containerName, port, dirToServe string) error {
	// note we must prefix "./" to the dirToServe as the mount requires an absolute path
	htmlMountArg := "type=bind,source=./" + dirToServe + ",target=/usr/share/nginx/html,readonly"
	configMountArg := "type=bind,source=./wasm-test-suite/nginx-test.conf,target=/usr/share/nginx/config/config.d/nginx-test.conf,readonly"	
	portArg := port + ":80"
	// start the nginx container and attached it to the default hosts bridge network
	return dockerCmdV("run", "--name", containerName, "--mount", htmlmountArg, "--mount", configMountArg, "-p", portArg, "-d", "nginx")
}

func startNginxContainerWithNetwork(containerName, port, dirToServe, networkName string) error {
	// note we must prefix "./" to the dirToServe as the mount requires an absolute path
	htmlMountArg := "type=bind,source=./" + dirToServe + ",target=/usr/share/nginx/html,readonly"
	configMountArg := "type=bind,source=./wasm-test-suite/nginx-test.conf,target=/usr/share/nginx/config/config.d/nginx-test.conf,readonly"	
	portArg := port + ":80"
	// start the nginx container and attach it to the new vugu-net network
	return dockerCmdV("run", "--name", containerName, "--network", networkName, "--mount", htmlmountArg, "--mount", configMountArg, "-p", portArg, "-d", "nginx")
}

func dockerPullImage(imageName string) error {
	return dockerCmdV("pull", imageName)
}

func connectContainerToHostNetwork(containerName string) error {
	return dockerCmdV("network", "connect", "bridge", containerName)
}

func startChromeDpContainer() error {
	return dockerCmdV("run", "-d", "-t", "-p", "9222:9222", "--name", VuguChromeDpContainerName, "--network", VuguContainerNetworkName, "chromedp/headless-shell")
}

func dockerCmdV(args ...string) error {
	return sh.RunV("docker", args...)
}
