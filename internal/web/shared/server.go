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
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/danielriddell21/unum/internal/theme"
)

type IndexData struct {
	DarkTheme      string
	LightTheme     string
	Version        string
	ThemeData      template.JS
	UmamiEnabled   bool
	UmamiWebsiteID string
}

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

type Bind struct {
	Host     string
	Port     int
	Public   bool
	AutoOpen bool
}

func ResolveBind(port int) (Bind, error) {
	b := Bind{Host: "localhost", Port: port, AutoOpen: true}

	envPort := os.Getenv("PORT")
	if envPort != "" {
		if n, err := strconv.Atoi(envPort); err == nil {
			b.Port = n
		}
	}

	// Binding beyond loopback is opt-in. UNUM_BIND names the interface
	// explicitly; PORT alone is not enough, since unrelated dev tooling exports
	// it and a stray value must never publish the user's data to the network.
	switch {
	case os.Getenv("UNUM_BIND") != "":
		b.Host = os.Getenv("UNUM_BIND")
	case envPort != "" && os.Getenv("UNUM_ENV") != "":
		b.Host = "0.0.0.0"
	}

	b.Public = !isLoopbackHost(b.Host)
	if b.Public {
		b.AutoOpen = false
	}

	if b.Port == 0 {
		n, err := FreePort()
		if err != nil {
			return Bind{}, err
		}
		b.Port = n
	}
	return b, nil
}

func (b Bind) Addr() string {
	return net.JoinHostPort(b.Host, strconv.Itoa(b.Port))
}

func (b Bind) URL() string {
	return "http://" + b.Addr()
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func FreePort() (int, error) {
	l, err := net.Listen("tcp", "localhost:0") //nolint:noctx // net.Listen has no context-aware variant; localhost-only binding, not user-controlled
	if err != nil {
		return 0, fmt.Errorf("listen: %w", err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port, nil
}

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

func PrintStartupBanner(toolName, url string) {
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4FF"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A"))
	_, _ = fmt.Fprintf(os.Stderr, "%s  → %s\n",
		accent.Render("[ UNUM ] "+toolName),
		accent.Render(url),
	)
	_, _ = fmt.Fprintf(os.Stderr, "%s\n", muted.Render("Press Ctrl+C to stop"))
}

func PrintBindWarning(b Bind) {
	if !b.Public {
		return
	}
	warn := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAF00"))
	_, _ = fmt.Fprintf(os.Stderr, "%s\n", warn.Render(
		"[ UNUM ] warning: listening on "+b.Host+" — reachable from the network, with no authentication",
	))
}

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

func ServeSharedAsset(path, contentType string) http.HandlerFunc {
	return ServeAsset(Assets, path, contentType)
}

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

func RegisterMetrics(mux *http.ServeMux, b Bind) {
	// Only where something actually scrapes it: the hosted deployment, or an
	// explicit local opt-in. A loopback dev server has no scraper.
	if !b.Public && os.Getenv("UNUM_METRICS") == "" {
		return
	}
	mux.Handle("/metrics", promhttp.Handler())
}
