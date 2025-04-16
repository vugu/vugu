package main

import "github.com/vugu/vugu"

type Root struct {
	Something int
	Success   bool
}

func (c *Root) OnClickRun(event vugu.DOMEvent, n int) {
	c.Success = !c.Success
	c.Something += n
}
