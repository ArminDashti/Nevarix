CREATE TABLE IF NOT EXISTS agent_apps (
    agent_id     TEXT NOT NULL,
    name         TEXT NOT NULL,
    version      TEXT NOT NULL DEFAULT '',
    publisher    TEXT NOT NULL DEFAULT '',
    install_date TEXT NOT NULL DEFAULT '',
    source       TEXT NOT NULL DEFAULT '',
    updated_at   TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (agent_id, name)
);

CREATE TABLE IF NOT EXISTS agent_events (
    id          BIGSERIAL PRIMARY KEY,
    agent_id    TEXT NOT NULL,
    kind        TEXT NOT NULL,
    message     TEXT NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_agent_events_agent_ts
    ON agent_events (agent_id, occurred_at DESC);

CREATE TABLE IF NOT EXISTS core_actions (
    id         BIGSERIAL PRIMARY KEY,
    kind       TEXT NOT NULL,
    detail     TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
);
