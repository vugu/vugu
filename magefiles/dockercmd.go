package main

import (
	"os"

	"github.com/magefile/mage/sh"
)

func cleanupContainers() {
	// remove any bridge network that we may have created previously - before removing any containers that my be left over
	_ = cleanupContainerNetwork(VuguContainerNetworkName)
	_ = cleanupContainer(VuguNginxContainerName)
	_ = cleanupContainer(VuguChromeDpContainerName)
}

func cleanupContainer(containerName string) error {
	dockerCmdV("stop", containerName)
	// remove the vugu-nginx container - again ignore the error
	dockerCmdV("rm", containerName)
	return nil
}

func cleanupContainerNetwork(networkName string) error {
	// remove the vugu-nginx container - again ignore the error
	_ = dockerCmdV("network", "rm", networkName)
	return nil
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
func stopLegacyWasmTestSuiteContainer() {
	_ = dockerCmdV("stop", LegacyWasmTestSuiteContainerName)
	_ = dockerCmdV("rm", LegacyWasmTestSuiteContainerName)
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

func startNginxContainer(dirToServe string) error {
	// note we must prefix "./" to the dirToServe as the mount requires an absolute path
	mountArg := "type=bind,source=./" + dirToServe + ",target=/usr/share/nginx/html,readonly"
	// start the nginx container and attach it to the new vugu-net network
	return dockerCmdV("run", "--name", VuguNginxContainerName, "--network", VuguContainerNetworkName, "--mount", mountArg, "-p", "8888:80", "-d", "nginx")
}

func startLocalNginxContainer(dirToServe string) error {
	// note we must prefix "./" to the dirToServe as the mount requires an absolute path
	mountArg := "type=bind,source=./" + dirToServe + ",target=/usr/share/nginx/html,readonly"
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
