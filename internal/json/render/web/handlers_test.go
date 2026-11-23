package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/danielriddell21/unum/internal/json/lens/query"
)

func TestHandleQuery_ValidExpression(t *testing.T) {
	root := mustParse(t, `{"name": "alice", "age": 30}`)
	handler := handleQuery(newQueryDoc(root))

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
	handler := handleQuery(newQueryDoc(root))

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
	handler := handleQuery(newQueryDoc(root))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/query", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET status %d, want 405", w.Code)
	}
}

func TestHandleQuery_InvalidJSONBody(t *testing.T) {
	root := mustParse(t, `{}`)
	handler := handleQuery(newQueryDoc(root))

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

func postQuery(ctx context.Context, t *testing.T, handler http.HandlerFunc, expr string) queryResponse {
	t.Helper()
	body, err := json.Marshal(queryRequest{Expr: expr})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/query", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	var resp queryResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

func TestHandleQuery_Timeout(t *testing.T) {
	handler := handleQuery(newQueryDoc(mustParse(t, `{}`)))

	// The request context carries the deadline here so the test need not wait
	// out the handler's own timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	resp := postQuery(ctx, t, handler, "[repeat(1)]")
	if resp.Error != query.ErrTimeout.Error() {
		t.Errorf("error=%q, want %q", resp.Error, query.ErrTimeout)
	}
	if resp.Result != "" {
		t.Errorf("result=%q, want empty", resp.Result)
	}
}

func TestHandleQuery_ClientDisconnectStopsQuery(t *testing.T) {
	handler := handleQuery(newQueryDoc(mustParse(t, `{}`)))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan queryResponse, 1)
	go func() { done <- postQuery(ctx, t, handler, "[repeat(1)]") }()

	select {
	case resp := <-done:
		if resp.Error == "" {
			t.Error("a cancelled request should not produce a result")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("handler kept running after the request context was cancelled")
	}
}

func TestHandleQuery_OutputTruncated(t *testing.T) {
	handler := handleQuery(newQueryDoc(mustParse(t, `{}`)))

	resp := postQuery(context.Background(), t, handler, `"x" * 20000000`)
	if !strings.Contains(resp.Error, "output truncated") {
		t.Errorf("error=%q, want an output truncated message", resp.Error)
	}
	if resp.Result != "" {
		t.Errorf("result has %d bytes, want none", len(resp.Result))
	}
}

func TestHandleQuery_ReusesAnyForm(t *testing.T) {
	doc := newQueryDoc(mustParse(t, `{"a": [1, 2, 3]}`))
	handler := handleQuery(doc)

	first := postQuery(context.Background(), t, handler, ".a | length")
	cached := doc.anyForm()
	second := postQuery(context.Background(), t, handler, ".a[0]")

	if first.Error != "" || second.Error != "" {
		t.Fatalf("unexpected errors: %q, %q", first.Error, second.Error)
	}
	if strings.TrimSpace(first.Result) != "3" || strings.TrimSpace(second.Result) != "1" {
		t.Errorf("results %q, %q", first.Result, second.Result)
	}
	// Maps are reference types, so equal pointers mean the same conversion
	// served both queries.
	if reflect.ValueOf(doc.anyForm()).Pointer() != reflect.ValueOf(cached).Pointer() {
		t.Error("any form was rebuilt between queries")
	}
}

func TestQueryDoc_AnyFormConvertsOnce(t *testing.T) {
	doc := newQueryDoc(mustParse(t, `{"k": 1}`))
	if doc.input != nil {
		t.Fatal("conversion should be lazy")
	}

	var wg sync.WaitGroup
	ptrs := make([]uintptr, 8)
	for i := range ptrs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ptrs[i] = reflect.ValueOf(doc.anyForm()).Pointer()
		}()
	}
	wg.Wait()

	for _, p := range ptrs {
		if p != ptrs[0] {
			t.Fatal("concurrent callers saw different conversions")
		}
	}
}

func TestHandleKeyedQuery_UsesCachedDoc(t *testing.T) {
	key := "keyed-query-test"
	uploads.put(key, newQueryDoc(mustParse(t, `{"v": 7}`)), nil)

	body, err := json.Marshal(queryRequest{Expr: ".v"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/query?key="+key, bytes.NewReader(body))
	w := httptest.NewRecorder()
	handleKeyedQuery()(w, req)

	var resp queryResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(resp.Result) != "7" {
		t.Errorf("result=%q error=%q", resp.Result, resp.Error)
	}
}
