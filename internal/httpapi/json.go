package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"syscall"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil && !isBrokenPipe(err) {
		log.Printf("write json: %v", err)
	}
}

func isBrokenPipe(err error) bool {
	return errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET)
}
