package main

import (
	"fmt"
	"slices"
)

type Root struct {
	Show bool  `vugu:"data"`
	List []int `vugu:"data"`
}

func (c *Root) Init() {
	c.Show = true
	c.List = []int{1, 2, 3, 4, 5}
}

func (c *Root) ToggleShow() bool {
	c.Show = !c.Show
	fmt.Printf("c.Show: %v\n", c.Show)
	return c.Show
}

func (c *Root) Push() {
	c.List = slices.Insert(c.List, len(c.List), len(c.List)+1)
	fmt.Printf("Push c.List: %v\n", c.List)
}

func (c *Root) Pop() {
	if len(c.List) > 0 {
		c.List = slices.Delete(c.List, len(c.List)-1, len(c.List))
	}
	fmt.Printf("Pop  c.List: %v\n", c.List)
}

func (c *Root) Reverse() {
	slices.Reverse(c.List)
}
