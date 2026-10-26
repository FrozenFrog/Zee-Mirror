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
	lang := s.GetUserLanguage(callback.From.ID)

	if buttons, ok := helpCategoryButtons[action]; ok {
		text := service.ProfessionalMessage(i18n.T(lang, "help_"+action), i18n.T(lang, "helptext_choose_command"))
		sendHelpMessage(s, callback, lang, text, helpCategoryKeyboard(lang, buttons))
		return
	}
	if page, ok := helpCommandPages[action]; ok {
		sendHelpMessage(s, callback, lang, page.render(lang), helpCommandKeyboard(lang, page.category))
		return
	}

	var text string
	var keyboard tgbotapi.InlineKeyboardMarkup
	backToHelp := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_back"), "help:main"),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_close"), "help:"+CmdClose),
	)

	switch action {
	case "main":
		text = service.GetHelpMainText(lang)
		keyboard = service.GetHelpKeyboard(lang)

	case "settings":
		text = GetHelpSettings(lang)
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_settings"), "settings:open"),
			),
			backToHelp,
		)

	case "all":
		text = getHelpAllCommands(lang)
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
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "helptext_closed")))
		return

	default:
		return
	}

	sendHelpMessage(s, callback, lang, text, keyboard)
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

func sendHelpMessage(s *service.BotService, callback *tgbotapi.CallbackQuery, lang, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	if len(keyboard.InlineKeyboard) > 0 {
		editMsg.ReplyMarkup = &keyboard
	}
	_, err := s.Bot.Send(editMsg)
	if err != nil {
		slog.Error("Error editing help message", "error", err)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "helptext_menu_error")))
	} else {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
	}
}

// helpButton opens a command's help page; its label is the catalog key
// helptext_btn_<command>.
type helpButton struct {
	action string
}

func (b helpButton) label(lang string) string {
	return i18n.T(lang, "helptext_btn_"+strings.TrimPrefix(b.action, "cmd_"))
}

func helpButtons(actions ...string) []helpButton {
	buttons := make([]helpButton, len(actions))
	for i, a := range actions {
		buttons[i] = helpButton{action: a}
	}
	return buttons
}

// helpCategoryKeyboard lists a category's commands two per row, then a back-to-help row.
func helpCategoryKeyboard(lang string, buttons []helpButton) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(buttons); i += 2 {
		row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(buttons[i].label(lang), "help:"+buttons[i].action))
		if i+1 < len(buttons) {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(buttons[i+1].label(lang), "help:"+buttons[i+1].action))
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

var (
	downloadHelpButtons = helpButtons("cmd_mirror", "cmd_leech", "cmd_ytdlp", "cmd_ytdlpleech",
		"cmd_viking", "cmd_torrent", "cmd_clone", "cmd_batch", "cmd_search")
	monitorHelpButtons = helpButtons("cmd_status", "cmd_stats", "cmd_system", "cmd_health",
		"cmd_logs", "cmd_ping", "cmd_speed")
	filesHelpButtons = helpButtons("cmd_ls", "cmd_mkdir", "cmd_rm", "cmd_mv", "cmd_share", "cmd_find")
	mediaHelpButtons = helpButtons("cmd_extractaudio", "cmd_compress", "cmd_thumbnail", "cmd_screenshots",
		"cmd_subtitle", "cmd_hardsub", "cmd_rescale", "cmd_convert", "cmd_mediainfo")
	taskHelpButtons     = helpButtons("cmd_status", "cmd_cancel", "cmd_cancelall")
	storageHelpButtons  = helpButtons("cmd_storages", "cmd_setstorage")
	adminHelpButtons    = helpButtons("cmd_authorize", "cmd_unauthorize", "cmd_users", "cmd_setlogchannel", "cmd_setalertchannel")
	recoveryHelpButtons = helpButtons("cmd_recover", "cmd_recoverystatus")
)

// helpCategoryButtons maps a help category (also its help_<category> title
// key) to the commands it lists.
var helpCategoryButtons = map[string][]helpButton{
	"download": downloadHelpButtons,
	"monitor":  monitorHelpButtons,
	"files":    filesHelpButtons,
	"media":    mediaHelpButtons,
	"task":     taskHelpButtons,
	"storage":  storageHelpButtons,
	"admin":    adminHelpButtons,
	"recovery": recoveryHelpButtons,
}

type helpCommandPage struct {
	category string // the page's back button returns to this category
	render   func(lang string) string
}

var helpCommandPages = map[string]helpCommandPage{
	"cmd_mirror":          {"download", GetHelpMirror},
	"cmd_leech":           {"download", GetHelpLeech},
	"cmd_ytdlp":           {"download", GetHelpYTDLP},
	"cmd_ytdlpleech":      {"download", GetHelpYTDLPLeech},
	"cmd_viking":          {"download", GetHelpViking},
	"cmd_torrent":         {"download", GetHelpTorrent},
	"cmd_clone":           {"download", GetHelpClone},
	"cmd_batch":           {"download", GetHelpBatch},
	"cmd_search":          {"download", GetHelpSearch},
	"cmd_status":          {"monitor", GetHelpStatus},
	"cmd_stats":           {"monitor", GetHelpStats},
	"cmd_system":          {"monitor", GetHelpSystem},
	"cmd_health":          {"monitor", GetHelpHealth},
	"cmd_logs":            {"monitor", GetHelpLogs},
	"cmd_ping":            {"monitor", GetHelpPing},
	"cmd_speed":           {"monitor", GetHelpSpeed},
	"cmd_ls":              {"files", GetHelpLs},
	"cmd_mkdir":           {"files", GetHelpMkdir},
	"cmd_rm":              {"files", GetHelpRm},
	"cmd_mv":              {"files", GetHelpMv},
	"cmd_share":           {"files", GetHelpShare},
	"cmd_find":            {"files", GetHelpFind},
	"cmd_extractaudio":    {"media", GetHelpExtractAudio},
	"cmd_compress":        {"media", GetHelpCompress},
	"cmd_thumbnail":       {"media", GetHelpThumbnail},
	"cmd_screenshots":     {"media", GetHelpScreenshots},
	"cmd_subtitle":        {"media", GetHelpSubtitle},
	"cmd_hardsub":         {"media", GetHelpHardsub},
	"cmd_rescale":         {"media", GetHelpRescale},
	"cmd_convert":         {"media", GetHelpConvert},
	"cmd_mediainfo":       {"media", GetHelpMediaInfo},
	"cmd_cancel":          {"task", GetHelpCancel},
	"cmd_cancelall":       {"task", GetHelpCancelAll},
	"cmd_storages":        {"storage", GetHelpStorages},
	"cmd_setstorage":      {"storage", GetHelpSetStorage},
	"cmd_authorize":       {"admin", GetHelpAuthorize},
	"cmd_unauthorize":     {"admin", GetHelpUnauthorize},
	"cmd_users":           {"admin", GetHelpUsers},
	"cmd_setalertchannel": {"admin", GetHelpSetAlertChannel},
	"cmd_setlogchannel":   {"admin", GetHelpSetLogChannel},
	"cmd_recover":         {"recovery", GetHelpRecover},
	"cmd_recoverystatus":  {"recovery", GetHelpRecoveryStatus},
}
