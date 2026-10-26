package config

import (
	"context"
	"encoding/hex"
	"maps"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

type memStore map[string]string

func (m memStore) AllSettings(context.Context) (map[string]string, error) {
	return maps.Clone(m), nil
}

func (m memStore) Set(_ context.Context, key, value string) error {
	m[key] = value
	return nil
}

func (m memStore) DeleteSetting(_ context.Context, key string) error {
	delete(m, key)
	return nil
}

func setTestKey(t *testing.T) {
	t.Helper()
	t.Setenv("ENCRYPTION_KEY", strings.Repeat("cd", 32))
}

// Every key LoadConfig reads must be a bootstrap or a registered setting,
// otherwise it is never saved to the database nor shown on the dashboard.
func TestSettingKeysCoverLoadConfig(t *testing.T) {
	src, err := os.ReadFile("config.go")
	if err != nil {
		t.Fatal(err)
	}
	read := map[string]bool{}
	for _, m := range regexp.MustCompile(`src\.(?:str|strOr|int|bool)\("([A-Z0-9_]+)"`).FindAllStringSubmatch(string(src), -1) {
		read[m[1]] = true
	}
	for name := range read {
		if _, ok := findSettingKey(name); !ok && !isBootstrapKey(name) {
			t.Errorf("%s is read by LoadConfig but missing from settingKeys", name)
		}
	}
	for _, k := range settingKeys {
		if !read[k.name] {
			t.Errorf("%s is in settingKeys but LoadConfig never reads it", k.name)
		}
		if k.name != strings.ToUpper(k.name) {
			t.Errorf("%s must be uppercase", k.name)
		}
	}
}

func TestResolveDBDriver(t *testing.T) {
	tests := []struct{ driver, dsn, want string }{
		{"", "", "sqlite"},
		{"", "/app/data/zee.db", "sqlite"},
		{"", "postgres://u:p@h/db", "postgres"},
		{"", "postgresql://u:p@h/db", "postgres"},
		{"Postgres", "", "postgres"},
		{"sqlite", "postgres://u:p@h/db", "sqlite"},
		{"mysql", "", "mysql"},
	}
	for _, tt := range tests {
		if got := resolveDBDriver(tt.driver, tt.dsn); got != tt.want {
			t.Errorf("resolveDBDriver(%q, %q) = %q, want %q", tt.driver, tt.dsn, got, tt.want)
		}
	}
}

func TestValidateBootstrap(t *testing.T) {
	key := []byte(strings.Repeat("k", 32))
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"sqlite", Config{DBDriver: "sqlite", EncryptionKey: key}, false},
		{"postgres", Config{DBDriver: "postgres", DatabaseURL: "postgres://h/db", EncryptionKey: key}, false},
		{"postgres without url", Config{DBDriver: "postgres", EncryptionKey: key}, true},
		{"unknown driver", Config{DBDriver: "mysql", EncryptionKey: key}, true},
		{"no key", Config{DBDriver: "sqlite"}, true},
	}
	for _, tt := range tests {
		if err := tt.cfg.ValidateBootstrap(); (err != nil) != tt.wantErr {
			t.Errorf("%s: ValidateBootstrap() = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestSyncEnvSavesDeclaredSettings(t *testing.T) {
	setTestKey(t)
	t.Setenv("BOT_TOKEN", "123:secret")
	t.Setenv("OWNER_ID", "42")
	t.Setenv("RCLONE_DEST", "")
	t.Setenv("DATABASE_URL", "postgres://u:p@h/db")
	store := memStore{"RCLONE_DEST": "db:/dest", "OWNER_ID": "42"}

	written, err := SyncEnv(context.Background(), store)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(written, "BOT_TOKEN") || slices.Contains(written, "OWNER_ID") {
		t.Errorf("written = %v, want BOT_TOKEN but not the unchanged OWNER_ID", written)
	}
	if !strings.HasPrefix(store["BOT_TOKEN"], "enc:") {
		t.Errorf("secret stored unencrypted: %q", store["BOT_TOKEN"])
	}
	if store["RCLONE_DEST"] != "db:/dest" {
		t.Errorf("empty env must not overwrite the database, got %q", store["RCLONE_DEST"])
	}
	if _, ok := store["DATABASE_URL"]; ok {
		t.Error("bootstrap key DATABASE_URL must never be stored")
	}

	// A second run sees the same (decrypted) value and writes nothing.
	if written, err = SyncEnv(context.Background(), store); err != nil || slices.Contains(written, "BOT_TOKEN") {
		t.Errorf("second sync wrote %v, err %v", written, err)
	}
}

func TestSyncEnvRefusesPlaintextSecretWithoutKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "")
	t.Setenv("BOT_TOKEN", "123:secret")
	store := memStore{}
	if _, err := SyncEnv(context.Background(), store); err == nil {
		t.Fatal("expected an error without ENCRYPTION_KEY")
	}
	if _, ok := store["BOT_TOKEN"]; ok {
		t.Error("secret written in plaintext")
	}
}

func TestSaveSettings(t *testing.T) {
	setTestKey(t)
	for _, k := range []string{"BOT_TOKEN", "OWNER_ID", "RCLONE_DEST", "MAX_RETRIES", "WEB_DASHBOARD_TOKEN"} {
		t.Setenv(k, "")
	}
	t.Setenv("LOG_LEVEL", "debug")
	key, _ := hex.DecodeString(strings.Repeat("cd", 32))
	src := source{key: key}
	oldToken, _ := src.encode("BOT_TOKEN", "old")
	store := memStore{"BOT_TOKEN": oldToken, "MAX_RETRIES": "3", "RCLONE_DEST": "a:/"}

	text := strings.Join([]string{
		"# comment",
		"BOT_TOKEN=" + redacted,
		"OWNER_ID=42",
		`RCLONE_DEST="b:/x"`,
		"MAX_RETRIES=",
		"WEB_DASHBOARD_TOKEN=new-token-0123456789",
		"LOG_LEVEL=info",
	}, "\n")
	changed, overridden, err := SaveSettings(context.Background(), store, text)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"OWNER_ID", "RCLONE_DEST", "MAX_RETRIES", "WEB_DASHBOARD_TOKEN"}; !slices.Equal(changed, want) {
		t.Errorf("changed = %v, want %v", changed, want)
	}
	if !slices.Equal(overridden, []string{"LOG_LEVEL"}) {
		t.Errorf("overridden = %v", overridden)
	}
	if store["BOT_TOKEN"] != oldToken {
		t.Error("redacted secret must keep the stored value")
	}
	if store["RCLONE_DEST"] != "b:/x" || store["OWNER_ID"] != "42" {
		t.Errorf("values not saved: %v", store)
	}
	if _, ok := store["MAX_RETRIES"]; ok {
		t.Error("empty value must delete the row")
	}
	if got := src.decrypt("WEB_DASHBOARD_TOKEN", store["WEB_DASHBOARD_TOKEN"]); got != "new-token-0123456789" || !strings.HasPrefix(store["WEB_DASHBOARD_TOKEN"], "enc:") {
		t.Errorf("WEB_DASHBOARD_TOKEN stored as %q", store["WEB_DASHBOARD_TOKEN"])
	}

	for _, bad := range []string{"DATABASE_URL=x", "NOPE=1", "no equals sign"} {
		if _, _, err := SaveSettings(context.Background(), store, bad); err == nil {
			t.Errorf("SaveSettings(%q) should fail", bad)
		}
	}
}

func TestFormatSettingsHidesSecrets(t *testing.T) {
	setTestKey(t)
	t.Setenv("BOT_TOKEN", "")
	t.Setenv("OWNER_ID", "")
	t.Setenv("LOG_LEVEL", "debug")
	t.Cleanup(func() { dbSettings.Store(nil) })
	dbSettings.Store(&map[string]string{"BOT_TOKEN": "enc:abcd", "OWNER_ID": "42"})

	out := FormatSettings()
	for _, want := range []string{"BOT_TOKEN=" + redacted + "\n", "OWNER_ID=42\n", "# (env) LOG_LEVEL=debug\n", "# RCLONE_DEST=\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("FormatSettings() missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "enc:abcd") || strings.Contains(out, "ENCRYPTION_KEY") {
		t.Errorf("FormatSettings() leaks a secret or bootstrap key:\n%s", out)
	}
}
