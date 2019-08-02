// +build wasm

package main

import (
	"log"

	"github.com/vugu/vugu"
)

func main() {

	println("Entering main()")
	defer println("Exiting main()")

	// runMemTest()

	rootBuilder := &Root{}

	buildEnv, err := vugu.NewBuildEnv(rootBuilder)
	if err != nil {
		log.Fatal(err)
	}

	renderer, err := vugu.NewJSRenderer("#root_mount_parent")
	if err != nil {
		log.Fatal(err)
	}

	for ok := true; ok; ok = renderer.EventWait() {

		buildOut, err := buildEnv.BuildRoot()
		if err != nil {
			panic(err)
		}

		err = renderer.Render(buildOut)
		if err != nil {
			panic(err)
		}
	}
	
}
