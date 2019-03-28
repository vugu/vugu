// +build !wasm

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/vugu/vugu/simplehttp"
)

func main() {

	wd, _ := os.Getwd()
	l := ":8875"
	log.Printf("Starting HTTP Server at %q", l)
	h := simplehttp.New(wd, true)
	log.Fatal(http.ListenAndServe(l, h))

}
