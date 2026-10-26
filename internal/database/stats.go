package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"
	"zee-mirror/internal/domain"
)

func (db *DB) GetBotStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalTasks, completedTasks, failedTasks int
	var totalBandwidth int64

	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks").Scan(&totalTasks); err != nil {
		slog.Error("Database error in GetBotStats count", "error", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE status = 'completed'").Scan(&completedTasks); err != nil {
		slog.Error("Database error in GetBotStats completed", "error", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE status = 'failed'").Scan(&failedTasks); err != nil {
		slog.Error("Database error in GetBotStats failed", "error", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COALESCE(SUM(total_size), 0) FROM tasks WHERE status = 'completed'").Scan(&totalBandwidth); err != nil {
		slog.Error("Database error in GetBotStats bandwidth", "error", err)
	}

	stats["total_tasks"] = totalTasks
	stats["completed_tasks"] = completedTasks
	stats["failed_tasks"] = failedTasks
	stats["total_bandwidth"] = totalBandwidth

	return stats, nil
}

type UserStats = domain.UserStats

func (db *DB) GetUserStats(ctx context.Context, userID int64) (*UserStats, error) {
	stats := &UserStats{UserID: userID}

	if err := db.QueryRowContext(ctx, "SELECT COALESCE(username, '') FROM users WHERE id = $1", userID).Scan(&stats.Username); err != nil && !errors.Is(err, sql.ErrNoRows) {
		slog.Error("Database error in GetUserStats username", "error", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE user_id = $1", userID).Scan(&stats.TotalDownloads); err != nil {
		slog.Error("Database error in GetUserStats total", "error", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND status = 'completed'", userID).Scan(&stats.SuccessfulTasks); err != nil {
		slog.Error("Database error in GetUserStats successful", "error", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND status = 'failed'", userID).Scan(&stats.FailedTasks); err != nil {
		slog.Error("Database error in GetUserStats failed", "error", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COALESCE(SUM(total_size), 0) FROM tasks WHERE user_id = $1 AND status = 'completed'", userID).Scan(&stats.TotalBandwidth); err != nil {
		slog.Error("Database error in GetUserStats bandwidth", "error", err)
	}
	var lastActive sql.NullTime
	if err := db.QueryRowContext(ctx, "SELECT created_at FROM tasks WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1", userID).Scan(&lastActive); err != nil && !errors.Is(err, sql.ErrNoRows) {
		slog.Error("Database error in GetUserStats last active", "error", err)
	}
	if lastActive.Valid {
		stats.LastActive = db.fromDB(lastActive.Time)
	}
	if stats.LastActive.IsZero() {
		stats.LastActive = time.Now()
	}

	return stats, nil
}

type DailyStats = domain.DailyStats

// todayRange returns the start of today and of tomorrow in local time,
// converted for comparison with created_at.
func (db *DB) todayRange() (time.Time, time.Time) {
	y, m, d := time.Now().Date()
	start := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	return db.timeArg(start), db.timeArg(start.AddDate(0, 0, 1))
}

func (db *DB) GetTodayStats(ctx context.Context) (*DailyStats, error) {
	stats := &DailyStats{Date: time.Now()}
	start, end := db.todayRange()

	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE created_at >= $1 AND created_at < $2", start, end).Scan(&stats.TotalTasks); err != nil {
		slog.Error("Database error in GetTodayStats total", "error", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE created_at >= $1 AND created_at < $2 AND status = 'completed'", start, end).Scan(&stats.CompletedTasks); err != nil {
		slog.Error("Database error in GetTodayStats completed", "error", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE created_at >= $1 AND created_at < $2 AND status = 'failed'", start, end).Scan(&stats.FailedTasks); err != nil {
		slog.Error("Database error in GetTodayStats failed", "error", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COALESCE(SUM(total_size), 0) FROM tasks WHERE created_at >= $1 AND created_at < $2 AND status = 'completed'", start, end).Scan(&stats.TotalBandwidth); err != nil {
		slog.Error("Database error in GetTodayStats", "error", err)
	}

	return stats, nil
}

func (db *DB) GetUserTodayStats(ctx context.Context, userID int64) (*DailyStats, error) {
	stats := &DailyStats{Date: time.Now()}
	start, end := db.todayRange()

	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND created_at >= $2 AND created_at < $3", userID, start, end).Scan(&stats.TotalTasks); err != nil {
		slog.Error("Database error in GetUserTodayStats total", "error", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND created_at >= $2 AND created_at < $3 AND status = 'completed'", userID, start, end).Scan(&stats.CompletedTasks); err != nil {
		slog.Error("Database error in GetUserTodayStats completed", "error", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND created_at >= $2 AND created_at < $3 AND status = 'failed'", userID, start, end).Scan(&stats.FailedTasks); err != nil {
		slog.Error("Database error in GetUserTodayStats failed", "error", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COALESCE(SUM(total_size), 0) FROM tasks WHERE user_id = $1 AND created_at >= $2 AND created_at < $3 AND status = 'completed'", userID, start, end).Scan(&stats.TotalBandwidth); err != nil {
		slog.Error("Database error in GetUserTodayStats bandwidth", "error", err)
	}

	return stats, nil
}

func (db *DB) GetWeeklyStats(ctx context.Context) ([]DailyStats, error) {
	return db.dailyStats(ctx, 7)
}

func (db *DB) GetMonthlyStats(ctx context.Context) ([]DailyStats, error) {
	return db.dailyStats(ctx, 30)
}

// dailyStats returns one entry per local calendar day for the last `days`
// days, oldest first. Rows are bucketed in Go: SQL DATE() would use UTC on
// SQLite and the wall clock on Postgres.
func (db *DB) dailyStats(ctx context.Context, days int) ([]DailyStats, error) {
	y, m, d := time.Now().Date()
	tomorrow := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	first := tomorrow.AddDate(0, 0, -days)

	stats := make([]DailyStats, days)
	index := make(map[string]int, days)
	for i := range stats {
		day := first.AddDate(0, 0, i)
		stats[i].Date = day
		index[day.Format("2006-01-02")] = i
	}

	rows, err := db.QueryContext(ctx, `
		SELECT created_at, status, COALESCE(total_size, 0)
		FROM tasks WHERE created_at >= $1 AND created_at < $2
	`, db.timeArg(first), db.timeArg(tomorrow))
	if err != nil {
		slog.Error("Database error in dailyStats", "days", days, "error", err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var createdAt time.Time
		var status string
		var size int64
		if err := rows.Scan(&createdAt, &status, &size); err != nil {
			slog.Error("Database error scanning daily stats row", "error", err)
			continue
		}
		i, ok := index[db.fromDB(createdAt).Format("2006-01-02")]
		if !ok {
			continue
		}
		stats[i].TotalTasks++
		switch status {
		case "completed":
			stats[i].CompletedTasks++
			stats[i].TotalBandwidth += size
		case "failed":
			stats[i].FailedTasks++
		}
	}
	return stats, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
