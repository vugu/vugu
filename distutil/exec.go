package distutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Must panics if error.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustEnvExec is like MustExec but also sets specific environment variables.
func MustEnvExec(env2 []string, name string, arg ...string) string {

	env := os.Environ()

	b, err := exec.Command("go", "env", "GOPATH").CombinedOutput()
	if err != nil {
		panic(err)
	}
	goBinDir := filepath.Join(strings.TrimSpace(string(b)), "bin")
	_, err = os.Stat(goBinDir)
	if err == nil { // if dir exists, let's try adding it to PATH env
		for i := 0; i < len(env); i++ {
			if strings.HasPrefix(env[i], "PATH=") {
				env[i] = fmt.Sprintf("%s%c%s", env[i], os.PathListSeparator, goBinDir)
				goto donePath
			}
		}
		// no path... maybe we shoule add it?  pretty strange environment with no path, not putting anything here for now
	}
donePath:

	env = append(env, env2...)

	cmd := exec.Command(name, arg...)
	cmd.Env = env

	b, err = cmd.CombinedOutput()
	if err != nil {
		err2 := fmt.Errorf("error running: %s %v; err=%v; output:\n%s\n", name, arg, err, b)
		fmt.Print(err2)
		panic(err2)
	}

	return string(b)
}

// MustExec wraps exec.Command(...).CombinedOutput() with certain differences.
// If `go env GOPATH`/bin exists it is added to the PATH environment variable.
// Upon error the output of the command and the error will be printed and it will panic.
// Upon success the output is returned as a string.
func MustExec(name string, arg ...string) string {
	return MustEnvExec(nil, name, arg...)
}
