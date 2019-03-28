package distutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WasmExecJsPath find wasm_exec.js in the local Go distribution and return it's path.
// Return error if not found.
func WasmExecJsPath() (string, error) {

	b, err := exec.Command("go", "env", "GOROOT").CombinedOutput()
	if err != nil {
		return "", err
	}
	bstr := strings.TrimSpace(string(b))
	if bstr == "" {
		return "", fmt.Errorf("failed to find wasm_exec.js, empty path from `go env GOROOT`")
	}

	p := filepath.Join(bstr, "misc/wasm/wasm_exec.js")
	_, err = os.Stat(p)
	if err != nil {
		return "", err
	}

	return p, nil
}

// MustWasmExecJsPath find wasm_exec.js in the local Go distribution and return it's path.
// Panic if not found.
func MustWasmExecJsPath() string {
	s, err := WasmExecJsPath()
	if err != nil {
		panic(err)
	}
	return s
}
