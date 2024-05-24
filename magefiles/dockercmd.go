package main

import (
	"os"

	"github.com/magefile/mage/sh"
)

const GoLangCiLintImageName = "golangci/golangci-lint:latest-alpine"
const TinyGoImageName = "tinygo/tinygo:latest"
const WasmTestSuiteImageName = "vugu/wasm-test-suite:latest"
const NginxImageName = "nginx:latest"
const ChromeDpHeadlessShellImageName = "chromedp/headless-shell:latest"

const WasmTestSuiteContainerName = "wasm-test-suite"
const VuguNginxContainerName = "vugu-nginx"
const VuguChromeDpContainerName = "vugu-chromedp"

const VuguContainerNetworkName = "vugu-net"

const WasmTestSuiteDir = "./new-tests"
const LegacyWasmTestSuiteDir = "./wasm-test-suite"

func cleanupContainers() {
	// remove any bridge network that we may have created previously - before removing any containers that my be left over
	_ = cleanupContainerNetwork(VuguContainerNetworkName)
	_ = cleanupContainer(VuguNginxContainerName)
	_ = cleanupContainer(VuguChromeDpContainerName)
}

func cleanupContainer(containerName string) error {
	if err := dockerCmdV("stop", containerName); err != nil {
		return err
	}
	// remove the vugu-nginx container - again ignore the error
	return dockerCmdV("rm", containerName)
}

func cleanupContainerNetwork(networkName string) error {
	// remove the vugu-nginx container - again ignore the error
	return dockerCmdV("network", "rm", networkName)
}

func runGolangCiLint() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	dashVArg := cwd + ":/app"
	return dockerCmdV("run", "--rm", "-v", dashVArg, "-w", "/app", GoLangCiLintImageName, "golangci-lint", "run", "-v")
}

// we INTEND to ignore any errors here, as the containers might not be running.
func stopWasmTestSuiteContainer() {
	_ = dockerCmdV("stop", WasmTestSuiteContainerName)
	_ = dockerCmdV("rm", WasmTestSuiteContainerName)
}

func startWasmTestSuiteContainer() error {
	return dockerCmdV("run", "-d", "-t", "-p", "9222:9222", "-p", "8846:8846", "--name", WasmTestSuiteContainerName, WasmTestSuiteImageName)
}

func buildWasmTestSuiteContainer() error {
	return dockerCmdV("build", "-t", WasmTestSuiteImageName, ".")
}

func createContainerNetwork(networkName string) error {
	return dockerCmdV("network", "create", "--driver", "bridge", networkName)
}

func startNginxContainer(dirToServe string) error {
	// not NO "./" in the dirToServe
	mountArg := "type=bind,source=" + dirToServe + ",target=/usr/share/nginx/html,readonly"
	// start the nginx container and attach it to the new vugu-net network
	return dockerCmdV("run", "--name", VuguNginxContainerName, "--network", VuguContainerNetworkName, "--mount", mountArg, "-p", "8888:80", "-d", "nginx")
}

func startLocalNginxContainer(dirToServe string) error {
	// not NO "./" in the dirToServe
	mountArg := "type=bind,source=" + dirToServe + ",target=/usr/share/nginx/html,readonly"
	// start the nginx container and attached it to the default hosts bridge network
	return dockerCmdV("run", "--name", VuguNginxContainerName, "--mount", mountArg, "-p", "8888:80", "-d", "nginx")
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
