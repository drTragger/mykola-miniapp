package httpapi

import (
	"log"
	"net/http"

	"github.com/drTragger/mykola-miniapp/internal/web"
)

func NewRouter() (http.Handler, error) {
	staticFiles, err := web.StaticFS()
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.FS(staticFiles))

	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/metrics", metricsHandler)
	mux.Handle("/", fileServer)

	return logRequests(mux), nil
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
