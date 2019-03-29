/*
	Package vugu provides core functionality including vugu->go codegen and in-browser DOM syncing running in WebAssembly.  See http://www.vugu.org/

	Since Vugu projects can have both client-side (running in WebAssembly) as well as server-side functionality many of the
	items in this package are available in both environments.  Some however are either only available or only generally useful
	in one environment.

	Common

	Components and Registration...

	VGNode and friends for virtual DOM

	Data hashing stuff...

	Clients

	JSEnv for in-browser DOM synchronization.

	DOM Events:

	Event Environment:

	Servers

	ParserGo and ParserGoPkg

	StaticHTMLEnv for server-side HTML generation and testing.

*/
package vugu
