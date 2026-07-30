package main

type Root struct {
	ShowC1 bool   `vugu:"data"`
	ShowC2 bool   `vugu:"data"`
	C1Log  string `vugu:"data"`
	C2Log  string `vugu:"data"`
}

func (c *Root) Init() {
	c.ShowC1 = true
	c.ShowC2 = true
}
