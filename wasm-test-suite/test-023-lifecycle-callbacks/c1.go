package main

import (
	"fmt"

	"github.com/vugu/vugu"
)

type C1 struct {
	Parent *Root
}

// this version has contexts for each call

func (c *C1) Init(ctx vugu.InitCtx) {
	c.appendToLog("got C1.Init(ctx)")
}

func (c *C1) Compute(ctx vugu.ComputeCtx) {
	c.appendToLog("got C1.Compute(ctx)")
}

func (c *C1) Rendered(ctx vugu.RenderedCtx) {
	//c.appendToLog("got C1.Rendered(ctx)")
	c.appendToLog(fmt.Sprintf("got C1.Rendered(ctx)[first=%t]", ctx.First()))
}

func (c *C1) Destroy(ctx vugu.DestroyCtx) {
	c.appendToLog("got C1.Destroy(ctx)")
}

func (c *C1) Log() string {
	return c.Parent.C1Log
}

func (c *C1) appendToLog(s string) {
	c.Parent.C1Log = c.Parent.C1Log + s + "\n"

}
