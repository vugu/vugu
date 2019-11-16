package main

import (
	"fmt"

	"github.com/vugu/vjson"
	"github.com/vugu/vugu"
)

func main() {

	var r vjson.RawMessage
	var be vugu.BuildEnv

	// log.Printf("hello there!")
	fmt.Printf("hello testpgm: %v %v\n", r, be)
	// println("blah blah")
}
