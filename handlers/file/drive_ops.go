package file

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os/exec"
	"strings"
	"time"
	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleDriveMkdir(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}
	lang := s.GetUserLanguage(message.From.ID)
	if args == "" {
		s.Reply(message, i18n.T(lang, "drive_usage_mkdir"))
		return
	}

	folderPath := s.TaskManager.RcloneDest + "/" + args

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	configPath := s.TaskManager.ConfigDir + "/rclone.conf"
	cmd := exec.CommandContext(ctx, "rclone", "mkdir", folderPath, "--config", configPath)

	if err := cmd.Run(); err != nil {
		s.Reply(message, i18n.T(lang, "drive_mkdir_failed", utils.EscapeMarkdownV2(err.Error())))
		return
	}

	s.Reply(message, i18n.T(lang, "drive_mkdir_ok", utils.EscapeMarkdownV2(args)))
}

func HandleDriveDelete(s *service.BotService, message *tgbotapi.Message, args string) {
	lang := s.GetUserLanguage(message.From.ID)
	if !s.IsAdmin(message.From.ID) {
		s.Reply(message, i18n.T(lang, "drive_rm_denied"))
		return
	}

	if args == "" {
		s.Reply(message, i18n.T(lang, "drive_usage_rm"))
		return
	}

	targetPath := s.TaskManager.RcloneDest + "/" + args

	msg := tgbotapi.NewMessage(message.Chat.ID,
		i18n.T(lang, "drive_rm_confirm", utils.EscapeMarkdownV2(targetPath)))
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.ReplyMarkup = deleteConfirmKeyboard(s, lang, args)
	_, _ = s.Bot.Send(msg)
}

func HandleDriveMove(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	lang := s.GetUserLanguage(message.From.ID)
	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		s.Reply(message, i18n.T(lang, "drive_usage_mv"))
		return
	}

	source := s.TaskManager.RcloneDest + "/" + parts[0]
	dest := s.TaskManager.RcloneDest + "/" + parts[1]

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	configPath := s.TaskManager.ConfigDir + "/rclone.conf"
	cmd := exec.CommandContext(ctx, "rclone", "moveto", source, dest, "--config", configPath)

	if err := cmd.Run(); err != nil {
		s.Reply(message, i18n.T(lang, "drive_mv_failed", utils.EscapeMarkdownV2(err.Error())))
		return
	}

	s.Reply(message, i18n.T(lang, "drive_mv_ok",
		utils.EscapeMarkdownV2(parts[0]),
		utils.EscapeMarkdownV2(parts[1])))
}

func HandleDriveShare(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	lang := s.GetUserLanguage(message.From.ID)
	if args == "" {
		s.Reply(message, i18n.T(lang, "drive_usage_share"))
		return
	}

	targetPath := s.TaskManager.RcloneDest + "/" + args

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, i18n.T(lang, "drive_share_creating"))
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, _ := s.Bot.Send(statusMsg)

	var link string
	if s.Config.IndexURL != "" {
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
		link = fmt.Sprintf("%s/%s", baseURL, encodedPath)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		configPath := s.TaskManager.ConfigDir + "/rclone.conf"
		cmd := exec.CommandContext(ctx, "rclone", "link", targetPath, "--config", configPath)
		output, err := cmd.Output()

		if err != nil {
			editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sent.MessageID,
				i18n.T(lang, "drive_share_failed", utils.EscapeMarkdownV2(err.Error())))
			editMsg.ParseMode = tgbotapi.ModeMarkdownV2
			_, _ = s.Bot.Send(editMsg)
			return
		}
		link = strings.TrimSpace(string(output))
	}

	text := i18n.T(lang, "drive_share_ok",
		utils.EscapeMarkdownV2(args),
		utils.EscapeMarkdownV2(link))

	editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sent.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
}

func HandleDriveSearch(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}
	lang := s.GetUserLanguage(message.From.ID)
	if args == "" {
		s.Reply(message, i18n.T(lang, "drive_usage_find"))
		return
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, i18n.T(lang, "drive_find_searching"))
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, _ := s.Bot.Send(statusMsg)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	configPath := s.TaskManager.ConfigDir + "/rclone.conf"
	searchPath := s.TaskManager.RcloneDest

	slog.Info("Searching drive", "query", args, "path", searchPath)

	cmd := exec.CommandContext(ctx, "rclone", "lsjson", searchPath, "--config", configPath,
		"--include", "*"+args+"*", "-R", "--max-depth", "5", "--ignore-case")
	output, err := cmd.Output()

	if err != nil {
		editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sent.MessageID,
			i18n.T(lang, "drive_find_failed", utils.EscapeMarkdownV2(utils.RcloneError(err).Error())))
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
		_, _ = s.Bot.Send(editMsg)
		return
	}

	var rawFiles []service.DriveFile
	if err := json.Unmarshal(output, &rawFiles); err != nil {
		editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sent.MessageID,
			i18n.T(lang, "drive_find_parse_failed"))
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
		_, _ = s.Bot.Send(editMsg)
		return
	}

	var files []service.DriveFile
	queryLower := strings.ToLower(args)
	for _, f := range rawFiles {
		if strings.Contains(strings.ToLower(f.Name), queryLower) {
			files = append(files, f)
		}
	}

	text := formatSearchResults(lang, args, files)
	keyboard := buildSearchNavigationKeyboard(s, lang, files)
	editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sent.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	if len(keyboard.InlineKeyboard) > 0 {
		editMsg.ReplyMarkup = &keyboard
	}
	_, _ = s.Bot.Send(editMsg)
}

func formatSearchResults(lang, query string, files []service.DriveFile) string {
	var text strings.Builder

	text.WriteString(i18n.T(lang, "drive_find_title") + "\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")
	text.WriteString(i18n.T(lang, "drive_find_keyword", utils.EscapeMarkdownV2(query)) + "\n\n")

	if len(files) == 0 {
		text.WriteString(i18n.T(lang, "drive_find_none") + "\n")
	} else {
		maxShow := 20
		if len(files) > maxShow {
			text.WriteString(i18n.T(lang, "drive_find_showing", maxShow, len(files)) + "\n\n")
		}

		text.WriteString(i18n.T(lang, "drive_find_header") + "\n")

		for i, f := range files {
			if i >= maxShow {
				break
			}
			icon := IconFolder
			if !f.IsDir {
				icon = getFileIcon(f.Name)
			}
			size := ""
			if !f.IsDir {
				size = fmt.Sprintf(" \\(%s\\)", utils.EscapeMarkdownV2(utils.FormatBytes(f.Size)))
			}
			path := f.Path
			if path == "" {
				path = f.Name
			}
			fmt.Fprintf(&text, "%d\\. %s `%s`%s\n", i+1, icon, utils.EscapeMarkdownV2(utils.TruncateString(path, 45)), size)
		}
	}

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")
	return text.String()
}
