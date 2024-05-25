module github.com/vugu/vugu/legacy-wasm-test-suite/test

go 1.21.4

replace github.com/vugu/vugu => ../..

//replace github.com/vugu/vgrouter => ../../../vgrouter

require (
	github.com/vugu/vgrouter v0.0.0-20200725205318-eeb478c42e5d
	github.com/vugu/vjson v0.0.0-20200505061711-f9cbed27d3d9
	github.com/vugu/vugu v0.3.0
)

require github.com/vugu/xxhash v0.0.0-20191111030615-ed24d0179019 // indirect
