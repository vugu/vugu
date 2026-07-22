//go:build wasm && js

package main

import (
	"fmt"

	"github.com/vugu/vugu/v2"
)

type Thing struct {
	Parent *Root
}

type clickEvt struct {
	e vugu.DOMEvent
}

func (c *Thing) HandleButtonClick(event vugu.DOMEvent) {
	clickEvt := clickEvt{e: event}
	c.Parent.ShowText = fmt.Sprintf("%s", clickEvt)
}

func (e clickEvt) String() string {
	return "This is the ClickEvent"
}
