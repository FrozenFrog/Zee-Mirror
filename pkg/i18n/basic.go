package i18n

// Basic handlers: stats, dashboard, system, join, notify and schedule.
//
// Key naming (enforced by TestBasicKeys):
//   - keys containing "_toast", "_btn", "_plain" or "_hdr" are plain text
//     (callback toasts, button labels, plain-text messages, card titles that the
//     caller escapes itself) and must not contain MarkdownV2 escapes;
//   - every other key is a MarkdownV2 message stored pre-escaped.
func init() {
	register(map[string]map[string]string{
		"vi": {
			// stats
			"stats_load_failed":         "❌ *Không lấy được thống kê*",
			"stats_fallback_plain":      "❌ Không định dạng được thống kê (lỗi Markdown)\n\nDự phòng: thống kê rút gọn\nTổng Task: %d",
			"stats_btn_my":              "👤 Của tôi",
			"stats_btn_today":           "📊 Hôm nay",
			"stats_btn_week":            "📈 Tuần",
			"stats_btn_month":           "📉 Tháng",
			"stats_title":               "📊 *THỐNG KÊ DASHBOARD*",
			"stats_sec_global":          "🌐 *THỐNG KÊ TOÀN CỤC*",
			"stats_total_tasks":         "📥 Tổng Task: `%d`",
			"stats_completed":           "✅ Hoàn thành: `%d`",
			"stats_failed":              "❌ Thất bại: `%d`",
			"stats_success_rate":        "📊 Tỷ lệ thành công: `%.1f%%`",
			"stats_bandwidth":           "📈 Băng thông: `%s`",
			"stats_bandwidth_alt":       "📊 Băng thông: `%s`",
			"stats_sec_today_global":    "📅 *HÔM NAY \\(TOÀN CỤC\\)*",
			"stats_tasks_short":         "📥 Task: `%d`",
			"stats_sec_today_me":        "👤 *HÔM NAY CỦA TÔI*",
			"stats_sec_total_me":        "👤 *TỔNG THỐNG KÊ CỦA TÔI*",
			"stats_last_active_short":   "🕐 Hoạt động: `%s`",
			"stats_none_user":           "📭 *Bạn chưa có thống kê nào*",
			"stats_user_title":          "👤 *THỐNG KÊ NGƯỜI DÙNG*",
			"stats_sec_user_info":       "🆔 *THÔNG TIN NGƯỜI DÙNG*",
			"stats_user_name":           "👤 Tên: @%s",
			"stats_user_id":             "🆔 ID: `%d`",
			"stats_last_active":         "🕐 Hoạt động gần nhất: `%s`",
			"stats_sec_activity":        "📊 *HOẠT ĐỘNG*",
			"stats_succeeded":           "✅ Thành công: `%d`",
			"stats_none":                "📭 *Chưa có thống kê*",
			"stats_daily_title":         "📅 *THỐNG KÊ HÔM NAY*",
			"stats_sec_global_activity": "🌐 *HOẠT ĐỘNG TOÀN CỤC*",
			"stats_avg_speed":           "⚡ Tốc độ TB: `%s`",
			"stats_peak_concurrent":     "🔥 Đồng thời cao nhất: `%d`",
			"stats_sec_my_today":        "👤 *HOẠT ĐỘNG CỦA BẠN HÔM NAY*",
			"stats_weekly_title":        "📈 *THỐNG KÊ TUẦN*",
			"stats_sec_7days":           "🗓️ *7 NGÀY QUA*",
			"stats_day_row":             "%s `%s` : %d Task",
			"stats_sec_summary":         "📊 *TỔNG KẾT*",
			"stats_total_line":          "📥 Tổng: `%d` Task",
			"stats_monthly_title":       "📉 *THỐNG KÊ THÁNG*",
			"stats_sec_by_week":         "📅 *THEO TUẦN*",
			"stats_week_row":            "Tuần %d: `%d` Task\n└ `%s`",
			"stats_err_my":              "❌ *Không lấy được thống kê của bạn*",
			"stats_err_today":           "❌ *Không lấy được thống kê hôm nay*",
			"stats_err_week":            "❌ *Không lấy được thống kê tuần*",
			"stats_err_month":           "❌ *Không lấy được thống kê tháng*",
			"stats_toast_refreshed":     "🔄 Đã làm mới thống kê!",

			// shared basic bits
			"basic_btn_refresh":         "🔄 Làm mới",
			"basic_toast_closed":        "Đã đóng",
			"basic_pinging":             "🏓 *Đang ping\\.\\.\\.*",
			"basic_pong":                "🏓 *Pong\\!* `%v`",
			"basic_speedtest_running":   "🚀 *Đang chạy Speedtest\\.\\.\\.*",
			"basic_speedtest_error":     "❌ *Lỗi Speedtest*\n\n`%s`",
			"basic_speedtest_result":    "🚀 *Kết quả Speedtest*",
			"basic_name_plain":          "bạn",
			"basic_help_fallback_plain": "📖 Hướng dẫn sử dụng\n\nVui lòng chọn danh mục bên dưới:",

			// system
			"sys_btn_detail":       "📊 Chi tiết",
			"sys_btn_logs":         "📈 Logs",
			"sys_btn_cleanup":      "🧹 Dọn dẹp",
			"sys_toast_refreshed":  "🔄 Đã làm mới",
			"sys_toast_detail":     "📊 Xem chi tiết",
			"sys_toast_sending":    "Đang gửi file...",
			"sys_toast_owner_only": "❌ Chỉ dành cho Owner",
			"sys_toast_cleanup":    "🧹 Đã bắt đầu dọn dẹp",
			"sys_logs_denied_hdr":  "TỪ CHỐI TRUY CẬP",
			"sys_logs_denied":      "Chỉ Admin mới được xem log\\.",
			"sys_logs_missing_hdr": "KHÔNG TÌM THẤY FILE",
			"sys_logs_missing":     "Không tìm thấy file log hoặc file chưa được tạo\\.",
			"sys_logs_read_hdr":    "LỖI ĐỌC FILE",
			"sys_logs_read_err":    "Không đọc được log: %s",
			"sys_logs_write_hdr":   "LỖI GHI FILE",
			"sys_logs_write_err":   "Không tạo được file tạm: %s",
			"sys_logs_send_hdr":    "LỖI GỬI FILE",
			"sys_logs_send_err":    "Không gửi được file: %s",
			"sys_logs_caption":     "📝 *Log ứng dụng:* `%s`",

			// dashboard
			"bdash_toast_page":      "Trang %d",
			"bdash_status_hdr":      "TRẠNG THÁI",
			"bdash_status_none":     "Không có Task nào đang chạy\\.",
			"bdash_toast_empty":     "📭 Trống",
			"bdash_toast_updated":   "🔄 Đã cập nhật",
			"bdash_toast_pong":      "🏓 Pong!",
			"bdash_toast_speedtest": "🚀 Đang kiểm tra tốc độ...",
			"bdash_toast_error":     "❌ Lỗi",
			"bdash_toast_stats":     "📊 Thống kê",
			"bdash_storage_title":   "💾 *Nhà cung cấp lưu trữ khả dụng*",
			"bdash_storage_current": "*Hiện tại:* `%s`",
			"bdash_toast_storage":   "💾 Lưu trữ",
			"bdash_toast_system":    "🖥️ Hệ thống",
			"bdash_files_loading":   "🔍 *Đang tải danh sách file\\.\\.\\.*",
			"bdash_files_error":     "❌ *Không tải được danh sách file*\n\nLỗi: %s",
			"bdash_toast_files":     "📁 Danh sách Drive",
			"bdash_media_hdr":       "XỬ LÝ MEDIA",
			"bdash_media_body": "Các thao tác media cho video/audio:\n\n" +
				"• `/extractaudio` ─ Tách audio từ video\n" +
				"• `/compress` ─ Nén dung lượng video\n" +
				"• `/thumbnail` ─ Tạo thumbnail\n" +
				"• `/screenshots` ─ Chụp nhiều ảnh màn hình\n" +
				"• `/subtitle` ─ Soft\\-sub \\(Gắn subtitle\\)\n" +
				"• `/hardsub` ─ Hard\\-sub \\(Ghép cứng subtitle\\)\n" +
				"• `/rescale` ─ Đổi độ phân giải video\n" +
				"• `/convert` ─ Chuyển đổi định dạng file\n" +
				"• `/mediainfo` ─ Thông tin chi tiết file\n\n" +
				"💡 *Cách dùng:* Reply vào file video bằng lệnh tương ứng",
			"bdash_toast_media":  "🎵 Menu Media",
			"bdash_unknown_hdr":  "KHÔNG XÁC ĐỊNH",
			"bdash_unknown_body": "Thao tác không xác định\\.",

			"bdash_mode_mirror_hdr": "CHẾ ĐỘ MIRROR",
			"bdash_mode_mirror_body": "Tải file thẳng lên Google Drive\\.\n\n" +
				"📌 *CÁCH DÙNG*\n" +
				"• Reply vào file bằng `/mirror`\n" +
				"• Hoặc gửi: `/mirror URL`\n\n" +
				"🏷️ *FLAGS*\n" +
				"• `\\-z` ─ Nén Zip trước khi tải lên\n" +
				"• `\\-uz` ─ Giải nén sau khi tải xuống\n" +
				"• `\\-p PASS` ─ Mật khẩu zip",
			"bdash_mode_leech_hdr": "CHẾ ĐỘ LEECH",
			"bdash_mode_leech_body": "Tải file từ URL về server\\.\n\n" +
				"✅ *HỖ TRỢ*\n" +
				"• HTTP/HTTPS • FTP • Magnet\n\n" +
				"📌 *CÁCH DÙNG*\n" +
				"• `/leech` \\<URL\\>",
			"bdash_mode_ytdlp_hdr": "CHẾ ĐỘ YT-DLP",
			"bdash_mode_ytdlp_body": "Tải video từ hơn 1000 trang web\\.\n\n" +
				"✅ *HỖ TRỢ*\n" +
				"• YouTube • Twitter • TikTok • v\\.v\\.\n\n" +
				"📌 *CÁCH DÙNG*\n" +
				"• `/ytdlp` \\<URL\\>",
			"bdash_mode_torrent_hdr": "CHẾ ĐỘ TORRENT",
			"bdash_mode_torrent_body": "Tải qua Magnet hoặc file Torrent\\.\n\n" +
				"📌 *CÁCH DÙNG*\n" +
				"• `/torrent` \\<magnet\\_link\\>\n" +
				"• Reply vào file `.torrent` bằng `/torrent`",
			"bdash_mode_clone_hdr": "CHẾ ĐỘ CLONE",
			"bdash_mode_clone_body": "Clone file/thư mục Google Drive\\.\n\n" +
				"📌 *CÁCH DÙNG*\n" +
				"• `/clone` \\<GDRIVE\\_URL\\>",
			"bdash_mode_batch_hdr": "CHẾ ĐỘ BATCH",
			"bdash_mode_batch_body": "Tải nhiều URL cùng lúc\\.\n\n" +
				"📌 *CÁCH DÙNG*\n" +
				"`/batch`\n" +
				"`URL1`\n" +
				"`URL2`",
			"bdash_mode_search_hdr": "CHẾ ĐỘ TÌM KIẾM",
			"bdash_mode_search_body": "Tìm Torrent từ nhiều nguồn\\.\n\n" +
				"📌 *CÁCH DÙNG*\n" +
				"• `/search` \\<keyword\\>",
			"bdash_mode_hardsub_hdr": "CHẾ ĐỘ HARD-SUB",
			"bdash_mode_hardsub_body": "Ghép cứng subtitle vĩnh viễn vào video\\.\n\n" +
				"📌 *CÁCH DÙNG*\n" +
				"• Reply vào video bằng `/hardsub subtitle.srt`\n" +
				"• Hoặc: `/hardsub video.mp4 subtitle.srt`\n\n" +
				"🔥 *THÔNG TIN*\n" +
				"Quá trình này cần re\\-encode video nên sẽ mất nhiều thời gian hơn so với soft\\-sub\\.",
			"bdash_mode_rescale_hdr": "CHẾ ĐỘ RESCALE",
			"bdash_mode_rescale_body": "Đổi độ phân giải video \\(Transcoding\\)\\.\n\n" +
				"📌 *CÁCH DÙNG*\n" +
				"• Reply vào video bằng `/rescale 720p`\n" +
				"• Hoặc: `/rescale video.mp4 1280x720`\n\n" +
				"🏷️ *PRESET*\n" +
				"`4k`, `2k`, `1080p`, `720p`, `480p`, `360p` hoặc định dạng tùy chỉnh `WxH`\\.",

			// join
			"join_usage":        "⚠️ *Sai định dạng*\n\nDùng: `/join <link invite>`\nVí dụ: `/join https://t.me/+FnboT0dqFiozNzA1`",
			"join_hash_error":   "❌ *Lỗi:* Không lấy được hash mời từ liên kết",
			"join_trying":       "🔄 *Đang thử tham gia channel\\.\\.\\.*\nHash: `%s`",
			"join_trying_plain": "🔄 Đang thử tham gia channel...\nHash: %s",
			"join_already":      "✅ *Đã tham gia\\!* Bot đã ở trong channel này rồi\\.",
			"join_failed":       "❌ *Tham gia thất bại*\nLỗi: `%s`",
			"join_success":      "✅ *Thành công\\!* %s",

			// notify
			"notify_cmd_usage_plain":       "🔔 WEBHOOK THÔNG BÁO\n\nDùng: /notify <webhook_url>\nXóa: /notify off\n\nWebhook sẽ nhận POST JSON khi Task của bạn hoàn thành/thất bại.",
			"notify_cmd_off_plain":         "🔕 Đã tắt webhook thông báo (%d Task đang chạy)",
			"notify_cmd_invalid_url_plain": "❌ URL không hợp lệ — phải là http/https",
			"notify_cmd_saved":             "✅ Đã lưu webhook cho %d Task đang chạy\n\nURL: `%s`\nWebhook sẽ nhận POST JSON khi Task hoàn thành/thất bại\\.",

			// schedule
			"schedule_usage":        "Cách dùng: `/schedule <type> <HH:MM> <url> [filename]`\nLoại: mirror, zip, unzip, mp3, audio, video",
			"schedule_invalid_type": "Loại không hợp lệ: %s\\. Hợp lệ: mirror, zip, unzip, mp3, audio, video",
			"schedule_invalid_time": "Định dạng giờ không hợp lệ\\. Dùng HH:MM \\(24 giờ\\)",
			"schedule_save_failed":  "Không thể lên lịch Task",
			"schedule_ok":           "Đã lên lịch Task lúc %s",
		},
		"en": {
			// stats
			"stats_load_failed":         "❌ *Could not load statistics*",
			"stats_fallback_plain":      "❌ Could not format statistics (Markdown error)\n\nFallback: short statistics\nTotal Tasks: %d",
			"stats_btn_my":              "👤 Mine",
			"stats_btn_today":           "📊 Today",
			"stats_btn_week":            "📈 Week",
			"stats_btn_month":           "📉 Month",
			"stats_title":               "📊 *DASHBOARD STATISTICS*",
			"stats_sec_global":          "🌐 *GLOBAL STATISTICS*",
			"stats_total_tasks":         "📥 Total Tasks: `%d`",
			"stats_completed":           "✅ Completed: `%d`",
			"stats_failed":              "❌ Failed: `%d`",
			"stats_success_rate":        "📊 Success rate: `%.1f%%`",
			"stats_bandwidth":           "📈 Bandwidth: `%s`",
			"stats_bandwidth_alt":       "📊 Bandwidth: `%s`",
			"stats_sec_today_global":    "📅 *TODAY \\(GLOBAL\\)*",
			"stats_tasks_short":         "📥 Tasks: `%d`",
			"stats_sec_today_me":        "👤 *MY TODAY*",
			"stats_sec_total_me":        "👤 *MY TOTAL STATISTICS*",
			"stats_last_active_short":   "🕐 Active: `%s`",
			"stats_none_user":           "📭 *You have no statistics yet*",
			"stats_user_title":          "👤 *USER STATISTICS*",
			"stats_sec_user_info":       "🆔 *USER INFORMATION*",
			"stats_user_name":           "👤 Name: @%s",
			"stats_user_id":             "🆔 ID: `%d`",
			"stats_last_active":         "🕐 Last active: `%s`",
			"stats_sec_activity":        "📊 *ACTIVITY*",
			"stats_succeeded":           "✅ Succeeded: `%d`",
			"stats_none":                "📭 *No statistics yet*",
			"stats_daily_title":         "📅 *TODAY'S STATISTICS*",
			"stats_sec_global_activity": "🌐 *GLOBAL ACTIVITY*",
			"stats_avg_speed":           "⚡ Avg speed: `%s`",
			"stats_peak_concurrent":     "🔥 Peak concurrent: `%d`",
			"stats_sec_my_today":        "👤 *YOUR ACTIVITY TODAY*",
			"stats_weekly_title":        "📈 *WEEKLY STATISTICS*",
			"stats_sec_7days":           "🗓️ *LAST 7 DAYS*",
			"stats_day_row":             "%s `%s` : %d Tasks",
			"stats_sec_summary":         "📊 *SUMMARY*",
			"stats_total_line":          "📥 Total: `%d` Tasks",
			"stats_monthly_title":       "📉 *MONTHLY STATISTICS*",
			"stats_sec_by_week":         "📅 *BY WEEK*",
			"stats_week_row":            "Week %d: `%d` Tasks\n└ `%s`",
			"stats_err_my":              "❌ *Could not load your statistics*",
			"stats_err_today":           "❌ *Could not load today's statistics*",
			"stats_err_week":            "❌ *Could not load weekly statistics*",
			"stats_err_month":           "❌ *Could not load monthly statistics*",
			"stats_toast_refreshed":     "🔄 Statistics refreshed!",

			// shared basic bits
			"basic_btn_refresh":         "🔄 Refresh",
			"basic_toast_closed":        "Closed",
			"basic_pinging":             "🏓 *Pinging\\.\\.\\.*",
			"basic_pong":                "🏓 *Pong\\!* `%v`",
			"basic_speedtest_running":   "🚀 *Running Speedtest\\.\\.\\.*",
			"basic_speedtest_error":     "❌ *Speedtest error*\n\n`%s`",
			"basic_speedtest_result":    "🚀 *Speedtest result*",
			"basic_name_plain":          "friend",
			"basic_help_fallback_plain": "📖 User guide\n\nPlease choose a category below:",

			// system
			"sys_btn_detail":       "📊 Details",
			"sys_btn_logs":         "📈 Logs",
			"sys_btn_cleanup":      "🧹 Clean up",
			"sys_toast_refreshed":  "🔄 Refreshed",
			"sys_toast_detail":     "📊 Viewing details",
			"sys_toast_sending":    "Sending file...",
			"sys_toast_owner_only": "❌ Owner only",
			"sys_toast_cleanup":    "🧹 Cleanup started",
			"sys_logs_denied_hdr":  "ACCESS DENIED",
			"sys_logs_denied":      "Only Admins can view logs\\.",
			"sys_logs_missing_hdr": "FILE NOT FOUND",
			"sys_logs_missing":     "Log file not found or not created yet\\.",
			"sys_logs_read_hdr":    "FILE READ ERROR",
			"sys_logs_read_err":    "Could not read log: %s",
			"sys_logs_write_hdr":   "FILE WRITE ERROR",
			"sys_logs_write_err":   "Could not create temp file: %s",
			"sys_logs_send_hdr":    "FILE SEND ERROR",
			"sys_logs_send_err":    "Could not send file: %s",
			"sys_logs_caption":     "📝 *Application log:* `%s`",

			// dashboard
			"bdash_toast_page":      "Page %d",
			"bdash_status_hdr":      "STATUS",
			"bdash_status_none":     "No Tasks are running\\.",
			"bdash_toast_empty":     "📭 Empty",
			"bdash_toast_updated":   "🔄 Updated",
			"bdash_toast_pong":      "🏓 Pong!",
			"bdash_toast_speedtest": "🚀 Testing speed...",
			"bdash_toast_error":     "❌ Error",
			"bdash_toast_stats":     "📊 Statistics",
			"bdash_storage_title":   "💾 *Available storage providers*",
			"bdash_storage_current": "*Current:* `%s`",
			"bdash_toast_storage":   "💾 Storage",
			"bdash_toast_system":    "🖥️ System",
			"bdash_files_loading":   "🔍 *Loading file list\\.\\.\\.*",
			"bdash_files_error":     "❌ *Could not load the file list*\n\nError: %s",
			"bdash_toast_files":     "📁 Drive files",
			"bdash_media_hdr":       "MEDIA PROCESSING",
			"bdash_media_body": "Media operations for video/audio:\n\n" +
				"• `/extractaudio` ─ Extract audio from video\n" +
				"• `/compress` ─ Reduce video file size\n" +
				"• `/thumbnail` ─ Create a thumbnail\n" +
				"• `/screenshots` ─ Capture multiple screenshots\n" +
				"• `/subtitle` ─ Soft\\-sub \\(attach subtitles\\)\n" +
				"• `/hardsub` ─ Hard\\-sub \\(burn subtitles in\\)\n" +
				"• `/rescale` ─ Change video resolution\n" +
				"• `/convert` ─ Convert file format\n" +
				"• `/mediainfo` ─ Detailed file information\n\n" +
				"💡 *Usage:* Reply to a video file with the matching command",
			"bdash_toast_media":  "🎵 Media menu",
			"bdash_unknown_hdr":  "UNKNOWN",
			"bdash_unknown_body": "Unknown action\\.",

			"bdash_mode_mirror_hdr": "MIRROR MODE",
			"bdash_mode_mirror_body": "Upload files straight to Google Drive\\.\n\n" +
				"📌 *USAGE*\n" +
				"• Reply to a file with `/mirror`\n" +
				"• Or send: `/mirror URL`\n\n" +
				"🏷️ *FLAGS*\n" +
				"• `\\-z` ─ Zip before uploading\n" +
				"• `\\-uz` ─ Unzip after downloading\n" +
				"• `\\-p PASS` ─ Zip password",
			"bdash_mode_leech_hdr": "LEECH MODE",
			"bdash_mode_leech_body": "Download files from a URL to the server\\.\n\n" +
				"✅ *SUPPORTED*\n" +
				"• HTTP/HTTPS • FTP • Magnet\n\n" +
				"📌 *USAGE*\n" +
				"• `/leech` \\<URL\\>",
			"bdash_mode_ytdlp_hdr": "YT-DLP MODE",
			"bdash_mode_ytdlp_body": "Download videos from over 1000 websites\\.\n\n" +
				"✅ *SUPPORTED*\n" +
				"• YouTube • Twitter • TikTok • etc\\.\n\n" +
				"📌 *USAGE*\n" +
				"• `/ytdlp` \\<URL\\>",
			"bdash_mode_torrent_hdr": "TORRENT MODE",
			"bdash_mode_torrent_body": "Download via a Magnet link or a Torrent file\\.\n\n" +
				"📌 *USAGE*\n" +
				"• `/torrent` \\<magnet\\_link\\>\n" +
				"• Reply to a `.torrent` file with `/torrent`",
			"bdash_mode_clone_hdr": "CLONE MODE",
			"bdash_mode_clone_body": "Clone Google Drive files/folders\\.\n\n" +
				"📌 *USAGE*\n" +
				"• `/clone` \\<GDRIVE\\_URL\\>",
			"bdash_mode_batch_hdr": "BATCH MODE",
			"bdash_mode_batch_body": "Download multiple URLs at once\\.\n\n" +
				"📌 *USAGE*\n" +
				"`/batch`\n" +
				"`URL1`\n" +
				"`URL2`",
			"bdash_mode_search_hdr": "SEARCH MODE",
			"bdash_mode_search_body": "Search Torrents from multiple sources\\.\n\n" +
				"📌 *USAGE*\n" +
				"• `/search` \\<keyword\\>",
			"bdash_mode_hardsub_hdr": "HARD-SUB MODE",
			"bdash_mode_hardsub_body": "Permanently burn subtitles into the video\\.\n\n" +
				"📌 *USAGE*\n" +
				"• Reply to a video with `/hardsub subtitle.srt`\n" +
				"• Or: `/hardsub video.mp4 subtitle.srt`\n\n" +
				"🔥 *INFO*\n" +
				"This process re\\-encodes the video, so it takes longer than soft\\-sub\\.",
			"bdash_mode_rescale_hdr": "RESCALE MODE",
			"bdash_mode_rescale_body": "Change video resolution \\(transcoding\\)\\.\n\n" +
				"📌 *USAGE*\n" +
				"• Reply to a video with `/rescale 720p`\n" +
				"• Or: `/rescale video.mp4 1280x720`\n\n" +
				"🏷️ *PRESET*\n" +
				"`4k`, `2k`, `1080p`, `720p`, `480p`, `360p` or a custom `WxH` format\\.",

			// join
			"join_usage":        "⚠️ *Invalid format*\n\nUsage: `/join <invite link>`\nExample: `/join https://t.me/+FnboT0dqFiozNzA1`",
			"join_hash_error":   "❌ *Error:* Could not extract the invite hash from the link",
			"join_trying":       "🔄 *Trying to join the channel\\.\\.\\.*\nHash: `%s`",
			"join_trying_plain": "🔄 Trying to join the channel...\nHash: %s",
			"join_already":      "✅ *Already joined\\!* The bot is already in this channel\\.",
			"join_failed":       "❌ *Join failed*\nError: `%s`",
			"join_success":      "✅ *Success\\!* %s",

			// notify
			"notify_cmd_usage_plain":       "🔔 NOTIFICATION WEBHOOK\n\nUsage: /notify <webhook_url>\nRemove: /notify off\n\nThe webhook receives a JSON POST when your Task completes/fails.",
			"notify_cmd_off_plain":         "🔕 Notification webhook disabled (%d running Tasks)",
			"notify_cmd_invalid_url_plain": "❌ Invalid URL — must be http/https",
			"notify_cmd_saved":             "✅ Webhook saved for %d running Tasks\n\nURL: `%s`\nThe webhook will receive a JSON POST when a Task completes/fails\\.",

			// schedule
			"schedule_usage":        "Usage: `/schedule <type> <HH:MM> <url> [filename]`\nTypes: mirror, zip, unzip, mp3, audio, video",
			"schedule_invalid_type": "Invalid type: %s\\. Valid: mirror, zip, unzip, mp3, audio, video",
			"schedule_invalid_time": "Invalid time format\\. Use HH:MM \\(24\\-hour\\)",
			"schedule_save_failed":  "Could not schedule the Task",
			"schedule_ok":           "Task scheduled for %s",
		},
		"id": {
			// stats
			"stats_load_failed":         "❌ *Gagal memuat statistik*",
			"stats_fallback_plain":      "❌ Gagal memformat statistik (kesalahan Markdown)\n\nCadangan: statistik ringkas\nTotal Task: %d",
			"stats_btn_my":              "👤 Saya",
			"stats_btn_today":           "📊 Hari ini",
			"stats_btn_week":            "📈 Minggu",
			"stats_btn_month":           "📉 Bulan",
			"stats_title":               "📊 *STATISTIK DASHBOARD*",
			"stats_sec_global":          "🌐 *STATISTIK GLOBAL*",
			"stats_total_tasks":         "📥 Total Task: `%d`",
			"stats_completed":           "✅ Selesai: `%d`",
			"stats_failed":              "❌ Gagal: `%d`",
			"stats_success_rate":        "📊 Tingkat keberhasilan: `%.1f%%`",
			"stats_bandwidth":           "📈 Bandwidth: `%s`",
			"stats_bandwidth_alt":       "📊 Bandwidth: `%s`",
			"stats_sec_today_global":    "📅 *HARI INI \\(GLOBAL\\)*",
			"stats_tasks_short":         "📥 Task: `%d`",
			"stats_sec_today_me":        "👤 *HARI INI SAYA*",
			"stats_sec_total_me":        "👤 *TOTAL STATISTIK SAYA*",
			"stats_last_active_short":   "🕐 Aktif: `%s`",
			"stats_none_user":           "📭 *Anda belum memiliki statistik*",
			"stats_user_title":          "👤 *STATISTIK PENGGUNA*",
			"stats_sec_user_info":       "🆔 *INFORMASI PENGGUNA*",
			"stats_user_name":           "👤 Nama: @%s",
			"stats_user_id":             "🆔 ID: `%d`",
			"stats_last_active":         "🕐 Terakhir aktif: `%s`",
			"stats_sec_activity":        "📊 *AKTIVITAS*",
			"stats_succeeded":           "✅ Berhasil: `%d`",
			"stats_none":                "📭 *Belum ada statistik*",
			"stats_daily_title":         "📅 *STATISTIK HARI INI*",
			"stats_sec_global_activity": "🌐 *AKTIVITAS GLOBAL*",
			"stats_avg_speed":           "⚡ Kecepatan rata\\-rata: `%s`",
			"stats_peak_concurrent":     "🔥 Puncak bersamaan: `%d`",
			"stats_sec_my_today":        "👤 *AKTIVITAS ANDA HARI INI*",
			"stats_weekly_title":        "📈 *STATISTIK MINGGUAN*",
			"stats_sec_7days":           "🗓️ *7 HARI TERAKHIR*",
			"stats_day_row":             "%s `%s` : %d Task",
			"stats_sec_summary":         "📊 *RINGKASAN*",
			"stats_total_line":          "📥 Total: `%d` Task",
			"stats_monthly_title":       "📉 *STATISTIK BULANAN*",
			"stats_sec_by_week":         "📅 *PER MINGGU*",
			"stats_week_row":            "Minggu %d: `%d` Task\n└ `%s`",
			"stats_err_my":              "❌ *Gagal memuat statistik Anda*",
			"stats_err_today":           "❌ *Gagal memuat statistik hari ini*",
			"stats_err_week":            "❌ *Gagal memuat statistik mingguan*",
			"stats_err_month":           "❌ *Gagal memuat statistik bulanan*",
			"stats_toast_refreshed":     "🔄 Statistik diperbarui!",

			// shared basic bits
			"basic_btn_refresh":         "🔄 Segarkan",
			"basic_toast_closed":        "Ditutup",
			"basic_pinging":             "🏓 *Melakukan ping\\.\\.\\.*",
			"basic_pong":                "🏓 *Pong\\!* `%v`",
			"basic_speedtest_running":   "🚀 *Menjalankan Speedtest\\.\\.\\.*",
			"basic_speedtest_error":     "❌ *Kesalahan Speedtest*\n\n`%s`",
			"basic_speedtest_result":    "🚀 *Hasil Speedtest*",
			"basic_name_plain":          "teman",
			"basic_help_fallback_plain": "📖 Panduan penggunaan\n\nSilakan pilih kategori di bawah:",

			// system
			"sys_btn_detail":       "📊 Detail",
			"sys_btn_logs":         "📈 Log",
			"sys_btn_cleanup":      "🧹 Bersihkan",
			"sys_toast_refreshed":  "🔄 Disegarkan",
			"sys_toast_detail":     "📊 Lihat detail",
			"sys_toast_sending":    "Mengirim file...",
			"sys_toast_owner_only": "❌ Hanya untuk Owner",
			"sys_toast_cleanup":    "🧹 Pembersihan dimulai",
			"sys_logs_denied_hdr":  "AKSES DITOLAK",
			"sys_logs_denied":      "Hanya Admin yang dapat melihat log\\.",
			"sys_logs_missing_hdr": "FILE TIDAK DITEMUKAN",
			"sys_logs_missing":     "File log tidak ditemukan atau belum dibuat\\.",
			"sys_logs_read_hdr":    "KESALAHAN MEMBACA FILE",
			"sys_logs_read_err":    "Tidak dapat membaca log: %s",
			"sys_logs_write_hdr":   "KESALAHAN MENULIS FILE",
			"sys_logs_write_err":   "Tidak dapat membuat file sementara: %s",
			"sys_logs_send_hdr":    "KESALAHAN MENGIRIM FILE",
			"sys_logs_send_err":    "Tidak dapat mengirim file: %s",
			"sys_logs_caption":     "📝 *Log aplikasi:* `%s`",

			// dashboard
			"bdash_toast_page":      "Halaman %d",
			"bdash_status_hdr":      "STATUS",
			"bdash_status_none":     "Tidak ada Task yang sedang berjalan\\.",
			"bdash_toast_empty":     "📭 Kosong",
			"bdash_toast_updated":   "🔄 Diperbarui",
			"bdash_toast_pong":      "🏓 Pong!",
			"bdash_toast_speedtest": "🚀 Menguji kecepatan...",
			"bdash_toast_error":     "❌ Kesalahan",
			"bdash_toast_stats":     "📊 Statistik",
			"bdash_storage_title":   "💾 *Penyedia penyimpanan yang tersedia*",
			"bdash_storage_current": "*Saat ini:* `%s`",
			"bdash_toast_storage":   "💾 Penyimpanan",
			"bdash_toast_system":    "🖥️ Sistem",
			"bdash_files_loading":   "🔍 *Memuat daftar file\\.\\.\\.*",
			"bdash_files_error":     "❌ *Gagal memuat daftar file*\n\nKesalahan: %s",
			"bdash_toast_files":     "📁 Daftar Drive",
			"bdash_media_hdr":       "PEMROSESAN MEDIA",
			"bdash_media_body": "Operasi media untuk video/audio:\n\n" +
				"• `/extractaudio` ─ Ekstrak audio dari video\n" +
				"• `/compress` ─ Kompres ukuran video\n" +
				"• `/thumbnail` ─ Buat thumbnail\n" +
				"• `/screenshots` ─ Ambil beberapa screenshot\n" +
				"• `/subtitle` ─ Soft\\-sub \\(pasang subtitle\\)\n" +
				"• `/hardsub` ─ Hard\\-sub \\(tanam subtitle permanen\\)\n" +
				"• `/rescale` ─ Ubah resolusi video\n" +
				"• `/convert` ─ Konversi format file\n" +
				"• `/mediainfo` ─ Informasi detail file\n\n" +
				"💡 *Cara pakai:* Reply ke file video dengan perintah yang sesuai",
			"bdash_toast_media":  "🎵 Menu Media",
			"bdash_unknown_hdr":  "TIDAK DIKENAL",
			"bdash_unknown_body": "Aksi tidak dikenal\\.",

			"bdash_mode_mirror_hdr": "MODE MIRROR",
			"bdash_mode_mirror_body": "Unggah file langsung ke Google Drive\\.\n\n" +
				"📌 *CARA PAKAI*\n" +
				"• Reply ke file dengan `/mirror`\n" +
				"• Atau kirim: `/mirror URL`\n\n" +
				"🏷️ *FLAGS*\n" +
				"• `\\-z` ─ Kompres Zip sebelum diunggah\n" +
				"• `\\-uz` ─ Ekstrak setelah diunduh\n" +
				"• `\\-p PASS` ─ Password zip",
			"bdash_mode_leech_hdr": "MODE LEECH",
			"bdash_mode_leech_body": "Unduh file dari URL ke server\\.\n\n" +
				"✅ *DIDUKUNG*\n" +
				"• HTTP/HTTPS • FTP • Magnet\n\n" +
				"📌 *CARA PAKAI*\n" +
				"• `/leech` \\<URL\\>",
			"bdash_mode_ytdlp_hdr": "MODE YT-DLP",
			"bdash_mode_ytdlp_body": "Unduh video dari lebih dari 1000 situs web\\.\n\n" +
				"✅ *DIDUKUNG*\n" +
				"• YouTube • Twitter • TikTok • dll\\.\n\n" +
				"📌 *CARA PAKAI*\n" +
				"• `/ytdlp` \\<URL\\>",
			"bdash_mode_torrent_hdr": "MODE TORRENT",
			"bdash_mode_torrent_body": "Unduh melalui Magnet atau file Torrent\\.\n\n" +
				"📌 *CARA PAKAI*\n" +
				"• `/torrent` \\<magnet\\_link\\>\n" +
				"• Reply ke file `.torrent` dengan `/torrent`",
			"bdash_mode_clone_hdr": "MODE CLONE",
			"bdash_mode_clone_body": "Clone file/folder Google Drive\\.\n\n" +
				"📌 *CARA PAKAI*\n" +
				"• `/clone` \\<GDRIVE\\_URL\\>",
			"bdash_mode_batch_hdr": "MODE BATCH",
			"bdash_mode_batch_body": "Unduh banyak URL sekaligus\\.\n\n" +
				"📌 *CARA PAKAI*\n" +
				"`/batch`\n" +
				"`URL1`\n" +
				"`URL2`",
			"bdash_mode_search_hdr": "MODE PENCARIAN",
			"bdash_mode_search_body": "Cari Torrent dari berbagai sumber\\.\n\n" +
				"📌 *CARA PAKAI*\n" +
				"• `/search` \\<keyword\\>",
			"bdash_mode_hardsub_hdr": "MODE HARD-SUB",
			"bdash_mode_hardsub_body": "Tanam subtitle secara permanen ke dalam video\\.\n\n" +
				"📌 *CARA PAKAI*\n" +
				"• Reply ke video dengan `/hardsub subtitle.srt`\n" +
				"• Atau: `/hardsub video.mp4 subtitle.srt`\n\n" +
				"🔥 *INFO*\n" +
				"Proses ini perlu re\\-encode video sehingga membutuhkan waktu lebih lama dibanding soft\\-sub\\.",
			"bdash_mode_rescale_hdr": "MODE RESCALE",
			"bdash_mode_rescale_body": "Ubah resolusi video \\(Transcoding\\)\\.\n\n" +
				"📌 *CARA PAKAI*\n" +
				"• Reply ke video dengan `/rescale 720p`\n" +
				"• Atau: `/rescale video.mp4 1280x720`\n\n" +
				"🏷️ *PRESET*\n" +
				"`4k`, `2k`, `1080p`, `720p`, `480p`, `360p` atau format kustom `WxH`\\.",

			// join
			"join_usage":        "⚠️ *Format salah*\n\nPenggunaan: `/join <link undangan>`\nContoh: `/join https://t.me/+FnboT0dqFiozNzA1`",
			"join_hash_error":   "❌ *Kesalahan:* Tidak dapat mengambil hash undangan dari tautan",
			"join_trying":       "🔄 *Mencoba bergabung ke channel\\.\\.\\.*\nHash: `%s`",
			"join_trying_plain": "🔄 Mencoba bergabung ke channel...\nHash: %s",
			"join_already":      "✅ *Sudah bergabung\\!* Bot sudah berada di channel ini\\.",
			"join_failed":       "❌ *Gagal bergabung*\nKesalahan: `%s`",
			"join_success":      "✅ *Berhasil\\!* %s",

			// notify
			"notify_cmd_usage_plain":       "🔔 WEBHOOK NOTIFIKASI\n\nPenggunaan: /notify <webhook_url>\nHapus: /notify off\n\nWebhook akan menerima POST JSON saat Task Anda selesai/gagal.",
			"notify_cmd_off_plain":         "🔕 Webhook notifikasi dinonaktifkan (%d Task berjalan)",
			"notify_cmd_invalid_url_plain": "❌ URL tidak valid — harus http/https",
			"notify_cmd_saved":             "✅ Webhook disimpan untuk %d Task yang berjalan\n\nURL: `%s`\nWebhook akan menerima POST JSON saat Task selesai/gagal\\.",

			// schedule
			"schedule_usage":        "Cara pakai: `/schedule <type> <HH:MM> <url> [filename]`\nJenis: mirror, zip, unzip, mp3, audio, video",
			"schedule_invalid_type": "Jenis tidak valid: %s\\. Yang valid: mirror, zip, unzip, mp3, audio, video",
			"schedule_invalid_time": "Format jam tidak valid\\. Gunakan HH:MM \\(24 jam\\)",
			"schedule_save_failed":  "Tidak dapat menjadwalkan Task",
			"schedule_ok":           "Task dijadwalkan pada %s",
		},
		"ja": {
			// stats
			"stats_load_failed":         "❌ *統計を取得できませんでした*",
			"stats_fallback_plain":      "❌ 統計を整形できませんでした（Markdown エラー）\n\n代替表示: 簡易統計\n合計タスク: %d",
			"stats_btn_my":              "👤 自分",
			"stats_btn_today":           "📊 今日",
			"stats_btn_week":            "📈 週",
			"stats_btn_month":           "📉 月",
			"stats_title":               "📊 *ダッシュボード統計*",
			"stats_sec_global":          "🌐 *全体の統計*",
			"stats_total_tasks":         "📥 合計タスク: `%d`",
			"stats_completed":           "✅ 完了: `%d`",
			"stats_failed":              "❌ 失敗: `%d`",
			"stats_success_rate":        "📊 成功率: `%.1f%%`",
			"stats_bandwidth":           "📈 帯域幅: `%s`",
			"stats_bandwidth_alt":       "📊 帯域幅: `%s`",
			"stats_sec_today_global":    "📅 *今日（全体）*",
			"stats_tasks_short":         "📥 タスク: `%d`",
			"stats_sec_today_me":        "👤 *自分の今日*",
			"stats_sec_total_me":        "👤 *自分の統計合計*",
			"stats_last_active_short":   "🕐 稼働: `%s`",
			"stats_none_user":           "📭 *まだ統計がありません*",
			"stats_user_title":          "👤 *ユーザー統計*",
			"stats_sec_user_info":       "🆔 *ユーザー情報*",
			"stats_user_name":           "👤 名前: @%s",
			"stats_user_id":             "🆔 ID: `%d`",
			"stats_last_active":         "🕐 最終アクティブ: `%s`",
			"stats_sec_activity":        "📊 *アクティビティ*",
			"stats_succeeded":           "✅ 成功: `%d`",
			"stats_none":                "📭 *統計はまだありません*",
			"stats_daily_title":         "📅 *今日の統計*",
			"stats_sec_global_activity": "🌐 *全体のアクティビティ*",
			"stats_avg_speed":           "⚡ 平均速度: `%s`",
			"stats_peak_concurrent":     "🔥 最大同時実行数: `%d`",
			"stats_sec_my_today":        "👤 *今日のあなたのアクティビティ*",
			"stats_weekly_title":        "📈 *週間統計*",
			"stats_sec_7days":           "🗓️ *過去7日間*",
			"stats_day_row":             "%s `%s` : %d タスク",
			"stats_sec_summary":         "📊 *まとめ*",
			"stats_total_line":          "📥 合計: `%d` タスク",
			"stats_monthly_title":       "📉 *月間統計*",
			"stats_sec_by_week":         "📅 *週ごと*",
			"stats_week_row":            "第%d週: `%d` タスク\n└ `%s`",
			"stats_err_my":              "❌ *あなたの統計を取得できませんでした*",
			"stats_err_today":           "❌ *今日の統計を取得できませんでした*",
			"stats_err_week":            "❌ *週間統計を取得できませんでした*",
			"stats_err_month":           "❌ *月間統計を取得できませんでした*",
			"stats_toast_refreshed":     "🔄 統計を更新しました！",

			// shared basic bits
			"basic_btn_refresh":         "🔄 更新",
			"basic_toast_closed":        "閉じました",
			"basic_pinging":             "🏓 *Ping 中…*",
			"basic_pong":                "🏓 *Pong！* `%v`",
			"basic_speedtest_running":   "🚀 *Speedtest を実行中…*",
			"basic_speedtest_error":     "❌ *Speedtest エラー*\n\n`%s`",
			"basic_speedtest_result":    "🚀 *Speedtest 結果*",
			"basic_name_plain":          "ユーザー",
			"basic_help_fallback_plain": "📖 使い方ガイド\n\n下のカテゴリを選択してください:",

			// system
			"sys_btn_detail":       "📊 詳細",
			"sys_btn_logs":         "📈 ログ",
			"sys_btn_cleanup":      "🧹 クリーンアップ",
			"sys_toast_refreshed":  "🔄 更新しました",
			"sys_toast_detail":     "📊 詳細を表示",
			"sys_toast_sending":    "ファイルを送信中…",
			"sys_toast_owner_only": "❌ Owner 専用です",
			"sys_toast_cleanup":    "🧹 クリーンアップを開始しました",
			"sys_logs_denied_hdr":  "アクセス拒否",
			"sys_logs_denied":      "ログを閲覧できるのは Admin のみです。",
			"sys_logs_missing_hdr": "ファイルが見つかりません",
			"sys_logs_missing":     "ログファイルが見つからないか、まだ作成されていません。",
			"sys_logs_read_hdr":    "ファイル読み取りエラー",
			"sys_logs_read_err":    "ログを読み取れませんでした: %s",
			"sys_logs_write_hdr":   "ファイル書き込みエラー",
			"sys_logs_write_err":   "一時ファイルを作成できませんでした: %s",
			"sys_logs_send_hdr":    "ファイル送信エラー",
			"sys_logs_send_err":    "ファイルを送信できませんでした: %s",
			"sys_logs_caption":     "📝 *アプリケーションログ:* `%s`",

			// dashboard
			"bdash_toast_page":      "%d ページ目",
			"bdash_status_hdr":      "ステータス",
			"bdash_status_none":     "実行中のタスクはありません。",
			"bdash_toast_empty":     "📭 空です",
			"bdash_toast_updated":   "🔄 更新しました",
			"bdash_toast_pong":      "🏓 Pong!",
			"bdash_toast_speedtest": "🚀 速度を測定中…",
			"bdash_toast_error":     "❌ エラー",
			"bdash_toast_stats":     "📊 統計",
			"bdash_storage_title":   "💾 *利用可能なストレージプロバイダー*",
			"bdash_storage_current": "*現在:* `%s`",
			"bdash_toast_storage":   "💾 ストレージ",
			"bdash_toast_system":    "🖥️ システム",
			"bdash_files_loading":   "🔍 *ファイル一覧を読み込み中…*",
			"bdash_files_error":     "❌ *ファイル一覧を読み込めませんでした*\n\nエラー: %s",
			"bdash_toast_files":     "📁 Drive 一覧",
			"bdash_media_hdr":       "メディア処理",
			"bdash_media_body": "動画/音声のメディア操作:\n\n" +
				"• `/extractaudio` ─ 動画から音声を抽出\n" +
				"• `/compress` ─ 動画の容量を圧縮\n" +
				"• `/thumbnail` ─ サムネイルを作成\n" +
				"• `/screenshots` ─ 複数のスクリーンショットを撮影\n" +
				"• `/subtitle` ─ Soft\\-sub（字幕を付加）\n" +
				"• `/hardsub` ─ Hard\\-sub（字幕を焼き込み）\n" +
				"• `/rescale` ─ 動画の解像度を変更\n" +
				"• `/convert` ─ ファイル形式を変換\n" +
				"• `/mediainfo` ─ ファイルの詳細情報\n\n" +
				"💡 *使い方:* 動画ファイルに該当コマンドで返信してください",
			"bdash_toast_media":  "🎵 メディアメニュー",
			"bdash_unknown_hdr":  "不明",
			"bdash_unknown_body": "不明な操作です。",

			"bdash_mode_mirror_hdr": "Mirror モード",
			"bdash_mode_mirror_body": "ファイルを直接 Google Drive にアップロードします。\n\n" +
				"📌 *使い方*\n" +
				"• ファイルに `/mirror` で返信\n" +
				"• または送信: `/mirror URL`\n\n" +
				"🏷️ *FLAGS*\n" +
				"• `\\-z` ─ アップロード前に Zip 圧縮\n" +
				"• `\\-uz` ─ ダウンロード後に解凍\n" +
				"• `\\-p PASS` ─ Zip のパスワード",
			"bdash_mode_leech_hdr": "Leech モード",
			"bdash_mode_leech_body": "URL からサーバーへファイルをダウンロードします。\n\n" +
				"✅ *対応*\n" +
				"• HTTP/HTTPS • FTP • Magnet\n\n" +
				"📌 *使い方*\n" +
				"• `/leech` \\<URL\\>",
			"bdash_mode_ytdlp_hdr": "YT-DLP モード",
			"bdash_mode_ytdlp_body": "1000 以上のサイトから動画をダウンロードします。\n\n" +
				"✅ *対応*\n" +
				"• YouTube • Twitter • TikTok • など\n\n" +
				"📌 *使い方*\n" +
				"• `/ytdlp` \\<URL\\>",
			"bdash_mode_torrent_hdr": "Torrent モード",
			"bdash_mode_torrent_body": "Magnet または Torrent ファイルでダウンロードします。\n\n" +
				"📌 *使い方*\n" +
				"• `/torrent` \\<magnet\\_link\\>\n" +
				"• `.torrent` ファイルに `/torrent` で返信",
			"bdash_mode_clone_hdr": "Clone モード",
			"bdash_mode_clone_body": "Google Drive のファイル/フォルダーをクローンします。\n\n" +
				"📌 *使い方*\n" +
				"• `/clone` \\<GDRIVE\\_URL\\>",
			"bdash_mode_batch_hdr": "Batch モード",
			"bdash_mode_batch_body": "複数の URL を一度にダウンロードします。\n\n" +
				"📌 *使い方*\n" +
				"`/batch`\n" +
				"`URL1`\n" +
				"`URL2`",
			"bdash_mode_search_hdr": "検索モード",
			"bdash_mode_search_body": "複数のソースから Torrent を検索します。\n\n" +
				"📌 *使い方*\n" +
				"• `/search` \\<keyword\\>",
			"bdash_mode_hardsub_hdr": "Hard-sub モード",
			"bdash_mode_hardsub_body": "字幕を動画に永続的に焼き込みます。\n\n" +
				"📌 *使い方*\n" +
				"• 動画に `/hardsub subtitle.srt` で返信\n" +
				"• または: `/hardsub video.mp4 subtitle.srt`\n\n" +
				"🔥 *情報*\n" +
				"この処理では動画の再エンコードが必要なため、soft\\-sub よりも時間がかかります。",
			"bdash_mode_rescale_hdr": "Rescale モード",
			"bdash_mode_rescale_body": "動画の解像度を変更します（トランスコード）。\n\n" +
				"📌 *使い方*\n" +
				"• 動画に `/rescale 720p` で返信\n" +
				"• または: `/rescale video.mp4 1280x720`\n\n" +
				"🏷️ *PRESET*\n" +
				"`4k`, `2k`, `1080p`, `720p`, `480p`, `360p` またはカスタム形式 `WxH`。",

			// join
			"join_usage":        "⚠️ *形式が正しくありません*\n\n使い方: `/join <招待リンク>`\n例: `/join https://t.me/+FnboT0dqFiozNzA1`",
			"join_hash_error":   "❌ *エラー:* リンクから招待ハッシュを取得できませんでした",
			"join_trying":       "🔄 *チャンネルへの参加を試行中…*\nHash: `%s`",
			"join_trying_plain": "🔄 チャンネルへの参加を試行中…\nHash: %s",
			"join_already":      "✅ *参加済みです！* ボットはすでにこのチャンネルに参加しています。",
			"join_failed":       "❌ *参加に失敗しました*\nエラー: `%s`",
			"join_success":      "✅ *成功！* %s",

			// notify
			"notify_cmd_usage_plain":       "🔔 通知 WEBHOOK\n\n使い方: /notify <webhook_url>\n解除: /notify off\n\nタスクが完了/失敗すると、Webhook に JSON が POST されます。",
			"notify_cmd_off_plain":         "🔕 通知 Webhook を無効にしました（実行中のタスク %d 件）",
			"notify_cmd_invalid_url_plain": "❌ URL が無効です — http/https である必要があります",
			"notify_cmd_saved":             "✅ 実行中のタスク %d 件に Webhook を保存しました\n\nURL: `%s`\nタスクが完了/失敗すると、Webhook に JSON が POST されます。",

			// schedule
			"schedule_usage":        "使い方: `/schedule <type> <HH:MM> <url> [filename]`\n種類: mirror, zip, unzip, mp3, audio, video",
			"schedule_invalid_type": "無効な種類です: %s。有効な種類: mirror, zip, unzip, mp3, audio, video",
			"schedule_invalid_time": "時刻の形式が正しくありません。HH:MM（24 時間表記）で指定してください",
			"schedule_save_failed":  "タスクをスケジュールできませんでした",
			"schedule_ok":           "%s にタスクをスケジュールしました",
		},
		"zh": {
			// stats
			"stats_load_failed":         "❌ *无法获取统计数据*",
			"stats_fallback_plain":      "❌ 无法格式化统计数据（Markdown 错误）\n\n备用显示：简要统计\n总任务数：%d",
			"stats_btn_my":              "👤 我的",
			"stats_btn_today":           "📊 今天",
			"stats_btn_week":            "📈 本周",
			"stats_btn_month":           "📉 本月",
			"stats_title":               "📊 *仪表盘统计*",
			"stats_sec_global":          "🌐 *全局统计*",
			"stats_total_tasks":         "📥 总任务数: `%d`",
			"stats_completed":           "✅ 已完成: `%d`",
			"stats_failed":              "❌ 失败: `%d`",
			"stats_success_rate":        "📊 成功率: `%.1f%%`",
			"stats_bandwidth":           "📈 流量: `%s`",
			"stats_bandwidth_alt":       "📊 流量: `%s`",
			"stats_sec_today_global":    "📅 *今天（全局）*",
			"stats_tasks_short":         "📥 任务: `%d`",
			"stats_sec_today_me":        "👤 *我的今天*",
			"stats_sec_total_me":        "👤 *我的统计汇总*",
			"stats_last_active_short":   "🕐 活跃: `%s`",
			"stats_none_user":           "📭 *您还没有任何统计数据*",
			"stats_user_title":          "👤 *用户统计*",
			"stats_sec_user_info":       "🆔 *用户信息*",
			"stats_user_name":           "👤 名称: @%s",
			"stats_user_id":             "🆔 ID: `%d`",
			"stats_last_active":         "🕐 最近活跃: `%s`",
			"stats_sec_activity":        "📊 *活动*",
			"stats_succeeded":           "✅ 成功: `%d`",
			"stats_none":                "📭 *暂无统计数据*",
			"stats_daily_title":         "📅 *今天的统计*",
			"stats_sec_global_activity": "🌐 *全局活动*",
			"stats_avg_speed":           "⚡ 平均速度: `%s`",
			"stats_peak_concurrent":     "🔥 并发峰值: `%d`",
			"stats_sec_my_today":        "👤 *您今天的活动*",
			"stats_weekly_title":        "📈 *每周统计*",
			"stats_sec_7days":           "🗓️ *最近 7 天*",
			"stats_day_row":             "%s `%s` : %d 个任务",
			"stats_sec_summary":         "📊 *汇总*",
			"stats_total_line":          "📥 总计: `%d` 个任务",
			"stats_monthly_title":       "📉 *每月统计*",
			"stats_sec_by_week":         "📅 *按周*",
			"stats_week_row":            "第 %d 周: `%d` 个任务\n└ `%s`",
			"stats_err_my":              "❌ *无法获取您的统计数据*",
			"stats_err_today":           "❌ *无法获取今天的统计数据*",
			"stats_err_week":            "❌ *无法获取每周统计数据*",
			"stats_err_month":           "❌ *无法获取每月统计数据*",
			"stats_toast_refreshed":     "🔄 统计已刷新！",

			// shared basic bits
			"basic_btn_refresh":         "🔄 刷新",
			"basic_toast_closed":        "已关闭",
			"basic_pinging":             "🏓 *正在 Ping…*",
			"basic_pong":                "🏓 *Pong！* `%v`",
			"basic_speedtest_running":   "🚀 *正在运行 Speedtest…*",
			"basic_speedtest_error":     "❌ *Speedtest 错误*\n\n`%s`",
			"basic_speedtest_result":    "🚀 *Speedtest 结果*",
			"basic_name_plain":          "朋友",
			"basic_help_fallback_plain": "📖 使用指南\n\n请选择下方的类别：",

			// system
			"sys_btn_detail":       "📊 详情",
			"sys_btn_logs":         "📈 日志",
			"sys_btn_cleanup":      "🧹 清理",
			"sys_toast_refreshed":  "🔄 已刷新",
			"sys_toast_detail":     "📊 查看详情",
			"sys_toast_sending":    "正在发送文件…",
			"sys_toast_owner_only": "❌ 仅限 Owner",
			"sys_toast_cleanup":    "🧹 已开始清理",
			"sys_logs_denied_hdr":  "拒绝访问",
			"sys_logs_denied":      "只有 Admin 才能查看日志。",
			"sys_logs_missing_hdr": "未找到文件",
			"sys_logs_missing":     "未找到日志文件，或文件尚未创建。",
			"sys_logs_read_hdr":    "读取文件出错",
			"sys_logs_read_err":    "无法读取日志：%s",
			"sys_logs_write_hdr":   "写入文件出错",
			"sys_logs_write_err":   "无法创建临时文件：%s",
			"sys_logs_send_hdr":    "发送文件出错",
			"sys_logs_send_err":    "无法发送文件：%s",
			"sys_logs_caption":     "📝 *应用日志:* `%s`",

			// dashboard
			"bdash_toast_page":      "第 %d 页",
			"bdash_status_hdr":      "状态",
			"bdash_status_none":     "没有正在运行的任务。",
			"bdash_toast_empty":     "📭 空",
			"bdash_toast_updated":   "🔄 已更新",
			"bdash_toast_pong":      "🏓 Pong!",
			"bdash_toast_speedtest": "🚀 正在测速…",
			"bdash_toast_error":     "❌ 错误",
			"bdash_toast_stats":     "📊 统计",
			"bdash_storage_title":   "💾 *可用的存储提供商*",
			"bdash_storage_current": "*当前:* `%s`",
			"bdash_toast_storage":   "💾 存储",
			"bdash_toast_system":    "🖥️ 系统",
			"bdash_files_loading":   "🔍 *正在加载文件列表…*",
			"bdash_files_error":     "❌ *无法加载文件列表*\n\n错误：%s",
			"bdash_toast_files":     "📁 Drive 列表",
			"bdash_media_hdr":       "媒体处理",
			"bdash_media_body": "视频/音频的媒体操作：\n\n" +
				"• `/extractaudio` ─ 从视频提取音频\n" +
				"• `/compress` ─ 压缩视频体积\n" +
				"• `/thumbnail` ─ 生成缩略图\n" +
				"• `/screenshots` ─ 截取多张截图\n" +
				"• `/subtitle` ─ Soft\\-sub（挂载字幕）\n" +
				"• `/hardsub` ─ Hard\\-sub（硬字幕压制）\n" +
				"• `/rescale` ─ 更改视频分辨率\n" +
				"• `/convert` ─ 转换文件格式\n" +
				"• `/mediainfo` ─ 文件详细信息\n\n" +
				"💡 *用法:* 用对应命令回复视频文件",
			"bdash_toast_media":  "🎵 媒体菜单",
			"bdash_unknown_hdr":  "未知",
			"bdash_unknown_body": "未知操作。",

			"bdash_mode_mirror_hdr": "Mirror 模式",
			"bdash_mode_mirror_body": "将文件直接上传到 Google Drive。\n\n" +
				"📌 *用法*\n" +
				"• 用 `/mirror` 回复文件\n" +
				"• 或发送：`/mirror URL`\n\n" +
				"🏷️ *FLAGS*\n" +
				"• `\\-z` ─ 上传前压缩为 Zip\n" +
				"• `\\-uz` ─ 下载后解压\n" +
				"• `\\-p PASS` ─ Zip 密码",
			"bdash_mode_leech_hdr": "Leech 模式",
			"bdash_mode_leech_body": "从 URL 下载文件到服务器。\n\n" +
				"✅ *支持*\n" +
				"• HTTP/HTTPS • FTP • Magnet\n\n" +
				"📌 *用法*\n" +
				"• `/leech` \\<URL\\>",
			"bdash_mode_ytdlp_hdr": "YT-DLP 模式",
			"bdash_mode_ytdlp_body": "从 1000 多个网站下载视频。\n\n" +
				"✅ *支持*\n" +
				"• YouTube • Twitter • TikTok • 等\n\n" +
				"📌 *用法*\n" +
				"• `/ytdlp` \\<URL\\>",
			"bdash_mode_torrent_hdr": "Torrent 模式",
			"bdash_mode_torrent_body": "通过 Magnet 或 Torrent 文件下载。\n\n" +
				"📌 *用法*\n" +
				"• `/torrent` \\<magnet\\_link\\>\n" +
				"• 用 `/torrent` 回复 `.torrent` 文件",
			"bdash_mode_clone_hdr": "Clone 模式",
			"bdash_mode_clone_body": "克隆 Google Drive 文件/文件夹。\n\n" +
				"📌 *用法*\n" +
				"• `/clone` \\<GDRIVE\\_URL\\>",
			"bdash_mode_batch_hdr": "Batch 模式",
			"bdash_mode_batch_body": "一次下载多个 URL。\n\n" +
				"📌 *用法*\n" +
				"`/batch`\n" +
				"`URL1`\n" +
				"`URL2`",
			"bdash_mode_search_hdr": "搜索模式",
			"bdash_mode_search_body": "从多个来源搜索 Torrent。\n\n" +
				"📌 *用法*\n" +
				"• `/search` \\<keyword\\>",
			"bdash_mode_hardsub_hdr": "Hard-sub 模式",
			"bdash_mode_hardsub_body": "将字幕永久压制进视频。\n\n" +
				"📌 *用法*\n" +
				"• 用 `/hardsub subtitle.srt` 回复视频\n" +
				"• 或：`/hardsub video.mp4 subtitle.srt`\n\n" +
				"🔥 *说明*\n" +
				"此过程需要重新编码视频，因此比 soft\\-sub 耗时更长。",
			"bdash_mode_rescale_hdr": "Rescale 模式",
			"bdash_mode_rescale_body": "更改视频分辨率（转码）。\n\n" +
				"📌 *用法*\n" +
				"• 用 `/rescale 720p` 回复视频\n" +
				"• 或：`/rescale video.mp4 1280x720`\n\n" +
				"🏷️ *PRESET*\n" +
				"`4k`、`2k`、`1080p`、`720p`、`480p`、`360p` 或自定义格式 `WxH`。",

			// join
			"join_usage":        "⚠️ *格式错误*\n\n用法：`/join <邀请链接>`\n示例：`/join https://t.me/+FnboT0dqFiozNzA1`",
			"join_hash_error":   "❌ *错误:* 无法从链接中获取邀请哈希",
			"join_trying":       "🔄 *正在尝试加入频道…*\nHash: `%s`",
			"join_trying_plain": "🔄 正在尝试加入频道…\nHash: %s",
			"join_already":      "✅ *已加入！* 机器人已在此频道中。",
			"join_failed":       "❌ *加入失败*\n错误：`%s`",
			"join_success":      "✅ *成功！* %s",

			// notify
			"notify_cmd_usage_plain":       "🔔 通知 WEBHOOK\n\n用法：/notify <webhook_url>\n删除：/notify off\n\n当您的任务完成/失败时，Webhook 会收到 JSON POST 请求。",
			"notify_cmd_off_plain":         "🔕 已关闭通知 Webhook（%d 个运行中的任务）",
			"notify_cmd_invalid_url_plain": "❌ URL 无效 — 必须是 http/https",
			"notify_cmd_saved":             "✅ 已为 %d 个运行中的任务保存 Webhook\n\nURL：`%s`\n任务完成/失败时，Webhook 会收到 JSON POST 请求。",

			// schedule
			"schedule_usage":        "用法：`/schedule <type> <HH:MM> <url> [filename]`\n类型：mirror, zip, unzip, mp3, audio, video",
			"schedule_invalid_type": "类型无效：%s。有效类型：mirror, zip, unzip, mp3, audio, video",
			"schedule_invalid_time": "时间格式无效。请使用 HH:MM（24 小时制）",
			"schedule_save_failed":  "无法安排任务",
			"schedule_ok":           "已安排任务，时间：%s",
		},
	})
}
