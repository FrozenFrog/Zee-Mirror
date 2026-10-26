package basic

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"zee-mirror/internal/domain"
	"zee-mirror/internal/service"
	"zee-mirror/pkg/chart"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	CmdRefresh = "refresh"
)

type UserStats = domain.UserStats
type DailyStats = domain.DailyStats

func HandleStats(s *service.BotService, message *tgbotapi.Message) {
	if !s.CanUse(message) {
		return
	}

	ctx := context.Background()
	lang := s.GetUserLanguage(message.From.ID)
	stats, err := s.DB.GetBotStats(ctx)
	if err != nil {
		s.Reply(message, i18n.T(lang, "stats_load_failed"))
		return
	}

	userStats, _ := s.DB.GetUserStats(ctx, message.From.ID)
	dailyStats, _ := s.DB.GetTodayStats(ctx)
	userDailyStats, _ := s.DB.GetUserTodayStats(ctx, message.From.ID)
	weeklyStats, _ := s.DB.GetWeeklyStats(ctx)

	slog.Info("Generating stats", "userID", message.From.ID)

	text := FormatStatsMessage(lang, stats, userStats, dailyStats, userDailyStats)

	keyboard := GetStatsKeyboard(lang)

	chartBytes, err := chart.GenerateWeeklyStatsChart(weeklyStats)
	if err == nil && len(chartBytes) > 0 {
		msg := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FileBytes{Name: "stats.png", Bytes: chartBytes})
		msg.Caption = text
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		if len(keyboard.InlineKeyboard) > 0 {
			msg.ReplyMarkup = keyboard
		}

		sentMsg, err := s.Bot.Send(msg)
		if err != nil {
			slog.Error("Error sending stats photo", "error", err, "userID", message.From.ID)

			fallbackMsg := tgbotapi.NewMessage(message.Chat.ID, text)
			fallbackMsg.ParseMode = tgbotapi.ModeMarkdownV2
			if len(keyboard.InlineKeyboard) > 0 {
				fallbackMsg.ReplyMarkup = keyboard
			}
			if sentTextMsg, err := s.Bot.Send(fallbackMsg); err == nil {
				s.AutoDeleteMessage(message.Chat.ID, sentTextMsg.MessageID, 60*time.Second)
			}
		} else {
			s.AutoDeleteMessage(message.Chat.ID, sentMsg.MessageID, 60*time.Second)
		}
	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		if len(keyboard.InlineKeyboard) > 0 {
			msg.ReplyMarkup = keyboard
		}
		sentMsg, err := s.Bot.Send(msg)
		if err != nil {
			slog.Error("Error sending stats message", "error", err, "userID", message.From.ID)
			msg.ParseMode = ""
			msg.Text = i18n.T(lang, "stats_fallback_plain", getIntValue(stats, "total_tasks"))
			_, _ = s.Bot.Send(msg)
		} else {
			s.AutoDeleteMessage(message.Chat.ID, sentMsg.MessageID, 60*time.Second)
		}
	}
}

func GetStatsKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "stats_btn_my"), "stats:my"),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "stats_btn_today"), "stats:today"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "stats_btn_week"), "stats:weekly"),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "stats_btn_month"), "stats:monthly"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "basic_btn_refresh"), "stats:"+CmdRefresh),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_close"), "stats:"+CmdClose),
		),
	)
}

// writeLine appends the translated line for key followed by a newline.
func writeLine(b *strings.Builder, lang, key string, args ...any) {
	b.WriteString(i18n.T(lang, key, args...))
	b.WriteString("\n")
}

func FormatStatsMessage(lang string, stats map[string]interface{}, userStats *UserStats, dailyStats *DailyStats, userDailyStats *DailyStats) string {
	var text strings.Builder

	writeLine(&text, lang, "stats_title")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")

	totalTasks := getIntValue(stats, "total_tasks")
	completedTasks := getIntValue(stats, "completed_tasks")
	failedTasks := getIntValue(stats, "failed_tasks")
	totalBandwidth := getInt64Value(stats, "total_bandwidth")

	writeLine(&text, lang, "stats_sec_global")
	writeLine(&text, lang, "stats_total_tasks", totalTasks)
	writeLine(&text, lang, "stats_completed", completedTasks)
	writeLine(&text, lang, "stats_failed", failedTasks)
	writeLine(&text, lang, "stats_success_rate", calculateSuccessRate(completedTasks, totalTasks))
	writeLine(&text, lang, "stats_bandwidth", utils.EscapeMarkdownV2Code(utils.FormatBytes(totalBandwidth)))
	text.WriteString("\n")

	if dailyStats != nil {
		writeLine(&text, lang, "stats_sec_today_global")
		writeLine(&text, lang, "stats_tasks_short", dailyStats.TotalTasks)
		writeLine(&text, lang, "stats_completed", dailyStats.CompletedTasks)
		writeLine(&text, lang, "stats_bandwidth", utils.EscapeMarkdownV2Code(utils.FormatBytes(dailyStats.TotalBandwidth)))
		text.WriteString("\n")
	}

	if userDailyStats != nil && userDailyStats.TotalTasks > 0 {
		writeLine(&text, lang, "stats_sec_today_me")
		writeLine(&text, lang, "stats_tasks_short", userDailyStats.TotalTasks)
		writeLine(&text, lang, "stats_bandwidth_alt", utils.EscapeMarkdownV2Code(utils.FormatBytes(userDailyStats.TotalBandwidth)))
		text.WriteString("\n")
	}

	if userStats != nil {
		writeLine(&text, lang, "stats_sec_total_me")
		writeLine(&text, lang, "stats_total_tasks", userStats.TotalDownloads)
		writeLine(&text, lang, "stats_bandwidth_alt", utils.EscapeMarkdownV2Code(utils.FormatBytes(userStats.TotalBandwidth)))
		if !userStats.LastActive.IsZero() {
			writeLine(&text, lang, "stats_last_active_short", utils.EscapeMarkdownV2Code(userStats.LastActive.Format("15:04")))
		}
	}

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")
	return text.String()
}

func formatUserStatsDetailed(lang string, stats *UserStats) string {
	if stats == nil {
		return i18n.T(lang, "stats_none_user")
	}

	var text strings.Builder
	writeLine(&text, lang, "stats_user_title")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")

	writeLine(&text, lang, "stats_sec_user_info")
	writeLine(&text, lang, "stats_user_name", utils.EscapeMarkdownV2(stats.Username))
	writeLine(&text, lang, "stats_user_id", stats.UserID)
	if !stats.LastActive.IsZero() {
		writeLine(&text, lang, "stats_last_active", utils.EscapeMarkdownV2Code(stats.LastActive.Format("02 Jan 15:04")))
	}
	text.WriteString("\n")

	writeLine(&text, lang, "stats_sec_activity")
	writeLine(&text, lang, "stats_total_tasks", stats.TotalDownloads)
	writeLine(&text, lang, "stats_succeeded", stats.SuccessfulTasks)
	writeLine(&text, lang, "stats_failed", stats.FailedTasks)
	writeLine(&text, lang, "stats_bandwidth", utils.EscapeMarkdownV2Code(utils.FormatBytes(stats.TotalBandwidth)))

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")
	return text.String()
}

// formatDailyStatsDetailed renders today's statistics page.
func formatDailyStatsDetailed(lang string, stats *DailyStats, userStats *DailyStats) string {
	if stats == nil {
		return i18n.T(lang, "stats_none")
	}

	var text strings.Builder
	writeLine(&text, lang, "stats_daily_title")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")

	fmt.Fprintf(&text, "📆 *%s*\n\n", stats.Date.Format("02 Jan 2006"))

	writeLine(&text, lang, "stats_sec_global_activity")
	writeLine(&text, lang, "stats_total_tasks", stats.TotalTasks)
	writeLine(&text, lang, "stats_completed", stats.CompletedTasks)
	writeLine(&text, lang, "stats_failed", stats.FailedTasks)
	writeLine(&text, lang, "stats_success_rate", calculateSuccessRate(stats.CompletedTasks, stats.TotalTasks))
	writeLine(&text, lang, "stats_bandwidth", utils.EscapeMarkdownV2Code(utils.FormatBytes(stats.TotalBandwidth)))
	if stats.AverageSpeed > 0 {
		writeLine(&text, lang, "stats_avg_speed", utils.EscapeMarkdownV2Code(utils.FormatSpeed(stats.AverageSpeed)))
	}
	if stats.PeakConcurrent > 0 {
		writeLine(&text, lang, "stats_peak_concurrent", stats.PeakConcurrent)
	}

	if userStats != nil {
		text.WriteString("\n")
		writeLine(&text, lang, "stats_sec_my_today")
		writeLine(&text, lang, "stats_total_tasks", userStats.TotalTasks)
		writeLine(&text, lang, "stats_succeeded", userStats.CompletedTasks)
		writeLine(&text, lang, "stats_failed", userStats.FailedTasks)
		writeLine(&text, lang, "stats_bandwidth", utils.EscapeMarkdownV2Code(utils.FormatBytes(userStats.TotalBandwidth)))
	}

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")
	return text.String()
}

func formatWeeklyStats(lang string, stats []DailyStats) string {
	var text strings.Builder
	writeLine(&text, lang, "stats_weekly_title")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")

	totalTasks := 0
	totalCompleted := 0
	totalFailed := 0

	var totalBandwidth int64

	writeLine(&text, lang, "stats_sec_7days")

	for _, day := range stats {
		totalTasks += day.TotalTasks
		totalCompleted += day.CompletedTasks
		totalFailed += day.FailedTasks
		totalBandwidth += day.TotalBandwidth

		icon := "▫️"
		if day.TotalTasks > 0 {
			icon = "▪️"
		}

		writeLine(&text, lang, "stats_day_row", icon, day.Date.Format("02/01"), day.TotalTasks)
	}

	text.WriteString("\n")
	writeLine(&text, lang, "stats_sec_summary")
	writeLine(&text, lang, "stats_total_line", totalTasks)
	writeLine(&text, lang, "stats_completed", totalCompleted)
	writeLine(&text, lang, "stats_bandwidth", utils.EscapeMarkdownV2Code(utils.FormatBytes(totalBandwidth)))

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")
	return text.String()
}

func formatMonthlyStats(lang string, stats []DailyStats) string {
	var text strings.Builder
	writeLine(&text, lang, "stats_monthly_title")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")

	totalTasks := 0
	totalCompleted := 0
	totalFailed := 0
	var totalBandwidth int64

	weekStats := make(map[int]struct {
		tasks     int
		completed int
		bandwidth int64
	})

	var weeks []int

	for _, day := range stats {
		totalTasks += day.TotalTasks
		totalCompleted += day.CompletedTasks
		totalFailed += day.FailedTasks
		totalBandwidth += day.TotalBandwidth

		_, week := day.Date.ISOWeek()
		if _, exists := weekStats[week]; !exists {
			weeks = append(weeks, week)
		}

		ws := weekStats[week]
		ws.tasks += day.TotalTasks
		ws.completed += day.CompletedTasks
		ws.bandwidth += day.TotalBandwidth
		weekStats[week] = ws
	}

	writeLine(&text, lang, "stats_sec_by_week")

	for _, week := range weeks {
		ws := weekStats[week]
		writeLine(&text, lang, "stats_week_row", week, ws.tasks, utils.EscapeMarkdownV2Code(utils.FormatBytes(ws.bandwidth)))
	}

	text.WriteString("\n")
	writeLine(&text, lang, "stats_sec_summary")
	writeLine(&text, lang, "stats_total_line", totalTasks)
	writeLine(&text, lang, "stats_completed", totalCompleted)
	writeLine(&text, lang, "stats_bandwidth", utils.EscapeMarkdownV2Code(utils.FormatBytes(totalBandwidth)))

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")
	return text.String()
}

func HandleStatsCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) < 2 {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		return
	}

	ctx := context.Background()
	lang := s.GetUserLanguage(callback.From.ID)
	action := parts[1]
	var text string

	switch action {
	case "my":
		userStats, err := s.DB.GetUserStats(ctx, callback.From.ID)
		if err != nil {
			text = i18n.T(lang, "stats_err_my")
		} else {
			text = formatUserStatsDetailed(lang, userStats)
		}

	case "today":
		dailyStats, err := s.DB.GetTodayStats(ctx)
		userDailyStats, _ := s.DB.GetUserTodayStats(ctx, callback.From.ID)
		if err != nil {
			text = i18n.T(lang, "stats_err_today")
		} else {
			text = formatDailyStatsDetailed(lang, dailyStats, userDailyStats)
		}

	case "weekly":
		weeklyStats, err := s.DB.GetWeeklyStats(ctx)
		if err != nil {
			text = i18n.T(lang, "stats_err_week")
		} else {
			text = formatWeeklyStats(lang, weeklyStats)
		}

	case "monthly":
		monthlyStats, err := s.DB.GetMonthlyStats(ctx)
		if err != nil {
			text = i18n.T(lang, "stats_err_month")
		} else {
			text = formatMonthlyStats(lang, monthlyStats)
		}

	case CmdRefresh:
		stats, _ := s.DB.GetBotStats(ctx)
		userStats, _ := s.DB.GetUserStats(ctx, callback.From.ID)
		dailyStats, _ := s.DB.GetTodayStats(ctx)
		userDailyStats, _ := s.DB.GetUserTodayStats(ctx, callback.From.ID)
		text = FormatStatsMessage(lang, stats, userStats, dailyStats, userDailyStats)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "stats_toast_refreshed")))

	case CmdClose:
		deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
		_, _ = s.Bot.Request(deleteMsg)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "basic_toast_closed")))
		return
	}

	if text != "" {
		keyboard := GetStatsKeyboard(lang)
		if action != CmdRefresh {
			keyboard = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_back"), "stats:"+CmdRefresh),
				tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_close"), "stats:"+CmdClose),
			))
		}

		if callback.Message.Photo != nil {
			editMsg := tgbotapi.NewEditMessageCaption(callback.Message.Chat.ID, callback.Message.MessageID, text)
			editMsg.ParseMode = tgbotapi.ModeMarkdownV2
			editMsg.ReplyMarkup = &keyboard
			_, _ = s.Bot.Send(editMsg)
		} else {
			editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
			editMsg.ParseMode = tgbotapi.ModeMarkdownV2
			editMsg.ReplyMarkup = &keyboard
			_, _ = s.Bot.Send(editMsg)
		}
	}
}

func calculateSuccessRate(completed, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(completed) / float64(total) * 100
}

func getIntValue(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}

func getInt64Value(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int64:
			return val
		case int:
			return int64(val)
		case float64:
			return int64(val)
		}
	}
	return 0
}
