package main

type Root struct {
	Clicked string
}

func (c *Root) Items() []string {
	return []string{"a", "b", "c", "d", "e"}
}
