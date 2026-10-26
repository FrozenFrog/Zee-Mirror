package i18n

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"testing"
)

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

// Every supported language must translate every key; the fallback is a
// safety net, not a substitute for a translation.
func TestAllLanguagesCoverAllKeys(t *testing.T) {
	for _, l := range Languages {
		msgs, ok := translations[l.Code]
		if !ok {
			t.Fatalf("language %q has no catalog", l.Code)
		}
		for key := range translations[DefaultLang] {
			if _, ok := msgs[key]; !ok {
				t.Errorf("key %q missing from %q", key, l.Code)
			}
		}
	}
	if len(translations) != len(Languages) {
		t.Errorf("translations has %d languages, Languages lists %d", len(translations), len(Languages))
	}
}

var verbRe = regexp.MustCompile(`%(?:\[\d+\])?[-+# 0]*\d*(?:\.\d+)?[a-zA-Z%]`)

func verbs(s string) []string {
	var out []string
	for _, v := range verbRe.FindAllString(s, -1) {
		if v == "%%" {
			continue
		}
		out = append(out, regexp.MustCompile(`\[\d+\]`).ReplaceAllString(v, ""))
	}
	slices.Sort(out)
	return out
}

// Translations must take the same fmt arguments as DefaultLang, otherwise
// T renders %!d(string=…) garbage.
func TestPlaceholdersMatch(t *testing.T) {
	for key, base := range translations[DefaultLang] {
		want := verbs(base)
		for lang, msgs := range translations {
			if msg, ok := msgs[key]; ok && !slices.Equal(verbs(msg), want) {
				t.Errorf("%s/%s: verbs %v, want %v", lang, key, verbs(msg), want)
			}
		}
	}
}

// mdv2Unescaped returns the first MarkdownV2 reserved character that is not
// escaped outside code spans and links, or 0.
func mdv2Unescaped(s string) rune {
	r := []rune(s)
	inCode := false
	for i := 0; i < len(r); i++ {
		c := r[i]
		switch {
		case c == '\\':
			i++
		case c == '`':
			inCode = !inCode
		case inCode:
		case c == '%':
			if m := verbRe.FindStringIndex(string(r[i:])); m != nil && m[0] == 0 {
				i += len([]rune(string(r[i:])[:m[1]])) - 1
			}
		case c == ']' && i+1 < len(r) && r[i+1] == '(':
			for i < len(r) && r[i] != ')' {
				i++
			}
		case c < 127 && containsRune(".!()-=+#|{}>", c):
			return c
		}
	}
	return 0
}

func containsRune(set string, c rune) bool {
	for _, x := range set {
		if x == c {
			return true
		}
	}
	return false
}

// A key whose DefaultLang text contains MarkdownV2 escapes is sent with
// ParseMode MarkdownV2; an unescaped reserved character in any translation
// makes Telegram reject the whole message.
func TestMarkdownV2Escaping(t *testing.T) {
	for key, base := range translations[DefaultLang] {
		if !regexp.MustCompile(`\\[^\\]`).MatchString(base) {
			continue
		}
		for lang, msgs := range translations {
			if msg, ok := msgs[key]; ok {
				if c := mdv2Unescaped(msg); c != 0 {
					t.Errorf("%s/%s: unescaped %q in %q", lang, key, c, msg)
				}
			}
		}
	}
}

func TestMdv2Unescaped(t *testing.T) {
	for in, want := range map[string]rune{
		`Done\.`:              0,
		`Done.`:               '.',
		"`a.b` ok":            0,
		`%d files\.`:          0,
		`%-10s x`:             0,
		`[link](https://a.b)`: 0,
		`Sub-task`:            '-',
		"*bold* _it_ ~s~ 完了。": 0,
	} {
		if got := mdv2Unescaped(in); got != want {
			t.Errorf("mdv2Unescaped(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalize(t *testing.T) {
	for in, want := range map[string]string{
		"en": "en", "en-US": "en", "ZH-hans": "zh", "pt-br": "", "": "", "vi": "vi",
	} {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
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

func TestErrorMessage(t *testing.T) {
	cause := errors.New("boom")
	err := fmt.Errorf("wrapped: %w", WrapError(cause, "task_not_found", "x1"))
	if got := Message("vi", err); got != "Không tìm thấy task/batch `x1`" {
		t.Errorf("Message(vi) = %q", got)
	}
	if got := err.Error(); got != "wrapped: Task/Batch `x1` not found" {
		t.Errorf("Error() = %q", got)
	}
	if !errors.Is(err, cause) {
		t.Error("cause not reachable")
	}
	if got := Message("en", cause); got != "boom" {
		t.Errorf("plain error = %q", got)
	}
	if _, ok := Lookup("en", "no_such_key"); ok {
		t.Error("Lookup found missing key")
	}
}
