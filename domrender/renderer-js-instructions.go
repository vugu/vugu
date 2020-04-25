package domrender

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

// NOTE: I looked at using Protobuf for this, and in some ways it makes sense.  The main issue though is that it brings in
// a bunch of code I don't necessarily need, particularly on the JS side.  There are only a few data types that need to be
// encoded and it's not too big a deal.  Whereas with protobuf I immediately bring in 250k of protobuf JS code, the vast
// majority of which is not needed.  So I'm proceeding with the idea that the encoding/decoding is simple enough to just do
// by hand.  I hope I'm right.  -bgp

// NOTE: I needed a single concise word which means, essentially "make it so".  The idea being that the element described
// should exist, and if it does not, update/replace whatever is there so it is.  Unable to find a suitable term in the
// English language, I've chosen the word "Picard" for this purpose.  UPDATE: Alas, this didn't pan out, but it was worth
// a try ;)

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
	// opcodeSelectParent     uint8 = 11 // select parent element
	// opcodePicardFirstChild uint8 = 12 // ensure an element first child and select element

	opcodeMoveToFirstChild uint8 = 20 // move node selection to first child (doesn't have to exist)
	opcodeSetElement       uint8 = 21 // assign current selected node as an element of the specified type
	// opcodeSetElementAttr      uint8 = 22 // set attribute on current element
	opcodeSetText                   uint8 = 23 // assign current selected node as text with specified content
	opcodeSetComment                uint8 = 24 // assign current selected node as comment with specified content
	opcodeMoveToParent              uint8 = 25 // move node selection to parent
	opcodeMoveToNextSibling         uint8 = 26 // move node selection to next sibling (doesn't have to exist)
	opcodeRemoveOtherEventListeners uint8 = 27 // remove all event listeners from currently selected element that were not just set
	opcodeSetEventListener          uint8 = 28 // assign event listener to currently selected element
	opcodeSetInnerHTML              uint8 = 29 // set the innerHTML for an element

	opcodeSetCSSTag          uint8 = 30 // write a CSS (style or link) tag
	opcodeRemoveOtherCSSTags uint8 = 31 // remove any CSS tags that have not been written since the last call
	opcodeSetJSTag           uint8 = 32 // write a JS (script) tag
	opcodeRemoveOtherJSTags  uint8 = 33 // remove any JS tags that have not been written since the last call

	opcodeSetProperty     uint8 = 35 // assign a JS property to the current element
	opcodeSelectQuery     uint8 = 36 // select an element
	opcodeBufferInnerHTML uint8 = 37 // pass chunked text to set as inner html, complete with opcodeSetInnerHTML

	opcodeSetAttrNSStr uint8 = 38 // assign attribute string to the current selected namespaced element
	opcodeSetElementNS uint8 = 39 // assign current selected node as an element of the specified type in the specified namespace

	opcodeCallback            uint8 = 40 // issue callback, sends just callbackID
	opcodeCallbackLastElement uint8 = 41 // issue callback with callbackID and most recent element reference

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
	logWriter    io.Writer // set to non-nil to enable debug log output
}

var errDoesNotFit = errors.New("requested instruction does not fit in the buffer")

func (il *instructionList) logf(f string, args ...interface{}) error {
	if il.logWriter == nil {
		return nil
	}
	if !strings.HasSuffix(f, "\n") {
		f += "\n"
	}
	_, err := fmt.Fprintf(il.logWriter, "domrender ildebug: "+f, args...)
	return err
}

func (il *instructionList) flush() error {
	il.logf("flush() calling flushBufFunc")
	err := il.flushBufFunc(il)
	if err != nil {
		return err
	}
	il.pos = 0
	il.logf("flush() completed")
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
	if il.pos+l > len(il.buf)-1 {
		return errDoesNotFit
	}
	return nil
}

func (il *instructionList) writeEnd() {
	il.logf("writeEnd[%d]()", opcodeEnd)
	il.buf[il.pos] = opcodeEnd
	il.pos++
}

func (il *instructionList) writeClearEl() error {
	il.logf("writeClearEl[%d]()", opcodeClearEl)

	err := il.checkLenAndFlush(1)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeClearEl)

	return nil
}

func (il *instructionList) writeRemoveOtherAttrs() error {

	il.logf("writeRemoveOtherAttrs[%d]()", opcodeRemoveOtherAttrs)

	err := il.checkLenAndFlush(1)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeRemoveOtherAttrs)

	return nil
}

func (il *instructionList) writeSetAttrStr(name, value string) error {

	il.logf("writeSetAttrStr[%d](name=%q, value=%q)", opcodeSetAttrStr, name, value)

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

func (il *instructionList) writeSetAttrNSStr(namespace, name, value string) error {

	il.logf("writeSetAttrNSStr[%d](ns=%q, name=%q, value=%q)", opcodeSetAttrNSStr, namespace, name, value)

	size := len(namespace) + len(name) + len(value) + 9

	err := il.checkLenAndFlush(size)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSetAttrNSStr)
	il.writeValString(namespace)
	il.writeValString(name)
	il.writeValString(value)

	return nil
}

func (il *instructionList) writeSelectQuery(selector string) error {

	il.logf("writeSelectQuery[%d](selector=%q)", opcodeSelectQuery, selector)

	err := il.checkLenAndFlush(5 + len(selector))
	if err != nil {
		return err
	}
	il.writeValUint8(opcodeSelectQuery)
	il.writeValString(selector)
	return nil
}

func (il *instructionList) writeSelectMountPoint(selector, nodeName string) error {

	il.logf("writeSelectMountPoint[%d](selector=%q, nodeName=%q)", opcodeSelectMountPoint, selector, nodeName)

	err := il.checkLenAndFlush(len(selector) + len(nodeName) + 9)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSelectMountPoint)
	il.writeValString(selector)
	il.writeValString(nodeName)

	return nil

}

func (il *instructionList) writeMoveToFirstChild() error {

	il.logf("writeMoveToFirstChild[%d]()", opcodeMoveToFirstChild)

	err := il.checkLenAndFlush(1)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeMoveToFirstChild)

	return nil
}

func (il *instructionList) writeSetElement(nodeName string) error {

	il.logf("writeSetElement[%d](nodeName=%q)", opcodeSetElement, nodeName)

	err := il.checkLenAndFlush(len(nodeName) + 5)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSetElement)
	il.writeValString(nodeName)

	return nil

}

func (il *instructionList) writeSetElementNS(nodeName, namespace string) error {

	il.logf("writeSetElementNS[%d](nodeName=%q, ns=%q)", opcodeSetElementNS, nodeName, namespace)

	size := len(nodeName) + len(namespace) + 9
	err := il.checkLenAndFlush(size)

	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSetElementNS)
	il.writeValString(nodeName)
	il.writeValString(namespace)

	return nil

}

func (il *instructionList) writeSetText(text string) error {

	il.logf("writeSetText[%d](text=%q)", opcodeSetText, text)

	err := il.checkLenAndFlush(len(text) + 5)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSetText)
	il.writeValString(text)

	return nil

}

func (il *instructionList) writeSetComment(comment string) error {

	il.logf("writeSetComment[%d](comment=%q)", opcodeSetComment, comment)

	err := il.checkLenAndFlush(len(comment) + 5)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSetComment)
	il.writeValString(comment)

	return nil

}

func (il *instructionList) writeMoveToParent() error {

	il.logf("writeMoveToParent[%d]()", opcodeMoveToParent)

	err := il.checkLenAndFlush(1)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeMoveToParent)

	return nil
}

func (il *instructionList) writeMoveToNextSibling() error {

	il.logf("writeMoveToNextSibling[%d]()", opcodeMoveToNextSibling)

	err := il.checkLenAndFlush(1)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeMoveToNextSibling)

	return nil
}

func (il *instructionList) writeSetInnerHTML(html string) error {

	il.logf("writeSetInnerHTML[%d](html=%q)", opcodeSetInnerHTML, html)

	// Make sure there is room to write at least one byte
	// (1 byte for opcode, 4 bytes for string length, 1 byte of data)
	// [This further ensures that maxLen - il.pos > 0]
	err := il.checkLenAndFlush(6)
	if err != nil {
		return err
	}

	remaining := html
	maxLen := len(il.buf) - 6
	for len(remaining) > maxLen-il.pos {
		chunk := remaining[:maxLen-il.pos]
		remaining = remaining[maxLen-il.pos:]
		err := il.checkLenAndFlush(len(chunk) + 5)
		if err != nil {
			return err
		}

		il.writeValUint8(opcodeBufferInnerHTML)
		il.writeValString(chunk)
		il.flush()
	}

	err = il.checkLenAndFlush(len(remaining) + 5)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSetInnerHTML)
	il.writeValString(remaining)

	return nil
}

func (il *instructionList) writeSetEventListener(positionID []byte, eventType string, capture, passive bool) error {

	il.logf("writeSetInnerHTML[%d](positionID=%q, eventType=%q, capture=%v, passive=%v)", opcodeSetEventListener, positionID, eventType, capture, passive)

	err := il.checkLenAndFlush(len(positionID) + len(eventType) + 11)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSetEventListener)
	il.writeValBytes(positionID)
	il.writeValString(eventType)

	captureB := uint8(0)
	if capture {
		captureB = 1
	}
	il.writeValUint8(captureB)

	passiveB := uint8(0)
	if passive {
		passiveB = 1
	}
	il.writeValUint8(passiveB)

	return nil

}

func (il *instructionList) writeRemoveOtherEventListeners(positionID []byte) error {

	il.logf("writeRemoveOtherEventListeners[%d](positionID=%q)", opcodeRemoveOtherEventListeners, positionID)

	err := il.checkLenAndFlush(5 + len(positionID))
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeRemoveOtherEventListeners)
	il.writeValBytes(positionID)

	return nil

}

func (il *instructionList) writeSetCSSTag(elementName string, textContent []byte, attrPairs []string) error {

	il.logf("writeSetCSSTag[%d](elementName=%q, textContext=%q, attrPairs=%#v)", opcodeSetCSSTag, elementName, textContent, attrPairs)

	if len(attrPairs) > 254 {
		return fmt.Errorf("attrPairs is %d, too large, max is 254", len(attrPairs))
	}

	var al = 0
	for _, s := range attrPairs {
		al += len(s) + 4
	}

	var l = 1 + // opcode
		al + // attrs
		// 8 + // hashCode
		1 + // 1 byte for number of strings to read
		len(elementName) + 4 +
		len(textContent) + 4

	err := il.checkLenAndFlush(l)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSetCSSTag)
	// il.writeValUint64(hashCode)
	il.writeValString(elementName)
	il.writeValBytes(textContent)
	il.writeValUint8(uint8(len(attrPairs)))
	for _, s := range attrPairs {
		il.writeValString(s)
	}

	return nil

}

func (il *instructionList) writeRemoveOtherCSSTags() error {

	il.logf("writeRemoveOtherCSSTags[%d]()", opcodeRemoveOtherCSSTags)

	err := il.checkLenAndFlush(1)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeRemoveOtherCSSTags)

	return nil
}

func (il *instructionList) writeSetProperty(key string, jsonValue []byte) error {

	il.logf("writeSetProperty[%d](key=%q, jsonValue=%q)", opcodeSetProperty, key, jsonValue)

	size := len(key) + len(jsonValue) + 9

	err := il.checkLenAndFlush(size)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeSetProperty)
	il.writeValString(key)
	il.writeValBytes(jsonValue)

	return nil
}

func (il *instructionList) writeCallback(callbackID uint32) error {

	il.logf("writeCallback[%d](callbackID=%v)", opcodeCallback, callbackID)

	size := 5

	err := il.checkLenAndFlush(size)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeCallback)
	il.writeValUint32(callbackID)

	return nil
}

func (il *instructionList) writeCallbackLastElement(callbackID uint32) error {

	il.logf("writeCallbackLastElement[%d](callbackID=%v)", opcodeCallbackLastElement, callbackID)

	size := 5

	err := il.checkLenAndFlush(size)
	if err != nil {
		return err
	}

	il.writeValUint8(opcodeCallbackLastElement)
	il.writeValUint32(callbackID)

	return nil
}

func (il *instructionList) writeValUint8(b uint8) {
	il.buf[il.pos] = b
	il.pos++
}

func (il *instructionList) writeValUint32(v uint32) {
	binary.BigEndian.PutUint32(il.buf[il.pos:il.pos+4], v)
	il.pos += 4
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

func (il *instructionList) writeValBytes(s []byte) {

	lenstr := len(s)
	pos := il.pos

	// write length as uint32
	binary.BigEndian.PutUint32(il.buf[pos:pos+4], uint32(lenstr))

	// copy bytes directly from string into buf
	copy(il.buf[pos+4:pos+4+lenstr], s)

	il.pos = pos + 4 + lenstr
}

// // "element and text" pattern (used for script, style, link) goes like:
// // string - element name
// // string - text content (zero length means no text content)
// // uint32 - number of attributes
// // string... - string pairs of key and then value for attributes (number of pairs is number of attributes above, so 1 attr would be 1 in the uint32 above and 2 string - 2 would mean 4 strings, etc.)
// func (il *instructionList) writeValElementAndText(elName, textContent string, attrKV []string) error {

// 	return nil
// }
