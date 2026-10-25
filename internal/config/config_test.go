package config

import (
	"context"
	"encoding/hex"
	"slices"
	"strings"
	"testing"

	"zee-mirror/internal/crypto"
)

func TestValidate_MissingBotTokens(t *testing.T) {
	cfg := &Config{
		OwnerID:    12345,
		RcloneDest: "gdrive:/test",
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing BotTokens")
	}
	if err.Error() != "BOT_TOKEN or BOT_TOKENS must be set" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidate_MissingOwnerID(t *testing.T) {
	cfg := &Config{
		BotTokens:  []string{"token1"},
		RcloneDest: "gdrive:/test",
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing OwnerID")
	}
	if err.Error() != "OWNER_ID must be set" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidate_EmptyRcloneDest_Defaults(t *testing.T) {
	cfg := &Config{
		BotTokens: []string{"token1"},
		OwnerID:   12345,
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cfg.RcloneDest != "gdrive:/MirrorBot" {
		t.Errorf("expected default rclone dest, got %s", cfg.RcloneDest)
	}
}

func TestValidate_AllFieldsPresent(t *testing.T) {
	cfg := &Config{
		BotTokens:  []string{"token1"},
		OwnerID:    12345,
		RcloneDest: "gdrive:/MyFolder",
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cfg.RcloneDest != "gdrive:/MyFolder" {
		t.Errorf("RcloneDest should not be overridden, got %s", cfg.RcloneDest)
	}
}

func TestValidate_SingleBotToken(t *testing.T) {
	cfg := &Config{
		BotTokens: []string{"single-token"},
		OwnerID:   111,
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("single token should be valid: %v", err)
	}
}

func TestValidate_MultipleBotTokens(t *testing.T) {
	cfg := &Config{
		BotTokens: []string{"token1", "token2", "token3"},
		OwnerID:   111,
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("multiple tokens should be valid: %v", err)
	}
}

func TestValidate_OIDC(t *testing.T) {
	base := Config{BotTokens: []string{"token1"}, OwnerID: 12345}
	tests := []struct {
		name    string
		mutate  func(c *Config)
		wantErr bool
	}{
		{"disabled", func(*Config) {}, false},
		{"complete", func(c *Config) {
			c.OIDCIssuer, c.OIDCClientID, c.OIDCRedirectURL = "https://id.example", "client", "https://bot.example/cb"
		}, false},
		{"missing client id", func(c *Config) {
			c.OIDCIssuer, c.OIDCRedirectURL = "https://id.example", "https://bot.example/cb"
		}, true},
		{"missing redirect", func(c *Config) {
			c.OIDCIssuer, c.OIDCClientID = "https://id.example", "client"
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := base
			tt.mutate(&cfg)
			if err := cfg.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfig_OIDCFromEnv(t *testing.T) {
	t.Setenv("OIDC_ISSUER", "https://pocketid.example/")
	t.Setenv("OIDC_CLIENT_ID", "zee")
	t.Setenv("OIDC_ALLOWED_GROUPS", "zee-admin, ops")
	t.Setenv("OIDC_REDIRECT_URL", "")
	t.Setenv("WEB_DASHBOARD_URL", "https://bot.example/")

	cfg := LoadConfig()
	if cfg.OIDCIssuer != "https://pocketid.example" {
		t.Errorf("OIDCIssuer = %q", cfg.OIDCIssuer)
	}
	if cfg.OIDCRedirectURL != "https://bot.example"+OIDCCallbackPath {
		t.Errorf("OIDCRedirectURL = %q", cfg.OIDCRedirectURL)
	}
	if len(cfg.OIDCAllowedGroups) != 2 || cfg.OIDCAllowedGroups[1] != "ops" {
		t.Errorf("OIDCAllowedGroups = %v", cfg.OIDCAllowedGroups)
	}
	if cfg.OIDCProviderName != "Pocket-ID" {
		t.Errorf("OIDCProviderName = %q", cfg.OIDCProviderName)
	}
}

func TestValidate_WebhookRequiresSecret(t *testing.T) {
	cfg := &Config{
		BotTokens:  []string{"token1"},
		OwnerID:    12345,
		UseWebhook: true,
		WebhookURL: "https://example.com/hook",
		RcloneDest: "gdrive:/test",
	}
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "WEBHOOK_SECRET") {
		t.Errorf("expected WEBHOOK_SECRET error, got: %v", err)
	}
	cfg.WebhookSecret = "s3cret"
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error with secret set: %v", err)
	}
}

func TestUseSettingsLayersDatabaseUnderEnvironment(t *testing.T) {
	keyHex := strings.Repeat("ab", 32)
	key, _ := hex.DecodeString(keyHex)
	enc, err := crypto.Encrypt([]byte("123:abc"), key)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("ENCRYPTION_KEY", keyHex)
	t.Setenv("BOT_TOKEN", "")
	t.Setenv("OWNER_ID", "") // empty env must not hide the database value
	t.Setenv("RCLONE_DEST", "env:/dest")
	t.Cleanup(func() {
		dbSettings.Store(nil)
		settingsSource.Store(nil)
	})

	cfg, err := UseSettings(context.Background(), func(context.Context) (map[string]string, error) {
		return map[string]string{
			"BOT_TOKEN":   "enc:" + enc,
			"OWNER_ID":    "42",
			"RCLONE_DEST": "db:/dest",
			"MAX_RETRIES": "7",
			"cookies":     "not a config key",
		}, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.BotToken != "123:abc" || !slices.Equal(cfg.BotTokens, []string{"123:abc"}) {
		t.Errorf("encrypted BOT_TOKEN not decrypted: %q %v", cfg.BotToken, cfg.BotTokens)
	}
	if cfg.OwnerID != 42 {
		t.Errorf("OwnerID = %d, want 42 from database", cfg.OwnerID)
	}
	if cfg.RcloneDest != "env:/dest" {
		t.Errorf("RcloneDest = %q, environment should win", cfg.RcloneDest)
	}
	if cfg.MaxRetries != 7 {
		t.Errorf("MaxRetries = %d, want 7 from database", cfg.MaxRetries)
	}
}

func TestEncryptedValueWithoutKeyIsDropped(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "")
	t.Setenv("BOT_TOKEN", "enc:AAAA")
	if got := LoadConfig().BotToken; got != "" {
		t.Errorf("BotToken = %q, want empty when the key is missing", got)
	}
}
