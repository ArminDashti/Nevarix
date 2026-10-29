# nevarix-agent

Golang host metrics agent for Nevarix. Collects CPU, memory, physical SSD/HDD, NICs, motherboard, Docker, and hardware inventory, then sends snapshots to `nevarix-core` over REST or gRPC.

All timestamps are **UTC** (`collectedAt` is RFC3339 with `Z`).

The nested `system` object is **schemaVersion 2**. Top-level CPU/memory/disks/bandwidth fields stay for older dashboards.

History keys in ring buffers: `cpu`, `memory`, `nic:<name>`, `disk:<id>`, `docker:<containerId>`.

NIC and DIMM temperatures are often `null` on consumer hardware. Docker is skipped (non-fatal) when the CLI/socket is missing. Docker socket access is privileged.

## Quick start

```bash
cp .env.example .env
# set NEVARIX_AGENT_TOKEN to match AGENT_INGEST_TOKEN on nevarix-core
go run ./cmd/agent
```

## Install (Windows or Linux)

From the repo root, build the CLI and daemon, then install as an OS service (Administrator on Windows, root on Linux):

```bash
go build -o bin/nevarix-agent ./cmd/agent    # add .exe on Windows
go build -o bin/nvx-agent ./cmd/nvx-agent      # add .exe on Windows
sudo ./bin/nvx-agent install                   # omit sudo on Windows (elevated shell)
```

Defaults: binaries under `/opt/NevarixAgent` (Linux) or `%ProgramFiles%\NevarixAgent` (Windows); data under `/var/lib/NevarixAgent` or `%LOCALAPPDATA%\NevarixAgent`. The installer sets `NEVARIX_AGENT_DATA_DIR` and `NEVARIX_OUTBOX_PATH`. Set `NEVARIX_AGENT_TOKEN` in the service environment (Windows: machine env + service restart; Linux: `/etc/nevarix-agent/environment`) before the agent can ingest.

## Environment

| Variable | Default | Description |
|----------|---------|-------------|
| `NEVARIX_CORE_URL` | `http://127.0.0.1:8090` | Base URL of nevarix-core (REST) |
| `NEVARIX_TRANSPORT` | `auto` | `rest`, `grpc`, or `auto` (`auto` uses gRPC when `NEVARIX_GRPC_ADDR` is set) |
| `NEVARIX_GRPC_ADDR` | (empty) | gRPC address (default `host(core):50051` when transport is `grpc`) |
| `NEVARIX_OUTBOX_PATH` | `data/agent-outbox.db` | SQLite outbox; snapshots stay until Core acks them |
| `NEVARIX_AGENT_TOKEN` | (required) | Bearer token for ingest |
| `NEVARIX_AGENT_ID` | hostname | Stable agent id |
| `NEVARIX_HOSTNAME` | OS hostname | Reported hostname |
| `COLLECT_INTERVAL_SECONDS` | `5` | Collect/send interval |
| `NEVARIX_HISTORY_POINTS` | `60` | Local ring-buffer length (min 10) |
| `NEVARIX_DOCKER_MAX_CONTAINERS` | `50` | Cap on container stats per tick |

## Endpoints

- REST: `POST {NEVARIX_CORE_URL}/api/v1/ingest/metrics` (a row is deleted from the outbox only after a 2xx)
- gRPC: `nevarix.agent.v1.AgentService/IngestMetrics`

Header / metadata: `Authorization: Bearer <NEVARIX_AGENT_TOKEN>`
