module github.com/vugu/vugu/v2/examples/html-form

replace github.com/vugu/vugu/v2 => ../..

go 1.23

require (
	github.com/vugu/vjson v0.0.0-20200505061711-f9cbed27d3d9
	github.com/vugu/vugu/v2 v2.0.0-00010101000000-000000000000
	golang.org/x/text v0.21.0
)

require github.com/vugu/xxhash v0.0.0-20191111030615-ed24d0179019 // indirect
