package httpapi

import (
	"net/http"

	"github.com/drTragger/mykola-miniapp/internal/system"
)

func vpnSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, system.VPNSummaryResponse{
			OK:    false,
			Error: "method not allowed",
		})
		return
	}

	data, err := system.GetVPNSummary()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, system.VPNSummaryResponse{
			OK:    false,
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, data)
}
