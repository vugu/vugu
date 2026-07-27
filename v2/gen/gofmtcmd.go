package gen

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func gofmt(pgm string) (string, error) {
	// build up command to run
	cmd := exec.Command("gofmt")

	// I need to capture output
	var fmtOutput bytes.Buffer
	cmd.Stderr = &fmtOutput
	cmd.Stdout = &fmtOutput

	// also set up input pipe
	read, write := io.Pipe()
	defer write.Close() // make sure this always gets closed, it is safe to call more than once
	cmd.Stdin = read

	// copy down environment variables
	cmd.Env = os.Environ()
	// force wasm,js target
	cmd.Env = append(cmd.Env, "GOOS=js")
	cmd.Env = append(cmd.Env, "GOARCH=wasm")

	// start gofmt
	if err := cmd.Start(); err != nil {
		return pgm, fmt.Errorf("can't run gofmt: %v", err)
	}

	// stream in the raw source
	if _, err := write.Write([]byte(pgm)); err != nil && err != io.ErrClosedPipe {
		return pgm, fmt.Errorf("gofmt failed: %v", err)
	}

	write.Close()

	// wait until gofmt is done
	if err := cmd.Wait(); err != nil {
		return pgm, fmt.Errorf("go fmt error %v; full output: %s", err, fmtOutput.String())
	}

	return fmtOutput.String(), nil
}
