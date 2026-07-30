//go:build js && wasm

package main

import (
	"fmt"

	"github.com/vugu/vugu"
	"github.com/vugu/vugu/domrender"
)

func main() {
	// this is the id of the div where the wasm will be loaded.
	// this MUST match the div id defined in the index.html file (circa line 14)
	mountPoint := "#vugu_mount_point"
	fmt.Printf("Entering main(), -mount-point=%q\n", mountPoint) // thjs goes to the javascript console in the browser
	defer fmt.Printf("Exiting main()\n")

	renderer, err := domrender.New(mountPoint)
	if err != nil {
		panic(err)
	}
	defer renderer.Release()

	buildEnv, err := vugu.NewBuildEnv(renderer.EventEnv())
	if err != nil {
		panic(err)
	}

	// By default 'Root` is the name of the top most component
	// but in this case we have renamed it to "Parent"
	// This struct is defined in parent.go.
	// The generated code that allows methods on property on an instance of Root to be called or accessed
	// is in parent_gen.go
	parentBuilder := &Parent{}

	// wait until we get a render event
	for ok := true; ok; ok = renderer.EventWait() {
		// build a new DOM tree to update the page state
		buildResults := buildEnv.RunBuild(parentBuilder)
		// render the new DOM tree to update the browser display
		err = renderer.Render(buildResults)
		if err != nil {
			panic(err)
		}
	}

}
