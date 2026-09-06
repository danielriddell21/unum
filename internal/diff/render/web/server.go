package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/danielriddell21/unum/internal/diff/format"
	"github.com/danielriddell21/unum/internal/diff/node"
	"github.com/danielriddell21/unum/internal/diff/parse"
	"github.com/danielriddell21/unum/internal/telemetry"
	"github.com/danielriddell21/unum/internal/web/shared"
)

//go:embed assets/*
var assets embed.FS

const (
	contentTypeCSS = "text/css"
	contentTypeJS  = "application/javascript"
	apiDiffPath    = "/api/diff"
)

type Options struct {
	Port       int
	Quiet      bool
	DarkTheme  string
	LightTheme string
	Version    string
	Tel        *telemetry.Telemetry
}

func Start(d *node.Diff, opts Options) error {
	bind, err := shared.ResolveBind(opts.Port)
	if err != nil {
		return fmt.Errorf("web: %w", err)
	}
	addr := bind.Addr()
	url := bind.URL()

	payload := buildPayload(d)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("web: marshal payload: %w", err)
	}

	idx := shared.NewIndexData(opts.DarkTheme, opts.LightTheme, opts.Version)
	mux := http.NewServeMux()
	mux.HandleFunc("/shared.css", shared.ServeSharedAsset("assets/shared.css", contentTypeCSS))
	mux.HandleFunc("/shared.js", shared.ServeSharedAsset("assets/shared.js", contentTypeJS))
	mux.HandleFunc("/style.css", shared.ServeAsset(assets, "assets/style.css", contentTypeCSS))
	mux.HandleFunc("/app.js", shared.ServeAsset(assets, "assets/app.js", contentTypeJS))
	mux.HandleFunc(apiDiffPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleServerDiff(opts.Tel)(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payloadBytes)
	})
	mux.HandleFunc("/", shared.ServeTemplate(assets, "assets/index.html")(idx))
	shared.RegisterMetrics(mux, bind)
	shared.RegisterUmamiProxy(mux)

	srv := &http.Server{Addr: addr, Handler: otelhttp.NewHandler(mux, "unum-diff"), ReadHeaderTimeout: 10 * time.Second}

	shared.PrintBindWarning(bind)
	shared.PrintStartupBanner("diff viewer", url)

	if bind.AutoOpen {
		go shared.OpenBrowser(url)
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

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

type postDiffRequest struct {
	NameA    string `json:"nameA"`
	ContentA string `json:"contentA"`
	NameB    string `json:"nameB"`
	ContentB string `json:"contentB"`
	Format   string `json:"format"`
}

func StartServer(opts Options) error {
	bind, err := shared.ResolveBind(opts.Port)
	if err != nil {
		return fmt.Errorf("web: %w", err)
	}
	addr := bind.Addr()
	url := bind.URL()

	d := shared.NewIndexData(opts.DarkTheme, opts.LightTheme, opts.Version)
	mux := http.NewServeMux()
	mux.HandleFunc("/shared.css", shared.ServeSharedAsset("assets/shared.css", contentTypeCSS))
	mux.HandleFunc("/shared.js", shared.ServeSharedAsset("assets/shared.js", contentTypeJS))
	mux.HandleFunc("/style.css", shared.ServeAsset(assets, "assets/style.css", contentTypeCSS))
	mux.HandleFunc("/app.js", shared.ServeAsset(assets, "assets/app.js", contentTypeJS))
	mux.HandleFunc(apiDiffPath, handleServerDiff(opts.Tel))
	mux.HandleFunc("/", shared.ServeTemplate(assets, "assets/index.html")(d))
	shared.RegisterMetrics(mux, bind)
	shared.RegisterUmamiProxy(mux)

	srv := &http.Server{Addr: addr, Handler: otelhttp.NewHandler(mux, "unum-diff"), ReadHeaderTimeout: 10 * time.Second}

	shared.PrintBindWarning(bind)
	shared.PrintStartupBanner("diff viewer", url)

	if bind.AutoOpen {
		go shared.OpenBrowser(url)
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

func handleServerDiff(tel *telemetry.Telemetry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx, span := tel.Tracer().Start(r.Context(), "diff.compute")
		defer span.End()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "read body")
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}
		var req postDiffRequest
		if err := json.Unmarshal(body, &req); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid JSON body")
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if req.ContentA == "" || req.ContentB == "" {
			http.Error(w, "contentA and contentB required", http.StatusBadRequest)
			return
		}

		dataA := []byte(req.ContentA)
		dataB := []byte(req.ContentB)
		diffFormat := format.Parse(req.Format, req.NameA, req.NameB)

		var diff *node.Diff
		switch diffFormat {
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
			span.RecordError(err)
			span.SetStatus(codes.Error, "diff")
			http.Error(w, "diff: "+err.Error(), http.StatusUnprocessableEntity)
			return
		}
		diff.FileA = req.NameA
		diff.FileB = req.NameB
		diff.Format = diffFormat

		span.SetAttributes(
			attribute.Int("diff.size_bytes_a", len(dataA)),
			attribute.Int("diff.size_bytes_b", len(dataB)),
			attribute.String("diff.format", diffFormat.String()),
			attribute.Int("diff.added", diff.Added),
			attribute.Int("diff.removed", diff.Removed),
			attribute.Int("diff.modified", diff.Modified),
		)
		tel.TrackEvent("diff-compute", apiDiffPath, map[string]string{
			"format":  diffFormat.String(),
			"added":   strconv.Itoa(diff.Added),
			"removed": strconv.Itoa(diff.Removed),
		})
		_ = ctx // used by Tracer().Start above

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
