package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr             string
	GRPCAddr             string
	DatabaseURL          string
	CORSOrigins          []string
	AgentIngestToken     string
	MetricsHistoryPoints int
}

func Load() Config {
	return Config{
		HTTPAddr:             getenv("HTTP_ADDR", "127.0.0.1:8090"),
		GRPCAddr:             getenv("GRPC_ADDR", "127.0.0.1:50051"),
		DatabaseURL:          getenv("DATABASE_URL", "postgres://nevarix:nevarix@127.0.0.1:5432/nevarix?sslmode=disable"),
		CORSOrigins:          splitCSV(getenv("CORS_ORIGINS", "http://127.0.0.1:5173,http://localhost:5173")),
		AgentIngestToken:     getenv("AGENT_INGEST_TOKEN", ""),
		MetricsHistoryPoints: getenvInt("METRICS_HISTORY_POINTS", 60),
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

func splitCSV(raw string) []string {
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
