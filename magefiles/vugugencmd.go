//go:build mage

package main

import "github.com/magefile/mage/sh"

func runVugugenInCurrentDir() error {
	return sh.RunV("vugugen")
}
