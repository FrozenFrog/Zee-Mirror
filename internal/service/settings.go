package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Settings struct {
	DefaultMode        string
	YTDLPQuality       string
	Mu                 sync.RWMutex
	AutoDeleteMessages bool
}

func NewSettings() *Settings {
	return &Settings{
		AutoDeleteMessages: true,
		DefaultMode:        string(TypeMirror),
		YTDLPQuality:       "1080p",
	}
}

func (s *BotService) HandleSettings(message *tgbotapi.Message) {
	text := s.formatSettingsMessage(message.From.ID)
	keyboard := s.getSettingsKeyboard(message.From.ID)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = MarkdownV2
	msg.ReplyMarkup = keyboard
	sentMsg, err := s.Bot.Send(msg)
	if err != nil {
		slog.Error("Failed to send settings menu", "error", err)
		return
	}
	s.AutoDeleteMessage(message.Chat.ID, sentMsg.MessageID, 60*time.Second)
}

func (s *BotService) HandleLanguage(message *tgbotapi.Message) {
	lang := s.GetUserLanguage(message.From.ID)
	text := i18n.T(lang, "settings_select_language")

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = MarkdownV2
	msg.ReplyMarkup = languageKeyboard()
	msg.ReplyParameters.MessageID = message.MessageID
	if _, err := s.Bot.Send(msg); err != nil {
		slog.Error("Failed to send language menu", "error", err)
	}
}

func (s *BotService) HandleSettingsCallback(callback *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) < 2 {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		return
	}

	ctx := context.Background()
	action := parts[1]
	s.Settings.Mu.Lock()
	if val, err := s.DB.Get(ctx, "auto_delete_messages"); err == nil {
		s.Settings.AutoDeleteMessages = (val == "true")
	}
	if val, err := s.DB.Get(ctx, "default_mode"); err == nil {
		s.Settings.DefaultMode = val
	}
	s.Settings.Mu.Unlock()

	switch action {
	case "auto_delete":
		s.Settings.Mu.Lock()
		s.Settings.AutoDeleteMessages = !s.Settings.AutoDeleteMessages
		status := s.Settings.AutoDeleteMessages
		s.Settings.Mu.Unlock()

		_ = s.DB.Set(ctx, "auto_delete_messages", fmt.Sprintf("%v", status))

		if status {
			_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "✅ Tự động xóa: BẬT"))
		} else {
			_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "❌ Tự động xóa: TẮT"))
		}

	case "default_mirror":
		s.Settings.Mu.Lock()
		s.Settings.DefaultMode = string(TypeMirror)
		s.Settings.Mu.Unlock()
		_ = s.DB.Set(ctx, "default_mode", string(TypeMirror))
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "📥 Mặc định: Mirror"))

	case "default_leech":
		s.Settings.Mu.Lock()
		s.Settings.DefaultMode = string(TypeLeech)
		s.Settings.Mu.Unlock()
		_ = s.DB.Set(ctx, "default_mode", string(TypeLeech))
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "🔗 Mặc định: Leech"))

	case "quality_720":
		s.Settings.Mu.Lock()
		s.Settings.YTDLPQuality = "720p"
		s.Settings.Mu.Unlock()
		_ = s.DB.Set(ctx, "ytdlp_quality", "720p")
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "📺 Chất lượng: 720p"))

	case "quality_1080":
		s.Settings.Mu.Lock()
		s.Settings.YTDLPQuality = "1080p"
		s.Settings.Mu.Unlock()
		_ = s.DB.Set(ctx, "ytdlp_quality", "1080p")
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "📺 Chất lượng: 1080p"))

	case "quality_2160":
		s.Settings.Mu.Lock()
		s.Settings.YTDLPQuality = "2160p"
		s.Settings.Mu.Unlock()
		_ = s.DB.Set(ctx, "ytdlp_quality", "2160p")
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "📺 Chất lượng: 4K (2160p)"))

	case "lang_vi":
		_ = s.DB.SetLanguage(ctx, callback.From.ID, "vi")
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "🇻🇳 Ngôn ngữ: Tiếng Việt"))

	case "lang_id":
		_ = s.DB.SetLanguage(ctx, callback.From.ID, "id")
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "🇮🇩 Bahasa: Indonesia"))

	case "lang_en":
		_ = s.DB.SetLanguage(ctx, callback.From.ID, "en")
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "🇺🇸 Language: English"))

	case "lang_ja":
		_ = s.DB.SetLanguage(ctx, callback.From.ID, "ja")
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "🇯🇵 言語: 日本語"))

	case "lang_zh":
		_ = s.DB.SetLanguage(ctx, callback.From.ID, "zh")
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "🇨🇳 语言: 简体中文"))

	default:
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
	}

	text := s.formatSettingsMessage(callback.From.ID)
	keyboard := s.getSettingsKeyboard(callback.From.ID)

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	editMsg.ParseMode = MarkdownV2
	editMsg.ReplyMarkup = &keyboard
	_, _ = s.Bot.Send(editMsg)
}

func (s *BotService) formatSettingsMessage(userID int64) string {
	s.Settings.Mu.RLock()
	defer s.Settings.Mu.RUnlock()

	lang := s.GetUserLanguage(userID)

	autoDeleteStatus := "❌ TẮT"
	if s.Settings.AutoDeleteMessages {
		autoDeleteStatus = "✅ BẬT"
	}

	defaultModeEmoji := "📥"
	if s.Settings.DefaultMode == ModeLeech {
		defaultModeEmoji = "🔗"
	}

	langFlag := "🇻🇳"
	langName := "Tiếng Việt"
	switch lang {
	case "id":
		langFlag = "🇮🇩"
		langName = "Indonesia"
	case "en":
		langFlag = "🇺🇸"
		langName = "English"
	case "ja":
		langFlag = "🇯🇵"
		langName = "日本語"
	case "zh":
		langFlag = "🇨🇳"
		langName = "简体中文"
	}

	cookiesStatus := i18n.T(lang, "settings_cookies_missing")
	if _, err := os.Stat(filepath.Join(s.Config.ConfigDir, "cookies.txt")); err == nil {
		cookiesStatus = i18n.T(lang, "settings_cookies_installed")
	}

	return fmt.Sprintf("⚙️ *%s*\n\n"+
		"*%s:* %s\n"+
		"_%s_\n\n"+
		"*%s:* %s %s\n"+
		"_%s_\n\n"+
		"*%s:* 📺 %s\n"+
		"_%s_\n\n"+
		"*%s:* %s\n"+
		"_%s_\n\n"+
		"*%s:* %s %s\n\n"+
		"%s",
		i18n.T(lang, "settings_title"),
		i18n.T(lang, "settings_auto_delete"), autoDeleteStatus,
		i18n.T(lang, "settings_auto_delete_desc"),
		i18n.T(lang, "settings_default_mode"), defaultModeEmoji, utils.EscapeMarkdownV2(s.Settings.DefaultMode),
		i18n.T(lang, "settings_default_mode_desc"),
		i18n.T(lang, "settings_ytdlp_quality"), utils.EscapeMarkdownV2(s.Settings.YTDLPQuality),
		i18n.T(lang, "settings_ytdlp_quality_desc"),
		i18n.T(lang, "settings_cookies"), cookiesStatus,
		i18n.T(lang, "settings_cookies_desc"),
		i18n.T(lang, "settings_language"), langFlag, langName,
		i18n.T(lang, "settings_footer"),
	)
}

// Callback data is "settings:<action>", handled by HandleSettingsCallback.
func (s *BotService) getSettingsKeyboard(userID int64) tgbotapi.InlineKeyboardMarkup {
	lang := s.GetUserLanguage(userID)

	s.Settings.Mu.RLock()
	autoDelete := "❌ " + i18n.T(lang, "settings_auto_delete")
	if s.Settings.AutoDeleteMessages {
		autoDelete = "✅ " + i18n.T(lang, "settings_auto_delete")
	}
	s.Settings.Mu.RUnlock()

	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(autoDelete, "settings:auto_delete"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📥 "+i18n.T(lang, "settings_set_mirror"), "settings:default_mirror"),
			tgbotapi.NewInlineKeyboardButtonData("📤 "+i18n.T(lang, "settings_set_leech"), "settings:default_leech"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📺 720p", "settings:quality_720"),
			tgbotapi.NewInlineKeyboardButtonData("📺 1080p", "settings:quality_1080"),
			tgbotapi.NewInlineKeyboardButtonData("📺 4K", "settings:quality_2160"),
		),
	}
	rows = append(rows, languageKeyboard().InlineKeyboard...)
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func languageKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇻🇳 Tiếng Việt", "settings:lang_vi"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇺🇸 English", "settings:lang_en"),
			tgbotapi.NewInlineKeyboardButtonData("🇮🇩 Indonesia", "settings:lang_id"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇯🇵 日本語", "settings:lang_ja"),
			tgbotapi.NewInlineKeyboardButtonData("🇨🇳 简体中文", "settings:lang_zh"),
		),
	)
}
