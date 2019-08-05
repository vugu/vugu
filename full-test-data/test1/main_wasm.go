// +build wasm

package main

import (
	"log"
	"fmt"
	"flag"

	"github.com/vugu/vugu"
)

func main() {

	mountPoint := flag.String("mount-point", "#root_mount_point", "The query selector for the mount point for the root component, if it is not a full HTML component")
	flag.Parse()

	fmt.Printf("Entering main(): %#v\n", *mountPoint)
	defer fmt.Printf("Exiting main()\n")

	// runMemTest()

	rootBuilder := &Root{}

	buildEnv, err := vugu.NewBuildEnv(rootBuilder)
	if err != nil {
		log.Fatal(err)
	}

	renderer, err := vugu.NewJSRenderer(*mountPoint)
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
