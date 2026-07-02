package main

import (
	"github.com/vugu/vgrouter"
	"github.com/vugu/vugu/v2"
)

type Root struct {
	vgrouter.NavigatorRef

	Body vugu.Builder
}
