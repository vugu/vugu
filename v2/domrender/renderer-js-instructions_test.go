package domrender

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteSetInnerHTML(t *testing.T) {
	tests := []struct {
		bufferSize   int
		position     int
		htmlString   []byte
		sentBuffers  [][]byte
		outputBuffer []byte
		description  string
	}{
		{
			bufferSize:   10,
			position:     0,
			htmlString:   []byte{1, 2, 3, 4},
			sentBuffers:  nil,
			outputBuffer: []byte{29, 0, 0, 0, 4, 1, 2, 3, 4, 0}, // setInner, len 4, 1,2,3,4 end 0
			description:  "a small message which does not need to be buffered",
		},
		{
			bufferSize:   10,
			position:     3,
			htmlString:   []byte{1, 2, 3, 4},
			sentBuffers:  [][]byte{{255, 255, 255, 37, 0, 0, 0, 1, 1, 0}}, // setBuffer, len 1, 1 end 0
			outputBuffer: []byte{29, 0, 0, 0, 3, 2, 3, 4, 0, 0},           // setInner, len 3, 2,3,4 end 0
			description:  "a small message which does need to be buffered because of offsets",
		},
		{
			bufferSize:   10,
			position:     0,
			htmlString:   []byte{1, 2, 3, 4, 5, 6},
			sentBuffers:  [][]byte{{37, 0, 0, 0, 4, 1, 2, 3, 4, 0}}, // setBuffer, len 4, 1,2,3,4 end 0
			outputBuffer: []byte{29, 0, 0, 0, 2, 5, 6, 0, 0, 0},     // setInner, len 2, 5,6, end 0, empty
			description:  "a message which needs to be split across two buffers",
		},
		{
			bufferSize: 10,
			position:   8,
			htmlString: []byte{1, 2, 3, 4, 5, 6},
			sentBuffers: [][]byte{
				{255, 255, 255, 255, 255, 255, 255, 255, 0, 0}, // previous buffer
				{37, 0, 0, 0, 4, 1, 2, 3, 4, 0},                // setBuffer, len 4, 1,2,3,4 end 0
			},
			outputBuffer: []byte{29, 0, 0, 0, 2, 5, 6, 0, 0, 0}, // setInner, len 2, 5,6, end 0, empty
			description:  "a message which needs to be split across two buffers, that won't fit in the original buffer",
		},
		{
			bufferSize: 10,
			position:   0,
			htmlString: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			sentBuffers: [][]byte{
				{37, 0, 0, 0, 4, 1, 2, 3, 4, 0},     // setBuffer, len 4, 1,2,3,4 end 0
				{37, 0, 0, 0, 4, 5, 6, 7, 8, 0},     // setBuffer, len 4, 5,6,7,8 end 0
				{37, 0, 0, 0, 4, 9, 10, 11, 12, 0},  // setBuffer, len 4, 9,10,11,12 end 0
				{37, 0, 0, 0, 4, 13, 14, 15, 16, 0}, // setBuffer, len 4, 13,14,15,16 end 0
			},
			outputBuffer: []byte{29, 0, 0, 0, 4, 17, 18, 19, 20, 0}, // setInner, len 2, 5,6, end 0
			description:  "a big message which needs to be split across four buffers",
		},
	}

	for _, test := range tests {
		buffer := make([]byte, test.bufferSize)
		var sent [][]byte

		il := newInstructionList(buffer, func(il *instructionList) error {
			// On send, add the ending 0 opCode
			buffer[il.pos] = 0

			// Save the old buffer as sent
			data := make([]byte, len(buffer))
			copy(data, buffer)
			sent = append(sent, data)

			// Zero the buffer (not necessary, but makes the tests more readable)
			for i := range buffer {
				buffer[i] = 0
			}
			return nil
		})

		// Prepend detectable garbage if there is a position set
		for i := 0; i < test.position; i++ {
			buffer[i] = 255
		}
		il.pos = test.position

		err := il.writeSetInnerHTML(string(test.htmlString))
		assert.NoError(t, err, test.description)
		assert.Equal(t, test.sentBuffers, sent, test.description)
		assert.Equal(t, test.outputBuffer, buffer, test.description)
	}
}
