package database

import (
	"context"
	"time"
)

func (db *DB) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := db.QueryRowContext(ctx, "SELECT COALESCE(value, '') FROM settings WHERE key = $1", key).Scan(&value)
	return value, err
}

func (db *DB) Set(ctx context.Context, key, value string) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, $3)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, key, value, db.timeArg(time.Now()))
	return err
}

// AllSettings returns every row of the settings table.
func (db *DB) AllSettings(ctx context.Context) (map[string]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT key, COALESCE(value, '') FROM settings")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		out[key] = value
	}
	return out, rows.Err()
}

// DeleteSetting removes a row from the settings table.
func (db *DB) DeleteSetting(ctx context.Context, key string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM settings WHERE key = $1", key)
	return err
}
