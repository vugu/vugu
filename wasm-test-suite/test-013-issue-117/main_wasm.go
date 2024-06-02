//go:build wasm
// +build wasm

package main

import (
	"fmt"

	"github.com/vugu/vugu"
	"github.com/vugu/vugu/domrender"
)

func main() {

	var mountPoint *string
	{
		mp := "#vugu_mount_point"
		mountPoint = &mp
	}

	fmt.Printf("Entering main(), -mount-point=%q\n", *mountPoint)

	renderer, err := domrender.New(*mountPoint)
	if err != nil {
		panic(err)
	}

	buildEnv, err := vugu.NewBuildEnv(renderer.EventEnv())
	if err != nil {
		panic(err)
	}

	rootBuilder := vuguSetup(buildEnv, renderer.EventEnv())

	for ok := true; ok; ok = renderer.EventWait() {

		buildResults := buildEnv.RunBuild(rootBuilder)

		err = renderer.Render(buildResults)
		if err != nil {
			panic(err)
		}
	}

}
