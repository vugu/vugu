package main

type C2 struct {
	Parent *Root
}

// alternate version without context for simplicity

func (c *C2) Init() {
	c.appendToLog("got C2.Init()")
}

func (c *C2) Compute() {
	c.appendToLog("got C2.Compute()")
}

func (c *C2) Rendered() {
	c.appendToLog("got C2.Rendered()")
}

func (c *C2) Destroy() {
	c.appendToLog("got C2.Destroy()")
}

func (c *C2) Log() string {
	return c.Parent.C2Log
}

func (c *C2) appendToLog(s string) {
	c.Parent.C2Log = c.Parent.C2Log + s + "\n"
}
