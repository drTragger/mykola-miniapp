package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/drTragger/mykola-miniapp/internal/config"
	"github.com/drTragger/mykola-miniapp/internal/httpapi"
	"github.com/drTragger/mykola-miniapp/internal/metrics"
	"github.com/drTragger/mykola-miniapp/internal/system"
	"github.com/drTragger/mykola-miniapp/internal/telegram"
	"github.com/drTragger/mykola-miniapp/internal/ups"
)

func main() {
	cfg := config.Load()

	if err := metrics.InitHistory(); err != nil {
		log.Fatal(err)
	}

	if err := ups.InitHistory(); err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	metrics.StartBackgroundRefresh(ctx, 5*time.Second)
	ups.StartBackgroundRefresh(ctx, 10*time.Second)
	system.StartBackgroundRefresh(ctx, 15*time.Second)

	go runTelegramBot(ctx, cfg)

	handler, err := httpapi.NewRouter()
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:              cfg.App.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Println("Mini App server started on", cfg.App.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down HTTP server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("Shutdown complete")
}

func runTelegramBot(ctx context.Context, cfg config.Config) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("telegram bot panic recovered: %v", r)
		}
	}()

	telegram.StartBot(ctx, cfg)
}
