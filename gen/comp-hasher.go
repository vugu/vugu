package gen

import (
	"encoding/binary"
	"sync"

	"github.com/vugu/xxhash"
)

var compHashCounts = make(map[string]int, 8)
var compHashMU sync.Mutex

// compHashCounted returns a hash value for the given string and
// maintaining a count of the number of times that string was passed.
// Each time a given s is provided a different hash will be returned,
// but invocations from one program run to the next should return
// the same series.  (This avoids unnecessary changes in generated files
// when produced with the same input.)
func compHashCounted(s string) uint64 {
	compHashMU.Lock()
	c := compHashCounts[s]
	compHashCounts[s] = c + 1
	compHashMU.Unlock()

	h := xxhash.New()
	h.WriteString(s)
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(c))
	h.Write(b[:])
	return h.Sum64()

}
