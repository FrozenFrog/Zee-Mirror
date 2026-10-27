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
	message := new(BotService).FormatHealthCheckLang("en", checks)
	for _, version := range []string{"1\\.36\\.0", "1\\.37\\.0", "1\\.75\\.1", "1\\.75\\.2", "2026\\.10\\.26"} {
		if !strings.Contains(message, version) {
			t.Errorf("missing version %q in health check", version)
		}
	}
}
