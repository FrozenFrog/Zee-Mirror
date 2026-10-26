package i18n

// Media tools (extract audio, compress, thumbnail, screenshots, subtitle,
// hardsub, rescale, convert, mediainfo). Key prefix: media_.
//
// Conventions in this file:
//   - media_title_*, media_info_title: plain text; the handler escapes them
//     when it wraps them in a bold MarkdownV2 header.
//   - media_btn_*, media_toast_*: plain text (button labels, callback toasts).
//   - media_err_*: plain text of errors created in the service layer; the
//     handler escapes them when it shows them.
//   - everything else is MarkdownV2, stored pre-escaped.
func init() {
	register(map[string]map[string]string{
		"vi": {
			// Headers wrapped by the handler (plain text).
			"media_title_invalid_format": "SAI ĐỊNH DẠNG",
			"media_title_no_audio":       "KHÔNG CÓ AUDIO",
			"media_title_ffmpeg_error":   "LỖI FFMPEG",
			"media_title_extract_done":   "ĐÃ TÁCH AUDIO",
			"media_title_compress_done":  "ĐÃ NÉN VIDEO",
			"media_title_subtitle_done":  "ĐÃ GẮN SUBTITLE",
			"media_title_convert_done":   "ĐÃ CHUYỂN ĐỔI MEDIA",
			"media_title_hardsub_done":   "ĐÃ GHÉP CỨNG SUBTITLE",
			"media_title_rescale_done":   "ĐÃ ĐỔI ĐỘ PHÂN GIẢI VIDEO",

			// Buttons and toasts (plain text).
			"media_btn_upload_drive":      "📤 Tải lên Drive",
			"media_toast_unknown_action":  "Thao tác media không xác định",
			"media_toast_link_expired":    "❌ Liên kết đã hết hạn hoặc không hợp lệ!",
			"media_toast_file_deleted":    "❌ File đã bị xóa khỏi server!",
			"media_toast_mirror_starting": "📤 Đang bắt đầu Mirror lên Drive...",

			// Errors from the service layer (plain text).
			"media_err_duration_fetch":   "không lấy được thời lượng video: %s",
			"media_err_duration_parse":   "không đọc được thời lượng video: %s",
			"media_err_duration_invalid": "thời lượng video không hợp lệ",
			"media_err_no_screenshots":   "Không tạo được ảnh chụp màn hình nào",

			// Shared.
			"media_usage_header":           "⚠️ *Sai định dạng*",
			"media_status_send_failed":     "❌ *Không gửi được tin nhắn trạng thái*",
			"media_download_failed":        "❌ *Tải file thất bại*",
			"media_download_failed_detail": "❌ *Tải file thất bại*\nLỗi: %s",

			// Done messages: output file, optional detail, upload hint.
			"media_done_audio":    "🎵 Output: `%s`\\!\n\nBấm nút bên dưới để tải lên Drive\\.",
			"media_done_video":    "🎬 Output: `%s`\\!\n\nBấm nút bên dưới để tải lên Drive\\.",
			"media_done_file":     "📄 Output: `%s`\\!\n\nBấm nút bên dưới để tải lên Drive\\.",
			"media_done_compress": "🎬 Output: `%s`\\!\n📊 Chất lượng: %s\n\nBấm nút bên dưới để tải lên Drive\\.",
			"media_done_rescale":  "🎬 Output: `%s`\\!\n📐 Độ phân giải: %s\n\nBấm nút bên dưới để tải lên Drive\\.",

			// Extract audio.
			"media_extract_usage":    "Hãy reply vào một file video bằng /extractaudio hoặc dùng:\n`/extractaudio path`",
			"media_extract_progress": "🎵 *Đang tách audio\\.\\.\\.*",
			"media_no_audio":         "Video này không có track audio để tách\\.",
			"media_extract_failed":   "Tách audio thất bại: %s\n\nOutput \\(phần cuối\\):\n`%s`",

			// Compress.
			"media_compress_usage":    "Hãy reply vào một file video bằng /compress hoặc dùng:\n`/compress path/to/video [quality]`\n\nChất lượng: low, medium, high \\(mặc định: medium\\)",
			"media_compress_progress": "🗜️ *Đang nén video\\.\\.\\.*",
			"media_compress_failed":   "❌ *Nén video thất bại*\n\nLỗi: %s\n\nOutput \\(phần cuối\\):\n`%s`",

			// Thumbnail.
			"media_thumbnail_usage":    "Hãy reply vào một file video bằng /thumbnail hoặc dùng:\n`/thumbnail [timestamp]`\n\nVí dụ: `/thumbnail 00:01:30`",
			"media_thumbnail_progress": "🖼️ *Đang tạo thumbnail\\.\\.\\.*",
			"media_thumbnail_failed":   "❌ *Tạo thumbnail thất bại*\n\nLỗi: %s\n\nOutput \\(phần cuối\\):\n`%s`",
			"media_thumbnail_caption":  "🖼️ Thumbnail từ `%s` tại %s",

			// Embed subtitle.
			"media_subtitle_usage":    "Dùng: `/subtitle video.mp4 subtitle.srt` hoặc reply vào video bằng `/subtitle subtitle.srt`",
			"media_subtitle_progress": "📝 *Đang gắn subtitle\\.\\.\\.*",
			"media_subtitle_failed":   "❌ *Gắn subtitle thất bại*\n\nLỗi: %s\n\nOutput \\(phần cuối\\):\n`%s`",

			// Convert.
			"media_convert_usage":       "Dùng: `/convert input.mp4 mkv` hoặc reply vào file bằng `/convert mkv`",
			"media_convert_unsupported": "❌ *Định dạng không được hỗ trợ*\n\nHỗ trợ: mp4, mkv, avi, mov, webm, mp3, aac, flac, wav",
			"media_convert_progress":    "🔄 *Đang chuyển đổi sang %s\\.\\.\\.*",
			"media_convert_failed":      "❌ *Chuyển đổi định dạng thất bại*\n\nLỗi: %s\n\nOutput \\(phần cuối\\):\n`%s`",

			// Media info (handler).
			"media_mediainfo_usage":        "Hãy reply vào một file media hoặc dùng:\n`/mediainfo path/to/file`",
			"media_mediainfo_failed":       "❌ *Không lấy được thông tin*\n\nLỗi: %s",
			"media_mediainfo_parse_failed": "❌ *Không phân tích được metadata của file*",

			// Screenshots.
			"media_screenshots_usage":    "Dùng: `/screenshots video.mp4 [count]` hoặc reply vào video bằng `/screenshots [count]`",
			"media_screenshots_progress": "📸 *Đang tạo %d ảnh chụp màn hình\\.\\.\\.*",
			"media_screenshots_failed":   "❌ *Tạo ảnh chụp màn hình thất bại*\n\nLỗi: %s",

			// Hardsub.
			"media_hardsub_usage":    "Dùng: `/hardsub video.mp4 subtitle.srt` hoặc reply vào video bằng `/hardsub subtitle.srt`",
			"media_hardsub_progress": "🔥 *Đang ghép cứng subtitle \\(Hard\\-sub\\)\\.\\.\\.*\n_Quá trình này mất nhiều thời gian hơn vì phải re\\-encode video\\._",
			"media_hardsub_failed":   "❌ *Ghép cứng subtitle thất bại*\n\nLỗi: %s\n\nOutput \\(phần cuối\\):\n`%s`",

			// Rescale.
			"media_rescale_usage_reply": "Hãy reply vào video bằng `/rescale 1280x720` hoặc `/rescale 720p`",
			"media_rescale_usage_path":  "Dùng: `/rescale video.mp4 1280x720` hoặc `/rescale video.mp4 720p`",
			"media_rescale_invalid":     "❌ *Độ phân giải không hợp lệ*\n\nDùng định dạng: `1280x720` hoặc `720p`",
			"media_rescale_progress":    "📐 *Đang đổi độ phân giải video sang %s\\.\\.\\.*",
			"media_rescale_failed":      "❌ *Đổi độ phân giải video thất bại*\n\nLỗi: %s\n\nOutput \\(phần cuối\\):\n`%s`",

			// Media info report (service.FormatMediaInfo); one key per line.
			"media_info_title":        "✨ THÔNG TIN MEDIA ✨",
			"media_info_file":         "📄 *File:* `%s`",
			"media_info_size":         "⚖️ *Dung lượng:* `%s`",
			"media_info_duration":     "🕒 *Thời lượng:* `%s`",
			"media_info_bitrate":      "📊 *Bitrate tổng:* `%s/s`",
			"media_info_format":       "📦 *Định dạng:* `%s`",
			"media_info_video_stream": "🎬 *LUỒNG VIDEO \\#%d*",
			"media_info_codec":        " • *Codec:* `%s`",
			"media_info_resolution":   " • *Độ phân giải:* `%dx%d`",
			"media_info_aspect":       " • *Tỉ lệ khung hình:* `%s`",
			"media_info_pixfmt":       " • *Định dạng pixel:* `%s`",
			"media_info_fps":          " • *Tốc độ khung hình:* `%.2f fps`",
			"media_info_audio_stream": "🎵 *LUỒNG AUDIO \\#%d*",
			"media_info_channels":     " • *Kênh:* `%d` \\(%s\\)",
			"media_info_sample_rate":  " • *Tần số lấy mẫu:* `%.1f kHz`",
			"media_info_audio_rate":   " • *Bitrate:* `%s/s`",
		},
		"en": {
			"media_title_invalid_format": "INVALID FORMAT",
			"media_title_no_audio":       "NO AUDIO",
			"media_title_ffmpeg_error":   "FFMPEG ERROR",
			"media_title_extract_done":   "AUDIO EXTRACTED",
			"media_title_compress_done":  "VIDEO COMPRESSED",
			"media_title_subtitle_done":  "SUBTITLE ADDED",
			"media_title_convert_done":   "MEDIA CONVERTED",
			"media_title_hardsub_done":   "SUBTITLE BURNED IN",
			"media_title_rescale_done":   "VIDEO RESCALED",

			"media_btn_upload_drive":      "📤 Upload to Drive",
			"media_toast_unknown_action":  "Unknown media action",
			"media_toast_link_expired":    "❌ Link expired or invalid!",
			"media_toast_file_deleted":    "❌ File was deleted from the server!",
			"media_toast_mirror_starting": "📤 Starting Mirror to Drive...",

			"media_err_duration_fetch":   "failed to get video duration: %s",
			"media_err_duration_parse":   "failed to read video duration: %s",
			"media_err_duration_invalid": "invalid video duration",
			"media_err_no_screenshots":   "No screenshots could be generated",

			"media_usage_header":           "⚠️ *Invalid format*",
			"media_status_send_failed":     "❌ *Could not send the status message*",
			"media_download_failed":        "❌ *File download failed*",
			"media_download_failed_detail": "❌ *File download failed*\nError: %s",

			"media_done_audio":    "🎵 Output: `%s`\\!\n\nTap the button below to upload to Drive\\.",
			"media_done_video":    "🎬 Output: `%s`\\!\n\nTap the button below to upload to Drive\\.",
			"media_done_file":     "📄 Output: `%s`\\!\n\nTap the button below to upload to Drive\\.",
			"media_done_compress": "🎬 Output: `%s`\\!\n📊 Quality: %s\n\nTap the button below to upload to Drive\\.",
			"media_done_rescale":  "🎬 Output: `%s`\\!\n📐 Resolution: %s\n\nTap the button below to upload to Drive\\.",

			"media_extract_usage":    "Reply to a video file with /extractaudio or use:\n`/extractaudio path`",
			"media_extract_progress": "🎵 *Extracting audio\\.\\.\\.*",
			"media_no_audio":         "This video has no audio track to extract\\.",
			"media_extract_failed":   "Audio extraction failed: %s\n\nOutput \\(last lines\\):\n`%s`",

			"media_compress_usage":    "Reply to a video file with /compress or use:\n`/compress path/to/video [quality]`\n\nQuality: low, medium, high \\(default: medium\\)",
			"media_compress_progress": "🗜️ *Compressing video\\.\\.\\.*",
			"media_compress_failed":   "❌ *Video compression failed*\n\nError: %s\n\nOutput \\(last lines\\):\n`%s`",

			"media_thumbnail_usage":    "Reply to a video file with /thumbnail or use:\n`/thumbnail [timestamp]`\n\nExample: `/thumbnail 00:01:30`",
			"media_thumbnail_progress": "🖼️ *Creating thumbnail\\.\\.\\.*",
			"media_thumbnail_failed":   "❌ *Thumbnail creation failed*\n\nError: %s\n\nOutput \\(last lines\\):\n`%s`",
			"media_thumbnail_caption":  "🖼️ Thumbnail from `%s` at %s",

			"media_subtitle_usage":    "Usage: `/subtitle video.mp4 subtitle.srt` or reply to a video with `/subtitle subtitle.srt`",
			"media_subtitle_progress": "📝 *Adding subtitle\\.\\.\\.*",
			"media_subtitle_failed":   "❌ *Adding subtitle failed*\n\nError: %s\n\nOutput \\(last lines\\):\n`%s`",

			"media_convert_usage":       "Usage: `/convert input.mp4 mkv` or reply to a file with `/convert mkv`",
			"media_convert_unsupported": "❌ *Unsupported format*\n\nSupported: mp4, mkv, avi, mov, webm, mp3, aac, flac, wav",
			"media_convert_progress":    "🔄 *Converting to %s\\.\\.\\.*",
			"media_convert_failed":      "❌ *Format conversion failed*\n\nError: %s\n\nOutput \\(last lines\\):\n`%s`",

			"media_mediainfo_usage":        "Reply to a media file or use:\n`/mediainfo path/to/file`",
			"media_mediainfo_failed":       "❌ *Could not get media info*\n\nError: %s",
			"media_mediainfo_parse_failed": "❌ *Could not parse the file metadata*",

			"media_screenshots_usage":    "Usage: `/screenshots video.mp4 [count]` or reply to a video with `/screenshots [count]`",
			"media_screenshots_progress": "📸 *Generating %d screenshot\\(s\\)\\.\\.\\.*",
			"media_screenshots_failed":   "❌ *Screenshot generation failed*\n\nError: %s",

			"media_hardsub_usage":    "Usage: `/hardsub video.mp4 subtitle.srt` or reply to a video with `/hardsub subtitle.srt`",
			"media_hardsub_progress": "🔥 *Burning in subtitle \\(Hard\\-sub\\)\\.\\.\\.*\n_This takes longer because the video has to be re\\-encoded\\._",
			"media_hardsub_failed":   "❌ *Burning in subtitle failed*\n\nError: %s\n\nOutput \\(last lines\\):\n`%s`",

			"media_rescale_usage_reply": "Reply to a video with `/rescale 1280x720` or `/rescale 720p`",
			"media_rescale_usage_path":  "Usage: `/rescale video.mp4 1280x720` or `/rescale video.mp4 720p`",
			"media_rescale_invalid":     "❌ *Invalid resolution*\n\nUse the format: `1280x720` or `720p`",
			"media_rescale_progress":    "📐 *Rescaling video to %s\\.\\.\\.*",
			"media_rescale_failed":      "❌ *Video rescaling failed*\n\nError: %s\n\nOutput \\(last lines\\):\n`%s`",

			"media_info_title":        "✨ MEDIA INFO ✨",
			"media_info_file":         "📄 *File:* `%s`",
			"media_info_size":         "⚖️ *Size:* `%s`",
			"media_info_duration":     "🕒 *Duration:* `%s`",
			"media_info_bitrate":      "📊 *Total bitrate:* `%s/s`",
			"media_info_format":       "📦 *Format:* `%s`",
			"media_info_video_stream": "🎬 *VIDEO STREAM \\#%d*",
			"media_info_codec":        " • *Codec:* `%s`",
			"media_info_resolution":   " • *Resolution:* `%dx%d`",
			"media_info_aspect":       " • *Aspect ratio:* `%s`",
			"media_info_pixfmt":       " • *Pixel format:* `%s`",
			"media_info_fps":          " • *Frame rate:* `%.2f fps`",
			"media_info_audio_stream": "🎵 *AUDIO STREAM \\#%d*",
			"media_info_channels":     " • *Channels:* `%d` \\(%s\\)",
			"media_info_sample_rate":  " • *Sample rate:* `%.1f kHz`",
			"media_info_audio_rate":   " • *Bitrate:* `%s/s`",
		},
		"id": {
			"media_title_invalid_format": "FORMAT SALAH",
			"media_title_no_audio":       "TIDAK ADA AUDIO",
			"media_title_ffmpeg_error":   "KESALAHAN FFMPEG",
			"media_title_extract_done":   "AUDIO DIEKSTRAK",
			"media_title_compress_done":  "VIDEO DIKOMPRESI",
			"media_title_subtitle_done":  "SUBTITLE DITAMBAHKAN",
			"media_title_convert_done":   "MEDIA DIKONVERSI",
			"media_title_hardsub_done":   "SUBTITLE DITANAMKAN",
			"media_title_rescale_done":   "RESOLUSI VIDEO DIUBAH",

			"media_btn_upload_drive":      "📤 Unggah ke Drive",
			"media_toast_unknown_action":  "Aksi media tidak dikenal",
			"media_toast_link_expired":    "❌ Tautan kedaluwarsa atau tidak valid!",
			"media_toast_file_deleted":    "❌ File sudah dihapus dari server!",
			"media_toast_mirror_starting": "📤 Memulai Mirror ke Drive...",

			"media_err_duration_fetch":   "gagal mendapatkan durasi video: %s",
			"media_err_duration_parse":   "gagal membaca durasi video: %s",
			"media_err_duration_invalid": "durasi video tidak valid",
			"media_err_no_screenshots":   "Tidak ada screenshot yang berhasil dibuat",

			"media_usage_header":           "⚠️ *Format salah*",
			"media_status_send_failed":     "❌ *Gagal mengirim pesan status*",
			"media_download_failed":        "❌ *Gagal mengunduh file*",
			"media_download_failed_detail": "❌ *Gagal mengunduh file*\nKesalahan: %s",

			"media_done_audio":    "🎵 Output: `%s`\\!\n\nTekan tombol di bawah untuk mengunggah ke Drive\\.",
			"media_done_video":    "🎬 Output: `%s`\\!\n\nTekan tombol di bawah untuk mengunggah ke Drive\\.",
			"media_done_file":     "📄 Output: `%s`\\!\n\nTekan tombol di bawah untuk mengunggah ke Drive\\.",
			"media_done_compress": "🎬 Output: `%s`\\!\n📊 Kualitas: %s\n\nTekan tombol di bawah untuk mengunggah ke Drive\\.",
			"media_done_rescale":  "🎬 Output: `%s`\\!\n📐 Resolusi: %s\n\nTekan tombol di bawah untuk mengunggah ke Drive\\.",

			"media_extract_usage":    "Balas file video dengan /extractaudio atau gunakan:\n`/extractaudio path`",
			"media_extract_progress": "🎵 *Mengekstrak audio\\.\\.\\.*",
			"media_no_audio":         "Video ini tidak memiliki trek audio untuk diekstrak\\.",
			"media_extract_failed":   "Ekstrak audio gagal: %s\n\nOutput \\(bagian akhir\\):\n`%s`",

			"media_compress_usage":    "Balas file video dengan /compress atau gunakan:\n`/compress path/to/video [quality]`\n\nKualitas: low, medium, high \\(default: medium\\)",
			"media_compress_progress": "🗜️ *Mengompres video\\.\\.\\.*",
			"media_compress_failed":   "❌ *Kompresi video gagal*\n\nKesalahan: %s\n\nOutput \\(bagian akhir\\):\n`%s`",

			"media_thumbnail_usage":    "Balas file video dengan /thumbnail atau gunakan:\n`/thumbnail [timestamp]`\n\nContoh: `/thumbnail 00:01:30`",
			"media_thumbnail_progress": "🖼️ *Membuat thumbnail\\.\\.\\.*",
			"media_thumbnail_failed":   "❌ *Pembuatan thumbnail gagal*\n\nKesalahan: %s\n\nOutput \\(bagian akhir\\):\n`%s`",
			"media_thumbnail_caption":  "🖼️ Thumbnail dari `%s` pada %s",

			"media_subtitle_usage":    "Gunakan: `/subtitle video.mp4 subtitle.srt` atau balas video dengan `/subtitle subtitle.srt`",
			"media_subtitle_progress": "📝 *Menambahkan subtitle\\.\\.\\.*",
			"media_subtitle_failed":   "❌ *Penambahan subtitle gagal*\n\nKesalahan: %s\n\nOutput \\(bagian akhir\\):\n`%s`",

			"media_convert_usage":       "Gunakan: `/convert input.mp4 mkv` atau balas file dengan `/convert mkv`",
			"media_convert_unsupported": "❌ *Format tidak didukung*\n\nDidukung: mp4, mkv, avi, mov, webm, mp3, aac, flac, wav",
			"media_convert_progress":    "🔄 *Mengonversi ke %s\\.\\.\\.*",
			"media_convert_failed":      "❌ *Konversi format gagal*\n\nKesalahan: %s\n\nOutput \\(bagian akhir\\):\n`%s`",

			"media_mediainfo_usage":        "Balas file media atau gunakan:\n`/mediainfo path/to/file`",
			"media_mediainfo_failed":       "❌ *Gagal mendapatkan informasi*\n\nKesalahan: %s",
			"media_mediainfo_parse_failed": "❌ *Gagal membaca metadata file*",

			"media_screenshots_usage":    "Gunakan: `/screenshots video.mp4 [count]` atau balas video dengan `/screenshots [count]`",
			"media_screenshots_progress": "📸 *Membuat %d screenshot\\.\\.\\.*",
			"media_screenshots_failed":   "❌ *Pembuatan screenshot gagal*\n\nKesalahan: %s",

			"media_hardsub_usage":    "Gunakan: `/hardsub video.mp4 subtitle.srt` atau balas video dengan `/hardsub subtitle.srt`",
			"media_hardsub_progress": "🔥 *Menanamkan subtitle secara permanen \\(Hard\\-sub\\)\\.\\.\\.*\n_Proses ini lebih lama karena video harus di\\-encode ulang\\._",
			"media_hardsub_failed":   "❌ *Penanaman subtitle gagal*\n\nKesalahan: %s\n\nOutput \\(bagian akhir\\):\n`%s`",

			"media_rescale_usage_reply": "Balas video dengan `/rescale 1280x720` atau `/rescale 720p`",
			"media_rescale_usage_path":  "Gunakan: `/rescale video.mp4 1280x720` atau `/rescale video.mp4 720p`",
			"media_rescale_invalid":     "❌ *Resolusi tidak valid*\n\nGunakan format: `1280x720` atau `720p`",
			"media_rescale_progress":    "📐 *Mengubah resolusi video ke %s\\.\\.\\.*",
			"media_rescale_failed":      "❌ *Perubahan resolusi video gagal*\n\nKesalahan: %s\n\nOutput \\(bagian akhir\\):\n`%s`",

			"media_info_title":        "✨ INFO MEDIA ✨",
			"media_info_file":         "📄 *File:* `%s`",
			"media_info_size":         "⚖️ *Ukuran:* `%s`",
			"media_info_duration":     "🕒 *Durasi:* `%s`",
			"media_info_bitrate":      "📊 *Total bitrate:* `%s/s`",
			"media_info_format":       "📦 *Format:* `%s`",
			"media_info_video_stream": "🎬 *STREAM VIDEO \\#%d*",
			"media_info_codec":        " • *Codec:* `%s`",
			"media_info_resolution":   " • *Resolusi:* `%dx%d`",
			"media_info_aspect":       " • *Rasio aspek:* `%s`",
			"media_info_pixfmt":       " • *Format piksel:* `%s`",
			"media_info_fps":          " • *Frame rate:* `%.2f fps`",
			"media_info_audio_stream": "🎵 *STREAM AUDIO \\#%d*",
			"media_info_channels":     " • *Kanal:* `%d` \\(%s\\)",
			"media_info_sample_rate":  " • *Sample rate:* `%.1f kHz`",
			"media_info_audio_rate":   " • *Bitrate:* `%s/s`",
		},
		"ja": {
			"media_title_invalid_format": "形式エラー",
			"media_title_no_audio":       "音声なし",
			"media_title_ffmpeg_error":   "FFMPEGエラー",
			"media_title_extract_done":   "音声を抽出しました",
			"media_title_compress_done":  "動画を圧縮しました",
			"media_title_subtitle_done":  "字幕を追加しました",
			"media_title_convert_done":   "メディアを変換しました",
			"media_title_hardsub_done":   "字幕を焼き込みました",
			"media_title_rescale_done":   "動画の解像度を変更しました",

			"media_btn_upload_drive":      "📤 Driveにアップロード",
			"media_toast_unknown_action":  "不明なメディア操作です",
			"media_toast_link_expired":    "❌ リンクの有効期限が切れているか、無効です！",
			"media_toast_file_deleted":    "❌ ファイルはサーバーから削除されました！",
			"media_toast_mirror_starting": "📤 DriveへのMirrorを開始しています...",

			"media_err_duration_fetch":   "動画の長さを取得できませんでした: %s",
			"media_err_duration_parse":   "動画の長さを読み取れませんでした: %s",
			"media_err_duration_invalid": "動画の長さが無効です",
			"media_err_no_screenshots":   "スクリーンショットを1枚も作成できませんでした",

			"media_usage_header":           "⚠️ *形式が正しくありません*",
			"media_status_send_failed":     "❌ *ステータスメッセージを送信できませんでした*",
			"media_download_failed":        "❌ *ファイルのダウンロードに失敗しました*",
			"media_download_failed_detail": "❌ *ファイルのダウンロードに失敗しました*\nエラー: %s",

			"media_done_audio":    "🎵 Output: `%s`！\n\n下のボタンを押すとDriveにアップロードします。",
			"media_done_video":    "🎬 Output: `%s`！\n\n下のボタンを押すとDriveにアップロードします。",
			"media_done_file":     "📄 Output: `%s`！\n\n下のボタンを押すとDriveにアップロードします。",
			"media_done_compress": "🎬 Output: `%s`！\n📊 品質: %s\n\n下のボタンを押すとDriveにアップロードします。",
			"media_done_rescale":  "🎬 Output: `%s`！\n📐 解像度: %s\n\n下のボタンを押すとDriveにアップロードします。",

			"media_extract_usage":    "動画ファイルに /extractaudio で返信するか、次のように使用してください：\n`/extractaudio path`",
			"media_extract_progress": "🎵 *音声を抽出中\\.\\.\\.*",
			"media_no_audio":         "この動画には抽出できる音声トラックがありません。",
			"media_extract_failed":   "音声の抽出に失敗しました: %s\n\nOutput（末尾）：\n`%s`",

			"media_compress_usage":    "動画ファイルに /compress で返信するか、次のように使用してください：\n`/compress path/to/video [quality]`\n\n品質: low, medium, high（デフォルト: medium）",
			"media_compress_progress": "🗜️ *動画を圧縮中\\.\\.\\.*",
			"media_compress_failed":   "❌ *動画の圧縮に失敗しました*\n\nエラー: %s\n\nOutput（末尾）：\n`%s`",

			"media_thumbnail_usage":    "動画ファイルに /thumbnail で返信するか、次のように使用してください：\n`/thumbnail [timestamp]`\n\n例: `/thumbnail 00:01:30`",
			"media_thumbnail_progress": "🖼️ *サムネイルを作成中\\.\\.\\.*",
			"media_thumbnail_failed":   "❌ *サムネイルの作成に失敗しました*\n\nエラー: %s\n\nOutput（末尾）：\n`%s`",
			"media_thumbnail_caption":  "🖼️ `%s` の %s 時点のサムネイル",

			"media_subtitle_usage":    "使い方: `/subtitle video.mp4 subtitle.srt`、または動画に `/subtitle subtitle.srt` で返信してください",
			"media_subtitle_progress": "📝 *字幕を追加中\\.\\.\\.*",
			"media_subtitle_failed":   "❌ *字幕の追加に失敗しました*\n\nエラー: %s\n\nOutput（末尾）：\n`%s`",

			"media_convert_usage":       "使い方: `/convert input.mp4 mkv`、またはファイルに `/convert mkv` で返信してください",
			"media_convert_unsupported": "❌ *サポートされていない形式です*\n\n対応形式: mp4, mkv, avi, mov, webm, mp3, aac, flac, wav",
			"media_convert_progress":    "🔄 *%s に変換中\\.\\.\\.*",
			"media_convert_failed":      "❌ *形式の変換に失敗しました*\n\nエラー: %s\n\nOutput（末尾）：\n`%s`",

			"media_mediainfo_usage":        "メディアファイルに返信するか、次のように使用してください：\n`/mediainfo path/to/file`",
			"media_mediainfo_failed":       "❌ *情報を取得できませんでした*\n\nエラー: %s",
			"media_mediainfo_parse_failed": "❌ *ファイルのメタデータを解析できませんでした*",

			"media_screenshots_usage":    "使い方: `/screenshots video.mp4 [count]`、または動画に `/screenshots [count]` で返信してください",
			"media_screenshots_progress": "📸 *スクリーンショットを %d 枚作成中\\.\\.\\.*",
			"media_screenshots_failed":   "❌ *スクリーンショットの作成に失敗しました*\n\nエラー: %s",

			"media_hardsub_usage":    "使い方: `/hardsub video.mp4 subtitle.srt`、または動画に `/hardsub subtitle.srt` で返信してください",
			"media_hardsub_progress": "🔥 *字幕を焼き込み中 \\(Hard\\-sub\\)\\.\\.\\.*\n_動画を再エンコードするため、通常より時間がかかります。_",
			"media_hardsub_failed":   "❌ *字幕の焼き込みに失敗しました*\n\nエラー: %s\n\nOutput（末尾）：\n`%s`",

			"media_rescale_usage_reply": "動画に `/rescale 1280x720` または `/rescale 720p` で返信してください",
			"media_rescale_usage_path":  "使い方: `/rescale video.mp4 1280x720` または `/rescale video.mp4 720p`",
			"media_rescale_invalid":     "❌ *解像度が無効です*\n\n次の形式で指定してください: `1280x720` または `720p`",
			"media_rescale_progress":    "📐 *動画の解像度を %s に変更中\\.\\.\\.*",
			"media_rescale_failed":      "❌ *動画の解像度変更に失敗しました*\n\nエラー: %s\n\nOutput（末尾）：\n`%s`",

			"media_info_title":        "✨ メディア情報 ✨",
			"media_info_file":         "📄 *ファイル:* `%s`",
			"media_info_size":         "⚖️ *サイズ:* `%s`",
			"media_info_duration":     "🕒 *再生時間:* `%s`",
			"media_info_bitrate":      "📊 *総ビットレート:* `%s/s`",
			"media_info_format":       "📦 *フォーマット:* `%s`",
			"media_info_video_stream": "🎬 *ビデオストリーム \\#%d*",
			"media_info_codec":        " • *コーデック:* `%s`",
			"media_info_resolution":   " • *解像度:* `%dx%d`",
			"media_info_aspect":       " • *アスペクト比:* `%s`",
			"media_info_pixfmt":       " • *ピクセルフォーマット:* `%s`",
			"media_info_fps":          " • *フレームレート:* `%.2f fps`",
			"media_info_audio_stream": "🎵 *オーディオストリーム \\#%d*",
			"media_info_channels":     " • *チャンネル:* `%d` （%s）",
			"media_info_sample_rate":  " • *サンプルレート:* `%.1f kHz`",
			"media_info_audio_rate":   " • *ビットレート:* `%s/s`",
		},
		"zh": {
			"media_title_invalid_format": "格式错误",
			"media_title_no_audio":       "没有音频",
			"media_title_ffmpeg_error":   "FFMPEG 错误",
			"media_title_extract_done":   "音频已提取",
			"media_title_compress_done":  "视频已压缩",
			"media_title_subtitle_done":  "字幕已添加",
			"media_title_convert_done":   "媒体已转换",
			"media_title_hardsub_done":   "硬字幕已合成",
			"media_title_rescale_done":   "视频分辨率已调整",

			"media_btn_upload_drive":      "📤 上传到 Drive",
			"media_toast_unknown_action":  "未知的媒体操作",
			"media_toast_link_expired":    "❌ 链接已过期或无效！",
			"media_toast_file_deleted":    "❌ 文件已从服务器删除！",
			"media_toast_mirror_starting": "📤 正在开始 Mirror 到 Drive...",

			"media_err_duration_fetch":   "无法获取视频时长: %s",
			"media_err_duration_parse":   "无法读取视频时长: %s",
			"media_err_duration_invalid": "视频时长无效",
			"media_err_no_screenshots":   "未能生成任何截图",

			"media_usage_header":           "⚠️ *格式错误*",
			"media_status_send_failed":     "❌ *无法发送状态消息*",
			"media_download_failed":        "❌ *文件下载失败*",
			"media_download_failed_detail": "❌ *文件下载失败*\n错误：%s",

			"media_done_audio":    "🎵 Output: `%s`！\n\n点击下方按钮上传到 Drive。",
			"media_done_video":    "🎬 Output: `%s`！\n\n点击下方按钮上传到 Drive。",
			"media_done_file":     "📄 Output: `%s`！\n\n点击下方按钮上传到 Drive。",
			"media_done_compress": "🎬 Output: `%s`！\n📊 质量：%s\n\n点击下方按钮上传到 Drive。",
			"media_done_rescale":  "🎬 Output: `%s`！\n📐 分辨率：%s\n\n点击下方按钮上传到 Drive。",

			"media_extract_usage":    "请回复视频文件并使用 /extractaudio，或使用：\n`/extractaudio path`",
			"media_extract_progress": "🎵 *正在提取音频\\.\\.\\.*",
			"media_no_audio":         "此视频没有可提取的音轨。",
			"media_extract_failed":   "音频提取失败：%s\n\nOutput（末尾）：\n`%s`",

			"media_compress_usage":    "请回复视频文件并使用 /compress，或使用：\n`/compress path/to/video [quality]`\n\n质量：low、medium、high（默认：medium）",
			"media_compress_progress": "🗜️ *正在压缩视频\\.\\.\\.*",
			"media_compress_failed":   "❌ *视频压缩失败*\n\n错误：%s\n\nOutput（末尾）：\n`%s`",

			"media_thumbnail_usage":    "请回复视频文件并使用 /thumbnail，或使用：\n`/thumbnail [timestamp]`\n\n示例：`/thumbnail 00:01:30`",
			"media_thumbnail_progress": "🖼️ *正在生成缩略图\\.\\.\\.*",
			"media_thumbnail_failed":   "❌ *缩略图生成失败*\n\n错误：%s\n\nOutput（末尾）：\n`%s`",
			"media_thumbnail_caption":  "🖼️ `%s` 在 %s 处的缩略图",

			"media_subtitle_usage":    "用法：`/subtitle video.mp4 subtitle.srt`，或回复视频并使用 `/subtitle subtitle.srt`",
			"media_subtitle_progress": "📝 *正在添加字幕\\.\\.\\.*",
			"media_subtitle_failed":   "❌ *字幕添加失败*\n\n错误：%s\n\nOutput（末尾）：\n`%s`",

			"media_convert_usage":       "用法：`/convert input.mp4 mkv`，或回复文件并使用 `/convert mkv`",
			"media_convert_unsupported": "❌ *不支持的格式*\n\n支持：mp4, mkv, avi, mov, webm, mp3, aac, flac, wav",
			"media_convert_progress":    "🔄 *正在转换为 %s\\.\\.\\.*",
			"media_convert_failed":      "❌ *格式转换失败*\n\n错误：%s\n\nOutput（末尾）：\n`%s`",

			"media_mediainfo_usage":        "请回复媒体文件，或使用：\n`/mediainfo path/to/file`",
			"media_mediainfo_failed":       "❌ *无法获取信息*\n\n错误：%s",
			"media_mediainfo_parse_failed": "❌ *无法解析文件元数据*",

			"media_screenshots_usage":    "用法：`/screenshots video.mp4 [count]`，或回复视频并使用 `/screenshots [count]`",
			"media_screenshots_progress": "📸 *正在生成 %d 张截图\\.\\.\\.*",
			"media_screenshots_failed":   "❌ *截图生成失败*\n\n错误：%s",

			"media_hardsub_usage":    "用法：`/hardsub video.mp4 subtitle.srt`，或回复视频并使用 `/hardsub subtitle.srt`",
			"media_hardsub_progress": "🔥 *正在合成硬字幕 \\(Hard\\-sub\\)\\.\\.\\.*\n_由于需要重新编码视频，此过程耗时较长。_",
			"media_hardsub_failed":   "❌ *硬字幕合成失败*\n\n错误：%s\n\nOutput（末尾）：\n`%s`",

			"media_rescale_usage_reply": "请回复视频并使用 `/rescale 1280x720` 或 `/rescale 720p`",
			"media_rescale_usage_path":  "用法：`/rescale video.mp4 1280x720` 或 `/rescale video.mp4 720p`",
			"media_rescale_invalid":     "❌ *分辨率无效*\n\n请使用格式：`1280x720` 或 `720p`",
			"media_rescale_progress":    "📐 *正在将视频分辨率调整为 %s\\.\\.\\.*",
			"media_rescale_failed":      "❌ *视频分辨率调整失败*\n\n错误：%s\n\nOutput（末尾）：\n`%s`",

			"media_info_title":        "✨ 媒体信息 ✨",
			"media_info_file":         "📄 *文件：* `%s`",
			"media_info_size":         "⚖️ *大小：* `%s`",
			"media_info_duration":     "🕒 *时长：* `%s`",
			"media_info_bitrate":      "📊 *总比特率：* `%s/s`",
			"media_info_format":       "📦 *格式：* `%s`",
			"media_info_video_stream": "🎬 *视频流 \\#%d*",
			"media_info_codec":        " • *编码：* `%s`",
			"media_info_resolution":   " • *分辨率：* `%dx%d`",
			"media_info_aspect":       " • *宽高比：* `%s`",
			"media_info_pixfmt":       " • *像素格式：* `%s`",
			"media_info_fps":          " • *帧率：* `%.2f fps`",
			"media_info_audio_stream": "🎵 *音频流 \\#%d*",
			"media_info_channels":     " • *声道：* `%d` （%s）",
			"media_info_sample_rate":  " • *采样率：* `%.1f kHz`",
			"media_info_audio_rate":   " • *比特率：* `%s/s`",
		},
	})
}
