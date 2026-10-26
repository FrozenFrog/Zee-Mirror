package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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

	lang := s.GetUserLanguage(callback.From.ID)
	toast := func(text string) {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, text))
	}

	switch action {
	case "auto_delete":
		s.Settings.Mu.Lock()
		s.Settings.AutoDeleteMessages = !s.Settings.AutoDeleteMessages
		status := s.Settings.AutoDeleteMessages
		s.Settings.Mu.Unlock()

		_ = s.DB.Set(ctx, "auto_delete_messages", fmt.Sprintf("%v", status))

		if status {
			toast(i18n.T(lang, "settings_toast_auto_delete_on"))
		} else {
			toast(i18n.T(lang, "settings_toast_auto_delete_off"))
		}

	case "default_mirror":
		s.Settings.Mu.Lock()
		s.Settings.DefaultMode = string(TypeMirror)
		s.Settings.Mu.Unlock()
		_ = s.DB.Set(ctx, "default_mode", string(TypeMirror))
		toast(i18n.T(lang, "settings_toast_default_mode", "📥", "Mirror"))

	case "default_leech":
		s.Settings.Mu.Lock()
		s.Settings.DefaultMode = string(TypeLeech)
		s.Settings.Mu.Unlock()
		_ = s.DB.Set(ctx, "default_mode", string(TypeLeech))
		toast(i18n.T(lang, "settings_toast_default_mode", "🔗", "Leech"))

	case "quality_720":
		s.Settings.Mu.Lock()
		s.Settings.YTDLPQuality = "720p"
		s.Settings.Mu.Unlock()
		_ = s.DB.Set(ctx, "ytdlp_quality", "720p")
		toast(i18n.T(lang, "settings_toast_quality", "720p"))

	case "quality_1080":
		s.Settings.Mu.Lock()
		s.Settings.YTDLPQuality = "1080p"
		s.Settings.Mu.Unlock()
		_ = s.DB.Set(ctx, "ytdlp_quality", "1080p")
		toast(i18n.T(lang, "settings_toast_quality", "1080p"))

	case "quality_2160":
		s.Settings.Mu.Lock()
		s.Settings.YTDLPQuality = "2160p"
		s.Settings.Mu.Unlock()
		_ = s.DB.Set(ctx, "ytdlp_quality", "2160p")
		toast(i18n.T(lang, "settings_toast_quality", "4K (2160p)"))

	default:
		l, ok := languageFromAction(action)
		if !ok {
			toast("")
			break
		}
		_ = s.DB.SetLanguage(ctx, callback.From.ID, l.Code)
		toast(languageSetToast(l))
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

	autoDeleteStatus := i18n.T(lang, "settings_off")
	if s.Settings.AutoDeleteMessages {
		autoDeleteStatus = i18n.T(lang, "settings_on")
	}

	defaultModeEmoji := "📥"
	if s.Settings.DefaultMode == ModeLeech {
		defaultModeEmoji = "🔗"
	}

	current := languageInfo(lang)
	langFlag, langName := current.Flag, utils.EscapeMarkdownV2(current.Name)

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

// languageKeyboard lists i18n.Languages with callback data
// "settings:lang_<code>": the first (default) language on its own row, the
// rest two per row. Names are native, so the picker is the same for everyone.
func languageKeyboard() tgbotapi.InlineKeyboardMarkup {
	button := func(l i18n.Language) tgbotapi.InlineKeyboardButton {
		return tgbotapi.NewInlineKeyboardButtonData(l.Flag+" "+l.Name, "settings:lang_"+l.Code)
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	if len(i18n.Languages) > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button(i18n.Languages[0])))
	}
	for i := 1; i < len(i18n.Languages); i += 2 {
		row := tgbotapi.NewInlineKeyboardRow(button(i18n.Languages[i]))
		if i+1 < len(i18n.Languages) {
			row = append(row, button(i18n.Languages[i+1]))
		}
		rows = append(rows, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// languageFromAction maps a "lang_<code>" settings action to its language.
func languageFromAction(action string) (i18n.Language, bool) {
	code, ok := strings.CutPrefix(action, "lang_")
	if !ok {
		return i18n.Language{}, false
	}
	for _, l := range i18n.Languages {
		if l.Code == code {
			return l, true
		}
	}
	return i18n.Language{}, false
}

// languageInfo describes code, falling back to DefaultLang.
func languageInfo(code string) i18n.Language {
	var def i18n.Language
	for _, l := range i18n.Languages {
		if l.Code == code {
			return l
		}
		if l.Code == i18n.DefaultLang {
			def = l
		}
	}
	return def
}

// languageSetToast confirms a language change, in the new language (plain
// text for a callback toast).
func languageSetToast(l i18n.Language) string {
	return i18n.T(l.Code, "lang_set_toast", l.Flag, l.Name)
}
