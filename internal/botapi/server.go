// Package botapi runs the local Telegram Bot API server (tdlib/telegram-bot-api)
// as a child process, so its api_id/api_hash can come from the database
// instead of a separate container's environment.
package botapi

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	// Binary is the executable name looked up in PATH.
	Binary = "telegram-bot-api"
	// DataDir matches the path the rest of the code expects local files under.
	DataDir = "/var/lib/telegram-bot-api"
	// Endpoint is the TELEGRAM_API format pointing at the local server.
	Endpoint = "http://127.0.0.1:8081/bot%s/%s"

	listenAddr   = "127.0.0.1"
	listenPort   = 8081
	readyTimeout = 30 * time.Second
	stopTimeout  = 10 * time.Second
)

type Server struct {
	cmd      *exec.Cmd
	done     chan struct{}
	stopping atomic.Bool
}

// Available reports whether the telegram-bot-api binary is installed.
func Available() bool {
	_, err := exec.LookPath(Binary)
	return err == nil
}

// Start launches the server and waits until it accepts HTTP connections.
func Start(apiID int, apiHash, dataDir string) (*Server, error) {
	if apiID == 0 || apiHash == "" {
		return nil, errors.New("APP_ID and APP_HASH are required for the local Bot API server")
	}
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return nil, fmt.Errorf("create bot api dir: %w", err)
	}

	// #nosec G204 -- fixed binary and internal arguments, no user input
	cmd := exec.Command(Binary,
		"--local",
		"--http-ip-address="+listenAddr,
		"--http-port="+strconv.Itoa(listenPort),
		"--dir="+dataDir,
	)
	// Credentials go through the environment so they don't show up in ps.
	cmd.Env = append(os.Environ(),
		"TELEGRAM_API_ID="+strconv.Itoa(apiID),
		"TELEGRAM_API_HASH="+apiHash,
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s: %w", Binary, err)
	}

	s := &Server{cmd: cmd, done: make(chan struct{})}
	go streamOutput(stdout)
	go streamOutput(stderr)
	go func() {
		if err := cmd.Wait(); !s.stopping.Load() {
			slog.Error("Local Bot API server exited unexpectedly", "error", err)
		}
		close(s.done)
	}()

	if err := s.waitReady(); err != nil {
		s.Stop()
		return nil, err
	}
	slog.Info("Local Bot API server started", "addr", fmt.Sprintf("%s:%d", listenAddr, listenPort), "dir", dataDir)
	return s, nil
}

func (s *Server) waitReady() error {
	url := fmt.Sprintf("http://%s:%d/", listenAddr, listenPort)
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(readyTimeout)
	for time.Now().Before(deadline) {
		select {
		case <-s.done:
			return errors.New("local Bot API server exited during startup")
		default:
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := client.Do(req)
		cancel()
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("local Bot API server not ready after %s", readyTimeout)
}

// Stop asks the server to exit and kills it if it doesn't within stopTimeout.
func (s *Server) Stop() {
	if s == nil || s.cmd.Process == nil {
		return
	}
	select {
	case <-s.done:
		return
	default:
	}
	slog.Info("Stopping local Bot API server...")
	s.stopping.Store(true)
	_ = s.cmd.Process.Signal(syscall.SIGTERM)
	select {
	case <-s.done:
	case <-time.After(stopTimeout):
		_ = s.cmd.Process.Kill()
		<-s.done
	}
}

func streamOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			slog.Debug("telegram-bot-api", "output", line)
		}
	}
}
