package main 

import "log"
import "syscall/js"
import "time"

func runMemTest() {

	log.Printf("runMemTest start")
	defer log.Printf("runMemTest end")

	// var eval = js.Global().Get("eval")
	// log.Printf("eval = %v", eval)

	var g = js.Global()

	// 25us per call
	// g.Call("eval", "function testf1(a) { return a + 'f1'; }")

	// 42us per call
	g.Call("eval", "function testf1(a) { var e = document.createElement('div'); e.innerHTML = a + 'f1'; document.body.appendChild(e); }")

	var blah = g.Call("testf1", "blah").String()
	log.Printf("blah = %q", blah)

	n := 1000
	t := time.Now()
	for i := 0; i < n; i++ {
		_ = g.Call("testf1", "bleh").String()
	}
	log.Printf("avg call time: %v", time.Since(t) / time.Duration(n))

}