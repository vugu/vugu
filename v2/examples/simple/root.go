package main

import "time"

type Root struct {
	ShowWasm bool `vugu:"data"`
	ShowGo   bool `vugu:"data"`
	ShowVugu bool `vugu:"data"`
}

func (r *Root) After2020() bool {
	return time.Now().Year() >= 2020
}
