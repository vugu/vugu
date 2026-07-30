package vugu

import (
	"log"
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
