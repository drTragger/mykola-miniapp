package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/drTragger/mykola-miniapp/internal/metrics"
)

func metricsHistoryHandler(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	points, err := metrics.GetHistory(limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(metrics.HistoryResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(metrics.HistoryResponse{OK: true, Points: points})
}
