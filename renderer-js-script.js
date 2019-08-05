
(function() {

	if (window.vuguRender) { return; } // only once

    const opcodeEnd = 0         // no more instructions in this buffer
    // const opcodeClearRefmap = 1 // clear the reference map, all following instructions must not reference prior IDs
    const opcodeClearEl = 1 // clear the currently selected element
    // const opcodeSetHTMLRef = 2  // assign ref for html tag
    // const opcodeSetHeadRef = 3  // assign ref for head tag
    // const opcodeSetBodyRef = 4  // assign ref for body tag
    // const opcodeSelectRef = 5   // select element by ref
    const opcodeSetAttrStr = 6  // assign attribute string to the current selected element
    const opcodeSelectMountPoint = 7 // selects the mount point element and pushes to the stack - the first time by selector but every subsequent time it will reuse the element from before (because the selector may not match after it's been synced over, it's id etc), also make sure it's of this element name and recreate if so
	// const opcodePicardFirstChildElement = 8  // ensure an element first child and push onto element stack
	// const opcodePicardFirstChildText    = 9  // ensure a text first child and push onto element stack
	// const opcodePicardFirstChildComment = 10 // ensure a comment first child and push onto element stack
	const opcodeSelectParent                   = 11 // pop from the element stack
	const opcodePicardFirstChild = 12  // ensure an element first child and push onto element stack

    // Decoder provides our binary decoding.
    // Using a class because that's what all the cool JS kids are doing these days.
    class Decoder {

        constructor(dataView, offset) {
            this.dataView = dataView;
            this.offset = offset || 0;
            return this;
        }

        // readUint8 reads a single byte, 0-255
        readUint8() {
            var ret = this.dataView.getUint8(this.offset);
            this.offset++;
            return ret;
        }

        // readRefToString reads a 64-bit unsigned int ref but returns it as a hex string
        readRefToString() {
            // read in two 32-bit parts, BigInt is not yet well supported
            var ret = this.dataView.getUint32(this.offset).toString(16).padStart(8, "0") +
                this.dataView.getUint32(this.offset + 4).toString(16).padStart(8, "0");
            this.offset += 8;
            return ret;
        }

        // readString is 4 bytes length followed by utf chars
        readString() {
            var len = this.dataView.getUint32(this.offset);
            var ret = utf8decoder.decode(new DataView(this.dataView.buffer, this.dataView.byteOffset + this.offset + 4, len));
            this.offset += len + 4;
            return ret;
        }

    }

    let utf8decoder = new TextDecoder();

	window.vuguRender = function(buffer) { 
        
        // NOTE: vuguRender must not automatically reset anything between calls.
        // Since a series of instructions might get cut off due to buffer end, we
        // need to be able to just pick right up with the next call where we left off.
        // The caller decides when to reset things by sending the appropriate
        // instruction(s).

		let state = window.vuguRenderState || {};
		window.vuguRenderState = state;

		console.log("vuguRender called", buffer);

		let bufferView = new DataView(buffer.buffer, buffer.byteOffset, buffer.byteLength);

        var decoder = new Decoder(bufferView, 0);
        
        // state.refMap = state.refMap || {};
        // state.curRef = state.curRef || ""; // current reference number (as a hex string)
        // state.curRefEl = state.curRefEl || null; // current reference element
        // state.elStack = state.elStack || []; // stack of elements as we traverse the DOM tree
        state.el = state.el || null; // currently selected element
        state.mountPointEl = state.mountPointEl || null; // mount point element

        instructionLoop: while (true) {

			let opcode = decoder.readUint8();

            switch (opcode) {

                case opcodeEnd: {
                    break instructionLoop;
                }
    
                // case opcodeClearRefmap:
                //     state.refMap = {};
                //     state.curRef = "";
                //     state.curRefEl = null;
                //     break;

                case opcodeClearEl: {
                    state.el = null;
                    break;
                }
        
                // case opcodeSetHTMLRef:
                //     var refstr = decoder.readRefToString();
                //     state.refMap[refstr] = document.querySelector("html");
                //     break;

                // case opcodeSelectRef:
                //     var refstr = decoder.readRefToString();
                //     state.curRef = refstr;
                //     state.curRefEl = state.refMap[refstr];
                //     if (!state.curRefEl) {
                //         throw "opcodeSelectRef: refstr does not exist - " + refstr;
                //     }
                //     break;

                case opcodeSetAttrStr: {
                    let el = state.el;
                    if (!el) {
                        return "opcodeSetAttrStr: no current reference";
                    }
                    let attrName = decoder.readString();
                    let attrValue = decoder.readString();
                    el.setAttribute(attrName, attrValue);
                    // console.log("setting attr", attrName, attrValue, el)
                    break;
                }

                case opcodeSelectMountPoint: {

                    // select mount point using selector or if it was done earlier re-use the one from before
                    let selector = decoder.readString();
                    let nodeName = decoder.readString();
                    // console.log("GOT HERE selector,nodeName = ", selector, nodeName);
                    // console.log("state.mountPointEl", state.mountPointEl);
                    if (state.mountPointEl) {
                        state.el = state.mountPointEl;
                        // state.elStack.push(state.mountPointEl);
                    } else {
                        let el = document.querySelector(selector);
                        if (!el) {
                            throw "mount point selector not found: " + selector;
                        }
                        state.mountPointEl = el;
                        // state.elStack.push(el);
                        state.el = el;
                    }

                    let el = state.el;

                    // make sure it's the right element name and replace if not
                    if (el.nodeName.toUpperCase() != nodeName.toUpperCase()) {

                        var newEl = document.createElement(nodeName);
                        el.parentNode.replaceChild(newEl, el);

                        state.mountPointEl = newEl;
                        el = newEl;

                    }

                    state.el = el;

                    break;
                }

                case opcodePicardFirstChild: {

            		let nodeType = decoder.readUint8();
                    let data = decoder.readString();

                    let oldFirstChildEl = state.el.firstChild;

                    let newFirstChildEl = null;

                    let needsCreate = true;
                    if (oldFirstChildEl) {
                        // node types from Go are https://godoc.org/golang.org/x/net/html#NodeType
                        // whereas node types in DOM are https://developer.mozilla.org/en-US/docs/Web/API/Node/nodeType

                        // text
                        if (nodeType == 1 && oldFirstChildEl.nodeType == 3) {
                            needsCreate = false;
                        } else 
                        // element
                        if (nodeType == 3 && oldFirstChildEl.nodeType == 1) {
                            needsCreate = false;
                        } else 
                        // comment
                        if (nodeType == 4 && oldFirstChildEl.nodeType == 8) {
                            needsCreate = false;
                        }

                    }

                    if (needsCreate) {

                        switch (nodeType) {
                            case 1: {
                                newFirstChildEl = document.createTextNode(data);
                                break;
                            }
                            case 3: {
                                newFirstChildEl = document.createElement(data);
                                break;
                            }
                            case 4: {
                                newFirstChildEl = document.createComment(data);
                                break;
                            }
                        }
    
                    }

                    if (newFirstChildEl) {
                        if (oldFirstChildEl) {
                            state.el.replaceChild(newFirstChildEl, oldFirstChildEl);
                        } else {
                            state.el.appendChild(newFirstChildEl);
                        }
                        state.el = newFirstChildEl;
                    } else {
                        state.el = oldFirstChildEl;
                    }

                    break;
                }

                // case opcodePicardFirstChildElement: {
                //     // ensure an element first child and select

                //     let el = state.el;
                //     let nextEl = el.firstChild;
                //     if (!nextEl) {
                //         nextEl = 
                //     }
                //     state.el = el;

                //     break;
                // }

                // case opcodePicardFirstChildText: {
                //     // ensure a text first child and select
                //     break;
                // }

                // case opcodePicardFirstChildComment: {
                //     // ensure a comment first child and select
                //     break;
                // }

                case opcodeSelectParent: {
                    // select parent
                    state.el = state.el.parentNode;
                    break;
                }

                default: {
                    console.error("found invalid opcode", opcode);
                    return;
                }
            }

		}

	}

})()
