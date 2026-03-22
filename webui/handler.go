package webui

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func NewHandler() (http.Handler, error) {
	staticFiles, err := staticFS()
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.FS(staticFiles))

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":   true,
			"time": time.Now().Format(time.RFC3339),
		})
	})

	mux.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, MetricsResponse{
				OK:    false,
				Error: "method not allowed",
			})
			return
		}

		metrics, err := collectMetrics()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, MetricsResponse{
				OK:    false,
				Error: err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, metrics)
	})

	mux.Handle("/", fileServer)

	return logRequests(mux), nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
