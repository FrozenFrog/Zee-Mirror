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
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleDriveMkdir(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}
	if args == "" {
		s.Reply(message, "⚠️ *Sai định dạng*\n\nCách dùng: `/mkdir tên_thư_mục`")
		return
	}

	folderPath := s.TaskManager.RcloneDest + "/" + args

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	configPath := s.TaskManager.ConfigDir + "/rclone.conf"
	cmd := exec.CommandContext(ctx, "rclone", "mkdir", folderPath, "--config", configPath)

	if err := cmd.Run(); err != nil {
		s.Reply(message, fmt.Sprintf("❌ *Tạo thư mục thất bại*\n\nLỗi: %s", utils.EscapeMarkdownV2(err.Error())))
		return
	}

	s.Reply(message, fmt.Sprintf("✅ *Đã tạo thư mục*\n\n📁 `%s`", utils.EscapeMarkdownV2(args)))
}

func HandleDriveDelete(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.IsAdmin(message.From.ID) {
		s.Reply(message, "❌ *Truy cập bị từ chối*\nChỉ Admin mới có thể xóa file\\.")
		return
	}

	if args == "" {
		s.Reply(message, "⚠️ *Sai định dạng*\n\nCách dùng: `/rm tên_file_hoặc_thư_mục`")
		return
	}

	targetPath := s.TaskManager.RcloneDest + "/" + args

	msg := tgbotapi.NewMessage(message.Chat.ID,
		fmt.Sprintf("⚠️ *Xác nhận xóa*\n\nBạn có chắc muốn xóa:\n📁 `%s`\\?",
			utils.EscapeMarkdownV2(targetPath)))
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(msg)
}

func HandleDriveMove(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		s.Reply(message, "⚠️ *Sai định dạng*\n\nCách dùng: `/mv nguồn đích`")
		return
	}

	source := s.TaskManager.RcloneDest + "/" + parts[0]
	dest := s.TaskManager.RcloneDest + "/" + parts[1]

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	configPath := s.TaskManager.ConfigDir + "/rclone.conf"
	cmd := exec.CommandContext(ctx, "rclone", "moveto", source, dest, "--config", configPath)

	if err := cmd.Run(); err != nil {
		s.Reply(message, fmt.Sprintf("❌ *Di chuyển file thất bại*\n\nLỗi: %s", utils.EscapeMarkdownV2(err.Error())))
		return
	}

	s.Reply(message, fmt.Sprintf("✅ *Đã di chuyển file*\n\n📄 `%s`\n➡️ `%s`",
		utils.EscapeMarkdownV2(parts[0]),
		utils.EscapeMarkdownV2(parts[1])))
}

func HandleDriveShare(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	if args == "" {
		s.Reply(message, "⚠️ *Sai định dạng*\n\nCách dùng: `/share tên_file`")
		return
	}

	targetPath := s.TaskManager.RcloneDest + "/" + args

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, "🔗 *Đang tạo link chia sẻ\\.\\.\\.*")
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
				fmt.Sprintf("❌ *Tạo link thất bại*\n\nLỗi: %s", utils.EscapeMarkdownV2(err.Error())))
			editMsg.ParseMode = tgbotapi.ModeMarkdownV2
			_, _ = s.Bot.Send(editMsg)
			return
		}
		link = strings.TrimSpace(string(output))
	}

	text := fmt.Sprintf("✅ *Đã tạo link chia sẻ*\n\n📄 File: `%s`\n🔗 Link: %s",
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
	if args == "" {
		s.Reply(message, "⚠️ *Sai định dạng*\n\nCách dùng: `/find từ_khóa`")
		return
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, "🔍 *Đang tìm kiếm file\\.\\.\\.*")
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
			fmt.Sprintf("❌ *Tìm kiếm file thất bại*\n\nLỗi: %s", utils.EscapeMarkdownV2(err.Error())))
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
		_, _ = s.Bot.Send(editMsg)
		return
	}

	var rawFiles []service.DriveFile
	if err := json.Unmarshal(output, &rawFiles); err != nil {
		editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sent.MessageID,
			"❌ *Không đọc được kết quả tìm kiếm*")
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

	text := formatSearchResults(args, files)
	keyboard := buildSearchNavigationKeyboard(s, files)
	editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sent.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	if len(keyboard.InlineKeyboard) > 0 {
		editMsg.ReplyMarkup = &keyboard
	}
	_, _ = s.Bot.Send(editMsg)
}

func formatSearchResults(query string, files []service.DriveFile) string {
	var text strings.Builder

	text.WriteString("🔍 *KẾT QUẢ TÌM KIẾM*\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")
	fmt.Fprintf(&text, "🔎 Từ khóa: `%s`\n\n", utils.EscapeMarkdownV2(query))

	if len(files) == 0 {
		text.WriteString("📭 _Không tìm thấy file nào_\n")
	} else {
		maxShow := 20
		if len(files) > maxShow {
			fmt.Fprintf(&text, "📊 Hiển thị %d trên %d kết quả\n\n", maxShow, len(files))
		}

		text.WriteString("📂 *FILE TÌM THẤY*\n")

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
