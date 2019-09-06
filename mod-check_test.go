package vugu

import (
	"log"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModCheckerChangeCounter(t *testing.T) {

	assert := assert.New(t)

	var wl WidgetList
	wl.SetList([]Widget{
		{ID: 1, Name: "Willy"},
		{ID: 2, Name: "Nilly"},
		{ID: 3, Name: "Silly"},
	})

	var mc ModTracker
	mc.TrackNext()

	changed := mc.ModCheckAll(&wl)
	assert.True(changed)
	log.Printf("changed(1) = %v", changed)

	mc.TrackNext()

	wl.Changed()
	changed = mc.ModCheckAll(&wl)
	assert.True(changed)
	log.Printf("changed(2) = %v", changed)

	mc.TrackNext()

	changed = mc.ModCheckAll(&wl)
	assert.False(changed)
	log.Printf("changed(3) = %v", changed)

	mc.TrackNext()

	changed = mc.ModCheckAll(&wl)
	assert.False(changed)
	log.Printf("changed(4) = %v", changed)

	mc.TrackNext()

	wl.SetList([]Widget{
		{ID: 4, Name: "Billy"},
		{ID: 5, Name: "Lilly"},
		{ID: 6, Name: "Milly"},
	})

	changed = mc.ModCheckAll(&wl)
	assert.True(changed)
	log.Printf("changed(5) = %v", changed)

	mc.TrackNext()

	changed = mc.ModCheckAll(&wl)
	assert.False(changed)
	log.Printf("changed(6) = %v", changed)

}

type Widget struct {
	ID   uint64
	Name string
}

type WidgetList struct {
	items []Widget
	ChangeCounter
}

func (l *WidgetList) SetList(items []Widget) {
	l.items = items
	l.Changed()
}

func (l *WidgetList) Len() int {
	return len(l.items)
}

func (l *WidgetList) Index(idx int) Widget {
	return l.items[idx]
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

}

func TestModCheckerNumbers(t *testing.T) {

	// assert := assert.New(t)

	vbool := true
	vint := 1
	vint8 := 1
	vint16 := 1
	vint32 := 1
	vint64 := 1
	vuint := 1
	vuint8 := 1
	vuint16 := 1
	vuint32 := 1
	vuint64 := 1
	vfloat32 := 1
	vfloat64 := 1
	vcomplex64 := complex(1, 1)
	vcomplex128 := complex(1, 1)

	pointers := []interface{}{
		&vbool, &vint, &vint8, &vint16, &vint32, &vint64, &vuint, &vuint8, &vuint16, &vuint32, &vuint64,
		&vfloat32, &vfloat64, &vcomplex64, &vcomplex128,
	}

	assign := func(values ...interface{}) {
		for i := range pointers {
			p := pointers[i]
			v := values[i]
			log.Printf("GOT HERE: %v %T %#v | %T %#v", i, p, p, v, v)
			reflect.ValueOf(p).Elem().Set(reflect.ValueOf(v))
		}
	}

	var mt ModTracker

	assertMod := func(mods ...bool) {
		for i, p := range pointers {
			mod := mt.ModCheckAll(p)
			if mods[i] != mod {
				t.Errorf("mod for %T (value=%#v) was %v expected %v",
					p, reflect.ValueOf(p).Elem().Interface(), mod, mods[i])
			}
		}
	}

	mt.TrackNext()

	// assign all the same initial values
	assign(
		bool(true),
		int(1),
		int8(1),
		int16(1),
		int32(1),
		int64(1),
		uint(1),
		uint8(1),
		uint16(1),
		uint32(1),
		uint64(1),
		float32(1),
		float64(1),
		complex64(complex(1, 1)),
		complex128(complex(1, 1)),
	)

	assertMod(
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
	)

}
