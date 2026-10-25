-- Postgres never got the 000008 audit_logs migration; numbered 000011 so
-- databases already at version 10 still pick it up.
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    action TEXT NOT NULL,
    actor_id BIGINT DEFAULT 0,
    actor_name TEXT DEFAULT '',
    resource TEXT DEFAULT '',
    resource_id TEXT DEFAULT '',
    details TEXT DEFAULT '',
    ip_address TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs(actor_id);
