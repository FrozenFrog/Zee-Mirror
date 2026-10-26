// Package secretfs keeps rclone.conf and cookies.txt in RAM and mirrors them
// to the settings table.
//
// The files live in a private directory on tmpfs (/dev/shm by default). The
// usual paths under CONFIG_DIR become symlinks into that directory, so rclone,
// yt-dlp and aria2 keep using --config / --cookies unchanged and nothing secret
// is written to disk. Whenever one of those tools rewrites a file (rclone
// refreshing an OAuth token, yt-dlp saving updated cookies) the new content is
// persisted to the database as soon as the write settles.
package secretfs

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"zee-mirror/internal/crypto"
	"zee-mirror/internal/repository"

	"github.com/fsnotify/fsnotify"
)

const (
	RcloneConfig = "rclone.conf"
	Cookies      = "cookies.txt"

	// syncDelay lets a burst of writes (yt-dlp rewrites cookies line by line)
	// settle before the file is read back.
	syncDelay  = 200 * time.Millisecond
	retryDelay = 5 * time.Second
	encPrefix  = "enc:"
)

// settingKeys maps each managed file to its row in the settings table.
var settingKeys = map[string]string{
	RcloneConfig: repository.SettingRcloneConfig,
	Cookies:      repository.SettingCookies,
}

// allowDiskFallback lets dev machines without tmpfs (macOS) run on a regular
// temp dir. On Linux anything that is not RAM-backed is refused.
var allowDiskFallback = runtime.GOOS != "linux"

// Options configures Open.
type Options struct {
	// ConfigDir holds the paths tools are pointed at (CONFIG_DIR).
	ConfigDir string
	// SeedDir optionally holds read-only files (e.g. a mounted k8s Secret)
	// used only while the database has no value yet. They are never deleted.
	SeedDir string
	// RAMBaseDir is the tmpfs mount the private directory is created in.
	RAMBaseDir string
	// EncryptionKey enables AES-256-GCM encryption of the stored values.
	EncryptionKey []byte
}

type Store struct {
	repo      repository.SettingsRepository
	watcher   *fsnotify.Watcher
	hashes    map[string][sha256.Size]byte
	timers    map[string]*time.Timer
	configDir string
	seedDir   string
	ramDir    string
	key       []byte
	wg        sync.WaitGroup
	mu        sync.Mutex // guards hashes, timers and closed
	syncMu    sync.Mutex // serialises file <-> database transfers, guards removed
	closeOnce sync.Once
	closed    bool
	removed   bool
}

// Open restores the managed files from the database into a fresh RAM
// directory, links them into configDir and starts watching for changes.
// A regular file found in configDir is imported into the database (it takes
// precedence over the stored value) and then deleted from disk.
func Open(ctx context.Context, repo repository.SettingsRepository, opts Options) (*Store, error) {
	ramDir, err := makeRAMDir(opts.RAMBaseDir)
	if err != nil {
		return nil, err
	}

	s := &Store{
		repo:      repo,
		configDir: opts.ConfigDir,
		seedDir:   opts.SeedDir,
		ramDir:    ramDir,
		key:       opts.EncryptionKey,
		hashes:    make(map[string][sha256.Size]byte),
		timers:    make(map[string]*time.Timer),
	}

	if len(s.key) == 0 {
		slog.Warn("ENCRYPTION_KEY not set: rclone config and cookies are stored unencrypted in the database")
	}

	for name := range settingKeys {
		if err = s.restore(ctx, name); err != nil {
			s.cleanup()
			return nil, fmt.Errorf("restore %s: %w", name, err)
		}
	}

	s.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		s.cleanup()
		return nil, fmt.Errorf("create secrets watcher: %w", err)
	}
	if err := s.watcher.Add(ramDir); err != nil {
		_ = s.watcher.Close()
		s.cleanup()
		return nil, fmt.Errorf("watch %s: %w", ramDir, err)
	}

	s.wg.Add(1)
	go s.watch()

	slog.Info("Rclone config and cookies mounted in RAM", "dir", ramDir)
	return s, nil
}

// Set replaces a managed file, e.g. when the owner uploads a new cookies.txt.
// Empty data removes the file.
func (s *Store) Set(ctx context.Context, name string, data []byte) error {
	if _, ok := settingKeys[name]; !ok {
		return fmt.Errorf("secretfs: unmanaged file %q", name)
	}

	s.syncMu.Lock()
	defer s.syncMu.Unlock()

	if s.removed {
		return errors.New("secretfs: store is closed")
	}
	if err := s.persist(ctx, name, data); err != nil {
		return err
	}
	return s.writeRAM(name, data)
}

// Close flushes pending changes to the database, then wipes the RAM directory
// and the symlinks pointing into it.
func (s *Store) Close() {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		s.closed = true
		for _, t := range s.timers {
			t.Stop()
		}
		s.mu.Unlock()

		for name := range settingKeys {
			s.sync(name)
		}

		_ = s.watcher.Close()
		s.wg.Wait()
		s.cleanup()
	})
}

func makeRAMDir(baseDir string) (string, error) {
	ok, err := isRAMBacked(baseDir)
	if err != nil || !ok {
		if !allowDiskFallback {
			return "", fmt.Errorf("secrets dir %q is not RAM-backed (tmpfs/ramfs), refusing to put rclone config and cookies on disk; point SECRETS_DIR at a tmpfs mount (statfs error: %v)", baseDir, err)
		}
		slog.Warn("No RAM-backed secrets dir on this platform, using a disk temp dir (development only)", "requested", baseDir)
		if _, statErr := os.Stat(baseDir); statErr != nil {
			baseDir = os.TempDir()
		}
	}

	dir, err := os.MkdirTemp(baseDir, "zee-mirror-")
	if err != nil {
		return "", fmt.Errorf("create secrets dir: %w", err)
	}
	return dir, nil
}

func (s *Store) restore(ctx context.Context, name string) error {
	data, err := s.load(ctx, name)
	if err != nil {
		return err
	}

	linkPath := filepath.Join(s.configDir, name)
	legacy, found, err := readRegularFile(linkPath)
	if err != nil {
		return err
	}
	if found {
		if len(legacy) > 0 {
			if err := s.persist(ctx, name, legacy); err != nil {
				return fmt.Errorf("import %s: %w", linkPath, err)
			}
			data = legacy
		}
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("remove %s after import: %w", linkPath, err)
		}
		slog.Info("Imported file into the database and removed it from disk", "file", linkPath, "bytes", len(legacy))
	}

	if len(data) == 0 && s.seedDir != "" {
		seedPath := filepath.Join(s.seedDir, name)
		seed, err := os.ReadFile(seedPath) // #nosec G304 -- operator-configured dir plus a fixed file name
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		if len(seed) > 0 {
			if err := s.persist(ctx, name, seed); err != nil {
				return fmt.Errorf("seed from %s: %w", seedPath, err)
			}
			data = seed
			slog.Info("Seeded empty database value from file", "file", seedPath)
		}
	}

	if err := s.writeRAM(name, data); err != nil {
		return err
	}
	return s.link(name)
}

func readRegularFile(path string) ([]byte, bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	// Symlinks left over from a previous run are simply replaced.
	if !info.Mode().IsRegular() {
		return nil, false, nil
	}
	data, err := os.ReadFile(path) // #nosec G304 -- path is CONFIG_DIR plus a fixed file name
	return data, err == nil, err
}

// link points configDir/name at the RAM copy. The link stays in place even
// while the RAM file is absent, so `rclone config` and cookie saves create the
// file in RAM instead of on disk.
func (s *Store) link(name string) error {
	linkPath := filepath.Join(s.configDir, name)
	if err := os.Remove(linkPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return os.Symlink(filepath.Join(s.ramDir, name), linkPath)
}

// writeRAM atomically replaces the RAM copy and records its hash first so the
// watcher does not echo the write back to the database.
func (s *Store) writeRAM(name string, data []byte) error {
	s.mu.Lock()
	s.hashes[name] = sha256.Sum256(data)
	s.mu.Unlock()

	target := filepath.Join(s.ramDir, name)
	if len(data) == 0 {
		if err := os.Remove(target); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}

	tmp, err := os.CreateTemp(s.ramDir, "."+name+"-*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), target)
}

func (s *Store) load(ctx context.Context, name string) ([]byte, error) {
	value, err := s.repo.Get(ctx, settingKeys[name])
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(value, encPrefix) {
		return []byte(value), nil
	}
	if len(s.key) == 0 {
		return nil, errors.New("value is encrypted but ENCRYPTION_KEY is not set")
	}
	return crypto.Decrypt(strings.TrimPrefix(value, encPrefix), s.key)
}

func (s *Store) persist(ctx context.Context, name string, data []byte) error {
	value := string(data)
	if len(s.key) > 0 && len(data) > 0 {
		enc, err := crypto.Encrypt(data, s.key)
		if err != nil {
			return err
		}
		value = encPrefix + enc
	}
	return s.repo.Set(ctx, settingKeys[name], value)
}

func (s *Store) watch() {
	defer s.wg.Done()
	for {
		select {
		case ev, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			// rclone saves via temp file + rename, so the final name shows up
			// as Create; yt-dlp truncates and writes in place.
			if name := filepath.Base(ev.Name); settingKeys[name] != "" {
				s.schedule(name, syncDelay)
			}
		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			slog.Warn("Secrets watcher error", "error", err)
		}
	}
}

func (s *Store) schedule(name string, delay time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	if t, ok := s.timers[name]; ok {
		t.Reset(delay)
		return
	}
	s.timers[name] = time.AfterFunc(delay, func() { s.sync(name) })
}

// sync copies the RAM file to the database if it changed since the last
// transfer. A missing file is stored as empty.
func (s *Store) sync(name string) {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()

	// After cleanup the files are gone on purpose; never mirror that.
	if s.removed {
		return
	}

	data, err := os.ReadFile(filepath.Join(s.ramDir, name))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		slog.Warn("Failed to read secret file from RAM", "file", name, "error", err)
		return
	}

	sum := sha256.Sum256(data)
	s.mu.Lock()
	unchanged := s.hashes[name] == sum
	s.mu.Unlock()
	if unchanged {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.persist(ctx, name, data); err != nil {
		slog.Error("Failed to save secret file to the database, will retry", "file", name, "error", err)
		s.schedule(name, retryDelay)
		return
	}

	s.mu.Lock()
	s.hashes[name] = sum
	s.mu.Unlock()
	slog.Info("Secret file changed, saved to the database", "file", name, "bytes", len(data))
}

func (s *Store) cleanup() {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()
	s.removed = true

	for name := range settingKeys {
		linkPath := filepath.Join(s.configDir, name)
		if target, err := os.Readlink(linkPath); err == nil && strings.HasPrefix(target, s.ramDir+string(filepath.Separator)) {
			_ = os.Remove(linkPath)
		}
	}
	if err := os.RemoveAll(s.ramDir); err != nil {
		slog.Warn("Failed to remove secrets dir", "dir", s.ramDir, "error", err)
	}
}
