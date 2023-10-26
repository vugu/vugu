package main

import (
	"fmt"

	"github.com/vugu/vjson"
	"github.com/vugu/vugu"
	"github.com/vugu/vugu/domrender"
)

func main() {

	var (
		r  vjson.RawMessage
		be vugu.BuildEnv
		jr domrender.JSRenderer
	)
	// use of jr need to be to investigate

	// log.Printf("hello there!")
	fmt.Printf("hello testpgm: %v %v %v\n", r, be, jr) //nolint
	fmt.Printf("hello testpgm: %v\n", jr)              //nolint
	// println("blah blah")
}
