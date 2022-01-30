package devutil

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DefaultTinygoDockerImage is used as the docker image for Tinygo unless overridden.
var DefaultTinygoDockerImage = "tinygo/tinygo:0.22.0"

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
	return &TinygoCompiler{
		logWriter:         os.Stderr,
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
	beforeFunc         func() error
	generateCmdFunc    func() *exec.Cmd
	buildCmdFunc       func(outpath string) *exec.Cmd
	dockerBuildCmdFunc func(outpath string) *exec.Cmd
	afterFunc          func(outpath string, err error) error
	logWriter          io.Writer
	dlTmpGopath        string   // temporary directory that we download dependencies into with go get
	tinygoDockerImage  string   // docker image name to use, if empty then it is run directly
	wasmExecJS         []byte   // contents of wasm_exec.js
	tinygoArgs         []string // additional arguments to pass to the tinygo build cmd
}

// Close performs any cleanup.  For now it removes the temporary directory created by NewTinygoCompiler.
func (c *TinygoCompiler) Close() error {
	return nil
}

// SetTinygoArgs sets arguments to be passed to tinygo, e.g. -no-debug
func (c *TinygoCompiler) SetTinygoArgs(tinygoArgs ...string) *TinygoCompiler {
	c.tinygoArgs = tinygoArgs
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
	return c.SetBuildCmdFunc(func(outpath string) *exec.Cmd {
		cmd := exec.Command("tinygo", "build", "-target=wasm", "-o", outpath, ".")
		cmd.Dir = dir
		return cmd
	}).SetDockerBuildCmdFunc(func(outpath string) *exec.Cmd {
		buildDir := dir
		buildDirAbs, err := filepath.Abs(buildDir)
		if err != nil {
			panic(err)
		}
		buildDirAbs, err = filepath.EvalSymlinks(buildDirAbs) // Mac OS /var -> /var/private bs
		if err != nil {
			panic(err)
		}

		tinygoDockerImage := c.tinygoDockerImage

		// run tinygo via docker
		// example: docker run --rm \
		// -v /:/src \
		// -w /src/`pwd` \
		// tinygo/tinygo:0.22.0 tinygo build -o /root/go/src/example.com/tgtest1/out.wasm \
		// -target=wasm .

		args := make([]string, 0, 20)
		args = append(args, "run", "--rm")
		args = append(args, "-v", "/:/src") // map dir for dependencies
		args = append(args, "-e", "HOME=/tmp")
		args = append(args, fmt.Sprintf("--user=%d", os.Getuid()))
		args = append(args, "-w", "/src"+buildDirAbs)

		args = append(args, tinygoDockerImage)
		args = append(args, "tinygo", "build")
		args = append(args, "-o", "/src/"+outpath)
		args = append(args, "-target=wasm")
		args = append(args, c.tinygoArgs...)
		args = append(args, ".")

		return exec.Command("docker", args...)
	})
}

// SetBuildCmdFunc provides a function to create the exec.Cmd used when running `go build`.
// It overrides any other build-related setting.
func (c *TinygoCompiler) SetBuildCmdFunc(cmdf func(outpath string) *exec.Cmd) *TinygoCompiler {
	c.buildCmdFunc = cmdf
	return c
}

// SetDockerBuildCmdFunc provides a function to create the exec.Cmd used when running
// `tinygo build` in docker.
func (c *TinygoCompiler) SetDockerBuildCmdFunc(cmdf func(outpath string) *exec.Cmd) *TinygoCompiler {
	c.dockerBuildCmdFunc = cmdf
	return c
}

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

	if c.buildCmdFunc == nil {
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

	tmpf, err := ioutil.TempFile("", "WasmCompiler")
	if err != nil {
		return "", logerr(fmt.Errorf("WasmCompiler: error creating temporary file: %w", err))
	}

	outpath = tmpf.Name()

	err = tmpf.Close()
	if err != nil {
		return outpath, logerr(fmt.Errorf("WasmCompiler: error closing temporary file: %w", err))
	}

	os.Remove(outpath)

	if c.tinygoDockerImage == "" {
		cmd := c.buildCmdFunc(outpath)
		b, err := cmd.CombinedOutput()
		if err != nil {
			return "", logerr(fmt.Errorf("TinygoCompiler: build error: %w; cmd.args: %v, full output:\n%s", err, cmd.Args, b))
		}
		fmt.Fprintf(c.logWriter, "TinygoCompiler: successful build\n")

	} else {
		cmd := c.dockerBuildCmdFunc(outpath)
		b, err := cmd.CombinedOutput()
		if err != nil {
			return "", logerr(fmt.Errorf("TinygoCompiler: build error: %w; cmd.args: %v, full output:\n%s", err, cmd.Args, b))
		}
		fmt.Fprintf(c.logWriter, "TinygoCompiler: successful build. Output: %s\n", b)
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
	args = append(args, "cat `tinygo env TINYGOROOT`/targets/wasm_exec.js")

	cmd := exec.Command("docker", args...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("TinygoCompiler: wasm_exec.js error (cmd=docker %v): %w; full output:\n%s", args, err, b)
	}
	// fmt.Fprintf(c.logWriter, "TinygoCompiler: successful wasm_exec.js: docker %v; output: %s\n", args, b)

	c.wasmExecJS = b

	return bytes.NewReader(c.wasmExecJS), nil

}
