//go:build mage

package main

import "github.com/magefile/mage/sh"

func runVugugenInCurrentDirSkipMain() error {
	return sh.RunV("vugu", "gen", "--skip-main")
}

func runVugugenInCurrentDir() error {
	return sh.RunV("vugu", "gen")
}
