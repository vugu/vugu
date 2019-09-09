package vugu

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// CompKey is the key used to identify and look up a component instance.
type CompKey struct {
	ID      uint64      // unique ID for this instance of a component, randomly generated and embeded into source code
	IterKey interface{} // optional iteration key to distinguish the same component reference in source code but different loop iterations
}

// MakeCompKeyID generates a value for CompKey.ID based on the current unix timestamp in seconds for the top 32 bits and
// the bottom 32 bits populated from crypto/rand
func MakeCompKeyID() uint64 {
	var ret = uint64(time.Now().Unix()) << 32
	b := make([]byte, 4)
	rand.Read(b)
	ret |= uint64(binary.BigEndian.Uint32(b))
	return ret
}
