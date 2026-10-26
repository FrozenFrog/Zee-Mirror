package media

import (
	"strings"
	"testing"

	"zee-mirror/pkg/i18n"
)

// The Vietnamese rendering must stay byte-identical to the strings that were
// hardcoded before the i18n migration.
func TestVietnameseMessagesUnchanged(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			"extract usage",
			GetErrorMessage(i18n.T("vi", "media_title_invalid_format"), i18n.T("vi", "media_extract_usage")),
			"❌ *SAI ĐỊNH DẠNG*\n\nHãy reply vào một file video bằng /extractaudio hoặc dùng:\n`/extractaudio path`",
		},
		{
			"compress usage",
			usageMessage("vi", "media_compress_usage"),
			"⚠️ *Sai định dạng*\n\nHãy reply vào một file video bằng /compress hoặc dùng:\n`/compress path/to/video [quality]`\n\nChất lượng: low, medium, high \\(mặc định: medium\\)",
		},
		{
			"hardsub progress",
			i18n.T("vi", "media_hardsub_progress"),
			"🔥 *Đang ghép cứng subtitle \\(Hard\\-sub\\)\\.\\.\\.*\n_Quá trình này mất nhiều thời gian hơn vì phải re\\-encode video\\._",
		},
		{
			"ffmpeg failure",
			GetErrorMessage(i18n.T("vi", "media_title_ffmpeg_error"), ffmpegFailure("vi", "media_extract_failed", errTest("exit status 1"), []byte("line"))),
			"❌ *LỖI FFMPEG*\n\nTách audio thất bại: exit status 1\n\nOutput \\(phần cuối\\):\n`line`",
		},
		{
			"rescale done",
			GetSuccessMessage(i18n.T("vi", "media_title_rescale_done"), i18n.T("vi", "media_done_rescale", "a.mp4", "720p")),
			"✅ *ĐÃ ĐỔI ĐỘ PHÂN GIẢI VIDEO*\n\n🎬 Output: `a.mp4`\\!\n📐 Độ phân giải: 720p\n\nBấm nút bên dưới để tải lên Drive\\.",
		},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s:\n got %q\nwant %q", tt.name, tt.got, tt.want)
		}
	}
}

type errTest string

func (e errTest) Error() string { return string(e) }

func TestGetErrorMessageEscapesTitle(t *testing.T) {
	got := GetErrorMessage("FFMPEG (v2.0)", "body")
	if want := "❌ *FFMPEG \\(v2\\.0\\)*\n\nbody"; got != want {
		t.Errorf("GetErrorMessage = %q, want %q", got, want)
	}
}

func TestUploadKeyboardIsLocalized(t *testing.T) {
	for _, l := range i18n.Languages {
		kb := uploadKeyboard(l.Code, "abc")
		btn := kb.InlineKeyboard[0][0]
		if btn.Text != i18n.T(l.Code, "media_btn_upload_drive") || btn.CallbackData == nil || *btn.CallbackData != "media_m:abc" {
			t.Errorf("%s: unexpected button %+v", l.Code, btn)
		}
		if strings.HasPrefix(btn.Text, "media_") {
			t.Errorf("%s: untranslated button text %q", l.Code, btn.Text)
		}
	}
}
