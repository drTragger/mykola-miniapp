package httpapi

import (
	"net/http"

	"github.com/drTragger/mykola-miniapp/internal/system"
)

func systemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, system.Response{
			OK:    false,
			Error: "method not allowed",
		})
		return
	}

	data, err := system.GetSnapshot()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, system.Response{
			OK:    false,
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, data)
}
