package media

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"zee-mirror/internal/service"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func GetErrorMessage(title, content string) string {
	return fmt.Sprintf("❌ *%s*\n\n%s", title, content)
}

func GetSuccessMessage(title, content string) string {
	return fmt.Sprintf("✅ *%s*\n\n%s", title, content)
}

func HandleExtractAudio(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	fileID, fileName := s.GetFileFromMessage(message)
	inputPath := args

	if fileID == "" && inputPath == "" {
		s.Reply(message, GetErrorMessage("SAI ĐỊNH DẠNG", "Hãy reply vào một file video bằng /extractaudio hoặc dùng:\n`/extractaudio path`"))
		return
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, "🎵 *Đang tách audio\\.\\.\\.*")
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, err := s.Bot.Send(statusMsg)
	if err != nil {
		slog.Error("Failed to send status message", "error", err)
		s.Reply(message, "❌ *Không gửi được tin nhắn trạng thái*")
		return
	}

	if fileID != "" {
		var downloadErr error
		inputPath, downloadErr = s.DownloadTelegramFile(fileID, fileName)
		if downloadErr != nil {
			slog.Error("Telegram download failed during audio extraction", "error", downloadErr)
			s.EditMessage(sent.Chat.ID, sent.MessageID, fmt.Sprintf("❌ *Tải file thất bại*\nLỗi: %s", utils.EscapeMarkdownV2(downloadErr.Error())))
			return
		}
	}

	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".mp3"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if hasAudio, errAudio := s.HasAudioStream(inputPath); errAudio != nil || !hasAudio {
		s.EditMessage(sent.Chat.ID, sent.MessageID, GetErrorMessage("KHÔNG CÓ AUDIO", "Video này không có track audio để tách\\."))
		return
	}

	inputDir := filepath.Dir(inputPath)
	inputName := filepath.Base(inputPath)
	outputName := filepath.Base(outputPath)

	cmd := execCommand(ctx, "ffmpeg", "-i", inputName, "-vn", "-acodec", "libmp3lame", "-q:a", "2", outputName, "-y")
	cmd.Dir = inputDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.EditMessage(sent.Chat.ID, sent.MessageID, GetErrorMessage("LỖI FFMPEG", fmt.Sprintf("Tách audio thất bại: %s\n\nOutput \\(phần cuối\\):\n`%s`",
			utils.EscapeMarkdownV2(err.Error()),
			utils.EscapeMarkdownV2Code(utils.GetLastLines(string(output), 15)))))
		return
	}

	content := fmt.Sprintf("🎵 Output: `%s`\\!\n\nBấm nút bên dưới để tải lên Drive\\.", utils.EscapeMarkdownV2Code(filepath.Base(outputPath)))
	pathID := s.StorePath(outputPath)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📤 Tải lên Drive", "media_m:"+pathID),
		),
	)
	s.EditMessageWithMarkup(sent.Chat.ID, sent.MessageID, GetSuccessMessage("ĐÃ TÁCH AUDIO", content), &keyboard)
}

func HandleCompressVideo(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	fileID, fileName := s.GetFileFromMessage(message)
	inputPath := args
	quality := "medium"

	if parts := strings.Fields(args); len(parts) > 0 {
		if fileID == "" {
			inputPath = parts[0]
			if len(parts) > 1 {
				quality = parts[1]
			}
		} else {
			quality = parts[0]
		}
	}

	if fileID == "" && inputPath == "" {
		s.Reply(message, "⚠️ *Sai định dạng*\n\nHãy reply vào một file video bằng /compress hoặc dùng:\n`/compress path/to/video [quality]`\n\nChất lượng: low, medium, high \\(mặc định: medium\\)")
		return
	}

	crf := "28"
	switch quality {
	case "low":
		crf = "32"
	case "high":
		crf = "23"
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, "🗜️ *Đang nén video\\.\\.\\.*")
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, err := s.Bot.Send(statusMsg)
	if err != nil {
		slog.Error("Failed to send status message", "error", err)
		s.Reply(message, "❌ *Không gửi được tin nhắn trạng thái*")
		return
	}

	if fileID != "" {
		var downloadErr error
		inputPath, downloadErr = s.DownloadTelegramFile(fileID, fileName)
		if downloadErr != nil {
			s.EditMessage(sent.Chat.ID, sent.MessageID, "❌ *Tải file thất bại*")
			return
		}
	}

	ext := filepath.Ext(inputPath)
	outputPath := strings.TrimSuffix(inputPath, ext) + "_compressed" + ext

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	inputDir := filepath.Dir(inputPath)
	inputName := filepath.Base(inputPath)
	outputName := filepath.Base(outputPath)

	cmd := execCommand(ctx, "ffmpeg", "-i", inputName,
		"-c:v", "libx264", "-crf", crf, "-preset", "medium",
		"-c:a", "aac", "-b:a", "128k",
		outputName, "-y")
	cmd.Dir = inputDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		s.EditMessage(sent.Chat.ID, sent.MessageID, fmt.Sprintf("❌ *Nén video thất bại*\n\nLỗi: %s\n\nOutput \\(phần cuối\\):\n`%s`",
			utils.EscapeMarkdownV2(err.Error()),
			utils.EscapeMarkdownV2Code(utils.GetLastLines(string(output), 15))))
		return
	}

	content := fmt.Sprintf("🎬 Output: `%s`\\!\n📊 Chất lượng: %s\n\nBấm nút bên dưới để tải lên Drive\\.",
		utils.EscapeMarkdownV2Code(filepath.Base(outputPath)),
		utils.EscapeMarkdownV2(quality))
	pathID := s.StorePath(outputPath)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📤 Tải lên Drive", "media_m:"+pathID),
		),
	)
	s.EditMessageWithMarkup(sent.Chat.ID, sent.MessageID, GetSuccessMessage("ĐÃ NÉN VIDEO", content), &keyboard)
}

func HandleGenerateThumbnail(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	fileID, fileName := s.GetFileFromMessage(message)
	timestamp := "00:00:05"
	inputPath := ""

	if args != "" {
		parts := strings.Fields(args)
		if fileID != "" {
			timestamp = parts[0]
		} else {
			if len(parts) > 0 {
				inputPath = parts[0]
			}
			if len(parts) > 1 {
				timestamp = parts[1]
			}
		}
	}

	if fileID == "" && inputPath == "" {
		s.Reply(message, "⚠️ *Sai định dạng*\n\nHãy reply vào một file video bằng /thumbnail hoặc dùng:\n`/thumbnail [timestamp]`\n\nVí dụ: `/thumbnail 00:01:30`")
		return
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, "🖼️ *Đang tạo thumbnail\\.\\.\\.*")
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, err := s.Bot.Send(statusMsg)
	if err != nil {
		slog.Error("Failed to send status message", "error", err)
		s.Reply(message, "❌ *Không gửi được tin nhắn trạng thái*")
		return
	}

	if fileID != "" {
		var downloadErr error
		inputPath, downloadErr = s.DownloadTelegramFile(fileID, fileName)
		if downloadErr != nil {
			s.EditMessage(sent.Chat.ID, sent.MessageID, "❌ *Tải file thất bại*")
			return
		}
	}

	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + "_thumb.jpg"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	inputDir := filepath.Dir(inputPath)
	inputName := filepath.Base(inputPath)
	outputName := filepath.Base(outputPath)

	cmd := execCommand(ctx, "ffmpeg", "-i", inputName,
		"-ss", timestamp, "-vframes", "1",
		"-q:v", "2", outputName, "-y")
	cmd.Dir = inputDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		s.EditMessage(sent.Chat.ID, sent.MessageID, fmt.Sprintf("❌ *Tạo thumbnail thất bại*\n\nLỗi: %s\n\nOutput \\(phần cuối\\):\n`%s`",
			utils.EscapeMarkdownV2(err.Error()),
			utils.EscapeMarkdownV2Code(utils.GetLastLines(string(output), 15))))
		return
	}

	photo := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FilePath(outputPath))
	photo.Caption = fmt.Sprintf("🖼️ Thumbnail từ `%s` tại %s", utils.EscapeMarkdownV2Code(filepath.Base(inputPath)), utils.EscapeMarkdownV2(timestamp))
	photo.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(photo)

	_, _ = s.Bot.Request(tgbotapi.NewDeleteMessage(sent.Chat.ID, sent.MessageID))
}

func HandleEmbedSubtitle(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	fileID, fileName := s.GetFileFromMessage(message)
	var videoPath, subPath string

	parts := strings.Fields(args)
	if fileID != "" {
		var err error
		videoPath, err = s.DownloadTelegramFile(fileID, fileName)
		if err != nil {
			s.Reply(message, "❌ *Tải file thất bại*")
			return
		}
		if len(parts) > 0 {
			subPath = parts[0]
		}
	} else {
		if len(parts) < 2 {
			s.Reply(message, "⚠️ *Sai định dạng*\n\nDùng: `/subtitle video.mp4 subtitle.srt` hoặc reply vào video bằng `/subtitle subtitle.srt`")
			return
		}
		videoPath = parts[0]
		subPath = parts[1]
	}

	if videoPath == "" || subPath == "" {
		s.Reply(message, "⚠️ *Sai định dạng*\n\nDùng: `/subtitle video.mp4 subtitle.srt` hoặc reply vào video bằng `/subtitle subtitle.srt`")
		return
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, "📝 *Đang gắn subtitle\\.\\.\\.*")
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, err := s.Bot.Send(statusMsg)
	if err != nil {
		slog.Error("Failed to send status message", "error", err)
		s.Reply(message, "❌ *Không gửi được tin nhắn trạng thái*")
		return
	}

	ext := filepath.Ext(videoPath)
	outputPath := strings.TrimSuffix(videoPath, ext) + "_subbed" + ext

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	inputDir := filepath.Dir(videoPath)

	cmd := execCommand(ctx, "ffmpeg", "-i", filepath.Base(videoPath),
		"-i", "file:"+subPath,
		"-c", "copy", "-c:s", "mov_text",
		filepath.Base(outputPath), "-y")
	cmd.Dir = inputDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		s.EditMessage(sent.Chat.ID, sent.MessageID, fmt.Sprintf("❌ *Gắn subtitle thất bại*\n\nLỗi: %s\n\nOutput \\(phần cuối\\):\n`%s`",
			utils.EscapeMarkdownV2(err.Error()),
			utils.EscapeMarkdownV2Code(utils.GetLastLines(string(output), 15))))
		return
	}

	content := fmt.Sprintf("🎬 Output: `%s`\\!\n\nBấm nút bên dưới để tải lên Drive\\.",
		utils.EscapeMarkdownV2Code(filepath.Base(outputPath)))
	pathID := s.StorePath(outputPath)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📤 Tải lên Drive", "media_m:"+pathID),
		),
	)
	s.EditMessageWithMarkup(sent.Chat.ID, sent.MessageID, GetSuccessMessage("ĐÃ GẮN SUBTITLE", content), &keyboard)
}

func HandleConvertFormat(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	fileID, fileName := s.GetFileFromMessage(message)
	var inputPath, targetFormat string

	parts := strings.Fields(args)
	if fileID != "" {
		var err error
		inputPath, err = s.DownloadTelegramFile(fileID, fileName)
		if err != nil {
			s.Reply(message, "❌ *Tải file thất bại*")
			return
		}
		if len(parts) > 0 {
			targetFormat = strings.ToLower(parts[0])
		}
	} else {
		if len(parts) < 2 {
			s.Reply(message, "⚠️ *Sai định dạng*\n\nDùng: `/convert input.mp4 mkv` hoặc reply vào file bằng `/convert mkv`")
			return
		}
		inputPath = parts[0]
		targetFormat = strings.ToLower(parts[1])
	}

	validFormats := map[string]bool{
		"mp4": true, "mkv": true, "avi": true, "mov": true, "webm": true,
		"mp3": true, "aac": true, "flac": true, "wav": true,
	}

	if !validFormats[targetFormat] {
		s.Reply(message, "❌ *Định dạng không được hỗ trợ*\n\nHỗ trợ: mp4, mkv, avi, mov, webm, mp3, aac, flac, wav")
		return
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("🔄 *Đang chuyển đổi sang %s\\.\\.\\.*", targetFormat))
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, err := s.Bot.Send(statusMsg)
	if err != nil {
		slog.Error("Failed to send status message", "error", err)
		s.Reply(message, "❌ *Không gửi được tin nhắn trạng thái*")
		return
	}

	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + "." + targetFormat

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	inputDir := filepath.Dir(inputPath)
	var cmd *exec.Cmd
	switch targetFormat {
	case "mp3", "aac", "flac", "wav":

		cmd = execCommand(ctx, "ffmpeg", "-i", filepath.Base(inputPath), "-vn", filepath.Base(outputPath), "-y")
	default:

		cmd = execCommand(ctx, "ffmpeg", "-i", filepath.Base(inputPath), "-c:v", "copy", "-c:a", "copy", filepath.Base(outputPath), "-y")
	}
	cmd.Dir = inputDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		s.EditMessage(sent.Chat.ID, sent.MessageID, fmt.Sprintf("❌ *Chuyển đổi định dạng thất bại*\n\nLỗi: %s\n\nOutput \\(phần cuối\\):\n`%s`",
			utils.EscapeMarkdownV2(err.Error()),
			utils.EscapeMarkdownV2Code(utils.GetLastLines(string(output), 15))))
		return
	}

	content := fmt.Sprintf("📄 Output: `%s`\\!\n\nBấm nút bên dưới để tải lên Drive\\.",
		utils.EscapeMarkdownV2Code(filepath.Base(outputPath)))
	pathID := s.StorePath(outputPath)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📤 Tải lên Drive", "media_m:"+pathID),
		),
	)
	s.EditMessageWithMarkup(sent.Chat.ID, sent.MessageID, GetSuccessMessage("ĐÃ CHUYỂN ĐỔI MEDIA", content), &keyboard)
}

func HandleMediaInfo(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	fileID, fileName := s.GetFileFromMessage(message)
	inputPath := args

	if fileID != "" {
		var err error
		inputPath, err = s.DownloadTelegramFile(fileID, fileName)
		if err != nil {
			s.Reply(message, "❌ *Tải file thất bại*")
			return
		}
	}

	if inputPath == "" {
		s.Reply(message, "⚠️ *Sai định dạng*\n\nHãy reply vào một file media hoặc dùng:\n`/mediainfo path/to/file`")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := execCommand(ctx, "ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", inputPath)
	output, err := cmd.Output()
	if err != nil {
		s.Reply(message, fmt.Sprintf("❌ *Không lấy được thông tin*\n\nLỗi: %s", utils.EscapeMarkdownV2(err.Error())))
		return
	}

	var probe service.FFProbeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		slog.Error("FFProbe JSON unmarshal failed", "error", err)
		s.Reply(message, "❌ *Không phân tích được metadata của file*")
		return
	}

	text := s.FormatMediaInfo(filepath.Base(inputPath), probe)
	s.Reply(message, text)
}

func HandleMediaMirrorCallback(s *service.BotService, cb *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) < 2 {
		return
	}

	pathID := parts[1]
	fullPath, ok := s.GetPath(pathID)
	if !ok {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(cb.ID, "❌ Liên kết đã hết hạn hoặc không hợp lệ!"))
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(cb.ID, "❌ File đã bị xóa khỏi server!"))
		return
	}

	_, _ = s.Bot.Request(tgbotapi.NewCallback(cb.ID, "📤 Đang bắt đầu Mirror lên Drive..."))

	s.TaskManager.Mu.Lock()
	s.TaskManager.LastStatusMsg[cb.Message.Chat.ID] = cb.Message.MessageID
	s.TaskManager.Mu.Unlock()

	url := "file://" + fullPath
	fileName := filepath.Base(fullPath)

	replyID := 0
	if cb.Message.ReplyToMessage != nil {
		replyID = cb.Message.ReplyToMessage.MessageID
	}

	task, err := s.TaskManager.CreateTask(service.TypeMirror, url, fileName, cb.Message.Chat.ID, 0, replyID, cb.From.ID, false, false, "", "", 0, "", false)
	if err != nil {
		s.HandleCreateTaskError(cb.Message.Chat.ID, cb.Message.MessageID, err)
		return
	}
	s.HandleAutoDelete(task)
	s.UpdateSharedDashboard(cb.Message.Chat.ID, false)
	slog.Info("Local media mirror task created", "taskID", task.ID, "path", fullPath)
}

func HandleScreenshots(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	fileID, fileName := s.GetFileFromMessage(message)
	count := 4
	inputPath := args

	parts := strings.Fields(args)
	if fileID != "" {
		var err error
		inputPath, err = s.DownloadTelegramFile(fileID, fileName)
		if err != nil {
			s.Reply(message, "❌ *Tải file thất bại*")
			return
		}
		if len(parts) > 0 {
			_, _ = fmt.Sscanf(parts[0], "%d", &count)
		}
	} else if len(parts) > 0 {
		inputPath = parts[0]
		if len(parts) > 1 {
			_, _ = fmt.Sscanf(parts[1], "%d", &count)
		}
	}

	if count > 10 {
		count = 10
	} else if count < 1 {
		count = 1
	}

	if inputPath == "" {
		s.Reply(message, "⚠️ *Sai định dạng*\n\nDùng: `/screenshots video.mp4 [count]` hoặc reply vào video bằng `/screenshots [count]`")
		return
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("📸 *Đang tạo %d ảnh chụp màn hình\\.\\.\\.*", count))
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, err := s.Bot.Send(statusMsg)
	if err != nil {
		slog.Error("Failed to send status message during screenshots", "error", err)
		s.Reply(message, "❌ *Không gửi được tin nhắn trạng thái*")
		return
	}

	screenshots, err := s.GenerateScreenshotsList(inputPath, count)
	if err != nil {
		s.EditMessage(sent.Chat.ID, sent.MessageID, fmt.Sprintf("❌ *Tạo ảnh chụp màn hình thất bại*\n\nLỗi: %s", utils.EscapeMarkdownV2(err.Error())))
		return
	}

	var mediaGroup []interface{}
	for _, ss := range screenshots {
		photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(ss))
		mediaGroup = append(mediaGroup, photo)
	}

	mediaMsg := tgbotapi.NewMediaGroup(message.Chat.ID, mediaGroup)
	_, _ = s.Bot.Send(mediaMsg)

	_, _ = s.Bot.Request(tgbotapi.NewDeleteMessage(sent.Chat.ID, sent.MessageID))
}

func HandleHardsub(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	fileID, fileName := s.GetFileFromMessage(message)
	var videoPath, subPath string

	parts := strings.Fields(args)
	if fileID != "" {
		var err error
		videoPath, err = s.DownloadTelegramFile(fileID, fileName)
		if err != nil {
			s.Reply(message, "❌ *Tải file thất bại*")
			return
		}
		if len(parts) > 0 {
			subPath = parts[0]
		}
	} else {
		if len(parts) < 2 {
			s.Reply(message, "⚠️ *Sai định dạng*\n\nDùng: `/hardsub video.mp4 subtitle.srt` hoặc reply vào video bằng `/hardsub subtitle.srt`")
			return
		}
		videoPath = parts[0]
		subPath = parts[1]
	}

	if videoPath == "" || subPath == "" {
		s.Reply(message, "⚠️ *Sai định dạng*\n\nDùng: `/hardsub video.mp4 subtitle.srt` hoặc reply vào video bằng `/hardsub subtitle.srt`")
		return
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, "🔥 *Đang ghép cứng subtitle \\(Hard\\-sub\\)\\.\\.\\.*\n_Quá trình này mất nhiều thời gian hơn vì phải re\\-encode video\\._")
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, err := s.Bot.Send(statusMsg)
	if err != nil {
		slog.Error("Failed to send status message", "error", err)
		return
	}

	ext := filepath.Ext(videoPath)
	outputPath := strings.TrimSuffix(videoPath, ext) + "_hardsub" + ext

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	absSubPath, _ := filepath.Abs(subPath)
	ffmpegSubPath := strings.ReplaceAll(absSubPath, "\\", "/")
	ffmpegSubPath = strings.ReplaceAll(ffmpegSubPath, ":", "\\:")

	cmd := execCommand(ctx, "ffmpeg", "-i", videoPath,
		"-vf", fmt.Sprintf("subtitles='%s'", ffmpegSubPath),
		"-c:v", "libx264", "-crf", "23", "-preset", "medium",
		"-c:a", "copy",
		outputPath, "-y")

	output, err := cmd.CombinedOutput()
	if err != nil {
		s.EditMessage(sent.Chat.ID, sent.MessageID, fmt.Sprintf("❌ *Ghép cứng subtitle thất bại*\n\nLỗi: %s\n\nOutput \\(phần cuối\\):\n`%s`",
			utils.EscapeMarkdownV2(err.Error()),
			utils.EscapeMarkdownV2Code(utils.GetLastLines(string(output), 15))))
		return
	}

	content := fmt.Sprintf("🎬 Output: `%s`\\!\n\nBấm nút bên dưới để tải lên Drive\\.",
		utils.EscapeMarkdownV2Code(filepath.Base(outputPath)))
	pathID := s.StorePath(outputPath)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📤 Tải lên Drive", "media_m:"+pathID),
		),
	)
	s.EditMessageWithMarkup(sent.Chat.ID, sent.MessageID, GetSuccessMessage("ĐÃ GHÉP CỨNG SUBTITLE", content), &keyboard)
}

func HandleRescale(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	fileID, fileName := s.GetFileFromMessage(message)
	var inputPath, resolution string

	parts := strings.Fields(args)
	if fileID != "" {
		if len(parts) == 0 {
			s.Reply(message, "⚠️ *Sai định dạng*\n\nHãy reply vào video bằng `/rescale 1280x720` hoặc `/rescale 720p`")
			return
		}
		var err error
		inputPath, err = s.DownloadTelegramFile(fileID, fileName)
		if err != nil {
			s.Reply(message, "❌ *Tải file thất bại*")
			return
		}
		resolution = parts[0]
	} else {
		if len(parts) < 2 {
			s.Reply(message, "⚠️ *Sai định dạng*\n\nDùng: `/rescale video.mp4 1280x720` hoặc `/rescale video.mp4 720p`")
			return
		}
		inputPath = parts[0]
		resolution = parts[1]
	}

	scale := ""
	switch strings.ToLower(resolution) {
	case "4k":
		scale = "3840:2160"
	case "2k":
		scale = "2560:1440"
	case "1080p":
		scale = "1920:1080"
	case "720p":
		scale = "1280:720"
	case "480p":
		scale = "854:480"
	case "360p":
		scale = "640:360"
	default:
		switch {
		case strings.Contains(resolution, "x"):
			scale = strings.Replace(resolution, "x", ":", 1)
		case strings.Contains(resolution, ":"):
			scale = resolution
		default:
			s.Reply(message, "❌ *Độ phân giải không hợp lệ*\n\nDùng định dạng: `1280x720` hoặc `720p`")
			return
		}
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("📐 *Đang đổi độ phân giải video sang %s\\.\\.\\.*", resolution))
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, err := s.Bot.Send(statusMsg)
	if err != nil {
		slog.Error("Failed to send status message", "error", err)
		return
	}

	ext := filepath.Ext(inputPath)
	outputPath := strings.TrimSuffix(inputPath, ext) + "_" + resolution + ext

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	cmd := execCommand(ctx, "ffmpeg", "-i", inputPath,
		"-vf", fmt.Sprintf("scale=%s:force_original_aspect_ratio=decrease,pad=%s:(ow-iw)/2:(oh-ih)/2", scale, scale),
		"-c:v", "libx264", "-crf", "23", "-preset", "medium",
		"-c:a", "copy",
		outputPath, "-y")

	output, err := cmd.CombinedOutput()
	if err != nil {
		s.EditMessage(sent.Chat.ID, sent.MessageID, fmt.Sprintf("❌ *Đổi độ phân giải video thất bại*\n\nLỗi: %s\n\nOutput \\(phần cuối\\):\n`%s`",
			utils.EscapeMarkdownV2(err.Error()),
			utils.EscapeMarkdownV2Code(utils.GetLastLines(string(output), 15))))
		return
	}

	content := fmt.Sprintf("🎬 Output: `%s`\\!\n📐 Độ phân giải: %s\n\nBấm nút bên dưới để tải lên Drive\\.",
		utils.EscapeMarkdownV2Code(filepath.Base(outputPath)),
		utils.EscapeMarkdownV2(resolution))
	pathID := s.StorePath(outputPath)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📤 Tải lên Drive", "media_m:"+pathID),
		),
	)
	s.EditMessageWithMarkup(sent.Chat.ID, sent.MessageID, GetSuccessMessage("ĐÃ ĐỔI ĐỘ PHÂN GIẢI VIDEO", content), &keyboard)
}
