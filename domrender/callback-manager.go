package domrender

import (
	"github.com/vugu/vugu"
)

// callbackManager handles the element lifecycle related events for browser DOM elements in order to
// implement vg-js-create and vg-js-populate and anything else like this.
type callbackManager struct {
	nextCallbackID  uint32 // large enough and representable as float64 without loss
	callbackInfoMap map[uint32]callbackInfo
}

type callbackInfoType int

const (
	callbackCreate = iota + 1
	callbackPopulate
)

// startRender prepares for the next render cycle
func (cm *callbackManager) startRender() {
	cm.nextCallbackID = 1
	if l := len(cm.callbackInfoMap); l > 0 {
		cm.callbackInfoMap = make(map[uint32]callbackInfo, l)
	} else {
		cm.callbackInfoMap = make(map[uint32]callbackInfo)
	}
}

func (cm *callbackManager) doneRender() {
}

// sets the create and populate handlers and returns a callbackID for each,
// if no instruction is needed for something then 0 will be returned for its ID
func (cm *callbackManager) addCreateAndPopulateHandlers(create, populate vugu.JSValueHandler) (uint32, uint32) {

	if create == nil && populate == nil {
		return 0, 0
	}

	// for either one to work, we need the element reference, so we always register the first one
	// (even if the create function is empty)
	cid := cm.nextCallbackID
	cm.nextCallbackID++
	cm.callbackInfoMap[cid] = callbackInfo{
		typ: callbackCreate,
		f:   create,
	}

	// if populate is not nil then set it up too with it's own ID and it's createID pointing
	// back to the other one (so it can get at the js.Value later)
	var pid uint32
	if populate != nil {
		pid = cm.nextCallbackID
		cm.nextCallbackID++
		cm.callbackInfoMap[pid] = callbackInfo{
			typ:      callbackPopulate,
			f:        populate,
			createID: cid,
		}
	}

	return cid, pid
}
