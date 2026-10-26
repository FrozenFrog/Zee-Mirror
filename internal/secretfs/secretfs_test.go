package secretfs

import (
	"context"
	"database/sql"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"zee-mirror/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memRepo struct {
	data map[string]string
	mu   sync.Mutex
}

func newMemRepo() *memRepo { return &memRepo{data: map[string]string{}} }

func (r *memRepo) Get(_ context.Context, key string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.data[key]
	if !ok {
		return "", sql.ErrNoRows
	}
	return v, nil
}

func (r *memRepo) Set(_ context.Context, key, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[key] = value
	return nil
}

func (r *memRepo) AllSettings(_ context.Context) (map[string]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return maps.Clone(r.data), nil
}

func (r *memRepo) DeleteSetting(_ context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, key)
	return nil
}

func (r *memRepo) value(key string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.data[key]
}

func openTest(t *testing.T, repo *memRepo, configDir string, key []byte) *Store {
	t.Helper()
	allowDiskFallback = true
	s, err := Open(context.Background(), repo, Options{ConfigDir: configDir, RAMBaseDir: t.TempDir(), EncryptionKey: key})
	require.NoError(t, err)
	return s
}

func TestOpenImportsLegacyFileAndLinksIt(t *testing.T) {
	repo := newMemRepo()
	configDir := t.TempDir()
	legacy := filepath.Join(configDir, RcloneConfig)
	require.NoError(t, os.WriteFile(legacy, []byte("[gdrive]\ntype = drive\n"), 0o600))

	s := openTest(t, repo, configDir, nil)
	defer s.Close()

	assert.Equal(t, "[gdrive]\ntype = drive\n", repo.value(repository.SettingRcloneConfig))

	info, err := os.Lstat(legacy)
	require.NoError(t, err)
	assert.NotZero(t, info.Mode()&os.ModeSymlink, "config path should now be a symlink into RAM")

	got, err := os.ReadFile(legacy)
	require.NoError(t, err)
	assert.Equal(t, "[gdrive]\ntype = drive\n", string(got))

	// No stored cookies: the link dangles so callers see the file as missing.
	_, err = os.Stat(filepath.Join(configDir, Cookies))
	assert.True(t, os.IsNotExist(err))
}

func TestExternalRewriteIsSavedToDatabase(t *testing.T) {
	repo := newMemRepo()
	require.NoError(t, repo.Set(context.Background(), repository.SettingRcloneConfig, "token = old\n"))
	configDir := t.TempDir()

	s := openTest(t, repo, configDir, nil)
	defer s.Close()

	// Mimic rclone's save: resolve the symlink, write a temp file next to the
	// target, rename it over the target.
	target, err := os.Readlink(filepath.Join(configDir, RcloneConfig))
	require.NoError(t, err)
	tmp := target + ".tmp123"
	require.NoError(t, os.WriteFile(tmp, []byte("token = new\n"), 0o600))
	require.NoError(t, os.Rename(tmp, target))

	assert.Eventually(t, func() bool { return repo.value(repository.SettingRcloneConfig) == "token = new\n" },
		3*time.Second, 20*time.Millisecond)

	// In-place writes through the symlink (yt-dlp cookie save) are caught too.
	require.NoError(t, os.WriteFile(filepath.Join(configDir, Cookies), []byte("# Netscape HTTP Cookie File\n"), 0o600))
	assert.Eventually(t, func() bool { return repo.value(repository.SettingCookies) == "# Netscape HTTP Cookie File\n" },
		3*time.Second, 20*time.Millisecond)
}

func TestEncryptedRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	repo := newMemRepo()
	configDir := t.TempDir()

	s := openTest(t, repo, configDir, key)
	require.NoError(t, s.Set(context.Background(), Cookies, []byte("secret-cookie")))
	stored := repo.value(repository.SettingCookies)
	assert.True(t, strings.HasPrefix(stored, encPrefix))
	assert.NotContains(t, stored, "secret-cookie")
	s.Close()

	s2 := openTest(t, repo, configDir, key)
	defer s2.Close()
	got, err := os.ReadFile(filepath.Join(configDir, Cookies))
	require.NoError(t, err)
	assert.Equal(t, "secret-cookie", string(got))
}

func TestCloseWipesRAMButKeepsDatabase(t *testing.T) {
	repo := newMemRepo()
	require.NoError(t, repo.Set(context.Background(), repository.SettingCookies, "c=1"))
	configDir := t.TempDir()

	s := openTest(t, repo, configDir, nil)
	ramDir := s.ramDir
	s.Close()

	_, err := os.Stat(ramDir)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Lstat(filepath.Join(configDir, Cookies))
	assert.True(t, os.IsNotExist(err))
	assert.Equal(t, "c=1", repo.value(repository.SettingCookies))
}

func TestSeedOnlyUsedWhenDatabaseEmpty(t *testing.T) {
	allowDiskFallback = true
	seedDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(seedDir, RcloneConfig), []byte("token = seed\n"), 0o600))
	repo := newMemRepo()
	configDir := t.TempDir()
	opts := Options{ConfigDir: configDir, SeedDir: seedDir, RAMBaseDir: t.TempDir()}

	s, err := Open(context.Background(), repo, opts)
	require.NoError(t, err)
	assert.Equal(t, "token = seed\n", repo.value(repository.SettingRcloneConfig))
	require.NoError(t, s.Set(context.Background(), RcloneConfig, []byte("token = refreshed\n")))
	s.Close()

	// Restart with the same (stale) seed: the refreshed token must survive.
	s, err = Open(context.Background(), repo, opts)
	require.NoError(t, err)
	defer s.Close()
	assert.Equal(t, "token = refreshed\n", repo.value(repository.SettingRcloneConfig))
	_, err = os.Stat(filepath.Join(seedDir, RcloneConfig))
	assert.NoError(t, err, "seed files are never deleted")
}
