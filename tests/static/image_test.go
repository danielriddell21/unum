package static_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	sampleJPG = filepath.Join("testdata", "sample.jpg")
	samplePNG = filepath.Join("testdata", "sample.png")
)

func TestImageReport(t *testing.T) {
	stdout, _, code := run("image", sampleJPG, flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}

	for _, want := range []string{"source", "format", "jpeg", "dimensions", "1200 × 800", "quality", "saving"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("report missing %q:\n%s", want, stdout)
		}
	}
}

func TestImageReportWritesNoFile(t *testing.T) {
	before, err := os.Stat(sampleJPG)
	if err != nil {
		t.Fatalf("stat fixture: %v", err)
	}

	if _, _, code := run("image", sampleJPG, flagNoColor); code != 0 {
		t.Fatalf(exitFmt, code)
	}

	after, err := os.Stat(sampleJPG)
	if err != nil {
		t.Fatalf("stat fixture: %v", err)
	}
	if before.Size() != after.Size() {
		t.Error("running without --output must not touch the source file")
	}
}

func TestImageNoColor(t *testing.T) {
	stdout, stderr, code := run("image", sampleJPG, flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if strings.Contains(stdout, "\x1b[") || strings.Contains(stderr, "\x1b[") {
		t.Error("--no-color output contains ANSI escape codes")
	}
}

func TestImageWritesSmallerFile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "small.jpg")

	stdout, _, code := run("image", sampleJPG, "-o", out, "-q", "60", flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}

	written, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.HasPrefix(written, []byte{0xFF, 0xD8, 0xFF}) {
		t.Error("output is not a jpeg")
	}

	source, err := os.Stat(sampleJPG)
	if err != nil {
		t.Fatalf("stat source: %v", err)
	}
	if int64(len(written)) >= source.Size() {
		t.Errorf("output is %d bytes, no smaller than the %d byte source", len(written), source.Size())
	}
	if !strings.Contains(stdout, "small.jpg") {
		t.Errorf("summary should name the output file:\n%s", stdout)
	}
}

func TestImageConvertsFormatFromOutputExtension(t *testing.T) {
	out := filepath.Join(t.TempDir(), "converted.png")

	if _, _, code := run("image", sampleJPG, "-o", out, flagNoColor); code != 0 {
		t.Fatalf(exitFmt, code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.HasPrefix(data, []byte{0x89, 'P', 'N', 'G'}) {
		t.Error("output should be a png, inferred from the -o extension")
	}
}

func TestImageExplicitFormatFlag(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.bin")

	if _, _, code := run("image", samplePNG, "-o", out, "--to", "jpeg", flagNoColor); code != 0 {
		t.Fatalf(exitFmt, code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
		t.Error("--to jpeg should win over the output extension")
	}
}

func TestImageWebPOutput(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.webp")

	if _, _, code := run("image", samplePNG, "-o", out, flagNoColor); code != 0 {
		t.Fatalf(exitFmt, code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("RIFF")) || !bytes.Contains(data[:16], []byte("WEBP")) {
		t.Errorf("output is not a webp container: % x", data[:min(16, len(data))])
	}
}

// Lossless webp should undercut the png it came from.
func TestImageWebPBeatsPNG(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.webp")

	if _, _, code := run("image", samplePNG, "-o", out, "--to", "webp", "-q", "100", flagNoColor); code != 0 {
		t.Fatalf(exitFmt, code)
	}

	written, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}
	source, err := os.Stat(samplePNG)
	if err != nil {
		t.Fatalf("stat source: %v", err)
	}
	if written.Size() >= source.Size() {
		t.Errorf("lossless webp is %d bytes, no smaller than the %d byte png", written.Size(), source.Size())
	}
}

func TestImageScale(t *testing.T) {
	out := filepath.Join(t.TempDir(), "half.jpg")

	stdout, _, code := run("image", sampleJPG, "-o", out, "--scale", "50%", flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if !strings.Contains(stdout, "600 × 400") {
		t.Errorf("expected half-size dimensions in the summary:\n%s", stdout)
	}
}

func TestImageMaxWidthKeepsAspect(t *testing.T) {
	out := filepath.Join(t.TempDir(), "thumb.jpg")

	stdout, _, code := run("image", sampleJPG, "-o", out, "--max-width", "300", flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if !strings.Contains(stdout, "300 × 200") {
		t.Errorf("expected the aspect ratio to be kept:\n%s", stdout)
	}
}

func TestImageTargetSize(t *testing.T) {
	out := filepath.Join(t.TempDir(), "budget.jpg")
	const budget = 30000

	if _, _, code := run("image", sampleJPG, "-o", out, "--target", "30kb", flagNoColor); code != 0 {
		t.Fatalf(exitFmt, code)
	}

	info, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}
	if info.Size() > budget {
		t.Errorf("output is %d bytes, over the %d byte budget", info.Size(), budget)
	}
	// A budget that is reachable should not be undershot by a wide margin.
	if info.Size() < budget/2 {
		t.Errorf("output is only %d bytes against a %d byte budget — quality was left on the table",
			info.Size(), budget)
	}
}

func TestImageTargetTooSmallStillWritesAFile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "tiny.jpg")

	stdout, _, code := run("image", sampleJPG, "-o", out, "--target", "1kb", flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("expected a best-effort file: %v", err)
	}
	if !strings.Contains(stdout, "could not reach") {
		t.Errorf("expected a note that the budget was missed:\n%s", stdout)
	}
}

func TestImagePNGQuantization(t *testing.T) {
	out := filepath.Join(t.TempDir(), "small.png")

	if _, _, code := run("image", samplePNG, "-o", out, "-q", "50", flagNoColor); code != 0 {
		t.Fatalf(exitFmt, code)
	}

	written, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}
	source, err := os.Stat(samplePNG)
	if err != nil {
		t.Fatalf("stat source: %v", err)
	}
	if written.Size() >= source.Size() {
		t.Errorf("quantized png is %d bytes, no smaller than the %d byte source",
			written.Size(), source.Size())
	}
}

func TestImageQuietSuppressesBootLine(t *testing.T) {
	_, stderr, code := run("image", sampleJPG, "--quiet", flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if strings.Contains(stderr, "[ UNUM ]") {
		t.Errorf("--quiet should suppress the boot line, got %q", stderr)
	}
}

func TestImageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"missing file", []string{"image", "nonexistent.jpg"}},
		{"not an image", []string{"image", filepath.Join("testdata", "sample.json")}},
		{"unknown output format", []string{"image", sampleJPG, "--to", "avif"}},
		{"scale above full size", []string{"image", sampleJPG, "--scale", "200%"}},
		{"scale of zero", []string{"image", sampleJPG, "--scale", "0"}},
		{"unparseable target", []string{"image", sampleJPG, "--target", "biggish"}},
		{"no file without web", []string{"image"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, stderr, code := run(append(tt.args, flagNoColor)...)
			if code == 0 {
				t.Fatal("expected a non-zero exit")
			}
			if stderr == "" {
				t.Error("expected an error message on stderr")
			}
		})
	}
}
