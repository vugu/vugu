module github.com/vugu/vugu/wasm-test-suite/test

replace github.com/vugu/vugu => ../..

go 1.21.3

require (
	github.com/vugu/vjson v0.0.0-20200505061711-f9cbed27d3d9
	github.com/vugu/vugu v0.1.1-0.20191208090309-fa72e903246b
)

require github.com/vugu/xxhash v0.0.0-20191111030615-ed24d0179019 // indirect
