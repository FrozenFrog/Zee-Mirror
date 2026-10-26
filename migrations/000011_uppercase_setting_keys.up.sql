-- Settings keys are uppercase, like the environment variables they mirror
-- (rclone_config -> RCLONE_CONFIG, dashboard_msg_<chat> -> DASHBOARD_MSG_<chat>).
-- When several spellings of a key exist, keep the uppercase row (the one the
-- bot read), or else the first spelling, and drop the others so the rename
-- cannot hit the primary key.
DELETE FROM settings
WHERE key <> UPPER(key)
  AND EXISTS (
    SELECT 1 FROM settings other
    WHERE UPPER(other.key) = UPPER(settings.key)
      AND other.key <> settings.key
      AND (other.key = UPPER(other.key) OR other.key < settings.key)
  );

UPDATE settings SET key = UPPER(key) WHERE key <> UPPER(key);
