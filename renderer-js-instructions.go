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

// NOTE: I needed a single concise word which means, essentially "make it so".  The idea being that the element described
// should exist and if it does not update/replace whatever is there so it is.  Unable to suitable term in the English language,
// I've chosen the word "Picard" for this purpose.

const (
	opcodeEnd uint8 = 0 // no more instructions in this buffer
	// opcodeClearRefmap      uint8 = 1 // clear the reference map, all following instructions must not reference prior IDs
	// opcodeClearElStack uint8 = 1 // clear the stack of elements
	opcodeClearEl uint8 = 1 // unset current element
	// opcodeSetHTMLRef       uint8 = 2 // assign ref for html tag
	// opcodeSetHeadRef       uint8 = 3 // assign ref for head tag
	// opcodeSetBodyRef       uint8 = 4 // assign ref for body tag
	// opcodeSelectRef        uint8 = 5 // select element by ref
	opcodeRemoveOtherAttrs uint8 = 5 // remove any elements for the current element that we didn't just set
	opcodeSetAttrStr       uint8 = 6 // assign attribute string to the current selected element
	opcodeSelectMountPoint uint8 = 7 // selects the mount point element and pushes to the stack - the first time by selector but every subsequent time it will reuse the element from before (because the selector may not match after it's been synced over, it's id etc), also make sure it's of this element name and recreate if so
	// opcodePicardFirstChildElement uint8 = 8  // ensure an element first child and select element
	// opcodePicardFirstChildText    uint8 = 9  // ensure a text first child and select element
	// opcodePicardFirstChildComment uint8 = 10 // ensure a comment first child and select element
	opcodeSelectParent     uint8 = 11 // select parent element
	opcodePicardFirstChild uint8 = 12 // ensure an element first child and select element

	opcodeMoveToFirstChild uint8 = 20 // move node selection to first child (doesn't have to exist)
	opcodeSetElement       uint8 = 21 // assign current selected node as an element of the specified type
	// opcodeSetElementAttr      uint8 = 22 // set attribute on current element
	opcodeSetText             uint8 = 23 // assign current selected node as text with specified content
	opcodeSetComment          uint8 = 24 // assign current selected node as comment with specified content
	opcodeMoveToParent        uint8 = 25 // move node selection to parent
	opcodeMoveToNextSibling   uint8 = 26 // move node selection to next sibling (doesn't have to exist)
	opcodeClearEventListeners uint8 = 27 // remove all event listeners from currently selected element
	opcodeSetEventListener    uint8 = 28 // assign event listener to currently selected element

)

// newInstructionList will create a new instance backed by the specified slice and with a clearBufFunc
// that is called when the buffer is about to overflow.
func newInstructionList(buf []byte, flushBufFunc func(il *instructionList) error) *instructionList {
	if buf == nil {
		panic("buf is nil")
	}
	if flushBufFunc == nil {
		panic("flushBufFunc is nil")
	}
	return &instructionList{
		buf:          buf,
		flushBufFunc: flushBufFunc,
	}
}

type instructionList struct {
	buf          []byte
	pos          int
	flushBufFunc func(il *instructionList) error
}

var errDoesNotFit = fmt.Errorf("requested instruction does not fit in the buffer")

func (il *instructionList) flush() error {
	err := il.flushBufFunc(il)
	if err != nil {
		return err
	}
	il.pos = 0
	return nil
}

// checkLenAndFlush calls checkLen(), if it fails attempts to flush the buffer and checkLen again, at which point any error is returned.
func (il *instructionList) checkLenAndFlush(l int) error {

	err := il.checkLen(l)
	if err != nil {

		if err == errDoesNotFit {
			err = il.flush()
			if err != nil {
				return err
			}
			err = il.checkLen(l)
		}
	}

	return err
}

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

// func (il *instructionList) writeClearRefmap() error {

// 	err := il.checkLenAndFlush(1)
// 	if err != nil {
// 		return err
// 	}

// 	il.writeValUint8(opcodeClearRefmap)

// 	return nil
// }

func (il *instructionList) writeClearEl() error {

	err := il.checkLenAndFlush(1)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeClearEl)

	return nil
}

// func (il *instructionList) writeSetHTMLRef(ref uint64) error {

// 	err := il.checkLenAndFlush(9)
// 	if err != nil {
// 		return err
// 	}

// 	il.writeValUint8(opcodeSetHTMLRef)
// 	il.writeValUint64(ref)

// 	return nil
// }

// func (il *instructionList) writeSelectRef(ref uint64) error {

// 	err := il.checkLenAndFlush(9)
// 	if err != nil {
// 		return err
// 	}

// 	il.writeValUint8(opcodeSelectRef)
// 	il.writeValUint64(ref)

// 	return nil
// }

func (il *instructionList) writeRemoveOtherAttrs() error {

	err := il.checkLenAndFlush(1)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeRemoveOtherAttrs)

	return nil
}

func (il *instructionList) writeSetAttrStr(name, value string) error {

	size := len(name) + len(value) + 9

	err := il.checkLenAndFlush(size)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSetAttrStr)
	il.writeValString(name)
	il.writeValString(value)

	return nil
}

func (il *instructionList) writeSelectMountPoint(selector, nodeName string) error {

	err := il.checkLenAndFlush(len(selector) + len(nodeName) + 9)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSelectMountPoint)
	il.writeValString(selector)
	il.writeValString(nodeName)

	return nil

}

func (il *instructionList) writePicardFirstChild(nodeType uint8, data string) error {

	// ensure an element first child and push onto element stack

	err := il.checkLenAndFlush(len(data) + 6)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodePicardFirstChild)
	il.writeValUint8(nodeType)
	il.writeValString(data)

	return nil

}

func (il *instructionList) writeMoveToFirstChild() error {

	err := il.checkLenAndFlush(1)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeMoveToFirstChild)

	return nil
}

func (il *instructionList) writeSetElement(nodeName string) error {

	err := il.checkLenAndFlush(len(nodeName) + 5)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSetElement)
	il.writeValString(nodeName)

	return nil

}

func (il *instructionList) writeSetText(text string) error {

	err := il.checkLenAndFlush(len(text) + 5)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSetText)
	il.writeValString(text)

	return nil

}

func (il *instructionList) writeSetComment(comment string) error {

	err := il.checkLenAndFlush(len(comment) + 5)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSetComment)
	il.writeValString(comment)

	return nil

}

func (il *instructionList) writeMoveToParent() error {

	err := il.checkLenAndFlush(1)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeMoveToParent)

	return nil
}

func (il *instructionList) writeMoveToNextSibling() error {

	err := il.checkLenAndFlush(1)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeMoveToNextSibling)

	return nil
}

// func (il *instructionList) writePicardFirstChildElement(nodeName string) error {

// 	// ensure an element first child and push onto element stack

// 	err := il.checkLenAndFlush(len(nodeName) + 5)
// 	if err != nil {
// 		return err
// 	}

// 	il.writeValUint8(opcodePicardFirstChildElement)
// 	il.writeValString(nodeName)

// 	return nil

// }

// func (il *instructionList) writePicardFirstChildText(text string) error {

// 	// ensure a text first child and push onto element stack

// 	err := il.checkLenAndFlush(len(text) + 5)
// 	if err != nil {
// 		return err
// 	}

// 	il.writeValUint8(opcodePicardFirstChildText)
// 	il.writeValString(text)

// 	return nil
// }

// func (il *instructionList) writePicardFirstChildComment(comment string) error {

// 	// ensure a comment first child and push onto element stack

// 	err := il.checkLenAndFlush(len(comment) + 5)
// 	if err != nil {
// 		return err
// 	}

// 	il.writeValUint8(opcodePicardFirstChildComment)
// 	il.writeValString(comment)

// 	return nil

// }

func (il *instructionList) writeSelectParent() error {

	// pop from the element stack

	err := il.checkLenAndFlush(1)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSelectParent)

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
