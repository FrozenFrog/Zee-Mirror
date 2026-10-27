package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"zee-mirror/internal/config"
	"zee-mirror/internal/repository"
	"zee-mirror/internal/toolversion"
)

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method == http.MethodPost {
		var settingsUpdate struct {
			DefaultMode        string `json:"DefaultMode"`
			AutoDeleteMessages bool   `json:"AutoDeleteMessages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&settingsUpdate); err != nil {
			http.Error(w, "Invalid settings", http.StatusBadRequest)
			return
		}

		s.Service.Settings.Mu.Lock()
		s.Service.Settings.AutoDeleteMessages = settingsUpdate.AutoDeleteMessages
		s.Service.Settings.DefaultMode = settingsUpdate.DefaultMode
		s.Service.Settings.Mu.Unlock()

		_ = s.Service.SettingsRepo.Set(ctx, repository.SettingAutoDeleteMessages, fmt.Sprintf("%v", settingsUpdate.AutoDeleteMessages))
		_ = s.Service.SettingsRepo.Set(ctx, repository.SettingDefaultMode, settingsUpdate.DefaultMode)
		w.WriteHeader(http.StatusOK)
		return
	}

	s.Service.Settings.Mu.RLock()
	defer s.Service.Settings.Mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.Service.Settings); err != nil {
		slog.Error("Failed to encode settings response", "error", err)
	}
}

// handleConfig shows and edits the settings table as KEY=value lines.
// Secrets are redacted, and the bootstrap keys (DB_DRIVER, DATABASE_URL,
// ENCRYPTION_KEY, CONFIG_DIR) stay environment-only.
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		if err := json.NewEncoder(w).Encode(map[string]string{"config": config.FormatSettings()}); err != nil {
			slog.Error("Failed to encode config response", "error", err)
		}
	case http.MethodPost:
		var req struct {
			Config string `json:"config"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
			return
		}

		changed, overridden, err := config.SaveSettings(r.Context(), s.Service.SettingsRepo, req.Config)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		slog.Info("Settings updated from dashboard", "keys", changed, "overriddenByEnv", overridden, "ip", getClientIP(r))
		s.applyConfig(config.Reload())
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "Settings saved. Some changes need a restart.",
			"changed":    changed,
			"overridden": overridden,
		})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) applyConfig(newCfg *config.Config) {
	s.Service.TaskManager.Mu.Lock()
	s.Service.TaskManager.Config = newCfg
	s.Service.TaskManager.MaxConcurrent = newCfg.MaxConcurrentDownloads
	s.Service.TaskManager.StopDuplicate = newCfg.StopDuplicate
	s.Service.TaskManager.Mu.Unlock()
}

func (s *Server) handleConfigReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	newCfg := config.Reload()
	s.applyConfig(newCfg)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":                   "ok",
		"message":                  "Config reloaded successfully",
		"max_concurrent_downloads": newCfg.MaxConcurrentDownloads,
		"stop_duplicate":           newCfg.StopDuplicate,
		"log_level":                newCfg.LogLevel,
	}); err != nil {
		slog.Error("Failed to encode config reload response", "error", err)
	}
}

func (s *Server) handleUpdateTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Tool string `json:"tool"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&request); err != nil && !errors.Is(err, io.EOF) {
			writeToolError(w, http.StatusBadRequest, "Invalid request", "")
			return
		}
	}
	if request.Tool == "" {
		request.Tool = "yt-dlp"
	}
	if request.Tool != "yt-dlp" && request.Tool != "rclone" {
		writeToolError(w, http.StatusBadRequest, "Unsupported tool", "")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	var cmdArgs []string
	var command string
	if request.Tool == "rclone" {
		home, err := os.UserHomeDir()
		if err != nil {
			writeToolError(w, http.StatusInternalServerError, err.Error(), "")
			return
		}
		path := filepath.Join(home, ".local", "bin", "rclone")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			writeToolError(w, http.StatusInternalServerError, err.Error(), "")
			return
		}
		command = "rclone"
		cmdArgs = []string{"selfupdate", "--output", path}
	} else {
		command = "python3"
		cmdArgs = []string{"-m", "pip", "install", "--user", "--break-system-packages", "--no-cache-dir", "--upgrade", "yt-dlp[default]"}
	}
	cmd := execCommand(ctx, command, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		writeToolError(w, http.StatusInternalServerError, err.Error(), strings.TrimSpace(string(output)))
		return
	}
	version := getToolVersion(ctx, request.Tool)
	result := map[string]interface{}{"success": true, "tool": request.Tool, "current": version.Current, "latest": version.Latest}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		slog.Error("Failed to encode tool update response", "error", err)
	}
}

var getToolVersion = toolversion.Get

func (s *Server) handleToolVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeToolError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(toolversion.All(ctx)); err != nil {
		slog.Error("Failed to encode tool versions", "error", err)
	}
}

func writeToolError(w http.ResponseWriter, status int, message, output string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message, "output": output})
}
