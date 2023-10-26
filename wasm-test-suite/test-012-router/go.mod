module github.com/vugu/vugu/wasm-test-suite/test

go 1.21.3

replace github.com/vugu/vugu => ../..

//replace github.com/vugu/vgrouter => ../../../vgrouter

require (
	github.com/vugu/vgrouter v0.0.0-20200329225024-3b01bdbe25fa
	github.com/vugu/vjson v0.0.0-20200505061711-f9cbed27d3d9
	github.com/vugu/vugu v0.1.0
)

require github.com/vugu/xxhash v0.0.0-20191111030615-ed24d0179019 // indirect
