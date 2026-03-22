package webui

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	initdata "github.com/telegram-mini-apps/init-data-golang"
)

type MeResponse struct {
	OK        bool   `json:"ok"`
	InitData  string `json:"initData,omitempty"`
	UserID    int64  `json:"userId,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Error     string `json:"error,omitempty"`
}

func NewHandler() (http.Handler, error) {
	staticFiles, err := staticFS()
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.FS(staticFiles))
	mux.Handle("/", fileServer)

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":   true,
			"time": time.Now().Format(time.RFC3339),
		})
	})

	mux.HandleFunc("/api/me", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, MeResponse{
				OK:    false,
				Error: "method not allowed",
			})
			return
		}

		var req struct {
			InitData string `json:"initData"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, MeResponse{
				OK:    false,
				Error: "invalid json",
			})
			return
		}

		if req.InitData == "" {
			writeJSON(w, http.StatusBadRequest, MeResponse{
				OK:    false,
				Error: "initData is required",
			})
			return
		}

		botToken := os.Getenv("BOT_TOKEN")
		if botToken == "" {
			writeJSON(w, http.StatusInternalServerError, MeResponse{
				OK:    false,
				Error: "BOT_TOKEN is not set",
			})
			return
		}

		if err := initdata.Validate(req.InitData, botToken, 24*time.Hour); err != nil {
			writeJSON(w, http.StatusUnauthorized, MeResponse{
				OK:    false,
				Error: "invalid initData: " + err.Error(),
			})
			return
		}

		parsed, err := initdata.Parse(req.InitData)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, MeResponse{
				OK:    false,
				Error: "failed to parse initData: " + err.Error(),
			})
			return
		}

		resp := MeResponse{
			OK:       true,
			InitData: req.InitData,
		}

		if parsed.User.ID != 0 {
			resp.UserID = parsed.User.ID
			resp.Username = parsed.User.Username
			resp.FirstName = parsed.User.FirstName
			resp.LastName = parsed.User.LastName
		}

		writeJSON(w, http.StatusOK, resp)
	})

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
