package vugu

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"sort"

	"github.com/cespare/xxhash"
)

type DataHasher interface {
	DataHash() uint64
}

func ComputeHash(i interface{}) uint64 {

	// FIXME: do we need to handle circular references here? (via pointer, via interface{}),
	// right now this will recurse forever and eventually stack overflow, not sure if this is
	// real world case or not.  Not sure if we should panic or gracefully handle it or what.

	if _, ok := i.(*reflect.Value); ok {
		panic("cannot compute hash on a *reflect.Value")
	}
	if _, ok := i.(reflect.Value); ok {
		panic("cannot compute hash on a reflect.Value")
	}

	if dh, ok := i.(DataHasher); ok {
		return dh.DataHash()
	}

	b8 := make([]byte, 8)

	// find dereferenced value
	v := reflect.ValueOf(i)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			// return xxhash.Sum64(b8)
			return 0
		}
		v = v.Elem()
	}

	// log.Printf("v = %#v", v)

	switch v.Kind() {

	case reflect.Bool:
		if v.Bool() {
			b8[0] = 1
		}
		return xxhash.Sum64(b8)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		binary.BigEndian.PutUint64(b8, uint64(v.Int()))
		return xxhash.Sum64(b8)

	case reflect.Float32, reflect.Float64:
		binary.BigEndian.PutUint64(b8, math.Float64bits(v.Float()))
		return xxhash.Sum64(b8)

	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		h := xxhash.New()
		binary.BigEndian.PutUint64(b8, math.Float64bits(real(c)))
		h.Write(b8)
		binary.BigEndian.PutUint64(b8, math.Float64bits(imag(c)))
		h.Write(b8)
		return h.Sum64()

	case reflect.Func, reflect.Chan, reflect.UnsafePointer:
		p := v.Pointer()
		binary.BigEndian.PutUint64(b8, uint64(p))
		return xxhash.Sum64(b8)

	case reflect.Interface:
		d := v.InterfaceData()
		h := xxhash.New()
		binary.BigEndian.PutUint64(b8, uint64(d[0]))
		h.Write(b8)
		binary.BigEndian.PutUint64(b8, uint64(d[1]))
		h.Write(b8)
		return h.Sum64()

	case reflect.Array, reflect.Slice:
		h := xxhash.New()
		l := v.Len()
		for i := 0; i < l; i++ {
			binary.BigEndian.PutUint64(b8, ComputeHash(v.Index(i).Interface()))
			h.Write(b8)
		}
		return h.Sum64()

	case reflect.Map:
		h := xxhash.New()

		// we have to do a little dance here to ensure the sequence is stable,
		// because MapKeys() is not - both per the doc and in practice

		mk := v.MapKeys()

		khmap := make(map[uint64]int, len(mk)) // index_into_mk -> hash
		for i, k := range mk {
			// NOTE: collisions will cause a key to be skipped - not great,
			// but possibly better than hashes being unstable for the same input.
			// (i.e. I'd prefer an obscure bug where the UI didn't update one time, as opposed to
			// one where the CPU is on fire because it's updating all the time)
			khmap[ComputeHash(k.Interface())] = i
		}

		hs := make([]uint64, 0, len(khmap))
		for h := range khmap {
			hs = append(hs, h)
		}
		sort.Slice(hs, func(i, j int) bool {
			return hs[i] < hs[j]
		})

		// log.Printf("START1")

		// now that we have a stable sequence:
		for _, i := range hs {

			// write key hash
			binary.BigEndian.PutUint64(b8, i)
			h.Write(b8)

			// compute and write value hash

			mapKey := mk[khmap[i]]
			mapValue := v.MapIndex(mapKey)

			vhash := ComputeHash(mapValue.Interface())
			binary.BigEndian.PutUint64(b8, vhash)
			h.Write(b8)

			// log.Printf("map: i_hash=%v, v_hash=%v", i, vhash)
		}

		// for _, k := range mk {
		// 	binary.BigEndian.PutUint64(b8, ComputeHash(k.Interface()))
		// 	h.Write(b8)
		// 	val := v.MapIndex(k)
		// 	binary.BigEndian.PutUint64(b8, ComputeHash(val.Interface()))
		// 	h.Write(b8)
		// }

		return h.Sum64()

	case reflect.Struct:
		h := xxhash.New()
		l := v.NumField()
		t := v.Type()
		for i := 0; i < l; i++ {
			if t.Field(i).PkgPath != "" {
				continue // skip unexported fields
			}
			binary.BigEndian.PutUint64(b8, ComputeHash(v.Field(i).Interface()))
			h.Write(b8)
		}
		return h.Sum64()

	case reflect.String:
		return xxhash.Sum64String(v.String())

	case reflect.Ptr:
		panic(fmt.Errorf("should not have gotten a pointer here"))

	default:
		// panic(fmt.Errorf("unknown kind %v", v.Kind()))
		// nop
		return 0
	}

	panic("unreachable code")
}
