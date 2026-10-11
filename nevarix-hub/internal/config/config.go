package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr      string
	DatabaseURL   string
	CORSOrigins   []string
	IngestToken   string
	HistoryPoints int
}

func Load() Config {
	return Config{
		HTTPAddr:      getenv("HUB_HTTP_ADDR", "127.0.0.1:8091"),
		DatabaseURL:   getenv("HUB_DATABASE_URL", "postgres://nevarix:nevarix@127.0.0.1:5432/nevarix?sslmode=disable"),
		CORSOrigins:   splitCSV(getenv("CORS_ORIGINS", "http://127.0.0.1:5173,http://localhost:5173")),
		IngestToken:   getenv("HUB_INGEST_TOKEN", ""),
		HistoryPoints: getenvInt("HUB_HISTORY_POINTS", 60),
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

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
