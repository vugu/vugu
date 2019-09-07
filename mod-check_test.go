package vugu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModCheckerStrings(t *testing.T) {

	assert := assert.New(t)

	var mt ModTracker

	mt.TrackNext()
	s := "testing" // initial value
	assert.True(mt.ModCheckAll(&s))

	mt.TrackNext()
	// no change
	assert.False(mt.ModCheckAll(&s))

	mt.TrackNext()
	s = "testing2" // different value
	assert.True(mt.ModCheckAll(&s))

	mt.TrackNext()
	s = "testing" // back to earlier value
	assert.True(mt.ModCheckAll(&s))

	mt.TrackNext()
	s = "testing" // same value
	assert.False(mt.ModCheckAll(&s))

	// run through it again with a byte slice

	mt.TrackNext()
	b := []byte("testing") // initial value
	assert.True(mt.ModCheckAll(&b))

	mt.TrackNext()
	// no change
	assert.False(mt.ModCheckAll(&b))

	mt.TrackNext()
	b = []byte("testing2") // different value
	assert.True(mt.ModCheckAll(&b))

	mt.TrackNext()
	b = []byte("testing") // back to earlier value
	assert.True(mt.ModCheckAll(&b))

	mt.TrackNext()
	b = []byte("testing") // same value
	assert.False(mt.ModCheckAll(&b))

}

func TestModCheckerBool(t *testing.T) {

	mt := NewModTracker()

	var v1, v2 bool
	check := func(vp *bool, newv bool, expectedMod bool) {
		mt.TrackNext()
		*vp = newv
		mod := mt.ModCheckAll(vp)
		if mod != expectedMod {
			t.Errorf("check(%#v, %#v, %#v) wrong mod result: %v", vp, newv, expectedMod, mod)
		}
	}

	check(&v1, false, true)
	check(&v1, false, false)
	check(&v1, true, true)
	check(&v1, true, false)
	check(&v1, false, true)
	check(&v1, false, false)
	check(&v1, false, false)
	check(&v2, false, true)
	check(&v2, true, true)
	check(&v2, false, true)
	check(&v2, false, false)

}

func TestModCheckerInt(t *testing.T) {

	mt := NewModTracker()

	var v1, v2 int
	check := func(vp *int, newv int, expectedMod bool) {
		mt.TrackNext()
		*vp = newv
		mod := mt.ModCheckAll(vp)
		if mod != expectedMod {
			t.Errorf("check(%#v, %#v, %#v) wrong mod result: %v", vp, newv, expectedMod, mod)
		}
	}

	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 2, true)
	check(&v1, 2, false)
	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 1, false)
	check(&v2, 1, true)
	check(&v2, 2, true)
	check(&v2, 1, true)
	check(&v2, 1, false)

}

func TestModCheckerInt8(t *testing.T) {

	mt := NewModTracker()

	var v1, v2 int8
	check := func(vp *int8, newv int8, expectedMod bool) {
		mt.TrackNext()
		*vp = newv
		mod := mt.ModCheckAll(vp)
		if mod != expectedMod {
			t.Errorf("check(%#v, %#v, %#v) wrong mod result: %v", vp, newv, expectedMod, mod)
		}
	}

	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 2, true)
	check(&v1, 2, false)
	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 1, false)
	check(&v2, 1, true)
	check(&v2, 2, true)
	check(&v2, 1, true)
	check(&v2, 1, false)

}

func TestModCheckerInt16(t *testing.T) {

	mt := NewModTracker()

	var v1, v2 int16
	check := func(vp *int16, newv int16, expectedMod bool) {
		mt.TrackNext()
		*vp = newv
		mod := mt.ModCheckAll(vp)
		if mod != expectedMod {
			t.Errorf("check(%#v, %#v, %#v) wrong mod result: %v", vp, newv, expectedMod, mod)
		}
	}

	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 2, true)
	check(&v1, 2, false)
	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 1, false)
	check(&v2, 1, true)
	check(&v2, 2, true)
	check(&v2, 1, true)
	check(&v2, 1, false)

}

func TestModCheckerInt32(t *testing.T) {

	mt := NewModTracker()

	var v1, v2 int32
	check := func(vp *int32, newv int32, expectedMod bool) {
		mt.TrackNext()
		*vp = newv
		mod := mt.ModCheckAll(vp)
		if mod != expectedMod {
			t.Errorf("check(%#v, %#v, %#v) wrong mod result: %v", vp, newv, expectedMod, mod)
		}
	}

	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 2, true)
	check(&v1, 2, false)
	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 1, false)
	check(&v2, 1, true)
	check(&v2, 2, true)
	check(&v2, 1, true)
	check(&v2, 1, false)

}

func TestModCheckerInt64(t *testing.T) {

	mt := NewModTracker()

	var v1, v2 int64
	check := func(vp *int64, newv int64, expectedMod bool) {
		mt.TrackNext()
		*vp = newv
		mod := mt.ModCheckAll(vp)
		if mod != expectedMod {
			t.Errorf("check(%#v, %#v, %#v) wrong mod result: %v", vp, newv, expectedMod, mod)
		}
	}

	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 2, true)
	check(&v1, 2, false)
	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 1, false)
	check(&v2, 1, true)
	check(&v2, 2, true)
	check(&v2, 1, true)
	check(&v2, 1, false)

}

func TestModCheckerUint(t *testing.T) {

	mt := NewModTracker()

	var v1, v2 uint
	check := func(vp *uint, newv uint, expectedMod bool) {
		mt.TrackNext()
		*vp = newv
		mod := mt.ModCheckAll(vp)
		if mod != expectedMod {
			t.Errorf("check(%#v, %#v, %#v) wrong mod result: %v", vp, newv, expectedMod, mod)
		}
	}

	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 2, true)
	check(&v1, 2, false)
	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 1, false)
	check(&v2, 1, true)
	check(&v2, 2, true)
	check(&v2, 1, true)
	check(&v2, 1, false)

}

func TestModCheckerUint8(t *testing.T) {

	mt := NewModTracker()

	var v1, v2 uint8
	check := func(vp *uint8, newv uint8, expectedMod bool) {
		mt.TrackNext()
		*vp = newv
		mod := mt.ModCheckAll(vp)
		if mod != expectedMod {
			t.Errorf("check(%#v, %#v, %#v) wrong mod result: %v", vp, newv, expectedMod, mod)
		}
	}

	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 2, true)
	check(&v1, 2, false)
	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 1, false)
	check(&v2, 1, true)
	check(&v2, 2, true)
	check(&v2, 1, true)
	check(&v2, 1, false)

}

func TestModCheckerUint16(t *testing.T) {

	mt := NewModTracker()

	var v1, v2 uint16
	check := func(vp *uint16, newv uint16, expectedMod bool) {
		mt.TrackNext()
		*vp = newv
		mod := mt.ModCheckAll(vp)
		if mod != expectedMod {
			t.Errorf("check(%#v, %#v, %#v) wrong mod result: %v", vp, newv, expectedMod, mod)
		}
	}

	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 2, true)
	check(&v1, 2, false)
	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 1, false)
	check(&v2, 1, true)
	check(&v2, 2, true)
	check(&v2, 1, true)
	check(&v2, 1, false)

}

func TestModCheckerUint32(t *testing.T) {

	mt := NewModTracker()

	var v1, v2 uint32
	check := func(vp *uint32, newv uint32, expectedMod bool) {
		mt.TrackNext()
		*vp = newv
		mod := mt.ModCheckAll(vp)
		if mod != expectedMod {
			t.Errorf("check(%#v, %#v, %#v) wrong mod result: %v", vp, newv, expectedMod, mod)
		}
	}

	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 2, true)
	check(&v1, 2, false)
	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 1, false)
	check(&v2, 1, true)
	check(&v2, 2, true)
	check(&v2, 1, true)
	check(&v2, 1, false)

}

func TestModCheckerUint64(t *testing.T) {

	mt := NewModTracker()

	var v1, v2 uint64
	check := func(vp *uint64, newv uint64, expectedMod bool) {
		mt.TrackNext()
		*vp = newv
		mod := mt.ModCheckAll(vp)
		if mod != expectedMod {
			t.Errorf("check(%#v, %#v, %#v) wrong mod result: %v", vp, newv, expectedMod, mod)
		}
	}

	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 2, true)
	check(&v1, 2, false)
	check(&v1, 1, true)
	check(&v1, 1, false)
	check(&v1, 1, false)
	check(&v2, 1, true)
	check(&v2, 2, true)
	check(&v2, 1, true)
	check(&v2, 1, false)

}
