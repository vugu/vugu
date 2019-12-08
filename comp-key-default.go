// +build !tinygo

package vugu

// CompKey is the key used to identify and look up a component instance.
type CompKey struct {
	ID      uint64      // unique ID for this instance of a component, randomly generated and embeded into source code
	IterKey interface{} // optional iteration key to distinguish the same component reference in source code but different loop iterations
}

// MakeCompKey creates a CompKey from the id and iteration key you provide.
// The purpose is to hide the implementation of CompKey as it can vary.
func MakeCompKey(id uint64, iterKey interface{}) CompKey {
	return CompKey{ID: id, IterKey: iterKey}
}
