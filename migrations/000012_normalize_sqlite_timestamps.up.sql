-- Normalise DATETIME values to UTC text.
--
-- Before this release the SQLite driver wrote time.Time with Go's
-- time.Time.String() ("2026-10-26 13:30:00.123 +0700 WIB m=+0.01") in the
-- app's local zone, and scheduled_at held RFC 3339 strings
-- ("2026-10-26T20:30:00+07:00"). SQLite's date functions cannot read the
-- first form, and neither form compares correctly as text with
-- datetime('now') or with each other. The app now writes UTC as
-- "2006-01-02 15:04:05.999999999+00:00"; rewrite old rows to match.
-- Values already in SQLite's own "YYYY-MM-DD HH:MM:SS" form (from
-- CURRENT_TIMESTAMP / datetime('now')) are UTC already and stay as they are.

-- users.created_at: time.Time.String() form
UPDATE users
SET created_at = strftime('%Y-%m-%d %H:%M:%f', substr(created_at, 1, 19)
        || substr(substr(created_at, 20), 1, instr(substr(created_at, 20), ' ') - 1)
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 1, 3)
        || ':'
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 4, 2)) || '+00:00'
WHERE typeof(created_at) = 'text'
  AND created_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9] [0-9][0-9]:[0-9][0-9]:[0-9][0-9]* [+-][0-9][0-9][0-9][0-9] *'
  AND strftime('%Y-%m-%d %H:%M:%f', substr(created_at, 1, 19)
        || substr(substr(created_at, 20), 1, instr(substr(created_at, 20), ' ') - 1)
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 1, 3)
        || ':'
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 4, 2)) IS NOT NULL;

-- users.created_at: RFC 3339 form
UPDATE users
SET created_at = strftime('%Y-%m-%d %H:%M:%f', created_at) || '+00:00'
WHERE typeof(created_at) = 'text'
  AND created_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]*'
  AND strftime('%Y-%m-%d %H:%M:%f', created_at) IS NOT NULL;

-- users.expires_at: time.Time.String() form
UPDATE users
SET expires_at = strftime('%Y-%m-%d %H:%M:%f', substr(expires_at, 1, 19)
        || substr(substr(expires_at, 20), 1, instr(substr(expires_at, 20), ' ') - 1)
        || substr(substr(expires_at, 20), instr(substr(expires_at, 20), ' ') + 1, 3)
        || ':'
        || substr(substr(expires_at, 20), instr(substr(expires_at, 20), ' ') + 4, 2)) || '+00:00'
WHERE typeof(expires_at) = 'text'
  AND expires_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9] [0-9][0-9]:[0-9][0-9]:[0-9][0-9]* [+-][0-9][0-9][0-9][0-9] *'
  AND strftime('%Y-%m-%d %H:%M:%f', substr(expires_at, 1, 19)
        || substr(substr(expires_at, 20), 1, instr(substr(expires_at, 20), ' ') - 1)
        || substr(substr(expires_at, 20), instr(substr(expires_at, 20), ' ') + 1, 3)
        || ':'
        || substr(substr(expires_at, 20), instr(substr(expires_at, 20), ' ') + 4, 2)) IS NOT NULL;

-- users.expires_at: RFC 3339 form
UPDATE users
SET expires_at = strftime('%Y-%m-%d %H:%M:%f', expires_at) || '+00:00'
WHERE typeof(expires_at) = 'text'
  AND expires_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]*'
  AND strftime('%Y-%m-%d %H:%M:%f', expires_at) IS NOT NULL;

-- tasks.created_at: time.Time.String() form
UPDATE tasks
SET created_at = strftime('%Y-%m-%d %H:%M:%f', substr(created_at, 1, 19)
        || substr(substr(created_at, 20), 1, instr(substr(created_at, 20), ' ') - 1)
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 1, 3)
        || ':'
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 4, 2)) || '+00:00'
WHERE typeof(created_at) = 'text'
  AND created_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9] [0-9][0-9]:[0-9][0-9]:[0-9][0-9]* [+-][0-9][0-9][0-9][0-9] *'
  AND strftime('%Y-%m-%d %H:%M:%f', substr(created_at, 1, 19)
        || substr(substr(created_at, 20), 1, instr(substr(created_at, 20), ' ') - 1)
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 1, 3)
        || ':'
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 4, 2)) IS NOT NULL;

-- tasks.created_at: RFC 3339 form
UPDATE tasks
SET created_at = strftime('%Y-%m-%d %H:%M:%f', created_at) || '+00:00'
WHERE typeof(created_at) = 'text'
  AND created_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]*'
  AND strftime('%Y-%m-%d %H:%M:%f', created_at) IS NOT NULL;

-- tasks.completed_at: time.Time.String() form
UPDATE tasks
SET completed_at = strftime('%Y-%m-%d %H:%M:%f', substr(completed_at, 1, 19)
        || substr(substr(completed_at, 20), 1, instr(substr(completed_at, 20), ' ') - 1)
        || substr(substr(completed_at, 20), instr(substr(completed_at, 20), ' ') + 1, 3)
        || ':'
        || substr(substr(completed_at, 20), instr(substr(completed_at, 20), ' ') + 4, 2)) || '+00:00'
WHERE typeof(completed_at) = 'text'
  AND completed_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9] [0-9][0-9]:[0-9][0-9]:[0-9][0-9]* [+-][0-9][0-9][0-9][0-9] *'
  AND strftime('%Y-%m-%d %H:%M:%f', substr(completed_at, 1, 19)
        || substr(substr(completed_at, 20), 1, instr(substr(completed_at, 20), ' ') - 1)
        || substr(substr(completed_at, 20), instr(substr(completed_at, 20), ' ') + 1, 3)
        || ':'
        || substr(substr(completed_at, 20), instr(substr(completed_at, 20), ' ') + 4, 2)) IS NOT NULL;

-- tasks.completed_at: RFC 3339 form
UPDATE tasks
SET completed_at = strftime('%Y-%m-%d %H:%M:%f', completed_at) || '+00:00'
WHERE typeof(completed_at) = 'text'
  AND completed_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]*'
  AND strftime('%Y-%m-%d %H:%M:%f', completed_at) IS NOT NULL;

-- audit_logs.created_at: time.Time.String() form
UPDATE audit_logs
SET created_at = strftime('%Y-%m-%d %H:%M:%f', substr(created_at, 1, 19)
        || substr(substr(created_at, 20), 1, instr(substr(created_at, 20), ' ') - 1)
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 1, 3)
        || ':'
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 4, 2)) || '+00:00'
WHERE typeof(created_at) = 'text'
  AND created_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9] [0-9][0-9]:[0-9][0-9]:[0-9][0-9]* [+-][0-9][0-9][0-9][0-9] *'
  AND strftime('%Y-%m-%d %H:%M:%f', substr(created_at, 1, 19)
        || substr(substr(created_at, 20), 1, instr(substr(created_at, 20), ' ') - 1)
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 1, 3)
        || ':'
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 4, 2)) IS NOT NULL;

-- audit_logs.created_at: RFC 3339 form
UPDATE audit_logs
SET created_at = strftime('%Y-%m-%d %H:%M:%f', created_at) || '+00:00'
WHERE typeof(created_at) = 'text'
  AND created_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]*'
  AND strftime('%Y-%m-%d %H:%M:%f', created_at) IS NOT NULL;

-- scheduled_tasks.scheduled_at: time.Time.String() form
UPDATE scheduled_tasks
SET scheduled_at = strftime('%Y-%m-%d %H:%M:%f', substr(scheduled_at, 1, 19)
        || substr(substr(scheduled_at, 20), 1, instr(substr(scheduled_at, 20), ' ') - 1)
        || substr(substr(scheduled_at, 20), instr(substr(scheduled_at, 20), ' ') + 1, 3)
        || ':'
        || substr(substr(scheduled_at, 20), instr(substr(scheduled_at, 20), ' ') + 4, 2)) || '+00:00'
WHERE typeof(scheduled_at) = 'text'
  AND scheduled_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9] [0-9][0-9]:[0-9][0-9]:[0-9][0-9]* [+-][0-9][0-9][0-9][0-9] *'
  AND strftime('%Y-%m-%d %H:%M:%f', substr(scheduled_at, 1, 19)
        || substr(substr(scheduled_at, 20), 1, instr(substr(scheduled_at, 20), ' ') - 1)
        || substr(substr(scheduled_at, 20), instr(substr(scheduled_at, 20), ' ') + 1, 3)
        || ':'
        || substr(substr(scheduled_at, 20), instr(substr(scheduled_at, 20), ' ') + 4, 2)) IS NOT NULL;

-- scheduled_tasks.scheduled_at: RFC 3339 form
UPDATE scheduled_tasks
SET scheduled_at = strftime('%Y-%m-%d %H:%M:%f', scheduled_at) || '+00:00'
WHERE typeof(scheduled_at) = 'text'
  AND scheduled_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]*'
  AND strftime('%Y-%m-%d %H:%M:%f', scheduled_at) IS NOT NULL;

-- scheduled_tasks.created_at: time.Time.String() form
UPDATE scheduled_tasks
SET created_at = strftime('%Y-%m-%d %H:%M:%f', substr(created_at, 1, 19)
        || substr(substr(created_at, 20), 1, instr(substr(created_at, 20), ' ') - 1)
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 1, 3)
        || ':'
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 4, 2)) || '+00:00'
WHERE typeof(created_at) = 'text'
  AND created_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9] [0-9][0-9]:[0-9][0-9]:[0-9][0-9]* [+-][0-9][0-9][0-9][0-9] *'
  AND strftime('%Y-%m-%d %H:%M:%f', substr(created_at, 1, 19)
        || substr(substr(created_at, 20), 1, instr(substr(created_at, 20), ' ') - 1)
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 1, 3)
        || ':'
        || substr(substr(created_at, 20), instr(substr(created_at, 20), ' ') + 4, 2)) IS NOT NULL;

-- scheduled_tasks.created_at: RFC 3339 form
UPDATE scheduled_tasks
SET created_at = strftime('%Y-%m-%d %H:%M:%f', created_at) || '+00:00'
WHERE typeof(created_at) = 'text'
  AND created_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]*'
  AND strftime('%Y-%m-%d %H:%M:%f', created_at) IS NOT NULL;
