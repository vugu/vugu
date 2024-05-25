package main

import (
	"github.com/vugu/vugu"
)

type Root struct{}

func (c *Root) AttributeList() []vugu.VGAttribute {
	return []vugu.VGAttribute{
		{
			Key: "class",
			Val: "widget",
		},
		{
			Key: "data-test",
			Val: "test",
		},
	}
}

func testFunc() []vugu.VGAttribute {
	return []vugu.VGAttribute{
		{
			Key: "class",
			Val: "funcwidget",
		},
		{
			Key: "data-test",
			Val: "functest",
		},
	}
}
