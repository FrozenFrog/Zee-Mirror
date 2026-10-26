package utils

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"zee-mirror/pkg/i18n"
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
		return i18n.WrapError(err, "core_rclone_token_expired", err.Error(), reason)
	}
	return fmt.Errorf("%w: %s", err, reason)
}
