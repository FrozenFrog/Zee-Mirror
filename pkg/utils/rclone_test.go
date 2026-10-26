package utils

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"zee-mirror/pkg/i18n"
)

func TestRcloneError(t *testing.T) {
	cmd := exec.Command("sh", "-c", `echo "2026/10/26 02:39:32 CRITICAL: Failed to create file system: couldn't fetch token: invalid_grant: maybe token expired?" >&2; exit 1`)
	_, err := cmd.Output()
	rerr := RcloneError(err)
	got := rerr.Error()
	for _, want := range []string{"exit status 1", "log in again", "couldn't fetch token"} {
		if !strings.Contains(got, want) {
			t.Errorf("RcloneError() = %q, missing %q", got, want)
		}
	}
	if vi := i18n.Message("vi", rerr); !strings.Contains(vi, "đăng nhập lại") || !strings.Contains(vi, "exit status 1") {
		t.Errorf("i18n.Message(vi) = %q", vi)
	}
	var exitErr *exec.ExitError
	if !errors.As(rerr, &exitErr) {
		t.Error("exit error not reachable")
	}
	if strings.Contains(got, "2026/10/26") {
		t.Errorf("RcloneError() kept the log timestamp: %q", got)
	}

	_, err = exec.Command("sh", "-c", "echo 'directory not found' >&2; exit 3").Output()
	if got := RcloneError(err).Error(); got != "exit status 3: directory not found" {
		t.Errorf("RcloneError() = %q", got)
	}
}
