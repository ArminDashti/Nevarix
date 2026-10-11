package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArminDashti/nevarix-agent/internal/client"
	"github.com/ArminDashti/nevarix-agent/internal/collector"
	"github.com/ArminDashti/nevarix-agent/internal/config"
	"github.com/ArminDashti/nevarix-agent/internal/health"
	"github.com/ArminDashti/nevarix-agent/internal/outbox"
	"github.com/joho/godotenv"
)

const flushBatch = 100

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if cfg.AgentToken == "" {
		log.Fatal("NEVARIX_AGENT_TOKEN is required")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	box, err := outbox.Open(cfg.OutboxPath)
	if err != nil {
		log.Fatalf("open outbox: %v", err)
	}
	defer box.Close()

	col := collector.New(cfg.AgentID, cfg.Hostname)
	api := client.New(cfg.HubURL, cfg.AgentToken)

	go func() {
		id := health.Identity{Agent: "nevarix-agent", AgentID: cfg.AgentID, Hostname: cfg.Hostname}
		if err := health.Serve(ctx, cfg.ListenAddr, id); err != nil {
			log.Printf("health listener: %v", err)
		}
	}()

	log.Printf("nevarix-agent starting agentId=%s hostname=%s hub=%s interval=%s outbox=%s (timestamps UTC)",
		cfg.AgentID, cfg.Hostname, cfg.HubURL, cfg.CollectInterval, cfg.OutboxPath)

	// Resend anything left from a previous run before collecting new data.
	if err := flush(ctx, box, api); err != nil {
		log.Printf("flush outbox: %v", err)
	}

	ticker := time.NewTicker(cfg.CollectInterval)
	defer ticker.Stop()

	for {
		if err := collectAndQueue(col, box); err != nil {
			log.Printf("collect: %v", err)
		}
		if err := flush(ctx, box, api); err != nil {
			log.Printf("send metrics (kept in outbox): %v", err)
		}
		select {
		case <-ctx.Done():
			log.Printf("nevarix-agent shutting down")
			return
		case <-ticker.C:
		}
	}
}

// collectAndQueue writes the snapshot to SQLite first, so a crash or Hub outage cannot drop it.
func collectAndQueue(col *collector.Collector, box *outbox.Store) error {
	snap, err := col.Collect()
	if err != nil {
		return err
	}
	body, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	return box.Enqueue(context.Background(), body)
}

// flush sends queued snapshots oldest first and deletes each row only after the Hub acks it.
// It stops at the first failure so the Hub receives snapshots in order.
func flush(ctx context.Context, box *outbox.Store, api *client.Client) error {
	items, err := box.Pending(ctx, flushBatch)
	if err != nil {
		return err
	}
	for _, it := range items {
		sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		err := api.PostMetrics(sendCtx, it.Payload)
		cancel()
		if err != nil {
			return err
		}
		if err := box.Ack(ctx, it.ID); err != nil {
			return err
		}
	}
	return nil
}
