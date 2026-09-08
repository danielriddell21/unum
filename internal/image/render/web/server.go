package web

import (
	"embed"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/danielriddell21/unum/internal/telemetry"
	"github.com/danielriddell21/unum/internal/web/shared"
)

//go:embed assets/*
var assets embed.FS

type Options struct {
	Port       int
	Quiet      bool
	DarkTheme  string
	LightTheme string
	Version    string
	Tel        *telemetry.Telemetry
	SourceName string
	SourceData []byte
	SourceMIME string
}

func Start(opts Options) error {
	bind, err := shared.ResolveBind(opts.Port)
	if err != nil {
		return fmt.Errorf("web: %w", err)
	}
	addr := bind.Addr()
	url := bind.URL()

	st := newStore()
	d := shared.NewIndexData(opts.DarkTheme, opts.LightTheme, opts.Version)

	mux := http.NewServeMux()
	mux.HandleFunc("/shared.css", shared.ServeSharedAsset("assets/shared.css", "text/css"))
	mux.HandleFunc("/shared.js", shared.ServeSharedAsset("assets/shared.js", "application/javascript"))
	mux.HandleFunc("/style.css", shared.ServeAsset(assets, "assets/style.css", "text/css"))
	mux.HandleFunc("/app.js", shared.ServeAsset(assets, "assets/app.js", "application/javascript"))
	mux.HandleFunc("/api/upload", handleUpload(opts.Tel, st))
	mux.HandleFunc("/api/optimize", handleOptimize(opts.Tel, st))
	mux.HandleFunc("/api/analyze", handleAnalyze(opts.Tel, st))
	mux.HandleFunc("/api/original", handleOriginal(st))
	mux.HandleFunc("/api/source", handleSource(opts))
	mux.HandleFunc("/", shared.ServeTemplate(assets, "assets/index.html")(d))
	shared.RegisterMetrics(mux, bind)
	shared.RegisterUmamiProxy(mux)

	srv := &http.Server{
		Addr:              addr,
		Handler:           otelhttp.NewHandler(mux, "unum-image"),
		ReadHeaderTimeout: 10 * time.Second,
	}

	shared.PrintBindWarning(bind)
	shared.PrintStartupBanner("image optimizer", url)

	if bind.AutoOpen {
		go shared.OpenBrowser(url)
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

func handleSource(opts Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(opts.SourceData) == 0 {
			http.Error(w, "no image loaded", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", opts.SourceMIME)
		w.Header().Set("X-Unum-Name", opts.SourceName)
		_, _ = w.Write(opts.SourceData)
	}
}
