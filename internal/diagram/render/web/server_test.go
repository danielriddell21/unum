package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func postRender(t *testing.T, s *server, lang, format, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost,
		"/api/render?lang="+lang+"&format="+format, strings.NewReader(body))
	w := httptest.NewRecorder()
	s.handleRender(w, req)
	return w
}

func TestHandleRenderD2SVG(t *testing.T) {
	s := &server{}
	w := postRender(t, s, "d2", "svg", "x -> y")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200 (%s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "<svg") {
		t.Errorf("expected <svg in body:\n%s", w.Body.String())
	}
}

func TestHandleRenderD2Drawio(t *testing.T) {
	s := &server{}
	w := postRender(t, s, "d2", "drawio", "a -> b")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "mxGraphModel") {
		t.Errorf("expected mxGraphModel in body:\n%s", w.Body.String())
	}
}

func TestHandleRenderMermaidNonFlowchart(t *testing.T) {
	s := &server{}
	w := postRender(t, s, "mermaid", "svg", "pie title T\n \"A\" : 10\n \"B\" : 20\n")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200 (%s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "<svg") {
		t.Errorf("expected <svg for a non-flowchart mermaid type:\n%s", w.Body.String())
	}
}

func TestHandleRenderUnknownLang(t *testing.T) {
	s := &server{}
	w := postRender(t, s, "cobol", "svg", "x -> y")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", w.Code)
	}
}

func TestHandleIndex(t *testing.T) {
	s := &server{opts: Options{Source: "x -> y", Lang: "d2", Version: "test"}}
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	s.handleIndex(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "RENDER_CONFIG") || !strings.Contains(body, "render") {
		t.Errorf("index missing expected content:\n%s", body)
	}
}
