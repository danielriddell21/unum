// Package web provides the HTTP server for the unum diff web UI.
// All static assets are embedded at build time — the binary is fully self-contained.
package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/danielriddell21/unum/internal/diff/format"
	"github.com/danielriddell21/unum/internal/diff/node"
	"github.com/danielriddell21/unum/internal/diff/parse"
	"github.com/danielriddell21/unum/internal/web/shared"
)

//go:embed assets/*
var assets embed.FS

// Options configures the diff web server.
type Options struct {
	Port       int
	Quiet      bool
	DarkTheme  string // cyber | matrix | dracula | nord
	LightTheme string // clean | solarized
}

// Start launches the diff web server, auto-opens the browser, and blocks until
// Ctrl+C.
func Start(d *node.Diff, opts Options) error {
	host := "localhost"
	autoOpen := true
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			opts.Port = n
		}
		host = "0.0.0.0"
		autoOpen = false
	}

	port := opts.Port
	if port == 0 {
		var err error
		port, err = shared.FreePort()
		if err != nil {
			return fmt.Errorf("web: cannot find free port: %w", err)
		}
	}

	addr := host + ":" + strconv.Itoa(port)
	url := "http://" + addr

	payload := buildPayload(d)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("web: marshal payload: %w", err)
	}

	idx := shared.IndexData{DarkTheme: opts.DarkTheme, LightTheme: opts.LightTheme}
	mux := http.NewServeMux()
	mux.HandleFunc("/shared.css", shared.ServeSharedAsset("assets/shared.css", "text/css"))
	mux.HandleFunc("/shared.js", shared.ServeSharedAsset("assets/shared.js", "application/javascript"))
	mux.HandleFunc("/style.css", shared.ServeAsset(assets, "assets/style.css", "text/css"))
	mux.HandleFunc("/app.js", shared.ServeAsset(assets, "assets/app.js", "application/javascript"))
	mux.HandleFunc("/api/diff", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payloadBytes)
	})
	mux.HandleFunc("/", shared.ServeTemplate(assets, "assets/index.html")(idx))

	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}

	shared.PrintStartupBanner("diff viewer", url)

	if autoOpen {
		go shared.OpenBrowser(url)
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
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

func buildPayload(d *node.Diff) *diffPayload {
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
	return p
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

// ─── Server mode (no pre-loaded diff) ────────────────────────────────────────

// postDiffRequest is the body accepted by POST /api/diff in server mode.
type postDiffRequest struct {
	NameA    string `json:"nameA"`
	ContentA string `json:"contentA"`
	NameB    string `json:"nameB"`
	ContentB string `json:"contentB"`
	Format   string `json:"format"` // "json", "yaml", "terraform", "text", or "" for auto-detect
}

// StartServer launches the diff web server in input mode with no pre-loaded diff.
// GET /api/diff returns 204; POST /api/diff computes a diff from submitted content.
func StartServer(opts Options) error {
	host := "localhost"
	autoOpen := true
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			opts.Port = n
		}
		host = "0.0.0.0"
		autoOpen = false
	}

	port := opts.Port
	if port == 0 {
		var err error
		port, err = shared.FreePort()
		if err != nil {
			return fmt.Errorf("web: cannot find free port: %w", err)
		}
	}

	addr := host + ":" + strconv.Itoa(port)
	url := "http://" + addr

	d := shared.IndexData{DarkTheme: opts.DarkTheme, LightTheme: opts.LightTheme}
	mux := http.NewServeMux()
	mux.HandleFunc("/shared.css", shared.ServeSharedAsset("assets/shared.css", "text/css"))
	mux.HandleFunc("/shared.js", shared.ServeSharedAsset("assets/shared.js", "application/javascript"))
	mux.HandleFunc("/style.css", shared.ServeAsset(assets, "assets/style.css", "text/css"))
	mux.HandleFunc("/app.js", shared.ServeAsset(assets, "assets/app.js", "application/javascript"))
	mux.HandleFunc("/api/diff", handleServerDiff())
	mux.HandleFunc("/", shared.ServeTemplate(assets, "assets/index.html")(d))

	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}

	shared.PrintStartupBanner("diff viewer", url)

	if autoOpen {
		go shared.OpenBrowser(url)
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

func handleServerDiff() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}
		var req postDiffRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if req.ContentA == "" || req.ContentB == "" {
			http.Error(w, "contentA and contentB required", http.StatusBadRequest)
			return
		}

		dataA := []byte(req.ContentA)
		dataB := []byte(req.ContentB)
		fmt_ := format.Parse(req.Format, req.NameA, req.NameB)

		var diff *node.Diff
		switch fmt_ {
		case node.FormatJSON:
			diff, err = parse.JSON(dataA, dataB)
		case node.FormatYAML:
			diff, err = parse.YAML(dataA, dataB)
		case node.FormatTerraform:
			diff, err = parse.Terraform(dataA)
		default:
			diff, err = parse.Text(dataA, dataB, 3)
		}
		if err != nil {
			http.Error(w, "diff: "+err.Error(), http.StatusUnprocessableEntity)
			return
		}
		diff.FileA = req.NameA
		diff.FileB = req.NameB
		diff.Format = fmt_

		payload := buildPayload(diff)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
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
