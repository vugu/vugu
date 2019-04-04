package vugufmt

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// UseGoFmt sets the formatter to use gofmt on x-go blocks.
// Set simplifyAST to true to simplify the AST. This is false
// by default for gofmt, and is the same as passing in -s for it.
func UseGoFmt(simplifyAST bool) func(*Formatter) {

	return func(f *Formatter) {
		f.ScriptFormatters["application/x-go"] = func(input []byte) ([]byte, *FmtError) {
			return runGoFmt(input, simplifyAST)
		}
	}
}

func runGoFmt(input []byte, simplify bool) ([]byte, *FmtError) {
	// build up command to run
	cmd := exec.Command("gofmt")

	if simplify {
		cmd.Args = []string{"-s"}
	}

	var resBuff, errBuff bytes.Buffer

	// I need to capture output
	cmd.Stderr = &errBuff
	cmd.Stdout = &resBuff

	// also set up input pipe
	cmd.Stdin = bytes.NewReader(input)

	// copy down environment variables
	cmd.Env = os.Environ()

	// start gofmt
	if err := cmd.Start(); err != nil {
		return input, &FmtError{Msg: fmt.Sprintf("can't run gofmt: %s", err.Error())}
	}

	// wait until gofmt is done
	err := cmd.Wait()

	// Get all the output
	res := resBuff.Bytes()

	// Wrap the output in an error.
	if err != nil {
		return input, fromGoFmt(string(errBuff.String()))
	}

	return res, nil
}
