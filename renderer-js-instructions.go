package vugu

import (
	"encoding/binary"
	"fmt"
)

// NOTE: I looked at using Protobuf for this, and in some ways it makes sense.  The main issue though is that it brings in
// a bunch of code I don't necessarily need, particularly on the JS side.  There are only a few data types that need to be
// encoded and it's not too big a deal.  Whereas with protobuf I immediately bring in 250k of protobuf JS code, the vast
// majority of which is not needed.  So I'm proceeding with the idea that the encoding/decoding is simple enough to just do
// by hand.  I hope I'm right.  -bgp

// NOTE: prefix "ri" is for "render instruction"

const (
	opcodeEnd         uint8 = 0 // no more instructions in this buffer
	opcodeClearRefmap uint8 = 1 // clear the reference map, all following instructions must not reference prior IDs
	opcodeSetHTMLRef  uint8 = 2 // assign ref for html tag
	opcodeSetHeadRef  uint8 = 3 // assign ref for head tag
	opcodeSetBodyRef  uint8 = 4 // assign ref for body tag
	opcodeSelectRef   uint8 = 5 // select element by ref
	opcodeSetAttrStr  uint8 = 6 // assign attribute string to the current selected element
)

// type riOneRef struct {
// 	opcode uint8
// 	ref    riUint64
// }

// var riOneRefSize = int(unsafe.Sizeof(riOneRef{}))

// func riOneRefAt(ptr *byte) *riOneRef {
// 	return (*riOneRef)(unsafe.Pointer(ptr))
// }

// // riUint64 is a 64-bit big-endian unsigned int.
// type riUint64 [8]byte

// // func (r *riUint64) newRiUint64(v uint64) riUint64 {
// // 	var ret riUint64
// // 	ret.Set(v)
// // 	return ret
// // }

// func (r *riUint64) Set(v uint64) {
// 	binary.BigEndian.PutUint64((*r)[:], v)
// }

// func (r *riUint64) Get() uint64 {
// 	return binary.BigEndian.Uint64((*r)[:])
// }

// type riUint32 [4]byte

// type riString struct {
// 	strlen  riUint32
// 	strdata byte
// }

func newInstructionList(buf []byte) *instructionList {
	return &instructionList{
		buf: buf,
	}
}

type instructionList struct {
	buf []byte
	pos int
}

var errDoesNotFit = fmt.Errorf("requested instruction does not fit in the buffer")

func (il *instructionList) checkLen(l int) error {
	if il.pos+l >= len(il.buf)-1 {
		return errDoesNotFit
	}
	return nil
}

func (il *instructionList) writeEnd() {
	il.buf[il.pos] = opcodeEnd
	il.pos++
}

func (il *instructionList) writeClearRefmap() error {

	err := il.checkLen(1)
	if err != nil {
		return err
	}
	// if il.pos+1 >= len(il.buf)-1 {
	// 	return errDoesNotFit
	// }

	il.writeValUint8(opcodeClearRefmap)

	// il.buf[il.pos] = opcodeClearRefmap
	// il.pos++

	return nil
}

func (il *instructionList) writeSetHTMLRef(ref uint64) error {

	err := il.checkLen(9)
	if err != nil {
		return err
	}

	// if il.pos+9 >= len(il.buf)-1 {
	// 	return errDoesNotFit
	// }

	il.writeValUint8(opcodeSetHTMLRef)
	il.writeValUint64(ref)

	// data := riOneRefAt(&il.buf[il.pos])
	// data.opcode = opcodeSetHTMLRef
	// data.ref.Set(ref)
	// il.pos += riOneRefSize

	return nil
}

func (il *instructionList) writeSelectRef(ref uint64) error {

	err := il.checkLen(9)
	if err != nil {
		return err
	}

	// if il.pos+9 >= len(il.buf)-1 {
	// 	return errDoesNotFit
	// }

	il.writeValUint8(opcodeSelectRef)
	il.writeValUint64(ref)

	// data := riOneRefAt(&il.buf[il.pos])
	// data.opcode = opcodeSelectRef
	// data.ref.Set(ref)
	// il.pos += riOneRefSize

	return nil
}

func (il *instructionList) writeSetAttrStr(name, value string) error {

	size := len(name) + len(value) + 9

	err := il.checkLen(size)
	if err != nil {
		return err
	}

	// if il.pos+size >= len(il.buf)-1 {
	// 	return errDoesNotFit
	// }

	il.writeValUint8(opcodeSetAttrStr)
	il.writeValString(name)
	il.writeValString(value)

	// il.buf[il.pos] = opcodeSetAttrStr
	// pos := il.pos + 1
	// pos = riWriteString(il.buf, pos, name)
	// pos = riWriteString(il.buf, pos, value)
	// il.pos = pos

	// namelen := len(name)
	// nameoffset := il.pos + 5
	// valuelen := len(name)
	// valueoffset := nameoffset + namelen + 5

	// il.buf[il.pos] = opcodeSetAttrStr
	// binary.BigEndian.PutUint32(il.buf[il.pos+1:il.pos+5], uint32(namelen))
	// copy()
	// binary.BigEndian.PutUint32(il.buf[il.pos+5+namelen:il.pos+9+namelen], uint32(valuelen))

	// data := riOneRefAt(&il.buf[il.pos])
	// data.opcode = opcodeSetAttrStr
	// data.ref.Set(ref)
	// il.pos += riOneRefSize

	return nil
}

func (il *instructionList) writeValUint8(b uint8) {
	il.buf[il.pos] = b
	il.pos++
}

func (il *instructionList) writeValUint64(ref uint64) {
	binary.BigEndian.PutUint64(il.buf[il.pos:il.pos+8], ref)
	il.pos += 8
}

func (il *instructionList) writeValString(s string) {

	lenstr := len(s)
	pos := il.pos

	// write length as uint32
	binary.BigEndian.PutUint32(il.buf[pos:pos+4], uint32(lenstr))

	// copy bytes directly from string into buf
	copy(il.buf[pos+4:pos+4+lenstr], s)

	il.pos = pos + 4 + lenstr
}

// func riWriteString(buf []byte, pos int, str string) int {

// 	lenstr := len(str)

// 	// write length as uint32
// 	binary.BigEndian.PutUint32(buf[pos:pos+4], uint32(lenstr))

// 	// copy bytes directly from string into buf
// 	copy(buf[pos+4:pos+4+lenstr], str)

// 	// return new position
// 	return pos + 4 + lenstr
// }
