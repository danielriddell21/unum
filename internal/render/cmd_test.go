package rendertool_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/unum/internal/config"
	rendertool "github.com/danielriddell21/unum/internal/render"
	"github.com/danielriddell21/unum/internal/telemetry"
)

func testCommand(t *testing.T) *cobra.Command {
	t.Helper()
	noColor, quiet := true, true
	return rendertool.Command(&noColor, &quiet, "test", telemetry.Init(config.Config{}, "test", "0"))
}

func writeSource(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCommandWritesOutput(t *testing.T) {
	d2 := writeSource(t, "s.d2", "a -> b\nb -> c\n")
	mmd := writeSource(t, "s.mmd", "flowchart TD\n A --> B\n")
	cases := []struct{ name, file, format string }{
		{"d2-svg", d2, "svg"},
		{"d2-drawio", d2, "drawio"},
		{"d2-png", d2, "png"},
		{"mermaid-svg", mmd, "svg"},
		{"mermaid-drawio", mmd, "drawio"},
		{"mermaid-png", mmd, "png"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "out")
			cmd := testCommand(t)
			cmd.SetArgs([]string{tc.file, "--format", tc.format, "-o", out})
			if err := cmd.Execute(); err != nil {
				t.Fatalf("execute: %v", err)
			}
			b, err := os.ReadFile(out)
			if err != nil {
				t.Fatalf("read output: %v", err)
			}
			if len(b) == 0 {
				t.Error("output file is empty")
			}
		})
	}
}

func TestCommandErrors(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"unknown-language", []string{writeSource(t, "x.txt", "hi")}},
		{"bad-format", []string{writeSource(t, "y.d2", "a -> b"), "--format", "pdf"}},
		{"missing-file", []string{"does-not-exist.d2"}},
		{"forced-lang-render-error", []string{writeSource(t, "z.d2", "a ->"), "--lang", "d2"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := testCommand(t)
			cmd.SetArgs(tc.args)
			cmd.SilenceErrors = true
			if err := cmd.Execute(); err == nil {
				t.Errorf("expected an error for %s", tc.name)
			}
		})
	}
}
