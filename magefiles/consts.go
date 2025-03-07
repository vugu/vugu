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

const GoRoot = "GOROOT"       // GOROOT env var name
const GoVersion = "GOVERSION" // GOVERSION env var name
// The Go 1.24 release notes state that the "wasm_exec.js1" file has been moved from:
// $GOROOT/misc/wasm to $GOROOT/lib/wasm.
// While we still build with a pre go1.24 versions we ned to account for the different paths
const WasmExecJSPathMiscDir = "misc" // used for pre go1.24
const WasmExecJSPathLibDir = "lib"   // used for go1.24 and onwards
const WasmExecJSPathWasmDir = "wasm"
const WasmExecJS = "wasm_exec.js"

const Go1_24_0SemVer = "v1.24.0"
