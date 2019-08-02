
(function() {

	if (window.vuguRender) { return; } // only once

    const opcodeEnd = 0         // no more instructions in this buffer
    const opcodeClearRefmap = 1 // clear the reference map, all following instructions must not reference prior IDs
    const opcodeSetHTMLRef = 2  // assign ref for html tag
    const opcodeSetHeadRef = 3  // assign ref for head tag
    const opcodeSetBodyRef = 4  // assign ref for body tag
    const opcodeSelectRef = 5   // select element by ref
    const opcodeSetAttrStr = 6  // assign attribute string to the current selected element

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

    window._vuguRefMap = {};

	window.vuguRender = function(buffer) {    

		let state = window.vuguRenderState || {};
		window.vuguRenderState = state;

		console.log("vuguRender called", buffer);

		let bufferView = new DataView(buffer.buffer, buffer.byteOffset, buffer.byteLength);

        var decoder = new Decoder(bufferView, 0);
        
        var refmap = window._vuguRefMap;

        var curref = ""; // current reference number
        var currefel = null; // current reference element

        instructionLoop: while (true) {

			let opcode = decoder.readUint8();

            switch (opcode) {

                case opcodeEnd:
                        break instructionLoop;
    
                case opcodeClearRefmap:
                    refmap = {};
                    window._vuguRefMap = refmap;
                    curref = "";
                    currefel = null;
                    break;

                case opcodeSetHTMLRef:
                    var refstr = decoder.readRefToString();
                    refmap[refstr] = document.querySelector("html");
                    break;

                case opcodeSelectRef:
                    var refstr = decoder.readRefToString();
                    curref = refstr;
                    currefel = refmap[refstr];
                    if (!currefel) {
                        console.error("opcodeSelectRef: refstr does not exist", refstr);
                    }
                    break;

                case opcodeSetAttrStr:
                    if (!currefel) {
                        console.error("opcodeSetAttrStr: no current reference");
                    }
                    var attrName = decoder.readString();
                    var attrValue = decoder.readString();
                    currefel.setAttribute(attrName, attrValue);
                    break;

                default:
                    console.error("found invalid opcode", opcode);
                    return;
            }

		}

	}

})()
