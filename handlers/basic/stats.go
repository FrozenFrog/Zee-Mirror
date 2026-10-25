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
	stats, err := s.DB.GetBotStats(ctx)
	if err != nil {
		s.Reply(message, "❌ *Không lấy được thống kê*")
		return
	}

	userStats, _ := s.DB.GetUserStats(ctx, message.From.ID)
	dailyStats, _ := s.DB.GetTodayStats(ctx)
	userDailyStats, _ := s.DB.GetUserTodayStats(ctx, message.From.ID)
	weeklyStats, _ := s.DB.GetWeeklyStats(ctx)

	slog.Info("Generating stats", "userID", message.From.ID)

	text := FormatStatsMessage(stats, userStats, dailyStats, userDailyStats)

	keyboard := GetStatsKeyboard()

	chartBytes, err := chart.GenerateWeeklyStatsChart(weeklyStats)
	if err == nil && len(chartBytes) > 0 {
		msg := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FileBytes{Name: "stats.png", Bytes: chartBytes})
		msg.Caption = text
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		msg.ReplyMarkup = keyboard

		sentMsg, err := s.Bot.Send(msg)
		if err != nil {
			slog.Error("Error sending stats photo", "error", err, "userID", message.From.ID)

			fallbackMsg := tgbotapi.NewMessage(message.Chat.ID, text)
			fallbackMsg.ParseMode = tgbotapi.ModeMarkdownV2
			fallbackMsg.ReplyMarkup = keyboard
			if sentTextMsg, err := s.Bot.Send(fallbackMsg); err == nil {
				s.AutoDeleteMessage(message.Chat.ID, sentTextMsg.MessageID, 60*time.Second)
			}
		} else {
			s.AutoDeleteMessage(message.Chat.ID, sentMsg.MessageID, 60*time.Second)
		}
	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		msg.ReplyMarkup = keyboard
		sentMsg, err := s.Bot.Send(msg)
		if err != nil {
			slog.Error("Error sending stats message", "error", err, "userID", message.From.ID)
			msg.ParseMode = ""
			msg.Text = "❌ Không định dạng được thống kê (lỗi Markdown)\n\nDự phòng: thống kê rút gọn\nTổng Task: " + fmt.Sprint(getIntValue(stats, "total_tasks"))
			_, _ = s.Bot.Send(msg)
		} else {
			s.AutoDeleteMessage(message.Chat.ID, sentMsg.MessageID, 60*time.Second)
		}
	}
}

func GetStatsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.InlineKeyboardMarkup{}
}

func FormatStatsMessage(stats map[string]interface{}, userStats *UserStats, dailyStats *DailyStats, userDailyStats *DailyStats) string {
	var text strings.Builder

	text.WriteString("📊 *THỐNG KÊ DASHBOARD*\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")

	totalTasks := getIntValue(stats, "total_tasks")
	completedTasks := getIntValue(stats, "completed_tasks")
	failedTasks := getIntValue(stats, "failed_tasks")
	totalBandwidth := getInt64Value(stats, "total_bandwidth")

	text.WriteString("🌐 *THỐNG KÊ TOÀN CỤC*\n")
	fmt.Fprintf(&text, "📥 Tổng Task: `%d`\n", totalTasks)
	fmt.Fprintf(&text, "✅ Hoàn thành: `%d`\n", completedTasks)
	fmt.Fprintf(&text, "❌ Thất bại: `%d`\n", failedTasks)
	fmt.Fprintf(&text, "📊 Tỷ lệ thành công: `%.1f%%`\n", calculateSuccessRate(completedTasks, totalTasks))
	fmt.Fprintf(&text, "📈 Băng thông: `%s`\n\n", utils.EscapeMarkdownV2Code(utils.FormatBytes(totalBandwidth)))

	if dailyStats != nil {
		text.WriteString("📅 *HÔM NAY \\(TOÀN CỤC\\)*\n")
		fmt.Fprintf(&text, "📥 Task: `%d`\n", dailyStats.TotalTasks)
		fmt.Fprintf(&text, "✅ Hoàn thành: `%d`\n", dailyStats.CompletedTasks)
		fmt.Fprintf(&text, "📈 Băng thông: `%s`\n\n", utils.EscapeMarkdownV2Code(utils.FormatBytes(dailyStats.TotalBandwidth)))
	}

	if userDailyStats != nil && userDailyStats.TotalTasks > 0 {
		text.WriteString("👤 *HÔM NAY CỦA TÔI*\n")
		fmt.Fprintf(&text, "📥 Task: `%d`\n", userDailyStats.TotalTasks)
		fmt.Fprintf(&text, "📊 Băng thông: `%s`\n\n", utils.EscapeMarkdownV2Code(utils.FormatBytes(userDailyStats.TotalBandwidth)))
	}

	if userStats != nil {
		text.WriteString("👤 *TỔNG THỐNG KÊ CỦA TÔI*\n")
		fmt.Fprintf(&text, "📥 Tổng Task: `%d`\n", userStats.TotalDownloads)
		fmt.Fprintf(&text, "📊 Băng thông: `%s`\n", utils.EscapeMarkdownV2Code(utils.FormatBytes(userStats.TotalBandwidth)))
		if !userStats.LastActive.IsZero() {
			fmt.Fprintf(&text, "🕐 Hoạt động: `%s`\n", utils.EscapeMarkdownV2Code(userStats.LastActive.Format("15:04")))
		}
	}

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")
	return text.String()
}

func formatUserStatsDetailed(stats *UserStats) string {
	if stats == nil {
		return "📭 *Bạn chưa có thống kê nào*"
	}

	var text strings.Builder
	text.WriteString("👤 *THỐNG KÊ NGƯỜI DÙNG*\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")

	text.WriteString("🆔 *THÔNG TIN NGƯỜI DÙNG*\n")
	fmt.Fprintf(&text, "👤 Tên: @%s\n", utils.EscapeMarkdownV2(stats.Username))
	fmt.Fprintf(&text, "🆔 ID: `%d`\n", stats.UserID)
	if !stats.LastActive.IsZero() {
		fmt.Fprintf(&text, "🕐 Hoạt động gần nhất: `%s`\n", utils.EscapeMarkdownV2Code(stats.LastActive.Format("02 Jan 15:04")))
	}
	text.WriteString("\n")

	text.WriteString("📊 *HOẠT ĐỘNG*\n")
	fmt.Fprintf(&text, "📥 Tổng Task: `%d`\n", stats.TotalDownloads)
	fmt.Fprintf(&text, "✅ Thành công: `%d`\n", stats.SuccessfulTasks)
	fmt.Fprintf(&text, "❌ Thất bại: `%d`\n", stats.FailedTasks)
	fmt.Fprintf(&text, "📈 Băng thông: `%s`\n", utils.EscapeMarkdownV2Code(utils.FormatBytes(stats.TotalBandwidth)))

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")
	return text.String()
}

func formatDailyStatsDetailed(stats *DailyStats, title string, userStats *DailyStats) string {
	if stats == nil {
		return "📭 *Chưa có thống kê*"
	}

	var text strings.Builder
	fmt.Fprintf(&text, "📅 *THỐNG KÊ %s*\n", strings.ToUpper(utils.EscapeMarkdownV2(title)))
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")

	fmt.Fprintf(&text, "📆 *%s*\n\n", stats.Date.Format("02 Jan 2006"))

	text.WriteString("🌐 *HOẠT ĐỘNG TOÀN CỤC*\n")
	fmt.Fprintf(&text, "📥 Tổng Task: `%d`\n", stats.TotalTasks)
	fmt.Fprintf(&text, "✅ Hoàn thành: `%d`\n", stats.CompletedTasks)
	fmt.Fprintf(&text, "❌ Thất bại: `%d`\n", stats.FailedTasks)
	fmt.Fprintf(&text, "📊 Tỷ lệ thành công: `%.1f%%`\n", calculateSuccessRate(stats.CompletedTasks, stats.TotalTasks))
	fmt.Fprintf(&text, "📈 Băng thông: `%s`\n", utils.EscapeMarkdownV2Code(utils.FormatBytes(stats.TotalBandwidth)))
	if stats.AverageSpeed > 0 {
		fmt.Fprintf(&text, "⚡ Tốc độ TB: `%s`\n", utils.EscapeMarkdownV2Code(utils.FormatSpeed(stats.AverageSpeed)))
	}
	if stats.PeakConcurrent > 0 {
		fmt.Fprintf(&text, "🔥 Đồng thời cao nhất: `%d`\n", stats.PeakConcurrent)
	}

	if userStats != nil {
		text.WriteString("\n👤 *HOẠT ĐỘNG CỦA BẠN HÔM NAY*\n")
		fmt.Fprintf(&text, "📥 Tổng Task: `%d`\n", userStats.TotalTasks)
		fmt.Fprintf(&text, "✅ Thành công: `%d`\n", userStats.CompletedTasks)
		fmt.Fprintf(&text, "❌ Thất bại: `%d`\n", userStats.FailedTasks)
		fmt.Fprintf(&text, "📈 Băng thông: `%s`\n", utils.EscapeMarkdownV2Code(utils.FormatBytes(userStats.TotalBandwidth)))
	}

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")
	return text.String()
}

func formatWeeklyStats(stats []DailyStats) string {
	var text strings.Builder
	text.WriteString("📈 *THỐNG KÊ TUẦN*\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")

	totalTasks := 0
	totalCompleted := 0
	totalFailed := 0

	var totalBandwidth int64

	text.WriteString("🗓️ *7 NGÀY QUA*\n")

	for _, day := range stats {
		totalTasks += day.TotalTasks
		totalCompleted += day.CompletedTasks
		totalFailed += day.FailedTasks
		totalBandwidth += day.TotalBandwidth

		icon := "▫️"
		if day.TotalTasks > 0 {
			icon = "▪️"
		}

		fmt.Fprintf(&text, "%s `%s` : %d Task\n",
			icon,
			day.Date.Format("02/01"),
			day.TotalTasks)
	}

	text.WriteString("\n📊 *TỔNG KẾT*\n")
	fmt.Fprintf(&text, "📥 Tổng: `%d` Task\n", totalTasks)
	fmt.Fprintf(&text, "✅ Hoàn thành: `%d`\n", totalCompleted)
	fmt.Fprintf(&text, "📈 Băng thông: `%s`\n", utils.EscapeMarkdownV2Code(utils.FormatBytes(totalBandwidth)))

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")
	return text.String()
}

func formatMonthlyStats(stats []DailyStats) string {
	var text strings.Builder
	text.WriteString("📉 *THỐNG KÊ THÁNG*\n")
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

	text.WriteString("📅 *THEO TUẦN*\n")

	for _, week := range weeks {
		ws := weekStats[week]
		fmt.Fprintf(&text, "Tuần %d: `%d` Task\n└ `%s`\n",
			week,
			ws.tasks,
			utils.EscapeMarkdownV2Code(utils.FormatBytes(ws.bandwidth)))
	}

	text.WriteString("\n📊 *TỔNG KẾT*\n")
	fmt.Fprintf(&text, "📥 Tổng: `%d` Task\n", totalTasks)
	fmt.Fprintf(&text, "✅ Hoàn thành: `%d`\n", totalCompleted)
	fmt.Fprintf(&text, "📈 Băng thông: `%s`\n", utils.EscapeMarkdownV2Code(utils.FormatBytes(totalBandwidth)))

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")
	return text.String()
}

func HandleStatsCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) < 2 {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		return
	}

	ctx := context.Background()
	action := parts[1]
	var text string

	switch action {
	case "my":
		userStats, err := s.DB.GetUserStats(ctx, callback.From.ID)
		if err != nil {
			text = "❌ *Không lấy được thống kê của bạn*"
		} else {
			text = formatUserStatsDetailed(userStats)
		}

	case "today":
		dailyStats, err := s.DB.GetTodayStats(ctx)
		userDailyStats, _ := s.DB.GetUserTodayStats(ctx, callback.From.ID)
		if err != nil {
			text = "❌ *Không lấy được thống kê hôm nay*"
		} else {
			text = formatDailyStatsDetailed(dailyStats, "Hôm nay", userDailyStats)
		}

	case "weekly":
		weeklyStats, err := s.DB.GetWeeklyStats(ctx)
		if err != nil {
			text = "❌ *Không lấy được thống kê tuần*"
		} else {
			text = formatWeeklyStats(weeklyStats)
		}

	case "monthly":
		monthlyStats, err := s.DB.GetMonthlyStats(ctx)
		if err != nil {
			text = "❌ *Không lấy được thống kê tháng*"
		} else {
			text = formatMonthlyStats(monthlyStats)
		}

	case CmdRefresh:
		stats, _ := s.DB.GetBotStats(ctx)
		userStats, _ := s.DB.GetUserStats(ctx, callback.From.ID)
		dailyStats, _ := s.DB.GetTodayStats(ctx)
		userDailyStats, _ := s.DB.GetUserTodayStats(ctx, callback.From.ID)
		text = FormatStatsMessage(stats, userStats, dailyStats, userDailyStats)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "🔄 Đã làm mới thống kê!"))

	case CmdClose:
		deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
		_, _ = s.Bot.Request(deleteMsg)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "Đã đóng"))
		return
	}

	if text != "" {
		if callback.Message.Photo != nil {
			editMsg := tgbotapi.NewEditMessageCaption(callback.Message.Chat.ID, callback.Message.MessageID, text)
			editMsg.ParseMode = tgbotapi.ModeMarkdownV2
			_, _ = s.Bot.Send(editMsg)
		} else {
			editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
			editMsg.ParseMode = tgbotapi.ModeMarkdownV2
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
