package basic

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"zee-mirror/internal/service"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleDashboardCallback(s *service.BotService, callback *tgbotapi.CallbackQuery) {
	parts := strings.Split(callback.Data, ":")
	if len(parts) < 2 {
		return
	}
	action := parts[1]

	if handleDashboardAction(s, callback, action, parts) {
		return
	}

	text := getDashboardModeText(action)

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
}

func handleDashboardAction(s *service.BotService, callback *tgbotapi.CallbackQuery, action string, parts []string) bool {
	switch action {
	case "page":
		handlePagination(s, callback, parts)
	case "status":
		handleStatusFromCallback(s, callback)
	case "ping":
		handlePingFromCallback(s, callback)
	case "speed":
		handleSpeedFromCallback(s, callback)
	case "stats":
		handleStatsFromCallback(s, callback)
	case "storage":
		handleStoragesFromCallback(s, callback)
	case "files":
		handleDriveListFromCallback(s, callback)
	case "media":
		sendMediaMenu(s, callback)
	case "system":
		handleSystemFromCallback(s, callback)
	case "back":
		lang := s.GetUserLanguage(callback.From.ID)
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
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "Đã đóng"))
	default:
		return false
	}
	return true
}

func handlePagination(s *service.BotService, callback *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) >= 3 {
		page, _ := strconv.Atoi(parts[2])
		s.TaskManager.Mu.Lock()
		s.TaskManager.StatusPages[callback.Message.Chat.ID] = page
		s.TaskManager.Mu.Unlock()
		s.UpdateSharedDashboard(callback.Message.Chat.ID, false)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, fmt.Sprintf("Trang %d", page+1)))
	}
}

func handleStatusFromCallback(s *service.BotService, callback *tgbotapi.CallbackQuery) {
	lang := s.GetUserLanguage(callback.From.ID)
	tasks := s.TaskManager.GetActiveTasks()

	if len(tasks) == 0 {
		content := service.GetErrorMessage("TRẠNG THÁI", "Không có Task nào đang chạy\\.")
		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, content)
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
		_, _ = s.Bot.Send(editMsg)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "📭 Trống"))
		return
	}

	text := service.GetStatusHeader()
	for _, task := range tasks {
		snapshot := task.GetSnapshot()
		text += service.FormatTaskProfessional(lang, snapshot) + "\n"
	}

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeHTML
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "🔄 Đã cập nhật"))
}

func handlePingFromCallback(s *service.BotService, callback *tgbotapi.CallbackQuery) {
	start := time.Now()
	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, "🏓 *Đang ping\\.\\.\\.*")
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)

	elapsed := time.Since(start)
	text := fmt.Sprintf("🏓 *Pong\\!* `%v`", elapsed.Round(time.Millisecond))

	finalEdit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	finalEdit.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(finalEdit)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "🏓 Pong!"))
}

func handleSpeedFromCallback(s *service.BotService, callback *tgbotapi.CallbackQuery) {
	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, "🚀 *Đang chạy Speedtest\\.\\.\\.*")
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)

	go func() {
		cmd := execCommand(context.Background(), "speedtest-cli", "--simple")
		output, err := cmd.CombinedOutput()

		var text string
		if err != nil {
			text = fmt.Sprintf("❌ *Lỗi Speedtest*\n\n`%s`", utils.EscapeMarkdownV2(err.Error()))
		} else {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			var result strings.Builder
			result.WriteString("🚀 *Kết quả Speedtest*\n\n")
			for _, line := range lines {
				fmt.Fprintf(&result, "• `%s`\n", utils.EscapeMarkdownV2(line))
			}
			text = result.String()
		}

		finalEdit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
		finalEdit.ParseMode = tgbotapi.ModeMarkdownV2
		_, _ = s.Bot.Send(finalEdit)
	}()
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "🚀 Đang kiểm tra tốc độ..."))
}

func handleStatsFromCallback(s *service.BotService, callback *tgbotapi.CallbackQuery) {
	ctx := context.Background()
	stats, err := s.DB.GetBotStats(ctx)
	if err != nil {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "❌ Lỗi"))
		return
	}

	userStats, _ := s.DB.GetUserStats(ctx, callback.From.ID)
	dailyStats, _ := s.DB.GetTodayStats(ctx)
	userDailyStats, _ := s.DB.GetUserTodayStats(ctx, callback.From.ID)

	text := FormatStatsMessage(stats, userStats, dailyStats, userDailyStats)

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "📊 Thống kê"))
}

func handleStoragesFromCallback(s *service.BotService, callback *tgbotapi.CallbackQuery) {
	providers, err := s.GetAvailableStorages()
	if err != nil {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "❌ Lỗi"))
		return
	}

	var text strings.Builder
	text.WriteString("💾 *Nhà cung cấp lưu trữ khả dụng*\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for _, p := range providers {
		fmt.Fprintf(&text, "%s *%s* \\(%s\\)\n", p.Icon, utils.EscapeMarkdownV2(p.Name), utils.EscapeMarkdownV2(p.Type))
	}

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━\n")
	text.WriteString("*Hiện tại:* `" + utils.EscapeMarkdownV2(s.TaskManager.RcloneDest) + "`\n\n")

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text.String())
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "💾 Lưu trữ"))
}

func handleSystemFromCallback(s *service.BotService, callback *tgbotapi.CallbackQuery) {
	stats := s.GetSystemStats()
	text := s.FormatSystemStats(stats)

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "🖥️ Hệ thống"))
}

func handleDriveListFromCallback(s *service.BotService, callback *tgbotapi.CallbackQuery) {
	basePath := strings.TrimSuffix(s.TaskManager.RcloneDest, "/")
	path := basePath

	loadingMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, "🔍 *Đang tải danh sách file\\.\\.\\.*")
	loadingMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(loadingMsg)

	files, err := s.ListDriveFiles(path)
	if err != nil {
		text := fmt.Sprintf("❌ *Không tải được danh sách file*\n\nLỗi: %s", utils.EscapeMarkdownV2(err.Error()))
		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
		_, _ = s.Bot.Send(editMsg)
		return
	}

	text := s.FormatDriveFileList("/", files)

	finalEdit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	finalEdit.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(finalEdit)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "📁 Danh sách Drive"))
}

func sendMediaMenu(s *service.BotService, callback *tgbotapi.CallbackQuery) {
	content := "Các thao tác media cho video/audio:\n\n" +
		"• `/extractaudio` ─ Tách audio từ video\n" +
		"• `/compress` ─ Nén dung lượng video\n" +
		"• `/thumbnail` ─ Tạo thumbnail\n" +
		"• `/screenshots` ─ Chụp nhiều ảnh màn hình\n" +
		"• `/subtitle` ─ Soft\\-sub \\(Gắn subtitle\\)\n" +
		"• `/hardsub` ─ Hard\\-sub \\(Ghép cứng subtitle\\)\n" +
		"• `/rescale` ─ Đổi độ phân giải video\n" +
		"• `/convert` ─ Chuyển đổi định dạng file\n" +
		"• `/mediainfo` ─ Thông tin chi tiết file\n\n" +
		"💡 *Cách dùng:* Reply vào file video bằng lệnh tương ứng"

	text := service.ProfessionalMessage("XỬ LÝ MEDIA", content)

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "🎵 Menu Media"))
}

func getDashboardModeText(action string) string {
	switch action {
	case "mirror":
		return service.ProfessionalMessage("CHẾ ĐỘ MIRROR",
			"Tải file thẳng lên Google Drive\\.\n\n"+
				"📌 *CÁCH DÙNG*\n"+
				"• Reply vào file bằng `/mirror`\n"+
				"• Hoặc gửi: `/mirror URL`\n\n"+
				"🏷️ *FLAGS*\n"+
				"• `\\-z` ─ Nén Zip trước khi tải lên\n"+
				"• `\\-uz` ─ Giải nén sau khi tải xuống\n"+
				"• `\\-p PASS` ─ Mật khẩu zip")

	case "leech":
		return service.ProfessionalMessage("CHẾ ĐỘ LEECH",
			"Tải file từ URL về server\\.\n\n"+
				"✅ *HỖ TRỢ*\n"+
				"• HTTP/HTTPS • FTP • Magnet\n\n"+
				"📌 *CÁCH DÙNG*\n"+
				"• `/leech` \\<URL\\>")

	case "ytdlp":
		return service.ProfessionalMessage("CHẾ ĐỘ YT-DLP",
			"Tải video từ hơn 1000 trang web\\.\n\n"+
				"✅ *HỖ TRỢ*\n"+
				"• YouTube • Twitter • TikTok • v\\.v\\.\n\n"+
				"📌 *CÁCH DÙNG*\n"+
				"• `/ytdlp` \\<URL\\>")

	case "torrent":
		return service.ProfessionalMessage("CHẾ ĐỘ TORRENT",
			"Tải qua Magnet hoặc file Torrent\\.\n\n"+
				"📌 *CÁCH DÙNG*\n"+
				"• `/torrent` \\<magnet\\_link\\>\n"+
				"• Reply vào file `.torrent` bằng `/torrent`")

	case "clone":
		return service.ProfessionalMessage("CHẾ ĐỘ CLONE",
			"Clone file/thư mục Google Drive\\.\n\n"+
				"📌 *CÁCH DÙNG*\n"+
				"• `/clone` \\<GDRIVE\\_URL\\>")

	case "batch":
		return service.ProfessionalMessage("CHẾ ĐỘ BATCH",
			"Tải nhiều URL cùng lúc\\.\n\n"+
				"📌 *CÁCH DÙNG*\n"+
				"`/batch`\n"+
				"`URL1`\n"+
				"`URL2`")

	case "search":
		return service.ProfessionalMessage("CHẾ ĐỘ TÌM KIẾM",
			"Tìm Torrent từ nhiều nguồn\\.\n\n"+
				"📌 *CÁCH DÙNG*\n"+
				"• `/search` \\<keyword\\>")

	case "hardsub":
		return service.ProfessionalMessage("CHẾ ĐỘ HARD-SUB",
			"Ghép cứng subtitle vĩnh viễn vào video\\.\n\n"+
				"📌 *CÁCH DÙNG*\n"+
				"• Reply vào video bằng `/hardsub subtitle.srt`\n"+
				"• Hoặc: `/hardsub video.mp4 subtitle.srt`\n\n"+
				"🔥 *THÔNG TIN*\n"+
				"Quá trình này cần re\\-encode video nên sẽ mất nhiều thời gian hơn so với soft\\-sub\\.")

	case "rescale":
		return service.ProfessionalMessage("CHẾ ĐỘ RESCALE",
			"Đổi độ phân giải video \\(Transcoding\\)\\.\n\n"+
				"📌 *CÁCH DÙNG*\n"+
				"• Reply vào video bằng `/rescale 720p`\n"+
				"• Hoặc: `/rescale video.mp4 1280x720`\n\n"+
				"🏷️ *PRESET*\n"+
				"`4k`, `2k`, `1080p`, `720p`, `480p`, `360p` hoặc định dạng tùy chỉnh `WxH`\\.")

	case "back":
		return ""

	default:
		return service.GetErrorMessage("KHÔNG XÁC ĐỊNH", "Thao tác không xác định\\.")
	}
}
