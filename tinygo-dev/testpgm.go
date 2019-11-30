package main

import (
	"fmt"

	"github.com/vugu/vjson"
	"github.com/vugu/vugu"
	"github.com/vugu/vugu/domrender"
)

func main() {

	var r vjson.RawMessage
	var be vugu.BuildEnv

	var jr domrender.JSRenderer

	// log.Printf("hello there!")
	fmt.Printf("hello testpgm: %v %v %v\n", r, be, jr)
	fmt.Printf("hello testpgm: %v\n", jr)
	// println("blah blah")
}
