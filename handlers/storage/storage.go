package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Provider struct {
	Name       string
	RemoteName string
	Type       string
	Icon       string
}

func GetAvailableStorages(s *service.BotService) ([]Provider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	configPath := s.TaskManager.ConfigDir + "/rclone.conf"

	cmd := execCommand(ctx, "rclone", "listremotes", "--config", configPath, "--long")
	output, err := cmd.Output()

	if err != nil {
		return nil, i18n.WrapError(err, "storage_err_list_remotes", err)
	}

	var providers []Provider
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := strings.TrimSuffix(parts[0], ":")
			remoteType := parts[1]

			provider := Provider{
				Name:       name,
				RemoteName: parts[0],
				Type:       remoteType,
				Icon:       getStorageIcon(remoteType),
			}
			providers = append(providers, provider)
		}
	}

	return providers, nil
}

func getStorageIcon(storageType string) string {
	switch strings.ToLower(storageType) {
	case "drive":
		return "📁"
	case "onedrive":
		return "☁️"
	case "dropbox":
		return "📦"
	case "mega":
		return "🔷"
	case "s3", "b2":
		return "🪣"
	case "ftp", "sftp":
		return "🖥️"
	case "webdav":
		return "🌐"
	default:
		return "💾"
	}
}

func HandleStorages(s *service.BotService, message *tgbotapi.Message) {
	if !s.CanUse(message) {
		return
	}

	lang := s.GetUserLanguage(message.From.ID)
	providers, err := GetAvailableStorages(s)
	if err != nil {
		s.Reply(message, i18n.T(lang, "storage_list_failed", utils.EscapeMarkdownV2(i18n.Message(lang, err))))
		return
	}

	if len(providers) == 0 {
		s.Reply(message, i18n.T(lang, "storage_none"))
		return
	}

	text, keyboard := storageListMessage(s, lang, providers)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.ReplyMarkup = keyboard
	_, _ = s.Bot.Send(msg)
}

func storageListMessage(s *service.BotService, lang string, providers []Provider) (string, tgbotapi.InlineKeyboardMarkup) {
	var text strings.Builder
	text.WriteString(i18n.T(lang, "storage_list_title") + "\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, p := range providers {
		fmt.Fprintf(&text, "%s *%s* \\(%s\\)\n", p.Icon, utils.EscapeMarkdownV2(p.Name), utils.EscapeMarkdownV2(p.Type))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(p.Icon+" "+p.Name, "storage:select:"+p.Name),
		))
	}

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━\n")
	text.WriteString(i18n.T(lang, "storage_current", utils.EscapeMarkdownV2Code(s.TaskManager.RcloneDest)) + "\n\n")
	text.WriteString(i18n.T(lang, "storage_pick"))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_close"), "storage:"+service.CmdClose),
	))
	return text.String(), tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// storageBackKeyboard returns to the actions of one storage.
func storageBackKeyboard(lang, storageName string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_back"), "storage:select:"+storageName),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_close"), "storage:"+service.CmdClose),
	))
}

func HandleStorageCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) < 2 {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		return
	}

	action := parts[1]
	storageName := ""
	if len(parts) > 2 {
		storageName = parts[2]
	}

	lang := s.GetUserLanguage(callback.From.ID)

	switch action {
	case "list":
		providers, err := GetAvailableStorages(s)
		if err != nil {
			_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "storage_toast_error")))
			return
		}
		text, keyboard := storageListMessage(s, lang, providers)
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, text, keyboard)
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
		_, _ = s.Bot.Send(editMsg)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
	case "select":
		handleStorageSelect(s, lang, callback, storageName)
	case "browse":
		handleStorageBrowse(s, lang, callback, storageName, "")
	case "info":
		handleStorageInfo(s, lang, callback, storageName)
	case service.CmdClose:
		deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
		_, _ = s.Bot.Request(deleteMsg)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "drive_toast_closed")))
	}
}

func handleStorageSelect(s *service.BotService, lang string, callback *tgbotapi.CallbackQuery, storageName string) {
	text := i18n.T(lang, "storage_select_text", utils.EscapeMarkdownV2(storageName))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "storage_btn_browse"), "storage:browse:"+storageName),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "storage_btn_info"), "storage:info:"+storageName),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_back"), "storage:list"),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_close"), "storage:"+service.CmdClose),
		),
	)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, text, keyboard)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
}

func handleStorageBrowse(s *service.BotService, lang string, callback *tgbotapi.CallbackQuery, storageName, path string) {
	remotePath := storageName + ":"
	if path != "" {
		remotePath = storageName + ":" + path
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	configPath := s.TaskManager.ConfigDir + "/rclone.conf"

	cmd := execCommand(ctx, "rclone", "lsjson", remotePath, "--config", configPath)
	output, err := cmd.Output()

	if err != nil {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "storage_toast_browse_fail")))
		return
	}

	type DriveFile struct {
		Name  string `json:"Name"`
		IsDir bool   `json:"IsDir"`
	}

	var files []DriveFile
	_ = json.Unmarshal(output, &files)

	var text strings.Builder
	fmt.Fprintf(&text, "📂 *%s*\n", utils.EscapeMarkdownV2(remotePath))
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for i, f := range files {
		if i >= 15 {
			text.WriteString("\n" + i18n.T(lang, "storage_browse_more", len(files)-15))
			break
		}
		icon := "📄"
		if f.IsDir {
			icon = "📁"
		}
		fmt.Fprintf(&text, "%s `%s`\n", icon, utils.EscapeMarkdownV2Code(utils.TruncateString(f.Name, 40)))
	}

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, text.String(), storageBackKeyboard(lang, storageName))
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
}

func handleStorageInfo(s *service.BotService, lang string, callback *tgbotapi.CallbackQuery, storageName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	configPath := s.TaskManager.ConfigDir + "/rclone.conf"

	cmd := execCommand(ctx, "rclone", "about", storageName+":", "--config", configPath, "--json")
	output, err := cmd.Output()

	var text strings.Builder
	text.WriteString(i18n.T(lang, "storage_info_title", utils.EscapeMarkdownV2(storageName)) + "\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	if err != nil {
		text.WriteString(i18n.T(lang, "storage_info_none") + "\n")
	} else {
		var info map[string]interface{}
		if jsonErr := json.Unmarshal(output, &info); jsonErr == nil {
			if total, ok := info["total"].(float64); ok {
				text.WriteString(i18n.T(lang, "storage_info_total", utils.EscapeMarkdownV2Code(utils.FormatBytes(int64(total)))) + "\n")
			}
			if used, ok := info["used"].(float64); ok {
				text.WriteString(i18n.T(lang, "storage_info_used", utils.EscapeMarkdownV2Code(utils.FormatBytes(int64(used)))) + "\n")
			}
			if free, ok := info["free"].(float64); ok {
				text.WriteString(i18n.T(lang, "storage_info_free", utils.EscapeMarkdownV2Code(utils.FormatBytes(int64(free)))) + "\n")
			}
		}
	}

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, text.String(), storageBackKeyboard(lang, storageName))
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
}

func HandleSetStorage(s *service.BotService, message *tgbotapi.Message, args string) {
	lang := s.GetUserLanguage(message.From.ID)
	if !s.IsAdmin(message.From.ID) {
		s.Reply(message, i18n.T(lang, "storage_set_denied"))
		return
	}

	if args == "" {
		s.Reply(message, i18n.T(lang, "storage_usage_set"))
		return
	}

	remoteName := strings.Split(args, ":")[0]
	providers, err := GetAvailableStorages(s)
	if err != nil {
		s.Reply(message, i18n.T(lang, "storage_set_verify_failed"))
		return
	}

	found := false
	for _, p := range providers {
		if p.Name == remoteName {
			found = true
			break
		}
	}

	if !found {
		s.Reply(message, i18n.T(lang, "storage_set_not_found", utils.EscapeMarkdownV2(remoteName)))
		return
	}

	s.TaskManager.RcloneDest = args
	ctx := context.Background()
	_ = s.DB.Set(ctx, "rclone_dest", args)

	s.Reply(message, i18n.T(lang, "storage_set_ok", utils.EscapeMarkdownV2Code(args)))
}
