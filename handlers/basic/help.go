package basic

import (
	"log/slog"
	"strings"
	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleHelpCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, action string) {
	if action == "" {
		parts := strings.SplitN(callback.Data, ":", 2)
		if len(parts) == 2 {
			action = parts[1]
		} else if callback.Data == "help" {
			action = "main"
		}
	}
	action = normalizeHelpAction(action)

	if handleDownloadHelp(s, callback, action) {
		return
	}
	if handleMonitorHelp(s, callback, action) {
		return
	}
	if handleFilesHelp(s, callback, action) {
		return
	}
	if handleMediaHelp(s, callback, action) {
		return
	}
	if handleTaskHelp(s, callback, action) {
		return
	}
	if handleStorageHelp(s, callback, action) {
		return
	}
	if handleAdminHelp(s, callback, action) {
		return
	}
	if handleRecoveryHelp(s, callback, action) {
		return
	}

	var text string
	var keyboard tgbotapi.InlineKeyboardMarkup
	lang := s.GetUserLanguage(callback.From.ID)
	backToHelp := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_back"), "help:main"),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_close"), "help:"+CmdClose),
	)

	switch action {
	case "main":
		text = service.GetHelpMainText(lang)
		keyboard = service.GetHelpKeyboard(lang)

	case "settings":
		text = GetHelpSettings()
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_settings"), "settings:open"),
			),
			backToHelp,
		)

	case "all":
		text = getHelpAllCommands()
		keyboard = tgbotapi.NewInlineKeyboardMarkup(backToHelp)

	case ActionBack:
		content := service.GetWelcomeMessage(lang, callback.From.FirstName)
		kb := service.GetStartKeyboard(lang)
		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, content)
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
		editMsg.ReplyMarkup = &kb
		_, _ = s.Bot.Send(editMsg)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		return

	case CmdClose:
		deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
		_, _ = s.Bot.Request(deleteMsg)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "Đã đóng"))
		return

	default:
		return
	}

	sendHelpMessage(s, callback, text, keyboard)
}

func normalizeHelpAction(action string) string {
	switch action {
	case "cat_download", "sub_dl_general", "sub_dl_video", "sub_dl_torrent", "sub_dl_adv":
		return "download"
	case "cat_monitor":
		return "monitor"
	case "cat_files":
		return "files"
	case "cat_media":
		return "media"
	case "cat_task":
		return "task"
	case "cat_storage":
		return "storage"
	case "cat_admin":
		return "admin"
	case "cat_recovery":
		return "recovery"
	case "cat_settings", "cmd_lang":
		return "settings"
	default:
		return action
	}
}

func sendHelpMessage(s *service.BotService, callback *tgbotapi.CallbackQuery, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	if len(keyboard.InlineKeyboard) > 0 {
		editMsg.ReplyMarkup = &keyboard
	}
	_, err := s.Bot.Send(editMsg)
	if err != nil {
		slog.Error("Error editing help message", "error", err)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "❌ Lỗi khi tải menu"))
	} else {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
	}
}

type helpButton struct {
	label, action string
}

// helpCategoryKeyboard lists a category's commands two per row, then a back-to-help row.
func helpCategoryKeyboard(lang string, buttons []helpButton) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(buttons); i += 2 {
		row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(buttons[i].label, "help:"+buttons[i].action))
		if i+1 < len(buttons) {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(buttons[i+1].label, "help:"+buttons[i+1].action))
		}
		rows = append(rows, row)
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_back"), "help:main"),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_close"), "help:"+CmdClose),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// helpCommandKeyboard is shown under a single command's help: back to its category, or close.
func helpCommandKeyboard(lang, category string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_back"), "help:"+category),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_close"), "help:"+CmdClose),
	))
}

func sendHelpPage(s *service.BotService, callback *tgbotapi.CallbackQuery, category, action, text string, buttons []helpButton) {
	lang := s.GetUserLanguage(callback.From.ID)
	keyboard := helpCommandKeyboard(lang, category)
	if action == category {
		keyboard = helpCategoryKeyboard(lang, buttons)
	}
	sendHelpMessage(s, callback, text, keyboard)
}

var downloadHelpButtons = []helpButton{
	{"📥 Mirror", "cmd_mirror"}, {"📤 Leech", "cmd_leech"},
	{"🎬 YT Mirror", "cmd_ytdlp"}, {"🎬 YT Leech", "cmd_ytdlpleech"},
	{"⚔️ Viking", "cmd_viking"}, {"🧲 Torrent", "cmd_torrent"},
	{"📋 Clone", "cmd_clone"}, {"📦 Batch", "cmd_batch"},
	{"🔍 Search", "cmd_search"},
}

func handleDownloadHelp(s *service.BotService, callback *tgbotapi.CallbackQuery, action string) bool {
	var text string

	switch action {
	case "download":
		text = service.ProfessionalMessage("📥 TẢI XUỐNG", "Chọn lệnh để xem chi tiết:")
	case "cmd_mirror":
		text = GetHelpMirror()
	case "cmd_leech":
		text = GetHelpLeech()
	case "cmd_ytdlp":
		text = GetHelpYTDLP()
	case "cmd_ytdlpleech":
		text = GetHelpYTDLPLeech()
	case "cmd_viking":
		text = GetHelpViking()
	case "cmd_torrent":
		text = GetHelpTorrent()
	case "cmd_clone":
		text = GetHelpClone()
	case "cmd_batch":
		text = GetHelpBatch()
	case "cmd_search":
		text = GetHelpSearch()
	default:
		return false
	}
	sendHelpPage(s, callback, "download", action, text, downloadHelpButtons)
	return true
}

var monitorHelpButtons = []helpButton{
	{"📊 Status", "cmd_status"}, {"📈 Stats", "cmd_stats"},
	{"🖥️ System", "cmd_system"}, {"🏥 Health", "cmd_health"},
	{"📜 Logs", "cmd_logs"}, {"🏓 Ping", "cmd_ping"},
	{"🚀 Speed", "cmd_speed"},
}

func handleMonitorHelp(s *service.BotService, callback *tgbotapi.CallbackQuery, action string) bool {
	var text string

	switch action {
	case "monitor":
		text = service.ProfessionalMessage("📊 THEO DÕI", "Chọn lệnh để xem chi tiết:")
	case "cmd_status":
		text = GetHelpStatus()
	case "cmd_stats":
		text = GetHelpStats()
	case "cmd_system":
		text = GetHelpSystem()
	case "cmd_health":
		text = GetHelpHealth()
	case "cmd_logs":
		text = GetHelpLogs()
	case "cmd_ping":
		text = GetHelpPing()
	case "cmd_speed":
		text = GetHelpSpeed()
	default:
		return false
	}
	sendHelpPage(s, callback, "monitor", action, text, monitorHelpButtons)
	return true
}

var filesHelpButtons = []helpButton{
	{"📂 List (ls)", "cmd_ls"}, {"📁 Mkdir", "cmd_mkdir"},
	{"🗑️ Remove", "cmd_rm"}, {"📦 Move", "cmd_mv"},
	{"🔗 Share", "cmd_share"}, {"🔍 Find", "cmd_find"},
}

func handleFilesHelp(s *service.BotService, callback *tgbotapi.CallbackQuery, action string) bool {
	var text string

	switch action {
	case "files":
		text = service.ProfessionalMessage("📁 FILE", "Chọn lệnh để xem chi tiết:")
	case "cmd_ls":
		text = GetHelpLs()
	case "cmd_mkdir":
		text = GetHelpMkdir()
	case "cmd_rm":
		text = GetHelpRm()
	case "cmd_mv":
		text = GetHelpMv()
	case "cmd_share":
		text = GetHelpShare()
	case "cmd_find":
		text = GetHelpFind()
	default:
		return false
	}
	sendHelpPage(s, callback, "files", action, text, filesHelpButtons)
	return true
}

var mediaHelpButtons = []helpButton{
	{"🎵 Extract Audio", "cmd_extractaudio"}, {"🗜️ Compress", "cmd_compress"},
	{"🖼️ Thumbnail", "cmd_thumbnail"}, {"📸 Screenshots", "cmd_screenshots"},
	{"💬 Soft-sub", "cmd_subtitle"}, {"🔥 Hard-sub", "cmd_hardsub"},
	{"📐 Rescale", "cmd_rescale"}, {"🔄 Convert", "cmd_convert"},
	{"ℹ️ Media Info", "cmd_mediainfo"},
}

func handleMediaHelp(s *service.BotService, callback *tgbotapi.CallbackQuery, action string) bool {
	var text string

	switch action {
	case "media":
		text = service.ProfessionalMessage("🎵 MEDIA", "Chọn lệnh để xem chi tiết:")
	case "cmd_extractaudio":
		text = GetHelpExtractAudio()
	case "cmd_compress":
		text = GetHelpCompress()
	case "cmd_thumbnail":
		text = GetHelpThumbnail()
	case "cmd_screenshots":
		text = GetHelpScreenshots()
	case "cmd_subtitle":
		text = GetHelpSubtitle()
	case "cmd_hardsub":
		text = GetHelpHardsub()
	case "cmd_rescale":
		text = GetHelpRescale()
	case "cmd_convert":
		text = GetHelpConvert()
	case "cmd_mediainfo":
		text = GetHelpMediaInfo()
	default:
		return false
	}
	sendHelpPage(s, callback, "media", action, text, mediaHelpButtons)
	return true
}

var taskHelpButtons = []helpButton{
	{"📊 Status", "cmd_status"}, {"❌ Cancel", "cmd_cancel"},
	{"🚫 Cancel All", "cmd_cancelall"},
}

func handleTaskHelp(s *service.BotService, callback *tgbotapi.CallbackQuery, action string) bool {
	var text string

	switch action {
	case "task":
		text = service.ProfessionalMessage("📋 TASK", "Chọn lệnh để xem chi tiết:")
	case "cmd_cancel":
		text = GetHelpCancel()
	case "cmd_cancelall":
		text = GetHelpCancelAll()
	default:
		return false
	}
	sendHelpPage(s, callback, "task", action, text, taskHelpButtons)
	return true
}

var storageHelpButtons = []helpButton{
	{"📋 Storages", "cmd_storages"}, {"⚙️ Set Storage", "cmd_setstorage"},
}

func handleStorageHelp(s *service.BotService, callback *tgbotapi.CallbackQuery, action string) bool {
	var text string

	switch action {
	case "storage":
		text = service.ProfessionalMessage("💾 LƯU TRỮ", "Chọn lệnh để xem chi tiết:")
	case "cmd_storages":
		text = GetHelpStorages()
	case "cmd_setstorage":
		text = GetHelpSetStorage()
	default:
		return false
	}
	sendHelpPage(s, callback, "storage", action, text, storageHelpButtons)
	return true
}

var adminHelpButtons = []helpButton{
	{"✅ Authorize", "cmd_authorize"}, {"❌ Unauthorize", "cmd_unauthorize"},
	{"👥 Users", "cmd_users"}, {"📜 Log Channel", "cmd_setlogchannel"},
	{"🚨 Alert Channel", "cmd_setalertchannel"},
}

func handleAdminHelp(s *service.BotService, callback *tgbotapi.CallbackQuery, action string) bool {
	var text string

	switch action {
	case "admin":
		text = service.ProfessionalMessage("👑 QUẢN TRỊ", "Chọn lệnh để xem chi tiết:")
	case "cmd_authorize":
		text = GetHelpAuthorize()
	case "cmd_unauthorize":
		text = GetHelpUnauthorize()
	case "cmd_users":
		text = GetHelpUsers()
	case "cmd_setalertchannel":
		text = GetHelpSetAlertChannel()
	case "cmd_setlogchannel":
		text = GetHelpSetLogChannel()
	default:
		return false
	}
	sendHelpPage(s, callback, "admin", action, text, adminHelpButtons)
	return true
}

var recoveryHelpButtons = []helpButton{
	{"🔄 Recover", "cmd_recover"}, {"📊 Status", "cmd_recoverystatus"},
}

func handleRecoveryHelp(s *service.BotService, callback *tgbotapi.CallbackQuery, action string) bool {
	var text string

	switch action {
	case "recovery":
		text = service.ProfessionalMessage("🔧 KHÔI PHỤC", "Chọn lệnh để xem chi tiết:")
	case "cmd_recover":
		text = GetHelpRecover()
	case "cmd_recoverystatus":
		text = GetHelpRecoveryStatus()
	default:
		return false
	}
	sendHelpPage(s, callback, "recovery", action, text, recoveryHelpButtons)
	return true
}
