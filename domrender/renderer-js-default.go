// +build !tinygo

package domrender

import (
	"fmt"
	"runtime/debug"
)

func (r *JSRenderer) sendEventWaitCh() {

	if panicr := recover(); panicr != nil {
		fmt.Println("handleRawDOMEvent caught panic", panicr)
		debug.PrintStack()

		// in error case send false to tell event loop to exit
		select {
		case r.eventWaitCh <- false:
		default:
		}
		return

	}

	// in normal case send true to the channel to tell the event loop it should render
	select {
	case r.eventWaitCh <- true:
	default:
	}
}
