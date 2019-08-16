
(function() {

	if (window.vuguRender) { return; } // only once

    const opcodeEnd = 0         // no more instructions in this buffer
    // const opcodeClearRefmap = 1 // clear the reference map, all following instructions must not reference prior IDs
    const opcodeClearEl = 1 // clear the currently selected element
    // const opcodeSetHTMLRef = 2  // assign ref for html tag
    // const opcodeSetHeadRef = 3  // assign ref for head tag
    // const opcodeSetBodyRef = 4  // assign ref for body tag
    // const opcodeSelectRef = 5   // select element by ref
	const opcodeRemoveOtherAttrs = 5 // remove any elements for the current element that we didn't just set
    const opcodeSetAttrStr = 6  // assign attribute string to the current selected element
    const opcodeSelectMountPoint = 7 // selects the mount point element and pushes to the stack - the first time by selector but every subsequent time it will reuse the element from before (because the selector may not match after it's been synced over, it's id etc), also make sure it's of this element name and recreate if so
	// const opcodePicardFirstChildElement = 8  // ensure an element first child and push onto element stack
	// const opcodePicardFirstChildText    = 9  // ensure a text first child and push onto element stack
	// const opcodePicardFirstChildComment = 10 // ensure a comment first child and push onto element stack
	// const opcodeSelectParent                   = 11 // pop from the element stack
	// const opcodePicardFirstChild = 12  // ensure an element first child and push onto element stack

    const opcodeMoveToFirstChild     = 20 // move node selection to first child (doesn't have to exist)
	const opcodeSetElement           = 21 // assign current selected node as an element of the specified type
	// const opcodeSetElementAttr       = 22 // set attribute on current element
	const opcodeSetText              = 23 // assign current selected node as text with specified content
	const opcodeSetComment           = 24 // assign current selected node as comment with specified content
	const opcodeMoveToParent         = 25 // move node selection to parent
	const opcodeMoveToNextSibling    = 26 // move node selection to next sibling (doesn't have to exist)
	const opcodeRemoveOtherEventListeners  = 27 // remove all event listeners from currently selected element that were not just set
	const opcodeSetEventListener     = 28 // assign event listener to currently selected element
    const opcodeSetInnerHTML         = 29 // set the innerHTML for an element

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

    window.vuguGetActiveEvent = function() {
        let state = window.vuguState || {}; window.vuguState = state;
        return state.activeEvent;
    }
    window.vuguGetActiveEventTarget = function() {
        let state = window.vuguState || {}; window.vuguState = state;
        return state.activeEvent && state.activeEvent.target;
    }
    window.vuguGetActiveEventCurrentTarget = function() {
        let state = window.vuguState || {}; window.vuguState = state;
        return state.activeEvent && state.activeEvent.currentTarget;
    }
    window.vuguActiveEventPreventDefault = function() {
        let state = window.vuguState || {}; window.vuguState = state;
        if (state.activeEvent && state.activeEvent.preventDefault) {
            state.activeEvent.preventDefault();
        }
    }
    window.vuguActiveEventStopPropagation = function() {
        let state = window.vuguState || {}; window.vuguState = state;
        if (state.activeEvent && state.activeEvent.stopPropagation) {
            state.activeEvent.stopPropagation();
        }
    }

	window.vuguSetEventHandlerAndBuffer = function(eventHandlerFunc, eventBuffer) { 
		let state = window.vuguState || {};
        window.vuguState = state;
        state.eventBuffer = eventBuffer;
        state.eventBufferView = new DataView(eventBuffer.buffer, eventBuffer.byteOffset, eventBuffer.byteLength);
        state.eventHandlerFunc = eventHandlerFunc;
    }

	window.vuguRender = function(buffer) { 
        
        // NOTE: vuguRender must not automatically reset anything between calls.
        // Since a series of instructions might get cut off due to buffer end, we
        // need to be able to just pick right up with the next call where we left off.
        // The caller decides when to reset things by sending the appropriate
        // instruction(s).

		let state = window.vuguState || {};
		window.vuguState = state;

		console.log("vuguRender called", buffer);

        let textEncoder = new TextEncoder();

		let bufferView = new DataView(buffer.buffer, buffer.byteOffset, buffer.byteLength);

        var decoder = new Decoder(bufferView, 0);
        
        // state.refMap = state.refMap || {};
        // state.curRef = state.curRef || ""; // current reference number (as a hex string)
        // state.curRefEl = state.curRefEl || null; // current reference element
        // state.elStack = state.elStack || []; // stack of elements as we traverse the DOM tree

        // mount point element
        state.mountPointEl = state.mountPointEl || null; 

        // currently selected element
        state.el = state.el || null;

        // specifies a "next" move for the current element, if used it must be followed by
        // one of opcodeSetElement, opcodeSetText, opcodeSetComment, which will create/replace/use existing
        // the element and put it in "el".  The point is this allow us to select nodes that may
        // not exist yet, knowing that the next call will specify what that node is.  It's more complex here
        // but makes it easier to generate instructions while walking a DOM tree.
        // Value is one of "first_child", "next_sibling"
        // (Parents always exist and so doesn't use this mechanism.)
        state.nextElMove = state.nextElMove || null;

        // keeps track of attributes that are being set on the current element, so we can remove any extras
        state.elAttrNames = state.elAttrNames || {};

        // map of positionID -> array of listener spec and handler function, for all elements
        state.eventHandlerMap = state.eventHandlerMap || {};
    
        // keeps track of event listeners that are being set on the current element, so we can remvoe any extras
        state.elEventKeys = state.elEventKeys || {};

        instructionLoop: while (true) {

            let opcode = decoder.readUint8();
            
            // console.log("processing opcode", opcode);
            // console.log("test_span_id: ", document.querySelector("#test_span_id"));

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
                    state.nextElMove = null;
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
                    state.elAttrNames[attrName] = true;
                    // console.log("setting attr", attrName, attrValue, el)
                    break;
                }

                case opcodeSelectMountPoint: {
                    
                    state.elAttrNames = {}; // reset attribute list
                    state.elEventKeys = {};

                    // select mount point using selector or if it was done earlier re-use the one from before
                    let selector = decoder.readString();
                    let nodeName = decoder.readString();
                    // console.log("GOT HERE selector,nodeName = ", selector, nodeName);
                    // console.log("state.mountPointEl", state.mountPointEl);
                    if (state.mountPointEl) {
                        console.log("opcodeSelectMountPoint: state.mountPointEl already exists, using it", state.mountPointEl, "parent is", state.mountPointEl.parentNode);
                        state.el = state.mountPointEl;
                        // state.elStack.push(state.mountPointEl);
                    } else {
                        console.log("opcodeSelectMountPoint: state.mountPointEl does not exist, using selector to find it", selector);
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

                        let newEl = document.createElement(nodeName);
                        el.parentNode.replaceChild(newEl, el);

                        state.mountPointEl = newEl;
                        el = newEl;

                    }

                    state.el = el;

                    state.nextElMove = null;

                    break;
                }

                // case opcodePicardFirstChild: {

            	// 	let nodeType = decoder.readUint8();
                //     let data = decoder.readString();

                //     let oldFirstChildEl = state.el.firstChild;

                //     let newFirstChildEl = null;

                //     let needsCreate = true;
                //     if (oldFirstChildEl) {
                //         // node types from Go are https://godoc.org/golang.org/x/net/html#NodeType
                //         // whereas node types in DOM are https://developer.mozilla.org/en-US/docs/Web/API/Node/nodeType

                //         // text
                //         if (nodeType == 1 && oldFirstChildEl.nodeType == 3) {
                //             needsCreate = false;
                //         } else 
                //         // element
                //         if (nodeType == 3 && oldFirstChildEl.nodeType == 1) {
                //             needsCreate = false;
                //         } else 
                //         // comment
                //         if (nodeType == 4 && oldFirstChildEl.nodeType == 8) {
                //             needsCreate = false;
                //         }

                //     }

                //     if (needsCreate) {

                //         switch (nodeType) {
                //             case 1: {
                //                 newFirstChildEl = document.createTextNode(data);
                //                 break;
                //             }
                //             case 3: {
                //                 newFirstChildEl = document.createElement(data);
                //                 break;
                //             }
                //             case 4: {
                //                 newFirstChildEl = document.createComment(data);
                //                 break;
                //             }
                //         }
    
                //     }

                //     if (newFirstChildEl) {
                //         if (oldFirstChildEl) {
                //             state.el.replaceChild(newFirstChildEl, oldFirstChildEl);
                //         } else {
                //             state.el.appendChild(newFirstChildEl);
                //         }
                //         state.el = newFirstChildEl;
                //     } else {
                //         state.el = oldFirstChildEl;
                //     }

                //     break;
                // }

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

                // remove any elements for the current element that we didn't just set
                case opcodeRemoveOtherAttrs: {

                    if (!state.el) {
                        throw "no element selected";
                    }

                    if (state.nextElMove) {
                        throw "cannot call opcodeRemoveOtherAttrs when nextElMove is set";
                    }

                    // build a list of attribute names to remove
                    let rmAttrNames = [];
                    for (let i = 0; i < state.el.attributes.length; i++) {
                        if (!state.elAttrNames[state.el.attributes[i].name]) {
                            rmAttrNames.push(state.el.attributes[i].name);
                        }
                    }

                    // remove them
                    for (let i = 0; i < rmAttrNames.length; i++) {
                        state.el.attributes.removeNamedItem(rmAttrNames[i]);
                    }

                    break;
                }

                // move node selection to parent
                case opcodeMoveToParent: {

                    // if first_child is next move then we just unset this
                    if (state.nextElMove == "first_child") {
                        state.nextElMove = null;
                    } else {
                        // otherwise we actually move and also reset nextElMove
                        state.el = state.el.parentNode;
                        state.nextElMove = null;
                    }

                    break;
                }

                // move node selection to first child (doesn't have to exist)
                case opcodeMoveToFirstChild: {

                    // if a next move already set, then we need to execute it before we can do this
                    if (state.nextElMove) {
                        if (state.nextElMove == "first_child") {
                            state.el = state.el.firstChild;
                            if (!state.el) { throw "unable to find state.el.firstChild"; }
                        } else if (state.nextElMove == "next_sibling") {
                            state.el = state.el.nextSibling;
                            if (!state.el) { throw "unable to find state.el.nextSibling"; }
                        }
                        state.nextElMove = null;
                    }

                    if (!state.el) { throw "must have current selection to use opcodeMoveToFirstChild"; }
                    state.nextElMove = "first_child";

                    break;
                }
                
                // move node selection to next sibling (doesn't have to exist)
                case opcodeMoveToNextSibling: {

                    // if a next move already set, then we need to execute it before we can do this
                    if (state.nextElMove) {
                        if (state.nextElMove == "first_child") {
                            state.el = state.el.firstChild;
                            if (!state.el) { throw "unable to find state.el.firstChild"; }
                        } else if (state.nextElMove == "next_sibling") {
                            state.el = state.el.nextSibling;
                            if (!state.el) { throw "unable to find state.el.nextSibling"; }
                        }
                        state.nextElMove = null;
                    }

                    if (!state.el) { throw "must have current selection to use opcodeMoveToNextSibling"; }
                    state.nextElMove = "next_sibling";

                    break;
                }
                
                // assign current selected node as an element of the specified type
                case opcodeSetElement: {
                    
                    let nodeName = decoder.readString();

                    state.elAttrNames = {};
                    state.elEventKeys = {};

                    // handle nextElMove cases

                    if (state.nextElMove == "first_child") {
                        state.nextElMove = null;
                        let newEl = state.el.firstChild;
                        if (newEl) { 
                            state.el = newEl; 
                            break; 
                        } else {
                            newEl = document.createElement(nodeName);
                            state.el.appendChild(newEl);
                            state.el = newEl;
                            break; // we're done here, since we just created the right element
                        }
                    } else if (state.nextElMove == "next_sibling") {
                        state.nextElMove = null;
                        let newEl = state.el.nextSibling;
                        if (newEl) { 
                            state.el = newEl; 
                            break; 
                        } else {
                            newEl = document.createElement(nodeName);
                            // console.log("HERE1", state.el);
                            // state.el.insertAdjacentElement(newEl, 'afterend');
                            state.el.parentNode.appendChild(newEl);
                            state.el = newEl;
                            break; // we're done here, since we just created the right element
                        }
                    } else if (state.nextElMove) {
                        throw "bad state.nextElMove value: " + state.nextElMove;
                    }

                    // if we get here we need to verify that state.el is in fact an element of the right type
                    // and replace if not

                    if (state.el.nodeType != 1 || state.el.nodeName.toUpperCase() != nodeName.toUpperCase()) {

                        let newEl = document.createElement(nodeName);
                        // throw "stopping here";
                        state.el.parentNode.replaceChild(newEl, state.el);
                        state.el = newEl;

                    }

                    break;
                }

                // assign current selected node as text with specified content
                case opcodeSetText: {

                    let content = decoder.readString();

                    // console.log("in opcodeSetText 1");

                    // handle nextElMove cases

                    if (state.nextElMove == "first_child") {
                        state.nextElMove = null;
                        let newEl = state.el.firstChild;
                        // console.log("in opcodeSetText 2");
                        if (newEl) { 
                            state.el = newEl; 
                            break;
                        } else {
                            let newEl = document.createTextNode(content);
                            state.el.appendChild(newEl);
                            state.el = newEl;
                            // console.log("in opcodeSetText 3");
                            break; // we're done here, since we just created the right element
                        }
                    } else if (state.nextElMove == "next_sibling") {
                        state.nextElMove = null;
                        let newEl = state.el.nextSibling;
                        // console.log("in opcodeSetText 4");
                        if (newEl) { 
                            state.el = newEl; 
                            break; 
                        } else {
                            let newEl = document.createTextNode(content);
                            // state.el.insertAdjacentElement(newEl, 'afterend');
                            state.el.parentNode.appendChild(newEl);
                            state.el = newEl;
                            // console.log("in opcodeSetText 5");
                            break; // we're done here, since we just created the right element
                        }
                    } else if (state.nextElMove) {
                        throw "bad state.nextElMove value: " + state.nextElMove;
                    }

                    // if we get here we need to verify that state.el is in fact a node of the right type
                    // and with right content and replace if not
                    // console.log("in opcodeSetText 6");

                    if (state.el.nodeType != 3) {

                        let newEl = document.createTextNode(content);
                        state.el.parentNode.replaceChild(newEl, state.el);
                        state.el = newEl;
                        // console.log("in opcodeSetText 7");

                    } else {
                        // console.log("in opcodeSetText 8");
                        state.el.textContent = content;
                    }
                    // console.log("in opcodeSetText 9");

                    break;
                }

                // assign current selected node as comment with specified content
                case opcodeSetComment: {
                    
                    let content = decoder.readString();

                    // handle nextElMove cases

                    if (state.nextElMove == "first_child") {
                        state.nextElMove = null;
                        let newEl = state.el.firstChild;
                        if (newEl) { 
                            state.el = newEl; 
                            break; 
                        } else {
                            let newEl = document.createComment(content);
                            state.el.appendChild(newEl);
                            state.el = newEl;
                            break; // we're done here, since we just created the right element
                        }
                    } else if (state.nextElMove == "next_sibling") {
                        state.nextElMove = null;
                        let newEl = state.el.nextSibling;
                        if (newEl) { 
                            state.el = newEl; 
                            break; 
                        } else {
                            let newEl = document.createComment(content);
                            // state.el.insertAdjacentElement(newEl, 'afterend');
                            state.el.parentNode.appendChild(newEl);
                            state.el = newEl;
                            break; // we're done here, since we just created the right element
                        }
                    } else if (state.nextElMove) {
                        throw "bad state.nextElMove value: " + state.nextElMove;
                    }

                    // if we get here we need to verify that state.el is in fact a node of the right type
                    // and with right content and replace if not

                    if (state.el.nodeType != 8) {

                        let newEl = document.createComment(content);
                        state.el.parentNode.replaceChild(newEl, state.el);
                        state.el = newEl;

                    } else {
                        state.el.textContent = content;
                    }

                    break;
                }

                case opcodeSetInnerHTML: {

                    let html = decoder.readString();

                    if (!state.el) { throw "opcodeSetInnerHTML must have currently selected element"; }
                    if (state.nextElMove) { throw "opcodeSetInnerHTML nextElMove must not be set"; }
                    if (state.el.nodeType != 1) { throw "opcodeSetInnerHTML currently selected element expected nodeType 1 but has: " + state.el.nodeType; }

                    state.el.innerHTML = html;

                    break;
                }

                // remove all event listeners from currently selected element that were not just set
                case opcodeRemoveOtherEventListeners: {
                    this.console.log("opcodeRemoveOtherEventListeners");

                    let positionID = decoder.readString();

                    // look at all registered events for this positionID
                    let emap = state.eventHandlerMap[positionID] || {};
                    // for any that we didn't just set, remove them
                    let toBeRemoved = [];
                    for (let k in emap) {
                        if (!state.elEventKeys[k]) {
                            toBeRemoved.push(k);
                        }
                    }

                    // for each one that was missing, we remove from emap and call removeEventListener
                    for (let i = 0; i < toBeRemoved.length; i++) {
                        let f = emap[k];
                        let k = toBeRemoved[i];
                        let kparts = k.split("|");
                        state.el.removeEventListener(kparts[0], f, {capture:!!kparts[1], passive:!!kparts[2]});
                        delete emap[k];
                    }

                    // if emap is empty now, remove the entry from eventHandlerMap altogether
                    if (Object.keys(emap).length == 0) {
                        delete state.eventHandlerMap[positionID];
                    } else {
                        state.eventHandlerMap[positionID] = emap;
                    }

                    break;
                }
            
                // assign event listener to currently selected element
                case opcodeSetEventListener: {
                    let positionID = decoder.readString();
                    let eventType = decoder.readString();
                    let capture = decoder.readUint8();
                    let passive = decoder.readUint8();

                    if (!state.el) {
                        throw "must have state.el set in order to call opcodeSetEventListener";
                    }

                    var eventKey = eventType + "|" + (capture?"1":"0") + "|" + (passive?"1":"0");
                    state.elEventKeys[eventKey] = true;

                    // map of positionID -> map of listener spec and handler function, for all elements
                    //state.eventHandlerMap
                    let emap = state.eventHandlerMap[positionID] || {};

                    // register function if not done already
                    let f = emap[eventKey];
                    if (!f) {
                        f = function(event) {

                            // set the active event, so the Go code and call back in and examine it if needed
                            state.activeEvent = event; 

                            let eventObj = {};
                            // console.log(event);
                            for (let i in event) {
                                let itype = typeof(event[i]);
                                // copy primitive values directly
                                if ((itype == "boolean" || itype == "number" || itype == "string") && true/*event.hasOwnProperty(i)*/) {
                                    eventObj[i] = event[i];
                                }
                            }

                            // also do the same for anything in "target"
                            if (event.target) {
                                eventObj.target = {};
                                let et = event.target;
                                for (let i in et) {
                                    let itype = typeof(et[i]);
                                    if ((itype == "boolean" || itype == "number" || itype == "string") && true/*et.hasOwnProperty(i)*/) {
                                        eventObj.target[i] = et[i];
                                    }
                                }
                            }
                            
                            // console.log(eventObj);
                            // console.log(JSON.stringify(eventObj));

                            let fullJSON = JSON.stringify({
                                
                                // include properties from event registration
                                position_id: positionID,
                                event_type: eventType,
                                capture: !!capture,
                                passive: !!passive,

                                // the event object data as extracted above
                                event_summary: eventObj,

                            });

                            // console.log(state.eventBuffer);

                            // write JSON to state.eventBuffer with zero char as termination

                            
                            let encodeResultBuffer = textEncoder.encode(fullJSON);
                            //console.log("encodeResult", encodeResult);
                            state.eventBuffer.set(encodeResultBuffer, 4); // copy encoded string to event buffer
                            // now write length using DataView as uint32
                            state.eventBufferView.setUint32(0, encodeResultBuffer.byteLength - encodeResultBuffer.byteOffset);

                            // let result = textEncoder.encodeInto(fullJSON, state.eventBuffer);
                            // let eventBufferDataView = new DataView(state.eventBuffer.buffer, state.eventBuffer.byteOffset, state.eventBuffer.byteLength);
                            // eventBufferDataView.setUint8(result.written, 0);

                            // write length after, since only now do we know the final length
                            // state.eventBufferView.setUint32(0, result.written);

                            // serialize event into the event buffer, somehow,
                            // and keep track of the target element, also consider grabbing
                            // the value or relevant properties as appropriate for form things
                            
                            state.eventHandlerFunc.call(null); // call with null this avoid unnecessary js.Value reference

                            // unset the active event
                            state.activeEvent = null;
                        };    
                        emap[eventKey] = f;

                        // this.console.log("addEventListener", eventType);
                        state.el.addEventListener(eventType, f, {capture:capture, passive:passive});
                    }

                    state.eventHandlerMap[positionID] = emap;

                    this.console.log("opcodeSetEventListener", positionID, eventType, capture, passive);
                    break;
                }
            
                // case opcodeSelectParent: {
                //     // select parent
                //     state.el = state.el.parentNode;
                //     break;
                // }

                default: {
                    console.error("found invalid opcode", opcode);
                    return;
                }
            }

		}

	}

})()
