package database

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"zee-mirror/internal/domain"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
)

// withLocal sets time.Local for the duration of the test. The app does the
// same at startup from TZ, and the time handling must not depend on it.
func withLocal(t *testing.T, loc *time.Location) {
	t.Helper()
	old := time.Local
	time.Local = loc
	t.Cleanup(func() { time.Local = old })
}

var testZones = []*time.Location{
	time.FixedZone("WIB", 7*3600),
	time.FixedZone("EST", -5*3600),
	time.UTC,
}

func sameInstant(a, b time.Time) bool {
	d := a.Sub(b)
	return d > -time.Millisecond && d < time.Millisecond
}

func TestNewDBRejectsUnknownDriver(t *testing.T) {
	for _, name := range []string{"", "pgx", "sqlite3", "mysql"} {
		if _, err := NewDB(name, t.TempDir(), "", testMigrationsDir); err == nil || !strings.Contains(err.Error(), "unsupported DB_DRIVER") {
			t.Errorf("NewDB(%q) err = %v, want unsupported DB_DRIVER", name, err)
		}
	}
}

func TestTimesRoundTrip(t *testing.T) {
	for _, loc := range testZones {
		t.Run(loc.String(), func(t *testing.T) {
			withLocal(t, loc)
			runOnEngines(t, func(t *testing.T, db *DB) {
				ctx := context.Background()
				created := time.Now().Add(-3 * time.Hour)
				expires := time.Now().Add(48 * time.Hour)
				completed := time.Now().Add(-time.Hour)

				if err := db.Upsert(ctx, domain.User{ID: 1, Username: "u", Role: "user", CreatedAt: created,
					ExpiresAt: sql.NullTime{Time: expires, Valid: true}}); err != nil {
					t.Fatalf("Upsert: %v", err)
				}
				u, err := db.GetByID(ctx, 1)
				if err != nil {
					t.Fatalf("GetByID: %v", err)
				}
				if !sameInstant(u.CreatedAt, created) || !u.ExpiresAt.Valid || !sameInstant(u.ExpiresAt.Time, expires) || !u.IsActive {
					t.Errorf("user times = %v / %v active=%v, want %v / %v active", u.CreatedAt, u.ExpiresAt, u.IsActive, created, expires)
				}

				// Expired an hour ago must be inactive in every zone.
				if err = db.SetExpiration(ctx, 1, time.Now().Add(-time.Hour)); err != nil {
					t.Fatalf("SetExpiration: %v", err)
				}
				if u, _ = db.GetByID(ctx, 1); u.IsActive {
					t.Errorf("user expired an hour ago is still active (expires_at %v)", u.ExpiresAt.Time)
				}

				if err = db.Save(ctx, domain.TaskRecord{ID: "t", Type: "mirror", Status: "completed", ChatID: 1, UserID: 1,
					CreatedAt: created, CompletedAt: sql.NullTime{Time: completed, Valid: true}}); err != nil {
					t.Fatalf("Save: %v", err)
				}
				tr, err := db.GetTaskByID(ctx, "t")
				if err != nil {
					t.Fatalf("GetTaskByID: %v", err)
				}
				if !sameInstant(tr.CreatedAt, created) || !tr.CompletedAt.Valid || !sameInstant(tr.CompletedAt.Time, completed) {
					t.Errorf("task times = %v / %v, want %v / %v", tr.CreatedAt, tr.CompletedAt.Time, created, completed)
				}

				if err = db.LogAudit(ctx, domain.AuditEntry{Action: "a"}); err != nil {
					t.Fatalf("LogAudit: %v", err)
				}
				entries, err := db.ListAuditLogs(ctx, 10, 0)
				if err != nil || len(entries) != 1 {
					t.Fatalf("ListAuditLogs = %v, %v", entries, err)
				}
				if d := time.Since(entries[0].CreatedAt); d < 0 || d > time.Minute {
					t.Errorf("audit created_at %v is %v from now", entries[0].CreatedAt, d)
				}
			})
		})
	}
}

func TestGetRecoverableInLocalZones(t *testing.T) {
	for _, loc := range testZones {
		t.Run(loc.String(), func(t *testing.T) {
			withLocal(t, loc)
			runOnEngines(t, func(t *testing.T, db *DB) {
				ctx := context.Background()
				now := time.Now()
				for _, tr := range []domain.TaskRecord{
					{ID: "1h", Status: "downloading", CreatedAt: now.Add(-1 * time.Hour)},
					{ID: "20h", Status: "queued", CreatedAt: now.Add(-20 * time.Hour)},
					{ID: "23h", Status: "uploading", CreatedAt: now.Add(-23 * time.Hour)},
					{ID: "25h", Status: "downloading", CreatedAt: now.Add(-25 * time.Hour)},
					{ID: "30h", Status: "downloading", CreatedAt: now.Add(-30 * time.Hour)},
					{ID: "done", Status: "completed", CreatedAt: now.Add(-1 * time.Hour)},
				} {
					tr.Type, tr.ChatID, tr.UserID = "mirror", 1, 1
					if err := db.Save(ctx, tr); err != nil {
						t.Fatalf("Save %s: %v", tr.ID, err)
					}
				}
				got, err := db.GetRecoverable(ctx)
				if err != nil {
					t.Fatalf("GetRecoverable: %v", err)
				}
				ids := map[string]bool{}
				for _, tr := range got {
					ids[tr.ID] = true
				}
				for _, want := range []string{"1h", "20h", "23h"} {
					if !ids[want] {
						t.Errorf("task %s missing from recoverable %v", want, ids)
					}
				}
				for _, notWant := range []string{"25h", "30h", "done"} {
					if ids[notWant] {
						t.Errorf("task %s should not be recoverable", notWant)
					}
				}
			})
		})
	}
}

func TestScheduledDueInLocalZones(t *testing.T) {
	for _, loc := range testZones {
		t.Run(loc.String(), func(t *testing.T) {
			withLocal(t, loc)
			runOnEngines(t, func(t *testing.T, db *DB) {
				ctx := context.Background()
				now := time.Now()
				// Same format as handlers/basic/schedule.go.
				for id, at := range map[string]time.Time{
					"due":    now.Add(-time.Minute),
					"later":  now.Add(time.Hour),
					"latest": now.Add(5 * time.Hour),
				} {
					if err := db.SaveScheduled(ctx, domain.ScheduledTask{ID: id, TaskType: "mirror", URL: "u", ChatID: 1, UserID: 1,
						Zip: true, ScheduledAt: at.Format(time.RFC3339)}); err != nil {
						t.Fatalf("SaveScheduled: %v", err)
					}
				}
				got, err := db.GetPendingScheduled(ctx)
				if err != nil {
					t.Fatalf("GetPendingScheduled: %v", err)
				}
				if len(got) != 1 || got[0].ID != "due" || !got[0].Zip || got[0].Unzip {
					t.Fatalf("GetPendingScheduled = %+v, want only 'due'", got)
				}
				at, err := time.Parse(time.RFC3339, got[0].ScheduledAt)
				if err != nil || !at.Equal(now.Add(-time.Minute).Truncate(time.Second)) {
					t.Errorf("ScheduledAt = %q (%v)", got[0].ScheduledAt, err)
				}
				if _, err := time.Parse(time.RFC3339, got[0].CreatedAt); err != nil {
					t.Errorf("CreatedAt = %q: %v", got[0].CreatedAt, err)
				}
				if err := db.MarkScheduledDone(ctx, "due", "task1"); err != nil {
					t.Fatalf("MarkScheduledDone: %v", err)
				}
				if got, _ := db.GetPendingScheduled(ctx); len(got) != 0 {
					t.Errorf("after MarkScheduledDone got %+v", got)
				}
			})
		})
	}
}

func TestDailyStatsUseLocalDays(t *testing.T) {
	for _, loc := range testZones {
		t.Run(loc.String(), func(t *testing.T) {
			withLocal(t, loc)
			runOnEngines(t, func(t *testing.T, db *DB) {
				ctx := context.Background()
				y, m, d := time.Now().Date()
				today := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
				for _, tr := range []domain.TaskRecord{
					// Just after local midnight: the previous day in UTC for zones east of UTC.
					{ID: "a", Status: "completed", TotalSize: 100, UserID: 7, CreatedAt: today.Add(time.Minute)},
					{ID: "b", Status: "failed", UserID: 7, CreatedAt: today.Add(time.Minute)},
					// Just before local midnight two days ago.
					{ID: "c", Status: "completed", TotalSize: 5, UserID: 8, CreatedAt: today.AddDate(0, 0, -1).Add(-time.Minute)},
					{ID: "old", Status: "completed", TotalSize: 1, UserID: 8, CreatedAt: today.AddDate(0, 0, -40)},
				} {
					tr.Type, tr.ChatID = "mirror", 1
					if err := db.Save(ctx, tr); err != nil {
						t.Fatalf("Save: %v", err)
					}
				}

				weekly, err := db.GetWeeklyStats(ctx)
				if err != nil || len(weekly) != 7 {
					t.Fatalf("GetWeeklyStats = %d entries, %v", len(weekly), err)
				}
				last, twoAgo := weekly[6], weekly[4]
				if !last.Date.Equal(today) || last.TotalTasks != 2 || last.CompletedTasks != 1 || last.FailedTasks != 1 || last.TotalBandwidth != 100 {
					t.Errorf("today = %+v", last)
				}
				if twoAgo.TotalTasks != 1 || twoAgo.TotalBandwidth != 5 {
					t.Errorf("two days ago = %+v (weekly %+v)", twoAgo, weekly)
				}

				monthly, err := db.GetMonthlyStats(ctx)
				if err != nil || len(monthly) != 30 {
					t.Fatalf("GetMonthlyStats = %d entries, %v", len(monthly), err)
				}
				total := 0
				for _, ds := range monthly {
					total += ds.TotalTasks
				}
				if total != 3 {
					t.Errorf("monthly total = %d, want 3", total)
				}

				todayStats, _ := db.GetTodayStats(ctx)
				if todayStats.TotalTasks != 2 || todayStats.CompletedTasks != 1 || todayStats.FailedTasks != 1 || todayStats.TotalBandwidth != 100 {
					t.Errorf("GetTodayStats = %+v", todayStats)
				}
				userToday, _ := db.GetUserTodayStats(ctx, 8)
				if userToday.TotalTasks != 0 {
					t.Errorf("GetUserTodayStats(8) = %+v, want none today", userToday)
				}

				us, _ := db.GetUserStats(ctx, 8)
				if us.TotalDownloads != 2 || us.SuccessfulTasks != 2 || us.TotalBandwidth != 6 ||
					!sameInstant(us.LastActive, today.AddDate(0, 0, -1).Add(-time.Minute)) {
					t.Errorf("GetUserStats = %+v", us)
				}
				bot, _ := db.GetBotStats(ctx)
				if bot["total_tasks"] != 4 || bot["total_bandwidth"] != int64(106) {
					t.Errorf("GetBotStats = %v", bot)
				}
			})
		})
	}
}

func TestDeleteOld(t *testing.T) {
	runOnEngines(t, func(t *testing.T, db *DB) {
		ctx := context.Background()
		now := time.Now()
		for _, tr := range []domain.TaskRecord{
			{ID: "old-done", Status: "completed", CreatedAt: now.AddDate(0, 0, -40)},
			{ID: "old-active", Status: "downloading", CreatedAt: now.AddDate(0, 0, -40)},
			{ID: "new-done", Status: "failed", CreatedAt: now.AddDate(0, 0, -1)},
		} {
			tr.Type, tr.ChatID, tr.UserID = "mirror", 1, 1
			if err := db.Save(ctx, tr); err != nil {
				t.Fatalf("Save: %v", err)
			}
		}
		// Same format as internal/service/cleanup.go.
		n, err := db.DeleteOld(ctx, now.AddDate(0, 0, -30).Format("2006-01-02 15:04:05"))
		if err != nil || n != 1 {
			t.Fatalf("DeleteOld = %d, %v; want 1", n, err)
		}
		if _, err := db.GetTaskByID(ctx, "old-done"); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("old-done still present: %v", err)
		}
		if _, err := db.DeleteOld(ctx, "garbage"); err == nil {
			t.Error("DeleteOld accepted an unparsable time")
		}
	})
}

func TestListTasksPaging(t *testing.T) {
	runOnEngines(t, func(t *testing.T, db *DB) {
		ctx := context.Background()
		base := time.Now().Add(-time.Hour)
		for i, id := range []string{"a", "b", "c", "d"} {
			if err := db.Save(ctx, domain.TaskRecord{ID: id, Type: "mirror", Status: "completed", ChatID: 1, UserID: 1,
				CreatedAt: base.Add(time.Duration(i) * time.Minute)}); err != nil {
				t.Fatalf("Save: %v", err)
			}
		}
		ids := func(f domain.TaskFilter) string {
			got, err := db.ListTasks(ctx, f)
			if err != nil {
				t.Fatalf("ListTasks(%+v): %v", f, err)
			}
			var s []string
			for _, tr := range got {
				s = append(s, tr.ID)
			}
			return strings.Join(s, ",")
		}
		if got := ids(domain.TaskFilter{Limit: 2, Offset: 1}); got != "c,b" {
			t.Errorf("limit 2 offset 1 = %q", got)
		}
		if got := ids(domain.TaskFilter{Offset: 3}); got != "a" {
			t.Errorf("offset 3 without limit = %q", got)
		}
		if got := ids(domain.TaskFilter{UserID: 1, Status: "completed", Limit: 1}); got != "d" {
			t.Errorf("filtered = %q", got)
		}
	})
}

func TestNullColumnsDoNotDropRows(t *testing.T) {
	runOnEngines(t, func(t *testing.T, db *DB) {
		ctx := context.Background()
		// A task row as written before quality/md5 existed, with NULLs.
		if _, err := db.ExecContext(ctx, `INSERT INTO tasks (id, type, status, chat_id, user_id, created_at, quality, md5)
			VALUES ('legacy', 'mirror', 'queued', 1, 1, $1, NULL, NULL)`, db.timeArg(time.Now())); err != nil {
			t.Fatalf("insert: %v", err)
		}
		if tr, err := db.GetTaskByID(ctx, "legacy"); err != nil || tr.Quality != "" {
			t.Errorf("GetTaskByID(legacy) = %+v, %v", tr, err)
		}
		if got, _ := db.GetActive(ctx); len(got) != 1 {
			t.Errorf("GetActive dropped the legacy row: %d", len(got))
		}
		if got, _ := db.GetRecoverable(ctx); len(got) != 1 {
			t.Errorf("GetRecoverable dropped the legacy row: %d", len(got))
		}

		// SetLanguage creates a user without a username.
		if err := db.SetLanguage(ctx, 42, "en"); err != nil {
			t.Fatalf("SetLanguage: %v", err)
		}
		if u, err := db.GetByID(ctx, 42); err != nil || u.Language != "en" || u.Username != "" {
			t.Errorf("GetByID(42) = %+v, %v", u, err)
		}
		if users, err := db.GetAll(ctx); err != nil || len(users) != 1 {
			t.Errorf("GetAll = %d users, %v", len(users), err)
		}
	})
}

func TestCheckpointRoundTrip(t *testing.T) {
	runOnEngines(t, func(t *testing.T, db *DB) {
		ctx := context.Background()
		if err := db.Save(ctx, domain.TaskRecord{ID: "cp", Type: "mirror", Status: "downloading", ChatID: 1, UserID: 1, CreatedAt: time.Now()}); err != nil {
			t.Fatalf("Save: %v", err)
		}
		at := time.Unix(1_790_000_123, 0)
		for _, p := range []float64{12.5, 33.333333} {
			if err := db.SaveCheckpoint(ctx, domain.TaskCheckpoint{TaskID: "cp", DownloadedBytes: 5_000_000_000, TotalBytes: 6_000_000_000, Progress: p, LastUpdate: at}); err != nil {
				t.Fatalf("SaveCheckpoint: %v", err)
			}
		}
		cp, err := db.GetCheckpoint(ctx, "cp")
		if err != nil || cp == nil {
			t.Fatalf("GetCheckpoint = %v, %v", cp, err)
		}
		if cp.DownloadedBytes != 5_000_000_000 || cp.Progress != 33.333333 || !cp.LastUpdate.Equal(at) {
			t.Errorf("checkpoint = %+v", cp)
		}
		// rate_limits.last_fetch holds Unix seconds with a fraction (internal/queue).
		lastFetch := 1_790_000_123.25
		if _, err := db.ExecContext(ctx, "INSERT INTO rate_limits (user_id, tokens, last_fetch, updated_at) VALUES ($1, $2, $3, $4)",
			int64(1), 3.5, lastFetch, db.timeArg(time.Now())); err != nil {
			t.Fatalf("insert rate_limits: %v", err)
		}
		var got float64
		if err := db.QueryRowContext(ctx, "SELECT last_fetch FROM rate_limits WHERE user_id = $1", int64(1)).Scan(&got); err != nil || got != lastFetch {
			t.Errorf("rate_limits.last_fetch = %v (%v), want %v", got, err, lastFetch)
		}
	})
}

// TestLegacySQLiteTimestamps builds a database the way earlier releases
// did (time.Time.String() values in the local zone, RFC 3339 scheduled_at)
// and checks that migration 000012 makes them usable.
func TestLegacySQLiteTimestamps(t *testing.T) {
	withLocal(t, time.FixedZone("WIB", 7*3600))
	dir := t.TempDir()
	path := filepath.Join(dir, "zee-mirror.db")

	legacy, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatal(err)
	}
	drv, err := sqlite.WithInstance(legacy, &sqlite.Config{})
	if err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs(testMigrationsDir)
	m, err := migrate.NewWithDatabaseInstance("file://"+filepath.ToSlash(abs), "sqlite", drv)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.Migrate(11); err != nil {
		t.Fatalf("migrate to 11: %v", err)
	}

	now := time.Now() // carries a monotonic reading, so String() ends in "m=+..."
	ctx := context.Background()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, execErr := legacy.ExecContext(ctx, q, args...); execErr != nil {
			t.Fatalf("%s: %v", q, execErr)
		}
	}
	exec("INSERT INTO users (id, username, role, created_at, expires_at) VALUES (1, 'u', 'user', ?, ?)", now.Add(-time.Hour), now.Add(-30*time.Minute))
	exec("INSERT INTO tasks (id, type, status, chat_id, user_id, created_at) VALUES ('recent', 'mirror', 'downloading', 1, 1, ?)", now.Add(-time.Hour))
	exec("INSERT INTO tasks (id, type, status, chat_id, user_id, created_at) VALUES ('stale', 'mirror', 'downloading', 1, 1, ?)", now.Add(-26*time.Hour))
	exec("INSERT INTO tasks (id, type, status, chat_id, user_id, created_at) VALUES ('zero', 'mirror', 'completed', 1, 1, ?)", time.Time{})
	exec("INSERT INTO audit_logs (action, created_at) VALUES ('a', CURRENT_TIMESTAMP)")
	exec("INSERT INTO scheduled_tasks (id, task_type, url, chat_id, user_id, scheduled_at, created_at) VALUES ('due', 'mirror', 'u', 1, 1, ?, datetime('now'))",
		now.Add(-time.Minute).Format(time.RFC3339))
	exec("INSERT INTO scheduled_tasks (id, task_type, url, chat_id, user_id, scheduled_at, created_at) VALUES ('later', 'mirror', 'u', 1, 1, ?, datetime('now'))",
		now.Add(time.Hour).Format(time.RFC3339))
	var raw string
	_ = legacy.QueryRow("SELECT CAST(created_at AS TEXT) FROM tasks WHERE id = 'recent'").Scan(&raw)
	if !strings.Contains(raw, " +0700 WIB") {
		t.Fatalf("legacy value %q is not in time.Time.String() form", raw)
	}
	_ = legacy.Close()

	db, err := NewDB("sqlite", dir, "", testMigrationsDir)
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer func() { _ = db.Close() }()

	_ = db.QueryRow("SELECT CAST(created_at AS TEXT) FROM tasks WHERE id = 'recent'").Scan(&raw)
	if !strings.HasSuffix(raw, "+00:00") {
		t.Errorf("created_at not normalised: %q", raw)
	}
	tr, err := db.GetTaskByID(ctx, "recent")
	if err != nil || !sameInstant(tr.CreatedAt, now.Add(-time.Hour)) {
		t.Errorf("recent created_at = %v (%v), want %v", tr.CreatedAt, err, now.Add(-time.Hour))
	}
	if tr, zerr := db.GetTaskByID(ctx, "zero"); zerr != nil || !tr.CreatedAt.IsZero() {
		t.Errorf("zero created_at = %v (%v)", tr.CreatedAt, zerr)
	}
	rec, err := db.GetRecoverable(ctx)
	if err != nil || len(rec) != 1 || rec[0].ID != "recent" {
		t.Errorf("GetRecoverable = %+v, %v; want only 'recent'", rec, err)
	}
	u, err := db.GetByID(ctx, 1)
	if err != nil || u.IsActive || !sameInstant(u.ExpiresAt.Time, now.Add(-30*time.Minute)) {
		t.Errorf("user = %+v, %v", u, err)
	}
	due, err := db.GetPendingScheduled(ctx)
	if err != nil || len(due) != 1 || due[0].ID != "due" {
		t.Errorf("GetPendingScheduled = %+v, %v; want only 'due'", due, err)
	}
	if logs, err := db.ListAuditLogs(ctx, 10, 0); err != nil || len(logs) != 1 || time.Since(logs[0].CreatedAt) > time.Minute {
		t.Errorf("audit logs = %+v, %v", logs, err)
	}
	if weekly, _ := db.GetWeeklyStats(ctx); weekly[6].TotalTasks != 1 {
		t.Errorf("weekly today = %+v", weekly[6])
	}
}

// TestDirtyVersionOneRetries covers a first migration that failed part way:
// the database is marked dirty at version 1 and has no tables.
func TestDirtyVersionOneRetries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "zee-mirror.db")
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = raw.Exec("CREATE TABLE schema_migrations (version uint64, dirty bool); INSERT INTO schema_migrations VALUES (1, 1);"); err != nil {
		t.Fatal(err)
	}
	_ = raw.Close()

	db, err := NewDB("sqlite", dir, "", testMigrationsDir)
	if err != nil {
		t.Fatalf("NewDB on dirty database: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.GetCount(context.Background()); err != nil {
		t.Errorf("users table missing after retry: %v", err)
	}
}
