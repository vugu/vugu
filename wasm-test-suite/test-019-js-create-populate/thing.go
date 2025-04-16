//go:build !js || !wasm

package main

import (
	"fmt"
	"log"

	"github.com/vugu/vugu/js"

	"github.com/vugu/vugu"
)

type Thing struct {
	Log []string `vugu:"data"`
}

func (c *Thing) handleJSCreate(value js.Value) {
	log.Printf("vg-js-create: %#v", value)
	log.Printf("vg-js-create found class: %s", value.Get("className").String())
	log.Printf("vg-js-create firstElementChild: %v", value.Get("firstElementChild"))
	// NOTE: We're modifying c.Log during the render pass, so we won't see this
	// until the next time around - this is generally not a good idea but it works
	// for the purposes of the test because we can't get the console output
	// via chromedp in wasm-suite_test.go
	c.Log = append(c.Log, fmt.Sprintf("vg-js-create className %s", value.Get("className").String()))
	c.Log = append(c.Log, fmt.Sprintf("vg-js-create firstElementChild %v", value.Get("firstElementChild")))
}

func (c *Thing) handleJSPopulate(value js.Value) {
	log.Printf("vg-js-populate: %#v", value)
	log.Printf("vg-js-populate firstElementChild: %v", value.Get("firstElementChild"))
	c.Log = append(c.Log, fmt.Sprintf("vg-js-populate className %s", value.Get("className").String()))
	c.Log = append(c.Log, fmt.Sprintf("vg-js-populate firstElementChild %v", value.Get("firstElementChild")))
}

func (c *Thing) handleButtonClick(event vugu.DOMEvent) {
}
