package vugu

import (
	"math/rand"
	"time"

	"github.com/vugu/xxhash"
)

// MakeCompKeyID forms a value for CompKey.ID from the given time the uint32 you provide for the lower 32 bits.
func MakeCompKeyID(t time.Time, data uint32) uint64 {
	var ret = uint64(t.Unix()) << 32
	ret |= uint64(data)
	return ret
}

// MakeCompKeyIDTimeHash forms a value for CompKey.ID from the given time and a hash of the bytes you provide.
func MakeCompKeyIDTimeHash(t time.Time, b []byte) uint64 {
	h := xxhash.New()
	h.Write(b)
	return MakeCompKeyID(t, uint32(h.Sum64()))
}

var compKeyRand *rand.Rand

// MakeCompKeyIDNowRand generates a value for CompKey.ID based on the current unix timestamp in seconds for the top 32 bits and
// the bottom 32 bits populated from a random source
func MakeCompKeyIDNowRand() uint64 {
	if compKeyRand == nil {
		compKeyRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	var ret = uint64(time.Now().Unix()) << 32
	ret |= uint64(compKeyRand.Int63() & 0xFFFFFFFF)
	return ret
}
