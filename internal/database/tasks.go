package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"zee-mirror/internal/domain"
)

type TaskRecord = domain.TaskRecord

// taskColumns is the SELECT list scanned by scanTask. Nullable text columns
// are wrapped in COALESCE: a NULL (for example quality on rows created before
// migration 000004) cannot be scanned into a string and would drop the row.
const taskColumns = "id, COALESCE(gid, ''), type, status, COALESCE(url, ''), COALESCE(file_name, ''), " +
	"COALESCE(local_path, ''), COALESCE(remote_path, ''), COALESCE(remote_url, ''), " +
	"COALESCE(total_size, 0), COALESCE(downloaded_size, 0), COALESCE(uploaded_size, 0), chat_id, user_id, created_at, " +
	"completed_at, COALESCE(zip, FALSE), COALESCE(unzip, FALSE), COALESCE(password, ''), COALESCE(error, ''), " +
	"COALESCE(retries, 0), COALESCE(quality, ''), COALESCE(md5, ''), notify_url"

type rowScanner interface {
	Scan(dest ...any) error
}

func (db *DB) scanTask(row rowScanner) (TaskRecord, error) {
	var t TaskRecord
	err := row.Scan(
		&t.ID, &t.GID, &t.Type, &t.Status, &t.URL, &t.FileName, &t.LocalPath, &t.RemotePath, &t.RemoteURL,
		&t.TotalSize, &t.DownloadedSize, &t.UploadedSize, &t.ChatID, &t.UserID, &t.CreatedAt,
		&t.CompletedAt, &t.Zip, &t.Unzip, &t.Password, &t.Error, &t.RetryCount, &t.Quality, &t.MD5, &t.NotifyURL,
	)
	if err != nil {
		return t, err
	}
	t.CreatedAt = db.fromDB(t.CreatedAt)
	t.CompletedAt = db.fromDBNull(t.CompletedAt)
	return t, nil
}

func (db *DB) queryTasks(ctx context.Context, caller, query string, args ...any) ([]TaskRecord, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tasks []TaskRecord
	for rows.Next() {
		t, err := db.scanTask(rows)
		if err != nil {
			slog.Error("Error scanning task", "caller", caller, "error", err)
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (db *DB) ListTasks(ctx context.Context, filter domain.TaskFilter) ([]TaskRecord, error) {
	query := "SELECT " + taskColumns + " FROM tasks WHERE 1=1"
	var args []any
	arg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if filter.UserID > 0 {
		query += " AND user_id = " + arg(filter.UserID)
	}
	if filter.Status != "" {
		query += " AND status = " + arg(filter.Status)
	}

	query += " ORDER BY created_at DESC"

	switch {
	case filter.Limit > 0:
		query += " LIMIT " + arg(filter.Limit)
	case filter.Offset > 0:
		// SQLite rejects OFFSET without LIMIT; -1 and ALL mean "no limit".
		if db.driver == driverPostgres {
			query += " LIMIT ALL"
		} else {
			query += " LIMIT -1"
		}
	}
	if filter.Offset > 0 {
		query += " OFFSET " + arg(filter.Offset)
	}

	return db.queryTasks(ctx, "ListTasks", query, args...)
}

func (db *DB) GetCompletedTaskByURL(ctx context.Context, url, quality string) (*TaskRecord, error) {
	tr, err := db.scanTask(db.QueryRowContext(ctx, `
		SELECT `+taskColumns+` FROM tasks WHERE url = $1 AND COALESCE(quality, '') = $2 AND status = 'completed'
		ORDER BY created_at DESC LIMIT 1
	`, url, quality))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &tr, nil
}

func (db *DB) Save(ctx context.Context, t TaskRecord) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO tasks (
			id, gid, type, status, url, file_name, local_path, remote_path, remote_url,
			total_size, downloaded_size, uploaded_size, chat_id, user_id, created_at,
			completed_at, zip, unzip, password, error, retries, quality, notify_url
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
		ON CONFLICT(id) DO UPDATE SET
			gid = excluded.gid,
			status = excluded.status,
			file_name = excluded.file_name,
			local_path = excluded.local_path,
			remote_path = excluded.remote_path,
			remote_url = excluded.remote_url,
			total_size = excluded.total_size,
			downloaded_size = excluded.downloaded_size,
			uploaded_size = excluded.uploaded_size,
			completed_at = excluded.completed_at,
			error = excluded.error,
			retries = excluded.retries,
			quality = excluded.quality
	`, t.ID, t.GID, t.Type, t.Status, t.URL, t.FileName, t.LocalPath, t.RemotePath, t.RemoteURL,
		t.TotalSize, t.DownloadedSize, t.UploadedSize, t.ChatID, t.UserID, db.timeArg(t.CreatedAt),
		db.nullTimeArg(t.CompletedAt), t.Zip, t.Unzip, t.Password, t.Error, t.RetryCount, t.Quality, t.NotifyURL)
	return err
}

func (db *DB) GetActive(ctx context.Context) ([]TaskRecord, error) {
	return db.queryTasks(ctx, "GetActive",
		"SELECT "+taskColumns+" FROM tasks WHERE status NOT IN ('completed', 'failed', 'cancelled')")
}

func (db *DB) GetTaskByID(ctx context.Context, id string) (*TaskRecord, error) {
	tr, err := db.scanTask(db.QueryRowContext(ctx, "SELECT "+taskColumns+" FROM tasks WHERE id = $1", id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &tr, nil
}

func (db *DB) UpdateStatus(ctx context.Context, taskID, status, errorMsg string) error {
	_, err := db.ExecContext(ctx, "UPDATE tasks SET status = $1, error = $2 WHERE id = $3", status, errorMsg, taskID)
	return err
}

func (db *DB) UpdateMD5(ctx context.Context, id, md5 string) error {
	_, err := db.ExecContext(ctx, "UPDATE tasks SET md5=$1 WHERE id=$2", md5, id)
	return err
}

func (db *DB) SetNotifyURLByUser(ctx context.Context, userID int64, url string) error {
	_, err := db.ExecContext(ctx, "UPDATE tasks SET notify_url=$1 WHERE user_id=$2 AND status NOT IN ('completed','failed','cancelled')", url, userID)
	return err
}

// parseLocalTime parses a caller-supplied time string. Strings without a zone
// are taken as time.Local.
func parseLocalTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999999-07:00"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	for _, layout := range []string{"2006-01-02 15:04:05.999999999", "2006-01-02T15:04:05.999999999", "2006-01-02 15:04", "2006-01-02"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised time %q", s)
}

// DeleteOld removes finished tasks created before the given time
// ("2006-01-02 15:04:05" in local time, or RFC 3339).
func (db *DB) DeleteOld(ctx context.Context, before string) (int, error) {
	cutoff, err := parseLocalTime(before)
	if err != nil {
		return 0, err
	}
	result, err := db.ExecContext(ctx, "DELETE FROM tasks WHERE status IN ('completed', 'failed', 'cancelled') AND created_at < $1", db.timeArg(cutoff))
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return int(affected), nil
}

func (db *DB) GetRecentLogs(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, type, status, COALESCE(file_name, ''), created_at, COALESCE(error, '')
		FROM tasks
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var logs []map[string]interface{}
	for rows.Next() {
		var id, taskType, status, fileName, errorStr string
		var createdAt time.Time
		if err := rows.Scan(&id, &taskType, &status, &fileName, &createdAt, &errorStr); err != nil {
			slog.Error("Error scanning task in GetRecentLogs", "error", err)
			continue
		}

		level := "info"
		message := fmt.Sprintf("[%s] %s - %s", taskType, fileName, status)
		switch status {
		case "failed":
			level = "error"
			message = fmt.Sprintf("[%s] %s - %s: %s", taskType, fileName, status, errorStr)
		case "completed":
			level = "success"
		}

		logs = append(logs, map[string]interface{}{
			"level":     level,
			"message":   message,
			"timestamp": db.fromDB(createdAt),
		})
	}
	return logs, rows.Err()
}
