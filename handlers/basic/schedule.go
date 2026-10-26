package basic

import (
	"context"
	"strings"
	"time"

	"zee-mirror/internal/domain"
	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

func HandleSchedule(s *service.BotService, message *tgbotapi.Message, args string) {
	lang := senderLang(s, message)
	parts := strings.Fields(args)
	if len(parts) < 3 {
		s.Reply(message, i18n.T(lang, "schedule_usage"))
		return
	}

	taskType := parts[0]
	timeStr := parts[1]
	url := parts[2]
	var fileName string
	if len(parts) > 3 {
		fileName = strings.Join(parts[3:], " ")
	}

	validTypes := map[string]bool{"mirror": true, "zip": true, "unzip": true, "mp3": true, "audio": true, "video": true}
	if !validTypes[taskType] {
		s.Reply(message, i18n.T(lang, "schedule_invalid_type", utils.EscapeMarkdownV2(taskType)))
		return
	}

	parsedTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		s.Reply(message, i18n.T(lang, "schedule_invalid_time"))
		return
	}

	now := time.Now()
	scheduledAt := time.Date(now.Year(), now.Month(), now.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, now.Location())
	if scheduledAt.Before(now) {
		scheduledAt = scheduledAt.Add(24 * time.Hour)
	}

	st := domain.ScheduledTask{
		ID:          uuid.New().String(),
		TaskType:    taskType,
		URL:         url,
		FileName:    fileName,
		ChatID:      message.Chat.ID,
		UserID:      message.From.ID,
		Zip:         taskType == "zip",
		Unzip:       taskType == "unzip",
		Password:    "",
		Quality:     "",
		ScheduledAt: scheduledAt.Format(time.RFC3339),
		Status:      "pending",
	}

	if err := s.DB.SaveScheduled(context.Background(), st); err != nil {
		s.Reply(message, i18n.T(lang, "schedule_save_failed"))
		return
	}

	s.Reply(message, i18n.T(lang, "schedule_ok", scheduledAt.Format("15:04 02/01/2006")))
}
