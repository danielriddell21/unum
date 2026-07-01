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

const (
	contentTypeHdr  = "Content-Type"
	contentTypeJSON = "application/json"
	contentTypeCSS  = "text/css"
	contentTypeJS   = "application/javascript"
	errKeyNotFound  = "key not found"
	apiUploadPath   = "/api/upload"
)

type Options struct {
	Port       int
	Filename   string
	Quiet      bool
	DarkTheme  string
	LightTheme string
	Version    string
	Tel        *telemetry.Telemetry
}

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
	mux.HandleFunc("/shared.css", shared.ServeSharedAsset("assets/shared.css", contentTypeCSS))
	mux.HandleFunc("/shared.js", shared.ServeSharedAsset("assets/shared.js", contentTypeJS))
	mux.HandleFunc("/style.css", shared.ServeAsset(assets, "assets/style.css", contentTypeCSS))
	mux.HandleFunc("/app.js", shared.ServeAsset(assets, "assets/app.js", contentTypeJS))
	mux.HandleFunc("/", shared.ServeTemplate(assets, "assets/index.html")(d))

	// API
	mux.HandleFunc("/api/tree", handleTree(payloadBytes))
	mux.HandleFunc(apiUploadPath, handleUpload(opts.Tel))
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

func toWebNode(n *node.Node) *webNode {
	wn := &webNode{
		Kind:  n.Kind.String(),
		Key:   n.Key,
		Index: n.Index,
		Raw:   n.Raw,
		Path:  n.Path(),
	}

	if n.Kind == node.KindString {
		var s string
		if json.Unmarshal([]byte(n.Raw), &s) == nil {
			wn.DisplayValue = s
		}
	}

	if v, ok := n.GetAnnotation(merkle.Lens, "hash"); ok {
		wn.MerkleHash, _ = v.(string)
	}

	if n.Kind == node.KindArray {
		wn.Stats = buildWebStats(n)
	}

	for _, c := range n.Children {
		wn.Children = append(wn.Children, toWebNode(c))
	}

	return wn
}

func buildWebStats(n *node.Node) *webStats {
	s, ok := stats.Get(n)
	if !ok {
		return nil
	}
	return &webStats{
		Count:        s.Count,
		NumericCount: s.NumericCount,
		Min:          s.Min,
		Max:          s.Max,
		Mean:         s.Mean,
		Stddev:       s.Stddev,
		P50:          s.P50,
		P95:          s.P95,
		P99:          s.P99,
	}
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

var rootCache sync.Map

var payloadCache sync.Map

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
	mux.HandleFunc("/shared.css", shared.ServeSharedAsset("assets/shared.css", contentTypeCSS))
	mux.HandleFunc("/shared.js", shared.ServeSharedAsset("assets/shared.js", contentTypeJS))
	mux.HandleFunc("/style.css", shared.ServeAsset(assets, "assets/style.css", contentTypeCSS))
	mux.HandleFunc("/app.js", shared.ServeAsset(assets, "assets/app.js", contentTypeJS))
	mux.HandleFunc("/api/tree", handleKeyedTree())
	mux.HandleFunc(apiUploadPath, handleUpload(opts.Tel))
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

func handleTree(payloadBytes []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if key := r.URL.Query().Get("key"); key != "" {
			v, ok := payloadCache.Load(key)
			if !ok {
				http.Error(w, errKeyNotFound, http.StatusNotFound)
				return
			}
			w.Header().Set(contentTypeHdr, contentTypeJSON)
			_, _ = w.Write(v.([]byte)) //nolint:gosec // v is always []byte; type assertion is safe
			return
		}
		w.Header().Set(contentTypeHdr, contentTypeJSON)
		_, _ = w.Write(payloadBytes)
	}
}

func handleKeyedTree() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		v, ok := payloadCache.Load(key)
		if !ok {
			http.Error(w, errKeyNotFound, http.StatusNotFound)
			return
		}
		w.Header().Set(contentTypeHdr, contentTypeJSON)
		_, _ = w.Write(v.([]byte)) //nolint:gosec // v is always []byte; type assertion is safe
	}
}

func handleKeyedQuery() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			writeJSON(w, queryResponse{Error: "no file context"})
			return
		}
		v, ok := rootCache.Load(key)
		if !ok {
			writeJSON(w, queryResponse{Error: errKeyNotFound})
			return
		}
		handleQuery(v.(*node.Node))(w, r) //nolint:gosec // v is always *node.Node; type assertion is safe
	}
}

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
		tel.TrackEvent("json-upload", apiUploadPath, map[string]string{
			"node_count": strconv.Itoa(nodeCount),
			"size_bytes": strconv.Itoa(len(req.Content)),
		})

		key := generateKey()
		rootCache.Store(key, root)
		payloadCache.Store(key, payloadBytes)

		w.Header().Set(contentTypeHdr, contentTypeJSON)
		_ = json.NewEncoder(w).Encode(map[string]string{"key": key})
	}
}

func generateKey() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
