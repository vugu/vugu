// +build tinygo

package vugu

type ModTracker struct {
	// old map[interface{}]mtResult
	// cur map[interface{}]mtResult
}

func (mt *ModTracker) TrackNext() {
	panic("TBD tinygo")
}

func (mt *ModTracker) ModCheckAll(values ...interface{}) (ret bool) {
	panic("TBD tinygo")
}
