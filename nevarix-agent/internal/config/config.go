package config

import (
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ArminDashti/nevarix-agent/internal/core"
)

type Config struct {
	CoreURL         string
	GRPCAddr        string
	Transport       string
	AgentToken      string
	AgentID         string
	Hostname        string
	CollectInterval time.Duration
	ListenAddr      string
	OutboxPath      string
	HistoryPoints   int
	DockerMax       int
}

func Load() Config {
	hostname := getenv("NEVARIX_HOSTNAME", "")
	if hostname == "" {
		hostname, _ = os.Hostname()
	}
	agentID := getenv("NEVARIX_AGENT_ID", "")
	if agentID == "" {
		agentID = hostname
	}
	coreURL := strings.TrimRight(getenv("NEVARIX_CORE_URL", "http://127.0.0.1:8090"), "/")
	if st, err := core.Load(core.StatePath()); err != nil {
		log.Printf("core state: %v", err)
	} else if u, ok := st.CoreURL(); ok {
		coreURL = u
	}
	transport := strings.ToLower(getenv("NEVARIX_TRANSPORT", "auto"))
	grpcAddr := getenv("NEVARIX_GRPC_ADDR", "")
	if transport == "grpc" && grpcAddr == "" {
		grpcAddr = deriveGRPC(coreURL)
	}
	if transport == "auto" {
		if grpcAddr != "" {
			transport = "grpc"
		} else {
			transport = "rest"
		}
	}
	return Config{
		CoreURL:         coreURL,
		GRPCAddr:        grpcAddr,
		Transport:       transport,
		AgentToken:      getenv("NEVARIX_AGENT_TOKEN", ""),
		AgentID:         agentID,
		Hostname:        hostname,
		CollectInterval: time.Duration(getenvInt("COLLECT_INTERVAL_SECONDS", 5)) * time.Second,
		ListenAddr:      getenv("NEVARIX_LISTEN_ADDR", ":9000"),
		OutboxPath:      getenv("NEVARIX_OUTBOX_PATH", "data/agent-outbox.db"),
		HistoryPoints:   getenvInt("NEVARIX_HISTORY_POINTS", 60),
		DockerMax:       getenvInt("NEVARIX_DOCKER_MAX_CONTAINERS", 50),
	}
}

func deriveGRPC(coreURL string) string {
	u, err := url.Parse(coreURL)
	if err != nil || u.Hostname() == "" {
		return "127.0.0.1:50051"
	}
	return net.JoinHostPort(u.Hostname(), "50051")
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
