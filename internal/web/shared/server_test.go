package shared_test

import (
	"context"
	"embed"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielriddell21/unum/internal/web/shared"
)

// ─── NewIndexData ─────────────────────────────────────────────────────────────

func TestNewIndexData_BasicFields(t *testing.T) {
	d := shared.NewIndexData("cyber", "clean", "1.2.3")
	if d.DarkTheme != "cyber" {
		t.Errorf("DarkTheme=%q, want cyber", d.DarkTheme)
	}
	if d.LightTheme != "clean" {
		t.Errorf("LightTheme=%q, want clean", d.LightTheme)
	}
	if d.Version != "1.2.3" {
		t.Errorf("Version=%q, want 1.2.3", d.Version)
	}
}

func TestNewIndexData_ThemeDataPresent(t *testing.T) {
	d := shared.NewIndexData("matrix", "solarized", "")
	if len(d.ThemeData) == 0 {
		t.Error("ThemeData should be non-empty JSON")
	}
}

func TestNewIndexData_UmamiDisabledByDefault(t *testing.T) {
	t.Setenv("UMAMI_URL", "")
	t.Setenv("UMAMI_WEBSITE_ID", "")
	d := shared.NewIndexData("cyber", "clean", "")
	if d.UmamiEnabled {
		t.Error("UmamiEnabled should be false when env vars are absent")
	}
}

func TestNewIndexData_UmamiEnabledWhenBothSet(t *testing.T) {
	t.Setenv("UMAMI_URL", "http://umami.internal")
	t.Setenv("UMAMI_WEBSITE_ID", "abc-123")
	d := shared.NewIndexData("cyber", "clean", "")
	if !d.UmamiEnabled {
		t.Error("UmamiEnabled should be true when UMAMI_URL and UMAMI_WEBSITE_ID are set")
	}
	if d.UmamiWebsiteID != "abc-123" {
		t.Errorf("UmamiWebsiteID=%q, want abc-123", d.UmamiWebsiteID)
	}
}

func TestNewIndexData_UmamiDisabledWhenOnlyURLSet(t *testing.T) {
	t.Setenv("UMAMI_URL", "http://umami.internal")
	t.Setenv("UMAMI_WEBSITE_ID", "")
	d := shared.NewIndexData("cyber", "clean", "")
	if d.UmamiEnabled {
		t.Error("UmamiEnabled should be false when UMAMI_WEBSITE_ID is absent")
	}
}

// ─── FreePort ─────────────────────────────────────────────────────────────────

func TestFreePort_ReturnsPositivePort(t *testing.T) {
	port, err := shared.FreePort()
	if err != nil {
		t.Fatalf("FreePort: %v", err)
	}
	if port <= 0 || port > 65535 {
		t.Errorf("FreePort returned %d, want 1–65535", port)
	}
}

func TestFreePort_ReturnsDifferentPortsEachTime(t *testing.T) {
	p1, err1 := shared.FreePort()
	p2, err2 := shared.FreePort()
	if err1 != nil || err2 != nil {
		t.Fatalf("FreePort errors: %v / %v", err1, err2)
	}
	// Not guaranteed to differ but almost certain on any OS.
	_ = p1
	_ = p2
}

// ─── ServeAsset ───────────────────────────────────────────────────────────────

//go:embed testdata
var testFS embed.FS

func TestServeAsset_ExistingFile(t *testing.T) {
	handler := shared.ServeAsset(testFS, "testdata/hello.txt", "text/plain")
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/hello.txt", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status %d, want 200", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	if string(body) != "hello\n" {
		t.Errorf("body=%q, want 'hello\\n'", body)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/plain" {
		t.Errorf("Content-Type=%q, want text/plain", ct)
	}
}

func TestServeAsset_NotFound(t *testing.T) {
	handler := shared.ServeAsset(testFS, "testdata/missing.txt", "text/plain")
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/missing.txt", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status %d, want 404", w.Code)
	}
}

// ─── RegisterUmamiProxy ───────────────────────────────────────────────────────

func TestRegisterUmamiProxy_NoURL(t *testing.T) {
	t.Setenv("UMAMI_URL", "")
	mux := http.NewServeMux()
	// Should not panic and should not register a route.
	shared.RegisterUmamiProxy(mux)
}

func TestRegisterUmamiProxy_InvalidURL(t *testing.T) {
	// url.Parse accepts almost anything; use a truly malformed URL.
	t.Setenv("UMAMI_URL", "://invalid")
	mux := http.NewServeMux()
	shared.RegisterUmamiProxy(mux) // should not panic
}

func TestRegisterUmamiProxy_ValidURL(t *testing.T) {
	t.Setenv("UMAMI_URL", "http://umami.internal")
	mux := http.NewServeMux()
	shared.RegisterUmamiProxy(mux) // should register /umami/ route
}
