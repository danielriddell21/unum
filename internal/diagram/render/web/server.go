package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"

	"github.com/danielriddell21/unum/internal/diagram/engine"
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
	Source     string
	Lang       string
	File       string
}

type server struct {
	opts Options
}

func Start(opts Options) error {
	bind, err := shared.ResolveBind(opts.Port)
	if err != nil {
		return fmt.Errorf("web: %w", err)
	}
	addr := bind.Addr()
	url := bind.URL()

	s := &server{opts: opts}
	mux := http.NewServeMux()
	mux.HandleFunc("/shared.css", shared.ServeSharedAsset("assets/shared.css", "text/css"))
	mux.HandleFunc("/shared.js", shared.ServeSharedAsset("assets/shared.js", "application/javascript"))
	mux.HandleFunc("/style.css", shared.ServeAsset(assets, "assets/style.css", "text/css"))
	mux.HandleFunc("/app.js", shared.ServeAsset(assets, "assets/app.js", "application/javascript"))
	mux.HandleFunc("/api/render", s.handleRender)
	mux.HandleFunc("/", s.handleIndex)
	shared.RegisterMetrics(mux, bind)
	shared.RegisterUmamiProxy(mux)

	srv := &http.Server{Addr: addr, Handler: otelhttp.NewHandler(mux, "unum-diagram"), ReadHeaderTimeout: 10 * time.Second}

	shared.PrintBindWarning(bind)
	shared.PrintStartupBanner("diagram renderer", url)
	if bind.AutoOpen {
		go shared.OpenBrowser(url)
	}
	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	raw, err := assets.ReadFile("assets/index.html")
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	tmpl, err := template.New("page").Parse(string(raw))
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	sourceJSON, _ := json.Marshal(s.opts.Source)
	data := struct {
		shared.IndexData
		Lang       string
		File       string
		SourceJSON template.JS
	}{
		IndexData:  shared.NewIndexData(s.opts.DarkTheme, s.opts.LightTheme, s.opts.Version),
		Lang:       s.opts.Lang,
		File:       s.opts.File,
		SourceJSON: template.JS(sourceJSON), //nolint:gosec // json-encoded file content, not executable markup
	}
	w.Header().Set("Content-Type", "text/html")
	_ = tmpl.Execute(w, data)
}

func (s *server) handleRender(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "svg"
	}
	themeName := r.URL.Query().Get("theme")
	if themeName == "" {
		themeName = s.opts.DarkTheme
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	ctx, span := s.opts.Tel.Tracer().Start(r.Context(), "diagram.web")
	defer span.End()
	span.SetAttributes(attribute.String("lang", lang), attribute.String("format", format))

	out, contentType, err := engine.Render(lang, format, string(body), engine.ThemeByName(themeName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.opts.Tel.TrackEvent("render-diagram", "/api/render", map[string]string{"lang": lang, "format": format})
	_ = ctx

	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(out) //nolint:gosec // local single-user renderer serving the user's own diagram under an explicit content type, not a multi-user web app
}
