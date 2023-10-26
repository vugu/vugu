module github.com/vugu/vugu/wasm-test-suite/test

go 1.21.3

replace github.com/vugu/vugu => ../..

require (
	github.com/vugu/vjson v0.0.0-20200505061711-f9cbed27d3d9
	github.com/vugu/vugu v0.1.1-0.20200406224150-50acda24c5ef
)

require (
	github.com/vugu/xxhash v0.0.0-20191111030615-ed24d0179019 // indirect
	golang.org/x/text v0.13.0 // indirect
)
