// Package web provides the HTTP server for the unum hash web UI.
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

	"github.com/danielriddell21/unum/internal/hash/types"
)

//go:embed assets/*
var assets embed.FS

// Options configures the hash web server.
type Options struct {
	Port  int
	Quiet bool
}

// deriveFn and historyFns are injected to avoid import cycles.
var (
	deriveFn        func(input string) types.Result
	appendHistoryFn func(input string) error
	loadHistoryFn   func() []types.HistoryEntry
)

// SetFuncs wires in the derivation and history functions.
func SetFuncs(
	derive func(string) types.Result,
	appendHistory func(string) error,
	loadHistory func() []types.HistoryEntry,
) {
	deriveFn = derive
	appendHistoryFn = appendHistory
	loadHistoryFn = loadHistory
}

// Start launches the hash web server, auto-opens the browser, and blocks until
// Ctrl+C.
func Start(opts Options) error {
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

	mux := http.NewServeMux()
	mux.HandleFunc("/style.css", serveAsset("assets/style.css", "text/css"))
	mux.HandleFunc("/app.js", serveAsset("assets/app.js", "application/javascript"))
	mux.HandleFunc("/api/derive", handleDerive())
	mux.HandleFunc("/api/history", handleHistory())
	mux.HandleFunc("/", serveAsset("assets/index.html", "text/html"))

	srv := &http.Server{Addr: addr, Handler: mux}

	_, _ = fmt.Fprintf(os.Stderr, "\033[38;5;51m[ UNUM ] HASH INTERFACE LIVE → %s\033[0m\n", url)
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

func handleDerive() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input := r.URL.Query().Get("input")
		if input == "" {
			http.Error(w, "missing input param", http.StatusBadRequest)
			return
		}
		if deriveFn == nil {
			http.Error(w, "derive not configured", http.StatusInternalServerError)
			return
		}
		result := deriveFn(input)
		if appendHistoryFn != nil {
			_ = appendHistoryFn(input)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}

func handleHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var entries []types.HistoryEntry
		if loadHistoryFn != nil {
			entries = loadHistoryFn()
		}
		if entries == nil {
			entries = []types.HistoryEntry{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(entries)
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
