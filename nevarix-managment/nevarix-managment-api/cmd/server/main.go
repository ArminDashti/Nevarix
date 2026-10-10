package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArminDashti/nevarix-api/internal/asn"
	"github.com/ArminDashti/nevarix-api/internal/config"
	"github.com/ArminDashti/nevarix-api/internal/dockerx"
	httpserver "github.com/ArminDashti/nevarix-api/internal/http"
	"github.com/ArminDashti/nevarix-api/internal/hubclient"
	"github.com/ArminDashti/nevarix-api/internal/metrics"
	"github.com/ArminDashti/nevarix-api/internal/softether"
	"github.com/ArminDashti/nevarix-api/internal/store"
	"github.com/joho/godotenv"
)

const softEtherKind = "softether"

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if cfg.HubIngestToken == "" {
		log.Fatal("HUB_INGEST_TOKEN is required")
	}

	box, err := store.Open(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer box.Close()

	collector := metrics.NewCollector(cfg.MetricsHistoryPoints)

	var dockerClient *dockerx.Client
	if dc, err := dockerx.New(); err != nil {
		log.Printf("docker client unavailable: %v", err)
	} else {
		dockerClient = dc
		defer dockerClient.Close()
	}

	resolver := asn.NewResolver()
	seClient := softether.New(cfg.SoftEtherContainer, cfg.SoftEtherPassword, cfg.SoftEtherHub, cfg.SoftEtherEnabled, resolver)
	hub := hubclient.New(cfg.HubURL, cfg.HubIngestToken)

	ctx, stopBackground := context.WithCancel(context.Background())
	defer stopBackground()
	if cfg.SoftEtherEnabled {
		go pollSoftEther(ctx, cfg, seClient, box, hub)
	}

	srv := httpserver.New(cfg, collector, dockerClient, seClient, hub)
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("nevarix-managment listening on http://%s", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	stopBackground()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}

// pollSoftEther queues each SoftEther snapshot in SQLite, then pushes the queue to Hub.
// A row is removed only after Hub acknowledges it, so Hub downtime loses nothing.
func pollSoftEther(ctx context.Context, cfg config.Config, se *softether.Client, box *store.Store, hub *hubclient.Client) {
	ticker := time.NewTicker(cfg.SoftEtherPollEvery)
	defer ticker.Stop()
	for {
		pollCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		sessions, err := se.ListOnlineSessions(pollCtx)
		if err != nil {
			log.Printf("softether poll: %v", err)
		} else if err := enqueueSoftEther(pollCtx, box, sessions); err != nil {
			log.Printf("softether enqueue: %v", err)
		}
		if err := flushSoftEther(pollCtx, box, hub); err != nil {
			log.Printf("softether flush: %v", err)
		}
		cancel()

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func enqueueSoftEther(ctx context.Context, box *store.Store, sessions []softether.OnlineSession) error {
	body, err := json.Marshal(map[string]any{"sessions": sessions})
	if err != nil {
		return err
	}
	return box.Enqueue(ctx, softEtherKind, body)
}

// flushSoftEther sends rows oldest first and stops at the first failure to keep ordering.
func flushSoftEther(ctx context.Context, box *store.Store, hub *hubclient.Client) error {
	rows, err := box.Pending(ctx, softEtherKind, 100)
	if err != nil {
		return err
	}
	for _, r := range rows {
		if err := hub.PostJSON(ctx, "/api/v1/ingest/softether", r.Payload); err != nil {
			return err
		}
		if err := box.Ack(ctx, r.ID); err != nil {
			return err
		}
	}
	return nil
}
