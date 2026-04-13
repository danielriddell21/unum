package shared

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/danielriddell21/unum/internal/theme"
)

// IndexData holds the theme config injected into HTML templates.
type IndexData struct {
	DarkTheme      string
	LightTheme     string
	Version        string
	ThemeData      template.JS // JSON: {dark:{cyber:{...},...}, light:{clean:{...},...}}
	UmamiEnabled   bool        // true if UMAMI_URL is set — enables JS tracking snippet
	UmamiWebsiteID string      // Umami website ID for the tracking snippet
}

// NewIndexData builds IndexData with all palette CSS vars pre-serialised.
func NewIndexData(dark, light, version string) IndexData {
	type themeSet struct {
		Dark  map[string]map[string]string `json:"dark"`
		Light map[string]map[string]string `json:"light"`
	}
	ts := themeSet{
		Dark: map[string]map[string]string{
			"cyber":   theme.ResolvePalette("cyber").ToCSSVars(),
			"matrix":  theme.ResolvePalette("matrix").ToCSSVars(),
			"dracula": theme.ResolvePalette("dracula").ToCSSVars(),
			"nord":    theme.ResolvePalette("nord").ToCSSVars(),
		},
		Light: map[string]map[string]string{
			"clean":     theme.ResolveLightPalette("clean").ToCSSVars(),
			"solarized": theme.ResolveLightPalette("solarized").ToCSSVars(),
		},
	}
	b, _ := json.Marshal(ts)
	umamiURL := os.Getenv("UMAMI_URL")
	return IndexData{
		DarkTheme:      dark,
		LightTheme:     light,
		Version:        version,
		ThemeData:      template.JS(b), //nolint:gosec // controlled palette data, not user input
		UmamiEnabled:   umamiURL != "" && os.Getenv("UMAMI_WEBSITE_ID") != "",
		UmamiWebsiteID: os.Getenv("UMAMI_WEBSITE_ID"),
	}
}

// ServeTemplate parses the named template from fs and serves it with data d.
func ServeTemplate(fs embed.FS, path string) func(d IndexData) http.HandlerFunc {
	return func(d IndexData) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			raw, err := fs.ReadFile(path)
			if err != nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			tmpl, err := template.New("page").Parse(string(raw))
			if err != nil {
				http.Error(w, "template error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			_ = tmpl.Execute(w, d)
		}
	}
}

// FreePort returns a random free TCP port on localhost.
func FreePort() (int, error) {
	l, err := net.Listen("tcp", "localhost:0") //nolint:noctx // net.Listen has no context-aware variant; localhost-only binding, not user-controlled
	if err != nil {
		return 0, fmt.Errorf("listen: %w", err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// OpenBrowser opens url in the default browser after a short delay.
func OpenBrowser(url string) {
	time.Sleep(300 * time.Millisecond)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url) //nolint:noctx // browser launcher; fire-and-forget subprocess, no timeout or cancellation needed
	case "darwin":
		cmd = exec.Command("open", url) //nolint:noctx // browser launcher; fire-and-forget subprocess, no timeout or cancellation needed
	default:
		cmd = exec.Command("xdg-open", url) //nolint:noctx // browser launcher; fire-and-forget subprocess, no timeout or cancellation needed
	}
	_ = cmd.Start()
}

// PrintStartupBanner writes the tool startup lines to stderr.
func PrintStartupBanner(toolName, url string) {
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4FF"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A"))
	_, _ = fmt.Fprintf(os.Stderr, "%s  → %s\n",
		accent.Render("[ UNUM ] "+toolName),
		accent.Render(url),
	)
	_, _ = fmt.Fprintf(os.Stderr, "%s\n", muted.Render("Press Ctrl+C to stop"))
}

// ServeAsset serves a file from fs at path with the given Content-Type.
func ServeAsset(fs embed.FS, path, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(path)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", contentType)
		_, _ = w.Write(data)
	}
}

// ServeSharedAsset serves a file from the shared Assets embed.
func ServeSharedAsset(path, contentType string) http.HandlerFunc {
	return ServeAsset(Assets, path, contentType)
}

// RegisterUmamiProxy registers /umami/* routes that reverse-proxy to the
// internal Umami instance. No-op if UMAMI_URL is empty.
func RegisterUmamiProxy(mux *http.ServeMux) {
	umamiURL := os.Getenv("UMAMI_URL")
	if umamiURL == "" {
		return
	}
	target, err := url.Parse(umamiURL)
	if err != nil {
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target) //nolint:gosec // UMAMI_URL is operator-controlled infrastructure config, not user input
	mux.Handle("/umami/", http.StripPrefix("/umami", proxy))
}

// RegisterMetrics registers the /metrics endpoint for Prometheus scraping.
func RegisterMetrics(mux *http.ServeMux) {
	mux.Handle("/metrics", promhttp.Handler())
}
