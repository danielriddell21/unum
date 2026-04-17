package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleQuery_ValidExpression(t *testing.T) {
	root := mustParse(t, `{"name": "alice", "age": 30}`)
	handler := handleQuery(root)

	body := `{"expr": ".name"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/query", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", w.Code)
	}
	var resp queryResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error != "" {
		t.Errorf("unexpected error: %s", resp.Error)
	}
	if !strings.Contains(resp.Result, "alice") {
		t.Errorf("result=%q, want to contain 'alice'", resp.Result)
	}
}

func TestHandleQuery_InvalidExpression(t *testing.T) {
	root := mustParse(t, `{"x": 1}`)
	handler := handleQuery(root)

	body := `{"expr": "!!invalid!!"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/query", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200 (errors are in body)", w.Code)
	}
	var resp queryResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error == "" {
		t.Error("expected error in response for invalid expression")
	}
}

func TestHandleQuery_MethodNotAllowed(t *testing.T) {
	root := mustParse(t, `{}`)
	handler := handleQuery(root)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/query", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET status %d, want 405", w.Code)
	}
}

func TestHandleQuery_InvalidJSONBody(t *testing.T) {
	root := mustParse(t, `{}`)
	handler := handleQuery(root)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/query", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	handler(w, req)

	var resp queryResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error == "" {
		t.Error("expected error for invalid JSON body")
	}
}

func TestWriteJSON_SetsContentType(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, map[string]string{"k": "v"})
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type=%q, want application/json", ct)
	}
}
