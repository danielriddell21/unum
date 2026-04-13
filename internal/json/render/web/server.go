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
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/danielriddell21/unum/internal/json/analyze"
	"github.com/danielriddell21/unum/internal/json/lens/merkle"
	"github.com/danielriddell21/unum/internal/json/lens/stats"
	"github.com/danielriddell21/unum/internal/json/lens/transform"
	"github.com/danielriddell21/unum/internal/json/lens/typegen"
	"github.com/danielriddell21/unum/internal/json/node"
	"github.com/danielriddell21/unum/internal/json/parse"
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
	d := shared.IndexData{DarkTheme: opts.DarkTheme, LightTheme: opts.LightTheme}
	mux.HandleFunc("/shared.css", shared.ServeSharedAsset("assets/shared.css", "text/css"))
	mux.HandleFunc("/shared.js", shared.ServeSharedAsset("assets/shared.js", "application/javascript"))
	mux.HandleFunc("/style.css", shared.ServeAsset(assets, "assets/style.css", "text/css"))
	mux.HandleFunc("/app.js", shared.ServeAsset(assets, "assets/app.js", "application/javascript"))
	mux.HandleFunc("/", shared.ServeTemplate(assets, "assets/index.html")(d))

	// API
	mux.HandleFunc("/api/tree", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payloadBytes)
	})
	mux.HandleFunc("/api/query", handleQuery(root))

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
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

// ─── Browser mode ─────────────────────────────────────────────────────────────

// rootCache stores parsed roots keyed by absolute file path or upload key.
var rootCache sync.Map // map[string]*node.Node

// payloadCache stores pre-marshalled payloads keyed by upload key.
var payloadCache sync.Map // map[string][]byte

// StartBrowser launches the web server in file-browser mode (no pre-selected
// file). The landing page lets the user navigate the filesystem and pick a
// .json file to explore.
func StartBrowser(opts Options) error {
	baseDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}

	host := "localhost"
	autoOpen := true
	if p := os.Getenv("PORT"); p != "" {
		if n, nerr := strconv.Atoi(p); nerr == nil {
			opts.Port = n
		}
		host = "0.0.0.0"
		autoOpen = false
	}

	port := opts.Port
	if port == 0 {
		port, err = shared.FreePort()
		if err != nil {
			return fmt.Errorf("web: cannot find free port: %w", err)
		}
	}

	addr := host + ":" + strconv.Itoa(port)
	url := "http://" + addr

	mux := http.NewServeMux()
	d := shared.IndexData{DarkTheme: opts.DarkTheme, LightTheme: opts.LightTheme}
	mux.HandleFunc("/shared.css", shared.ServeSharedAsset("assets/shared.css", "text/css"))
	mux.HandleFunc("/shared.js", shared.ServeSharedAsset("assets/shared.js", "application/javascript"))
	mux.HandleFunc("/style.css", shared.ServeAsset(assets, "assets/style.css", "text/css"))
	mux.HandleFunc("/app.js", shared.ServeAsset(assets, "assets/app.js", "application/javascript"))
	mux.HandleFunc("/explore", shared.ServeTemplate(assets, "assets/index.html")(d))
	mux.HandleFunc("/api/browse", handleBrowse(baseDir))
	mux.HandleFunc("/api/upload", handleUpload())
	mux.HandleFunc("/api/tree", handleDynamicTree(baseDir))
	mux.HandleFunc("/api/query", handleDynamicQuery(baseDir))
	mux.HandleFunc("/", shared.ServeTemplate(assets, "assets/browse.html")(d))

	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}

	shared.PrintStartupBanner("json explorer", url)

	if autoOpen {
		go shared.OpenBrowser(url)
	}
	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

// dirEntry is a serialisable filesystem entry for the browse API.
type dirEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
	Path  string `json:"path"`
}

func handleBrowse(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dir := r.URL.Query().Get("dir")
		if dir == "" {
			dir = baseDir
		}
		abs, err := filepath.Abs(dir)
		if err != nil || !strings.HasPrefix(abs, baseDir) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		entries, err := os.ReadDir(abs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var result []dirEntry
		for _, e := range entries {
			name := e.Name()
			if strings.HasPrefix(name, ".") {
				continue
			}
			if e.IsDir() || strings.HasSuffix(strings.ToLower(name), ".json") {
				result = append(result, dirEntry{
					Name:  name,
					IsDir: e.IsDir(),
					Path:  filepath.Join(abs, name),
				})
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}

func handleDynamicTree(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Upload key path — no filesystem access needed.
		if key := r.URL.Query().Get("key"); key != "" {
			v, ok := payloadCache.Load(key)
			if !ok {
				http.Error(w, "key not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(v.([]byte)) //nolint:gosec // v is always []byte here; type assertion is safe, no user-controlled input
			return
		}

		file := r.URL.Query().Get("file")
		if file == "" {
			http.Error(w, "missing file or key param", http.StatusBadRequest)
			return
		}
		abs, err := filepath.Abs(file)
		if err != nil || !strings.HasPrefix(abs, baseDir) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		data, err := os.ReadFile(abs) //nolint:gosec // path is resolved from user-provided filename; intentional file read, G304 suppressed globally but kept explicit here
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		root, err := parse.Parse(data)
		if err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		rootCache.Store(abs, root)

		payload, err := buildPayload(root, abs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		payload.SizeBytes = int64(len(data))

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payloadBytes)
	}
}

func handleDynamicQuery(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Upload key path.
		if key := r.URL.Query().Get("key"); key != "" {
			v, ok := rootCache.Load(key)
			if !ok {
				writeJSON(w, queryResponse{Error: "key not found"})
				return
			}
			handleQuery(v.(*node.Node))(w, r)
			return
		}

		file := r.URL.Query().Get("file")
		if file == "" {
			writeJSON(w, queryResponse{Error: "no file context"})
			return
		}
		abs, err := filepath.Abs(file)
		if err != nil || !strings.HasPrefix(abs, baseDir) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		v, ok := rootCache.Load(abs)
		if !ok {
			writeJSON(w, queryResponse{Error: "file not yet loaded — open it first"})
			return
		}
		handleQuery(v.(*node.Node))(w, r)
	}
}

// uploadRequest is the body accepted by POST /api/upload.
type uploadRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

func handleUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}
		var req uploadRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if req.Content == "" {
			http.Error(w, "content required", http.StatusBadRequest)
			return
		}
		root, err := parse.Parse([]byte(req.Content))
		if err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		payload, err := buildPayload(root, req.Filename)
		if err != nil {
			http.Error(w, "build payload: "+err.Error(), http.StatusInternalServerError)
			return
		}
		payload.SizeBytes = int64(len(req.Content))

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			http.Error(w, "marshal payload: "+err.Error(), http.StatusInternalServerError)
			return
		}

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
