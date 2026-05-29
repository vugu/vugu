package main

import "github.com/vugu/vugu/vgform"

type Root struct {
	FoodGroup       string
	Textarea1Value  string
	Inputtext1Value string
}

func (c *Root) SetStringPtrDefault(g *string, s string) vgform.StringPtr {
	return vgform.StringPtrDefault(g, s)
}

func (c *Root) SetSliceOptions() vgform.SliceOptions {
	return vgform.SliceOptions{"sandwich_group", "cow_group", "jungle_group", "butterfinger_group"}
}
