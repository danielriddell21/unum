package web

import (
	"encoding/json"
	"net/http"

	"github.com/danielriddell21/unum/internal/json/lens/query"
	"github.com/danielriddell21/unum/internal/json/node"
)

type queryRequest struct {
	Expr string `json:"expr"`
}

type queryResponse struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func handleQuery(root *node.Node) http.HandlerFunc {
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

		result, err := query.Execute(root, req.Expr)
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
