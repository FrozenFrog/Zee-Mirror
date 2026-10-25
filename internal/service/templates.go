package service

import (
	"fmt"
	"html"
	"net/url"
	"os"
	"strings"
	"time"
	"zee-mirror/internal/domain"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	LineSeparator    = "━━━━━━━━━━━━━━━━━━━━━━"
	CompactSeparator = "──────────────────────"
	RepoURL          = "https://github.com/ifauzeee/Zee-Mirror"
	UnknownSize      = "Không rõ"
)

func buildTaskStatusText(lang string, snapshot domain.TaskSnapshot, baseURL string) string {
	var text string
	switch snapshot.Status {
	case domain.StatusCompleted:
		duration := calculateDuration(snapshot)
		sizeStr := determineSizeString(snapshot)

		playlistLine := ""
		if snapshot.PlaylistCount > 0 {
			playlistLine = fmt.Sprintf("\n📋 *Playlist:* \\[%d/%d\\]", snapshot.PlaylistIndex, snapshot.PlaylistCount)
		}

		md5Line := ""
		if snapshot.MD5 != "" {
			md5Line = fmt.Sprintf("\n🔐 *MD5:* `%s`", utils.EscapeMarkdownV2Code(snapshot.MD5))
		}

		streamLine := ""
		if baseURL != "" && snapshot.FileName != "" {
			streamURL := strings.TrimRight(baseURL, "/") + "/stream/" + url.PathEscape(snapshot.FileName)
			streamLine = fmt.Sprintf("\n🌊 *Stream:* `%s`", utils.EscapeMarkdownV2Code(streamURL))
		}

		text = fmt.Sprintf("%s\n\n"+
			"📄 *Tên:* `%s`\n"+
			"📦 *Dung lượng:* `%s`\n"+
			"⏱ *Thời gian:* `%s`\n"+
			"📁 *Đường dẫn:* `%s`%s%s%s",
			i18n.T(lang, "status_completed"),
			utils.EscapeMarkdownV2Code(snapshot.FileName),
			utils.EscapeMarkdownV2Code(sizeStr),
			utils.EscapeMarkdownV2Code(utils.FormatDuration(duration)),
			utils.EscapeMarkdownV2Code(snapshot.RemotePath),
			playlistLine,
			md5Line,
			streamLine)
	case domain.StatusFailed:
		text = fmt.Sprintf("%s\n📄 `%s`\nLỗi: `%s`",
			i18n.T(lang, "status_failed"),
			utils.EscapeMarkdownV2Code(snapshot.FileName),
			utils.EscapeMarkdownV2Code(utils.TruncateString(snapshot.Error, 100)))
	default:
		return ""
	}
	return text
}

func calculateDuration(snapshot domain.TaskSnapshot) time.Duration {
	duration := snapshot.CompletedAt.Sub(snapshot.StartedAt)
	if snapshot.StartedAt.IsZero() {
		duration = snapshot.CompletedAt.Sub(snapshot.CreatedAt)
	}
	return duration
}

func determineSizeString(snapshot domain.TaskSnapshot) string {
	sizeStr := UnknownSize

	if snapshot.LocalPath != "" {
		if info, err := os.Stat(snapshot.LocalPath); err == nil && info.IsDir() {
			if dirSize, err := utils.CalculateDirSize(snapshot.LocalPath); err == nil && dirSize > 0 {
				sizeStr = utils.FormatBytes(dirSize)
			} else {
				if snapshot.TotalSize > 0 {
					sizeStr = utils.FormatBytes(snapshot.TotalSize)
				} else if snapshot.DownloadedSize > 0 {
					sizeStr = utils.FormatBytes(snapshot.DownloadedSize)
				}
			}
		} else {
			if snapshot.TotalSize > 0 {
				sizeStr = utils.FormatBytes(snapshot.TotalSize)
			} else if snapshot.DownloadedSize > 0 {
				sizeStr = utils.FormatBytes(snapshot.DownloadedSize)
			}
		}
	} else {
		if snapshot.TotalSize > 0 {
			sizeStr = utils.FormatBytes(snapshot.TotalSize)
		} else if snapshot.DownloadedSize > 0 {
			sizeStr = utils.FormatBytes(snapshot.DownloadedSize)
		}
	}

	return sizeStr
}

func ProfessionalMessage(title string, content string) string {
	return fmt.Sprintf("✨ *%s* ✨\n%s\n\n%s\n\n%s",
		utils.EscapeMarkdownV2(title),
		LineSeparator,
		content,
		LineSeparator)
}

// HelpDetailMessage builds a help page. The sections are MarkdownV2 (already
// escaped, with their own bold and code formatting); only title is escaped.
func HelpDetailMessage(title, kegunaan, caraPakai, contoh, extra string) string {
	content := ""

	if kegunaan != "" {
		content += "📝 *CÔNG DỤNG*\n" + kegunaan + "\n\n"
	}

	if caraPakai != "" {
		content += "📌 *CÁCH DÙNG*\n" + caraPakai + "\n\n"
	}

	if contoh != "" {
		content += "💡 *VÍ DỤ*\n" + contoh
	}

	if extra != "" {
		content += "\n\n" + extra
	}

	return ProfessionalMessage(title, content)
}

func GetWelcomeMessage(lang, userName string) string {
	content := i18n.T(lang, "welcome_content", utils.EscapeMarkdownV2(userName))
	return ProfessionalMessage(i18n.T(lang, "welcome_title"), content)
}

func GetStatusHeader() string {
	return fmt.Sprintf("📊 <b>TRẠNG THÁI TASK ĐANG CHẠY</b>\n%s\n\n", LineSeparator)
}

func FormatTaskProfessional(lang string, taskSnapshot domain.TaskSnapshot) string {
	emoji := utils.StatusEmoji(string(taskSnapshot.Status))
	bar := utils.ProgressBar(taskSnapshot.Progress, 10)

	processedSize := taskSnapshot.DownloadedSize
	if taskSnapshot.Status == domain.StatusUploading {
		processedSize = taskSnapshot.UploadedSize
	}

	totalSizeStr := utils.FormatBytes(taskSnapshot.TotalSize)
	if taskSnapshot.TotalSize == 0 {
		totalSizeStr = UnknownSize
	}

	msgLine := ""
	if taskSnapshot.ProcessingMessage != "" {
		msgLine = fmt.Sprintf("💬 <b>Trạng thái:</b> <i>%s</i>\n", html.EscapeString(taskSnapshot.ProcessingMessage))
	} else if taskSnapshot.Status == domain.StatusDownloading && taskSnapshot.Progress >= 100 {
		msgLine = "💬 <b>Trạng thái:</b> <i>Đang xử lý file...</i>\n"
	}

	playlistInfo := ""
	if taskSnapshot.PlaylistCount > 0 {
		playlistInfo = fmt.Sprintf("📋 <b>Playlist:</b> [%d/%d]\n", taskSnapshot.PlaylistIndex, taskSnapshot.PlaylistCount)
	}

	return fmt.Sprintf(
		"🏷️ <b>ID:</b> <code>%s</code> • %s <b>%s</b>\n"+
			"%s\n"+
			"%s"+
			"%s"+
			"📄 <b>File:</b> <code>%s</code>\n"+
			"📦 <b>Dung lượng:</b> <code>%s / %s</code>\n"+
			"⚡ <b>Tốc độ:</b> <code>%s</code> • ⏱️ <b>Còn lại:</b> <code>%s</code>\n"+
			"🚫 <b>Hủy:</b> /cancel_%s\n",
		html.EscapeString(taskSnapshot.ID),
		emoji,
		html.EscapeString(i18n.T(lang, strings.ToLower(string(taskSnapshot.Status)))),
		bar,
		msgLine,
		playlistInfo,
		html.EscapeString(utils.TruncateString(taskSnapshot.FileName, 35)),
		html.EscapeString(utils.FormatBytes(processedSize)),
		html.EscapeString(totalSizeStr),
		html.EscapeString(utils.FormatSpeed(taskSnapshot.Speed)),
		html.EscapeString(utils.FormatDuration(taskSnapshot.ETA)),
		html.EscapeString(taskSnapshot.ID),
	)
}

func GetStartKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_menu"), "help:main"),
		),
	)
}

func GetHelpMainText(lang string) string {
	content := i18n.T(lang, "help_content")
	return ProfessionalMessage(i18n.T(lang, "help_title"), content)
}

func GetHelpKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_download"), "help:download"),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_monitor"), "help:monitor"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_files"), "help:files"),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_media"), "help:media"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_task"), "help:task"),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_storage"), "help:storage"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_admin"), "help:admin"),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_recovery"), "help:recovery"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_settings"), "help:settings"),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_all"), "help:all"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_home"), "help:back"),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_close"), "help:close"),
		),
	)
}
