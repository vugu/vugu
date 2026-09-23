package main

import "github.com/vugu/vugu"

var _ vugu.DOMEvent // import fixer

// ClickHandler is the interface for things that can handle ClickEvent.
type ClickHandler interface {
	ClickHandle(event ClickEvent)
}

// ClickFunc implements ClickHandler as a function.
type ClickFunc func(event ClickEvent)

// ClickHandle implements the ClickHandler interface.
func (f ClickFunc) ClickHandle(event ClickEvent) { f(event) }

// assert ClickFunc implements ClickHandler
var _ ClickHandler = ClickFunc(nil)

