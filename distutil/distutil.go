/*
Package distutil has some useful functions for building your Vugu application's distribution

Rather than introducing third party build tools.  Authors of Vugu-based applications are
encouraged to build their distribution files (output which is run on the production server)
using a simple .go file which can be "go run"ed.  This package makes some common tasks simpler:

Copying a directory of static files from one location to another.  The destination directory
can safely be a child of the source directory.  Files which are up to date are not re-copied, for speed.

	// by default uses DefaultFileInclPattern, matches common static file extensions
	distutil.MustCopyDirFiltered(fromDir, toDir, nil)

You can also provide your own pattern to say which files to copy:

	distutil.MustCopyDirFiltered(fromDir, toDir, regexp.MustCompile(`[.](css|js|map|jpg|png|wasm)$`)

File the wasm_exec.js file included with your Go distribution and copy that:

	distutil.MustCopyFile(distutil.MustWasmExecJsPath(), filepath.Join(toDir, "wasm_exec.js"))

Run a command and automatically include $GOPATH/bin (defaults to $HOME/go/bin) to $PATH.
This makes it easy to ensure tools installed by "go get" are available during "go generate".
(The output of the command is returned as a string, panics on error.)

	fmt.Print(distutil.MustExec("go", "generate", "."))

Executing a command while overriding certain environment variables is also easy:

	fmt.Print(distutil.MustEnvExec(
		[]string{"GOOS=js", "GOARCH=wasm"},
		"go", "build", "-o", filepath.Join(outDir, "main.wasm"), "."))

*/
package distutil
