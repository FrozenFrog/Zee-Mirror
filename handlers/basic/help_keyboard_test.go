package basic

import (
	"strings"
	"testing"

	"zee-mirror/internal/service"
)

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
