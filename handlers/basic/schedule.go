package basic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"zee-mirror/internal/domain"
	"zee-mirror/internal/service"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

func HandleSchedule(s *service.BotService, message *tgbotapi.Message, args string) {
	parts := strings.Fields(args)
	if len(parts) < 3 {
		s.Reply(message, "Cách dùng: `/schedule <type> <HH:MM> <url> [filename]`\nLoại: mirror, zip, unzip, mp3, audio, video")
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
		s.Reply(message, fmt.Sprintf("Loại không hợp lệ: %s\\. Hợp lệ: mirror, zip, unzip, mp3, audio, video", utils.EscapeMarkdownV2(taskType)))
		return
	}

	parsedTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		s.Reply(message, "Định dạng giờ không hợp lệ\\. Dùng HH:MM \\(24 giờ\\)")
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
		s.Reply(message, "Không thể lên lịch Task")
		return
	}

	s.Reply(message, fmt.Sprintf("Đã lên lịch Task lúc %s", scheduledAt.Format("15:04 02/01/2006")))
}
