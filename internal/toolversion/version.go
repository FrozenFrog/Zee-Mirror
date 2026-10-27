package toolversion

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Version struct {
	Current string `json:"current"`
	Latest  string `json:"latest"`
}

var definitions = map[string]struct {
	command string
	args    []string
	url     string
}{
	"aria2":  {"aria2c", []string{"--version"}, "https://api.github.com/repos/aria2/aria2/releases/latest"},
	"rclone": {"rclone", []string{"version"}, "https://api.github.com/repos/rclone/rclone/releases/latest"},
	"yt-dlp": {"yt-dlp", []string{"--version"}, "https://pypi.org/pypi/yt-dlp/json"},
}

var versionPattern = regexp.MustCompile(`(?i)(?:aria2 version |rclone v?)([0-9][^\s,]*)`)

var latestCache = struct {
	sync.Mutex
	values map[string]cachedVersion
}{values: make(map[string]cachedVersion)}

type cachedVersion struct {
	version string
	expires time.Time
}

var client = &http.Client{Timeout: 5 * time.Second}

// Get reads the version from the executable actually selected by PATH and
// looks up the newest stable release. A lookup failure leaves Latest empty.
func Get(ctx context.Context, name string) Version {
	def, ok := definitions[name]
	if !ok {
		return Version{}
	}
	result := Version{}
	commandCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	output, err := exec.CommandContext(commandCtx, def.command, def.args...).Output()
	cancel()
	if err == nil {
		result.Current = parseCurrent(name, string(output))
	}
	result.Latest, _ = Latest(ctx, name)
	return result
}

func parseCurrent(name, output string) string {
	line := strings.TrimSpace(strings.SplitN(output, "\n", 2)[0])
	if name == "yt-dlp" {
		return line
	}
	if match := versionPattern.FindStringSubmatch(line); len(match) == 2 {
		return strings.TrimPrefix(match[1], "v")
	}
	return ""
}

func Latest(ctx context.Context, name string) (string, error) {
	def, ok := definitions[name]
	if !ok {
		return "", errors.New("unknown tool")
	}
	latestCache.Lock()
	if cached, ok := latestCache.values[name]; ok && time.Now().Before(cached.expires) {
		latestCache.Unlock()
		return cached.version, nil
	}
	latestCache.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, def.url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Zee-Mirror")
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", errors.New(response.Status)
	}
	var release struct {
		TagName string `json:"tag_name"`
		Info    struct {
			Version string `json:"version"`
		} `json:"info"`
	}
	// PyPI's project JSON includes all historical releases; yt-dlp's response
	// already exceeds 1 MiB.
	if err := json.NewDecoder(io.LimitReader(response.Body, 8<<20)).Decode(&release); err != nil {
		return "", err
	}
	version := release.Info.Version
	if name != "yt-dlp" {
		version = normalizeReleaseTag(release.TagName)
	}
	if version == "" {
		return "", errors.New("latest version missing from response")
	}
	latestCache.Lock()
	latestCache.values[name] = cachedVersion{version: version, expires: time.Now().Add(30 * time.Minute)}
	latestCache.Unlock()
	return version, nil
}

func normalizeReleaseTag(tag string) string {
	return strings.TrimPrefix(strings.TrimPrefix(tag, "release-"), "v")
}

func All(ctx context.Context) map[string]Version {
	results := make(map[string]Version, len(definitions))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, name := range []string{"aria2", "rclone", "yt-dlp"} {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			version := Get(ctx, name)
			mu.Lock()
			results[name] = version
			mu.Unlock()
		}(name)
	}
	wg.Wait()
	return results
}
