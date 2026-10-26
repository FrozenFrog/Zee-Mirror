package basic

import (
	"strings"
	"testing"

	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"
)

// Every category button opens a known page, every page is reachable from a
// category, and labels are translated in every language.
func TestHelpButtonsMatchPages(t *testing.T) {
	reachable := map[string]bool{}
	for cat, buttons := range helpCategoryButtons {
		for _, b := range buttons {
			reachable[b.action] = true
			if _, ok := helpCommandPages[b.action]; !ok {
				t.Errorf("%s button %q has no help page", cat, b.action)
			}
		}
		for _, l := range i18n.Languages {
			for _, row := range helpCategoryKeyboard(l.Code, buttons).InlineKeyboard {
				for _, b := range row {
					if b.Text == "" || strings.Contains(b.Text, "helptext_") {
						t.Errorf("%s/%s: untranslated button %q", l.Code, cat, b.Text)
					}
				}
			}
		}
	}
	for action, page := range helpCommandPages {
		if !reachable[action] {
			t.Errorf("help page %q is not reachable from any category", action)
		}
		if _, ok := helpCategoryButtons[page.category]; !ok {
			t.Errorf("help page %q returns to unknown category %q", action, page.category)
		}
	}
}

// Every help button must open a page the help handlers know about; this also
// guards against the keyboards being stubbed out again.
func TestHelpKeyboardsReachHandledActions(t *testing.T) {
	handled := map[string]bool{"main": true, "settings": true, "all": true, ActionBack: true, CmdClose: true}
	categories := map[string][]helpButton{
		"download": downloadHelpButtons, "monitor": monitorHelpButtons, "files": filesHelpButtons,
		"media": mediaHelpButtons, "task": taskHelpButtons, "storage": storageHelpButtons,
		"admin": adminHelpButtons, "recovery": recoveryHelpButtons,
	}
	for cat, buttons := range categories {
		handled[cat] = true
		for _, b := range buttons {
			handled[b.action] = true
		}
	}

	main := service.GetHelpKeyboard("vi")
	if len(main.InlineKeyboard) == 0 {
		t.Fatal("help keyboard is empty")
	}
	for _, row := range main.InlineKeyboard {
		for _, b := range row {
			action := strings.TrimPrefix(*b.CallbackData, "help:")
			if !handled[normalizeHelpAction(action)] {
				t.Errorf("help button %q -> %q is not handled", b.Text, *b.CallbackData)
			}
		}
	}

	start := service.GetStartKeyboard("vi")
	if len(start.InlineKeyboard) == 0 || *start.InlineKeyboard[0][0].CallbackData != "help:main" {
		t.Errorf("start keyboard should open help, got %+v", start.InlineKeyboard)
	}

	for cat, buttons := range categories {
		kb := helpCategoryKeyboard("vi", buttons)
		n := 0
		for _, row := range kb.InlineKeyboard {
			n += len(row)
		}
		if n != len(buttons)+2 {
			t.Errorf("%s keyboard has %d buttons, want %d", cat, n, len(buttons)+2)
		}
	}
}
