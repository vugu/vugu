package vugu

// ModChecker interface is implemented by types that want to implement their own modification tracking.
// The ModCheck method is passed a ModTracker (for use in checking child values for modification if needed),
// and the prior data value stored corresponding to this value (will be nil on the first call).
// ModCheck should return true if this instance is modified, as well as a new data value to be stored.
// It is up to the specific implementation to decide what type to use for oldData and newData and how to
// compare them.  In some cases it may be appropriate to return the value itself or a summarized version
// of it.  But for types that use lots of memory and would otherwise take too much time to traverse, a
// counter or other value can be used to indicate that some changed is occured, with the rest of the
// application using mutator methods to increment this value upon change.
type ModChecker interface {
	ModCheck(mt *ModTracker, oldData interface{}) (isModified bool, newData interface{})
}

// NewModTracker creates an empty ModTracker, calls TrackNext on it, and returns it.
func NewModTracker() *ModTracker {
	var ret ModTracker
	ret.TrackNext()
	return &ret
}

type mtResult struct {
	modified bool
	data     interface{}
}
