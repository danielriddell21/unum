package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"

	"github.com/danielriddell21/unum/internal/render/diagram"
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
}

type server struct {
	opts       Options
	once       sync.Once
	browser    *diagram.Browser
	browserErr error
}

func (s *server) getBrowser() (*diagram.Browser, error) {
	s.once.Do(func() {
		if !diagram.BrowserAvailable() {
			s.browserErr = fmt.Errorf("mermaid and png need a Chromium browser: install one or set UNUM_CHROMIUM_BIN")
			return
		}
		s.browser, s.browserErr = diagram.NewBrowser()
	})
	return s.browser, s.browserErr
}

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

	s := &server{opts: opts}
	mux := http.NewServeMux()
	mux.HandleFunc("/shared.css", shared.ServeSharedAsset("assets/shared.css", "text/css"))
	mux.HandleFunc("/shared.js", shared.ServeSharedAsset("assets/shared.js", "application/javascript"))
	mux.HandleFunc("/style.css", shared.ServeAsset(assets, "assets/style.css", "text/css"))
	mux.HandleFunc("/app.js", shared.ServeAsset(assets, "assets/app.js", "application/javascript"))
	mux.HandleFunc("/api/render", s.handleRender)
	mux.HandleFunc("/", s.handleIndex)
	shared.RegisterMetrics(mux)
	shared.RegisterUmamiProxy(mux)

	srv := &http.Server{Addr: addr, Handler: otelhttp.NewHandler(mux, "unum-render"), ReadHeaderTimeout: 10 * time.Second}

	shared.PrintStartupBanner("render", url)
	if autoOpen {
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
		SourceJSON template.JS
	}{
		IndexData:  shared.NewIndexData(s.opts.DarkTheme, s.opts.LightTheme, s.opts.Version),
		Lang:       s.opts.Lang,
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	ctx, span := s.opts.Tel.Tracer().Start(r.Context(), "render.web")
	defer span.End()
	span.SetAttributes(attribute.String("lang", lang), attribute.String("format", format))

	var b *diagram.Browser
	if diagram.NeedsBrowser(lang, format) {
		bb, err := s.getBrowser()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		b = bb
	}

	out, contentType, err := diagram.Render(lang, format, string(body), b, diagram.ThemeByName(s.opts.DarkTheme))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.opts.Tel.TrackEvent("render-web", "/api/render", map[string]string{"lang": lang, "format": format})
	_ = ctx

	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(out)
}
