package main

type Root struct{}

func (c *Root) Lang() string {
	return "en"
}
func (c *Root) HeadClass() string {
	return "head-class"
}
