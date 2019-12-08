// +build tinygo

package vugu

import "fmt"

// CompKey is a string in TinyGo for the time being
type CompKey string

// MakeCompKey creates a CompKey from the id and iteration key you provide.
// The purpose is to hide the implementation of CompKey as it can vary.
func MakeCompKey(id uint64, iterKey interface{}) CompKey {
	return CompKey(fmt.Sprintf("%x:%v", id, iterKey))
}
