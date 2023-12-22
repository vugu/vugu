package domrender

import (
	"fmt"

	"github.com/vugu/vugu"
	js "github.com/vugu/vugu/js"
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

type callbackInfo struct {
	typ callbackInfoType    // create or populate
	f   vugu.JSValueHandler // to be called when we get the js.Value
	el  js.Value            // the js element
	// for populate, the ID of the corresponding create callbackInfo
	// (so we can grab the element from it)
	createID uint32
}

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

// callback is call when we get a callback from the render script
func (cm *callbackManager) callback(this js.Value, args []js.Value) interface{} {

	if len(args) < 1 {
		panic(fmt.Errorf("no args passed to callbackManager.callback"))
	}

	cbIDVal := args[0]
	cbID := uint32(cbIDVal.Int())
	if cbID == 0 {
		panic(fmt.Errorf("callbackManager.callback got zero callback ID"))
	}

	cbInfo := cm.callbackInfoMap[cbID]
	switch cbInfo.typ {

	case callbackCreate:
		jsVal := args[1]
		cbInfo.el = jsVal
		cm.callbackInfoMap[cbID] = cbInfo
		// f can be nil when vg-js-populate is set but vg-js-create is not,
		// we still need the element to be recorded above
		if cbInfo.f != nil {
			cbInfo.f.JSValueHandle(jsVal)
		}

	case callbackPopulate:
		createID := cbInfo.createID
		if createID == 0 {
			panic(fmt.Errorf("callbackManager.callback handling populate found createID==0"))
		}
		createInfo := cm.callbackInfoMap[createID]
		jsVal := createInfo.el
		if jsVal.IsUndefined() { // if we got all the way here, there should always be a value
			panic(fmt.Errorf("callbackManager.callback handling populate got undefined jsVal"))
		}
		// and f should always be set, panic if not so we can track down why
		if cbInfo.f == nil {
			panic(fmt.Errorf("callbackManager.callback handling populate got nil JSvalueHandler"))
		}
		cbInfo.f.JSValueHandle(jsVal)

	default:
		panic(fmt.Errorf("unknown callback type %#v", cbInfo.typ))

	}

	return nil
}
