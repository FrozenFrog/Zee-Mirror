package i18n

import "testing"

// Every key must exist in DefaultLang, since other languages fall back to it.
func TestDefaultLangCoversAllKeys(t *testing.T) {
	def := translations[DefaultLang]
	for lang, msgs := range translations {
		for key := range msgs {
			if _, ok := def[key]; !ok {
				t.Errorf("key %q (%s) missing from %q", key, lang, DefaultLang)
			}
		}
	}
}

func TestTFallback(t *testing.T) {
	tests := []struct {
		lang, key, want string
	}{
		{"", "help_close", "✖️ Đóng"},
		{"xx", "help_close", "✖️ Đóng"},
		{"EN", "help_close", "✖️ Close"},
		{"ja", "help_content", translations[DefaultLang]["help_content"]},
		{"vi", "no_such_key", "no_such_key"},
	}
	for _, tt := range tests {
		if got := T(tt.lang, tt.key); got != tt.want {
			t.Errorf("T(%q, %q) = %q, want %q", tt.lang, tt.key, got, tt.want)
		}
	}
	if got := T("vi", "task_cancelled", "abc"); got != "Đã hủy task `abc`\\." {
		t.Errorf("formatted = %q", got)
	}
}
