package basic

import (
	"fmt"
	"strings"

	"zee-mirror/internal/service"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleNotify registers (or clears with "off") a webhook URL that gets a POST
// whenever one of the user's tasks reaches a terminal state.
func HandleNotify(s *service.BotService, m *tgbotapi.Message, args string) {
	if m.From == nil {
		return
	}
	arg := strings.TrimSpace(args)
	if arg == "" {
		notifyReply(s, m, "🔔 WEBHOOK THÔNG BÁO\n\nDùng: /notify <webhook_url>\nXóa: /notify off\n\nWebhook sẽ nhận POST JSON khi Task của bạn hoàn thành/thất bại.")
		return
	}

	if strings.EqualFold(arg, "off") {
		n := s.SetUserNotify(m.From.ID, "")
		notifyReply(s, m, fmt.Sprintf("🔕 Đã tắt webhook thông báo (%d Task đang chạy)", n))
		return
	}

	if !service.ValidNotifyURL(arg) {
		notifyReply(s, m, "❌ URL không hợp lệ — phải là http/https")
		return
	}

	n := s.SetUserNotify(m.From.ID, arg)
	text := fmt.Sprintf("✅ Đã lưu webhook cho %d Task đang chạy\n\nURL: `%s`\nWebhook sẽ nhận POST JSON khi Task hoàn thành/thất bại\\.",
		n, utils.EscapeMarkdownV2Code(arg))
	detail := tgbotapi.NewMessage(m.Chat.ID, text)
	detail.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(detail)
}

func notifyReply(s *service.BotService, m *tgbotapi.Message, text string) {
	_, _ = s.Bot.Send(tgbotapi.NewMessage(m.Chat.ID, text))
}
