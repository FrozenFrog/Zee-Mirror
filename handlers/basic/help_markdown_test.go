package basic

import (
	"fmt"
	"strings"
	"testing"

	"zee-mirror/pkg/i18n"
)

// checkMarkdownV2 reports the first construct Telegram would reject in a
// MarkdownV2 text: an unescaped reserved character or an unclosed entity.
// It covers the subset the help pages use (bold, italic, strikethrough,
// inline code), so links and pre blocks are reported as errors.
func checkMarkdownV2(text string) error {
	const reserved = "_*[]()~`>#+-=|{}.!"
	runes := []rune(text)
	open := map[rune]int{}
	inCode := false
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		if c == '\\' {
			if i+1 >= len(runes) || runes[i+1] > 126 {
				return fmt.Errorf("dangling backslash at %d", i)
			}
			i++
			continue
		}
		if inCode {
			if c == '`' {
				inCode = false
			}
			continue
		}
		switch c {
		case '`':
			inCode = true
		case '*', '_', '~':
			open[c] ^= 1
		default:
			if strings.ContainsRune(reserved, c) {
				return fmt.Errorf("unescaped %q at %d: %q", c, i, around(runes, i))
			}
		}
	}
	if inCode {
		return fmt.Errorf("unclosed inline code")
	}
	for c, n := range open {
		if n != 0 {
			return fmt.Errorf("unclosed %q entity", c)
		}
	}
	return nil
}

func around(runes []rune, i int) string {
	return string(runes[max(0, i-20):min(len(runes), i+20)])
}

func TestCheckMarkdownV2(t *testing.T) {
	valid := []string{"*a* `x.y` \\. _b_", "`a\\`b`", "~s~"}
	invalid := []string{"a.b", "*a", "`a", "(x)", "a\\"}
	for _, s := range valid {
		if err := checkMarkdownV2(s); err != nil {
			t.Errorf("%q: unexpected error %v", s, err)
		}
	}
	for _, s := range invalid {
		if checkMarkdownV2(s) == nil {
			t.Errorf("%q: expected an error", s)
		}
	}
}

// allHelpPages returns every help page renderer, keyed by name.
func allHelpPages() map[string]func(lang string) string {
	pages := map[string]func(string) string{
		"settings": GetHelpSettings, "notify": GetHelpNotify,
		"all_empty": func(lang string) string { return GenerateAllCommandsHelp(lang, nil) },
		"all": func(lang string) string {
			return GenerateAllCommandsHelp(lang, []HelpCommandEntry{
				{Name: "ytdlp_leech", Description: "Tải (YT-DLP) về Telegram.", Category: "download", Aliases: []string{"yl", "y_l"}},
				{Name: "stats", Description: "Thống kê", Category: "monitor"},
				{Name: "mirror", Category: "download", Aliases: []string{"m"}},
			})
		},
	}
	for action, page := range helpCommandPages {
		pages[action] = page.render
	}
	for cat, buttons := range helpCategoryButtons {
		pages["category_"+cat] = func(lang string) string {
			var sb strings.Builder
			for _, b := range buttons {
				sb.WriteString(b.label(lang))
			}
			return sb.String()
		}
	}
	return pages
}

func TestHelpPagesAreValidMarkdownV2(t *testing.T) {
	for _, l := range i18n.Languages {
		for name, page := range allHelpPages() {
			if strings.HasPrefix(name, "category_") {
				continue // button labels are plain text
			}
			text := page(l.Code)
			if err := checkMarkdownV2(text); err != nil {
				t.Errorf("%s help page %s: %v", l.Code, name, err)
			}
		}
	}
}

// Every help page and button must be translated, never falling back to a
// raw catalog key or an empty section.
func TestHelpPagesTranslated(t *testing.T) {
	pageNames := []string{"settings", "notify"}
	for action := range helpCommandPages {
		pageNames = append(pageNames, strings.TrimPrefix(action, "cmd_"))
	}
	for _, l := range i18n.Languages {
		for _, page := range pageNames {
			for _, part := range []string{"title", "desc", "usage"} {
				if s, ok := i18n.Lookup(l.Code, "helptext_"+page+"_"+part); !ok || s == "" {
					t.Errorf("%s: helptext_%s_%s missing", l.Code, page, part)
				}
			}
		}
		for name, page := range allHelpPages() {
			if text := page(l.Code); text == "" || strings.Contains(text, "helptext_") {
				t.Errorf("%s help page %s renders %q", l.Code, name, text)
			}
		}
	}
	if GetHelpMirror("en") == GetHelpMirror("vi") {
		t.Error("English mirror help equals Vietnamese")
	}
	if !strings.Contains(GetHelpMirror("vi"), "Tải file từ URL rồi upload thẳng lên Google Drive") {
		t.Error("Vietnamese mirror help changed")
	}
}

func TestAllCommandsHelpLocalizesDescriptions(t *testing.T) {
	cmds := []HelpCommandEntry{
		{Name: "mirror", Description: "fallback", Category: "download"},
		{Name: "nokey", Description: "Plain fallback", Category: "download"},
	}
	en := GenerateAllCommandsHelp("en", cmds)
	if !strings.Contains(en, "Upload a file from a URL to Google Drive") {
		t.Errorf("en description not localized: %s", en)
	}
	if !strings.Contains(en, "Plain fallback") {
		t.Errorf("missing fallback description: %s", en)
	}
	if !strings.Contains(GenerateAllCommandsHelp("vi", cmds), "Upload file từ URL lên Google Drive") {
		t.Error("vi description not localized")
	}
}
