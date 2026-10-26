package downloader

import (
	"testing"

	"zee-mirror/pkg/i18n"
)

func TestProgressUpdateLocalized(t *testing.T) {
	var up ProgressUpdate
	up.SetMessage("plugin_ytdlp_merging")
	up.SetError("plugin_ytdlp_hardsub_failed", "boom")

	if up.Message != i18n.T(i18n.DefaultLang, "plugin_ytdlp_merging") {
		t.Errorf("Message = %q, want DefaultLang text", up.Message)
	}
	if got := up.MessageIn("en"); got != "Merging video/audio..." {
		t.Errorf("MessageIn(en) = %q", got)
	}
	if got := up.ErrorIn("en"); got != "Burning in subtitles failed: boom" {
		t.Errorf("ErrorIn(en) = %q", got)
	}

	plain := ProgressUpdate{Message: "raw", Error: "ERROR: x"}
	if plain.MessageIn("en") != "raw" || plain.ErrorIn("en") != "ERROR: x" {
		t.Error("untranslated updates must pass through unchanged")
	}
}
