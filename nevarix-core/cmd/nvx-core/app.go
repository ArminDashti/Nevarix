package main

import (
	"strconv"

	"github.com/ArminDashti/nevarix-core/internal/lifecycle"
)

// coreApp is the lifecycle descriptor for the core server (docker compose or native process).
var coreApp = lifecycle.App{
	Name:        "core",
	CLI:         "nvx-core",
	Binary:      "nevarix-core",
	Pkg:         "./cmd/server",
	Image:       "nevarix-core:latest",
	Dockerfile:  "dockerfile",
	Compose:     "docker-compose.yml",
	DefaultPort: 8181,
	PortEnv:     "PUBLISH_PORT",
	URLPath:     "/",
	NativeEnv: func(port int) []string {
		return []string{"HTTP_ADDR=127.0.0.1:" + strconv.Itoa(port)}
	},
}
