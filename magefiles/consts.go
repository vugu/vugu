//go:build mage

package main

const DockerCmdExeName = "docker"
const GitCmdExeName = "git"

const GoLangCiLintImageName = "golangci/golangci-lint:latest-alpine"
const TinyGoImageName = "tinygo/tinygo:latest"
const LegacyWasmTestSuiteImageName = "vugu/wasm-test-suite:latest"
const NginxImageName = "nginx:latest"
const ChromeDpHeadlessShellImageName = "chromedp/headless-shell:latest"

const LegacyWasmTestSuiteContainerName = "wasm-test-suite"
const VuguNginxWasmTestsContainerName = "vugu-nginx"
const VuguNginxExamplesContainerName = "vugu-nginx-examples"
const VuguChromeDpContainerName = "vugu-chromedp"

const VuguContainerNetworkName = "vugu-net"

const VuguNginxWasmTestsContainerPort = "8888"
const VuguNginxExamplesContainerPort = "8889"

const WasmTestSuiteDir = "wasm-test-suite"
const LegacyWasmTestSuiteDir = "legacy-wasm-test-suite"
const ExamplesDir = "examples"
const MagefilesDir = "magefiles"

const GoRoot = "GOROOT" // GOROOT env var name
const WasmExecJSPathMiscDir = "misc"
const WasmExecJSPathWasmDir = "wasm"
const WasmExecJS = "wasm_exec.js"
