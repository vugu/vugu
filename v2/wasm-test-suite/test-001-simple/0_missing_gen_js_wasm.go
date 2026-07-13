//go:build js && wasm

package main

import "github.com/vugu/vugu/v2"

var _ vugu.DOMEvent // import fixer

// Root is a Vugu component and implements the vugu.Builder interface.
type Root struct {}

