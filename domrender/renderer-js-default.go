package domrender

import (
	"fmt"

	"github.com/vugu/vugu"
)

func (r *JSRenderer) sendEventWaitCh() {

	if panicr := recover(); panicr != nil {
		fmt.Println("handleRawDOMEvent caught panic", panicr)
		//		debug.PrintStack()

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

// Render is a render function.
func (r *JSRenderer) Render(buildResults *vugu.BuildResults) error {

	// acquire read lock so events are not changing data while Render is in progress
	r.eventRWMU.RLock()
	defer r.eventRWMU.RUnlock()

	err := r.render(buildResults)
	return err
}
