package vugu

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompKey(t *testing.T) {

	assert := assert.New(t)

	id := MakeCompKeyIDNowRand()
	log.Printf("id=%#v", id)

	m := make(map[CompKey]bool, 5000)
	for i := 0; i < 5000; i++ {
		ck := CompKey{ID: MakeCompKeyIDNowRand()}
		if m[ck] {
			t.Logf("CompKey %#v found duplicate", ck)
			t.Fail()
		}
		m[ck] = true
	}

	// verify that different IterKey values are distinct
	m = make(map[CompKey]bool)
	ck1 := CompKey{ID: MakeCompKeyIDNowRand(), IterKey: int(123)}
	m[ck1] = true
	assert.False(m[CompKey{ID: ck1.ID, IterKey: nil}])
	assert.False(m[CompKey{ID: ck1.ID, IterKey: int(122)}])
	assert.False(m[CompKey{ID: ck1.ID, IterKey: uint(123)}])
	assert.True(m[CompKey{ID: ck1.ID, IterKey: int(123)}])

}
