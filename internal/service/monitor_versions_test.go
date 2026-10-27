package service

import (
	"strings"
	"testing"
)

func TestFormatHealthCheckIncludesToolVersions(t *testing.T) {
	checks := []HealthCheck{
		{Name: "Aria2", Healthy: true, Message: "available", CurrentVersion: "1.36.0", LatestVersion: "1.37.0"},
		{Name: "Rclone", Healthy: true, Message: "available", CurrentVersion: "1.75.1", LatestVersion: "1.75.2"},
		{Name: "yt-dlp", Healthy: true, Message: "available", CurrentVersion: "2026.10.26", LatestVersion: "2026.10.26"},
	}
	message := new(BotService).FormatHealthCheckLang("vi", checks)
	for _, line := range []string{
		"└ Bản đang dùng: 1\\.36\\.0",
		"└ Mới nhất: 1\\.37\\.0",
		"└ Bản đang dùng: 1\\.75\\.1",
		"└ Mới nhất: 1\\.75\\.2",
		"└ Bản đang dùng: 2026\\.10\\.26",
		"└ Mới nhất: 2026\\.10\\.26",
	} {
		if !strings.Contains(message, line) {
			t.Errorf("missing line %q in health check", line)
		}
	}
}
