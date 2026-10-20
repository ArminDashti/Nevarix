# nevarix-core

Single Go service for Nevarix: agent connectors, management connectors, and the core.

## Layout

```
nevarix-core/
├── agent-connectors/       # talks to nevarix-agent installed on client hosts
│   ├── auth/               # shared ingest token (REST + gRPC)
│   ├── discovery/          # finds agents by probing GET /health
│   ├── restapi/            # POST /api/v1/ingest/metrics
│   ├── grpcapi/            # AgentService.IngestMetrics
│   ├── proto/              # agent.proto (source of truth)
│   └── pb/                 # generated code (buf generate)
├── management-connectors/  # talks to the management panel
│   ├── service/            # shared use cases (validation, endpoint merge, scan)
│   ├── restapi/            # /api/v1/... panel routes
│   ├── grpcapi/            # ManagementService
│   ├── proto/              # management.proto (source of truth)
│   └── pb/                 # generated code (buf generate)
├── core/                   # config, model, PostgreSQL store, local metrics, Docker
└── cmd/server/             # wires everything into one process
```

Dependency direction: `cmd` → connectors → `core`. `core` never imports a connector.

## Transports

| Client | REST | gRPC |
| --- | --- | --- |
| nevarix-agent (client hosts) | `POST /api/v1/ingest/metrics` | `nevarix.agent.v1.AgentService/IngestMetrics` |
| Management panel | `/api/v1/...` (all panel routes) | `nevarix.management.v1.ManagementService` |

Agent calls need `Authorization: Bearer <AGENT_INGEST_TOKEN>` on both transports.
Local-host and Docker routes are REST-only.

## Regenerate gRPC code

Requires `buf`, `protoc-gen-go` and `protoc-gen-go-grpc` on `PATH`.

```bash
cd agent-connectors && buf generate
cd management-connectors && buf generate
```

## Quick start

```bash
cp .env.example .env
go run ./cmd/server
```

REST listens on `http://127.0.0.1:8090`, gRPC on `127.0.0.1:50051`. Timestamps are **UTC**.

## Environment

See `.env.example`.
