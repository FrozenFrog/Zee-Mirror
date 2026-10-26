package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"zee-mirror/internal/repository"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx stdlib registration
)

const (
	driverSQLite   = "sqlite"
	driverPostgres = "postgres"
)

type DB struct {
	*sql.DB
	driver string
}

var _ repository.TaskRepository = (*DB)(nil)
var _ repository.UserRepository = (*DB)(nil)
var _ repository.SettingsRepository = (*DB)(nil)
var _ repository.ScheduledTaskRepository = (*DB)(nil)
var _ repository.AuditRepository = (*DB)(nil)

func NewDB(driverName, configDir, dsn, migrationsDir string) (*DB, error) {
	switch driverName {
	case driverPostgres:
		return newPostgresDB(dsn, migrationsDir)
	case driverSQLite:
		return newSQLiteDB(configDir, dsn, migrationsDir)
	default:
		return nil, fmt.Errorf("unsupported DB_DRIVER %q (use sqlite or postgres)", driverName)
	}
}

// sqliteDSN appends the connection options to the database path.
// _time_format=sqlite makes the driver write time.Time values as
// "2006-01-02 15:04:05.999999999-07:00" instead of time.Time.String(), so
// stored values can be parsed by SQLite's date functions.
func sqliteDSN(dbPath string) string {
	sep := "?"
	if strings.Contains(dbPath, "?") {
		sep = "&"
	}
	return dbPath + sep + "_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(1)&_time_format=sqlite"
}

func newSQLiteDB(configDir, dsn, migrationsDir string) (*DB, error) {
	dbPath := dsn
	if dbPath == "" {
		dbPath = filepath.Join(configDir, "zee-mirror.db")
	}

	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0750); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", sqliteDSN(dbPath))
	if err != nil {
		return nil, err
	}

	// One connection serialises every statement. Never run a query on db
	// while rows from another query are still open, or it waits forever.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	instance := &DB{DB: db, driver: driverSQLite}
	if err := instance.RunMigrations(migrationsDir); err != nil {
		slog.Error("Database migration failed", "error", err)
		_ = db.Close()
		return nil, err
	}

	return instance, nil
}

func newPostgresDB(dsn, migrationsDir string) (*DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is required for postgres driver")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	instance := &DB{DB: db, driver: driverPostgres}
	if err := instance.RunMigrations(migrationsDir); err != nil {
		slog.Error("Database migration failed", "error", err)
		_ = db.Close()
		return nil, err
	}

	return instance, nil
}

// Time columns hold different things per engine, so every time.Time that is
// written goes through timeArg and every one that is read through fromDB.
//
//   - SQLite has no time type. Times are stored as UTC text, which compares
//     correctly as text and matches datetime('now') and CURRENT_TIMESTAMP.
//   - Postgres columns are TIMESTAMP (without time zone). pgx drops the zone
//     and stores the wall clock, so existing rows hold the app's local time
//     (TZ). New rows keep that convention, and reads put the zone back.
//
// Comparisons against "now" pass time.Now() through timeArg instead of using
// NOW() or datetime('now'), whose zone can differ from the stored values.
func (db *DB) timeArg(t time.Time) time.Time {
	if db.driver == driverPostgres {
		return t.In(time.Local)
	}
	return t.UTC()
}

// nullTimeArg is timeArg for nullable columns.
func (db *DB) nullTimeArg(t sql.NullTime) any {
	if !t.Valid {
		return nil
	}
	return db.timeArg(t.Time)
}

// fromDB returns a time read from a DATETIME/TIMESTAMP column in time.Local.
func (db *DB) fromDB(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	if db.driver == driverPostgres {
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	}
	return t.In(time.Local)
}

func (db *DB) fromDBNull(t sql.NullTime) sql.NullTime {
	if t.Valid {
		t.Time = db.fromDB(t.Time)
	}
	return t
}

func (db *DB) RunMigrations(migrationsDir string) error {
	var driver database.Driver
	var err error

	migrateDir := migrationsDir
	switch db.driver {
	case driverPostgres:
		migrateDir = filepath.Join(migrationsDir, "postgres")
		driver, err = postgres.WithInstance(db.DB, &postgres.Config{})
		if err != nil {
			return fmt.Errorf("failed to create postgres migrate driver: %w", err)
		}
	default:
		driver, err = sqlite.WithInstance(db.DB, &sqlite.Config{})
		if err != nil {
			return fmt.Errorf("failed to create sqlite migrate driver: %w", err)
		}
	}

	if absMigrateDir, absErr := filepath.Abs(migrateDir); absErr == nil {
		migrateDir = filepath.ToSlash(absMigrateDir)
	}
	if _, statErr := os.Stat(migrateDir); statErr != nil {
		return fmt.Errorf("%s migrations directory not found: %w", db.driver, statErr)
	}

	// m is not closed: closing it would also close db.DB.
	m, err := migrate.NewWithDatabaseInstance("file://"+migrateDir, db.driver, driver)
	if err != nil {
		return err
	}

	version, dirty, verr := m.Version()
	switch {
	case errors.Is(verr, migrate.ErrNilVersion):
		slog.Info("Database is empty, creating schema", "driver", db.driver)
	case verr != nil:
		return fmt.Errorf("failed to read migration version: %w", verr)
	default:
		slog.Info("Database schema found, applying pending migrations", "driver", db.driver, "version", version, "dirty", dirty)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		var dirtyErr migrate.ErrDirty
		if errors.As(err, &dirtyErr) && dirtyErr.Version == 1 {
			// Each migration runs in one transaction on both engines, so a
			// failed 000001 left nothing behind. It only uses IF NOT EXISTS,
			// so reset to "no version" and run it again rather than marking
			// it applied.
			slog.Warn("Database is dirty at version 1, resetting and retrying the migrations")
			if errForce := m.Force(database.NilVersion); errForce != nil {
				return fmt.Errorf("failed to force migration: %w", errForce)
			}
			if errRetry := m.Up(); errRetry != nil && !errors.Is(errRetry, migrate.ErrNoChange) {
				return fmt.Errorf("migration failed after force: %w", errRetry)
			}
			slog.Info("Database migration fixed and completed successfully")
			return nil
		}
		return err
	}

	slog.Info("Database migrations completed successfully")
	return nil
}

func (db *DB) Ping(ctx context.Context) error {
	return db.PingContext(ctx)
}
