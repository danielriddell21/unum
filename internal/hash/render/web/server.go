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

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"

	"github.com/danielriddell21/unum/internal/hash/types"
	"github.com/danielriddell21/unum/internal/telemetry"
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
	Version    string
	Tel        *telemetry.Telemetry
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

	d := shared.NewIndexData(opts.DarkTheme, opts.LightTheme, opts.Version)
	mux := http.NewServeMux()
	mux.HandleFunc("/shared.css", shared.ServeSharedAsset("assets/shared.css", "text/css"))
	mux.HandleFunc("/shared.js", shared.ServeSharedAsset("assets/shared.js", "application/javascript"))
	mux.HandleFunc("/style.css", shared.ServeAsset(assets, "assets/style.css", "text/css"))
	mux.HandleFunc("/app.js", shared.ServeAsset(assets, "assets/app.js", "application/javascript"))
	mux.HandleFunc("/api/derive", handleDerive(opts.Tel))
	mux.HandleFunc("/", shared.ServeTemplate(assets, "assets/index.html")(d))
	shared.RegisterMetrics(mux)
	shared.RegisterUmamiProxy(mux)

	srv := &http.Server{Addr: addr, Handler: otelhttp.NewHandler(mux, "unum-hash"), ReadHeaderTimeout: 10 * time.Second}

	shared.PrintStartupBanner("hash deriver", url)

	if autoOpen {
		go shared.OpenBrowser(url)
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

func handleDerive(tel *telemetry.Telemetry) http.HandlerFunc {
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

		ctx, span := tel.Tracer().Start(r.Context(), "hash.derive")
		defer span.End()

		result := deriveFn(input)
		span.SetAttributes(attribute.String("hash.input_length", strconv.Itoa(len(input))))
		tel.TrackEvent("hash-derive", "/api/derive", map[string]string{
			"input_length": strconv.Itoa(len(input)),
		})
		_ = ctx // used by Tracer().Start above

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}
