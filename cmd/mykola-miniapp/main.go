package main

import (
	"log"
	"net/http"
	"time"

	"github.com/drTragger/mykola-miniapp/internal/config"
	"github.com/drTragger/mykola-miniapp/internal/httpapi"
	"github.com/drTragger/mykola-miniapp/internal/metrics"
)

func main() {
	cfg := config.Load()

	metrics.StartBackgroundRefresh(5 * time.Second)

	handler, err := httpapi.NewRouter()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Mini App server started on", cfg.AppAddr)

	if err := http.ListenAndServe(cfg.AppAddr, handler); err != nil {
		log.Fatal(err)
	}
}
