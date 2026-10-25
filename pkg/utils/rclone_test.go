package utils

import (
	"os/exec"
	"strings"
	"testing"
)

func TestRcloneError(t *testing.T) {
	cmd := exec.Command("sh", "-c", `echo "2026/10/26 02:39:32 CRITICAL: Failed to create file system: couldn't fetch token: invalid_grant: maybe token expired?" >&2; exit 1`)
	_, err := cmd.Output()
	got := RcloneError(err).Error()
	for _, want := range []string{"exit status 1", "đăng nhập lại", "couldn't fetch token"} {
		if !strings.Contains(got, want) {
			t.Errorf("RcloneError() = %q, missing %q", got, want)
		}
	}
	if strings.Contains(got, "2026/10/26") {
		t.Errorf("RcloneError() kept the log timestamp: %q", got)
	}

	_, err = exec.Command("sh", "-c", "echo 'directory not found' >&2; exit 3").Output()
	if got := RcloneError(err).Error(); got != "exit status 3: directory not found" {
		t.Errorf("RcloneError() = %q", got)
	}
}
