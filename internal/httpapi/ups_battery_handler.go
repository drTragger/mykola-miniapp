package httpapi

import (
	"net/http"

	"github.com/drTragger/mykola-miniapp/internal/ups"
)

func upsBatteryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ups.BatteryResponse{
			OK:    false,
			Error: "method not allowed",
		})
		return
	}

	data, err := ups.GetBatterySnapshot()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ups.BatteryResponse{
			OK:    false,
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, data)
}
