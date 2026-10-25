package basic

import (
	"fmt"
	"strings"
	"time"
	"zee-mirror/internal/service"
	"zee-mirror/internal/userbot"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleJoin(s *service.BotService, m *tgbotapi.Message, args string) {
	if args == "" {
		reply := tgbotapi.NewMessage(m.Chat.ID, "⚠️ *Sai định dạng*\n\nDùng: `/join <link invite>`\nVí dụ: `/join https://t.me/+FnboT0dqFiozNzA1`")
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
		msg := tgbotapi.NewMessage(m.Chat.ID, "❌ *Lỗi:* Không lấy được hash mời từ liên kết")
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		msg.ReplyParameters.MessageID = m.MessageID
		s.AutoDeleteMessage(m.Chat.ID, m.MessageID, 30*time.Second)
		_, _ = s.Bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("🔄 *Đang thử tham gia channel\\.\\.\\.*\nHash: `%s`", hash))
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	sentMsg, err := s.Bot.Send(msg)
	if err != nil {
		msg = tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("🔄 Đang thử tham gia channel...\nHash: %s", hash))
		sentMsg, _ = s.Bot.Send(msg)
	}

	ub := userbot.GetInstance(s.Config)
	result, err := ub.JoinChat(hash)

	text := ""
	if err != nil {
		if strings.Contains(err.Error(), "USER_ALREADY_PARTICIPANT") {
			text = "✅ *Đã tham gia\\!* Bot đã ở trong channel này rồi\\."
		} else {
			text = fmt.Sprintf("❌ *Tham gia thất bại*\nLỗi: `%s`", utils.EscapeMarkdownV2(err.Error()))
		}
	} else {
		if result == "Already joined" {
			text = "✅ *Đã tham gia\\!* Bot đã ở trong channel này rồi\\."
		} else {
			text = fmt.Sprintf("✅ *Thành công\\!* %s", utils.EscapeMarkdownV2(result))
		}
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
