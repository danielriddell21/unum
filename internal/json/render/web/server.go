// Package web provides the HTTP server for the unum JSON web UI.
// All static assets are embedded at build time — the binary is fully self-contained.
package web

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	ometric "go.opentelemetry.io/otel/metric"

	"github.com/danielriddell21/unum/internal/json/analyze"
	"github.com/danielriddell21/unum/internal/json/lens/merkle"
	"github.com/danielriddell21/unum/internal/json/lens/stats"
	"github.com/danielriddell21/unum/internal/json/lens/transform"
	"github.com/danielriddell21/unum/internal/json/lens/typegen"
	"github.com/danielriddell21/unum/internal/json/node"
	"github.com/danielriddell21/unum/internal/json/parse"
	"github.com/danielriddell21/unum/internal/telemetry"
	"github.com/danielriddell21/unum/internal/web/shared"
)

//go:embed assets/*
var assets embed.FS

// Options configures the web server.
type Options struct {
	Port       int    // 0 = find a free port
	Filename   string // original filename for display
	Quiet      bool
	DarkTheme  string // cyber | matrix | dracula | nord
	LightTheme string // clean | solarized
	Version    string
	Tel        *telemetry.Telemetry
}

// Start launches the web server, auto-opens the browser, and blocks until the
// user presses Ctrl+C.
func Start(root *node.Node, opts Options) error {
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

	// Pre-compute all analysis outputs once (not per-request)
	payload, err := buildPayload(root, opts.Filename)
	if err != nil {
		return fmt.Errorf("web: build payload: %w", err)
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("web: marshal payload: %w", err)
	}

	mux := http.NewServeMux()

	// Static assets
	d := shared.NewIndexData(opts.DarkTheme, opts.LightTheme, opts.Version)
	mux.HandleFunc("/shared.css", shared.ServeSharedAsset("assets/shared.css", "text/css"))
	mux.HandleFunc("/shared.js", shared.ServeSharedAsset("assets/shared.js", "application/javascript"))
	mux.HandleFunc("/style.css", shared.ServeAsset(assets, "assets/style.css", "text/css"))
	mux.HandleFunc("/app.js", shared.ServeAsset(assets, "assets/app.js", "application/javascript"))
	mux.HandleFunc("/", shared.ServeTemplate(assets, "assets/index.html")(d))

	// API
	mux.HandleFunc("/api/tree", func(w http.ResponseWriter, r *http.Request) {
		if key := r.URL.Query().Get("key"); key != "" {
			v, ok := payloadCache.Load(key)
			if !ok {
				http.Error(w, "key not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(v.([]byte)) //nolint:gosec // v is always []byte; type assertion is safe
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payloadBytes)
	})
	mux.HandleFunc("/api/upload", handleUpload(opts.Tel))
	mux.HandleFunc("/api/query", handleQuery(root))
	shared.RegisterMetrics(mux)
	shared.RegisterUmamiProxy(mux)

	srv := &http.Server{
		Addr:              addr,
		Handler:           otelhttp.NewHandler(mux, "unum-json"),
		ReadHeaderTimeout: 10 * time.Second,
	}

	shared.PrintStartupBanner("json explorer", url)

	if autoOpen {
		go shared.OpenBrowser(url)
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

// ─── Payload ──────────────────────────────────────────────────────────────────

// treePayload is the full JSON object served at /api/tree.
type treePayload struct {
	Filename   string         `json:"filename"`
	NodeCount  int            `json:"nodeCount"`
	MaxDepth   int            `json:"maxDepth"`
	SizeBytes  int64          `json:"sizeBytes"`
	Tree       *webNode       `json:"tree"`
	YAML       string         `json:"yaml"`
	Typegen    map[string]any `json:"typegen"`
	MerkleRoot string         `json:"merkleRoot,omitempty"`
}

// webNode is a serializable tree node for the frontend.
type webNode struct {
	Kind         string     `json:"kind"`
	Key          string     `json:"key,omitempty"`
	Index        int        `json:"index"`
	Raw          string     `json:"raw,omitempty"`
	DisplayValue string     `json:"displayValue,omitempty"`
	Path         string     `json:"path"`
	Children     []*webNode `json:"children,omitempty"`
	MerkleHash   string     `json:"merkleHash,omitempty"`
	Stats        *webStats  `json:"stats,omitempty"`
}

type webStats struct {
	Count        int     `json:"count"`
	NumericCount int     `json:"numericCount"`
	Min          float64 `json:"min"`
	Max          float64 `json:"max"`
	Mean         float64 `json:"mean"`
	Stddev       float64 `json:"stddev"`
	P50          float64 `json:"p50"`
	P95          float64 `json:"p95"`
	P99          float64 `json:"p99"`
}

func buildPayload(root *node.Node, filename string) (*treePayload, error) { //nolint:unparam // error return kept for future lens failures; removing it would require a signature change later
	// Run annotation lenses
	ctx := context.Background()

	statsA := stats.Analyzer{}
	merkleA := merkle.Analyzer{}

	_ = statsA.Run(ctx, root, analyze.Options{RunStats: true})
	_ = merkleA.Run(ctx, root, analyze.Options{RunMerkle: true})

	// YAML
	yamlBytes, _ := transform.ToYAML(root)

	// TypeGen
	goTypes, _ := typegen.Generate(root, typegen.Options{Target: typegen.TargetGo})
	tsTypes, _ := typegen.Generate(root, typegen.Options{Target: typegen.TargetTypeScript})
	schemaStr, _ := typegen.Generate(root, typegen.Options{Target: typegen.TargetJSONSchema})

	p := &treePayload{
		Filename:  filename,
		NodeCount: root.CountNodes(),
		MaxDepth:  maxDepth(root),
		Tree:      toWebNode(root),
		YAML:      string(yamlBytes),
		Typegen: map[string]any{
			"go":         goTypes,
			"ts":         tsTypes,
			"jsonschema": schemaStr,
		},
		MerkleRoot: merkle.RootHash(root),
	}

	return p, nil
}

func toWebNode(n *node.Node) *webNode { //nolint:gocognit // maps every node kind to its web representation; branching on kind is the algorithm
	wn := &webNode{
		Kind:  n.Kind.String(),
		Key:   n.Key,
		Index: n.Index,
		Raw:   n.Raw,
		Path:  n.Path(),
	}

	// Display value for strings (unquoted)
	if n.Kind == node.KindString {
		var s string
		if json.Unmarshal([]byte(n.Raw), &s) == nil {
			wn.DisplayValue = s
		}
	}

	// Merkle hash
	if v, ok := n.GetAnnotation(merkle.Lens, "hash"); ok {
		wn.MerkleHash, _ = v.(string)
	}

	// Stats
	if n.Kind == node.KindArray {
		if nc, ok := n.GetAnnotation(stats.Lens, "numeric_count"); ok {
			ws := &webStats{}
			ws.NumericCount = nc.(int)
			if v, ok := n.GetAnnotation(stats.Lens, "count"); ok {
				ws.Count = v.(int)
			}
			if v, ok := n.GetAnnotation(stats.Lens, "min"); ok {
				ws.Min = v.(float64)
			}
			if v, ok := n.GetAnnotation(stats.Lens, "max"); ok {
				ws.Max = v.(float64)
			}
			if v, ok := n.GetAnnotation(stats.Lens, "mean"); ok {
				ws.Mean = v.(float64)
			}
			if v, ok := n.GetAnnotation(stats.Lens, "stddev"); ok {
				ws.Stddev = v.(float64)
			}
			if v, ok := n.GetAnnotation(stats.Lens, "p50"); ok {
				ws.P50 = v.(float64)
			}
			if v, ok := n.GetAnnotation(stats.Lens, "p95"); ok {
				ws.P95 = v.(float64)
			}
			if v, ok := n.GetAnnotation(stats.Lens, "p99"); ok {
				ws.P99 = v.(float64)
			}
			wn.Stats = ws
		}
	}

	for _, c := range n.Children {
		wn.Children = append(wn.Children, toWebNode(c))
	}

	return wn
}

func maxDepth(n *node.Node) int {
	if len(n.Children) == 0 {
		return 0
	}
	max := 0
	for _, c := range n.Children {
		d := maxDepth(c)
		if d > max {
			max = d
		}
	}
	return max + 1
}

// ─── Input mode (no pre-selected file) ───────────────────────────────────────

// rootCache stores parsed roots keyed by upload key.
var rootCache sync.Map // map[string]*node.Node

// payloadCache stores pre-marshalled payloads keyed by upload key.
var payloadCache sync.Map // map[string][]byte

// StartServer launches the json web server with no pre-loaded file.
// GET /api/tree returns 204; the frontend shows a drop/paste zone.
// POST /api/upload accepts JSON content and returns a key for later retrieval.
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
	var err error
	if port == 0 {
		port, err = shared.FreePort()
		if err != nil {
			return fmt.Errorf("web: cannot find free port: %w", err)
		}
	}

	addr := host + ":" + strconv.Itoa(port)
	url := "http://" + addr

	d := shared.NewIndexData(opts.DarkTheme, opts.LightTheme, opts.Version)
	mux := http.NewServeMux()
	mux.HandleFunc("/shared.css", shared.ServeSharedAsset("assets/shared.css", "text/css"))
	mux.HandleFunc("/shared.js", shared.ServeSharedAsset("assets/shared.js", "application/javascript"))
	mux.HandleFunc("/style.css", shared.ServeAsset(assets, "assets/style.css", "text/css"))
	mux.HandleFunc("/app.js", shared.ServeAsset(assets, "assets/app.js", "application/javascript"))
	mux.HandleFunc("/api/tree", handleKeyedTree())
	mux.HandleFunc("/api/upload", handleUpload(opts.Tel))
	mux.HandleFunc("/api/query", handleKeyedQuery())
	mux.HandleFunc("/", shared.ServeTemplate(assets, "assets/index.html")(d))
	shared.RegisterMetrics(mux)
	shared.RegisterUmamiProxy(mux)

	srv := &http.Server{Addr: addr, Handler: otelhttp.NewHandler(mux, "unum-json"), ReadHeaderTimeout: 10 * time.Second}

	shared.PrintStartupBanner("json explorer", url)

	if autoOpen {
		go shared.OpenBrowser(url)
	}
	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

// handleKeyedTree serves GET /api/tree — returns 204 when no key is given
// (signals the frontend to show the upload panel), or the pre-built payload
// for a key returned by /api/upload.
func handleKeyedTree() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		v, ok := payloadCache.Load(key)
		if !ok {
			http.Error(w, "key not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(v.([]byte)) //nolint:gosec // v is always []byte; type assertion is safe
	}
}

// handleKeyedQuery serves POST /api/query — requires a key from /api/upload.
func handleKeyedQuery() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			writeJSON(w, queryResponse{Error: "no file context"})
			return
		}
		v, ok := rootCache.Load(key)
		if !ok {
			writeJSON(w, queryResponse{Error: "key not found"})
			return
		}
		handleQuery(v.(*node.Node))(w, r) //nolint:gosec // v is always *node.Node; type assertion is safe
	}
}

// uploadRequest is the body accepted by POST /api/upload.
type uploadRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

func handleUpload(tel *telemetry.Telemetry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx, span := tel.Tracer().Start(r.Context(), "json.upload.process")
		defer span.End()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "read body")
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}
		var req uploadRequest
		if err := json.Unmarshal(body, &req); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid JSON body")
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if req.Content == "" {
			http.Error(w, "content required", http.StatusBadRequest)
			return
		}
		root, err := parse.Parse([]byte(req.Content))
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "parse")
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		payload, err := buildPayload(root, req.Filename)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "build payload")
			http.Error(w, "build payload: "+err.Error(), http.StatusInternalServerError)
			return
		}
		payload.SizeBytes = int64(len(req.Content))

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "marshal payload")
			http.Error(w, "marshal payload: "+err.Error(), http.StatusInternalServerError)
			return
		}

		nodeCount := root.CountNodes()
		span.SetAttributes(
			attribute.Int("json.node_count", nodeCount),
			attribute.Int("json.size_bytes", len(req.Content)),
		)
		tel.M.WebUploads.Add(ctx, 1, ometric.WithAttributes(attribute.String("tool", "json")))
		tel.TrackEvent("json-upload", "/api/upload", map[string]string{
			"node_count": strconv.Itoa(nodeCount),
			"size_bytes": strconv.Itoa(len(req.Content)),
		})

		key := generateKey()
		rootCache.Store(key, root)
		payloadCache.Store(key, payloadBytes)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"key": key})
	}
}

func generateKey() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
