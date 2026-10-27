package i18n

// Download pipeline catalog: download handlers and wizard (dl_, wizard_),
// torrent menus and metadata (tor_), batch manager (batchmgr_), clone,
// upload and archive steps (clone_, upload_, archive_) and download plugins
// (plugin_).
//
// MarkdownV2 keys are stored pre-escaped; task errors, dashboard logs,
// progress notes, toasts and button labels are plain text.
// wizard_options must keep the literal "URL:" and "Zip:" labels in every
// language: the wizard reads its state back from the message text.

const (
	dlBatchUsageBlock   = "```\n/batch\nURL1\nURL2\nURL3\n```"
	dlBatchExampleBlock = "```\n/batch -name MyDownloads -z\n" +
		"[https://example.com/file1.zip](https://example.com/file1.zip)\n" +
		"[https://example.com/file2.mp4](https://example.com/file2.mp4)\n" +
		"[https://example.com/file3.rar](https://example.com/file3.rar)\n```"
)

func init() {
	register(map[string]map[string]string{
		"vi": {
			// handlers
			"dl_toast_download_starting":  "✅ Bắt đầu tải xuống...",
			"dl_playlist_empty":           "❌ *Lỗi*\n\nPlaylist trống hoặc không tìm thấy video nào\\.",
			"dl_playlist_received":        "✅ *Đã nhận Playlist*\n\n🏷️ *Tiêu đề:* %s\n📦 *Số mục:* %d\n\nĐang xử lý các mục trong Playlist\\.\\.\\.",
			"dl_tg_file_error":            "❌ *Lỗi:* %s",
			"dl_secret_save_failed":       "❌ *Không lưu được %s*",
			"dl_secret_updated":           "✅ *Đã cập nhật %s*",
			"dl_err_local_not_found":      "Không tìm thấy file cục bộ: %v",
			"dl_err_open_source":          "Không mở được file nguồn: %v",
			"dl_err_create_dest":          "Không tạo được file đích: %v",
			"dl_err_copy":                 "Sao chép file thất bại: %v",
			"dl_err_short_write":          "Sao chép file thất bại: ghi thiếu dữ liệu",
			"dl_viking_reply_unsupported": "❌ *Lỗi*\n\nHiện chưa hỗ trợ reply file cho Viking \\(chỉ hỗ trợ URL\\)\\.",
			"dl_viking_no_url":            "❌ *Lỗi*\n\nVui lòng cung cấp URL để tải lên ViKiNG FiLE\\.",

			// mirror wizard
			"wizard_welcome":            "Chào mừng đến với Mirror Wizard\\.\nHãy chọn phương thức nhập bạn muốn\\.",
			"wizard_btn_input_url":      "🔗 Nhập URL",
			"wizard_btn_cancel":         "✖️ Hủy",
			"wizard_btn_start":          "🚀 Bắt đầu Mirror",
			"wizard_prompt_url":         "🔗 Hãy gửi *URL* bạn muốn Mirror:",
			"wizard_prompt_rename":      "✏️ Hãy gửi *Tên mới* cho file này:",
			"wizard_invalid_url":        "❌ URL không hợp lệ\\. Hãy gửi link http/https hoặc Magnet\\.",
			"wizard_options":            "⚙️ *Tùy chọn Mirror*\n\n🔗 *URL:* `%s`\n📦 *Zip:* %s\n\nHãy thiết lập tùy chọn trước khi bắt đầu\\.",
			"wizard_toast_no_url":       "❌ Không đọc được URL",
			"wizard_toast_start_failed": "❌ Lỗi khi bắt đầu Task",
			"wizard_toast_started":      "🚀 Đã bắt đầu Task",

			// torrent menu
			"tor_session_invalid":   "❌ Phiên không hợp lệ",
			"tor_session_not_found": "❌ Không tìm thấy phiên hoặc phiên đã hết hạn",
			"tor_toast_start_all":   "✅ Bắt đầu tải tất cả file...",
			"tor_started_all":       "✅ *Đã bắt đầu tải*\n\nĐang tải tất cả file trong torrent\\.\\.\\.",
			"tor_still_fetching":    "⏳ Vẫn đang lấy metadata...",
			"tor_no_files":          "❌ Không lấy được danh sách file",
			"tor_selection_updated": "🔄 Đã cập nhật lựa chọn",
			"tor_select_one":        "⚠️ Hãy chọn ít nhất một file!",
			"tor_cancelled":         "❌ Đã hủy",
			"tor_invalid_action":    "❌ Thao tác không hợp lệ",

			// torrent metadata (dashboard logs and session errors)
			"tor_meta_start":             "📡 Bắt đầu tìm metadata...",
			"tor_meta_parsing":           "📋 Đang phân tích danh sách file...",
			"tor_meta_tmpdir_failed":     "❌ Không tạo được thư mục tạm.",
			"tor_meta_local":             "📄 Đang xử lý file .torrent cục bộ...",
			"tor_meta_connecting":        "🧲 Đang kết nối tới peer (Magnet Link)...",
			"tor_meta_done":              "✅ Đã lấy được metadata.",
			"tor_meta_failed":            "❌ Không lấy được file metadata .torrent.",
			"tor_meta_missing":           "❌ Không tìm thấy file metadata trên hệ thống.",
			"tor_meta_show_files_failed": "❌ Chạy lệnh show-files thất bại.",
			"tor_meta_empty":             "⚠️ Danh sách file đọc từ metadata bị trống.",
			"tor_err_tmpdir":             "không tạo được thư mục tạm: %v",
			"tor_err_metadata":           "lấy metadata thất bại",
			"tor_err_parse":              "Không lấy được danh sách file. Metadata có thể bị hỏng hoặc không đọc được.",

			// batch manager
			"batchmgr_no_urls":  "❌ *Lỗi*\n\nKhông tìm thấy URL hợp lệ nào\\.\n\nGõ `/batch` để xem cách dùng\\.",
			"batchmgr_creating": "🔄 *Đang tạo %d task thường\\.\\.\\.*",
			"batchmgr_help": "📦 *Hệ thống tải Batch*\n\n*Cách dùng:*\n" + dlBatchUsageBlock + "\n\n" +
				"*Flag tùy chọn:*\n" +
				"• `-name <tên\\>` \\- Tên batch\n" +
				"• `-z` hoặc `-zip` \\- Nén zip toàn bộ kết quả\n" +
				"• `-p <password\\>` \\- Mật khẩu cho file zip\n" +
				"• `-priority <1\\-10\\>` \\- Độ ưu tiên \\(mặc định: 5\\)\n\n" +
				"*Ví dụ:*\n" + dlBatchExampleBlock + "\n\n" +
				"*Tính năng:*\n" +
				"✅ Tải nhiều URL cùng lúc\n" +
				"✅ Nén toàn bộ kết quả vào một file zip\n" +
				"✅ Quản lý hàng đợi theo độ ưu tiên\n" +
				"✅ Theo dõi tiến độ từng file và tổng thể",
			"batchmgr_err_dir":      "Không tạo được thư mục batch: %v",
			"batchmgr_err_zip":      "Nén zip thất bại: %v",
			"batchmgr_err_upload":   "Tải lên thất bại: %v",
			"batchmgr_err_task_dir": "không tạo được thư mục task: %v",
			"batchmgr_err_aria2":    "aria2c thất bại: %v, output: %s",
			"batchmgr_err_no_file":  "không có file nào được tải xuống",
			"batchmgr_complete": "✅ *Tải Batch hoàn tất\\!*\n\n" +
				"🏷️ *Tên:* %s\n" +
				"📊 *Tổng số file:* %d\n" +
				"✅ *Thành công:* %d\n" +
				"❌ *Thất bại:* %d\n" +
				"📦 *Tổng dung lượng:* %s\n" +
				"⏱️ *Thời gian:* %s",
			"batchmgr_link":         "\n\n🔗 *Link:* [Tải xuống](%s)",
			"batchmgr_btn_download": "📥 Tải xuống",
			"batchmgr_not_found":    "❌ Không tìm thấy batch",
			"batchmgr_refreshed":    "🔄 Đã làm mới",
			"batchmgr_cancelled":    "🚫 Đã hủy batch",
			"batchmgr_closed":       "Đã đóng",

			// clone
			"clone_no_url":      "❌ *Lỗi*\n\nVui lòng cung cấp URL Google Drive để Clone\\.",
			"clone_invalid_url": "❌ *Lỗi*\n\nURL không phải link Google Drive hợp lệ\\.",
			"clone_err_no_id":   "Không trích xuất được ID từ URL Google Drive",
			"clone_err_info":    "Không lấy được thông tin Google Drive: %v",
			"clone_err_start":   "Không khởi động được rclone: %v",
			"clone_err_failed":  "Clone bằng rclone thất bại: %v",

			// upload
			"upload_err_no_file":          "không có file để tải lên",
			"upload_err_no_local_file":    "không có file cục bộ để tải lên",
			"upload_err_stat":             "không đọc được thông tin file: %v",
			"upload_err_open":             "không mở được file: %v",
			"upload_err_viking_url":       "không lấy được URL tải lên: %v",
			"upload_err_viking_part":      "phần %d thất bại: %v",
			"upload_err_viking_complete":  "không hoàn tất được việc tải lên: %v",
			"upload_err_tg_dir":           "không thể tải thư mục trực tiếp lên Telegram, hãy nén zip trước",
			"upload_err_tg_too_large":     "file quá lớn cho Telegram (tối đa 2GB)",
			"upload_err_tg_failed":        "tải lên Telegram thất bại: %v",
			"upload_err_rclone_failed":    "rclone thất bại: %v",
			"upload_rclone_token_expired": "Token Rclone đã hết hạn! Làm mới token bằng: docker exec -it zee-mirror-bot rclone config",

			// archive
			"archive_err_extract_dir": "không tạo được thư mục giải nén: %v",
			"archive_err_extract":     "7zz giải nén thất bại: %v, output: %s",
			"archive_err_zip":         "7zz nén zip thất bại: %v, output: %s",

			// plugins
			"plugin_rclone_start":           "không khởi động được rclone: %v",
			"plugin_drive_no_id":            "không tìm thấy ID Google Drive trong URL: %s",
			"plugin_drive_failed":           "rclone tải xuống thất bại: %v",
			"plugin_mega_no_tool":           "không tìm thấy mega-get hoặc megadl trên server. Vui lòng cài MEGAcmd hoặc megatools trong Dockerfile",
			"plugin_mega_downloading":       "Đang tải file từ Mega (không hiển thị được tiến độ ở chế độ CLI)...",
			"plugin_mega_failed":            "tải xuống từ Mega thất bại: %v, output: %s",
			"plugin_ytdlp_pipe":             "không lấy được %s pipe: %v",
			"plugin_ytdlp_start":            "không khởi động được yt-dlp: %v",
			"plugin_ytdlp_hardsub_running":  "Đang ghép phụ đề cứng...",
			"plugin_ytdlp_hardsub_failed":   "Ghép phụ đề cứng thất bại: %v",
			"plugin_ytdlp_hardsub_done":     "Đã ghép phụ đề cứng xong!",
			"plugin_ytdlp_hardsub_no_files": "không tìm thấy file video hoặc phụ đề để ghép",
			"plugin_ytdlp_ffmpeg_failed":    "ffmpeg thất bại: %v, output: %s",
			"plugin_ytdlp_extract_audio":    "Đang tách/chuyển đổi audio...",
			"plugin_ytdlp_merging":          "Đang ghép video/audio...",
			"plugin_ytdlp_converting":       "Đang chuyển đổi định dạng video...",
			"plugin_ytdlp_metadata":         "Đang thêm metadata...",
			"plugin_torrent_read_local":     "không đọc được file torrent cục bộ: %v",
			"plugin_torrent_add":            "aria2 rpc thêm task thất bại: %v",
			"plugin_torrent_removed":        "task đã bị xóa khỏi aria2",
		},
		"en": {
			"dl_toast_download_starting":  "✅ Starting download...",
			"dl_playlist_empty":           "❌ *Error*\n\nThe playlist is empty or no videos were found\\.",
			"dl_playlist_received":        "✅ *Playlist Received*\n\n🏷️ *Title:* %s\n📦 *Items:* %d\n\nProcessing playlist items\\.\\.\\.",
			"dl_tg_file_error":            "❌ *Error:* %s",
			"dl_secret_save_failed":       "❌ *Failed to save %s*",
			"dl_secret_updated":           "✅ *Updated %s*",
			"dl_err_local_not_found":      "Local file not found: %v",
			"dl_err_open_source":          "Failed to open source file: %v",
			"dl_err_create_dest":          "Failed to create destination file: %v",
			"dl_err_copy":                 "File copy failed: %v",
			"dl_err_short_write":          "File copy failed: short write",
			"dl_viking_reply_unsupported": "❌ *Error*\n\nReplying to a file is not supported for Viking yet \\(URLs only\\)\\.",
			"dl_viking_no_url":            "❌ *Error*\n\nPlease provide a URL to upload to ViKiNG FiLE\\.",

			"wizard_welcome":            "Welcome to the Mirror Wizard\\.\nChoose how you want to provide input\\.",
			"wizard_btn_input_url":      "🔗 Enter URL",
			"wizard_btn_cancel":         "✖️ Cancel",
			"wizard_btn_start":          "🚀 Start Mirror",
			"wizard_prompt_url":         "🔗 Send the *URL* you want to Mirror:",
			"wizard_prompt_rename":      "✏️ Send the *new name* for this file:",
			"wizard_invalid_url":        "❌ Invalid URL\\. Please send an http/https or Magnet link\\.",
			"wizard_options":            "⚙️ *Mirror Options*\n\n🔗 *URL:* `%s`\n📦 *Zip:* %s\n\nSet your options before starting\\.",
			"wizard_toast_no_url":       "❌ Could not read the URL",
			"wizard_toast_start_failed": "❌ Failed to start Task",
			"wizard_toast_started":      "🚀 Task started",

			"tor_session_invalid":   "❌ Invalid session",
			"tor_session_not_found": "❌ Session not found or expired",
			"tor_toast_start_all":   "✅ Downloading all files...",
			"tor_started_all":       "✅ *Download started*\n\nDownloading all files in the torrent\\.\\.\\.",
			"tor_still_fetching":    "⏳ Still fetching metadata...",
			"tor_no_files":          "❌ Could not get the file list",
			"tor_selection_updated": "🔄 Selection updated",
			"tor_select_one":        "⚠️ Select at least one file!",
			"tor_cancelled":         "❌ Cancelled",
			"tor_invalid_action":    "❌ Invalid action",

			"tor_meta_start":             "📡 Looking up metadata...",
			"tor_meta_parsing":           "📋 Parsing the file list...",
			"tor_meta_tmpdir_failed":     "❌ Failed to create a temporary folder.",
			"tor_meta_local":             "📄 Processing local .torrent file...",
			"tor_meta_connecting":        "🧲 Connecting to peers (Magnet Link)...",
			"tor_meta_done":              "✅ Metadata fetched.",
			"tor_meta_failed":            "❌ Could not fetch the .torrent metadata file.",
			"tor_meta_missing":           "❌ Metadata file not found on the system.",
			"tor_meta_show_files_failed": "❌ The show-files command failed.",
			"tor_meta_empty":             "⚠️ The file list read from the metadata is empty.",
			"tor_err_tmpdir":             "failed to create temporary folder: %v",
			"tor_err_metadata":           "metadata fetch failed",
			"tor_err_parse":              "Could not get the file list. The metadata may be corrupted or unreadable.",

			"batchmgr_no_urls":  "❌ *Error*\n\nNo valid URLs found\\.\n\nType `/batch` to see usage\\.",
			"batchmgr_creating": "🔄 *Creating %d regular tasks\\.\\.\\.*",
			"batchmgr_help": "📦 *Batch Download System*\n\n*Usage:*\n" + dlBatchUsageBlock + "\n\n" +
				"*Optional flags:*\n" +
				"• `-name <name\\>` \\- Batch name\n" +
				"• `-z` or `-zip` \\- Zip all results\n" +
				"• `-p <password\\>` \\- Password for the zip file\n" +
				"• `-priority <1\\-10\\>` \\- Priority \\(default: 5\\)\n\n" +
				"*Example:*\n" + dlBatchExampleBlock + "\n\n" +
				"*Features:*\n" +
				"✅ Download many URLs at once\n" +
				"✅ Zip all results into a single file\n" +
				"✅ Queue management by priority\n" +
				"✅ Track progress per file and overall",
			"batchmgr_err_dir":      "Failed to create batch folder: %v",
			"batchmgr_err_zip":      "Zip compression failed: %v",
			"batchmgr_err_upload":   "Upload failed: %v",
			"batchmgr_err_task_dir": "failed to create task folder: %v",
			"batchmgr_err_aria2":    "aria2c failed: %v, output: %s",
			"batchmgr_err_no_file":  "no file was downloaded",
			"batchmgr_complete": "✅ *Batch Download Complete\\!*\n\n" +
				"🏷️ *Name:* %s\n" +
				"📊 *Total files:* %d\n" +
				"✅ *Succeeded:* %d\n" +
				"❌ *Failed:* %d\n" +
				"📦 *Total size:* %s\n" +
				"⏱️ *Duration:* %s",
			"batchmgr_link":         "\n\n🔗 *Link:* [Download](%s)",
			"batchmgr_btn_download": "📥 Download",
			"batchmgr_not_found":    "❌ Batch not found",
			"batchmgr_refreshed":    "🔄 Refreshed",
			"batchmgr_cancelled":    "🚫 Batch cancelled",
			"batchmgr_closed":       "Closed",

			"clone_no_url":      "❌ *Error*\n\nPlease provide a Google Drive URL to Clone\\.",
			"clone_invalid_url": "❌ *Error*\n\nThe URL is not a valid Google Drive link\\.",
			"clone_err_no_id":   "Could not extract the ID from the Google Drive URL",
			"clone_err_info":    "Failed to get Google Drive info: %v",
			"clone_err_start":   "Failed to start rclone: %v",
			"clone_err_failed":  "rclone clone failed: %v",

			"upload_err_no_file":          "no file to upload",
			"upload_err_no_local_file":    "no local file to upload",
			"upload_err_stat":             "failed to read file info: %v",
			"upload_err_open":             "failed to open file: %v",
			"upload_err_viking_url":       "failed to get upload URL: %v",
			"upload_err_viking_part":      "part %d failed: %v",
			"upload_err_viking_complete":  "failed to complete upload: %v",
			"upload_err_tg_dir":           "cannot upload a folder directly to Telegram, zip it first",
			"upload_err_tg_too_large":     "file too large for Telegram (max 2GB)",
			"upload_err_tg_failed":        "Telegram upload failed: %v",
			"upload_err_rclone_failed":    "rclone failed: %v",
			"upload_rclone_token_expired": "Rclone token expired! Refresh it with: docker exec -it zee-mirror-bot rclone config",

			"archive_err_extract_dir": "failed to create extraction folder: %v",
			"archive_err_extract":     "7zz extraction failed: %v, output: %s",
			"archive_err_zip":         "7zz zip compression failed: %v, output: %s",

			"plugin_rclone_start":           "failed to start rclone: %v",
			"plugin_drive_no_id":            "no Google Drive ID in URL: %s",
			"plugin_drive_failed":           "rclone download failed: %v",
			"plugin_mega_no_tool":           "mega-get or megadl is not installed on the server. Please install MEGAcmd or megatools in the Dockerfile",
			"plugin_mega_downloading":       "Downloading from Mega (progress is not available in CLI mode)...",
			"plugin_mega_failed":            "Mega download failed: %v, output: %s",
			"plugin_ytdlp_pipe":             "failed to get %s pipe: %v",
			"plugin_ytdlp_start":            "failed to start yt-dlp: %v",
			"plugin_ytdlp_hardsub_running":  "Burning in subtitles...",
			"plugin_ytdlp_hardsub_failed":   "Burning in subtitles failed: %v",
			"plugin_ytdlp_hardsub_done":     "Subtitles burned in!",
			"plugin_ytdlp_hardsub_no_files": "no video or subtitle file to merge",
			"plugin_ytdlp_ffmpeg_failed":    "ffmpeg failed: %v, output: %s",
			"plugin_ytdlp_extract_audio":    "Extracting/converting audio...",
			"plugin_ytdlp_merging":          "Merging video/audio...",
			"plugin_ytdlp_converting":       "Converting video format...",
			"plugin_ytdlp_metadata":         "Adding metadata...",
			"plugin_torrent_read_local":     "failed to read local torrent file: %v",
			"plugin_torrent_add":            "aria2 rpc failed to add task: %v",
			"plugin_torrent_removed":        "task was removed from aria2",
		},
		"id": {
			"dl_toast_download_starting":  "✅ Memulai download...",
			"dl_playlist_empty":           "❌ *Error*\n\nPlaylist kosong atau tidak ada video yang ditemukan\\.",
			"dl_playlist_received":        "✅ *Playlist Diterima*\n\n🏷️ *Judul:* %s\n📦 *Jumlah item:* %d\n\nMemproses item Playlist\\.\\.\\.",
			"dl_tg_file_error":            "❌ *Error:* %s",
			"dl_secret_save_failed":       "❌ *Gagal menyimpan %s*",
			"dl_secret_updated":           "✅ *%s diperbarui*",
			"dl_err_local_not_found":      "File lokal tidak ditemukan: %v",
			"dl_err_open_source":          "Gagal membuka file sumber: %v",
			"dl_err_create_dest":          "Gagal membuat file tujuan: %v",
			"dl_err_copy":                 "Gagal menyalin file: %v",
			"dl_err_short_write":          "Gagal menyalin file: data tertulis tidak lengkap",
			"dl_viking_reply_unsupported": "❌ *Error*\n\nReply file belum didukung untuk Viking \\(hanya URL\\)\\.",
			"dl_viking_no_url":            "❌ *Error*\n\nBerikan URL untuk diupload ke ViKiNG FiLE\\.",

			"wizard_welcome":            "Selamat datang di Mirror Wizard\\.\nPilih metode input yang Anda inginkan\\.",
			"wizard_btn_input_url":      "🔗 Masukkan URL",
			"wizard_btn_cancel":         "✖️ Batal",
			"wizard_btn_start":          "🚀 Mulai Mirror",
			"wizard_prompt_url":         "🔗 Silakan kirim *URL* yang ingin Anda Mirror:",
			"wizard_prompt_rename":      "✏️ Silakan kirim *Nama Baru* untuk file ini:",
			"wizard_invalid_url":        "❌ URL tidak valid\\. Kirim link http/https atau Magnet\\.",
			"wizard_options":            "⚙️ *Opsi Mirror*\n\n🔗 *URL:* `%s`\n📦 *Zip:* %s\n\nAtur opsi sebelum memulai\\.",
			"wizard_toast_no_url":       "❌ Tidak dapat membaca URL",
			"wizard_toast_start_failed": "❌ Gagal memulai Task",
			"wizard_toast_started":      "🚀 Task dimulai",

			"tor_session_invalid":   "❌ Sesi tidak valid",
			"tor_session_not_found": "❌ Sesi tidak ditemukan atau sudah kedaluwarsa",
			"tor_toast_start_all":   "✅ Mulai download semua file...",
			"tor_started_all":       "✅ *Download dimulai*\n\nMendownload semua file dalam torrent\\.\\.\\.",
			"tor_still_fetching":    "⏳ Masih mengambil metadata...",
			"tor_no_files":          "❌ Gagal mendapatkan daftar file",
			"tor_selection_updated": "🔄 Pilihan diperbarui",
			"tor_select_one":        "⚠️ Pilih setidaknya satu file!",
			"tor_cancelled":         "❌ Dibatalkan",
			"tor_invalid_action":    "❌ Aksi tidak valid",

			"tor_meta_start":             "📡 Mulai mencari metadata...",
			"tor_meta_parsing":           "📋 Menganalisa daftar file...",
			"tor_meta_tmpdir_failed":     "❌ Gagal membuat folder sementara.",
			"tor_meta_local":             "📄 Memproses file .torrent lokal...",
			"tor_meta_connecting":        "🧲 Menghubungkan ke peer (Magnet Link)...",
			"tor_meta_done":              "✅ Metadata berhasil diambil.",
			"tor_meta_failed":            "❌ Gagal mengambil file metadata .torrent.",
			"tor_meta_missing":           "❌ File metadata tidak ditemukan di sistem.",
			"tor_meta_show_files_failed": "❌ Perintah show-files gagal.",
			"tor_meta_empty":             "⚠️ Daftar file dari metadata kosong.",
			"tor_err_tmpdir":             "gagal membuat folder sementara: %v",
			"tor_err_metadata":           "gagal mengambil metadata",
			"tor_err_parse":              "Gagal mendapatkan daftar file. Metadata mungkin rusak atau tidak dapat dibaca.",

			"batchmgr_no_urls":  "❌ *Error*\n\nTidak ada URL valid yang ditemukan\\.\n\nKetik `/batch` untuk melihat cara pakai\\.",
			"batchmgr_creating": "🔄 *Membuat %d task biasa\\.\\.\\.*",
			"batchmgr_help": "📦 *Sistem Download Batch*\n\n*Cara pakai:*\n" + dlBatchUsageBlock + "\n\n" +
				"*Flag opsional:*\n" +
				"• `-name <nama\\>` \\- Nama batch\n" +
				"• `-z` atau `-zip` \\- Zip semua hasil\n" +
				"• `-p <password\\>` \\- Password untuk file zip\n" +
				"• `-priority <1\\-10\\>` \\- Prioritas \\(default: 5\\)\n\n" +
				"*Contoh:*\n" + dlBatchExampleBlock + "\n\n" +
				"*Fitur:*\n" +
				"✅ Download banyak URL sekaligus\n" +
				"✅ Zip semua hasil menjadi satu file\n" +
				"✅ Manajemen antrean berdasarkan prioritas\n" +
				"✅ Pantau progres tiap file dan keseluruhan",
			"batchmgr_err_dir":      "Gagal membuat folder batch: %v",
			"batchmgr_err_zip":      "Kompresi zip gagal: %v",
			"batchmgr_err_upload":   "Upload gagal: %v",
			"batchmgr_err_task_dir": "gagal membuat folder task: %v",
			"batchmgr_err_aria2":    "aria2c gagal: %v, output: %s",
			"batchmgr_err_no_file":  "tidak ada file yang terdownload",
			"batchmgr_complete": "✅ *Download Batch Selesai\\!*\n\n" +
				"🏷️ *Nama:* %s\n" +
				"📊 *Total file:* %d\n" +
				"✅ *Berhasil:* %d\n" +
				"❌ *Gagal:* %d\n" +
				"📦 *Total ukuran:* %s\n" +
				"⏱️ *Durasi:* %s",
			"batchmgr_link":         "\n\n🔗 *Link:* [Download](%s)",
			"batchmgr_btn_download": "📥 Download",
			"batchmgr_not_found":    "❌ Batch tidak ditemukan",
			"batchmgr_refreshed":    "🔄 Diperbarui",
			"batchmgr_cancelled":    "🚫 Batch dibatalkan",
			"batchmgr_closed":       "Ditutup",

			"clone_no_url":      "❌ *Error*\n\nBerikan URL Google Drive untuk di\\-Clone\\.",
			"clone_invalid_url": "❌ *Error*\n\nURL bukan link Google Drive yang valid\\.",
			"clone_err_no_id":   "Gagal mengambil ID dari URL Google Drive",
			"clone_err_info":    "Gagal mendapatkan info Google Drive: %v",
			"clone_err_start":   "Gagal menjalankan rclone: %v",
			"clone_err_failed":  "Clone dengan rclone gagal: %v",

			"upload_err_no_file":          "tidak ada file untuk diupload",
			"upload_err_no_local_file":    "tidak ada file lokal untuk diupload",
			"upload_err_stat":             "gagal membaca info file: %v",
			"upload_err_open":             "gagal membuka file: %v",
			"upload_err_viking_url":       "gagal mendapatkan URL upload: %v",
			"upload_err_viking_part":      "bagian %d gagal: %v",
			"upload_err_viking_complete":  "gagal menyelesaikan upload: %v",
			"upload_err_tg_dir":           "tidak dapat mengupload folder langsung ke Telegram, zip terlebih dahulu",
			"upload_err_tg_too_large":     "file terlalu besar untuk Telegram (maks 2GB)",
			"upload_err_tg_failed":        "upload ke Telegram gagal: %v",
			"upload_err_rclone_failed":    "rclone gagal: %v",
			"upload_rclone_token_expired": "Token Rclone kedaluwarsa! Perbarui token dengan: docker exec -it zee-mirror-bot rclone config",

			"archive_err_extract_dir": "gagal membuat folder ekstraksi: %v",
			"archive_err_extract":     "ekstraksi 7zz gagal: %v, output: %s",
			"archive_err_zip":         "kompresi zip 7zz gagal: %v, output: %s",

			"plugin_rclone_start":           "gagal menjalankan rclone: %v",
			"plugin_drive_no_id":            "ID Google Drive tidak ada di URL: %s",
			"plugin_drive_failed":           "download rclone gagal: %v",
			"plugin_mega_no_tool":           "mega-get atau megadl belum terpasang di server. Silakan pasang MEGAcmd atau megatools di Dockerfile",
			"plugin_mega_downloading":       "Mendownload file dari Mega (progres tidak tersedia di mode CLI)...",
			"plugin_mega_failed":            "download dari Mega gagal: %v, output: %s",
			"plugin_ytdlp_pipe":             "gagal mendapatkan pipe %s: %v",
			"plugin_ytdlp_start":            "gagal menjalankan yt-dlp: %v",
			"plugin_ytdlp_hardsub_running":  "Menempelkan subtitle permanen...",
			"plugin_ytdlp_hardsub_failed":   "Gagal menempelkan subtitle permanen: %v",
			"plugin_ytdlp_hardsub_done":     "Subtitle permanen selesai ditempel!",
			"plugin_ytdlp_hardsub_no_files": "file video atau subtitle untuk digabung tidak ada",
			"plugin_ytdlp_ffmpeg_failed":    "ffmpeg gagal: %v, output: %s",
			"plugin_ytdlp_extract_audio":    "Mengekstrak/mengonversi audio...",
			"plugin_ytdlp_merging":          "Menggabungkan video/audio...",
			"plugin_ytdlp_converting":       "Mengonversi format video...",
			"plugin_ytdlp_metadata":         "Menambahkan metadata...",
			"plugin_torrent_read_local":     "gagal membaca file torrent lokal: %v",
			"plugin_torrent_add":            "aria2 rpc gagal menambahkan task: %v",
			"plugin_torrent_removed":        "task telah dihapus dari aria2",
		},
		"ja": {
			"dl_toast_download_starting":  "✅ ダウンロードを開始します...",
			"dl_playlist_empty":           "❌ *エラー*\n\nプレイリストが空か、動画が見つかりませんでした。",
			"dl_playlist_received":        "✅ *プレイリストを受信しました*\n\n🏷️ *タイトル:* %s\n📦 *項目数:* %d\n\nプレイリストの項目を処理中…",
			"dl_tg_file_error":            "❌ *エラー:* %s",
			"dl_secret_save_failed":       "❌ *%s を保存できませんでした*",
			"dl_secret_updated":           "✅ *%s を更新しました*",
			"dl_err_local_not_found":      "ローカルファイルが見つかりません: %v",
			"dl_err_open_source":          "ソースファイルを開けませんでした: %v",
			"dl_err_create_dest":          "出力先ファイルを作成できませんでした: %v",
			"dl_err_copy":                 "ファイルのコピーに失敗しました: %v",
			"dl_err_short_write":          "ファイルのコピーに失敗しました: 書き込みが不完全です",
			"dl_viking_reply_unsupported": "❌ *エラー*\n\nViking はまだファイルへの返信に対応していません（URLのみ対応）。",
			"dl_viking_no_url":            "❌ *エラー*\n\nViKiNG FiLE にアップロードする URL を指定してください。",

			"wizard_welcome":            "Mirror Wizard へようこそ。\n入力方法を選択してください。",
			"wizard_btn_input_url":      "🔗 URLを入力",
			"wizard_btn_cancel":         "✖️ キャンセル",
			"wizard_btn_start":          "🚀 Mirror を開始",
			"wizard_prompt_url":         "🔗 Mirror したい *URL* を送信してください:",
			"wizard_prompt_rename":      "✏️ このファイルの *新しい名前* を送信してください:",
			"wizard_invalid_url":        "❌ 無効なURLです。http/https または Magnet リンクを送信してください。",
			"wizard_options":            "⚙️ *Mirror オプション*\n\n🔗 *URL:* `%s`\n📦 *Zip:* %s\n\n開始する前にオプションを設定してください。",
			"wizard_toast_no_url":       "❌ URLを読み取れませんでした",
			"wizard_toast_start_failed": "❌ Task の開始に失敗しました",
			"wizard_toast_started":      "🚀 Task を開始しました",

			"tor_session_invalid":   "❌ 無効なセッションです",
			"tor_session_not_found": "❌ セッションが見つからないか期限切れです",
			"tor_toast_start_all":   "✅ すべてのファイルのダウンロードを開始します...",
			"tor_started_all":       "✅ *ダウンロード開始*\n\nトレント内のすべてのファイルをダウンロード中…",
			"tor_still_fetching":    "⏳ メタデータを取得中です...",
			"tor_no_files":          "❌ ファイル一覧を取得できませんでした",
			"tor_selection_updated": "🔄 選択を更新しました",
			"tor_select_one":        "⚠️ 少なくとも1つのファイルを選択してください！",
			"tor_cancelled":         "❌ キャンセルしました",
			"tor_invalid_action":    "❌ 無効な操作です",

			"tor_meta_start":             "📡 メタデータの検索を開始...",
			"tor_meta_parsing":           "📋 ファイル一覧を解析中...",
			"tor_meta_tmpdir_failed":     "❌ 一時フォルダを作成できませんでした。",
			"tor_meta_local":             "📄 ローカルの .torrent ファイルを処理中...",
			"tor_meta_connecting":        "🧲 ピアに接続中 (Magnet Link)...",
			"tor_meta_done":              "✅ メタデータを取得しました。",
			"tor_meta_failed":            "❌ .torrent メタデータファイルを取得できませんでした。",
			"tor_meta_missing":           "❌ システム上にメタデータファイルがありません。",
			"tor_meta_show_files_failed": "❌ show-files コマンドが失敗しました。",
			"tor_meta_empty":             "⚠️ メタデータから読み取ったファイル一覧が空です。",
			"tor_err_tmpdir":             "一時フォルダを作成できませんでした: %v",
			"tor_err_metadata":           "メタデータの取得に失敗しました",
			"tor_err_parse":              "ファイル一覧を取得できませんでした。メタデータが破損しているか読み取れない可能性があります。",

			"batchmgr_no_urls":  "❌ *エラー*\n\n有効なURLが見つかりませんでした。\n\n`/batch` と入力すると使い方を表示します。",
			"batchmgr_creating": "🔄 *通常の Task を %d 件作成中…*",
			"batchmgr_help": "📦 *Batch ダウンロードシステム*\n\n*使い方:*\n" + dlBatchUsageBlock + "\n\n" +
				"*オプションのフラグ:*\n" +
				"• `-name <名前\\>` \\- Batch 名\n" +
				"• `-z` または `-zip` \\- すべての結果を zip に圧縮\n" +
				"• `-p <password\\>` \\- zip ファイルのパスワード\n" +
				"• `-priority <1\\-10\\>` \\- 優先度（デフォルト: 5）\n\n" +
				"*例:*\n" + dlBatchExampleBlock + "\n\n" +
				"*機能:*\n" +
				"✅ 複数の URL を同時にダウンロード\n" +
				"✅ すべての結果を1つの zip ファイルに圧縮\n" +
				"✅ 優先度によるキュー管理\n" +
				"✅ ファイルごとと全体の進捗を追跡",
			"batchmgr_err_dir":      "Batch フォルダを作成できませんでした: %v",
			"batchmgr_err_zip":      "zip 圧縮に失敗しました: %v",
			"batchmgr_err_upload":   "アップロードに失敗しました: %v",
			"batchmgr_err_task_dir": "Task フォルダを作成できませんでした: %v",
			"batchmgr_err_aria2":    "aria2c が失敗しました: %v, output: %s",
			"batchmgr_err_no_file":  "ダウンロードされたファイルがありません",
			"batchmgr_complete": "✅ *Batch ダウンロード完了！*\n\n" +
				"🏷️ *名前:* %s\n" +
				"📊 *ファイル総数:* %d\n" +
				"✅ *成功:* %d\n" +
				"❌ *失敗:* %d\n" +
				"📦 *合計サイズ:* %s\n" +
				"⏱️ *所要時間:* %s",
			"batchmgr_link":         "\n\n🔗 *リンク:* [ダウンロード](%s)",
			"batchmgr_btn_download": "📥 ダウンロード",
			"batchmgr_not_found":    "❌ Batch が見つかりません",
			"batchmgr_refreshed":    "🔄 更新しました",
			"batchmgr_cancelled":    "🚫 Batch をキャンセルしました",
			"batchmgr_closed":       "閉じました",

			"clone_no_url":      "❌ *エラー*\n\nClone する Google Drive の URL を指定してください。",
			"clone_invalid_url": "❌ *エラー*\n\nURL が有効な Google Drive リンクではありません。",
			"clone_err_no_id":   "Google Drive の URL から ID を抽出できませんでした",
			"clone_err_info":    "Google Drive の情報を取得できませんでした: %v",
			"clone_err_start":   "rclone を起動できませんでした: %v",
			"clone_err_failed":  "rclone による Clone に失敗しました: %v",

			"upload_err_no_file":          "アップロードするファイルがありません",
			"upload_err_no_local_file":    "アップロードするローカルファイルがありません",
			"upload_err_stat":             "ファイル情報を読み取れませんでした: %v",
			"upload_err_open":             "ファイルを開けませんでした: %v",
			"upload_err_viking_url":       "アップロードURLを取得できませんでした: %v",
			"upload_err_viking_part":      "パート %d が失敗しました: %v",
			"upload_err_viking_complete":  "アップロードを完了できませんでした: %v",
			"upload_err_tg_dir":           "フォルダを直接 Telegram にアップロードできません。先に zip で圧縮してください",
			"upload_err_tg_too_large":     "Telegram には大きすぎるファイルです（最大 2GB）",
			"upload_err_tg_failed":        "Telegram へのアップロードに失敗しました: %v",
			"upload_err_rclone_failed":    "rclone が失敗しました: %v",
			"upload_rclone_token_expired": "Rclone のトークンが期限切れです！次のコマンドで更新してください: docker exec -it zee-mirror-bot rclone config",

			"archive_err_extract_dir": "展開先フォルダを作成できませんでした: %v",
			"archive_err_extract":     "7zz の展開に失敗しました: %v, output: %s",
			"archive_err_zip":         "7zz の zip 圧縮に失敗しました: %v, output: %s",

			"plugin_rclone_start":           "rclone を起動できませんでした: %v",
			"plugin_drive_no_id":            "URL に Google Drive の ID がありません: %s",
			"plugin_drive_failed":           "rclone のダウンロードに失敗しました: %v",
			"plugin_mega_no_tool":           "サーバーに mega-get または megadl がインストールされていません。Dockerfile に MEGAcmd または megatools をインストールしてください",
			"plugin_mega_downloading":       "Mega からダウンロード中（CLI モードでは進捗を表示できません）...",
			"plugin_mega_failed":            "Mega からのダウンロードに失敗しました: %v, output: %s",
			"plugin_ytdlp_pipe":             "%s パイプを取得できませんでした: %v",
			"plugin_ytdlp_start":            "yt-dlp を起動できませんでした: %v",
			"plugin_ytdlp_hardsub_running":  "字幕を焼き付け中...",
			"plugin_ytdlp_hardsub_failed":   "字幕の焼き付けに失敗しました: %v",
			"plugin_ytdlp_hardsub_done":     "字幕の焼き付けが完了しました！",
			"plugin_ytdlp_hardsub_no_files": "結合する動画または字幕ファイルがありません",
			"plugin_ytdlp_ffmpeg_failed":    "ffmpeg が失敗しました: %v, output: %s",
			"plugin_ytdlp_extract_audio":    "音声を抽出/変換中...",
			"plugin_ytdlp_merging":          "動画/音声を結合中...",
			"plugin_ytdlp_converting":       "動画形式を変換中...",
			"plugin_ytdlp_metadata":         "メタデータを追加中...",
			"plugin_torrent_read_local":     "ローカルの torrent ファイルを読み取れませんでした: %v",
			"plugin_torrent_add":            "aria2 rpc で Task を追加できませんでした: %v",
			"plugin_torrent_removed":        "Task が aria2 から削除されました",
		},
		"zh": {
			"dl_toast_download_starting":  "✅ 开始下载...",
			"dl_playlist_empty":           "❌ *错误*\n\n播放列表为空或未找到任何视频。",
			"dl_playlist_received":        "✅ *已收到播放列表*\n\n🏷️ *标题:* %s\n📦 *条目数:* %d\n\n正在处理播放列表条目…",
			"dl_tg_file_error":            "❌ *错误:* %s",
			"dl_secret_save_failed":       "❌ *无法保存 %s*",
			"dl_secret_updated":           "✅ *已更新 %s*",
			"dl_err_local_not_found":      "未找到本地文件: %v",
			"dl_err_open_source":          "无法打开源文件: %v",
			"dl_err_create_dest":          "无法创建目标文件: %v",
			"dl_err_copy":                 "文件复制失败: %v",
			"dl_err_short_write":          "文件复制失败: 写入数据不完整",
			"dl_viking_reply_unsupported": "❌ *错误*\n\nViking 暂不支持回复文件（仅支持 URL）。",
			"dl_viking_no_url":            "❌ *错误*\n\n请提供要上传到 ViKiNG FiLE 的 URL。",

			"wizard_welcome":            "欢迎使用 Mirror Wizard。\n请选择您想要的输入方式。",
			"wizard_btn_input_url":      "🔗 输入 URL",
			"wizard_btn_cancel":         "✖️ 取消",
			"wizard_btn_start":          "🚀 开始 Mirror",
			"wizard_prompt_url":         "🔗 请发送您要 Mirror 的 *URL*:",
			"wizard_prompt_rename":      "✏️ 请发送此文件的 *新名称*:",
			"wizard_invalid_url":        "❌ 无效的 URL。请发送 http/https 或 Magnet 链接。",
			"wizard_options":            "⚙️ *Mirror 选项*\n\n🔗 *URL:* `%s`\n📦 *Zip:* %s\n\n请在开始前设置选项。",
			"wizard_toast_no_url":       "❌ 无法读取 URL",
			"wizard_toast_start_failed": "❌ 启动 Task 失败",
			"wizard_toast_started":      "🚀 Task 已开始",

			"tor_session_invalid":   "❌ 会话无效",
			"tor_session_not_found": "❌ 会话不存在或已过期",
			"tor_toast_start_all":   "✅ 开始下载所有文件...",
			"tor_started_all":       "✅ *下载已开始*\n\n正在下载种子中的所有文件…",
			"tor_still_fetching":    "⏳ 仍在获取元数据...",
			"tor_no_files":          "❌ 无法获取文件列表",
			"tor_selection_updated": "🔄 已更新选择",
			"tor_select_one":        "⚠️ 请至少选择一个文件！",
			"tor_cancelled":         "❌ 已取消",
			"tor_invalid_action":    "❌ 无效操作",

			"tor_meta_start":             "📡 开始查找元数据...",
			"tor_meta_parsing":           "📋 正在解析文件列表...",
			"tor_meta_tmpdir_failed":     "❌ 无法创建临时目录。",
			"tor_meta_local":             "📄 正在处理本地 .torrent 文件...",
			"tor_meta_connecting":        "🧲 正在连接节点 (Magnet Link)...",
			"tor_meta_done":              "✅ 已获取元数据。",
			"tor_meta_failed":            "❌ 无法获取 .torrent 元数据文件。",
			"tor_meta_missing":           "❌ 系统中没有元数据文件。",
			"tor_meta_show_files_failed": "❌ show-files 命令执行失败。",
			"tor_meta_empty":             "⚠️ 从元数据读取的文件列表为空。",
			"tor_err_tmpdir":             "无法创建临时目录: %v",
			"tor_err_metadata":           "获取元数据失败",
			"tor_err_parse":              "无法获取文件列表。元数据可能已损坏或无法读取。",

			"batchmgr_no_urls":  "❌ *错误*\n\n未找到任何有效的 URL。\n\n输入 `/batch` 查看用法。",
			"batchmgr_creating": "🔄 *正在创建 %d 个普通 Task…*",
			"batchmgr_help": "📦 *Batch 下载系统*\n\n*用法:*\n" + dlBatchUsageBlock + "\n\n" +
				"*可选参数:*\n" +
				"• `-name <名称\\>` \\- Batch 名称\n" +
				"• `-z` 或 `-zip` \\- 将所有结果压缩为 zip\n" +
				"• `-p <password\\>` \\- zip 文件密码\n" +
				"• `-priority <1\\-10\\>` \\- 优先级（默认: 5）\n\n" +
				"*示例:*\n" + dlBatchExampleBlock + "\n\n" +
				"*功能:*\n" +
				"✅ 同时下载多个 URL\n" +
				"✅ 将所有结果压缩为一个 zip 文件\n" +
				"✅ 按优先级管理队列\n" +
				"✅ 跟踪每个文件及整体进度",
			"batchmgr_err_dir":      "无法创建 Batch 目录: %v",
			"batchmgr_err_zip":      "zip 压缩失败: %v",
			"batchmgr_err_upload":   "上传失败: %v",
			"batchmgr_err_task_dir": "无法创建 Task 目录: %v",
			"batchmgr_err_aria2":    "aria2c 失败: %v, output: %s",
			"batchmgr_err_no_file":  "没有下载到任何文件",
			"batchmgr_complete": "✅ *Batch 下载完成！*\n\n" +
				"🏷️ *名称:* %s\n" +
				"📊 *文件总数:* %d\n" +
				"✅ *成功:* %d\n" +
				"❌ *失败:* %d\n" +
				"📦 *总大小:* %s\n" +
				"⏱️ *耗时:* %s",
			"batchmgr_link":         "\n\n🔗 *链接:* [下载](%s)",
			"batchmgr_btn_download": "📥 下载",
			"batchmgr_not_found":    "❌ 未找到 Batch",
			"batchmgr_refreshed":    "🔄 已刷新",
			"batchmgr_cancelled":    "🚫 Batch 已取消",
			"batchmgr_closed":       "已关闭",

			"clone_no_url":      "❌ *错误*\n\n请提供要 Clone 的 Google Drive URL。",
			"clone_invalid_url": "❌ *错误*\n\n该 URL 不是有效的 Google Drive 链接。",
			"clone_err_no_id":   "无法从 Google Drive URL 中提取 ID",
			"clone_err_info":    "无法获取 Google Drive 信息: %v",
			"clone_err_start":   "无法启动 rclone: %v",
			"clone_err_failed":  "使用 rclone Clone 失败: %v",

			"upload_err_no_file":          "没有可上传的文件",
			"upload_err_no_local_file":    "没有可上传的本地文件",
			"upload_err_stat":             "无法读取文件信息: %v",
			"upload_err_open":             "无法打开文件: %v",
			"upload_err_viking_url":       "无法获取上传 URL: %v",
			"upload_err_viking_part":      "第 %d 部分失败: %v",
			"upload_err_viking_complete":  "无法完成上传: %v",
			"upload_err_tg_dir":           "无法直接上传文件夹到 Telegram，请先压缩为 zip",
			"upload_err_tg_too_large":     "文件对 Telegram 来说过大（最大 2GB）",
			"upload_err_tg_failed":        "上传到 Telegram 失败: %v",
			"upload_err_rclone_failed":    "rclone 失败: %v",
			"upload_rclone_token_expired": "Rclone 令牌已过期！请使用以下命令刷新: docker exec -it zee-mirror-bot rclone config",

			"archive_err_extract_dir": "无法创建解压目录: %v",
			"archive_err_extract":     "7zz 解压失败: %v, output: %s",
			"archive_err_zip":         "7zz zip 压缩失败: %v, output: %s",

			"plugin_rclone_start":           "无法启动 rclone: %v",
			"plugin_drive_no_id":            "URL 中没有 Google Drive ID: %s",
			"plugin_drive_failed":           "rclone 下载失败: %v",
			"plugin_mega_no_tool":           "服务器上没有安装 mega-get 或 megadl。请在 Dockerfile 中安装 MEGAcmd 或 megatools",
			"plugin_mega_downloading":       "正在从 Mega 下载（CLI 模式下无法显示进度）...",
			"plugin_mega_failed":            "从 Mega 下载失败: %v, output: %s",
			"plugin_ytdlp_pipe":             "无法获取 %s 管道: %v",
			"plugin_ytdlp_start":            "无法启动 yt-dlp: %v",
			"plugin_ytdlp_hardsub_running":  "正在压制硬字幕...",
			"plugin_ytdlp_hardsub_failed":   "压制硬字幕失败: %v",
			"plugin_ytdlp_hardsub_done":     "硬字幕压制完成！",
			"plugin_ytdlp_hardsub_no_files": "没有可合并的视频或字幕文件",
			"plugin_ytdlp_ffmpeg_failed":    "ffmpeg 失败: %v, output: %s",
			"plugin_ytdlp_extract_audio":    "正在提取/转换音频...",
			"plugin_ytdlp_merging":          "正在合并视频/音频...",
			"plugin_ytdlp_converting":       "正在转换视频格式...",
			"plugin_ytdlp_metadata":         "正在添加元数据...",
			"plugin_torrent_read_local":     "无法读取本地 torrent 文件: %v",
			"plugin_torrent_add":            "aria2 rpc 添加 Task 失败: %v",
			"plugin_torrent_removed":        "Task 已从 aria2 中移除",
		},
	})
}
