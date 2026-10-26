package config

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"zee-mirror/internal/crypto"

	"github.com/joho/godotenv"
)

// Bootstrap keys are needed before the database is open, so they only come
// from the environment (or a .env file) and are never stored in it.
var bootstrapKeys = []string{"DB_DRIVER", "DATABASE_URL", "ENCRYPTION_KEY", "CONFIG_DIR"}

type settingKey struct {
	name   string
	secret bool
}

// settingKeys lists every key LoadConfig reads besides the bootstrap keys.
// Each may be set in the environment, a .env file or the settings table.
// Secret values are stored encrypted.
var settingKeys = []settingKey{
	{name: "BOT_TOKEN", secret: true},
	{name: "BOT_TOKENS", secret: true},
	{name: "OWNER_ID"},
	{name: "AUTHORIZED_USERS"},
	{name: "AUTHORIZED_CHATS"},
	{name: "TELEGRAM_API"},
	{name: "LOCAL_BOT_API"},
	{name: "TG_GETFILE_TIMEOUT"},
	{name: "APP_ID"},
	{name: "APP_HASH", secret: true},
	{name: "USER_SESSION_STRING", secret: true},
	{name: "WEB_DASHBOARD_URL"},
	{name: "WEB_DASHBOARD_TOKEN", secret: true},
	{name: "DASHBOARD_PORT_INTERNAL"},
	{name: "OIDC_ISSUER"},
	{name: "OIDC_CLIENT_ID"},
	{name: "OIDC_CLIENT_SECRET", secret: true},
	{name: "OIDC_REDIRECT_URL"},
	{name: "OIDC_PROVIDER_NAME"},
	{name: "OIDC_ALLOWED_GROUPS"},
	{name: "USE_WEBHOOK"},
	{name: "WEBHOOK_URL"},
	{name: "WEBHOOK_SECRET", secret: true},
	{name: "RCLONE_DEST"},
	{name: "RCLONE_DEST_FALLBACKS"},
	{name: "RCLONE_TRANSFERS"},
	{name: "RCLONE_CHECKERS"},
	{name: "RCLONE_DRIVE_CHUNK_SIZE"},
	{name: "RCLONE_BUFFER_SIZE"},
	{name: "RCLONE_PACER_MIN_SLEEP"},
	{name: "RCLONE_PACER_BURST"},
	{name: "RCLONE_LOG_LEVEL"},
	{name: "ARIA2_RPC_URL"},
	{name: "ARIA2_RPC_SECRET", secret: true},
	{name: "DOWNLOAD_DIR"},
	{name: "SECRETS_DIR"},
	{name: "SECRETS_SEED_DIR"},
	{name: "MAX_CONCURRENT_DOWNLOADS"},
	{name: "MAX_RETRIES"},
	{name: "STOP_DUPLICATE"},
	{name: "SMART_AUTO_ORGANIZATION"},
	{name: "AUTO_CLEANUP_DAYS"},
	{name: "DEFAULT_MAX_DAILY_TASKS"},
	{name: "DEFAULT_MAX_DAILY_BANDWIDTH"},
	{name: "INDEX_URL"},
	{name: "VIKING_USER_HASH", secret: true},
	{name: "LOG_LEVEL"},
	{name: "LOG_FORMAT"},
	{name: "APP_ENV"},
	{name: "SENTRY_DSN", secret: true},
	{name: "TZ"},
}

func isBootstrapKey(name string) bool {
	return slices.Contains(bootstrapKeys, name)
}

func findSettingKey(name string) (settingKey, bool) {
	i := slices.IndexFunc(settingKeys, func(k settingKey) bool { return k.name == name })
	if i < 0 {
		return settingKey{}, false
	}
	return settingKeys[i], true
}

// IsSecret reports whether a setting is stored encrypted and hidden from the
// dashboard.
func IsSecret(name string) bool {
	k, ok := findSettingKey(name)
	return ok && k.secret
}

// SettingsStore is the settings table.
type SettingsStore interface {
	AllSettings(ctx context.Context) (map[string]string, error)
	Set(ctx context.Context, key, value string) error
	DeleteSetting(ctx context.Context, key string) error
}

var errNoKey = errors.New("ENCRYPTION_KEY is not set")

// encode prepares a value for the settings table: secrets are encrypted
// unless they already are.
func (s source) encode(name, val string) (string, error) {
	if !IsSecret(name) || strings.HasPrefix(val, "enc:") {
		return val, nil
	}
	if s.key == nil {
		return "", errNoKey
	}
	enc, err := crypto.Encrypt([]byte(val), s.key)
	if err != nil {
		return "", err
	}
	return "enc:" + enc, nil
}

// SyncEnv writes every setting declared in the environment or a .env file
// into the settings table, so the database keeps the last known value when
// the variable is later removed. Unchanged values are left alone. It returns
// the keys it wrote.
func SyncEnv(ctx context.Context, store SettingsStore) ([]string, error) {
	stored, err := store.AllSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}
	src := newSource()
	var written []string
	for _, k := range settingKeys {
		val := os.Getenv(k.name)
		if val == "" {
			continue
		}
		plain := src.decrypt(k.name, val)
		if old, ok := stored[k.name]; ok && src.decrypt(k.name, old) == plain {
			continue
		}
		enc, err := src.encode(k.name, plain)
		if err != nil {
			return written, fmt.Errorf("store %s: %w", k.name, err)
		}
		if err = store.Set(ctx, k.name, enc); err != nil {
			return written, fmt.Errorf("store %s: %w", k.name, err)
		}
		written = append(written, k.name)
	}
	return written, nil
}

// redacted stands in for secret values shown on the dashboard. Sending it
// back keeps the stored value.
const redacted = "***"

// FormatSettings renders the settings as KEY=value lines for the dashboard
// editor. Secrets are redacted. Keys set in the environment are commented
// out, since the environment overrides whatever is saved here.
func FormatSettings() string {
	src := newSource()
	var b strings.Builder
	b.WriteString("# Settings stored in the database (KEY=value, one per line).\n")
	b.WriteString("# " + redacted + " keeps a hidden secret; KEY= with no value removes it.\n")
	b.WriteString("# Lines marked (env) come from the environment or .env and override the database.\n\n")
	for _, k := range settingKeys {
		env := os.Getenv(k.name)
		val := env
		if val == "" {
			val = src.settings[k.name]
		}
		if val != "" && k.secret {
			val = redacted
		}
		switch {
		case env != "":
			fmt.Fprintf(&b, "# (env) %s=%s\n", k.name, val)
		case val == "":
			fmt.Fprintf(&b, "# %s=\n", k.name)
		default:
			fmt.Fprintf(&b, "%s=%s\n", k.name, val)
		}
	}
	return b.String()
}

// SaveSettings applies KEY=value lines from the dashboard editor to the
// settings table. Comments and blank lines are ignored, a redacted secret is
// kept, and an empty value deletes the row. It returns the keys it changed
// and those skipped because the environment overrides them.
func SaveSettings(ctx context.Context, store SettingsStore, text string) (changed, overridden []string, err error) {
	values := make(map[string]string)
	var order []string
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, val, ok := strings.Cut(line, "=")
		name = strings.TrimSpace(name)
		if !ok {
			return nil, nil, fmt.Errorf("line %d: expected KEY=value", i+1)
		}
		if isBootstrapKey(name) {
			return nil, nil, fmt.Errorf("line %d: %s can only be set in the environment", i+1, name)
		}
		if _, known := findSettingKey(name); !known {
			return nil, nil, fmt.Errorf("line %d: unknown setting %q", i+1, name)
		}
		if _, dup := values[name]; !dup {
			order = append(order, name)
		}
		values[name] = unquote(strings.TrimSpace(val))
	}

	stored, err := store.AllSettings(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("read settings: %w", err)
	}
	src := newSource()
	for _, name := range order {
		val := values[name]
		if os.Getenv(name) != "" {
			overridden = append(overridden, name)
			continue
		}
		old, exists := stored[name]
		switch {
		case val == redacted && IsSecret(name):
			continue
		case val == "":
			if !exists {
				continue
			}
			err = store.DeleteSetting(ctx, name)
		case exists && src.decrypt(name, old) == val:
			continue
		default:
			var enc string
			if enc, err = src.encode(name, val); err == nil {
				err = store.Set(ctx, name, enc)
			}
		}
		if err != nil {
			return changed, overridden, fmt.Errorf("save %s: %w", name, err)
		}
		changed = append(changed, name)
	}
	return changed, overridden, nil
}

func unquote(v string) string {
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
		return v[1 : len(v)-1]
	}
	return v
}

const defaultConfigDir = "/app/config"

// loadDotEnv reads ./.env, then CONFIG_DIR/.env. Neither overrides a
// variable that is already set.
func loadDotEnv() {
	dir := os.Getenv("CONFIG_DIR")
	if dir == "" {
		dir = defaultConfigDir
	}
	for _, f := range []string{".env", filepath.Join(dir, ".env")} {
		err := godotenv.Load(f)
		switch {
		case err == nil:
			slog.Debug("Loaded settings file", "file", f)
		case !os.IsNotExist(err):
			slog.Warn("Failed to load settings file", "file", f, "error", err)
		}
	}
}
