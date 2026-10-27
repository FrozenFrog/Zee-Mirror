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
	RepoURL          = "https://github.com/FrozenFrog/Zee-Mirror"
)

func buildTaskStatusText(lang string, snapshot domain.TaskSnapshot, baseURL string) string {
	var text string
	switch snapshot.Status {
	case domain.StatusCompleted:
		duration := calculateDuration(snapshot)
		sizeStr := determineSizeString(lang, snapshot)
		originalLink := fmt.Sprintf("<a href=\"%s\">🔗 *Original Link:*</a>", utils.EscapeMarkdownV2Code(snapshot.URL))

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

		text = fmt.Sprintf("%s\n\n%s\n%s\n%s\n%s\n%s%s%s%s",
			i18n.T(lang, "status_completed"),
			i18n.T(lang, "task_done_name", utils.EscapeMarkdownV2Code(snapshot.FileName)),
			i18n.T(lang, "task_done_size", utils.EscapeMarkdownV2Code(sizeStr)),
			originalLink,
			i18n.T(lang, "task_done_time", utils.EscapeMarkdownV2Code(utils.FormatDuration(duration))),
			i18n.T(lang, "task_done_path", utils.EscapeMarkdownV2Code(snapshot.RemotePath)),
			playlistLine,
			md5Line,
			streamLine)
	case domain.StatusFailed:
		text = fmt.Sprintf("%s\n📄 `%s`\n%s",
			i18n.T(lang, "status_failed"),
			utils.EscapeMarkdownV2Code(snapshot.FileName),
			i18n.T(lang, "task_failed_error",
				utils.EscapeMarkdownV2Code(utils.TruncateString(keyOrText(lang, snapshot.Error), 100))))
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

func determineSizeString(lang string, snapshot domain.TaskSnapshot) string {
	sizeStr := UnknownSizeLang(lang)

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
	return HelpDetailMessageLang(i18n.DefaultLang, title, kegunaan, caraPakai, contoh, extra)
}

// HelpDetailMessageLang is HelpDetailMessage with section headers in lang.
func HelpDetailMessageLang(lang, title, kegunaan, caraPakai, contoh, extra string) string {
	content := ""

	if kegunaan != "" {
		content += i18n.T(lang, "help_section_purpose") + "\n" + kegunaan + "\n\n"
	}

	if caraPakai != "" {
		content += i18n.T(lang, "help_section_usage") + "\n" + caraPakai + "\n\n"
	}

	if contoh != "" {
		content += i18n.T(lang, "help_section_example") + "\n" + contoh
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

// GetStatusHeader is GetStatusHeaderLang in DefaultLang.
func GetStatusHeader() string {
	return GetStatusHeaderLang(i18n.DefaultLang)
}

// GetStatusHeaderLang is the HTML header of the status dashboard.
func GetStatusHeaderLang(lang string) string {
	return fmt.Sprintf("%s\n%s\n\n", i18n.T(lang, "status_dashboard_header"), LineSeparator)
}

// UnknownSizeLang is the placeholder for an unknown size in lang (plain text).
func UnknownSizeLang(lang string) string {
	return i18n.T(lang, "unit_unknown_size")
}

// StatusLabel names a task status in lang (plain text); unknown statuses are
// shown capitalized.
func StatusLabel(lang string, status domain.TaskStatus) string {
	if s, ok := i18n.Lookup(lang, "task_state_"+strings.ToLower(string(status))); ok {
		return s
	}
	return utils.FormatStatus(string(status))
}

// keyOrText renders s in lang when it is a catalog key, so downloaders and
// uploaders without a user context can report progress messages and errors
// as keys (plain-text ones, without format arguments). Any other text is
// returned unchanged.
func keyOrText(lang, s string) string {
	if s == "" || strings.ContainsAny(s, " \n") {
		return s
	}
	if msg, ok := i18n.Lookup(lang, s); ok {
		return msg
	}
	return s
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
		totalSizeStr = UnknownSizeLang(lang)
	}

	msgLine := ""
	if taskSnapshot.ProcessingMessage != "" {
		msgLine = i18n.T(lang, "task_card_status", html.EscapeString(keyOrText(lang, taskSnapshot.ProcessingMessage))) + "\n"
	} else if taskSnapshot.Status == domain.StatusDownloading && taskSnapshot.Progress >= 100 {
		msgLine = i18n.T(lang, "task_card_status", html.EscapeString(i18n.T(lang, "task_card_processing_file"))) + "\n"
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
			"%s\n%s\n%s\n%s\n",
		html.EscapeString(taskSnapshot.ID),
		emoji,
		html.EscapeString(StatusLabel(lang, taskSnapshot.Status)),
		bar,
		msgLine,
		playlistInfo,
		i18n.T(lang, "task_card_file", html.EscapeString(utils.TruncateString(taskSnapshot.FileName, 35))),
		i18n.T(lang, "task_card_size",
			html.EscapeString(utils.FormatBytes(processedSize)),
			html.EscapeString(totalSizeStr)),
		i18n.T(lang, "task_card_speed_eta",
			html.EscapeString(utils.FormatSpeed(taskSnapshot.Speed)),
			html.EscapeString(utils.FormatDuration(taskSnapshot.ETA))),
		i18n.T(lang, "task_card_cancel", html.EscapeString(cancelCommandID(taskSnapshot.ID))),
	)
}

// cancelCommandID converts separators unsupported by Telegram bot-command
// entities while keeping the task ID recognizable to the user.
func cancelCommandID(taskID string) string {
	return strings.ReplaceAll(taskID, "-", "_")
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
