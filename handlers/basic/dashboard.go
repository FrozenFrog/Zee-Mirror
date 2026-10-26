package basic

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleDashboardCallback(s *service.BotService, callback *tgbotapi.CallbackQuery) {
	parts := strings.Split(callback.Data, ":")
	if len(parts) < 2 {
		return
	}
	action := parts[1]
	lang := s.GetUserLanguage(callback.From.ID)

	if handleDashboardAction(s, callback, lang, action, parts) {
		return
	}

	text := getDashboardModeText(lang, action)

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
}

func handleDashboardAction(s *service.BotService, callback *tgbotapi.CallbackQuery, lang, action string, parts []string) bool {
	switch action {
	case "page":
		handlePagination(s, callback, lang, parts)
	case "status":
		handleStatusFromCallback(s, callback, lang)
	case "ping":
		handlePingFromCallback(s, callback, lang)
	case "speed":
		handleSpeedFromCallback(s, callback, lang)
	case "stats":
		handleStatsFromCallback(s, callback, lang)
	case "storage":
		handleStoragesFromCallback(s, callback, lang)
	case "files":
		handleDriveListFromCallback(s, callback, lang)
	case "media":
		sendMediaMenu(s, callback, lang)
	case "system":
		handleSystemFromCallback(s, callback, lang)
	case "back":
		content := service.GetWelcomeMessage(lang, callback.From.FirstName)
		kb := service.GetStartKeyboard(lang)
		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, content)
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
		if len(kb.InlineKeyboard) > 0 {
			editMsg.ReplyMarkup = &kb
		}
		_, _ = s.Bot.Send(editMsg)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
	case "close":
		deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
		_, _ = s.Bot.Request(deleteMsg)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "basic_toast_closed")))
	default:
		return false
	}
	return true
}

func handlePagination(s *service.BotService, callback *tgbotapi.CallbackQuery, lang string, parts []string) {
	if len(parts) >= 3 {
		page, _ := strconv.Atoi(parts[2])
		s.TaskManager.Mu.Lock()
		s.TaskManager.StatusPages[callback.Message.Chat.ID] = page
		s.TaskManager.Mu.Unlock()
		s.UpdateSharedDashboard(callback.Message.Chat.ID, false)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "bdash_toast_page", page+1)))
	}
}

func handleStatusFromCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, lang string) {
	tasks := s.TaskManager.GetActiveTasks()

	if len(tasks) == 0 {
		content := errorCard(lang, "bdash_status_hdr", "bdash_status_none")
		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, content)
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
		_, _ = s.Bot.Send(editMsg)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "bdash_toast_empty")))
		return
	}

	text := service.GetStatusHeaderLang(lang)
	for _, task := range tasks {
		snapshot := task.GetSnapshot()
		text += service.FormatTaskProfessional(lang, snapshot) + "\n"
	}

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeHTML
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "bdash_toast_updated")))
}

func handlePingFromCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, lang string) {
	start := time.Now()
	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, i18n.T(lang, "basic_pinging"))
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)

	elapsed := time.Since(start)
	text := i18n.T(lang, "basic_pong", elapsed.Round(time.Millisecond))

	finalEdit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	finalEdit.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(finalEdit)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "bdash_toast_pong")))
}

func handleSpeedFromCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, lang string) {
	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, i18n.T(lang, "basic_speedtest_running"))
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)

	go func() {
		cmd := execCommand(context.Background(), "speedtest-cli", "--simple")
		output, err := cmd.CombinedOutput()

		var text string
		if err != nil {
			text = i18n.T(lang, "basic_speedtest_error", utils.EscapeMarkdownV2(err.Error()))
		} else {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			var result strings.Builder
			result.WriteString(i18n.T(lang, "basic_speedtest_result") + "\n\n")
			for _, line := range lines {
				fmt.Fprintf(&result, "• `%s`\n", utils.EscapeMarkdownV2(line))
			}
			text = result.String()
		}

		finalEdit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
		finalEdit.ParseMode = tgbotapi.ModeMarkdownV2
		_, _ = s.Bot.Send(finalEdit)
	}()
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "bdash_toast_speedtest")))
}

func handleStatsFromCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, lang string) {
	ctx := context.Background()
	stats, err := s.DB.GetBotStats(ctx)
	if err != nil {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "bdash_toast_error")))
		return
	}

	userStats, _ := s.DB.GetUserStats(ctx, callback.From.ID)
	dailyStats, _ := s.DB.GetTodayStats(ctx)
	userDailyStats, _ := s.DB.GetUserTodayStats(ctx, callback.From.ID)

	text := FormatStatsMessage(lang, stats, userStats, dailyStats, userDailyStats)

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "bdash_toast_stats")))
}

func handleStoragesFromCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, lang string) {
	providers, err := s.GetAvailableStorages()
	if err != nil {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "bdash_toast_error")))
		return
	}

	var text strings.Builder
	text.WriteString(i18n.T(lang, "bdash_storage_title") + "\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for _, p := range providers {
		fmt.Fprintf(&text, "%s *%s* \\(%s\\)\n", p.Icon, utils.EscapeMarkdownV2(p.Name), utils.EscapeMarkdownV2(p.Type))
	}

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━\n")
	text.WriteString(i18n.T(lang, "bdash_storage_current", utils.EscapeMarkdownV2(s.TaskManager.RcloneDest)) + "\n\n")

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text.String())
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "bdash_toast_storage")))
}

func handleSystemFromCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, lang string) {
	stats := s.GetSystemStats()
	text := s.FormatSystemStatsLang(lang, stats)

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "bdash_toast_system")))
}

func handleDriveListFromCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, lang string) {
	basePath := strings.TrimSuffix(s.TaskManager.RcloneDest, "/")
	path := basePath

	loadingMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, i18n.T(lang, "bdash_files_loading"))
	loadingMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(loadingMsg)

	files, err := s.ListDriveFiles(path)
	if err != nil {
		text := i18n.T(lang, "bdash_files_error", utils.EscapeMarkdownV2(i18n.Message(lang, err)))
		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
		_, _ = s.Bot.Send(editMsg)
		return
	}

	text := s.FormatDriveFileListLang(lang, "/", files)

	finalEdit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	finalEdit.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(finalEdit)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "bdash_toast_files")))
}

func sendMediaMenu(s *service.BotService, callback *tgbotapi.CallbackQuery, lang string) {
	text := service.ProfessionalMessage(i18n.T(lang, "bdash_media_hdr"), i18n.T(lang, "bdash_media_body"))

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "bdash_toast_media")))
}

// dashboardModes are the dashboard actions that show a usage card; each has
// bdash_mode_<action>_hdr and bdash_mode_<action>_body keys.
var dashboardModes = map[string]bool{
	"mirror": true, "leech": true, "ytdlp": true, "torrent": true, "clone": true,
	"batch": true, "search": true, "hardsub": true, "rescale": true,
}

func getDashboardModeText(lang, action string) string {
	switch {
	case dashboardModes[action]:
		return service.ProfessionalMessage(
			i18n.T(lang, "bdash_mode_"+action+"_hdr"),
			i18n.T(lang, "bdash_mode_"+action+"_body"))
	case action == "back":
		return ""
	default:
		return errorCard(lang, "bdash_unknown_hdr", "bdash_unknown_body")
	}
}
