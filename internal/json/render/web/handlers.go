package web

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/danielriddell21/unum/internal/json/lens/query"
)

const queryTimeout = 5 * time.Second

type queryRequest struct {
	Expr string `json:"expr"`
}

type queryResponse struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func handleQuery(doc *queryDoc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req queryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, queryResponse{Error: "invalid request: " + err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), queryTimeout)
		defer cancel()

		result, err := query.ExecuteAny(ctx, doc.anyForm(), req.Expr)
		if err != nil {
			writeJSON(w, queryResponse{Error: err.Error()})
			return
		}
		writeJSON(w, queryResponse{Result: result})
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
