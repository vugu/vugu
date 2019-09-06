package vugu

import "fmt"

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

// type ModCheckedString string
// func (s ModCheckedString) ModCheck(mt *ModTracker, oldData interface{}) (isModified bool, newData interface{}) {
// }

type mtResult struct {
	modified bool
	data     interface{}
}

// ModTracker tracks modifications and maintains the appropriate state for this.
type ModTracker struct {
	old map[interface{}]mtResult
	cur map[interface{}]mtResult
}

// TrackNext moves the "current" information to the "old" position - starting a new round of change tracking.
// Calls to ModCheckAll after calling TrackNext will compare current values to the values from before the call to TrackNext.
// It is generally called once at the start of each build and render cycle.  This method must be called at least
// once before doing modification checks.
func (mt *ModTracker) TrackNext() {

	// lazy initialize
	if mt.old == nil {
		mt.old = make(map[interface{}]mtResult)
	}
	if mt.cur == nil {
		mt.cur = make(map[interface{}]mtResult)
	}

	// remove all elements from old map
	for k := range mt.old {
		delete(mt.old, k)
	}

	// and swap them
	mt.old, mt.cur = mt.cur, mt.old

}

// ModCheckAll performs a modification check on the values provided.
// For values implementing the ModChecker interface, the ModCheck method will be called.
// Otherwise some built-in types are supported including all primitive single-value types -
// bool, int/uint and all variations, both float types, both complex types, string.
// Pointers to supported types are supported.
// Arrays, slices and maps using supported types are supported.
// Structs cannot be passed by value, but pointers to structs will be checked by
// checking each field with a struct tag like `vugu:"modcheck"`.  Slices and arrays of
// structs are okay, since their members have a stable position in memory and a pointer
// can be taken.  Maps using structs however must use pointers to them
// (restriction applies to both keys and values) to be supported.
// Passing an unsupported type will result in a panic.
func (mt *ModTracker) ModCheckAll(values ...interface{}) (ret bool) {

	for _, v := range values {

		// check if we've already done a mod check on v
		curres, ok := mt.cur[v]
		if ok {
			ret = ret || curres.modified
			continue
		}

		// look up the old data
		oldres := mt.old[v]

		// the result of the mod check on v goes here
		var mod bool
		var newdata interface{}

		// see if it implements the ModChecker interface
		mc, ok := v.(ModChecker)
		if ok {
			mod, newdata = mc.ModCheck(mt, oldres.data)
		} else {

			// support for certain built-in types
			switch v.(type) {
			// hm, better to do this or use reflection?  Probably need to use reflection for things like struct pointers
			default:
				panic(fmt.Errorf("not yet implemented"))
			}

		}

		ret = ret || mod
		mt.cur[v] = mtResult{modified: mod, data: newdata}

	}

	return ret
}

// ModCheck(oldData interface{}) (isModified bool, newData interface{})

// hm, this may not work - what happens if we call ModCheck twice in a row?? we need a clear
// way to demark the "old" and "new" versions of data - it might be that the comparison and
// the gather of the new value need to be separate methods?  Or can ModTracker somehow
// prevent ModCheck from being call twice in the same pass (or prevent that from being an issue)

// actually that might work - if there is a call on ModTracker that moves "new" to "old", it
// woudl be pretty clear - and then calling this would only update the "new" value.  Also the
// presence of something in the "new" data would mean it's already been called in this pass
// and so can be deduplicated.
