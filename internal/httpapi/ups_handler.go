package httpapi

import (
	"net/http"

	"github.com/drTragger/mykola-miniapp/internal/ups"
)

func upsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ups.Response{
			OK:    false,
			Error: "method not allowed",
		})
		return
	}

	data, err := ups.GetSnapshot()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ups.Response{
			OK:    false,
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, data)
}
