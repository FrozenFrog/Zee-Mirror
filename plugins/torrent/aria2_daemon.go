package torrent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Aria2Daemon struct {
	Cmd       *exec.Cmd
	ConfigDir string
	RPCSecret string
}

var ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func NewAria2Daemon(configDir, rpcSecret string) *Aria2Daemon {
	return &Aria2Daemon{
		ConfigDir: configDir,
		RPCSecret: rpcSecret,
	}
}

func (d *Aria2Daemon) Start() error {
	slog.Info("Starting aria2c daemon...")

	sessionPath := filepath.Join(d.ConfigDir, "aria2.session")
	_ = os.WriteFile(sessionPath, []byte(""), 0600)

	args := []string{
		"--enable-rpc",
		"--rpc-listen-all=false",
		"--rpc-listen-port=6800",
		"--save-session=" + sessionPath,
		"--save-session-interval=60",
		"--max-concurrent-downloads=3",
		"--check-certificate=false",
	}

	if d.RPCSecret != "" {
		args = append(args, "--rpc-secret="+d.RPCSecret)
	}

	// #nosec G702 -- args are internal constants plus optional configured secret; no user input, no shell
	d.Cmd = execCommand(context.Background(), "aria2c", args...)
	stdoutPipe, err := d.Cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get aria2c stdout pipe: %v", err)
	}
	stderrPipe, err := d.Cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get aria2c stderr pipe: %v", err)
	}

	if err := d.Cmd.Start(); err != nil {
		return fmt.Errorf("failed to start aria2c: %v", err)
	}

	go streamAria2Output(stdoutPipe, false)
	go streamAria2Output(stderrPipe, true)

	time.Sleep(2 * time.Second)
	slog.Info("aria2c daemon started successfully")

	return nil
}

func (d *Aria2Daemon) Stop() {
	if d.Cmd != nil && d.Cmd.Process != nil {
		slog.Info("Stopping aria2c daemon...")
		_ = d.Cmd.Process.Kill()
	}
}

func streamAria2Output(r io.Reader, isErr bool) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		line = ansiEscapePattern.ReplaceAllString(line, "")

		if isErr || strings.Contains(line, "[WARN]") {
			slog.Warn("aria2c", "output", line)
			continue
		}

		if strings.HasPrefix(line, "[#") ||
			strings.HasPrefix(line, "*** Download Progress Summary") ||
			strings.HasPrefix(line, "===") ||
			strings.HasPrefix(line, "---") ||
			strings.HasPrefix(line, "FILE:") {
			continue
		}

		slog.Info("aria2c", "output", line)
	}
	if err := scanner.Err(); err != nil {
		slog.Warn("aria2c output stream error", "error", err)
	}
}
