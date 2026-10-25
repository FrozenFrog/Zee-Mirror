package config

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"zee-mirror/internal/crypto"
	"zee-mirror/pkg/utils"

	"github.com/joho/godotenv"
)

type Config struct {
	RclonePacerBurst         string
	RcloneTransfers          string
	RcloneLogLevel           string
	RcloneDest               string
	DownloadDir              string
	ConfigDir                string
	DashboardToken           string
	LogLevel                 string
	LogFormat                string
	AppEnv                   string
	SentryDSN                string
	BotToken                 string
	DatabaseURL              string
	SecretsDir               string
	SecretsSeedDir           string
	RcloneBufferSize         string
	AppHash                  string
	RclonePacerMinSleep      string
	Aria2RPCURL              string
	Aria2RPCSecret           string
	WebhookURL               string
	VikingUserHash           string
	UserSessionString        string
	WebhookSecret            string
	IndexURL                 string
	TelegramAPI              string
	DashboardURL             string
	RcloneDriveChunkSize     string
	RcloneCheckers           string
	DBDriver                 string
	TimeZone                 string
	EncryptionKey            []byte
	RcloneDestFallbacks      []string
	BotTokens                []string
	AuthorizedUsers          []int64
	MaxRetries               int
	AutoCleanupDays          int
	DefaultMaxDailyTasks     int
	TGGetFileTimeout         int
	AppID                    int
	MaxConcurrentDownloads   int
	OwnerID                  int64
	DefaultMaxDailyBandwidth int64
	DashboardPort            int
	SmartAutoOrganization    bool
	StopDuplicate            bool
	UseWebhook               bool
	LocalBotAPI              bool
}

func LoadConfig() *Config {
	src := newSource()
	cfg := &Config{
		BotToken:                 src.str("BOT_TOKEN"),
		TelegramAPI:              src.str("TELEGRAM_API"),
		RcloneDest:               src.str("RCLONE_DEST"),
		DownloadDir:              src.strOr("DOWNLOAD_DIR", "/app/downloads"),
		ConfigDir:                src.strOr("CONFIG_DIR", "/app/config"),
		MaxConcurrentDownloads:   src.int("MAX_CONCURRENT_DOWNLOADS", 3),
		DashboardToken:           src.strOr("WEB_DASHBOARD_TOKEN", "zee-mirror-secret"),
		DashboardPort:            src.int("DASHBOARD_PORT_INTERNAL", 8080),
		LogLevel:                 src.strOr("LOG_LEVEL", "info"),
		LogFormat:                src.strOr("LOG_FORMAT", "text"),
		AppEnv:                   src.strOr("APP_ENV", ""),
		SentryDSN:                src.str("SENTRY_DSN"),
		SmartAutoOrganization:    src.bool("SMART_AUTO_ORGANIZATION", false),
		IndexURL:                 src.str("INDEX_URL"),
		DashboardURL:             src.strOr("WEB_DASHBOARD_URL", "127.0.0.1"),
		DefaultMaxDailyTasks:     src.int("DEFAULT_MAX_DAILY_TASKS", -1),
		DefaultMaxDailyBandwidth: utils.ParseBytesString(src.strOr("DEFAULT_MAX_DAILY_BANDWIDTH", "-1")),
		StopDuplicate:            src.bool("STOP_DUPLICATE", false),
		MaxRetries:               src.int("MAX_RETRIES", 3),
		VikingUserHash:           src.str("VIKING_USER_HASH"),
		AppID:                    src.int("APP_ID", 0),
		AppHash:                  src.str("APP_HASH"),
		UserSessionString:        src.str("USER_SESSION_STRING"),
		Aria2RPCURL:              src.strOr("ARIA2_RPC_URL", "http://localhost:6800/jsonrpc"),
		Aria2RPCSecret:           src.str("ARIA2_RPC_SECRET"),
		WebhookURL:               src.str("WEBHOOK_URL"),
		WebhookSecret:            src.strOr("WEBHOOK_SECRET", ""),
		UseWebhook:               src.bool("USE_WEBHOOK", false),
		RcloneTransfers:          src.strOr("RCLONE_TRANSFERS", "10"),
		RcloneCheckers:           src.strOr("RCLONE_CHECKERS", "20"),
		RcloneDriveChunkSize:     src.strOr("RCLONE_DRIVE_CHUNK_SIZE", "256M"),
		RcloneBufferSize:         src.strOr("RCLONE_BUFFER_SIZE", "128M"),
		RcloneDestFallbacks:      splitCSV(src.str("RCLONE_DEST_FALLBACKS")),
		RclonePacerMinSleep:      src.strOr("RCLONE_PACER_MIN_SLEEP", "10ms"),
		RclonePacerBurst:         src.strOr("RCLONE_PACER_BURST", "200"),
		RcloneLogLevel:           src.strOr("RCLONE_LOG_LEVEL", "NOTICE"),
		AutoCleanupDays:          src.int("AUTO_CLEANUP_DAYS", 30),
		DBDriver:                 src.strOr("DB_DRIVER", "postgres"),
		DatabaseURL:              src.str("DATABASE_URL"),
		SecretsDir:               src.strOr("SECRETS_DIR", "/dev/shm"),
		SecretsSeedDir:           src.str("SECRETS_SEED_DIR"),
		TGGetFileTimeout:         src.int("TG_GETFILE_TIMEOUT", 600),
		TimeZone:                 src.str("TZ"),
		LocalBotAPI:              src.bool("LOCAL_BOT_API", true),
		EncryptionKey:            src.key,
	}

	if ownerIDStr := src.str("OWNER_ID"); ownerIDStr != "" {
		if id, err := strconv.ParseInt(ownerIDStr, 10, 64); err == nil {
			cfg.OwnerID = id
		}
	}

	if authUsers := src.str("AUTHORIZED_USERS"); authUsers != "" {
		for _, idStr := range strings.Split(authUsers, ",") {
			idStr = strings.TrimSpace(idStr)
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				cfg.AuthorizedUsers = append(cfg.AuthorizedUsers, id)
			}
		}
	}

	if botTokens := src.str("BOT_TOKENS"); botTokens != "" {
		for _, t := range strings.Split(botTokens, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				cfg.BotTokens = append(cfg.BotTokens, src.decrypt("BOT_TOKENS", t))
			}
		}
	} else if cfg.BotToken != "" {
		cfg.BotTokens = []string{cfg.BotToken}
	}

	return cfg
}

var currentConfig atomic.Value

// SettingsSource returns config values stored in the database, keyed by the
// same names as the environment variables (BOT_TOKEN, OWNER_ID, ...).
type SettingsSource func(ctx context.Context) (map[string]string, error)

var (
	dbSettings     atomic.Pointer[map[string]string]
	settingsSource atomic.Pointer[SettingsSource]
	settingKeyRe   = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
)

func init() {
	_ = godotenv.Load()
	currentConfig.Store(LoadConfig())
}

// UseSettings layers database-stored values under the environment: a
// non-empty environment variable still wins. The source is kept so Reload
// picks up later database changes.
func UseSettings(ctx context.Context, src SettingsSource) (*Config, error) {
	if err := loadSettings(ctx, src); err != nil {
		return nil, err
	}
	settingsSource.Store(&src)
	cfg := LoadConfig()
	currentConfig.Store(cfg)
	return cfg, nil
}

func loadSettings(ctx context.Context, src SettingsSource) error {
	all, err := src(ctx)
	if err != nil {
		return fmt.Errorf("load settings from database: %w", err)
	}
	vals := make(map[string]string, len(all))
	for k, v := range all {
		if settingKeyRe.MatchString(k) {
			vals[k] = v
		}
	}
	dbSettings.Store(&vals)
	return nil
}

func Reload() *Config {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		slog.Warn("Failed to load .env file", "error", err)
	}
	if src := settingsSource.Load(); src != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := loadSettings(ctx, *src)
		cancel()
		if err != nil {
			slog.Warn("Failed to reload settings from database, using previous values", "error", err)
		}
	}
	newCfg := LoadConfig()
	if newCfg.BotToken == "" && len(newCfg.BotTokens) == 0 {
		slog.Error("Reloaded config has empty BOT_TOKEN, keeping previous config")
		return currentConfig.Load().(*Config)
	}
	currentConfig.Store(newCfg)
	slog.Info("Config reloaded successfully",
		"maxConcurrent", newCfg.MaxConcurrentDownloads,
		"stopDuplicate", newCfg.StopDuplicate,
		"logLevel", newCfg.LogLevel,
	)
	return newCfg
}

// source resolves a config key: non-empty environment variable first, then
// the database settings, then the fallback. "enc:" values are decrypted.
type source struct {
	settings map[string]string
	key      []byte
}

func newSource() source {
	var src source
	if m := dbSettings.Load(); m != nil {
		src.settings = *m
	}
	// The key itself only comes from the environment: storing it next to
	// the data it protects would defeat the purpose.
	if keyHex := os.Getenv("ENCRYPTION_KEY"); keyHex != "" {
		key, err := hex.DecodeString(keyHex)
		if err == nil && len(key) == 32 {
			src.key = key
		} else {
			slog.Warn("Invalid ENCRYPTION_KEY: must be 32-byte hex", "length", len(key))
		}
	}
	return src
}

func (s source) lookup(name string) (string, bool) {
	if v := os.Getenv(name); v != "" {
		return s.decrypt(name, v), true
	}
	if v, ok := s.settings[name]; ok && v != "" {
		return s.decrypt(name, v), true
	}
	return "", false
}

func (s source) str(name string) string {
	v, _ := s.lookup(name)
	return v
}

func (s source) strOr(name, fallback string) string {
	if v, ok := s.lookup(name); ok {
		return v
	}
	return fallback
}

func (s source) int(name string, fallback int) int {
	if n, err := strconv.Atoi(s.str(name)); err == nil {
		return n
	}
	return fallback
}

func (s source) bool(name string, fallback bool) bool {
	b, err := strconv.ParseBool(s.str(name))
	if err != nil {
		return fallback
	}
	return b
}

func (s source) decrypt(name, val string) string {
	if !strings.HasPrefix(val, "enc:") {
		return val
	}
	if s.key == nil {
		slog.Error("Config value is encrypted but ENCRYPTION_KEY is not set", "key", name)
		return ""
	}
	decoded, err := crypto.Decrypt(strings.TrimPrefix(val, "enc:"), s.key)
	if err != nil {
		slog.Error("Failed to decrypt config value", "key", name, "error", err)
		return ""
	}
	return string(decoded)
}

func splitCSV(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (c *Config) Validate() error {
	if len(c.BotTokens) == 0 {
		return fmt.Errorf("BOT_TOKEN or BOT_TOKENS must be set")
	}
	if c.OwnerID == 0 {
		return fmt.Errorf("OWNER_ID must be set")
	}
	if strings.EqualFold(c.AppEnv, "production") &&
		(c.DashboardToken == "" || c.DashboardToken == "zee-mirror-secret") {
		return fmt.Errorf("WEB_DASHBOARD_TOKEN must be set to a secure value when APP_ENV=production (empty or default token is not allowed)")
	}
	if c.UseWebhook && c.WebhookSecret == "" {
		return fmt.Errorf("WEBHOOK_SECRET must be set when USE_WEBHOOK=true (webhook endpoint is unauthenticated without it)")
	}
	if c.RcloneDest == "" {
		slog.Warn("RCLONE_DEST not set, using default: gdrive:/MirrorBot")
		c.RcloneDest = "gdrive:/MirrorBot"
	}
	return nil
}
