package vugu

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeHash(t *testing.T) {

	assert := assert.New(t)

	data := struct {
		FString    string
		FInt       int
		FFloat     float64
		FMap       map[string]bool
		FSlice     []string
		FNil       interface{}
		FStruct    struct{ Inner1 string }
		FStructPtr *bytes.Buffer
		unexported bool
	}{
		FString:    "string1",
		FInt:       10,
		FFloat:     10.0,
		FMap:       map[string]bool{"key1": true, "key2": false},
		FSlice:     []string{"Larry", "Moe", "Curly"},
		unexported: true,
	}

	// log.Printf("1-----")
	lasth := ComputeHash(&data)
	// log.Printf("2-----")
	assert.Equal(lasth, ComputeHash(&data))
	// log.Printf("3-----")
	lasth = ComputeHash(data)
	assert.Equal(lasth, ComputeHash(data))
	data.FString = "string2"
	assert.NotEqual(lasth, ComputeHash(data))
	data.FString = "string1"
	assert.Equal(lasth, ComputeHash(data))
	data.FMap = nil
	assert.NotEqual(lasth, ComputeHash(data))
	lasth = ComputeHash(data)
	data.unexported = false
	assert.Equal(lasth, ComputeHash(data))
	data.FStruct.Inner1 = "someval"
	assert.NotEqual(lasth, ComputeHash(data))
	lasth = ComputeHash(data)
	data.FStructPtr = &bytes.Buffer{}
	assert.NotEqual(lasth, ComputeHash(data))
	lasth = ComputeHash(data)
	data.FNil = "not nil any more"
	assert.NotEqual(lasth, ComputeHash(data))

	// log.Printf("HERE1")
	data.FMap = map[string]bool{"key1": true, "key2": false}
	lasth = ComputeHash(data)
	data.FMap["key2"] = true
	assert.NotEqual(lasth, ComputeHash(data))
	data.FMap["key2"] = false
	assert.Equal(lasth, ComputeHash(data))
	data.FMap["key3"] = true
	assert.NotEqual(lasth, ComputeHash(data))
	delete(data.FMap, "key3")
	assert.Equal(lasth, ComputeHash(data))

}
