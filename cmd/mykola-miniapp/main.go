package main

import (
	"log"
	"net/http"

	"github.com/drTragger/mykola-miniapp/internal/config"
	"github.com/drTragger/mykola-miniapp/internal/httpapi"
)

func main() {
	cfg := config.Load()

	handler, err := httpapi.NewRouter()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Mini App server started on", cfg.AppAddr)

	if err := http.ListenAndServe(cfg.AppAddr, handler); err != nil {
		log.Fatal(err)
	}
}
