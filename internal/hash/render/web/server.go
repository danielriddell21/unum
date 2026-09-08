package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
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

type Options struct {
	Port       int
	Quiet      bool
	DarkTheme  string
	LightTheme string
	Version    string
	Tel        *telemetry.Telemetry
	Derive     func(input string) types.Result
}

func Start(opts Options) error {
	bind, err := shared.ResolveBind(opts.Port)
	if err != nil {
		return fmt.Errorf("web: %w", err)
	}
	addr := bind.Addr()
	url := bind.URL()

	d := shared.NewIndexData(opts.DarkTheme, opts.LightTheme, opts.Version)
	mux := http.NewServeMux()
	mux.HandleFunc("/shared.css", shared.ServeSharedAsset("assets/shared.css", "text/css"))
	mux.HandleFunc("/shared.js", shared.ServeSharedAsset("assets/shared.js", "application/javascript"))
	mux.HandleFunc("/style.css", shared.ServeAsset(assets, "assets/style.css", "text/css"))
	mux.HandleFunc("/app.js", shared.ServeAsset(assets, "assets/app.js", "application/javascript"))
	mux.HandleFunc("/api/derive", handleDerive(opts.Tel, opts.Derive))
	mux.HandleFunc("/", shared.ServeTemplate(assets, "assets/index.html")(d))
	shared.RegisterMetrics(mux, bind)
	shared.RegisterUmamiProxy(mux)

	srv := &http.Server{Addr: addr, Handler: otelhttp.NewHandler(mux, "unum-hash"), ReadHeaderTimeout: 10 * time.Second}

	shared.PrintBindWarning(bind)
	shared.PrintStartupBanner("hash deriver", url)

	if bind.AutoOpen {
		go shared.OpenBrowser(url)
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

func handleDerive(tel *telemetry.Telemetry, derive func(input string) types.Result) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input := r.URL.Query().Get("input")
		if input == "" {
			http.Error(w, "missing input param", http.StatusBadRequest)
			return
		}

		ctx, span := tel.Tracer().Start(r.Context(), "hash.derive")
		defer span.End()

		result := derive(input)
		span.SetAttributes(attribute.String("hash.input_length", strconv.Itoa(len(input))))
		tel.TrackEvent("hash-derive", "/api/derive", map[string]string{
			"input_length": strconv.Itoa(len(input)),
		})
		_ = ctx // used by Tracer().Start above

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}
