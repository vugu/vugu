package main

const GoLangCiLintImageName = "golangci/golangci-lint:latest-alpine"
const TinyGoImageName = "tinygo/tinygo:latest"
const LegacyWasmTestSuiteImageName = "vugu/wasm-test-suite:latest"
const NginxImageName = "nginx:latest"
const ChromeDpHeadlessShellImageName = "chromedp/headless-shell:latest"

const LegacyWasmTestSuiteContainerName = "wasm-test-suite"
const VuguNginxContainerName = "vugu-nginx"
const VuguChromeDpContainerName = "vugu-chromedp"

const VuguContainerNetworkName = "vugu-net"

const WasmTestSuiteDir = "wasm-test-suite"
const LegacyWasmTestSuiteDir = "legacy-wasm-test-suite"
const ExamplesDir = "examples"
