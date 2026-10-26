package file

import (
	"context"
	"log/slog"
	"os/exec"
	"strings"
	"time"
	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleDriveCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) < 2 {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		return
	}

	action := parts[1]
	switch action {
	case "cd", "c":
		handleCDCallback(s, callback, parts)
	case "home", "h":
		handleHomeCallback(s, callback)
	case "info", "i":
		handleInfoCallback(s, callback, parts)
	case CmdClose, "x":
		handleCloseCallback(s, callback)
	case "rm":
		handleAskDeleteCallback(s, callback, parts)
	case "confirm_delete", "df":
		handleConfirmDeleteCallback(s, callback, parts)
	case "cancel_delete", "xf":
		handleCancelDeleteCallback(s, callback)
	}
}

func resolveCallbackPath(s *service.BotService, parts []string) string {
	if len(parts) < 3 {
		return ""
	}
	fullPath := strings.Join(parts[2:], ":")
	if strings.HasPrefix(fullPath, "id:") {
		id := strings.TrimPrefix(fullPath, "id:")
		if cached, ok := s.GetPath(id); ok {
			return cached
		}
	}
	return fullPath
}

func handleCDCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, parts []string) {
	fullPath := resolveCallbackPath(s, parts)
	msg := &tgbotapi.Message{
		Chat: callback.Message.Chat,
		From: callback.From,
	}
	HandleDriveList(s, msg, fullPath, callback.Message.MessageID)
	lang := s.GetUserLanguage(callback.From.ID)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "drive_toast_opening")))
}

func handleHomeCallback(s *service.BotService, callback *tgbotapi.CallbackQuery) {
	msg := &tgbotapi.Message{
		Chat: callback.Message.Chat,
		From: callback.From,
	}
	HandleDriveList(s, msg, "", callback.Message.MessageID)
	lang := s.GetUserLanguage(callback.From.ID)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "drive_toast_home")))
}

func handleInfoCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, parts []string) {
	fullPath := resolveCallbackPath(s, parts)
	lang := s.GetUserLanguage(callback.From.ID)
	handleDriveFileInfoDetailed(s, lang, callback.Message.Chat.ID, callback.Message.MessageID, fullPath)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
}

func handleCloseCallback(s *service.BotService, callback *tgbotapi.CallbackQuery) {
	deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
	_, _ = s.Bot.Request(deleteMsg)
	lang := s.GetUserLanguage(callback.From.ID)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "drive_toast_closed")))
}

func handleConfirmDeleteCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, parts []string) {
	fullPath := resolveCallbackPath(s, parts)
	executeDelete(s, callback, fullPath)
}

func handleAskDeleteCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, parts []string) {
	fullPath := resolveCallbackPath(s, parts)
	lang := s.GetUserLanguage(callback.From.ID)
	text := i18n.T(lang, "drive_rm_confirm", utils.EscapeMarkdownV2Code(fullPath))
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, text, deleteConfirmKeyboard(s, lang, fullPath))
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
}

func handleCancelDeleteCallback(s *service.BotService, callback *tgbotapi.CallbackQuery) {
	deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
	_, _ = s.Bot.Request(deleteMsg)
	lang := s.GetUserLanguage(callback.From.ID)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "drive_toast_cancelled")))
}

func executeDelete(s *service.BotService, callback *tgbotapi.CallbackQuery, fileName string) {
	lang := s.GetUserLanguage(callback.From.ID)
	if !s.IsAdmin(callback.From.ID) {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "drive_toast_admin_only")))
		return
	}
	if fileName == "" || fileName == "/" || fileName == "." {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "drive_toast_bad_path")))
		return
	}

	targetPath := s.TaskManager.RcloneDest + "/" + fileName
	slog.Info("Attempting to delete drive file", "path", targetPath)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	configPath := s.TaskManager.ConfigDir + "/rclone.conf"

	cmd := exec.CommandContext(ctx, "rclone", "deletefile", targetPath, "--config", configPath)
	_, err := cmd.CombinedOutput()

	if err != nil {
		slog.Warn("deletefile failed, trying purge", "error", err, "path", targetPath)

		cmdPurge := exec.CommandContext(ctx, "rclone", "purge", targetPath, "--config", configPath)
		outputPurge, errPurge := cmdPurge.CombinedOutput()

		if errPurge != nil {
			slog.Error("purge also failed", "error", errPurge, "path", targetPath)
			_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "drive_toast_delete_failed")))

			errMsg := i18n.T(lang, "drive_rm_failed",
				utils.EscapeMarkdownV2(fileName),
				utils.EscapeMarkdownV2(errPurge.Error()),
				utils.EscapeMarkdownV2(utils.TruncateString(string(outputPurge), 100)))

			s.Reply(callback.Message, errMsg)
			return
		}
		slog.Info("Successfully purged path", "path", targetPath)
	} else {
		slog.Info("Successfully deleted file", "path", targetPath)
	}

	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "drive_toast_deleted")))

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID,
		i18n.T(lang, "drive_rm_done", utils.EscapeMarkdownV2(fileName)))
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2

	_, _ = s.Bot.Send(editMsg)
}
