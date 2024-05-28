# Vugu

[![Travis CI](https://travis-ci.org/vugu/vugu.svg?branch=master)](https://travis-ci.org/vugu/vugu)
[![GoDoc](https://godoc.org/github.com/vugu/vugu?status.svg)](https://godoc.org/github.com/vugu/vugu)
[![stability-experimental](https://img.shields.io/badge/stability-experimental-orange.svg)](https://github.com/emersion/stability-badges#experimental)
<!-- [![Go Report Card](https://goreportcard.com/badge/github.com/vugu/vugu)](https://goreportcard.com/report/github.com/vugu/vugu) -->

Vugu is an experimental library for web UIs written in Go and targeting webassembly.  Guide and docs at https://www.vugu.org.
Godoc at https://godoc.org/github.com/vugu/vugu.

If you've ever wanted to write a UI not in JS but pure Go... and run it in your browser, right now... That (experimental;) future is here!

Introducing Vugu (pronounced /ˈvuː.ɡuː/), a VueJS-inspired library in Go targeting wasm.

No node. No JS. No npm. No node_modules folder competing with your music library for disk space.

## Updates ♨

* 2024-05-25 Move to a [mage](https://magefile.org/) based build process
<img src="https://cdnjs.cloudflare.com/ajax/libs/ionicons/4.5.6/collection/build/ionicons/svg/logo-slack.svg" width="17" height="17"> Join the conversation: [Gophers on Slack](https://invite.slack.golangbridge.org/), channel #vugu

## Highlights

* Runs in-browser using WebAssembly
* Single-file components
* Vue-like markup syntax
* Write idiomatic Go code
* Rapid prototyping
* ~3 minute setup
* Standard Go build tools


## Start

Get started: http://www.vugu.org/doc/start

Still a work in progress, but a lot of things are already functional. Some work really well.

**You must have Go v1.22.3 as a minimum to use `vugu`. We require the for loop changes that were introduced in Go v1.22, and v1.22.3 was the lastest at the time writing.**


## Building `vugu` for Contributors

`vugu` now uses [mage](https://magefile.org/) to manage the build of the `vugu` tools - `vugugen`, `vugufmt` and `vgfrom`.
[Mage](https://magefile.org/) is also used to manage the testing process.

Please see the updated build instruction in the [Contributors Guide](https://github.com/vugu/vugu/blob/master/CONTRIBUTING.md)


## Running the examples

To run the examples you must have the [`mage`](https://magefile.org/) tool, `docker` and `goimports` installed. `vugu` uses `mage` to manage the build process.

The simplest way to install `mage` is:

```
git clone https://github.com/magefile/mage
cd mage
go run bootstrap.go
```

You must run `mage` from the module root of `vugu`, this is the directory where the top level `go.mod` exists.

You will also need the [`goimports`](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) tool installed. It is very likely that you have this installed already, normally as part of an editor plugin. If not then the it can be installed with:

```
go install golang.org/x/tools/cmd/goimports@latest
```

In order to run the examples you will also need `docker` installed. If you don't have `docker` installed then follow the [docker install instructions](https://docs.docker.com/engine/install/#licensing). Each example will be served by a local `nginx` container.

All of the examples are in the `examples` directory. Each sub-directory of `examples` contains a single example. Each example is it own Go module.

Building and serving all of the examples is a simple as:

```
cd path/to/vugu
mage examples
```

or

```
cd path/to/vugu
mage -v examples
```

Each example will be served at a URL of the form

```
http://localhost:8888/<name-of-example-directory>
```

For example to see the `fetch-and-display` example the URL would be:

```
http://localhost:8888/fetch-and-display
```

Or if you only want to run a single example use:

```
cd path/to/vugu
mage singleExample <name-of-example-module>
```

For example to serve just the `fetch-and-display` example the command would be:

```
cd path/to/vugu
mage singleExample github.com/vugu/vugu/example/fetch-and-display
```

### Creating a new example

If you need to create a new example the process is fairly straight forward. The critical point is to base it on a working example.

For example

```
cd /path/to/vugu
cp -r ./examples/fetch-and-display/ ./examples/my-new-example

```

The `cp` will copy everything in the directory including the critical local `.gitignore` file to the new example directory. Please make sire the `.gitignore` is present to ensure that vugu generated files are not submitted to the repository.

You then need to edit the `./examples/my-new-example/go.mod` to change the module name. ***This step is critical.***

The module name must be changed to match the example, so in this case the module name would be changed to `github.com/vugu/vugu/examples/my-new-example`

You can then edit the `root.vugu`, `root.go` as needed to support the example, or add more `*.vugu` and `*.go` files as necessary.

The files `main_wasm.go` and `wasm_exec.js` should not be edited.

The examples `index.html` file will need to edited in two distinct places. The first is circa line 11

```
<script src="/fetch-and-display/wasm_exec.js"></script>
```

To change the path to reflect the name of the example. In this case:

```
<script src="/my-new-example/wasm_exec.js"></script>
```


The second change is similar but reflects the path of the `main.wasm` binary. This is circa line 29

```
WebAssembly.instantiateStreaming(fetch("/fetch-and-display/main.wasm"), go.importObject).then((result) => {
```

which in this case would be changed to:

```
WebAssembly.instantiateStreaming(fetch("/my-new-example/main.wasm"), go.importObject).then((result) => {
```

The new example can then be built and served with:

```
cd /path/to/vugu
mage examples
```

Or individually like this:

```
cd /path/to/vugu
mage singleExample github.com/vugu/vugu/example/my-new-example
```


## Abbreviated Roadmap

- [x] Move to a Mage based build
- [x] Single-file components (looks similar to .vue); .vugu -> .go code generation.
- [x] Includes CSS in components.
- [x] Basic flow control with vg-if, vg-for and output with vg-content.
- [x] Dynamic attributes with `<tag :prop='expr'>`.
- [x] Nested components with dynamic properties
- [x] Efficiently syncs to browser DOM.
- [x] Static HTML output (great for tests).
- [x] DOM Events, click, etc.
- [x] Modification tracking to avoid unnecessary computation where possible.
- [x] Basic dev and prod server tooling, easy to get started
- [x] Rewrite everything so it is not so terrible internally
- [x] URL Router (in-progress)
- [x] Tinygo compilation support
- [x] Server-side rendering (works, needs more documentation and examples)
- [x] Go-only component events
- [x] Slots
- [ ] Component library(s) (wip!)
- [ ] Performance optimizations
- And much more...

## Notes

It's built **more like a library than a framework**.  While Vugu does do code generation for your .vugu component
files, (and will even output a default main_wasm.go for a new project and build your program automatically upon page refresh), 
fundamentally you are still in control.  Overall program flow, application wiring and initialization, the render loop
that keeps the page in sync with your components - you have control over all of that.
Frameworks call your code.  Vugu is a library, your code calls it (even if Vugu generates a bit of that for you in
the beginning to make things easier). One of the primary goals for Vugu, when it comes to developers first encountering it, 
was to make it very fast and easy to get started, but without imposing unnecessary limitations on how a project is structured.
Go build tooling (and now the module system) is awesome.  The idea is to leverage that to the furthest extent possible,
rather than reprogramming the wheel.

So you won't find a vugu command line tool that runs a development server, instead
you'll find in the docs an appropriate snippet of code you can paste in a file and `go run` yourself.  For the code
generation while there is an http.Handler that can do this upon page refresh, you also can (and should!) run `vugugen`
via `go generate`. There are many small decisions in Vugu which follow this philosophy: wherever reasonably possible,
just use the existing mechanism instead of inventing anew.  And keep doing that until there's proof that something
else is really needed.  So far it's been working well.  And it allows Vugu to focus on the specific things it 
brings to the table.

## Examples

Examples of implementations can be found into [examples repositories](https://github.com/orgs/vugu-examples/repositories)

## VSCode plugin

As most of your code will be in `.vugu` files, you will need to install [vscode-vugu](https://marketplace.visualstudio.com/items?itemName=binhonglee.vscode-vugu)
Credits goes to @binhonglee.