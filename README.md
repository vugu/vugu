# Vugu

Vugu is an experimental library for web UIs written in Go and targeting webassembly.  Guide and docs at http://www.vugu.org.
Godoc at https://godoc.org/github.com/vugu/vugu.

If you've ever wanted to a UI not in JS but pure Go... and run it in your browser right now... That (experimental;) future is here!

Introducing Vugu, a VueJS-inspired library in Go targeting wasm.

No node. No JS. No npm. No node_modules folder competing with your music library for disk space.

* Runs in-browser using WebAssembly
* Single-file components
* Vue-like markup syntax
* Write idiomatic Go code
* Rapid prototyping
* ~3 minute setup
* Standard Go build tools

Get started: http://www.vugu.org/doc/start

Still a work in progress, but a lot of things are already functional. Some work really well.

Abbreviated Roadmap:
- [x] Single-file components (looks similar to .vue); .vugu -> .go code generation.
- [x] Includes CSS in components.
- [x] Basic flow control with vg-if, vg-for and output with vg-html.
- [x] Dynamic attributes with `&lt;tag :prop='expr'>`.
- [x] Nested components with dynamic properties `&lt;my-custom-component>`.
- [x] Efficently syncs to browser DOM.
- [x] Static HTML output (great for tests).
- [x] DOM Events, click, etc.
- [x] Basic data hashing to avoid unnecessary computation where possible.
- [x] Basic dev and prod server tooling, easy to get started
- [ ] URL Router
- [ ] Server-side rendering (HTML generation works, needs URL Router to make it usable)
- [ ] Performance optimizations
- [ ] Go-only component events
- And much more...
