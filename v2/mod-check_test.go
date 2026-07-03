package vugu

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModCheckerWidget(t *testing.T) {

	// try a more complex but realistic example

	assert := assert.New(t)
	mt := NewModTracker()

	type Tag struct {
		Name  string `vugu:"data" json:"name"`
		Style int    `vugu:"data" json:"style"`
	}

	type Widget struct {
		ID           int64   `vugu:"data" json:"id"`
		Name         string  `vugu:"data" json:"name"`
		Description  *string `vugu:"data" json:"description"`
		Active       bool    `vugu:"data" json:"active"`
		StockCount   *int32  `vugu:"data" json:"stock_count"`
		Tags         []Tag   `vugu:"data" json:"tags"`
		FeaturedTags []*Tag  `vugu:"data" json:"-"` // calculated separately after loading
	}

	wjson := `
	[
		{
			"id": 100,
			"name": "Widgetron",
			"description": "This one is electronical",
			"active": true,
			"stock_count": 42,
			"tags": [{"name":"Electronical","style":314159},{"name":"Whatever","style":1}]
		},
		{
			"id": 101,
			"name": "Widgetcron",
			"description": "This one is chronical",
			"active": true,
			"stock_count": 2400,
			"tags": [{"name":"Something","style":2400},{"name":"Whatever","style":1}]
		},
		{
			"id": 102,
			"name": "Widgetevil",
			"description": "This one is diabolical",
			"active": false,
			"stock_count": 13,
			"tags": [{"name":"Something","style":666}]
		}
	]
	`

	var wlist []Widget
	assert.NoError(json.Unmarshal([]byte(wjson), &wlist))

	// log.Printf("wlist: %#v", wlist)

	mt.TrackNext()
	assert.True(mt.ModCheckAll(&wlist))
	assert.True(mt.ModCheckAll(&wlist))

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&wlist))

	// change the sequence
	mt.TrackNext()
	wlist[0], wlist[1] = wlist[1], wlist[0]
	assert.True(mt.ModCheckAll(&wlist))

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&wlist))

	// change a field value
	mt.TrackNext()
	wlist[0].Name = "Widgetcroninator"
	assert.True(mt.ModCheckAll(&wlist))

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&wlist))

	// change a pointer field
	mt.TrackNext()
	newStockCount := int32(2401)
	wlist[0].StockCount = &newStockCount
	assert.True(mt.ModCheckAll(&wlist))

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&wlist))

	// log.Printf("dumpA: %s", mt.dump())

	// change the value pointed to by a field
	mt.TrackNext()
	*wlist[0].StockCount = 2402
	assert.True(mt.ModCheckAll(&wlist))

	// log.Printf("dumpB: %s", mt.dump())

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&wlist))

	// set pointer field to nil
	mt.TrackNext()
	wlist[0].StockCount = nil
	assert.True(mt.ModCheckAll(&wlist))

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&wlist))

	// log.Printf("dump: %s", mt.dump())

	// now try it with a slice of struct pointers
	var wplist []*Widget
	assert.NoError(json.Unmarshal([]byte(wjson), &wplist))

	mt.TrackNext()
	assert.True(mt.ModCheckAll(&wplist))
	assert.True(mt.ModCheckAll(&wplist))

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&wplist))

	// change the sequence
	mt.TrackNext()
	wplist[0], wplist[1] = wplist[1], wplist[0]
	assert.True(mt.ModCheckAll(&wplist))

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&wplist))

	// change a field value
	mt.TrackNext()
	wplist[0].Name = "Widgetcroninator"
	assert.True(mt.ModCheckAll(&wplist))

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&wplist))

	// change a pointer field
	mt.TrackNext()
	newStockCount = int32(2401)
	wplist[0].StockCount = &newStockCount
	assert.True(mt.ModCheckAll(&wplist))

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&wplist))

	// log.Printf("dumpA: %s", mt.dump())

	// change the value pointed to by a field
	mt.TrackNext()
	*wplist[0].StockCount = 2402
	assert.True(mt.ModCheckAll(&wplist))

	// log.Printf("dumpB: %s", mt.dump())

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&wplist))

	// set pointer field to nil
	mt.TrackNext()
	wplist[0].StockCount = nil
	assert.True(mt.ModCheckAll(&wplist))

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&wplist))

	// set one of the struct elements to nil
	mt.TrackNext()
	wplist[0] = nil
	assert.True(mt.ModCheckAll(&wplist))

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&wplist))

}

func TestModCheckerStruct(t *testing.T) {
	assert := assert.New(t)
	mt := NewModTracker()

	var s1 struct {
		F1 string  `vugu:"data"`
		F2 int     `vugu:"data"`
		F3 float64 // not tagged
	}

	mt.TrackNext()
	assert.True(mt.ModCheckAll(&s1))

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&s1))

	mt.TrackNext()
	s1.F1 = "test1"
	assert.True(mt.ModCheckAll(&s1))

	mt.TrackNext()
	s1.F2 = 1
	assert.True(mt.ModCheckAll(&s1))

	mt.TrackNext()
	s1.F3 = 1.0 // field not tagged, should not cause modification
	assert.False(mt.ModCheckAll(&s1))

}

func TestModCheckerSliceArray(t *testing.T) {

	assert := assert.New(t)
	mt := NewModTracker()

	var a1 [3]string
	a1[0] = "s1"
	a1[1] = "s2"
	a1[2] = "s3"

	assert.True(mt.ModCheckAll(&a1))

	// log.Printf("state1: \n%s", mt.dump())

	mt.TrackNext()

	// log.Printf("state2: \n%s", mt.dump())

	assert.False(mt.ModCheckAll(&a1))

	// log.Printf("state3: \n%s", mt.dump())

	var s1 []string
	s1 = a1[:]

	mt.TrackNext()
	assert.True(mt.ModCheckAll(&s1))

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&s1))

	mt.TrackNext()
	s1 = a1[:2]
	assert.True(mt.ModCheckAll(&s1))

	// swap two elements
	mt.TrackNext()
	s1[0], s1[1] = s1[1], s1[0]
	assert.True(mt.ModCheckAll(&s1))

	mt.TrackNext()
	assert.False(mt.ModCheckAll(&s1))

}

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

	// check both
	mt.TrackNext()

	mt.TrackNext()
	b = []byte("testing")
	assert.True(mt.ModCheckAll(&b))
	s = "testing"
	assert.True(mt.ModCheckAll(&s))

	mt.TrackNext()
	b = []byte("testing")
	assert.False(mt.ModCheckAll(&b))
	s = "testing"
	assert.False(mt.ModCheckAll(&s))

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
