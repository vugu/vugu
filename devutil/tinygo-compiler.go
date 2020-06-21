package devutil

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/vugu/vugu/distutil"
)

// DefaultTinygoDockerImage is used as the docker image for Tinygo unless overridden.
var DefaultTinygoDockerImage = "vugu/tinygo-dev:latest" // this is temporary until we can smooth things out with tinygo/tinygo:latest

// MustNewTinygoCompiler is like NewTinygoCompiler but panics upon error.
func MustNewTinygoCompiler() *TinygoCompiler {
	c, err := NewTinygoCompiler()
	if err != nil {
		panic(err)
	}
	return c
}

// NewTinygoCompiler returns a new TinygoCompiler instance.
func NewTinygoCompiler() (*TinygoCompiler, error) {
	tmpDir, err := ioutil.TempDir("", "TinygoCompiler")
	if err != nil {
		return nil, err
	}
	tmpDirAbs, err := filepath.Abs(tmpDir)
	if err != nil {
		return nil, err
	}
	tmpDirAbs, err = filepath.EvalSymlinks(tmpDirAbs) // Mac OS /var -> /var/private bs
	if err != nil {
		return nil, err
	}
	return &TinygoCompiler{
		logWriter:         os.Stderr,
		dlTmpGopath:       tmpDirAbs,
		tinygoDockerImage: DefaultTinygoDockerImage,
	}, nil
}

// TinygoCompiler provides a convenient way to build a program via Tinygo into Wasm.
// This implementation by default uses Docker to download and run Tinygo, and provides methods
// to handle mapping local directories into the Tinygo docker filesystem and for
// making other dependencies available by calling `go get` on them.  This approach
// might change once Tinygo has module support, but for now the idea is it
// makes it reasonably convenient to integration Tinygo into the workflow for Vugu app.
type TinygoCompiler struct {
	beforeFunc      func() error
	generateCmdFunc func() *exec.Cmd
	// buildCmdFunc    func(outpath string) *exec.Cmd
	buildDir          string // directory with main pkg that we are building with Tinygo
	afterFunc         func(outpath string, err error) error
	logWriter         io.Writer
	dlTmpGopath       string            // temporary directory that we download dependencies into with go get
	goGetCmdList      [][]string        // `go get` commands to be run before building with Tinygo
	tinygoDockerImage string            // docker image name to use, if empty then it is run directly
	wasmExecJS        []byte            // contents of wasm_exec.js
	pkgReplaceMap     map[string]string // package replacements pkgName->directory
	tinygoArgs        []string          // additional arguments to pass to the tinygo build cmd
}

// Close performs any cleanup.  For now it removes the temporary directory created by NewTinygoCompiler.
func (c *TinygoCompiler) Close() error {
	return os.RemoveAll(c.dlTmpGopath)
}

// SetTinygoArgs sets arguments to be passed to tinygo, e.g. -no-debug
func (c *TinygoCompiler) SetTinygoArgs(tinygoArgs ...string) *TinygoCompiler {
	c.tinygoArgs = tinygoArgs
	return c
}

// AddGoGet adds a go get command to the list of dependencies.  Arguments are separated by whitespace.
func (c *TinygoCompiler) AddGoGet(goGetCmdLine string) *TinygoCompiler {
	return c.AddGoGetArgs(strings.Fields(goGetCmdLine))
}

// AddGoGetArgs is like AddGoGet but the args are explicitly separated in a string slice.
func (c *TinygoCompiler) AddGoGetArgs(goGetCmdParts []string) *TinygoCompiler {
	c.goGetCmdList = append(c.goGetCmdList, goGetCmdParts)
	return c
}

// AddPkgReplace adds a directory mapping for a package.  It provides similar functionality
// to go.mod's replace statement, but is implemented with a docker volume mapping.
// The dir will be run through filepath.Abs before adding it and will panic if that fails.
func (c *TinygoCompiler) AddPkgReplace(pkgName, dir string) *TinygoCompiler {
	if c.pkgReplaceMap == nil {
		c.pkgReplaceMap = make(map[string]string, 2)
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	c.pkgReplaceMap[pkgName] = dir
	return c
}

// SetLogWriter sets the writer to use for logging output.  Setting it to nil disables logging.
// The default from NewCompiler is os.Stderr
func (c *TinygoCompiler) SetLogWriter(w io.Writer) *TinygoCompiler {
	if w == nil {
		w = ioutil.Discard
	}
	c.logWriter = w
	return c
}

// SetDir sets both the build and generate directories.
func (c *TinygoCompiler) SetDir(dir string) *TinygoCompiler {
	return c.SetBuildDir(dir).SetGenerateDir(dir)
}

// SetBuildDir sets the directory of the main package, where `go build` will be run.
// Relative paths are okay and will be resolved with filepath.Abs.
func (c *TinygoCompiler) SetBuildDir(dir string) *TinygoCompiler {

	c.buildDir = dir
	return c

	// return c.SetBuildCmdFunc(func(outpath string) *exec.Cmd {
	// 	cmd := exec.Command("go", "build", "-o", outpath)
	// 	cmd.Dir = dir
	// 	cmd.Env = os.Environ()
	// 	cmd.Env = append(cmd.Env, "GOOS=js", "GOARCH=wasm")
	// 	return cmd
	// })
}

// // SetBuildCmdFunc provides a function to create the exec.Cmd used when running `go build`.
// // It overrides any other build-related setting.
// func (c *TinygoCompiler) SetBuildCmdFunc(cmdf func(outpath string) *exec.Cmd) *TinygoCompiler {
// 	c.buildCmdFunc = cmdf
// 	return c
// }

// SetGenerateDir sets the directory of where `go generate` will be run.
// Relative paths are okay and will be resolved with filepath.Abs.
func (c *TinygoCompiler) SetGenerateDir(dir string) *TinygoCompiler {
	return c.SetGenerateCmdFunc(func() *exec.Cmd {
		cmd := exec.Command("go", "generate")
		cmd.Dir = dir
		return cmd
	})
}

// SetGenerateCmdFunc provides a function to create the exec.Cmd used when running `go generate`.
// It overrides any other generate-related setting.
func (c *TinygoCompiler) SetGenerateCmdFunc(cmdf func() *exec.Cmd) *TinygoCompiler {
	c.generateCmdFunc = cmdf
	return c
}

// SetBeforeFunc specifies a function to be executed before anything else during Execute().
func (c *TinygoCompiler) SetBeforeFunc(f func() error) *TinygoCompiler {
	c.beforeFunc = f
	return c
}

// SetAfterFunc specifies a function to be executed after everthing else during Execute().
func (c *TinygoCompiler) SetAfterFunc(f func(outpath string, err error) error) *TinygoCompiler {
	c.afterFunc = f
	return c
}

// NoDocker is an alias for SetTinygoDockerImage("") and will result in the tinygo
// executable being run on the local system instead of via docker image.
func (c *TinygoCompiler) NoDocker() *TinygoCompiler {
	return c.SetTinygoDockerImage("")
}

// SetTinygoDockerImage will specify the docker image to use when invoking Tinygo.
// The default value is the value of when NewTinygoCompiler was called.
// If you specify an empty string then the "tinygo" command will be run directly
// on the local system.
func (c *TinygoCompiler) SetTinygoDockerImage(img string) *TinygoCompiler {
	c.tinygoDockerImage = img
	return c
}

// Execute runs the generate command (if any) and then invokes the Tinygo compiler
// and produces a wasm executable (or an error).
// The value of outpath is the absolute path to the output file on disk.
// It will be created with a temporary name and if no error is returned
// it is the caller's responsibility to delete the file when it is no longer needed.
// If an error occurs during any of the steps it will be returned with (possibly multi-line)
// descriptive output in it's error message, as produced by the underlying tool.
func (c *TinygoCompiler) Execute() (outpath string, err error) {

	logerr := func(e error) error {
		if e == nil {
			return nil
		}
		fmt.Fprintln(c.logWriter, e)
		return e
	}

	if c.buildDir == "" {
		return "", logerr(errors.New("TinygoCompiler: no build directory set, cannot continue (did you forget to call SetBulidDir?)"))
	}

	if c.beforeFunc != nil {
		err := c.beforeFunc()
		if err != nil {
			return "", logerr(err)
		}
	}

	if c.generateCmdFunc != nil {
		cmd := c.generateCmdFunc()
		b, err := cmd.CombinedOutput()
		if err != nil {
			return "", logerr(fmt.Errorf("TinygoCompiler: generate error: %w; full output:\n%s", err, b))
		}
		fmt.Fprintln(c.logWriter, "TinygoCompiler: Successful generate")
	}

	// run go get stuff
	for _, cmdA := range c.goGetCmdList {
		cmd := exec.Command(cmdA[0], cmdA[1:]...)
		cmd.Dir = c.dlTmpGopath
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "GOPATH="+c.dlTmpGopath)
		cmd.Env = append(cmd.Env, "GO111MODULE=off")
		b, err := cmd.CombinedOutput()
		if err != nil {
			return "", logerr(fmt.Errorf("TinygoCompiler: generate error: %w; full output:\n%s", err, b))
		}
		fmt.Fprintf(c.logWriter, "TinygoCompiler: Successful cmd: %v; output: %s\n", cmdA, b)
	}

	buildDir := c.buildDir
	buildDirAbs, err := filepath.Abs(buildDir)
	if err != nil {
		return "", logerr(err)
	}
	buildDirAbs, err = filepath.EvalSymlinks(buildDirAbs) // Mac OS /var -> /var/private bs
	if err != nil {
		return "", logerr(err)
	}

	// detect module info
	modDir, modName, dirSuffix, err := detectMod(buildDirAbs)
	if err != nil {
		return "", logerr(fmt.Errorf("TinygoCompiler: %w", err))
	}

	tmpBin := filepath.Join(c.dlTmpGopath, "bin")
	os.Mkdir(tmpBin, 0755) // create $GOPATH/bin if not there already
	tgWasmOutF, err := ioutil.TempFile(tmpBin, "tgwasmout")
	if err != nil {
		return "", logerr(err)
	}
	tgWasmOutF.Close()
	outpath = tgWasmOutF.Name()

	// map any other package replacements
	var pkgReplaceList []string
	if l := len(c.pkgReplaceMap); l > 0 {
		pkgReplaceList = make([]string, 0, l)
		for pkgName := range c.pkgReplaceMap {
			pkgReplaceList = append(pkgReplaceList, pkgName)
		}
		sort.Strings(pkgReplaceList) // prevent the command line shifting around from run to run without any reason
	}

	tinygoDockerImage := c.tinygoDockerImage
	if tinygoDockerImage == "" {

		// run tinygo directly on local system, first we must build the appropriate folder structure

		// make a new directory
		tgTmpGopath, err2 := ioutil.TempDir("", "vugu-tinygo-")
		if err2 != nil {
			return "", err2
		}
		defer func() {
			if err == nil { // if build was successful, remove temporary directory, otherwise we leave it for debugging
				os.RemoveAll(tgTmpGopath)
			}
		}()

		// copy everything into it, so we have one nice folder structure to give to Tinygo
		allFiles := regexp.MustCompile(`.*`)
		// downloaded dependencies
		err = distutil.CopyDirFiltered(c.dlTmpGopath, tgTmpGopath, allFiles)
		if err != nil {
			return "", fmt.Errorf("error while copying dependencies: %w", err)
		}

		// replacement dirs
		for _, pkgName := range pkgReplaceList {
			dir := c.pkgReplaceMap[pkgName]
			err = distutil.CopyDirFiltered(dir,
				filepath.Join(tgTmpGopath, "src", pkgName),
				allFiles)
			if err != nil {
				return "", fmt.Errorf("error while copying replacement dir %q: %w", dir, err)
			}
		}

		// main project source
		err = distutil.CopyDirFiltered(modDir,
			filepath.Join(tgTmpGopath, "src", modName),
			allFiles)
		if err != nil {
			return "", fmt.Errorf("error while copying build dir %q: %w", modDir, err)
		}

		// now we can execute tinygo
		args := make([]string, 0, 20)
		args = append(args, "build")
		args = append(args, "-o", outpath)
		args = append(args, "-target=wasm")
		args = append(args, c.tinygoArgs...)
		args = append(args, path.Join(modName, dirSuffix))
		cmd := exec.Command("tinygo", args...)
		cmd.Dir = tgTmpGopath
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "GOPATH="+tgTmpGopath)
		cmd.Env = append(cmd.Env, "GO111MODULE=off")
		b, err := cmd.CombinedOutput()
		if err != nil {
			return "", logerr(fmt.Errorf("TinygoCompiler: build error (cmd=tinygo %v): %w; full output:\n%s", args, err, b))
		}
		fmt.Fprintf(c.logWriter, "TinygoCompiler: successful build: tinygo %v; output: %s\n", args, b)

	} else {

		// run tinygo via docker
		// example: docker run --rm -eGOPATH=/root/go
		// -v`pwd`/tmp1:/root/go
		// -v`pwd`:/root/go/src/example.com/tgtest1
		// tinygo/tinygo:0.13.1 tinygo build -o /root/go/src/example.com/tgtest1/out.wasm
		// -target=wasm example.com/tgtest1/testapp

		args := make([]string, 0, 20)
		args = append(args, "run", "--rm")
		args = append(args, "-e", "GOPATH=/root/go")
		args = append(args, "-v", c.dlTmpGopath+":/root/go")       // map dir for dependencies
		args = append(args, "-v", modDir+":/root/go/src/"+modName) // map dir for main module

		// map any other package replacements
		for _, pkgName := range pkgReplaceList {
			dir := c.pkgReplaceMap[pkgName]
			args = append(args, "-v", dir+":/root/go/src/"+pkgName)
		}

		args = append(args, tinygoDockerImage)
		args = append(args, "tinygo", "build")
		args = append(args, "-o", "/root/go/bin/"+filepath.Base(outpath))
		args = append(args, "-target=wasm")
		args = append(args, c.tinygoArgs...)
		args = append(args, path.Join(modName, dirSuffix))

		cmd := exec.Command("docker", args...)
		b, err := cmd.CombinedOutput()
		if err != nil {
			return "", logerr(fmt.Errorf("TinygoCompiler: build error (cmd=docker %v): %w; full output:\n%s", args, err, b))
		}
		fmt.Fprintf(c.logWriter, "TinygoCompiler: successful build: docker %v; output: %s\n", args, b)

	}

	return outpath, nil

}

// WasmExecJS returns the contents of the wasm_exec.js file bundled with Tinygo.
func (c *TinygoCompiler) WasmExecJS() (r io.Reader, err error) {

	if c.wasmExecJS != nil {
		return bytes.NewReader(c.wasmExecJS), nil
	}

	tinygoDockerImage := c.tinygoDockerImage

	// direct way, not via docker
	if tinygoDockerImage == "" {

		cmd := exec.Command("tinygo", "env", "TINYGOROOT")
		resb, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("TinygoCompiler: WasmExecJS error getting TINYGOROOT: %w; full output:\n%s", err, resb)
		}

		wasmExecJSPath := filepath.Join(strings.TrimSpace(string(resb)), "targets/wasm_exec.js")
		b, err := ioutil.ReadFile(wasmExecJSPath)
		if err != nil {
			return nil, fmt.Errorf("TinygoCompiler: WasmExecJS error reading %q: %w", wasmExecJSPath, err)
		}

		c.wasmExecJS = b

		return bytes.NewReader(c.wasmExecJS), nil

	}

	// via docker

	args := make([]string, 0, 20)
	args = append(args, "run", "--rm", "-i")
	args = append(args, tinygoDockerImage)
	args = append(args, "/bin/bash", "-c")
	// different locations between tinygo and tinygo-dev, check them both
	args = append(args, `if [ -f /usr/local/tinygo/targets/wasm_exec.js ]; then cat /usr/local/tinygo/targets/wasm_exec.js; else cat /tinygo/targets/wasm_exec.js; fi`)

	cmd := exec.Command("docker", args...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("TinygoCompiler: wasm_exec.js error (cmd=docker %v): %w; full output:\n%s", args, err, b)
	}
	// fmt.Fprintf(c.logWriter, "TinygoCompiler: successful wasm_exec.js: docker %v; output: %s\n", args, b)

	c.wasmExecJS = b

	return bytes.NewReader(c.wasmExecJS), nil

}

// detectMod returns useful module information about a directory.
// Given "/path/to/mymod/some/app", and /path/to/mymod/go.mod has
// `module example.com/thismod`, this method will return:
// "/path/to/mymod",
// "example.com/thismod"
// "some/app"
// The modDir is the directory where go.mod lives.
// path.Join(modName, dirSuffix) is the import path of the input dir.
// filepath.Join(modDir, dirSuffix) is the same as the input dir.
// An error will be returned if go.mod cannot be found, is unreadable or if
// some other filesystem error occurs.
// Go module versions greater than 1 are not supported.
func detectMod(dir string) (modDir, modName, dirSuffix string, reterr error) {

	modDir = dir

	for {
		f, err := os.Open(filepath.Join(modDir, "go.mod"))

		if os.IsNotExist(err) {
			dirSuffix = path.Join(filepath.Base(modDir), dirSuffix)
			newModDir := filepath.Join(modDir, "..")
			newModDir, err := filepath.Abs(newModDir)
			if err != nil {
				reterr = err
				return
			}
			if modDir == newModDir {
				reterr = fmt.Errorf("no go.mod found for dir: %s", dir)
				return
			}
			modDir = newModDir
			continue

		} else if err != nil {
			reterr = err
			return

		}

		defer f.Close()
		modName, err = readModuleEntry(f)
		if err != nil {
			reterr = err
			return
		}

		return
	}

	panic("unreachable")
}

func readModuleEntry(r io.Reader) (string, error) {

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	ret := modulePath(b)
	if ret == "" {
		return "", errors.New("unable to determine module path from go.mod")
	}

	return ret, nil
}

// shamelessly stolen from: https://github.com/golang/vgo/blob/master/vendor/cmd/go/internal/modfile/read.go#L837
// ModulePath returns the module path from the gomod file text.
// If it cannot find a module path, it returns an empty string.
// It is tolerant of unrelated problems in the go.mod file.
func modulePath(mod []byte) string {
	for len(mod) > 0 {
		line := mod
		mod = nil
		if i := bytes.IndexByte(line, '\n'); i >= 0 {
			line, mod = line[:i], line[i+1:]
		}
		if i := bytes.Index(line, slashSlash); i >= 0 {
			line = line[:i]
		}
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, moduleStr) {
			continue
		}
		line = line[len(moduleStr):]
		n := len(line)
		line = bytes.TrimSpace(line)
		if len(line) == n || len(line) == 0 {
			continue
		}

		if line[0] == '"' || line[0] == '`' {
			p, err := strconv.Unquote(string(line))
			if err != nil {
				return "" // malformed quoted string or multiline module path
			}
			return p
		}

		return string(line)
	}
	return "" // missing module path
}

var (
	slashSlash = []byte("//")
	moduleStr  = []byte("module")
)

/*

old notes:

- user can specify Dockerfile, we give sensible default, this is where `go get` cmds live
- option to force rebuild of dockerfile to re-`go get` for updated deps
- run then just makes a container from this image

docker run --rm -v $(pwd):/src tinygo/tinygo:0.13.1 tinygo build -o wasm.wasm -target=wasm examples/wasm/export
docker run --rm -v $(pwd):/src testimg1:latest tinygo build -o wasm.wasm -target=wasm examples/wasm/export

docker run --rm -ti tinygo/tinygo:0.13.1 /bin/bash

GOPATH=`pwd` GO111MODULE=off go get github.com/vugu/html

docker run --rm -v`pwd`/tmp1:/root/go -v`pwd`:/root/go/src/example.com/tgtest1 tinygo/tinygo:0.13.1 tinygo build -o /root/go/src/example.com/tgtest1/out.wasm -target=wasm example.com/tgtest1/testapp
docker run --rm -ti -v`pwd`/tmp1:/root/go -v`pwd`:/root/go/src/example.com/tgtest1 tinygo/tinygo:0.13.1 /bin/bash

docker run --rm -eGOPATH=/root/go -v`pwd`/tmp1:/root/go -v`pwd`:/root/go/src/example.com/tgtest1 tinygo/tinygo:0.13.1 tinygo build -o /root/go/src/example.com/tgtest1/out.wasm -target=wasm example.com/tgtest1/testapp
docker run --rm -eGOPATH=/root/go -v`pwd`/tmp1:/root/go -v`pwd`:/root/go/src/example.com/tgtest1 tinygo/tinygo-dev:latest tinygo build -o /root/go/src/example.com/tgtest1/out.wasm -target=wasm example.com/tgtest1/testapp


# download dependencies into /my-app-gopath
GOPATH=/my-app-gopath GO111MODULE=off go get github.com/vugu/html

docker run --rm
	-v/my-app-gopath:/root/go
	-v`pwd`:/root/go/src/example.com/tgtest1
	-v/out:/out
	tinygo/tinygo-dev:latest
	tinygo build -o /out/out.wasm -target=wasm example.com/tgtest1/testapp



I’m taking another stab at building a Vugu app with Tinygo and I run into an error about a missing package with the tinygo-dev image where things work as expected with tinygo:0.13.1.  Just wanted to check if this is known/expected:
# download dependencies into /my-app-gopath
GOPATH=/my-app-gopath GO111MODULE=off go get github.com/vugu/html

docker run --rm
	-v/my-app-gopath:/root/go
	-v`pwd`:/root/go/src/example.com/tgtest1
	-v/out:/out
	tinygo/tinygo-dev:latest
	tinygo build -o /out/out.wasm -target=wasm example.com/tgtest1/testapp
The idea is to download the dependencies in the host environment using go get and then map those, plus the application being compiled into the appropriate place under /root/go in the container - that way tinygo build ... my/package/path  should work as expected.  The above gives:

*/
