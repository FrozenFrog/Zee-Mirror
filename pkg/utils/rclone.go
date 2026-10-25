package utils

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var rcloneLogPrefix = regexp.MustCompile(`^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} (CRITICAL|ERROR|NOTICE)\s*:\s*`)

// RcloneError adds the reason rclone printed on stderr to a failed command's
// error, which on its own only says "exit status 1". It expects the error of
// cmd.Output(), which captures stderr into *exec.ExitError.
func RcloneError(err error) error {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || len(exitErr.Stderr) == 0 {
		return err
	}

	lines := strings.Split(strings.TrimSpace(string(exitErr.Stderr)), "\n")
	reason := rcloneLogPrefix.ReplaceAllString(strings.TrimSpace(lines[len(lines)-1]), "")
	if len(reason) > 300 {
		reason = reason[:300] + "…"
	}
	if strings.Contains(reason, "invalid_grant") {
		return fmt.Errorf("%w: token của remote đã hết hạn hoặc bị thu hồi, cần đăng nhập lại (rclone config reconnect): %s", err, reason)
	}
	return fmt.Errorf("%w: %s", err, reason)
}
