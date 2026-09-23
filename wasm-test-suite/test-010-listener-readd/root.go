package main

import (
	"fmt"

	"github.com/vugu/vugu"
)

type Root struct {
	ActiveView string `vugu:"data"`
}

func (c *Root) SwitchView(event vugu.DOMEvent, viewname string) {
	fmt.Println("switching to " + viewname)
	go func() {
		ee := event.EventEnv()
		ee.Lock()
		c.ActiveView = viewname
		ee.UnlockRender()
	}()
}
