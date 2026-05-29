// +build tinygo

package vugu

type ModTracker struct {
	// old map[any]mtResult
	// cur map[any]mtResult
}

func (mt *ModTracker) TrackNext() {
	panic("TBD tinygo")
}

func (mt *ModTracker) ModCheckAll(values ...any) (ret bool) {
	panic("TBD tinygo")
}
