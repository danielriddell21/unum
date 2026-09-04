package web

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/config"
	"github.com/danielriddell21/unum/internal/image/optimize"
	"github.com/danielriddell21/unum/internal/telemetry"
)

func testTelemetry(t *testing.T) *telemetry.Telemetry {
	t.Helper()
	disabled := false
	return telemetry.Init(config.Config{Telemetry: &disabled}, "unum-test", "test")
}

func testImageBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 48))
	for y := range 48 {
		for x := range 64 {
			img.Set(x, y, color.RGBA{uint8(x * 4), uint8(y * 5), 120, 255})
		}
	}
	data, err := optimize.Encode(img, optimize.FormatJPEG, 90)
	if err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	return data
}

func uploadFixture(t *testing.T, tel *telemetry.Telemetry, st *store) string {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/upload?name=fixture.jpg", bytes.NewReader(testImageBytes(t)))
	handleUpload(tel, st)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("upload: status %d, body %s", rec.Code, rec.Body)
	}
	var resp uploadResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	return resp.ID
}

func TestHandleUpload(t *testing.T) {
	tel := testTelemetry(t)
	st := newStore()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/upload?name=fixture.jpg", bytes.NewReader(testImageBytes(t)))
	handleUpload(tel, st)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200; body %s", rec.Code, rec.Body)
	}

	var resp uploadResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID == "" {
		t.Error("expected an id")
	}
	if resp.Name != "fixture.jpg" {
		t.Errorf("Name = %q, want fixture.jpg", resp.Name)
	}
	if resp.Format != "jpeg" || resp.Width != 64 || resp.Height != 48 {
		t.Errorf("got %s %dx%d, want jpeg 64x48", resp.Format, resp.Width, resp.Height)
	}
}

func TestHandleUploadRejections(t *testing.T) {
	tel := testTelemetry(t)

	tests := []struct {
		name   string
		method string
		body   string
		want   int
	}{
		{"GET not allowed", http.MethodGet, "", http.StatusMethodNotAllowed},
		{"empty body", http.MethodPost, "", http.StatusBadRequest},
		{"not an image", http.MethodPost, "definitely not an image", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, "/api/upload", strings.NewReader(tt.body))
			handleUpload(tel, newStore())(rec, req)

			if rec.Code != tt.want {
				t.Errorf("status %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestHandleUploadDefaultsTheName(t *testing.T) {
	tel := testTelemetry(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/upload", bytes.NewReader(testImageBytes(t)))
	handleUpload(tel, newStore())(rec, req)

	var resp uploadResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Name != "image" {
		t.Errorf("Name = %q, want the default of image", resp.Name)
	}
}

func TestHandleOptimize(t *testing.T) {
	tel := testTelemetry(t)
	st := newStore()
	id := uploadFixture(t, tel, st)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/optimize?id="+id+"&quality=50&scale=50%25", nil)
	handleOptimize(tel, st)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200; body %s", rec.Code, rec.Body)
	}

	var resp optimizeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Quality != 50 {
		t.Errorf("Quality = %d, want 50", resp.Quality)
	}
	if resp.Width != 32 || resp.Height != 24 {
		t.Errorf("dimensions = %dx%d, want 32x24 at half scale", resp.Width, resp.Height)
	}
	if !strings.HasPrefix(resp.DataURL, "data:image/jpeg;base64,") {
		t.Errorf("DataURL prefix = %.40q, want a jpeg data url", resp.DataURL)
	}
	if resp.Bytes <= 0 {
		t.Error("expected a positive byte count")
	}
}

func TestHandleOptimizeFormatConversion(t *testing.T) {
	tel := testTelemetry(t)
	st := newStore()
	id := uploadFixture(t, tel, st)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/optimize?id="+id+"&format=png", nil)
	handleOptimize(tel, st)(rec, req)

	var resp optimizeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Format != "png" {
		t.Errorf("Format = %q, want png", resp.Format)
	}
	if !strings.HasPrefix(resp.DataURL, "data:image/png;base64,") {
		t.Errorf("DataURL prefix = %.40q, want a png data url", resp.DataURL)
	}
}

func TestHandleOptimizeRejections(t *testing.T) {
	tel := testTelemetry(t)
	st := newStore()
	id := uploadFixture(t, tel, st)

	tests := []struct {
		name  string
		query string
		want  int
	}{
		{"missing id", "", http.StatusBadRequest},
		{"unknown id", "?id=0123456789abcdef", http.StatusNotFound},
		{"bad format", "?id=" + id + "&format=avif", http.StatusBadRequest},
		{"bad scale", "?id=" + id + "&scale=500%25", http.StatusBadRequest},
		{"malformed query", "?id=" + id + "&scale=50%", http.StatusBadRequest},
		{"unwritable format", "?id=" + id + "&format=tiff", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/optimize"+tt.query, nil)
			handleOptimize(tel, st)(rec, req)

			if rec.Code != tt.want {
				t.Errorf("status %d, want %d; body %s", rec.Code, tt.want, rec.Body)
			}
		})
	}
}

func TestHandleAnalyze(t *testing.T) {
	tel := testTelemetry(t)
	st := newStore()
	id := uploadFixture(t, tel, st)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/analyze?id="+id, nil)
	handleAnalyze(tel, st)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rec.Code)
	}

	var resp analyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Steps) != len(optimize.DefaultLadder) {
		t.Fatalf("got %d steps, want %d", len(resp.Steps), len(optimize.DefaultLadder))
	}
	for _, step := range resp.Steps {
		if step.Bytes <= 0 {
			t.Errorf("step q%d reported %d bytes", step.Quality, step.Bytes)
		}
	}
}

func TestHandleSource(t *testing.T) {
	data := testImageBytes(t)

	t.Run("serves the launch image", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handleSource(Options{SourceName: "photo.jpg", SourceData: data, SourceMIME: "image/jpeg"})(
			rec, httptest.NewRequest(http.MethodGet, "/api/source", nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("status %d, want 200", rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); got != "image/jpeg" {
			t.Errorf("Content-Type = %q, want image/jpeg", got)
		}
		if got := rec.Header().Get("X-Unum-Name"); got != "photo.jpg" {
			t.Errorf("X-Unum-Name = %q, want photo.jpg", got)
		}
		if !bytes.Equal(rec.Body.Bytes(), data) {
			t.Error("body did not match the source bytes")
		}
	})

	t.Run("404 when launched with no file", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handleSource(Options{})(rec, httptest.NewRequest(http.MethodGet, "/api/source", nil))

		if rec.Code != http.StatusNotFound {
			t.Errorf("status %d, want 404", rec.Code)
		}
	})
}

func TestStoreDeduplicatesIdenticalUploads(t *testing.T) {
	st := newStore()
	data := testImageBytes(t)

	first, _, err := st.put("a.jpg", data)
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	second, _, err := st.put("b.jpg", data)
	if err != nil {
		t.Fatalf("put: %v", err)
	}

	if first != second {
		t.Errorf("identical bytes produced different ids: %q and %q", first, second)
	}
}

func TestStoreEvictsOldestBeyondCapacity(t *testing.T) {
	st := newStore()

	var ids []string
	for i := range maxStored + 2 {
		img := image.NewRGBA(image.Rect(0, 0, 8+i, 8))
		data, err := optimize.Encode(img, optimize.FormatPNG, 100)
		if err != nil {
			t.Fatalf("encode: %v", err)
		}
		id, _, err := st.put("x.png", data)
		if err != nil {
			t.Fatalf("put: %v", err)
		}
		ids = append(ids, id)
	}

	if _, ok := st.get(ids[0]); ok {
		t.Error("the oldest entry should have been evicted")
	}
	if _, ok := st.get(ids[len(ids)-1]); !ok {
		t.Error("the newest entry should still be present")
	}
}

func TestStoreRejectsUndecodableData(t *testing.T) {
	if _, _, err := newStore().put("bad", []byte("nope")); err == nil {
		t.Error("expected an error storing something that is not an image")
	}
}

func TestIntParam(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		fallback int
		want     int
	}{
		{"parses", "42", 7, 42},
		{"empty falls back", "", 7, 7},
		{"garbage falls back", "abc", 7, 7},
		{"zero", "0", 7, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := intParam(tt.raw, tt.fallback); got != tt.want {
				t.Errorf("intParam(%q, %d) = %d, want %d", tt.raw, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestDataURL(t *testing.T) {
	if got := dataURL("image/png", []byte{0, 1, 2}); got != "data:image/png;base64,AAEC" {
		t.Errorf("dataURL() = %q", got)
	}
}
