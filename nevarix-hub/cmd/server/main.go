package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/ArminDashti/nevarix-hub/internal/config"
	"github.com/ArminDashti/nevarix-hub/internal/httpapi"
	"github.com/ArminDashti/nevarix-hub/internal/store"
)

func main() {
	cfg := config.Load()
	if cfg.IngestToken == "" {
		log.Fatal("HUB_INGEST_TOKEN is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	st, err := store.Open(ctx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		log.Fatalf("open postgres: %v", err)
	}
	defer st.Close()

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.New(cfg, st).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("nevarix-hub listening on http://%s", cfg.HTTPAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
