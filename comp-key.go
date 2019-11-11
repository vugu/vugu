package vugu

import (
	"crypto/rand"
	"encoding/binary"
	"time"

	"github.com/vugu/xxhash"
)

// CompKey is the key used to identify and look up a component instance.
type CompKey struct {
	ID      uint64      // unique ID for this instance of a component, randomly generated and embeded into source code
	IterKey interface{} // optional iteration key to distinguish the same component reference in source code but different loop iterations
}

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

// MakeCompKeyIDNowRand generates a value for CompKey.ID based on the current unix timestamp in seconds for the top 32 bits and
// the bottom 32 bits populated from crypto/rand
func MakeCompKeyIDNowRand() uint64 {
	var ret = uint64(time.Now().Unix()) << 32
	b := make([]byte, 4)
	rand.Read(b)
	ret |= uint64(binary.BigEndian.Uint32(b))
	return ret
}
