ALTER TABLE core_settings
    ADD COLUMN IF NOT EXISTS snmp_community TEXT NOT NULL DEFAULT 'public',
    ADD COLUMN IF NOT EXISTS net_poll_interval_sec INTEGER NOT NULL DEFAULT 60;

CREATE TABLE IF NOT EXISTS net_devices (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                TEXT NOT NULL DEFAULT '',
    mgmt_ip             INET NOT NULL UNIQUE,
    vendor              TEXT NOT NULL DEFAULT 'unknown',
    snmp_version        TEXT NOT NULL DEFAULT '2c',
    snmp_community      TEXT NOT NULL DEFAULT 'public',
    poll_interval_sec   INTEGER NOT NULL DEFAULT 60,
    enabled             BOOLEAN NOT NULL DEFAULT TRUE,
    sys_descr           TEXT NOT NULL DEFAULT '',
    sys_object_id       TEXT NOT NULL DEFAULT '',
    last_ok_at          TIMESTAMPTZ,
    last_error          TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS net_device_latest (
    device_id           UUID PRIMARY KEY REFERENCES net_devices(id) ON DELETE CASCADE,
    cpu_percent         DOUBLE PRECISION NOT NULL DEFAULT 0,
    memory_used_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    uptime_seconds      BIGINT NOT NULL DEFAULT 0,
    bandwidth_rx_bps    DOUBLE PRECISION NOT NULL DEFAULT 0,
    bandwidth_tx_bps    DOUBLE PRECISION NOT NULL DEFAULT 0,
    collected_at        TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS net_interfaces (
    device_id       UUID NOT NULL REFERENCES net_devices(id) ON DELETE CASCADE,
    if_index        INTEGER NOT NULL,
    name            TEXT NOT NULL DEFAULT '',
    oper_status     INTEGER NOT NULL DEFAULT 0,
    speed_bps       BIGINT NOT NULL DEFAULT 0,
    in_octets       BIGINT NOT NULL DEFAULT 0,
    out_octets      BIGINT NOT NULL DEFAULT 0,
    rx_bps          DOUBLE PRECISION NOT NULL DEFAULT 0,
    tx_bps          DOUBLE PRECISION NOT NULL DEFAULT 0,
    sampled_at      TIMESTAMPTZ,
    PRIMARY KEY (device_id, if_index)
);

CREATE TABLE IF NOT EXISTS net_metric_points (
    id          BIGSERIAL PRIMARY KEY,
    device_id   UUID NOT NULL REFERENCES net_devices(id) ON DELETE CASCADE,
    ts          TIMESTAMPTZ NOT NULL,
    cpu_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    mem_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    rx_bps      DOUBLE PRECISION NOT NULL DEFAULT 0,
    tx_bps      DOUBLE PRECISION NOT NULL DEFAULT 0,
    UNIQUE (device_id, ts)
);

CREATE INDEX IF NOT EXISTS idx_net_metric_points_device_ts
    ON net_metric_points (device_id, ts DESC);

CREATE TABLE IF NOT EXISTS net_events (
    id          BIGSERIAL PRIMARY KEY,
    device_id   UUID REFERENCES net_devices(id) ON DELETE SET NULL,
    source_ip   INET,
    severity    TEXT NOT NULL DEFAULT 'info',
    kind        TEXT NOT NULL DEFAULT 'syslog',
    message     TEXT NOT NULL DEFAULT '',
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_net_events_received
    ON net_events (received_at DESC);

CREATE TABLE IF NOT EXISTS net_neighbors (
    device_id       UUID NOT NULL REFERENCES net_devices(id) ON DELETE CASCADE,
    local_if_index  INTEGER NOT NULL DEFAULT 0,
    remote_id       TEXT NOT NULL DEFAULT '',
    remote_name     TEXT NOT NULL DEFAULT '',
    protocol        TEXT NOT NULL DEFAULT 'lldp',
    PRIMARY KEY (device_id, local_if_index, remote_id)
);
