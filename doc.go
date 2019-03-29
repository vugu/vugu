/*
	Package vugu provides core functionality including vugu->go codegen and in-browser DOM syncing running in WebAssembly.  See http://www.vugu.org/

	Since Vugu projects can have both client-side (running in WebAssembly) as well as server-side functionality many of the
	items in this package are available in both environments.  Some however are either only available or only generally useful
	in one environment.

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

	Clients

	JSEnv for in-browser DOM synchronization.

	DOM Events:

	Event Environment:

	Servers

	ParserGo and ParserGoPkg

	StaticHTMLEnv for server-side HTML generation and testing.

*/
package vugu
