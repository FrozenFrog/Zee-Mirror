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

	slog.Info("Drive list request", "userID", message.From.ID, "args", args, "editMsgID", editMessageID)

	fullPath, relPath := resolveDrivePath(s, args)
	slog.Debug("Listing drive path", "fullPath", fullPath)

	sent, err := sendLoadingStatus(s, message.Chat.ID, editMessageID)
	if err != nil {
		slog.Error("Error preparing status message", "error", err)
		return
	}

	files, err := listDriveFiles(s, fullPath)
	if err != nil {
		handleDriveError(s, message.Chat.ID, sent.MessageID, "Không tải được danh sách file", err)
		return
	}

	slog.Debug("Drive list found items", "count", len(files))

	if tryJumpToFileInfo(s, message, sent.MessageID, relPath, files) {
		return
	}

	renderDriveList(s, message.Chat.ID, sent.MessageID, relPath, files)
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

func sendLoadingStatus(s *service.BotService, chatID int64, editMessageID int) (*tgbotapi.Message, error) {
	loadingText := "🔍 *Đang tải danh sách file\\.\\.\\.*"
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

func handleDriveError(s *service.BotService, chatID int64, messageID int, title string, err error) {
	slog.Error("Drive operation error", "title", title, "error", err)
	errorText := fmt.Sprintf("❌ *%s*\n\nLỗi: %s", title, utils.EscapeMarkdownV2(err.Error()))
	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, errorText)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
}

func tryJumpToFileInfo(s *service.BotService, message *tgbotapi.Message, messageID int, relPath string, files []service.DriveFile) bool {
	if len(files) == 1 && !files[0].IsDir {
		fileName := files[0].Name
		if relPath == fileName || strings.HasSuffix(relPath, "/"+fileName) {
			slog.Info("Target is a file, jumping to info view", "relPath", relPath, "messageID", messageID)
			handleDriveFileInfoDetailed(s, message.Chat.ID, messageID, relPath)
			return true
		}
	}
	return false
}

func renderDriveList(s *service.BotService, chatID int64, messageID int, relPath string, files []service.DriveFile) {
	currentPathForUI := relPath
	if !strings.HasPrefix(currentPathForUI, "/") {
		currentPathForUI = "/" + currentPathForUI
	}

	text := formatDriveFileList(currentPathForUI, files)
	keyboard := buildDriveNavigationKeyboard(s, files, relPath)

	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	if len(keyboard.InlineKeyboard) > 0 {
		editMsg.ReplyMarkup = &keyboard
	}

	if _, err := s.Bot.Send(editMsg); err != nil {
		slog.Error("Error sending final list message", "error", err)
		handleDriveError(s, chatID, messageID, "Không hiển thị được danh sách file", err)
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
		return nil, fmt.Errorf("rclone lsjson failed: %v", utils.RcloneError(err))
	}

	var files []service.DriveFile
	if err := json.Unmarshal(output, &files); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return files, nil
}

func formatDriveFileList(path string, files []service.DriveFile) string {
	var text strings.Builder

	text.WriteString("📂 *QUẢN LÝ FILE \\- GDRIVE*\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")
	fmt.Fprintf(&text, "📍 *ĐƯỜNG DẪN:* `%s`\n\n", utils.EscapeMarkdownV2Code(path))

	if len(files) == 0 {
		text.WriteString("📭 _Thư mục này trống_\n")
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
			text.WriteString("📁 *THƯ MỤC*\n")
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
			text.WriteString("📄 *FILE*\n")
			for _, f := range fileList {
				text.WriteString(f + "\n")
			}
			text.WriteString("\n")
		}

		text.WriteString("📊 *TỔNG KẾT*\n")
		fmt.Fprintf(&text, "📂 Thư mục: `%d`\n", folderCount)
		fmt.Fprintf(&text, "📄 File: `%d`\n", fileCount)
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

func buildDriveNavigationKeyboard(s *service.BotService, files []service.DriveFile, currentRelPath string) tgbotapi.InlineKeyboardMarkup {
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
		tgbotapi.NewInlineKeyboardButtonData("🔄 Làm mới", driveCallbackData(s, "c", currentRelPath)),
		tgbotapi.NewInlineKeyboardButtonData("🏠 Gốc", "dr:h"),
		tgbotapi.NewInlineKeyboardButtonData("✖️ Đóng", "dr:x"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func buildSearchNavigationKeyboard(s *service.BotService, files []service.DriveFile) tgbotapi.InlineKeyboardMarkup {
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
		tgbotapi.NewInlineKeyboardButtonData("✖️ Đóng", "dr:x"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func deleteConfirmKeyboard(s *service.BotService, relPath string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("✅ Xóa", driveCallbackData(s, "df", relPath)),
		tgbotapi.NewInlineKeyboardButtonData("❌ Hủy", "dr:xf"),
	))
}
