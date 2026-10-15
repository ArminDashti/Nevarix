-- Renames tables created under the old "hub" naming; no-op on fresh databases.
ALTER TABLE IF EXISTS hub_settings RENAME TO core_settings;
ALTER TABLE IF EXISTS hub_endpoints RENAME TO core_endpoints;

CREATE TABLE IF NOT EXISTS core_settings (
    id          SMALLINT PRIMARY KEY CHECK (id = 1),
    default_cidr TEXT NOT NULL,
    agent_port  INTEGER NOT NULL,
    timeout_ms  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS core_endpoints (
    ip        TEXT PRIMARY KEY,
    hostname  TEXT NOT NULL DEFAULT '',
    added_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_hosts (
    agent_id     TEXT PRIMARY KEY,
    hostname     TEXT NOT NULL DEFAULT '',
    last_seen_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_metric_latest (
    agent_id               TEXT PRIMARY KEY REFERENCES agent_hosts(agent_id) ON DELETE CASCADE,
    hostname               TEXT NOT NULL DEFAULT '',
    cpu_percent            DOUBLE PRECISION NOT NULL DEFAULT 0,
    cores_json             JSONB NOT NULL DEFAULT '[]'::jsonb,
    memory_total_bytes     BIGINT NOT NULL DEFAULT 0,
    memory_used_bytes      BIGINT NOT NULL DEFAULT 0,
    memory_available_bytes BIGINT NOT NULL DEFAULT 0,
    memory_used_percent    DOUBLE PRECISION NOT NULL DEFAULT 0,
    disks_json             JSONB NOT NULL DEFAULT '[]'::jsonb,
    bandwidth_rx_bps       DOUBLE PRECISION NOT NULL DEFAULT 0,
    bandwidth_tx_bps       DOUBLE PRECISION NOT NULL DEFAULT 0,
    uptime_seconds         BIGINT NOT NULL DEFAULT 0,
    system_json            JSONB NOT NULL DEFAULT '{}'::jsonb,
    collected_at           TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_metric_points (
    id          BIGSERIAL PRIMARY KEY,
    agent_id    TEXT NOT NULL REFERENCES agent_hosts(agent_id) ON DELETE CASCADE,
    ts          TIMESTAMPTZ NOT NULL,
    cpu_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    mem_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    rx_bps      DOUBLE PRECISION NOT NULL DEFAULT 0,
    tx_bps      DOUBLE PRECISION NOT NULL DEFAULT 0,
    UNIQUE (agent_id, ts)
);

CREATE INDEX IF NOT EXISTS idx_agent_metric_points_agent_ts
    ON agent_metric_points (agent_id, ts DESC);

CREATE TABLE IF NOT EXISTS softether_sessions (
    id              BIGSERIAL PRIMARY KEY,
    username        TEXT NOT NULL,
    client_ip       TEXT,
    asn             TEXT,
    download_bytes  BIGINT NOT NULL DEFAULT 0,
    upload_bytes    BIGINT NOT NULL DEFAULT 0,
    connected_at    TIMESTAMPTZ NOT NULL,
    disconnected_at TIMESTAMPTZ,
    duration_seconds BIGINT NOT NULL DEFAULT 0,
    session_key     TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_softether_sessions_username
    ON softether_sessions (username);

CREATE INDEX IF NOT EXISTS idx_softether_sessions_connected
    ON softether_sessions (connected_at DESC);

CREATE TABLE IF NOT EXISTS softether_user_stats (
    username               TEXT PRIMARY KEY,
    client_ip              TEXT,
    asn                    TEXT,
    download_bytes         BIGINT NOT NULL DEFAULT 0,
    upload_bytes           BIGINT NOT NULL DEFAULT 0,
    usage_duration_seconds BIGINT NOT NULL DEFAULT 0,
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
