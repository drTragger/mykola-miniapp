package httpapi

import (
	"net/http"
	"strconv"

	"github.com/drTragger/mykola-miniapp/internal/ups"
)

func upsHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ups.HistoryResponse{
			OK:    false,
			Error: "method not allowed",
		})
		return
	}

	limit := 288

	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err == nil && parsed > 0 && parsed <= 2000 {
			limit = parsed
		}
	}

	points, err := ups.GetHistory(limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ups.HistoryResponse{
			OK:    false,
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, ups.HistoryResponse{
		OK:     true,
		Points: points,
	})
}
