package sh

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

func TestOutCmd(t *testing.T) {
	cmd := OutCmd(os.Args[0], "-printArgs", "foo", "bar")
	out, err := cmd("baz", "bat")
	if err != nil {
		t.Fatal(err)
	}
	expected := "[foo bar baz bat]"
	if out != expected {
		t.Fatalf("expected %q but got %q", expected, out)
	}
}

func TestExitCode(t *testing.T) {
	ran, err := Exec(nil, nil, nil, os.Args[0], "-helper", "-exit", "99")
	if err == nil {
		t.Fatal("unexpected nil error from run")
	}
	if !ran {
		t.Error("ran returned as false, but should have been true")
	}
	code := ExitStatus(err)
	if code != 99 {
		t.Fatalf("expected exit status 99, but got %v", code)
	}
}

func TestEnv(t *testing.T) {
	env := "SOME_REALLY_LONG_MAGEFILE_SPECIFIC_THING"
	out := &bytes.Buffer{}
	ran, err := Exec(map[string]string{env: "foobar"}, out, nil, os.Args[0], "-printVar", env)
	if err != nil {
		t.Fatalf("unexpected error from runner: %#v", err)
	}
	if !ran {
		t.Error("expected ran to be true but was false.")
	}
	if out.String() != "foobar\n" {
		t.Errorf("expected foobar, got %q", out)
	}
}

func TestNotRun(t *testing.T) {
	ran, err := Exec(nil, nil, nil, "thiswontwork")
	if err == nil {
		t.Fatal("unexpected nil error")
	}
	if ran {
		t.Fatal("expected ran to be false but was true")
	}
}

func TestAutoExpand(t *testing.T) {
	t.Setenv("MAGE_FOOBAR", "baz")
	s, err := Output("echo", "$MAGE_FOOBAR")
	if err != nil {
		t.Fatal(err)
	}
	if s != "baz" {
		t.Fatalf(`Expected "baz" but got %q`, s)
	}
}

func TestCmdRanNilErr(t *testing.T) {
	if !CmdRan(nil) {
		t.Fatal("CmdRan(nil) should return true")
	}
}

func TestCmdRanNotFound(t *testing.T) {
	_, err := Exec(nil, nil, nil, "thiswontwork")
	if CmdRan(err) {
		t.Fatal("CmdRan should return false for not-found command")
	}
}

func TestExitStatusNil(t *testing.T) {
	code := ExitStatus(nil)
	if code != 0 {
		t.Fatalf("expected 0 for nil error, got %d", code)
	}
}

func TestExitStatusNonExecError(t *testing.T) {
	code := ExitStatus(errors.New("generic error"))
	if code != 1 {
		t.Fatalf("expected 1 for generic error, got %d", code)
	}
}

func TestExitStatusFromExec(t *testing.T) {
	_, err := Exec(nil, nil, nil, os.Args[0], "-helper", "-exit", "42")
	code := ExitStatus(err)
	if code != 42 {
		t.Fatalf("expected exit status 42, got %d", code)
	}
}

func TestRunCmd(t *testing.T) {
	echoHello := RunCmd("echo", "hello")
	err := echoHello("world")
	// RunWith directs output based on verbose, so just check no error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOutputWith(t *testing.T) {
	out, err := OutputWith(map[string]string{"MY_TEST_VAR": "xyz"}, os.Args[0], "-printVar", "MY_TEST_VAR")
	if err != nil {
		t.Fatal(err)
	}
	if out != "xyz" {
		t.Fatalf("expected 'xyz', got %q", out)
	}
}
