package httpapi

import (
	"net/http"

	"github.com/drTragger/mykola-miniapp/internal/metrics"
)

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, metrics.Response{
			OK:    false,
			Error: "method not allowed",
		})
		return
	}

	data, err := metrics.Collect()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, metrics.Response{
			OK:    false,
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, data)
}
