// Package tests contains functional CLI tests that exercise the compiled unum
// binary end-to-end. Run with: go test ./tests/...
package tests_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"
)

var unumBin string

func TestMain(m *testing.M) {
	// Tests run from the tests/ directory; step up to the repo root so that
	// ./cmd/unum and testdata/ resolve correctly.
	if err := os.Chdir(".."); err != nil {
		fmt.Fprintf(os.Stderr, "chdir to repo root: %v\n", err)
		os.Exit(1)
	}

	bin, err := buildBinary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	unumBin = bin

	code := m.Run()
	_ = os.Remove(unumBin)
	os.Exit(code)
}

// buildBinary compiles the unum binary to a temp file and returns its path.
func buildBinary() (string, error) {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	f, err := os.CreateTemp("", "unum-cli-test-*"+ext)
	if err != nil {
		return "", err
	}
	_ = f.Close()

	cmd := exec.Command("go", "build", "-o", f.Name(), "./cmd/unum")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%w\n%s", err, out)
	}
	return f.Name(), nil
}

// run executes the unum binary with the given args and returns stdout, stderr, and exit code.
func run(args ...string) (stdout, stderr string, code int) {
	cmd := exec.Command(unumBin, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			code = -1
		}
	}
	return
}
