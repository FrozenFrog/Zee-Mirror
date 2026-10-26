package basic

import (
	"strings"
	"time"
	"zee-mirror/internal/service"
	"zee-mirror/internal/userbot"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleJoin(s *service.BotService, m *tgbotapi.Message, args string) {
	lang := senderLang(s, m)

	if args == "" {
		reply := tgbotapi.NewMessage(m.Chat.ID, i18n.T(lang, "join_usage"))
		reply.ParseMode = tgbotapi.ModeMarkdownV2
		_, _ = s.Bot.Send(reply)
		return
	}

	var hash string
	switch {
	case strings.Contains(args, "+"):
		parts := strings.Split(args, "+")
		if len(parts) > 1 {
			hash = parts[1]
		}
	case strings.Contains(args, "joinchat/"):
		parts := strings.Split(args, "joinchat/")
		if len(parts) > 1 {
			hash = parts[1]
		}
	default:
		hash = args
	}

	hash = strings.TrimSpace(hash)
	hash = strings.Split(hash, "?")[0]
	hash = strings.Split(hash, "/")[0]

	if hash == "" {
		msg := tgbotapi.NewMessage(m.Chat.ID, i18n.T(lang, "join_hash_error"))
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		msg.ReplyParameters.MessageID = m.MessageID
		s.AutoDeleteMessage(m.Chat.ID, m.MessageID, 30*time.Second)
		_, _ = s.Bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, i18n.T(lang, "join_trying", hash))
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	sentMsg, err := s.Bot.Send(msg)
	if err != nil {
		msg = tgbotapi.NewMessage(m.Chat.ID, i18n.T(lang, "join_trying_plain", hash))
		sentMsg, _ = s.Bot.Send(msg)
	}

	ub := userbot.GetInstance(s.Config)
	result, err := ub.Join(hash)

	text := ""
	switch {
	case err != nil:
		text = i18n.T(lang, "join_failed", utils.EscapeMarkdownV2(i18n.Message(lang, err)))
	case result.AlreadyJoined:
		text = i18n.T(lang, "join_already")
	default:
		text = i18n.T(lang, "join_success", utils.EscapeMarkdownV2(result.Message(lang)))
	}

	if sentMsg.MessageID != 0 {
		edit := tgbotapi.NewEditMessageText(m.Chat.ID, sentMsg.MessageID, text)
		edit.ParseMode = tgbotapi.ModeMarkdownV2
		_, _ = s.Bot.Send(edit)
	} else {
		finalMsg := tgbotapi.NewMessage(m.Chat.ID, text)
		finalMsg.ParseMode = tgbotapi.ModeMarkdownV2
		_, _ = s.Bot.Send(finalMsg)
	}
}
