// Package web provides the HTTP server for the unum hash web UI.
// All static assets are embedded at build time — the binary is fully self-contained.
package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/danielriddell21/unum/internal/hash/types"
	"github.com/danielriddell21/unum/internal/web/shared"
)

//go:embed assets/*
var assets embed.FS

// Options configures the hash web server.
type Options struct {
	Port       int
	Quiet      bool
	DarkTheme  string // cyber | matrix | dracula | nord
	LightTheme string // clean | solarized
}

// deriveFn is injected to avoid import cycles.
var deriveFn func(input string) types.Result

// SetFuncs wires in the derivation function.
func SetFuncs(derive func(string) types.Result) {
	deriveFn = derive
}

// Start launches the hash web server, auto-opens the browser, and blocks until
// Ctrl+C.
func Start(opts Options) error {
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
	mux.HandleFunc("/api/derive", handleDerive())
	mux.HandleFunc("/", shared.ServeTemplate(assets, "assets/index.html")(d))

	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}

	shared.PrintStartupBanner("hash deriver", url)

	if autoOpen {
		go shared.OpenBrowser(url)
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
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
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}
