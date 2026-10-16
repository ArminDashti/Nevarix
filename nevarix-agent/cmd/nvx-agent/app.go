package main

import (
	"github.com/ArminDashti/nevarix-agent/internal/lifecycle"
)

// agentApp is the lifecycle descriptor for the agent (docker compose or native process).
var agentApp = lifecycle.App{
	Name:       "agent",
	CLI:        "nvx-agent",
	Binary:     "nevarix-agent",
	Pkg:        "./cmd/agent",
	Image:      "nevarix-agent:latest",
	Dockerfile: "dockerfile",
	Compose:    "docker-compose.yml",
}
