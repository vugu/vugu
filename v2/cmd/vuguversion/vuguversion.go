package main

import (
	"fmt"
	"runtime/debug"
)

func main() {
	rev, tag := computeRevision()
	fmt.Printf("rev: %v tag: %v\n", rev, tag)
}


func computeRevision() (string, string) {
	var (
		rev      = "unknown"
		tags     = "unknown"
		modified bool
	)

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Printf("reading BuildInfo failed\n")
		return rev, tags
	}
	fmt.Printf("%v\n", buildInfo.Settings)
	for _, v := range buildInfo.Settings {
		if v.Key == "vcs.revision" {
			rev = v.Value
		}
		if v.Key == "vcs.modified" {
			if v.Value == "true" {
				modified = true
			}
		}
		if v.Key == "-tags" {
			tags = v.Value
		}
		if v.Key == "-ldflags" {
			fmt.Printf("-ldflags: %v\n", v.Value) 
		}
	}
	if modified {
		return rev + "-modified", tags
	}
	return rev, tags
}