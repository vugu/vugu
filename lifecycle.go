package vugu

// BackgroundUpdater is an interface that if the component meets
// it, a channel will be passed to it early in the start sequence.
// This channel can be used to signal the event/redraw loop that you have
// completed some type of background update and you would like
// to be called, via Update(), to update the necessary parts of the
// UI state without messing up the render loop.  The object pushed into the
// channel by BackgroundInit() (or anything else that has access to the
// channel) will passed to the Update method.  Render will be called
// when Update returns.
type BackgroundUpdater interface {
	// This use of the interface{} as the signal means that you can't pass
	// null as the value because the infrastructure thinks you want to (or did)
	// close the channel.
	BackgroundInit(chan interface{})
	Update(interface{})
}

// If your component meets this interface it will be called
// on the CapabilityCheck  method after the WebAssembly is loaded and running
// but before we start the render loop.  The intent of the booleans
// is to try and pass some type of "capabilities list" so that an
// application that needs certain things from the browser
// (Webworkers? LocalStorage?) can take action immediately.
// If the Startup method responds with a string that is not "",
// that will be interpreted as a URL to send to the browser,
// probably a page like "sorry, upgrade your bromser."
type CapabilityChecker interface {
	CapabilityCheck(bool, bool, bool) string
}

// Starter is a life-cycle call for components that want a call immediately
// before the application begins.  The normal event loop.
//
// If you implement all the "start time" interfaces, the sequence of
// calls is
// 1) CapabilityCheck()
// 2) BackgroundInit()
// 3) Start()
type Starter interface {
	Start()
}

// Ender is a life-cycle call for components that want a call immediately
// before the application exits.
type Ender interface {
	End()
}
