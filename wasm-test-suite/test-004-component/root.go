package main

type Root struct {
	ItemCount int `vugu:"data"`
}

func (c *Root) BeforeBuild() {
	if c.ItemCount == 0 {
		c.ItemCount = 3
	}
}

func (c *Root) OnAdd() {
	c.ItemCount++
}
