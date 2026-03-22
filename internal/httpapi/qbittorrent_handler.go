package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/drTragger/mykola-miniapp/internal/config"
	"github.com/drTragger/mykola-miniapp/internal/qbittorrent"
)

type qbittorrentHandler struct {
	client *qbittorrent.Client
}

type torrentActionRequest struct {
	Hashes      []string `json:"hashes"`
	DeleteFiles bool     `json:"deleteFiles"`
}

func newQBittorrentHandler() (*qbittorrentHandler, error) {
	cfg := config.Load()

	client, err := qbittorrent.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &qbittorrentHandler{
		client: client,
	}, nil
}

func (h *qbittorrentHandler) list(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, qbittorrent.ListResponse{
			OK:    false,
			Error: "method not allowed",
		})
		return
	}

	torrents, err := h.client.ListTorrents()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, qbittorrent.ListResponse{
			OK:    false,
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, qbittorrent.ListResponse{
		OK:       true,
		Torrents: torrents,
	})
}

func (h *qbittorrentHandler) pause(w http.ResponseWriter, r *http.Request) {
	h.handleHashesAction(w, r, func(req torrentActionRequest) error {
		return h.client.Pause(req.Hashes)
	}, "paused")
}

func (h *qbittorrentHandler) resume(w http.ResponseWriter, r *http.Request) {
	h.handleHashesAction(w, r, func(req torrentActionRequest) error {
		return h.client.Resume(req.Hashes)
	}, "resumed")
}

func (h *qbittorrentHandler) delete(w http.ResponseWriter, r *http.Request) {
	h.handleHashesAction(w, r, func(req torrentActionRequest) error {
		return h.client.Delete(req.Hashes, req.DeleteFiles)
	}, "deleted")
}

func (h *qbittorrentHandler) handleHashesAction(
	w http.ResponseWriter,
	r *http.Request,
	action func(req torrentActionRequest) error,
	message string,
) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, qbittorrent.ActionResponse{
			OK:    false,
			Error: "method not allowed",
		})
		return
	}

	var req torrentActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, qbittorrent.ActionResponse{
			OK:    false,
			Error: "invalid request body",
		})
		return
	}

	if len(req.Hashes) == 0 {
		writeJSON(w, http.StatusBadRequest, qbittorrent.ActionResponse{
			OK:    false,
			Error: "hashes are required",
		})
		return
	}

	if err := action(req); err != nil {
		writeJSON(w, http.StatusInternalServerError, qbittorrent.ActionResponse{
			OK:    false,
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, qbittorrent.ActionResponse{
		OK:      true,
		Message: message,
	})
}
