package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"zee-mirror/internal/domain"
	"zee-mirror/internal/queue"
	"zee-mirror/internal/repository"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TaskRecovery struct {
	DB          repository.TaskRepository
	TaskManager *TaskManager
}

func NewTaskRecovery(db repository.TaskRepository, tm *TaskManager) *TaskRecovery {
	return &TaskRecovery{
		DB:          db,
		TaskManager: tm,
	}
}

func (tr *TaskRecovery) RecoverIncompleteTasks() error {
	slog.Info("Checking for incomplete tasks...")

	ctx := context.Background()
	tasks, err := tr.DB.GetRecoverable(ctx)
	if err != nil {
		return fmt.Errorf("%w: failed to get recoverable tasks: %v", domain.ErrStorage, err)
	}

	if len(tasks) == 0 {
		slog.Info("No incomplete tasks found")
		return nil
	}

	slog.Info("Found incomplete tasks", "count", len(tasks))

	recovered := 0
	skipped := 0

	for _, record := range tasks {
		if time.Since(record.CreatedAt) > 24*time.Hour {
			slog.Info("Skipping old task", "taskID", record.ID)
			_ = tr.DB.UpdateStatus(ctx, record.ID, string(domain.StatusExpired), "Task đã hết hạn")
			skipped++
			continue
		}

		task := tr.createTaskFromRecord(record)
		if task == nil {
			skipped++
			continue
		}

		slog.Info("Recovering task", "taskID", record.ID)

		tr.TaskManager.Mu.Lock()
		if _, alreadyExists := tr.TaskManager.Tasks[task.ID]; alreadyExists {
			tr.TaskManager.Mu.Unlock()
			slog.Info("Skipping already loaded task", "taskID", record.ID)
			skipped++
			continue
		}
		tr.TaskManager.Tasks[task.ID] = task
		tr.TaskManager.Mu.Unlock()

		tr.TaskManager.Queue.Enqueue(task, queue.PriorityLow)
		select {
		case tr.TaskManager.QueueSignal <- struct{}{}:
		default:
		}

		recovered++
	}

	slog.Info("Recovery summary", "recovered", recovered, "skipped", skipped)
	return nil
}

func (tr *TaskRecovery) createTaskFromRecord(record domain.TaskRecord) *Task {
	ctx, cancel := context.WithCancel(context.Background())

	task := &Task{
		Task: domain.Task{
			ID:         record.ID,
			GID:        record.GID,
			Type:       domain.TaskType(record.Type),
			Status:     domain.StatusQueued,
			URL:        record.URL,
			FileName:   record.FileName,
			ChatID:     record.ChatID,
			UserID:     record.UserID,
			CreatedAt:  time.Now().UTC(),
			Ctx:        ctx,
			CancelFunc: cancel,
			Zip:        record.Zip,
			Unzip:      record.Unzip,
			Password:   record.Password,
			Quality:    record.Quality,
			RetryCount: record.RetryCount,
			MaxRetries: tr.TaskManager.Config.MaxRetries,
		},
		DB: tr.DB,
	}

	return task
}

func (s *BotService) HandleRecoveryStatus(message *tgbotapi.Message) {
	if message.From.ID != s.Config.OwnerID {
		s.reply(message, "❌ *Truy cập bị từ chối*")
		return
	}

	ctx := context.Background()
	tasks, err := s.DB.GetRecoverable(ctx)
	if err != nil {
		s.reply(message, "❌ *Không lấy được dữ liệu khôi phục*")
		return
	}

	if len(tasks) == 0 {
		s.reply(message, "✅ *Không có task nào cần khôi phục*")
		return
	}

	var text strings.Builder
	text.WriteString("🔄 *Task có thể khôi phục*\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for i, t := range tasks {
		if i >= 10 {
			fmt.Fprintf(&text, "\n_\\.\\.\\. và %d task khác_", len(tasks)-10)
			break
		}
		fmt.Fprintf(&text, "• `%s` • %s\n", t.ID, utils.EscapeMarkdownV2(t.Type))
	}

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━\n")
	text.WriteString("Dùng /recover để khôi phục tất cả task")

	s.reply(message, text.String())
}

func (s *BotService) HandleRecover(message *tgbotapi.Message) {
	if message.From.ID != s.Config.OwnerID {
		s.reply(message, "❌ *Truy cập bị từ chối*")
		return
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, "🔄 *Đang khôi phục task\\.\\.\\.*")
	statusMsg.ParseMode = MarkdownV2
	sent, _ := s.Bot.Send(statusMsg)

	recovery := NewTaskRecovery(s.DB, s.TaskManager)
	if err := recovery.RecoverIncompleteTasks(); err != nil {
		s.EditMessage(sent.Chat.ID, sent.MessageID, fmt.Sprintf("❌ *Khôi phục thất bại*\n\n%s", utils.EscapeMarkdownV2(err.Error())))
		return
	}

	activeTasks := s.TaskManager.GetActiveTasks()
	s.EditMessage(sent.Chat.ID, sent.MessageID, fmt.Sprintf("✅ *Khôi phục hoàn tất*\n\nTask đang chạy: %d", len(activeTasks)))
}
