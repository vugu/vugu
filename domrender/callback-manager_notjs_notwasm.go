//go:build !js || !wasm

package domrender

import (
	"fmt"

	"github.com/vugu/vugu"
	js "github.com/vugu/vugu/js"
)

type callbackInfo struct {
	typ callbackInfoType    // create or populate
	f   vugu.JSValueHandler // to be called when we get the js.Value
	el  js.Value            // the js element
	// for populate, the ID of the corresponding create callbackInfo
	// (so we can grab the element from it)
	createID uint32
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
