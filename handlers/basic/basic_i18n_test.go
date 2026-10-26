package basic

import (
	"strings"
	"testing"
	"time"

	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"
)

// The Vietnamese rendering must stay byte-identical to the pre-i18n output.
func TestFormatStatsMessageViGolden(t *testing.T) {
	stats := map[string]interface{}{
		"total_tasks":     10,
		"completed_tasks": 8,
		"failed_tasks":    2,
		"total_bandwidth": int64(1024),
	}
	want := "📊 *THỐNG KÊ DASHBOARD*\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n\n" +
		"🌐 *THỐNG KÊ TOÀN CỤC*\n" +
		"📥 Tổng Task: `10`\n" +
		"✅ Hoàn thành: `8`\n" +
		"❌ Thất bại: `2`\n" +
		"📊 Tỷ lệ thành công: `80.0%`\n" +
		"📈 Băng thông: `" + utils.EscapeMarkdownV2Code(utils.FormatBytes(1024)) + "`\n\n" +
		"\n━━━━━━━━━━━━━━━━━━━━━━"
	if got := FormatStatsMessage("vi", stats, nil, nil, nil); got != want {
		t.Errorf("vi stats =\n%q\nwant\n%q", got, want)
	}
}

// No rendered text may leak a raw i18n key (a missing translation falls back
// to the key itself).
func TestStatsRenderingHasNoMissingKeys(t *testing.T) {
	day := DailyStats{
		Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), TotalTasks: 3, CompletedTasks: 2, FailedTasks: 1,
		TotalBandwidth: 2048, AverageSpeed: 1000, PeakConcurrent: 2,
	}
	user := &UserStats{Username: "tester", UserID: 7, TotalDownloads: 4, LastActive: time.Now()}
	stats := map[string]interface{}{"total_tasks": 1}

	for _, l := range i18n.Languages {
		texts := map[string]string{
			"stats":   FormatStatsMessage(l.Code, stats, user, &day, &day),
			"user":    formatUserStatsDetailed(l.Code, user),
			"none":    formatUserStatsDetailed(l.Code, nil),
			"daily":   formatDailyStatsDetailed(l.Code, &day, &day),
			"weekly":  formatWeeklyStats(l.Code, []DailyStats{day}),
			"monthly": formatMonthlyStats(l.Code, []DailyStats{day}),
		}
		for mode := range dashboardModes {
			texts["mode_"+mode] = getDashboardModeText(l.Code, mode)
		}
		texts["unknown"] = getDashboardModeText(l.Code, "nope")
		for _, row := range GetStatsKeyboard(l.Code).InlineKeyboard {
			for _, b := range row {
				texts["btn_"+b.Text] = b.Text
			}
		}
		for _, row := range systemKeyboard(l.Code).InlineKeyboard {
			for _, b := range row {
				texts["btn_"+b.Text] = b.Text
			}
		}

		for name, text := range texts {
			if text == "" {
				t.Errorf("%s/%s: empty", l.Code, name)
			}
			for _, prefix := range []string{"stats_", "sys_", "bdash_", "basic_", "join_", "notify_cmd_", "schedule_"} {
				if strings.Contains(text, prefix) {
					t.Errorf("%s/%s: leaked key prefix %q in %q", l.Code, name, prefix, text)
				}
			}
		}
	}
}

func TestDashboardModesDifferPerLanguage(t *testing.T) {
	for mode := range dashboardModes {
		if getDashboardModeText("vi", mode) == getDashboardModeText("en", mode) {
			t.Errorf("mode %q is not translated", mode)
		}
	}
}
