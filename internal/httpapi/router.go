package httpapi

import (
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/drTragger/mykola-miniapp/internal/config"
	"github.com/drTragger/mykola-miniapp/internal/web"
)

func NewRouter() (http.Handler, error) {
	cfg := config.Load()

	staticFiles, err := web.StaticFS()
	if err != nil {
		return nil, err
	}

	fileSystem := http.FS(staticFiles)
	fileServer := http.FileServer(fileSystem)

	mux := http.NewServeMux()
	apiMux := http.NewServeMux()

	mux.HandleFunc("/api/health", healthHandler)

	apiMux.HandleFunc("/api/metrics", metricsHandler)
	apiMux.HandleFunc("/api/ups", upsHandler)
	apiMux.HandleFunc("/api/ups/battery", upsBatteryHandler)
	apiMux.HandleFunc("/api/ups/history", upsHistoryHandler)
	apiMux.HandleFunc("/api/system", systemHandler)
	apiMux.HandleFunc("/api/vpn/summary", vpnSummaryHandler)

	qbHandler, err := newQBittorrentHandler()
	if err != nil {
		return nil, err
	}

	apiMux.HandleFunc("/api/qbittorrent/torrents", qbHandler.list)
	apiMux.HandleFunc("/api/qbittorrent/torrents/pause", qbHandler.pause)
	apiMux.HandleFunc("/api/qbittorrent/torrents/resume", qbHandler.resume)
	apiMux.HandleFunc("/api/qbittorrent/torrents/delete", qbHandler.delete)
	apiMux.HandleFunc("/api/qbittorrent/torrents/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/peers") {
			qbHandler.getTorrentPeers(w, r)
			return
		}

		http.NotFound(w, r)
	})

	mux.Handle("/api/", telegramAuthMiddleware(cfg, apiMux))

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
