package basic

import (
	"strings"

	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleNotify registers (or clears with "off") a webhook URL that gets a POST
// whenever one of the user's tasks reaches a terminal state.
func HandleNotify(s *service.BotService, m *tgbotapi.Message, args string) {
	if m.From == nil {
		return
	}
	lang := s.GetUserLanguage(m.From.ID)
	arg := strings.TrimSpace(args)
	if arg == "" {
		notifyReply(s, m, i18n.T(lang, "notify_cmd_usage_plain"))
		return
	}

	if strings.EqualFold(arg, "off") {
		n := s.SetUserNotify(m.From.ID, "")
		notifyReply(s, m, i18n.T(lang, "notify_cmd_off_plain", n))
		return
	}

	if !service.ValidNotifyURL(arg) {
		notifyReply(s, m, i18n.T(lang, "notify_cmd_invalid_url_plain"))
		return
	}

	n := s.SetUserNotify(m.From.ID, arg)
	text := i18n.T(lang, "notify_cmd_saved", n, utils.EscapeMarkdownV2Code(arg))
	detail := tgbotapi.NewMessage(m.Chat.ID, text)
	detail.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(detail)
}

func notifyReply(s *service.BotService, m *tgbotapi.Message, text string) {
	_, _ = s.Bot.Send(tgbotapi.NewMessage(m.Chat.ID, text))
}
