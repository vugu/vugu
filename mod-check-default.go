// +build !tinygo

package vugu

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/vugu/xxhash"
)

// type ModCheckedString string
// func (s ModCheckedString) ModCheck(mt *ModTracker, oldData interface{}) (isModified bool, newData interface{}) {
// }

// ModTracker tracks modifications and maintains the appropriate state for this.
type ModTracker struct {
	old map[interface{}]mtResult
	cur map[interface{}]mtResult
}

// TrackNext moves the "current" information to the "old" position - starting a new round of change tracking.
// Calls to ModCheckAll after calling TrackNext will compare current values to the values from before the call to TrackNext.
// It is generally called once at the start of each build and render cycle.  This method must be called at least
// once before doing modification checks.
func (mt *ModTracker) TrackNext() {

	// lazy initialize
	if mt.old == nil {
		mt.old = make(map[interface{}]mtResult)
	}
	if mt.cur == nil {
		mt.cur = make(map[interface{}]mtResult)
	}

	// remove all elements from old map
	for k := range mt.old {
		delete(mt.old, k)
	}

	// and swap them
	mt.old, mt.cur = mt.cur, mt.old

}

func (mt *ModTracker) dump() []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "-- cur (len=%d): --\n", len(mt.cur))
	for k, v := range mt.cur {
		fmt.Fprintf(&buf, " %#v = %#v\n", k, v)
	}
	fmt.Fprintf(&buf, "-- old (len=%d): --\n", len(mt.old))
	for k, v := range mt.old {
		fmt.Fprintf(&buf, " %#v = %#v\n", k, v)
	}
	return buf.Bytes()
}

// Otherwise pointers to some built-in types are supported including all primitive single-value types -
// bool, int/uint and all variations, both float types, both complex types, string.
// Pointers to supported types are supported.
// Arrays, slices and maps(nope!) using supported types are supported.
// Pointers to structs will be checked by
// checking each field with a struct tag like `vugu:"modcheck"`.  Slices and arrays of
// structs are okay, since their members have a stable position in memory and a pointer
// can be taken.  Maps using structs however must use pointers to them
// (restriction applies to both keys and values) to be supported.

// ModCheckAll performs a modification check on the values provided.
// For values implementing the ModChecker interface, the ModCheck method will be called.
// All values passed should be pointers to the types described below.
// Single-value primitive types are supported.  Structs are supported and
// and are traversed by calling ModCheckAll on each with the tag `vugu:"modcheck"`.
// Arrays and slices of supported types are supported, their length is compared as well
// as a pointer to each member.
// As a special case []byte is treated like a string.
// Maps are not supported at this time.
// Other weird and wonderful things like channels and funcs are not supported.
// Passing an unsupported type will result in a panic.
func (mt *ModTracker) ModCheckAll(values ...interface{}) (ret bool) {

	for _, v := range values {

		// check if we've already done a mod check on v
		curres, ok := mt.cur[v]
		if ok {
			ret = ret || curres.modified
			continue
		}

		// look up the old data
		oldres := mt.old[v]

		// the result of the mod check on v goes here
		var mod bool
		var newdata interface{}

		{
			// see if it implements the ModChecker interface
			mc, ok := v.(ModChecker)
			if ok {
				mod, newdata = mc.ModCheck(mt, oldres.data)
				goto handleData
			}

			// support for certain built-in types
			switch vt := v.(type) {

			case *string:
				oldval, ok := oldres.data.(string)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			case *[]byte: // special case of []byte, handled like string not slice
				oldval, ok := oldres.data.(string)
				vts := string(*vt)
				mod = !ok || oldval != vts
				newdata = vts
				goto handleData

			case *bool:
				oldval, ok := oldres.data.(bool)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			case *int:
				oldval, ok := oldres.data.(int)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			case *int8:
				oldval, ok := oldres.data.(int8)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			case *int16:
				oldval, ok := oldres.data.(int16)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			case *int32:
				oldval, ok := oldres.data.(int32)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			case *int64:
				oldval, ok := oldres.data.(int64)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			case *uint:
				oldval, ok := oldres.data.(uint)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			case *uint8:
				oldval, ok := oldres.data.(uint8)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			case *uint16:
				oldval, ok := oldres.data.(uint16)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			case *uint32:
				oldval, ok := oldres.data.(uint32)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			case *uint64:
				oldval, ok := oldres.data.(uint64)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			case *float32:
				oldval, ok := oldres.data.(float32)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			case *float64:
				oldval, ok := oldres.data.(float64)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			case *complex64:
				oldval, ok := oldres.data.(complex64)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			case *complex128:
				oldval, ok := oldres.data.(complex128)
				mod = !ok || oldval != *vt
				newdata = *vt
				goto handleData

			}

			// when the scalpel (type switch) doesn't do it,
			// gotta use the bonesaw (reflection)

			rv := reflect.ValueOf(v)

			// check pointer and deref
			if rv.Kind() != reflect.Ptr {
				panic(errors.New("type not implemented (pointer required): " + rv.String()))
			}
			rvv := rv.Elem()

			// slice and array are treated the same
			if rvv.Kind() == reflect.Slice || rvv.Kind() == reflect.Array {

				l := rvv.Len()

				// for slices and arrays we compute the hash of the raw contents,
				// takes care of length and sequence changes
				ha := xxhash.New()

				var el0t reflect.Type

				if l > 0 {
					// use the unsafe package to make a byte slice corresponding to the raw slice contents
					var bs []byte
					bsh := (*reflect.SliceHeader)(unsafe.Pointer(&bs))

					el0 := rvv.Index(0)
					el0t = el0.Type()

					bsh.Data = el0.Addr().Pointer() // point to first element of slice
					bsh.Len = l * int(el0t.Size())
					bsh.Cap = bsh.Len

					// hash it
					ha.Write(bs)
				}

				hashval := ha.Sum64()

				// use hashval as our data, and mark as modified if different
				oldval, ok := oldres.data.(uint64)
				mod = !ok || oldval != hashval
				newdata = hashval

				// for types that by definition have already been checked with the hash above, we're done
				if el0t != nil {
					switch el0t.Kind() {
					case reflect.Bool,
						reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
						reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
						reflect.Uintptr,
						reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
						reflect.String:
						goto handleData
					}
				}

				// recurse into each element and check, update mod as we go
				// NOTE: for "deep" element types that can have changes within them,
				// it's important to recurse into children even if mod is already
				// true, otherwise we'll never call ModCheckAll on these children and
				// never get an unmodified response

				for i := 0; i < l; i++ {

					// get pointer to the individual element and recurse into it
					elv := rvv.Index(i).Addr().Interface()
					mod = mt.ModCheckAll(elv) || mod
				}

				goto handleData

			}

			// for structs we iterate over the fields looked for tagged ones
			if rvv.Kind() == reflect.Struct {

				// just use bool(true) as the data value for a struct - this way it will always be modified the first time,
				// but every subsequent check will solely depend on the checks on it's fields
				oldval, ok := oldres.data.(bool)
				mod = !ok || oldval != true
				newdata = true

				rvvt := rvv.Type()
				for i := 0; i < rvvt.NumField(); i++ {
					// skip untagged fields
					if !hasTagPart(rvvt.Field(i).Tag.Get("vugu"), "data") {
						continue
					}

					// call ModCheckAll on pointer to field
					mod = mt.ModCheckAll(
						rvv.Field(i).Addr().Interface(),
					) || mod

				}

				goto handleData
			}

			// random stream of conciousness: should we check for a ModCheck implementation here and call it?
			// We might want to follow these pointers or rather recurse into them (if not nil?)
			// with a ModCheckAll.  Also look at how this ties into maps (if at all), since theoretically
			// we could compare a map by storing its length and basically doing for k, v := range m { ...ModCheckAll(&k,&v)... }
			// actually no that won't work because k and v will have different locations each time - but still could
			// potentially make some mapKeyValue struct that does what we need - but not vital to solve right now.
			// Pointers to pointers will probably come up first, and is likely why you are reading this comment.

			// pointer (meaning we were originally passed a pointer to a pointer)
			if rvv.Kind() == reflect.Ptr {

				// use pointer value as data...
				vv := rvv.Pointer()
				oldval, ok := oldres.data.(uintptr)
				mod = !ok || oldval != vv
				newdata = vv

				// ...but also recurse and call ModChecker with one level of pointer dereferencing, if not nil
				if !rvv.IsNil() {
					mod = mt.ModCheckAll(
						rvv.Interface(),
					) || mod
				}

				goto handleData
			}

			panic(errors.New("type not implemented: " + reflect.TypeOf(v).String()))

		}
	handleData:
		mt.cur[v] = mtResult{modified: mod, data: newdata}
		ret = ret || mod

	}

	return ret
}

// ModCheck(oldData interface{}) (isModified bool, newData interface{})

// hm, this may not work - what happens if we call ModCheck twice in a row?? we need a clear
// way to demark the "old" and "new" versions of data - it might be that the comparison and
// the gather of the new value need to be separate methods?  Or can ModTracker somehow
// prevent ModCheck from being call twice in the same pass (or prevent that from being an issue)

// actually that might work - if there is a call on ModTracker that moves "new" to "old", it
// woudl be pretty clear - and then calling this would only update the "new" value.  Also the
// presence of something in the "new" data would mean it's already been called in this pass
// and so can be deduplicated.

func hasTagPart(tagstr, part string) bool {
	for _, p := range strings.Split(tagstr, ",") {
		if p == part {
			return true
		}
	}
	return false
}
