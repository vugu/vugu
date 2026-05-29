package main

import "github.com/vugu/vugu"

type MyDataTable struct {
	AttrMap vugu.AttrMap

	DefaultSlot vugu.Builder
	AnotherSlot vugu.Builder
	SlotMap     map[string]vugu.Builder
}
