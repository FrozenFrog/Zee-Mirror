-- REAL is float4 in Postgres but an 8-byte double in SQLite. rate_limits.last_fetch
-- holds Unix seconds (about 1.8e9), which float4 can only store to the nearest
-- 128 seconds. Use DOUBLE PRECISION to match the SQLite schema.
ALTER TABLE rate_limits
    ALTER COLUMN tokens TYPE DOUBLE PRECISION,
    ALTER COLUMN last_fetch TYPE DOUBLE PRECISION;

ALTER TABLE task_checkpoints
    ALTER COLUMN progress TYPE DOUBLE PRECISION;
