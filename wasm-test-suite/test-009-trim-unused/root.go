package main

import (
	"fmt"
	"log"

	"github.com/vugu/vugu"
)

type Root struct {
	Numbers []string `vugu:"data"`
}

func (c *Root) Make(cnt int, event vugu.DOMEvent) {
	event.PreventDefault()
	go func() {
		ee := event.EventEnv()
		ee.Lock()
		defer ee.UnlockRender()

		log.Printf("show %d", cnt)
		c.Numbers = make([]string, cnt)
		for i := 0; i < cnt; i++ {
			c.Numbers[i] = fmt.Sprintf("n%dof%d", i+1, cnt)
		}
	}()
}
