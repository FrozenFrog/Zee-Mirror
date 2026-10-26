package service

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"zee-mirror/internal/domain"
	"zee-mirror/internal/metrics"
	"zee-mirror/internal/organizer"
	"zee-mirror/internal/uploader"
	"zee-mirror/pkg/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (s *BotService) UploadWithRclone(task *Task) error {
	task.SetStatus(StatusUploading)
	task.SetProgress(0)
	s.updateTaskStatus(task)

	lang := s.GetUserLanguage(task.UserID)
	err := s.RcloneUploader.Upload(task.Ctx, &task.Task, func(up uploader.ProgressUpdate) {
		task.UpdateFromUploadProgress(lang, up)
		s.updateTaskStatus(task)
	})

	if err != nil {
		return err
	}

	task.SetProgress(100)
	return nil
}

func (s *BotService) UploadToTelegram(task *Task) error {
	task.SetStatus(StatusUploading)
	task.SetProgress(0)
	s.updateTaskStatus(task)
	startTime := time.Now()

	filePath := task.LocalPath
	if filePath == "" {
		return i18n.WrapError(domain.ErrInvalidInput, "upload_err_no_local_file")
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return i18n.WrapError(domain.ErrStorage, "upload_err_stat", err)
	}

	if info.IsDir() {
		return i18n.WrapError(domain.ErrInvalidInput, "upload_err_tg_dir")
	}

	if info.Size() > domain.MaxTelegramUploadSize {
		return i18n.WrapError(domain.ErrInvalidInput, "upload_err_tg_too_large")
	}

	var msg tgbotapi.Chattable
	if organizer.IsVideoFile(filePath) {
		video := tgbotapi.NewVideo(task.ChatID, tgbotapi.FilePath(filePath))
		video.Caption = fmt.Sprintf("📄 %s", task.FileName)

		if thumb, errThumb := GenerateThumbnail(task.Ctx, filePath, s.TaskManager.DownloadDir); errThumb == nil {
			video.Thumb = tgbotapi.FilePath(thumb)
			defer func() {
				if errRemove := os.Remove(thumb); errRemove != nil {
					slog.Warn("Failed to remove thumbnail", "error", errRemove, "path", thumb)
				}
			}()
		}

		msg = video
	} else {
		doc := tgbotapi.NewDocument(task.ChatID, tgbotapi.FilePath(filePath))
		doc.Caption = fmt.Sprintf("📄 %s", task.FileName)
		msg = doc
	}

	task.SetProgress(50)
	s.updateTaskStatus(task)

	sentMsg, err := s.Bot.Send(msg)
	if err != nil {
		metrics.UploadDuration.WithLabelValues("telegram", "failed").Observe(time.Since(startTime).Seconds())
		return i18n.WrapError(domain.ErrExternal, "upload_err_tg_failed", err)
	}
	metrics.UploadDuration.WithLabelValues("telegram", "success").Observe(time.Since(startTime).Seconds())

	var fileID, tgFilePath string
	switch {
	case sentMsg.Document != nil:
		fileID = sentMsg.Document.FileID
	case sentMsg.Video != nil:
		fileID = sentMsg.Video.FileID
	case sentMsg.Audio != nil:
		fileID = sentMsg.Audio.FileID
	case sentMsg.Animation != nil:
		fileID = sentMsg.Animation.FileID
	default:
		if len(sentMsg.Photo) > 0 {
			fileID = sentMsg.Photo[0].FileID
		}
	}
	if fileID != "" {
		if tgFile, err := s.Bot.GetFile(tgbotapi.FileConfig{FileID: fileID}); err == nil {
			tgFilePath = tgFile.FilePath
		} else {
			slog.Warn("Failed to get Telegram file path", "error", err, "fileID", fileID)
		}
	}

	task.CompleteTelegramUpload(sentMsg.MessageID, info.Size(), fileID, tgFilePath)

	return nil
}
