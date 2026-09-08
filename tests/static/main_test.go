package static_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"
)

var unumBin string

var configDir string

func TestMain(m *testing.M) {
	if err := os.Chdir("../.."); err != nil {
		fmt.Fprintf(os.Stderr, "chdir to repo root: %v\n", err)
		os.Exit(1)
	}

	bin, err := buildBinary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	unumBin = bin

	configDir, err = os.MkdirTemp("", "unum-cli-config-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "temp config dir: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	_ = os.RemoveAll(configDir)
	_ = os.Remove(unumBin)
	os.Exit(code)
}

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

	cmd := exec.Command("go", "build", "-o", f.Name(), "./cmd/unum") //nolint:noctx // test build helper; no cancellation needed for a one-shot compile
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%w\n%s", err, out)
	}
	return f.Name(), nil
}

func run(args ...string) (stdout, stderr string, code int) {
	return runEnv(nil, args...)
}

// runEnv keeps every invocation pointed at a throwaway config dir with
// telemetry off, so tests never touch the developer's own unum config or
// hash history.
func runEnv(extraEnv []string, args ...string) (stdout, stderr string, code int) {
	cmd := exec.Command(unumBin, args...) //nolint:noctx // test runner invoking compiled binary; context not threaded through functional tests
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+configDir,
		"APPDATA="+configDir,
		"DO_NOT_TRACK=1",
		"PORT=",
		"UNUM_BIND=",
		"UNUM_ENV=",
	)
	cmd.Env = append(cmd.Env, extraEnv...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			code = exitErr.ExitCode()
		} else {
			code = -1
		}
	}
	return
}
