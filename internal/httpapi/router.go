package httpapi

import (
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/drTragger/mykola-miniapp/internal/web"
)

func NewRouter() (http.Handler, error) {
	staticFiles, err := web.StaticFS()
	if err != nil {
		return nil, err
	}

	fileSystem := http.FS(staticFiles)
	fileServer := http.FileServer(fileSystem)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/metrics", metricsHandler)

	mux.Handle("/assets/", fileServer)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		if path == "" || path == "index.html" {
			serveIndex(w, staticFiles)
			return
		}

		if strings.HasPrefix(path, "assets/") {
			fileServer.ServeHTTP(w, r)
			return
		}

		serveIndex(w, staticFiles)
	})

	return logRequests(mux), nil
}

func serveIndex(w http.ResponseWriter, staticFiles fs.FS) {
	file, err := staticFiles.Open("index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.Copy(w, file)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
