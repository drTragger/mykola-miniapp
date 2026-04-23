package httpapi

import (
	"io"
	"io/fs"
	"log"
	"net/http"
	"runtime/debug"
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

	fileServer := http.FileServer(http.FS(staticFiles))

	mux := http.NewServeMux()
	apiMux := http.NewServeMux()

	mux.HandleFunc("/api/health", healthHandler)

	apiMux.HandleFunc("/api/metrics", metricsHandler)
	apiMux.HandleFunc("/api/ups", upsHandler)
	apiMux.HandleFunc("/api/ups/battery", upsBatteryHandler)
	apiMux.HandleFunc("/api/ups/history", upsHistoryHandler)
	apiMux.HandleFunc("/api/system", systemHandler)
	apiMux.HandleFunc("/api/vpn/summary", vpnSummaryHandler)
	apiMux.HandleFunc("/api/processes", handleProcessesList)
	apiMux.HandleFunc("/api/processes/", handleProcessAction)

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

	return recoverMiddleware(logRequests(mux)), nil
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

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic in %s %s: %v\n%s", r.Method, r.URL.Path, rec, debug.Stack())
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
