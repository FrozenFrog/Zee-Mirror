package basic

import (
	"sort"
	"strings"
	"zee-mirror/internal/service"
)

type HelpCommandEntry struct {
	Name        string
	Description string
	Category    string
	Emoji       string
	Aliases     []string
}

var registeredCommands []HelpCommandEntry

func SetRegisteredCommands(commands []HelpCommandEntry) {
	registeredCommands = commands
}

func GetHelpMirror() string {
	return service.HelpDetailMessage(
		"📥 MIRROR",
		"Tải file từ URL rồi upload thẳng lên Google Drive\\.\nFile được lưu trên cloud storage, không lưu lại trên server\\.",
		"• `/mirror` hoặc `/m` <URL\\> ─ Tải và upload lên Drive\n• Reply vào file với `/mirror`\n\n⚙️ *FLAG TÙY CHỌN*\n• `\\-n <tên\\>` ─ Đổi tên file\n• `\\-z` ─ Nén zip trước khi upload\n• `\\-uz` ─ Giải nén sau khi tải xuống\n• `\\-p <pass\\>` ─ Mật khẩu zip/unzip",
		"`/mirror https://example\\.com/file\\.zip`\n`/mirror \\-n \"My File\\.zip\" https://example\\.com/file\\.zip`\n`/mirror \\-z https://example\\.com/folder`\n`/mirror \\-uz \\-p secret https://file\\.rar`",
		"",
	)
}

func GetHelpLeech() string {
	return service.HelpDetailMessage(
		"📤 LEECH",
		"Tải file từ URL về server bot\\.\nFile sẽ được gửi thẳng vào chat Telegram khi hoàn tất\\.",
		"• `/leech` hoặc `/l` <URL\\> ─ Tải về server\n• Reply vào file với `/leech`\n\n⚙️ *FLAG TÙY CHỌN*\n• `\\-n <tên\\>` ─ Đổi tên file\n• `\\-z` ─ Nén zip trước khi gửi\n• `\\-uz` ─ Giải nén sau khi tải xuống\n• `\\-p <pass\\>` ─ Mật khẩu zip/unzip",
		"`/leech https://example\\.com/video\\.mp4`\n`/leech \\-n \"My Video\\.mp4\" https://example\\.com/video\\.mp4`\n`/leech \\-uz https://archive\\.rar`",
		"",
	)
}

func GetHelpYTDLP() string {
	return service.HelpDetailMessage(
		"🎬 YT-DLP MIRROR",
		"Tải video/audio từ hơn 1000 trang web thẳng lên Drive\\.\nHỗ trợ: YouTube, TikTok, Twitter/X, Instagram, Facebook, Vimeo, v\\.v\\.",
		"• `/ytdlp` hoặc `/y` <URL\\> ─ Tải xuống, có chọn chất lượng\n\n⚙️ *FLAG TÙY CHỌN*\n• `\\-q <quality\\>` ─ Chọn chất lượng \\(best, 1080p, 720p\\)\n• `\\-n <tên\\>` ─ Đổi tên file\n• `\\-z` ─ Nén zip\n• `\\-p <pass\\>` ─ Mật khẩu zip\n\n✨ *TÍNH NĂNG*\n• Chọn chất lượng video \\(360p\\-8K\\)\n• Chỉ tải audio \\(MP3\\)\n• Phụ đề tự động",
		"`/ytdlp https://youtube\\.com/watch?v=xxxxx`\n`/ytdlp \\-q 1080p https://tiktok\\.com/@user/video/123`",
		"",
	)
}

func GetHelpYTDLPLeech() string {
	return service.HelpDetailMessage(
		"🎬 YT-DLP LEECH",
		"Tải video/audio từ hơn 1000 trang web thẳng về Telegram\\.\nHỗ trợ: YouTube, TikTok, Twitter/X, Instagram, Facebook, Vimeo, v\\.v\\.",
		"• `/ytdlpleech` hoặc `/yl` <URL\\> ─ Tải về Telegram\n\n⚙️ *FLAG TÙY CHỌN*\n• `\\-q <quality\\>` ─ Chọn chất lượng \\(best, 1080p, 720p\\)\n• `\\-n <tên\\>` ─ Đổi tên file\n• `\\-z` ─ Nén zip\n• `\\-p <pass\\>` ─ Mật khẩu zip\n\n✨ *TÍNH NĂNG*\n• Chọn chất lượng video \\(360p\\-8K\\)\n• Gửi dưới dạng video Telegram",
		"`/ytdlpleech https://youtube\\.com/watch?v=xxxxx`",
		"",
	)
}

func GetHelpViking() string {
	return service.HelpDetailMessage(
		"⚔️ VIKING FILE",
		"Tải từ URL rồi upload lên ViKiNG FiLE\\.\nHỗ trợ upload ẩn danh, hoặc upload theo tài khoản nếu đã đặt hash\\.",
		"• `/viking` hoặc `/v` <URL\\> ─ Upload lên Viking\n\n⚙️ *FLAG TÙY CHỌN*\n• `\\-n <tên\\>` ─ Đổi tên file\n• `\\-z` ─ Nén zip\n• `\\-p <pass\\>` ─ Mật khẩu zip",
		"`/viking https://example\\.com/file\\.zip`",
		"",
	)
}

func GetHelpTorrent() string {
	return service.HelpDetailMessage(
		"🧲 TORRENT",
		"Tải file qua Magnet link hoặc file torrent \\(\\.torrent\\)\\.\nBot sẽ tự động nhận diện khi bạn gửi Magnet link mà không kèm lệnh\\.",
		"• `/torrent` hoặc `/t` <magnet\\> ─ Tải từ Magnet\n• Reply vào file \\.torrent với `/torrent`\n• Gửi thẳng Magnet link \\(tự động nhận diện\\)\n\n⚙️ *FLAG TÙY CHỌN*\n• `\\-n <tên\\>` ─ Đổi tên file/thư mục\n• `\\-z` ─ Nén zip trước khi upload\n• `\\-uz` ─ Giải nén sau khi tải xuống\n• `\\-p <pass\\>` ─ Mật khẩu zip/unzip\n\n✨ *TÍNH NĂNG*\n• Hỗ trợ Magnet link và file \\.torrent\n• Tải đa kết nối\n• Tải tiếp khi bị gián đoạn",
		"`/torrent magnet:?xt=urn:btih:xxxxx`\n`/torrent \\-z magnet:?xt=urn:btih:xxxxx`",
		"💡 *TỰ ĐỘNG NHẬN DIỆN:* Bạn có thể gửi thẳng Magnet link, bot sẽ tự động xử lý\\.",
	)
}

func GetHelpClone() string {
	return service.HelpDetailMessage(
		"📋 CLONE",
		"Clone/sao chép file hoặc thư mục từ Google Drive sang Drive đích\\.",
		"• `/clone` hoặc `/cl` <drive\\_url\\> ─ Clone vào storage đang dùng\n\n🔗 *URL ĐƯỢC HỖ TRỢ*\n• `drive\\.google\\.com/file/d/xxx`\n• `drive\\.google\\.com/drive/folders/xxx`\n• `drive\\.google\\.com/open?id=xxx`",
		"`/clone https://drive\\.google\\.com/file/d/abc123`",
		"",
	)
}

func GetHelpBatch() string {
	return service.HelpDetailMessage(
		"📦 BATCH",
		"Tải nhiều URL cùng lúc trong một Batch\\.",
		"/batch\nURL1\nURL2\nURL3\n\n⚙️ *FLAG TÙY CHỌN* \\(dòng đầu tiên\\)\n• `\\-name <tên\\>` ─ Đặt tên Batch\n• `\\-z` / `\\-zip` ─ Nén zip toàn bộ kết quả\n• `\\-p <pass\\>` ─ Mật khẩu zip\n• `\\-priority <1\\-10\\>` ─ Đặt mức ưu tiên\n\n✨ *TÍNH NĂNG*\n• Tải song song\n• Theo dõi tiến độ từng URL\n• Tự động thử lại khi lỗi",
		"/batch\nhttps://file1\\.com/a\\.zip\nhttps://file2\\.com/b\\.zip\n\n`/batch \\-name DuLich \\-z`\n`https://photo\\.com/a\\.jpg`\n`https://photo\\.com/b\\.jpg`",
		"",
	)
}

func GetHelpSearch() string {
	return service.HelpDetailMessage(
		"🔍 TÌM TORRENT",
		"Tìm torrent từ nhiều nguồn/tracker\\.",
		"• `/search <từ khóa\\>` ─ Tìm torrent\n\n✨ *TÍNH NĂNG*\n• Tìm trên nhiều tracker\n• Lọc kết quả\n• Tải trực tiếp từ kết quả",
		"`/search ubuntu 22\\.04`\n`/search archlinux iso`",
		"",
	)
}

func GetHelpStatus() string {
	return service.HelpDetailMessage(
		"📊 TRẠNG THÁI",
		"Xem tất cả task đang chạy \\(tải xuống/upload\\)\\.",
		"• `/status` hoặc `/st` ─ Hiển thị tất cả task đang chạy\n\n📋 *THÔNG TIN HIỂN THỊ*\n• Tên file\n• Tiến độ \\(%\\)\n• Tốc độ tải xuống/upload\n• ETA \\(thời gian hoàn tất dự kiến\\)\n• Trạng thái \\(đang tải xuống/đang upload\\)\n\n🔘 *NÚT THAO TÁC*\n• Làm mới ─ Cập nhật trạng thái\n• Hủy ─ Hủy task",
		"`/status`",
		"",
	)
}

func GetHelpStats() string {
	return service.HelpDetailMessage(
		"📈 THỐNG KÊ",
		"Xem thống kê sử dụng bot\\.",
		"• `/stats` ─ Hiển thị thống kê\n\n📋 *THÔNG TIN HIỂN THỊ*\n• Tổng lượt tải xuống\n• Tổng lượt upload\n• Dữ liệu đã xử lý\n• Uptime của bot\n• Task hoàn tất/thất bại",
		"`/stats`",
		"",
	)
}

func GetHelpSystem() string {
	return service.HelpDetailMessage(
		"🖥️ HỆ THỐNG",
		"Xem thông tin chi tiết hệ thống server\\.",
		"• `/system` ─ Hiển thị thông tin hệ thống\n\n📋 *THÔNG TIN HIỂN THỊ*\n• Mức dùng CPU\n• Mức dùng RAM\n• Dung lượng ổ đĩa\n• Thống kê mạng\n• Uptime server\n• Thông tin hệ điều hành",
		"`/system`",
		"",
	)
}

func GetHelpHealth() string {
	return service.HelpDetailMessage(
		"🏥 KIỂM TRA HỆ THỐNG",
		"Kiểm tra tình trạng tất cả thành phần của bot\\.",
		"• `/health` ─ Kiểm tra tình trạng hệ thống\n\n🔍 *THÀNH PHẦN ĐƯỢC KIỂM TRA*\n• Aria2 daemon\n• Rclone\n• Database\n• Dung lượng ổ đĩa\n• Bộ nhớ\n• Kết nối mạng",
		"`/health`",
		"",
	)
}

func GetHelpLogs() string {
	return service.HelpDetailMessage(
		"📜 LOG",
		"Xem log hoạt động của bot\\.",
		"• `/logs` ─ Xem 50 dòng cuối\n• `/logs <n\\>` ─ Xem n dòng cuối",
		"`/logs` ─ Mặc định 50 dòng\n`/logs 100` ─ 100 dòng cuối\n`/logs 20` ─ 20 dòng cuối",
		"",
	)
}

func GetHelpPing() string {
	return service.HelpDetailMessage(
		"🏓 PING",
		"Kiểm tra độ trễ/thời gian phản hồi của bot\\.",
		"• `/ping` ─ Kiểm tra độ trễ\n\n📋 *THÔNG TIN HIỂN THỊ*\n• Thời gian phản hồi \\(ms\\)\n• Trạng thái bot",
		"`/ping`",
		"",
	)
}

func GetHelpSpeed() string {
	return service.HelpDetailMessage(
		"🚀 TỐC ĐỘ MẠNG",
		"Kiểm tra tốc độ internet của server\\.",
		"• `/speed` ─ Chạy speedtest\n\n📋 *THÔNG TIN HIỂN THỊ*\n• Tốc độ tải xuống\n• Tốc độ tải lên\n• Ping\n• Vị trí server",
		"`/speed`",
		"",
	)
}

func GetHelpLs() string {
	return service.HelpDetailMessage(
		"📂 DANH SÁCH (LS / DIR)",
		"Xem nội dung thư mục trên cloud storage\\.\nLệnh này hỗ trợ điều hướng tương tác bằng nút bấm\\.",
		"• `/ls` ─ Xem thư mục gốc\n• `/dir` ─ Alias của /ls\n• `/ls <đường dẫn\\>` ─ Xem thư mục cụ thể",
		"`/ls` ─ Thư mục gốc\n`/ls Movies` ─ Thư mục Movies\n`/ls Movies/2024` ─ Thư mục con",
		"💡 *MẸO:* Bấm vào tên thư mục trên nút để mở thư mục đó\\.",
	)
}

func GetHelpMkdir() string {
	return service.HelpDetailMessage(
		"📁 TẠO THƯ MỤC (MKDIR)",
		"Tạo thư mục mới trên cloud storage\\.",
		"• `/mkdir <tên\\_thư\\_mục\\>` ─ Tạo thư mục",
		"`/mkdir Movies`\n`/mkdir Backup/2024`",
		"",
	)
}

func GetHelpRm() string {
	return service.HelpDetailMessage(
		"🗑️ XÓA (RM)",
		"Xóa file hoặc thư mục trên cloud storage\\.",
		"• `/rm <tên\\>` ─ Xóa file/thư mục\n\n⚠️ *CẢNH BÁO*\nThao tác xóa là vĩnh viễn\\!",
		"`/rm old_file\\.zip`\n`/rm Temp/cache`",
		"",
	)
}

func GetHelpMv() string {
	return service.HelpDetailMessage(
		"📦 DI CHUYỂN (MV)",
		"Di chuyển file/thư mục sang vị trí khác hoặc đổi tên\\.",
		"• `/mv <nguồn\\> <đích\\>`",
		"`/mv file\\.zip Backup/`\n`/mv old\\.mp4 new\\.mp4`",
		"",
	)
}

func GetHelpShare() string {
	return service.HelpDetailMessage(
		"🔗 CHIA SẺ",
		"Tạo link chia sẻ cho file/thư mục\\.",
		"• `/share <tên\\>` ─ Tạo link chia sẻ",
		"`/share movie\\.mp4`\n`/share SharedFolder`",
		"",
	)
}

func GetHelpFind() string {
	return service.HelpDetailMessage(
		"🔍 TÌM FILE",
		"Tìm file đệ quy trên cloud storage\\.",
		"• `/find <từ khóa\\>` ─ Tìm file",
		"`/find movie`\n`/find \\.mp4` ─ Tìm tất cả file MP4",
		"",
	)
}

func GetHelpExtractAudio() string {
	return service.HelpDetailMessage(
		"🎵 TÁCH AUDIO",
		"Tách audio từ file video\\.",
		"• Reply vào video với `/extractaudio`\n\n📋 *ĐỊNH DẠNG ĐẦU RA*\nMP3 \\(LAME Codec\\)",
		"Reply vào video rồi gõ `/extractaudio`",
		"",
	)
}

func GetHelpCompress() string {
	return service.HelpDetailMessage(
		"🗜️ NÉN VIDEO",
		"Giảm dung lượng file video bằng FFmpeg\\.",
		"• Reply vào video với `/compress [low|medium|high]`\n\n✨ *TÍNH NĂNG*\n• Giảm bitrate video\n• Tự động điều chỉnh audio",
		"`/compress` \\(mặc định medium\\)\n`/compress high` \\(chất lượng tốt hơn\\)",
		"",
	)
}

func GetHelpThumbnail() string {
	return service.HelpDetailMessage(
		"🖼️ THUMBNAIL",
		"Tạo thumbnail từ video tại thời điểm chỉ định\\.",
		"• Reply vào video với `/thumbnail [HH:MM:SS]`\n\n✨ *TÍNH NĂNG*\n• Mặc định: giây thứ 5",
		"`/thumbnail` ─ Giây thứ 5\n`/thumbnail 00:01:30` ─ Phút 1:30",
		"",
	)
}

func GetHelpScreenshots() string {
	return service.HelpDetailMessage(
		"📸 CHỤP ẢNH MÀN HÌNH",
		"Tự động chụp nhiều ảnh màn hình từ video\\.",
		"• Reply vào video với `/screenshots [số lượng]`\n\n✨ *TÍNH NĂNG*\n• Từ 1 đến 10 ảnh",
		"`/screenshots` ─ Mặc định 4 ảnh\n`/screenshots 10` ─ Tối đa 10 ảnh",
		"",
	)
}

func GetHelpSubtitle() string {
	return service.HelpDetailMessage(
		"💬 PHỤ ĐỀ MỀM (SOFT-SUB)",
		"Nhúng track phụ đề vào container video \\(Soft\\-sub\\)\\.\nPhụ đề có thể bật/tắt khi xem\\.",
		"1\\. Gửi file phụ đề \\(\\.srt/\\.vtt\\)\n2\\. Reply vào video với `/subtitle`",
		"Reply vào video rồi gõ `/subtitle`",
		"",
	)
}

func GetHelpHardsub() string {
	return service.HelpDetailMessage(
		"🔥 PHỤ ĐỀ CỨNG (HARD-SUB)",
		"Ghi cứng phụ đề vào hình ảnh video \\(Hard\\-sub\\)\\.\nPhụ đề sẽ dính vĩnh viễn trên video và không thể tắt\\.",
		"1\\. Gửi file phụ đề \\(\\.srt/\\.ssa/\\.ass\\)\n2\\. Reply vào video với `/hardsub`",
		"Reply vào video rồi gõ `/hardsub`",
		"⚠️ *LƯU Ý:* Quá trình này mất khá lâu vì bot phải encode lại video\\.",
	)
}

func GetHelpRescale() string {
	return service.HelpDetailMessage(
		"📐 ĐỔI ĐỘ PHÂN GIẢI (RESCALE)",
		"Thay đổi độ phân giải/kích thước video \\(Transcode\\)\\.",
		"• Reply vào video với `/rescale <preset|WxH\\>`\n\n🏷️ *PRESET*\n`4k`, `2k`, `1080p`, `720p`, `480p`, `360p`",
		"Reply vào video với `/rescale 720p` hoặc `/rescale 1280x720`",
		"",
	)
}

func GetHelpConvert() string {
	return service.HelpDetailMessage(
		"🔄 CHUYỂN ĐỊNH DẠNG",
		"Chuyển đổi định dạng file video/audio mà không thay đổi chất lượng \\(Remux\\)\\.",
		"• Reply vào file với `/convert <định dạng\\>`\n\n📋 *HỖ TRỢ*\nmp4, mkv, avi, mov, webm, mp3, aac, flac, wav",
		"`/convert mp4`\n`/convert mp3`",
		"",
	)
}

func GetHelpMediaInfo() string {
	return service.HelpDetailMessage(
		"ℹ️ THÔNG TIN MEDIA",
		"Xem thông tin kỹ thuật chi tiết của file media bằng ffprobe\\.",
		"• Reply vào file với `/mediainfo`\n\n📋 *THÔNG TIN*\nCodec, độ phân giải, bitrate, thời lượng, streams",
		"Reply vào video rồi gõ `/mediainfo`",
		"",
	)
}

func GetHelpCancel() string {
	return service.HelpDetailMessage(
		"❌ HỦY TASK",
		"Hủy một task cụ thể\\.",
		"• `/cancel` hoặc `/c` <task\\_id\\> ─ Hủy task\n• Bấm nút Hủy trong tin nhắn trạng thái",
		"`/cancel abc123`",
		"💡 *MẸO:* Xem ID task bằng `/status`",
	)
}

func GetHelpCancelAll() string {
	return service.HelpDetailMessage(
		"🚫 HỦY TẤT CẢ",
		"Hủy TẤT CẢ task đang chạy\\.",
		"• `/cancelall` ─ Hủy tất cả task",
		"`/cancelall`",
		"⚠️ Mọi tiến trình tải xuống/upload đang chạy sẽ bị dừng\\!",
	)
}

func GetHelpStorages() string {
	return service.HelpDetailMessage(
		"📋 DANH SÁCH STORAGE",
		"Xem danh sách cloud storage hiện có\\.",
		"• `/storages` ─ Xem tất cả storage",
		"`/storages`",
		"",
	)
}

func GetHelpSetStorage() string {
	return service.HelpDetailMessage(
		"⚙️ CHỌN STORAGE",
		"Chọn storage đang dùng cho Mirror/Clone\\.",
		"• `/setstorage <tên\\>` ─ Chọn storage",
		"`/setstorage gdrive`",
		"",
	)
}

func GetHelpAuthorize() string {
	return service.HelpDetailMessage(
		"✅ CẤP QUYỀN",
		"Thêm người dùng mới được phép sử dụng bot\\.",
		"• `/authorize <user\\_id\\> [tên]` ─ Thêm người dùng\n• Reply vào tin nhắn của người dùng với `/authorize`",
		"`/authorize 123456789 Zee`",
		"",
	)
}

func GetHelpUnauthorize() string {
	return service.HelpDetailMessage(
		"❌ THU HỒI QUYỀN",
		"Thu hồi quyền truy cập bot của người dùng\\.",
		"• `/unauthorize <user\\_id\\>` ─ Thu hồi quyền",
		"`/unauthorize 123456789`",
		"",
	)
}

func GetHelpUsers() string {
	return service.HelpDetailMessage(
		"👥 NGƯỜI DÙNG",
		"Xem danh sách tất cả người dùng đã đăng ký trong database\\.",
		"• `/users` ─ Xem tất cả người dùng",
		"`/users`",
		"",
	)
}

func GetHelpSetAlertChannel() string {
	return service.HelpDetailMessage(
		"🚨 KÊNH CẢNH BÁO",
		"Đặt kênh nhận thông báo lỗi/sự kiện quan trọng\\.",
		"• `/setalertchannel <channel\\_id\\>`",
		"`/setalertchannel \\-1001234567890`",
		"",
	)
}

func GetHelpSetLogChannel() string {
	return service.HelpDetailMessage(
		"📜 KÊNH LOG",
		"Đặt kênh riêng để lưu log hoạt động\\.",
		"• `/setlogchannel <channel\\_id\\>`",
		"`/setlogchannel -1001234567890`",
		"💡 *LƯU Ý:* ID kênh thường bắt đầu bằng \\-100\\.",
	)
}

func GetHelpRecover() string {
	return service.HelpDetailMessage(
		"🔄 KHÔI PHỤC TASK",
		"Khôi phục các task bị dừng do bot khởi động lại\\.",
		"• `/recover` ─ Khôi phục task",
		"`/recover`",
		"",
	)
}

func GetHelpRecoveryStatus() string {
	return service.HelpDetailMessage(
		"📊 TRẠNG THÁI KHÔI PHỤC",
		"Xem thống kê các task có thể khôi phục\\.",
		"• `/recoverystatus` ─ Xem trạng thái",
		"`/recoverystatus`",
		"",
	)
}

func GetHelpSettings() string {
	return service.HelpDetailMessage(
		"⚙️ CÀI ĐẶT",
		"Tùy chỉnh tùy chọn và cấu hình của bot\\.",
		"• `/settings` ─ Mở menu cài đặt\n\n📋 *TÙY CHỌN*\n• Tự động xóa tin nhắn\n• Chế độ mặc định",
		"`/settings`",
		"",
	)
}

func getHelpAllCommands() string {
	return GenerateAllCommandsHelp(registeredCommands)
}

func escapeMarkdownV2(text string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	result := text
	for _, ch := range specialChars {
		result = strings.ReplaceAll(result, ch, "\\"+ch)
	}
	return result
}

func GenerateAllCommandsHelp(commands []HelpCommandEntry) string {
	if len(commands) == 0 {
		return service.ProfessionalMessage("TẤT CẢ LỆNH", "📋 *DANH SÁCH TẤT CẢ LỆNH CỦA BOT*\n\nChưa có lệnh nào được đăng ký\\.")
	}

	categoryOrder := []string{"download", "monitor", "files", "media", "task", "storage", "admin", "recovery", "general"}
	categoryNames := map[string]string{
		"download": "📥 TẢI XUỐNG",
		"monitor":  "📊 THEO DÕI & HỆ THỐNG",
		"files":    "📁 LƯU TRỮ & FILE",
		"media":    "🎞️ CÔNG CỤ MEDIA",
		"task":     "📋 QUẢN LÝ TASK",
		"storage":  "💾 LƯU TRỮ",
		"admin":    "👑 QUẢN TRỊ",
		"recovery": "🔧 KHÔI PHỤC",
		"general":  "🛠️ CHUNG",
	}

	byCategory := make(map[string][]HelpCommandEntry)
	for _, cmd := range commands {
		byCategory[cmd.Category] = append(byCategory[cmd.Category], cmd)
	}

	for cat := range byCategory {
		sort.Slice(byCategory[cat], func(i, j int) bool {
			return byCategory[cat][i].Name < byCategory[cat][j].Name
		})
	}

	var sb strings.Builder
	sb.WriteString("📋 *DANH SÁCH TẤT CẢ LỆNH CỦA BOT*\n")

	for _, cat := range categoryOrder {
		cmds, ok := byCategory[cat]
		if !ok || len(cmds) == 0 {
			continue
		}
		sb.WriteString("\n")
		sb.WriteString(categoryNames[cat])
		sb.WriteString("\n")
		for _, cmd := range cmds {
			sb.WriteString("• `/")
			sb.WriteString(cmd.Name)
			sb.WriteString("`")
			if len(cmd.Aliases) > 0 {
				sb.WriteString(" \\(")
				for i, alias := range cmd.Aliases {
					if i > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString("/")
					sb.WriteString(escapeMarkdownV2(alias))
				}
				sb.WriteString("\\)")
			}
			sb.WriteString(" ─ ")
			sb.WriteString(escapeMarkdownV2(cmd.Description))
			sb.WriteString("\n")
		}
	}

	return service.ProfessionalMessage("TẤT CẢ LỆNH", sb.String())
}

func GetHelpNotify() string {
	return service.HelpDetailMessage(
		"🔔 THÔNG BÁO WEBHOOK",
		"Đăng ký URL webhook để nhận thông báo tự động khi task hoàn tất hoặc thất bại\\.",
		"`/notify <webhook_url>` — Đăng ký webhook",
		"`/notify off` — Xóa webhook",
		"",
	)
}
