package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"zee-mirror/internal/domain"
)

// SaveScheduled stores a pending scheduled task. task.ScheduledAt must be
// RFC 3339 (or "2006-01-02 15:04:05" in local time).
func (db *DB) SaveScheduled(ctx context.Context, task domain.ScheduledTask) error {
	scheduledAt, err := parseLocalTime(task.ScheduledAt)
	if err != nil {
		return fmt.Errorf("scheduled_at: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO scheduled_tasks (id, task_type, url, file_name, chat_id, user_id, zip, unzip, password, quality, scheduled_at, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'pending', $12)
	`, task.ID, task.TaskType, task.URL, task.FileName, task.ChatID, task.UserID,
		boolToInt(task.Zip), boolToInt(task.Unzip), task.Password, task.Quality,
		db.timeArg(scheduledAt), db.timeArg(time.Now()))
	return err
}

// GetPendingScheduled returns the pending tasks that are due. ScheduledAt and
// CreatedAt come back as RFC 3339 in local time.
func (db *DB) GetPendingScheduled(ctx context.Context) ([]domain.ScheduledTask, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, task_type, url, file_name, chat_id, user_id, COALESCE(zip, 0), COALESCE(unzip, 0),
			COALESCE(password, ''), COALESCE(quality, ''), scheduled_at, status, COALESCE(task_id, ''), created_at
		FROM scheduled_tasks WHERE status = 'pending' AND scheduled_at <= $1
		ORDER BY scheduled_at
	`, db.timeArg(time.Now()))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tasks []domain.ScheduledTask
	for rows.Next() {
		var t domain.ScheduledTask
		var zipInt, unzipInt int
		var scheduledAt time.Time
		var createdAt sql.NullTime
		if err := rows.Scan(&t.ID, &t.TaskType, &t.URL, &t.FileName, &t.ChatID, &t.UserID, &zipInt, &unzipInt, &t.Password, &t.Quality, &scheduledAt, &t.Status, &t.TaskID, &createdAt); err != nil {
			return nil, err
		}
		t.Zip = zipInt == 1
		t.Unzip = unzipInt == 1
		t.ScheduledAt = db.fromDB(scheduledAt).Format(time.RFC3339)
		if createdAt.Valid {
			t.CreatedAt = db.fromDB(createdAt.Time).Format(time.RFC3339)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (db *DB) MarkScheduledDone(ctx context.Context, id, taskID string) error {
	_, err := db.ExecContext(ctx, "UPDATE scheduled_tasks SET status='done', task_id=$1 WHERE id=$2", taskID, id)
	return err
}

func (db *DB) DeleteScheduled(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM scheduled_tasks WHERE id=$1", id)
	return err
}
