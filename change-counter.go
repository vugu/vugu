package vugu

// ChangeCounter is a number that can be incremented to indicate modification.
// The ModCheck method is implemented using this number.
// A ChangeCounter can be used directly or embedded in a struct, with its
// methods calling Changed on each mutating operation, in order to
// track its modification.
type ChangeCounter int

// Changed increments the counter to indicate a modification.
func (c *ChangeCounter) Changed() {
	*c++
}

// ModCheck implements the ModChecker interface.
func (c *ChangeCounter) ModCheck(mt *ModTracker, oldData interface{}) (isModified bool, newData interface{}) {
	oldn, ok := oldData.(ChangeCounter)
	return (!ok) || (oldn != *c), *c
}
