package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"zee-mirror/internal/domain"
)

// historyTask is the dashboard's view of a stored task. domain.TaskRecord has
// no JSON tags, so encoding it directly yields "FileName"-style keys the
// dashboard cannot read.
type historyTask struct {
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	URL         string     `json:"url"`
	FileName    string     `json:"file_name"`
	Status      string     `json:"status"`
	MD5         string     `json:"md5"`
	Error       string     `json:"error,omitempty"`
	FileSize    int64      `json:"file_size"`
}

func newHistoryTask(t *domain.TaskRecord) historyTask {
	h := historyTask{
		CreatedAt: t.CreatedAt,
		ID:        t.ID,
		Type:      t.Type,
		URL:       t.URL,
		FileName:  t.FileName,
		Status:    t.Status,
		MD5:       t.MD5,
		Error:     t.Error,
		FileSize:  t.TotalSize,
	}
	if t.CompletedAt.Valid {
		h.CompletedAt = &t.CompletedAt.Time
	}
	return h
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req struct {
			URL      string `json:"url"`
			Type     string `json:"type"`
			FileName string `json:"fileName"`
			Password string `json:"password"`
			Quality  string `json:"quality"`
			Zip      bool   `json:"zip"`
			Unzip    bool   `json:"unzip"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if req.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}
		if req.Type == "" {
			req.Type = "mirror"
		}

		task, err := s.Service.TaskManager.CreateTask(domain.TaskType(req.Type), req.URL, req.FileName, 0, 0, 0, 0, req.Zip, req.Unzip, req.Password, req.Quality, 0, "", false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(task.GetSnapshot()); err != nil {
			slog.Error("Failed to encode created task response", "error", err)
		}
		return
	}

	if r.Method == http.MethodDelete {
		taskID := r.URL.Query().Get("id")
		if taskID != "" {
			s.Service.TaskManager.CancelTask(taskID)
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	tasks := s.Service.TaskManager.GetActiveTasks()
	var snapshots []interface{}
	for _, t := range tasks {
		snapshots = append(snapshots, t.GetSnapshot())
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(snapshots); err != nil {
		slog.Error("Failed to encode tasks response", "error", err)
	}
}

func (s *Server) handleGetTaskHistory(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}
	status := r.URL.Query().Get("status")
	userID, _ := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)

	filter := domain.TaskFilter{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
		Status: status,
	}

	tasks, err := s.Service.DB.ListTasks(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	history := make([]historyTask, 0, len(tasks))
	for i := range tasks {
		history = append(history, newHistoryTask(&tasks[i]))
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(history); err != nil {
		slog.Error("Failed to encode task history response", "error", err)
	}
}
