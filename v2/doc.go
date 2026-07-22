/*
	Package vugu provides core functionality including vugu->go codegen and in-browser DOM syncing running in WebAssembly.  See http://www.vugu.org/

	Since Vugu projects can have both client-side (running in WebAssembly) as well as server-side functionality many of the
	items in this package are available in both environments.  Some however are either only available or only generally useful
	in one environment.

	Common functionality includes the ComponentType interface, and ComponentInst struct corresponding to an instantiated componnet.
	VGNode and related structs are used to represent a virtual Document Object Model.  It is based on golang.org/x/net/html but
	with additional fields needed for Vugu.  Data hashing is performed by ComputeHash() and can be customized by implementing
	the DataHasher interface.

	Client-side code uses JSEnv to maintain a render loop and regenerate virtual DOM and efficiently synchronize it with
	the browser as needed.  DOMEvent is a wrapper around events from the browser and EventEnv is used to synchronize data
	access when writing event handler code that spawns goroutines.  Where appropriate, server-side stubs are available
	so components can be compiled for both client (WebAssembly) and server (server-side rendering and testing).

	Server-side code can use ParserGo and ParserGoPkg to parse .vugu files and code generate a corresponding .go file.
	StaticHTMLEnv can be used to generate static HTML, similar to the output of JSEnv but can be run on the server.
	Supported features are approximately the same minus event handling, unapplicable to static output.

*/
package vugu

/*

old notes:

	Common

	Components and Registration...

	VGNode and friends for virtual DOM:

	<b>Data hashing is perfomed with the ComputeHash() function.<b>  <em>It walks your data structure</en> and hashes the information as it goes.
	It uses xxhash internally and returns a uint64.  It is intended to be both fast and have good hash distribution to avoid
	collision-related bugs.

		someData := &struct{A string}{A:"test"}
		hv := ComputeHash(someData)

	If the DataHasher interface is implemented by a particular type then ComputeHash will just called it and hash it's return
	into the calculation.  Otherwise ComputeHash walks the data and finds primitive values and hashes them byte by bytes.
	Nil interfaces and nil pointers are skipped.

	Effective hashing is an important part of achieving good performance in Vugu applications, since the question "is this different
	than it was before" needs to be asked frequently.  The current experiment is to rely entirely on data hashing for change detection
	rather than implementing a data-binding system.

*/
