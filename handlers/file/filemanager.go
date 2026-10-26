package file

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	IconFolder = "📂"
	IconFile   = "📄"
	CmdClose   = "close"
)

func HandleDriveList(s *service.BotService, message *tgbotapi.Message, args string, editMessageID int) {
	if !s.CanUse(message) {
		return
	}

	lang := s.GetUserLanguage(message.From.ID)
	slog.Info("Drive list request", "userID", message.From.ID, "args", args, "editMsgID", editMessageID)

	fullPath, relPath := resolveDrivePath(s, args)
	slog.Debug("Listing drive path", "fullPath", fullPath)

	sent, err := sendLoadingStatus(s, lang, message.Chat.ID, editMessageID)
	if err != nil {
		slog.Error("Error preparing status message", "error", err)
		return
	}

	files, err := listDriveFiles(s, fullPath)
	if err != nil {
		handleDriveError(s, lang, message.Chat.ID, sent.MessageID, i18n.T(lang, "fm_err_load_list"), err)
		return
	}

	slog.Debug("Drive list found items", "count", len(files))

	if tryJumpToFileInfo(s, lang, message, sent.MessageID, relPath, files) {
		return
	}

	renderDriveList(s, lang, message.Chat.ID, sent.MessageID, relPath, files)
}

func resolveDrivePath(s *service.BotService, args string) (string, string) {
	basePath := strings.TrimSuffix(s.TaskManager.RcloneDest, "/")
	fullPath := basePath

	if args != "" {
		if strings.HasPrefix(args, "/") {
			remoteName := strings.Split(s.TaskManager.RcloneDest, ":")[0]
			fullPath = remoteName + ":" + args
		} else {
			fullPath = basePath + "/" + strings.TrimPrefix(args, "/")
		}
	}
	fullPath = strings.TrimSuffix(fullPath, "/")

	baseRel := ""
	if strings.Contains(s.TaskManager.RcloneDest, ":") {
		parts := strings.SplitN(s.TaskManager.RcloneDest, ":", 2)
		if len(parts) > 1 {
			baseRel = strings.Trim(parts[1], "/")
		}
	}

	relPath := ""
	if strings.Contains(fullPath, ":") {
		parts := strings.SplitN(fullPath, ":", 2)
		if len(parts) > 1 {
			fullPathAfterRemote := strings.Trim(parts[1], "/")
			if baseRel != "" && (fullPathAfterRemote == baseRel || strings.HasPrefix(fullPathAfterRemote, baseRel+"/")) {
				relPath = strings.TrimPrefix(fullPathAfterRemote, baseRel)
				relPath = strings.TrimPrefix(relPath, "/")
			} else {
				relPath = "/" + fullPathAfterRemote
			}
		}
	}
	return fullPath, relPath
}

func sendLoadingStatus(s *service.BotService, lang string, chatID int64, editMessageID int) (*tgbotapi.Message, error) {
	loadingText := i18n.T(lang, "fm_loading")
	if editMessageID != 0 {
		editMsg := tgbotapi.NewEditMessageText(chatID, editMessageID, loadingText)
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
		m, err := s.Bot.Send(editMsg)
		if err == nil {
			return &m, nil
		}
	}

	statusMsg := tgbotapi.NewMessage(chatID, loadingText)
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	m, err := s.Bot.Send(statusMsg)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func handleDriveError(s *service.BotService, lang string, chatID int64, messageID int, title string, err error) {
	slog.Error("Drive operation error", "title", title, "error", err)
	errorText := i18n.T(lang, "fm_error_fmt", utils.EscapeMarkdownV2(title), utils.EscapeMarkdownV2(i18n.Message(lang, err)))
	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, errorText)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
}

func tryJumpToFileInfo(s *service.BotService, lang string, message *tgbotapi.Message, messageID int, relPath string, files []service.DriveFile) bool {
	if len(files) == 1 && !files[0].IsDir {
		fileName := files[0].Name
		if relPath == fileName || strings.HasSuffix(relPath, "/"+fileName) {
			slog.Info("Target is a file, jumping to info view", "relPath", relPath, "messageID", messageID)
			handleDriveFileInfoDetailed(s, lang, message.Chat.ID, messageID, relPath)
			return true
		}
	}
	return false
}

func renderDriveList(s *service.BotService, lang string, chatID int64, messageID int, relPath string, files []service.DriveFile) {
	currentPathForUI := relPath
	if !strings.HasPrefix(currentPathForUI, "/") {
		currentPathForUI = "/" + currentPathForUI
	}

	text := formatDriveFileList(lang, currentPathForUI, files)
	keyboard := buildDriveNavigationKeyboard(s, lang, files, relPath)

	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	if len(keyboard.InlineKeyboard) > 0 {
		editMsg.ReplyMarkup = &keyboard
	}

	if _, err := s.Bot.Send(editMsg); err != nil {
		slog.Error("Error sending final list message", "error", err)
		handleDriveError(s, lang, chatID, messageID, i18n.T(lang, "fm_err_show_list"), err)
	}
}

func listDriveFiles(s *service.BotService, path string) ([]service.DriveFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	configPath := s.TaskManager.ConfigDir + "/rclone.conf"
	args := []string{
		"lsjson",
		path,
		"--config", configPath,
		"--no-modtime",
	}

	cmd := exec.CommandContext(ctx, "rclone", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, i18n.WrapError(err, "fm_err_rclone", utils.RcloneError(err))
	}

	var files []service.DriveFile
	if err := json.Unmarshal(output, &files); err != nil {
		return nil, i18n.WrapError(err, "fm_err_parse", err)
	}

	return files, nil
}

func formatDriveFileList(lang, path string, files []service.DriveFile) string {
	var text strings.Builder

	text.WriteString(i18n.T(lang, "fm_title") + "\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")
	text.WriteString(i18n.T(lang, "fm_path", utils.EscapeMarkdownV2Code(path)) + "\n\n")

	if len(files) == 0 {
		text.WriteString(i18n.T(lang, "fm_empty") + "\n")
	} else {
		folderCount := 0
		fileCount := 0

		var folders []string
		for _, f := range files {
			if f.IsDir {
				folderCount++
				folders = append(folders, fmt.Sprintf("%s `%s`", IconFolder, utils.EscapeMarkdownV2Code(utils.TruncateString(f.Name, 40))))
			}
		}

		if len(folders) > 0 {
			text.WriteString(i18n.T(lang, "fm_folders") + "\n")
			for _, f := range folders {
				text.WriteString(f + "\n")
			}
			text.WriteString("\n")
		}

		var fileList []string
		for _, f := range files {
			if !f.IsDir {
				fileCount++
				icon := getFileIcon(f.Name)
				size := utils.FormatBytes(f.Size)
				fileList = append(fileList, fmt.Sprintf("%s `%s` \\(%s\\)",
					icon,
					utils.EscapeMarkdownV2Code(utils.TruncateString(f.Name, 35)),
					utils.EscapeMarkdownV2(size)))
			}
		}

		if len(fileList) > 0 {
			text.WriteString(i18n.T(lang, "fm_files") + "\n")
			for _, f := range fileList {
				text.WriteString(f + "\n")
			}
			text.WriteString("\n")
		}

		text.WriteString(i18n.T(lang, "fm_summary") + "\n")
		text.WriteString(i18n.T(lang, "fm_count_folders", folderCount) + "\n")
		text.WriteString(i18n.T(lang, "fm_count_files", fileCount) + "\n")
	}

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")
	return text.String()
}

// driveCallbackData builds "dr:<action>:<path>", caching long paths behind an id
// to stay within Telegram's 64-byte callback_data limit.
func driveCallbackData(s *service.BotService, action, path string) string {
	data := "dr:" + action + ":" + path
	if len(data) > 60 {
		data = "dr:" + action + ":id:" + s.StorePath(path)
	}
	return data
}

func joinDrivePath(dir, name string) string {
	if dir == "" {
		return name
	}
	return strings.TrimSuffix(dir, "/") + "/" + name
}

func buildDriveNavigationKeyboard(s *service.BotService, lang string, files []service.DriveFile, currentRelPath string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	folderCount := 0
	for _, f := range files {
		if f.IsDir && folderCount < 10 {
			folderCount++
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%d. %s %s", folderCount, IconFolder, utils.TruncateString(f.Name, 25)),
				driveCallbackData(s, "c", joinDrivePath(currentRelPath, f.Name)),
			)))
		}
	}

	fileCount := 0
	for _, f := range files {
		if !f.IsDir && fileCount < 10 {
			fileCount++
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%d. %s %s", fileCount, getFileIcon(f.Name), utils.TruncateString(f.Name, 25)),
				driveCallbackData(s, "i", joinDrivePath(currentRelPath, f.Name)),
			)))
		}
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "fm_btn_refresh"), driveCallbackData(s, "c", currentRelPath)),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "fm_btn_root"), "dr:h"),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_close"), "dr:x"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func buildSearchNavigationKeyboard(s *service.BotService, lang string, files []service.DriveFile) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, f := range files {
		if i >= 20 {
			break
		}
		path := f.Path
		if path == "" {
			path = f.Name
		}
		action, icon := "i", getFileIcon(f.Name)
		if f.IsDir {
			action, icon = "c", IconFolder
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d. %s %s", i+1, icon, utils.TruncateString(filepath.Base(f.Name), 25)),
			driveCallbackData(s, action, path),
		)))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_close"), "dr:x"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func deleteConfirmKeyboard(s *service.BotService, lang, relPath string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "fm_btn_delete"), driveCallbackData(s, "df", relPath)),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "fm_btn_cancel"), "dr:xf"),
	))
}
