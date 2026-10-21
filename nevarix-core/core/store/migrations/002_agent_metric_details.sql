ALTER TABLE agent_metric_points
    ADD COLUMN IF NOT EXISTS details_json JSONB NOT NULL DEFAULT '{}'::jsonb;
