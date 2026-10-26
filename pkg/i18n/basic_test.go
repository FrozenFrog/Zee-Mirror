package i18n

import (
	"strings"
	"testing"
)

var basicPrefixes = []string{"stats_", "sys_", "bdash_", "basic_", "join_", "notify_cmd_", "schedule_"}

// basicPlain reports whether key is a plain-text key by the naming convention
// documented in basic.go.
func basicPlain(key string) bool {
	for _, marker := range []string{"_toast", "_btn", "_plain", "_hdr"} {
		if strings.Contains(key, marker) {
			return true
		}
	}
	return false
}

// mdv2Balanced reports whether bold markers and code fences are paired
// outside escapes, so Telegram does not reject the entity.
func mdv2Balanced(s string) bool {
	r := []rune(s)
	inCode := false
	stars := 0
	for i := 0; i < len(r); i++ {
		switch {
		case r[i] == '\\':
			i++
		case r[i] == '`':
			inCode = !inCode
		case r[i] == '*' && !inCode:
			stars++
		}
	}
	return !inCode && stars%2 == 0
}

// The basic handlers send these keys either as MarkdownV2 (stored
// pre-escaped) or as plain text; the key name says which, and every
// translation must respect it.
func TestBasicKeys(t *testing.T) {
	seen := 0
	for key := range translations[DefaultLang] {
		matched := false
		for _, p := range basicPrefixes {
			if strings.HasPrefix(key, p) {
				matched = true
			}
		}
		if !matched {
			continue
		}
		seen++
		for lang, msgs := range translations {
			msg := msgs[key]
			if basicPlain(key) {
				if strings.Contains(msg, `\`) {
					t.Errorf("%s/%s: plain text must not contain a backslash: %q", lang, key, msg)
				}
				continue
			}
			if c := mdv2Unescaped(msg); c != 0 {
				t.Errorf("%s/%s: unescaped %q in %q", lang, key, c, msg)
			}
			if !mdv2Balanced(msg) {
				t.Errorf("%s/%s: unbalanced * or ` in %q", lang, key, msg)
			}
		}
	}
	if seen == 0 {
		t.Fatal("no basic keys registered")
	}
}

func TestBasicFormats(t *testing.T) {
	for _, l := range Languages {
		if got := T(l.Code, "stats_success_rate", 87.5); !strings.Contains(got, "87.5%") {
			t.Errorf("%s: stats_success_rate = %q", l.Code, got)
		}
		if got := T(l.Code, "basic_pong", "12ms"); !strings.Contains(got, "`12ms`") {
			t.Errorf("%s: basic_pong = %q", l.Code, got)
		}
	}
}
