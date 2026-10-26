package i18n

import (
	"strings"
	"testing"
)

// Keys with these prefixes hold plain text (headers escaped by the handler,
// buttons, toasts, service errors); every other media_ key is MarkdownV2.
var mediaPlainPrefixes = []string{"media_title_", "media_btn_", "media_toast_", "media_err_", "media_info_title"}

func isMediaPlainKey(key string) bool {
	for _, p := range mediaPlainPrefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}

// MarkdownV2 media messages are stored pre-escaped, and a missing escape in
// any language makes Telegram reject the whole message. messages_test.go
// only checks keys whose Vietnamese text happens to contain a backslash, so
// check every media_ MarkdownV2 key here.
func TestMediaMarkdownV2Keys(t *testing.T) {
	for lang, msgs := range translations {
		for key, msg := range msgs {
			if !strings.HasPrefix(key, "media_") || isMediaPlainKey(key) {
				continue
			}
			if c := mdv2Unescaped(msg); c != 0 {
				t.Errorf("%s/%s: unescaped %q in %q", lang, key, c, msg)
			}
		}
	}
}

// Plain-text media keys are escaped (or sent unparsed) by the caller, so a
// backslash in the catalog would be shown literally.
func TestMediaPlainKeysHaveNoEscapes(t *testing.T) {
	for lang, msgs := range translations {
		for key, msg := range msgs {
			if strings.HasPrefix(key, "media_") && isMediaPlainKey(key) && strings.Contains(msg, `\`) {
				t.Errorf("%s/%s: plain text contains a backslash: %q", lang, key, msg)
			}
		}
	}
}

func TestMediaErrorMessage(t *testing.T) {
	err := NewError("media_err_duration_invalid")
	if got := Message("vi", err); got != "thời lượng video không hợp lệ" {
		t.Errorf("Message(vi) = %q", got)
	}
	if got := err.Error(); got != "invalid video duration" {
		t.Errorf("Error() = %q", got)
	}
}
