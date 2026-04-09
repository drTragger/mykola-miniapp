package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/drTragger/mykola-miniapp/internal/processes"
)

func handleProcessesList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	items, total, err := processes.ListProcesses(processes.ListParams{
		Search: q.Get("search"),
		Sort:   q.Get("sort"),
		Dir:    q.Get("dir"),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, processes.ListResponse{
			OK:    false,
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, processes.ListResponse{
		OK:    true,
		Items: items,
		Total: total,
	})
}

func handleProcessAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/processes/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[1] != "action" {
		http.NotFound(w, r)
		return
	}

	pid64, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, processes.ActionResponse{
			OK:    false,
			Error: "invalid pid",
		})
		return
	}

	var req processes.ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, processes.ActionResponse{
			OK:    false,
			Error: "invalid body",
		})
		return
	}

	if err := processes.SignalProcess(int32(pid64), req.Action); err != nil {
		writeJSON(w, http.StatusBadRequest, processes.ActionResponse{
			OK:    false,
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, processes.ActionResponse{
		OK:      true,
		Message: "action sent",
	})
}
