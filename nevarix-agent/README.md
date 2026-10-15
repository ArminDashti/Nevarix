# nevarix-agent

Golang host metrics agent for Nevarix. Collects CPU, memory, disk, bandwidth, and uptime, then POSTs to `nevarix-core`.

All timestamps are **UTC** (`collectedAt` is RFC3339 with `Z`).

## Quick start

```bash
cp .env.example .env
# set NEVARIX_AGENT_TOKEN to match AGENT_INGEST_TOKEN on nevarix-core
go run ./cmd/agent
```

## Environment

| Variable | Default | Description |
|----------|---------|-------------|
| `NEVARIX_CORE_URL` | `http://127.0.0.1:8090` | Base URL of nevarix-core (REST) |
| `NEVARIX_OUTBOX_PATH` | `data/agent-outbox.db` | SQLite outbox; snapshots stay until Core acks them |
| `NEVARIX_AGENT_TOKEN` | (required) | Bearer token for ingest |
| `NEVARIX_AGENT_ID` | hostname | Stable agent id |
| `NEVARIX_HOSTNAME` | OS hostname | Reported hostname |
| `COLLECT_INTERVAL_SECONDS` | `5` | Collect/send interval |

## Endpoint

`POST {NEVARIX_CORE_URL}/api/v1/ingest/metrics` (a row is deleted from the outbox only after a 2xx)

Header: `Authorization: Bearer <NEVARIX_AGENT_TOKEN>`
