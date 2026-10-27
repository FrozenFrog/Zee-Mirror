package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"zee-mirror/internal/config"
	"zee-mirror/internal/repository"
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

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	cmd := execCommand(ctx,
		"python3", "-m", "pip", "install",
		"--user",
		"--break-system-packages",
		"--no-cache-dir",
		"--upgrade",
		"yt-dlp[default]",
	)
	output, err := cmd.CombinedOutput()

	result := map[string]interface{}{
		"success": err == nil,
		"output":  string(output),
	}
	if err != nil {
		result["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		slog.Error("Failed to encode git status response", "error", err)
	}
}
