package vugu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildEnvCachedComponent(t *testing.T) {

	assert := assert.New(t)

	be, err := NewBuildEnv()
	assert.NoError(err)
	assert.NotNil(be)

	{ // just double check sane behavior for these keys
		k1 := MakeCompKey(1, 1)
		k2 := MakeCompKey(1, 1)
		assert.Equal(k1, k2)
		k3 := MakeCompKey(1, 2)
		assert.NotEqual(k1, k3)
		k4 := MakeCompKey(1, 1)
		assert.Equal(k1, k4)
	}

	rb1 := &rootb1{}

	// first run to intialize
	res := be.RunBuild(rb1)
	assert.NotNil(res)

	c := be.CachedComponent(MakeCompKey(1, 1))
	assert.Nil(c)
	assert.Nil(be.compCache[MakeCompKey(1, 1)])

	b1 := &testb1{}
	be.UseComponent(MakeCompKey(1, 1), b1)
	assert.NotNil(be.compUsed[MakeCompKey(1, 1)])

	// run another one
	res = be.RunBuild(rb1)
	assert.NotNil(res)

	// we should see b1 in the cache
	assert.NotNil(be.compCache[MakeCompKey(1, 1)])
	assert.Equal(b1, be.compCache[MakeCompKey(1, 1)])

	// TODO: but not in the used (not used for this pass)

	// TODO: now try to use it and make sure we can only get it once

}

type rootb1 struct{}

func (b *rootb1) Build(in *BuildIn) (out *BuildOut) {
	return &BuildOut{
		Out: []*VGNode{},
	}
}

type testb1 struct{}

func (b *testb1) Build(in *BuildIn) (out *BuildOut) {
	return &BuildOut{
		Out: []*VGNode{},
	}
}
