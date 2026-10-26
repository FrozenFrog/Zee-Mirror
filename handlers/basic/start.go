package basic

import (
	"log/slog"
	"time"
	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	ActionBack = "back"
	CmdClose   = "close"
)

// senderLang returns the language of the user who sent m, or the default when
// the message has no sender (channel posts).
func senderLang(s *service.BotService, m *tgbotapi.Message) string {
	if m == nil || m.From == nil {
		return i18n.DefaultLang
	}
	return s.GetUserLanguage(m.From.ID)
}

func StartHandler(s *service.BotService, message *tgbotapi.Message) {
	lang := s.GetUserLanguage(message.From.ID)
	userName := message.From.FirstName
	if userName == "" {
		userName = i18n.T(lang, "basic_name_plain")
	}

	welcomeText := service.GetWelcomeMessage(lang, userName)
	keyboard := service.GetStartKeyboard(lang)

	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	if len(keyboard.InlineKeyboard) > 0 {
		msg.ReplyMarkup = keyboard
	}

	if sentMsg, err := s.Bot.Send(msg); err != nil {
		slog.Error("Error sending welcome message", "error", err)
	} else {
		s.AutoDeleteMessage(message.Chat.ID, sentMsg.MessageID, 60*time.Second)
	}
}

func HelpHandler(s *service.BotService, message *tgbotapi.Message) {
	lang := s.GetUserLanguage(message.From.ID)
	helpText := service.GetHelpMainText(lang)
	keyboard := service.GetHelpKeyboard(lang)

	msg := tgbotapi.NewMessage(message.Chat.ID, helpText)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	if len(keyboard.InlineKeyboard) > 0 {
		msg.ReplyMarkup = keyboard
	}

	if sentMsg, err := s.Bot.Send(msg); err != nil {
		slog.Error("Error sending help message", "error", err)
		msg.ParseMode = ""
		msg.Text = i18n.T(lang, "basic_help_fallback_plain")
		_, _ = s.Bot.Send(msg)
	} else {
		s.AutoDeleteMessage(message.Chat.ID, sentMsg.MessageID, 60*time.Second)
	}
}
