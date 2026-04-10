package web

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielriddell21/unum/internal/hash/types"
)

// testDerive is a minimal derive function that doesn't import hash (avoids cycle in test).
func testDerive(input string) types.Result {
	h := sha256.Sum256([]byte(input))
	n := binary.BigEndian.Uint16(h[0:2])
	port := uint16(1024 + n%(65535-1024+1))
	color := "#" + hex.EncodeToString(h[0:3])
	short := hex.EncodeToString(h[0:4])
	return types.Result{
		Input: input,
		Port:  port,
		UUID:  fmt.Sprintf("test-uuid-%s", short),
		Color: color,
		Short: short,
		Emoji: "🦊",
		Phrase: "alpha-bravo-charlie",
	}
}

func testAppendHistory(_ string) error { return nil }
func testLoadHistory() []types.HistoryEntry {
	return []types.HistoryEntry{{Input: "test", Time: time.Now()}}
}

func init() {
	SetFuncs(testDerive, testAppendHistory, testLoadHistory)
}

func TestHandleDerive_ReturnsResult(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/derive?input=my-service", nil)
	w := httptest.NewRecorder()

	handleDerive()(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", w.Code)
	}

	var r types.Result
	if err := json.NewDecoder(w.Body).Decode(&r); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if r.Input != "my-service" {
		t.Errorf("Input = %q, want my-service", r.Input)
	}
	if r.Port < 1024 {
		t.Errorf("Port %d below minimum 1024", r.Port)
	}
}

func TestHandleDerive_MissingInput(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/derive", nil)
	w := httptest.NewRecorder()

	handleDerive()(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", w.Code)
	}
}

func TestHandleDerive_Deterministic(t *testing.T) {
	makeReq := func() types.Result {
		req := httptest.NewRequest(http.MethodGet, "/api/derive?input=test-svc", nil)
		w := httptest.NewRecorder()
		handleDerive()(w, req)
		var r types.Result
		_ = json.NewDecoder(w.Body).Decode(&r)
		return r
	}

	r1 := makeReq()
	r2 := makeReq()

	if r1.Port != r2.Port || r1.Color != r2.Color {
		t.Error("derive endpoint is not deterministic")
	}
}

func TestHandleHistory_ReturnsJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/history", nil)
	w := httptest.NewRecorder()

	handleHistory()(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var entries []types.HistoryEntry
	if err := json.NewDecoder(w.Body).Decode(&entries); err != nil {
		t.Fatalf("decode history: %v", err)
	}
}
