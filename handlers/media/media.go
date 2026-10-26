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
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GetErrorMessage and GetSuccessMessage wrap a MarkdownV2 body under a bold
// header. title is plain text (a translated i18n title) and is escaped here;
// content must already be MarkdownV2.
func GetErrorMessage(title, content string) string {
	return fmt.Sprintf("❌ *%s*\n\n%s", utils.EscapeMarkdownV2(title), content)
}

func GetSuccessMessage(title, content string) string {
	return fmt.Sprintf("✅ *%s*\n\n%s", utils.EscapeMarkdownV2(title), content)
}

// messageLang is the language of the user who sent m; anonymous senders (no
// From) get the default language.
func messageLang(s *service.BotService, m *tgbotapi.Message) string {
	if m.From == nil {
		return i18n.DefaultLang
	}
	return s.GetUserLanguage(m.From.ID)
}

// usageMessage builds the "invalid format" reply: a shared header plus the
// command-specific usage text.
func usageMessage(lang, usageKey string) string {
	return i18n.T(lang, "media_usage_header") + "\n\n" + i18n.T(lang, usageKey)
}

// uploadKeyboard is the "upload to Drive" button shown under every finished
// media operation.
func uploadKeyboard(lang, pathID string) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "media_btn_upload_drive"), "media_m:"+pathID),
		),
	)
	return &keyboard
}

// ffmpegFailure renders a localized "<operation> failed" message carrying the
// error and the tail of ffmpeg's output.
func ffmpegFailure(lang, key string, err error, output []byte) string {
	return i18n.T(lang, key,
		utils.EscapeMarkdownV2(err.Error()),
		utils.EscapeMarkdownV2Code(utils.GetLastLines(string(output), 15)))
}

func HandleExtractAudio(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	lang := messageLang(s, message)

	fileID, fileName := s.GetFileFromMessage(message)
	inputPath := args

	if fileID == "" && inputPath == "" {
		s.Reply(message, GetErrorMessage(i18n.T(lang, "media_title_invalid_format"), i18n.T(lang, "media_extract_usage")))
		return
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, i18n.T(lang, "media_extract_progress"))
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, err := s.Bot.Send(statusMsg)
	if err != nil {
		slog.Error("Failed to send status message", "error", err)
		s.Reply(message, i18n.T(lang, "media_status_send_failed"))
		return
	}

	if fileID != "" {
		var downloadErr error
		inputPath, downloadErr = s.DownloadTelegramFile(fileID, fileName)
		if downloadErr != nil {
			slog.Error("Telegram download failed during audio extraction", "error", downloadErr)
			s.EditMessage(sent.Chat.ID, sent.MessageID, i18n.T(lang, "media_download_failed_detail", utils.EscapeMarkdownV2(i18n.Message(lang, downloadErr))))
			return
		}
	}

	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".mp3"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if hasAudio, errAudio := s.HasAudioStream(inputPath); errAudio != nil || !hasAudio {
		s.EditMessage(sent.Chat.ID, sent.MessageID, GetErrorMessage(i18n.T(lang, "media_title_no_audio"), i18n.T(lang, "media_no_audio")))
		return
	}

	inputDir := filepath.Dir(inputPath)
	inputName := filepath.Base(inputPath)
	outputName := filepath.Base(outputPath)

	cmd := execCommand(ctx, "ffmpeg", "-i", inputName, "-vn", "-acodec", "libmp3lame", "-q:a", "2", outputName, "-y")
	cmd.Dir = inputDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.EditMessage(sent.Chat.ID, sent.MessageID, GetErrorMessage(i18n.T(lang, "media_title_ffmpeg_error"), ffmpegFailure(lang, "media_extract_failed", err, output)))
		return
	}

	content := i18n.T(lang, "media_done_audio", utils.EscapeMarkdownV2Code(filepath.Base(outputPath)))
	pathID := s.StorePath(outputPath)
	s.EditMessageWithMarkup(sent.Chat.ID, sent.MessageID, GetSuccessMessage(i18n.T(lang, "media_title_extract_done"), content), uploadKeyboard(lang, pathID))
}

func HandleCompressVideo(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	lang := messageLang(s, message)

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
		s.Reply(message, usageMessage(lang, "media_compress_usage"))
		return
	}

	crf := "28"
	switch quality {
	case "low":
		crf = "32"
	case "high":
		crf = "23"
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, i18n.T(lang, "media_compress_progress"))
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, err := s.Bot.Send(statusMsg)
	if err != nil {
		slog.Error("Failed to send status message", "error", err)
		s.Reply(message, i18n.T(lang, "media_status_send_failed"))
		return
	}

	if fileID != "" {
		var downloadErr error
		inputPath, downloadErr = s.DownloadTelegramFile(fileID, fileName)
		if downloadErr != nil {
			s.EditMessage(sent.Chat.ID, sent.MessageID, i18n.T(lang, "media_download_failed"))
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
		s.EditMessage(sent.Chat.ID, sent.MessageID, ffmpegFailure(lang, "media_compress_failed", err, output))
		return
	}

	content := i18n.T(lang, "media_done_compress",
		utils.EscapeMarkdownV2Code(filepath.Base(outputPath)),
		utils.EscapeMarkdownV2(quality))
	pathID := s.StorePath(outputPath)
	s.EditMessageWithMarkup(sent.Chat.ID, sent.MessageID, GetSuccessMessage(i18n.T(lang, "media_title_compress_done"), content), uploadKeyboard(lang, pathID))
}

func HandleGenerateThumbnail(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	lang := messageLang(s, message)

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
		s.Reply(message, usageMessage(lang, "media_thumbnail_usage"))
		return
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, i18n.T(lang, "media_thumbnail_progress"))
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, err := s.Bot.Send(statusMsg)
	if err != nil {
		slog.Error("Failed to send status message", "error", err)
		s.Reply(message, i18n.T(lang, "media_status_send_failed"))
		return
	}

	if fileID != "" {
		var downloadErr error
		inputPath, downloadErr = s.DownloadTelegramFile(fileID, fileName)
		if downloadErr != nil {
			s.EditMessage(sent.Chat.ID, sent.MessageID, i18n.T(lang, "media_download_failed"))
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
		s.EditMessage(sent.Chat.ID, sent.MessageID, ffmpegFailure(lang, "media_thumbnail_failed", err, output))
		return
	}

	photo := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FilePath(outputPath))
	photo.Caption = i18n.T(lang, "media_thumbnail_caption", utils.EscapeMarkdownV2Code(filepath.Base(inputPath)), utils.EscapeMarkdownV2(timestamp))
	photo.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(photo)

	_, _ = s.Bot.Request(tgbotapi.NewDeleteMessage(sent.Chat.ID, sent.MessageID))
}

func HandleEmbedSubtitle(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	lang := messageLang(s, message)

	fileID, fileName := s.GetFileFromMessage(message)
	var videoPath, subPath string

	parts := strings.Fields(args)
	if fileID != "" {
		var err error
		videoPath, err = s.DownloadTelegramFile(fileID, fileName)
		if err != nil {
			s.Reply(message, i18n.T(lang, "media_download_failed"))
			return
		}
		if len(parts) > 0 {
			subPath = parts[0]
		}
	} else {
		if len(parts) < 2 {
			s.Reply(message, usageMessage(lang, "media_subtitle_usage"))
			return
		}
		videoPath = parts[0]
		subPath = parts[1]
	}

	if videoPath == "" || subPath == "" {
		s.Reply(message, usageMessage(lang, "media_subtitle_usage"))
		return
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, i18n.T(lang, "media_subtitle_progress"))
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, err := s.Bot.Send(statusMsg)
	if err != nil {
		slog.Error("Failed to send status message", "error", err)
		s.Reply(message, i18n.T(lang, "media_status_send_failed"))
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
		s.EditMessage(sent.Chat.ID, sent.MessageID, ffmpegFailure(lang, "media_subtitle_failed", err, output))
		return
	}

	content := i18n.T(lang, "media_done_video", utils.EscapeMarkdownV2Code(filepath.Base(outputPath)))
	pathID := s.StorePath(outputPath)
	s.EditMessageWithMarkup(sent.Chat.ID, sent.MessageID, GetSuccessMessage(i18n.T(lang, "media_title_subtitle_done"), content), uploadKeyboard(lang, pathID))
}

func HandleConvertFormat(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	lang := messageLang(s, message)

	fileID, fileName := s.GetFileFromMessage(message)
	var inputPath, targetFormat string

	parts := strings.Fields(args)
	if fileID != "" {
		var err error
		inputPath, err = s.DownloadTelegramFile(fileID, fileName)
		if err != nil {
			s.Reply(message, i18n.T(lang, "media_download_failed"))
			return
		}
		if len(parts) > 0 {
			targetFormat = strings.ToLower(parts[0])
		}
	} else {
		if len(parts) < 2 {
			s.Reply(message, usageMessage(lang, "media_convert_usage"))
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
		s.Reply(message, i18n.T(lang, "media_convert_unsupported"))
		return
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, i18n.T(lang, "media_convert_progress", targetFormat))
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, err := s.Bot.Send(statusMsg)
	if err != nil {
		slog.Error("Failed to send status message", "error", err)
		s.Reply(message, i18n.T(lang, "media_status_send_failed"))
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
		s.EditMessage(sent.Chat.ID, sent.MessageID, ffmpegFailure(lang, "media_convert_failed", err, output))
		return
	}

	content := i18n.T(lang, "media_done_file", utils.EscapeMarkdownV2Code(filepath.Base(outputPath)))
	pathID := s.StorePath(outputPath)
	s.EditMessageWithMarkup(sent.Chat.ID, sent.MessageID, GetSuccessMessage(i18n.T(lang, "media_title_convert_done"), content), uploadKeyboard(lang, pathID))
}

func HandleMediaInfo(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	lang := messageLang(s, message)

	fileID, fileName := s.GetFileFromMessage(message)
	inputPath := args

	if fileID != "" {
		var err error
		inputPath, err = s.DownloadTelegramFile(fileID, fileName)
		if err != nil {
			s.Reply(message, i18n.T(lang, "media_download_failed"))
			return
		}
	}

	if inputPath == "" {
		s.Reply(message, usageMessage(lang, "media_mediainfo_usage"))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := execCommand(ctx, "ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", inputPath)
	output, err := cmd.Output()
	if err != nil {
		s.Reply(message, i18n.T(lang, "media_mediainfo_failed", utils.EscapeMarkdownV2(err.Error())))
		return
	}

	var probe service.FFProbeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		slog.Error("FFProbe JSON unmarshal failed", "error", err)
		s.Reply(message, i18n.T(lang, "media_mediainfo_parse_failed"))
		return
	}

	text := s.FormatMediaInfo(lang, filepath.Base(inputPath), probe)
	s.Reply(message, text)
}

func HandleMediaMirrorCallback(s *service.BotService, cb *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) < 2 {
		return
	}

	lang := s.GetUserLanguage(cb.From.ID)

	pathID := parts[1]
	fullPath, ok := s.GetPath(pathID)
	if !ok {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(cb.ID, i18n.T(lang, "media_toast_link_expired")))
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(cb.ID, i18n.T(lang, "media_toast_file_deleted")))
		return
	}

	_, _ = s.Bot.Request(tgbotapi.NewCallback(cb.ID, i18n.T(lang, "media_toast_mirror_starting")))

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
		s.HandleCreateTaskErrorLang(s.GetUserLanguage(cb.From.ID), cb.Message.Chat.ID, cb.Message.MessageID, err)
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

	lang := messageLang(s, message)

	fileID, fileName := s.GetFileFromMessage(message)
	count := 4
	inputPath := args

	parts := strings.Fields(args)
	if fileID != "" {
		var err error
		inputPath, err = s.DownloadTelegramFile(fileID, fileName)
		if err != nil {
			s.Reply(message, i18n.T(lang, "media_download_failed"))
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
		s.Reply(message, usageMessage(lang, "media_screenshots_usage"))
		return
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, i18n.T(lang, "media_screenshots_progress", count))
	statusMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, err := s.Bot.Send(statusMsg)
	if err != nil {
		slog.Error("Failed to send status message during screenshots", "error", err)
		s.Reply(message, i18n.T(lang, "media_status_send_failed"))
		return
	}

	screenshots, err := s.GenerateScreenshotsList(inputPath, count)
	if err != nil {
		s.EditMessage(sent.Chat.ID, sent.MessageID, i18n.T(lang, "media_screenshots_failed", utils.EscapeMarkdownV2(i18n.Message(lang, err))))
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

	lang := messageLang(s, message)

	fileID, fileName := s.GetFileFromMessage(message)
	var videoPath, subPath string

	parts := strings.Fields(args)
	if fileID != "" {
		var err error
		videoPath, err = s.DownloadTelegramFile(fileID, fileName)
		if err != nil {
			s.Reply(message, i18n.T(lang, "media_download_failed"))
			return
		}
		if len(parts) > 0 {
			subPath = parts[0]
		}
	} else {
		if len(parts) < 2 {
			s.Reply(message, usageMessage(lang, "media_hardsub_usage"))
			return
		}
		videoPath = parts[0]
		subPath = parts[1]
	}

	if videoPath == "" || subPath == "" {
		s.Reply(message, usageMessage(lang, "media_hardsub_usage"))
		return
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, i18n.T(lang, "media_hardsub_progress"))
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
		s.EditMessage(sent.Chat.ID, sent.MessageID, ffmpegFailure(lang, "media_hardsub_failed", err, output))
		return
	}

	content := i18n.T(lang, "media_done_video", utils.EscapeMarkdownV2Code(filepath.Base(outputPath)))
	pathID := s.StorePath(outputPath)
	s.EditMessageWithMarkup(sent.Chat.ID, sent.MessageID, GetSuccessMessage(i18n.T(lang, "media_title_hardsub_done"), content), uploadKeyboard(lang, pathID))
}

func HandleRescale(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	lang := messageLang(s, message)

	fileID, fileName := s.GetFileFromMessage(message)
	var inputPath, resolution string

	parts := strings.Fields(args)
	if fileID != "" {
		if len(parts) == 0 {
			s.Reply(message, usageMessage(lang, "media_rescale_usage_reply"))
			return
		}
		var err error
		inputPath, err = s.DownloadTelegramFile(fileID, fileName)
		if err != nil {
			s.Reply(message, i18n.T(lang, "media_download_failed"))
			return
		}
		resolution = parts[0]
	} else {
		if len(parts) < 2 {
			s.Reply(message, usageMessage(lang, "media_rescale_usage_path"))
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
			s.Reply(message, i18n.T(lang, "media_rescale_invalid"))
			return
		}
	}

	statusMsg := tgbotapi.NewMessage(message.Chat.ID, i18n.T(lang, "media_rescale_progress", utils.EscapeMarkdownV2(resolution)))
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
		s.EditMessage(sent.Chat.ID, sent.MessageID, ffmpegFailure(lang, "media_rescale_failed", err, output))
		return
	}

	content := i18n.T(lang, "media_done_rescale",
		utils.EscapeMarkdownV2Code(filepath.Base(outputPath)),
		utils.EscapeMarkdownV2(resolution))
	pathID := s.StorePath(outputPath)
	s.EditMessageWithMarkup(sent.Chat.ID, sent.MessageID, GetSuccessMessage(i18n.T(lang, "media_title_rescale_done"), content), uploadKeyboard(lang, pathID))
}
