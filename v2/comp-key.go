package vugu

import (
	"math/rand/v2"
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
	_, err := h.Write(b)
	if err != nil {
		panic(err)
	}
	return MakeCompKeyID(t, uint32(h.Sum64()))
}

var compKeyRand *rand.Rand

// MakeCompKeyIDNowRand generates a value for CompKey.ID using the newer v2 version of the math/rand package
func MakeCompKeyIDNowRand() uint64 {
	if compKeyRand == nil {
		compKeyRand = rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano())))
	}
	return compKeyRand.Uint64()
}
