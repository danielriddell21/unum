// Package web provides the HTTP server for the unum diff web UI.
// All static assets are embedded at build time — the binary is fully self-contained.
package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/danielriddell21/unum/internal/diff/node"
)

//go:embed assets/*
var assets embed.FS

// Options configures the diff web server.
type Options struct {
	Port  int
	Quiet bool
}

// Start launches the diff web server, auto-opens the browser, and blocks until
// Ctrl+C.
func Start(d *node.Diff, opts Options) error {
	port := opts.Port
	if port == 0 {
		var err error
		port, err = freePort()
		if err != nil {
			return fmt.Errorf("web: cannot find free port: %w", err)
		}
	}

	addr := "localhost:" + strconv.Itoa(port)
	url := "http://" + addr

	payload, err := buildPayload(d)
	if err != nil {
		return fmt.Errorf("web: build payload: %w", err)
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("web: marshal payload: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/style.css", serveAsset("assets/style.css", "text/css"))
	mux.HandleFunc("/app.js", serveAsset("assets/app.js", "application/javascript"))
	mux.HandleFunc("/api/diff", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payloadBytes)
	})
	mux.HandleFunc("/", serveAsset("assets/index.html", "text/html"))

	srv := &http.Server{Addr: addr, Handler: mux}

	_, _ = fmt.Fprintf(os.Stderr, "\033[38;5;51m[ UNUM ] DIFF VIEWER LIVE → %s\033[0m\n", url)
	_, _ = fmt.Fprintf(os.Stderr, "\033[38;5;240mPress Ctrl+C to stop\033[0m\n")

	go openBrowser(url)

	return srv.ListenAndServe()
}

func serveAsset(path, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := assets.ReadFile(path)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", contentType)
		_, _ = w.Write(data)
	}
}

func openBrowser(url string) {
	time.Sleep(300 * time.Millisecond)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// ─── Payload ──────────────────────────────────────────────────────────────────

type diffPayload struct {
	FileA    string       `json:"fileA"`
	FileB    string       `json:"fileB"`
	Format   string       `json:"format"`
	Added    int          `json:"added"`
	Removed  int          `json:"removed"`
	Modified int          `json:"modified"`
	Hunks    []webHunk    `json:"hunks"`
	Tree     *webDiffNode `json:"tree,omitempty"`
}

type webDiffNode struct {
	Kind     string         `json:"kind"`
	Path     string         `json:"path"`
	Key      string         `json:"key,omitempty"`
	Index    int            `json:"index"`
	OldValue string         `json:"oldValue,omitempty"`
	NewValue string         `json:"newValue,omitempty"`
	Children []*webDiffNode `json:"children,omitempty"`
}

type webHunk struct {
	OldStart int       `json:"oldStart"`
	OldCount int       `json:"oldCount"`
	NewStart int       `json:"newStart"`
	NewCount int       `json:"newCount"`
	Lines    []webLine `json:"lines"`
}

type webLine struct {
	Kind    string `json:"kind"`
	OldNum  int    `json:"oldNum"`
	NewNum  int    `json:"newNum"`
	Content string `json:"content"`
}

func buildPayload(d *node.Diff) (*diffPayload, error) {
	p := &diffPayload{
		FileA:    d.FileA,
		FileB:    d.FileB,
		Format:   formatString(d.Format),
		Added:    d.Added,
		Removed:  d.Removed,
		Modified: d.Modified,
	}
	if d.Root != nil {
		p.Tree = toWebDiffNode(d.Root)
	}
	for _, h := range d.Hunks {
		wh := webHunk{
			OldStart: h.OldStart,
			OldCount: h.OldCount,
			NewStart: h.NewStart,
			NewCount: h.NewCount,
		}
		for _, l := range h.Lines {
			wh.Lines = append(wh.Lines, webLine{
				Kind:    kindString(l.Kind),
				OldNum:  l.OldNum,
				NewNum:  l.NewNum,
				Content: l.Content,
			})
		}
		p.Hunks = append(p.Hunks, wh)
	}
	return p, nil
}

func toWebDiffNode(dn *node.DiffNode) *webDiffNode {
	if dn == nil {
		return nil
	}
	wn := &webDiffNode{
		Kind:     changeKindString(dn.Kind),
		Path:     dn.Path,
		Key:      dn.Key,
		Index:    dn.Index,
		OldValue: dn.OldValue,
		NewValue: dn.NewValue,
	}
	for _, c := range dn.Children {
		wn.Children = append(wn.Children, toWebDiffNode(c))
	}
	return wn
}

func changeKindString(k node.ChangeKind) string {
	switch k {
	case node.Added:
		return "added"
	case node.Removed:
		return "removed"
	case node.Modified:
		return "modified"
	default:
		return "unchanged"
	}
}

func kindString(k node.ChangeKind) string {
	switch k {
	case node.Added:
		return "added"
	case node.Removed:
		return "removed"
	default:
		return "unchanged"
	}
}

func formatString(f node.Format) string {
	switch f {
	case node.FormatJSON:
		return "json"
	case node.FormatYAML:
		return "yaml"
	case node.FormatTerraform:
		return "terraform"
	default:
		return "text"
	}
}
