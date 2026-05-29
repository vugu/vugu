/*
Package js is a drop-in replacement for syscall/js that provides identical behavior in a WebAssembly
environment, and useful non-functional behavior outside of WebAssembly.

To use it, simply import this package instead of "syscall/js" and use it in exactly the same way.
Your code will compile targeting either wasm or non-wasm environments.  In wasm, the functionality
is exactly the same, the calls and types are delegated directly to "syscall/js".  The compiler
will optimize away virtually (if not literally) all of the overhead associated with this aliasing.
When run outside of wasm, appropriate functionality indicates that the environment is not
availble: All js.Value instances are undefined.  Global(), Value.Get() always return undefined.
Value.Call(), FuncOf() and other such functions will panic.  Value.Truthy() will always return false.
For example, Global().Truthy() can be used to determine if the js environment is functional.

Rationale: Since syscall/js is only available when the GOOS is "js", this means programs which run server-side
cannot access that package and will fail to compile.  Since Vugu components are inherently closely
integrated with browsers and may often need to do things like declare variables of type js.Value,
this is a problem for components which are rendered not only in wasm client-side but also server-side.
Build tags can be used to provide multiple implementations of a components but this can become tedious.

Usually what you want is that the majority of your code which is not js-specific can be written once
and execute in the browser or on the server, and the relatively small amount of functionality that
is uses "js" will compile properly in both environments but just not be executed on the server.
Or allow for if statements to easily disable functionality not available server-side.
That's what this package provides.
*/
package js
