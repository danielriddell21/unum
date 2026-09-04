package web

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"go.opentelemetry.io/otel/attribute"

	"github.com/danielriddell21/unum/internal/image/optimize"
	"github.com/danielriddell21/unum/internal/telemetry"
)

const maxUploadBytes = 64 << 20

type uploadResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Format string `json:"format"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Bytes  int    `json:"bytes"`
}

type optimizeResponse struct {
	DataURL string  `json:"dataUrl"`
	Format  string  `json:"format"`
	Quality int     `json:"quality"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	Bytes   int     `json:"bytes"`
	Saving  float64 `json:"saving"`
}

type analyzeResponse struct {
	Steps []optimize.Step `json:"steps"`
}

func handleUpload(tel *telemetry.Telemetry, st *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}

		data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxUploadBytes))
		if err != nil {
			http.Error(w, "image too large", http.StatusRequestEntityTooLarge)
			return
		}
		if len(data) == 0 {
			http.Error(w, "empty upload", http.StatusBadRequest)
			return
		}

		name := r.URL.Query().Get("name")
		if name == "" {
			name = "image"
		}

		id, src, err := st.put(name, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, span := tel.Tracer().Start(r.Context(), "image.upload")
		span.SetAttributes(
			attribute.String("image.format", src.Format.String()),
			attribute.Int("image.bytes", src.Bytes),
		)
		span.End()
		tel.TrackEvent("image-upload", "/api/upload", map[string]string{
			"format": src.Format.String(),
		})

		writeJSON(w, uploadResponse{
			ID:     id,
			Name:   src.Name,
			Format: src.Format.String(),
			Width:  src.Width,
			Height: src.Height,
			Bytes:  src.Bytes,
		})
	}
}

func handleOptimize(tel *telemetry.Telemetry, st *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		src, opts, ok := requestOptions(w, r, st)
		if !ok {
			return
		}

		ctx, span := tel.Tracer().Start(r.Context(), "image.optimize")
		defer span.End()
		_ = ctx

		result, err := optimize.Optimize(src, opts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		span.SetAttributes(
			attribute.Int("image.output_bytes", result.Bytes),
			attribute.Int("image.quality", result.Quality),
		)

		writeJSON(w, optimizeResponse{
			DataURL: dataURL(result.Format.MIME(), result.Data),
			Format:  result.Format.String(),
			Quality: result.Quality,
			Width:   result.Width,
			Height:  result.Height,
			Bytes:   result.Bytes,
			Saving:  optimize.Saving(src.Bytes, result.Bytes),
		})
	}
}

func handleAnalyze(tel *telemetry.Telemetry, st *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		src, opts, ok := requestOptions(w, r, st)
		if !ok {
			return
		}

		_, span := tel.Tracer().Start(r.Context(), "image.analyze")
		defer span.End()

		steps, err := optimize.Ladder(src, opts, optimize.DefaultLadder)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		writeJSON(w, analyzeResponse{Steps: steps})
	}
}

func requestOptions(w http.ResponseWriter, r *http.Request, st *store) (optimize.Source, optimize.Options, bool) {
	// r.URL.Query() throws away every value when the query is malformed, which
	// would silently hand back defaults instead of reporting the bad request.
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "malformed query string", http.StatusBadRequest)
		return optimize.Source{}, optimize.Options{}, false
	}

	id := q.Get("id")
	if id == "" {
		http.Error(w, "missing id param", http.StatusBadRequest)
		return optimize.Source{}, optimize.Options{}, false
	}
	src, ok := st.get(id)
	if !ok {
		http.Error(w, "unknown image id — upload it again", http.StatusNotFound)
		return optimize.Source{}, optimize.Options{}, false
	}

	format := optimize.FormatUnknown
	if raw := q.Get("format"); raw != "" {
		parsed, valid := optimize.ParseFormat(raw)
		if !valid {
			http.Error(w, "unknown format "+raw, http.StatusBadRequest)
			return optimize.Source{}, optimize.Options{}, false
		}
		format = parsed
	}

	scale := 1.0
	if raw := q.Get("scale"); raw != "" {
		parsed, err := optimize.ParseScale(raw)
		if err != nil {
			http.Error(w, "scale "+err.Error(), http.StatusBadRequest)
			return optimize.Source{}, optimize.Options{}, false
		}
		scale = parsed
	}

	return src, optimize.Options{
		Quality:   intParam(q.Get("quality"), 80),
		Scale:     scale,
		MaxWidth:  intParam(q.Get("maxWidth"), 0),
		MaxHeight: intParam(q.Get("maxHeight"), 0),
		Format:    optimize.ResolveOutputFormat(format, src.Format),
	}, true
}

func intParam(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}

func dataURL(mime string, data []byte) string {
	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(data))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
