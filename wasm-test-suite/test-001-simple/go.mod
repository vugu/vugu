module github.com/vugu/vugu/wasm-test-suite/test-001-simple

replace github.com/vugu/vugu => ../..

go 1.23

toolchain go1.23.5

require (
	github.com/chromedp/chromedp v0.12.1
	github.com/stretchr/testify v1.10.0
	github.com/vugu/vjson v0.0.0-20200505061711-f9cbed27d3d9
	github.com/vugu/vugu v0.4.0
)

require (
	github.com/chromedp/cdproto v0.0.0-20250126231910-1730200a0f74 // indirect
	github.com/chromedp/sysutil v1.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/vugu/xxhash v0.0.0-20191111030615-ed24d0179019 // indirect
	golang.org/x/sys v0.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
