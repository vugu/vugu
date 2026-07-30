//go:build wasm && js

package main

import "github.com/vugu/vugu/v2"

type MyDataTable struct {
	AttrMap vugu.AttrMap

	DefaultSlot vugu.Builder
	AnotherSlot vugu.Builder
	SlotMap     map[string]vugu.Builder
}
