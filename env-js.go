// +build wasm

package vugu

// JSEnv is an environment that renders to DOM in webassembly applications.
type JSEnv struct {
	MountPoint       string
	ComponentTypeMap map[string]ComponentType // TODO: probably make this it's own type and have a global instance where things can register
}
