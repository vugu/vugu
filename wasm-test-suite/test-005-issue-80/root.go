package main

type Root struct{}

func (c *Root) Items() []string {
	return []string{"a", "b", "c", "d"}
}
