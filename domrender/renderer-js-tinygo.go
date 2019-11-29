// +build tinygo

package domrender

import (
	"errors"
	"fmt"
)

func (r *JSRenderer) sendEventWaitCh() {

	// NOTE: tinygo version omited recover and stack trace stuff for now

	// in normal case send true to the channel to tell the event loop it should render
	select {
	case r.eventWaitCh <- true:
	default:
	}
}

func errorf(f string, args ...interface{}) error {
	return errors.New(fmt.Sprintf(f, args...))
}
