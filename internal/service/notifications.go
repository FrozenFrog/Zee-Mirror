package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type NotificationService struct {
	Bot            *tgbotapi.BotAPI
	AlertChannelID int64
	OwnerID        int64
	mu             sync.RWMutex
	DiskThreshold  float64
}

func NewNotificationService(bot *tgbotapi.BotAPI, alertChannelID, ownerID int64) *NotificationService {
	return &NotificationService{
		Bot:            bot,
		AlertChannelID: alertChannelID,
		OwnerID:        ownerID,
		DiskThreshold:  85.0,
	}
}

func (n *NotificationService) SetAlertChannel(channelID int64) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.AlertChannelID = channelID
}

func (n *NotificationService) SendAlert(alertType, title, message string) {
	targetID := n.AlertChannelID
	if targetID == 0 {
		targetID = n.OwnerID
	}

	if targetID == 0 {
		return
	}

	// Alerts go to a channel or the owner, not to one user: DefaultLang.
	text := i18n.T(i18n.DefaultLang, "notify_alert",
		utils.EscapeMarkdownV2(title),
		alertType,
		utils.EscapeMarkdownV2(title),
		utils.EscapeMarkdownV2(message),
		utils.EscapeMarkdownV2(time.Now().Format("02 Jan 2006 15:04:05")),
	)

	msg := tgbotapi.NewMessage(targetID, text)
	msg.ParseMode = "MarkdownV2"
	if _, err := n.Bot.Send(msg); err != nil {
		slog.Error("Failed to send alert", "error", err, "targetID", targetID)
	}
}

func (n *NotificationService) AlertDiskUsage(usage float64, path string) {
	if usage < n.DiskThreshold {
		return
	}

	message := i18n.T(i18n.DefaultLang, "notify_disk_msg", usage, path)
	n.SendAlert("💾", i18n.T(i18n.DefaultLang, "notify_disk_title"), message)
}

func (n *NotificationService) AlertSystemHealth(cpuUsage, memUsage float64) {
	var alerts []string

	if cpuUsage > 90 {
		alerts = append(alerts, fmt.Sprintf("CPU: %.1f%%", cpuUsage))
	}
	if memUsage > 90 {
		alerts = append(alerts, i18n.T(i18n.DefaultLang, "notify_mem", memUsage))
	}

	if len(alerts) > 0 {
		message := i18n.T(i18n.DefaultLang, "notify_health_msg", strings.Join(alerts, "\n"))
		n.SendAlert("🖥️", i18n.T(i18n.DefaultLang, "notify_health_title"), message)
	}
}

func (s *BotService) HandleSetAlertChannel(message *tgbotapi.Message, args string) {
	lang := s.GetUserLanguage(message.From.ID)
	if message.From.ID != s.Config.OwnerID {
		s.reply(message, i18n.T(lang, "notify_owner_only"))
		return
	}

	channelID := message.Chat.ID

	if args != "" {
		var id int64
		if _, err := fmt.Sscanf(args, "%d", &id); err == nil {
			channelID = id
		}
	}

	if s.Notifications != nil {
		s.Notifications.SetAlertChannel(channelID)
	}

	ctx := context.Background()
	_ = s.DB.Set(ctx, "alert_channel_id", fmt.Sprintf("%d", channelID))

	s.reply(message, i18n.T(lang, "notify_channel_set", channelID))
}
