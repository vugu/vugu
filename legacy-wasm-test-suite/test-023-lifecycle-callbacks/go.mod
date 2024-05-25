module github.com/vugu/vugu/legacy-wasm-test-suite/test

go 1.21.4

replace github.com/vugu/vugu => ../..

require (
	github.com/vugu/vjson v0.0.0-20200505061711-f9cbed27d3d9
	github.com/vugu/vugu v0.0.0-00010101000000-000000000000
)

require github.com/vugu/xxhash v0.0.0-20191111030615-ed24d0179019 // indirect
