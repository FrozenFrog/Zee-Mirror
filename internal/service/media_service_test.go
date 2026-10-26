package service_test

import (
	"strings"
	"testing"

	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"
)

func sampleProbe() service.FFProbeOutput {
	return service.FFProbeOutput{
		Format: service.FFFormat{
			Filename:   "clip.mp4",
			FormatName: "mov,mp4",
			Duration:   "65.5",
			Size:       "1048576",
			BitRate:    "8000000",
		},
		Streams: []service.FFStream{
			{
				CodecType:          "video",
				CodecName:          "h264",
				Width:              1920,
				Height:             1080,
				DisplayAspectRatio: "16:9",
				PixFmt:             "yuv420p",
				RFrameRate:         "30000/1001",
			},
			{
				CodecType:     "audio",
				CodecName:     "aac",
				Channels:      2,
				ChannelLayout: "stereo",
				SampleRate:    "48000",
				BitRate:       "128000",
			},
		},
	}
}

func TestFormatMediaInfoLanguages(t *testing.T) {
	m := service.NewMediaService(nil)
	probe := sampleProbe()

	vi := m.FormatMediaInfo("vi", "clip.mp4", probe)
	for _, want := range []string{
		"THÔNG TIN MEDIA",
		"*Dung lượng:* `1.00 MB`",
		"*Thời lượng:* `1m 5s`",
		"*LUỒNG VIDEO \\#1*",
		"*Độ phân giải:* `1920x1080`",
		"*Tốc độ khung hình:* `29.97 fps`",
		"*LUỒNG AUDIO \\#1*",
		"*Kênh:* `2` \\(stereo\\)",
		"*Tần số lấy mẫu:* `48.0 kHz`",
	} {
		if !strings.Contains(vi, want) {
			t.Errorf("vi report missing %q:\n%s", want, vi)
		}
	}

	en := m.FormatMediaInfo("en", "clip.mp4", probe)
	for _, want := range []string{"MEDIA INFO", "*Size:* `1.00 MB`", "*VIDEO STREAM \\#1*", "*Channels:* `2` \\(stereo\\)"} {
		if !strings.Contains(en, want) {
			t.Errorf("en report missing %q:\n%s", want, en)
		}
	}
	if strings.Contains(en, "Dung lượng") {
		t.Errorf("en report still contains Vietnamese:\n%s", en)
	}

	// Every language must produce the same line structure.
	wantLines := strings.Count(vi, "\n")
	for _, l := range i18n.Languages {
		got := m.FormatMediaInfo(l.Code, "clip.mp4", probe)
		if n := strings.Count(got, "\n"); n != wantLines {
			t.Errorf("%s report has %d lines, want %d", l.Code, n, wantLines)
		}
		if strings.Contains(got, "media_info_") {
			t.Errorf("%s report contains an untranslated key:\n%s", l.Code, got)
		}
	}
}
