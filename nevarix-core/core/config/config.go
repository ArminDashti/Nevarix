package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr             string
	GRPCAddr             string
	DatabaseURL          string
	CORSOrigins          []string
	AgentIngestToken     string
	MetricsHistoryPoints int
	SoftEtherContainer   string
	SoftEtherPassword    string
	SoftEtherHub         string
	SoftEtherEnabled     bool
	SoftEtherPollEvery   time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr:             getenv("HTTP_ADDR", "127.0.0.1:8090"),
		GRPCAddr:             getenv("GRPC_ADDR", "127.0.0.1:50051"),
		DatabaseURL:          getenv("DATABASE_URL", "postgres://nevarix:nevarix@127.0.0.1:5432/nevarix?sslmode=disable"),
		CORSOrigins:          splitCSV(getenv("CORS_ORIGINS", "http://127.0.0.1:5173,http://localhost:5173")),
		AgentIngestToken:     getenv("AGENT_INGEST_TOKEN", ""),
		MetricsHistoryPoints: getenvInt("METRICS_HISTORY_POINTS", 60),
		SoftEtherContainer:   getenv("SOFTETHER_CONTAINER", "nevarix-softether"),
		SoftEtherPassword:    getenv("SOFTETHER_PASSWORD", "nevarix"),
		SoftEtherHub:         getenv("SOFTETHER_HUB", "DEFAULT"),
		SoftEtherEnabled:     getenvBool("SOFTETHER_ENABLED", false),
		SoftEtherPollEvery:   time.Duration(getenvInt("SOFTETHER_POLL_SECONDS", 30)) * time.Second,
	}
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	n, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	if err != nil {
		return fallback
	}
	return n
}

func getenvBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	return v == "1" || v == "true" || v == "yes"
}

func splitCSV(raw string) []string {
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
