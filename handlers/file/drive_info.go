package file

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleDriveFileInfoDetailed(s *service.BotService, lang string, chatID int64, messageID int, relPath string) {
	targetPath, configPath := resolvePaths(s, relPath)

	slog.Debug("Detailed drive info request", "relPath", relPath, "targetPath", targetPath)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	file := getDriveMetadata(ctx, targetPath, configPath)
	cloudLink := getCloudLink(ctx, s, targetPath, configPath)

	text := buildFileInfoMessage(lang, relPath, file)
	keyboard := buildFileInfoKeyboard(s, lang, relPath, cloudLink)

	s.SendOrEditMessage(chatID, messageID, text, keyboard)
}

func resolvePaths(s *service.BotService, relPath string) (string, string) {
	basePath := strings.TrimSuffix(s.TaskManager.RcloneDest, "/")
	cleanedRel := strings.Trim(relPath, "/")
	targetPath := basePath
	if cleanedRel != "" {
		targetPath = basePath + "/" + cleanedRel
	}
	configPath := s.TaskManager.ConfigDir + "/rclone.conf"
	return targetPath, configPath
}

func getDriveMetadata(ctx context.Context, targetPath, configPath string) service.DriveFile {
	cmd := exec.CommandContext(ctx, "rclone", "lsjson", targetPath, "--config", configPath)
	output, err := cmd.Output()
	if err != nil {
		slog.Error("rclone lsjson failed for info", "error", utils.RcloneError(err), "path", targetPath)
		return service.DriveFile{}
	}

	var files []service.DriveFile
	if errJSON := json.Unmarshal(output, &files); errJSON == nil && len(files) > 0 {
		file := files[0]
		slog.Debug("Drive metadata found", "name", file.Name, "size", file.Size)
		return file
	}
	return service.DriveFile{}
}

func getCloudLink(ctx context.Context, s *service.BotService, targetPath, configPath string) string {
	if s.Config.IndexURL != "" {
		return generateIndexLink(s, targetPath)
	}

	linkCmd := exec.CommandContext(ctx, "rclone", "link", targetPath, "--config", configPath)
	linkOutput, _ := linkCmd.Output()
	return strings.TrimSpace(string(linkOutput))
}

func generateIndexLink(s *service.BotService, targetPath string) string {
	targetPathSlash := strings.ReplaceAll(targetPath, "\\", "/")
	rcloneDestSlash := strings.ReplaceAll(s.TaskManager.RcloneDest, "\\", "/")
	rcloneDestSlash = strings.TrimRight(rcloneDestSlash, "/")

	var relPath string
	if strings.HasPrefix(targetPathSlash, rcloneDestSlash) {
		relPath = strings.TrimPrefix(targetPathSlash, rcloneDestSlash)
	} else {
		parts := strings.SplitN(targetPathSlash, ":", 2)
		if len(parts) > 1 {
			relPath = parts[1]
		} else {
			relPath = targetPathSlash
		}
	}

	relPath = strings.TrimLeft(relPath, "/")
	pathParts := strings.Split(relPath, "/")
	for i, part := range pathParts {
		pathParts[i] = url.PathEscape(part)
	}
	encodedPath := strings.Join(pathParts, "/")
	baseURL := strings.TrimRight(s.Config.IndexURL, "/")
	return fmt.Sprintf("%s/%s", baseURL, encodedPath)
}

func buildFileInfoMessage(lang, relPath string, file service.DriveFile) string {
	var text strings.Builder
	text.WriteString(i18n.T(lang, "fm_info_title", getFileIcon(relPath)) + "\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")
	text.WriteString(i18n.T(lang, "fm_info_name", utils.EscapeMarkdownV2(filepath.Base(relPath))) + "\n")

	if file.Name != "" {
		text.WriteString(i18n.T(lang, "fm_info_size", utils.EscapeMarkdownV2(utils.FormatBytes(file.Size))) + "\n")
		text.WriteString(i18n.T(lang, "fm_info_modified", utils.EscapeMarkdownV2(file.ModTime)) + "\n")
		text.WriteString(i18n.T(lang, "fm_info_mime", utils.EscapeMarkdownV2(file.MimeType)) + "\n")
	} else {
		text.WriteString(i18n.T(lang, "fm_info_restricted") + "\n")
	}

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")
	return text.String()
}

func buildFileInfoKeyboard(s *service.BotService, lang, relPath, cloudLink string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	if strings.HasPrefix(cloudLink, "http") {
		btnText := i18n.T(lang, "fm_btn_cloud_link")
		if s.Config.IndexURL != "" {
			btnText = i18n.T(lang, "fm_btn_index_link")
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonURL(btnText, cloudLink)))
	}

	dirPath := filepath.Dir(relPath)
	if dirPath == "." || dirPath == "/" {
		dirPath = ""
	}
	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "fm_btn_folder"), driveCallbackData(s, "c", dirPath)),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "fm_btn_rm"), driveCallbackData(s, "rm", relPath)),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_close"), "dr:x")),
	)
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func getFileIcon(filename string) string {
	lower := strings.ToLower(filename)

	switch {
	case strings.HasSuffix(lower, ".mp4"), strings.HasSuffix(lower, ".mkv"),
		strings.HasSuffix(lower, ".avi"), strings.HasSuffix(lower, ".mov"):
		return "🎬"
	case strings.HasSuffix(lower, ".mp3"), strings.HasSuffix(lower, ".flac"),
		strings.HasSuffix(lower, ".wav"), strings.HasSuffix(lower, ".aac"):
		return "🎵"
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"),
		strings.HasSuffix(lower, ".png"), strings.HasSuffix(lower, ".gif"):
		return "🖼️"
	case strings.HasSuffix(lower, ".pdf"):
		return "📕"
	case strings.HasSuffix(lower, ".doc"), strings.HasSuffix(lower, ".docx"):
		return IconFile
	case strings.HasSuffix(lower, ".xls"), strings.HasSuffix(lower, ".xlsx"):
		return "📊"
	case strings.HasSuffix(lower, ".zip"), strings.HasSuffix(lower, ".rar"),
		strings.HasSuffix(lower, ".7z"), strings.HasSuffix(lower, ".tar"):
		return "🗜️"
	case strings.HasSuffix(lower, ".exe"), strings.HasSuffix(lower, ".msi"):
		return "⚙️"
	case strings.HasSuffix(lower, ".iso"):
		return "💿"
	default:
		return IconFile
	}
}
