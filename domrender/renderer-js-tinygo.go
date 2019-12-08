// +build tinygo

package domrender

import "github.com/vugu/vugu"

func (r *JSRenderer) sendEventWaitCh() {

	// NOTE: tinygo version omited recover and stack trace stuff for now

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

	err := r.render(buildResults)

	// for now, not using defer in Tinygo
	r.eventRWMU.RUnlock()

	return err
}
