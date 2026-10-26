package search

import (
	"strings"
	"testing"
)

func TestCleanMagnetLink(t *testing.T) {
	if got := CleanMagnetLink("vi", ""); got != "Không có Magnet" {
		t.Errorf("vi empty = %q", got)
	}
	if got := CleanMagnetLink("en", ""); got != "No Magnet" {
		t.Errorf("en empty = %q", got)
	}
	if got := CleanMagnetLink("en", "magnet:?xt=urn:btih:abc"); got != "magnet:?xt=urn:btih:abc" {
		t.Errorf("magnet URI must pass through, got %q", got)
	}
	hash := strings.Repeat("a", 40)
	if got := CleanMagnetLink("ja", hash); got != "magnet:?xt=urn:btih:"+hash {
		t.Errorf("bare hash = %q", got)
	}
}

func TestProviderKeyboardLocalized(t *testing.T) {
	vi := providerKeyboard("vi", "sess")
	en := providerKeyboard("en", "sess")

	if got := vi.InlineKeyboard[1][1].Text; got != "🌍 Tất cả" {
		t.Errorf("vi all button = %q", got)
	}
	if got := en.InlineKeyboard[1][1].Text; got != "🌍 All" {
		t.Errorf("en all button = %q", got)
	}
	// callback data must not depend on the language
	for r := range vi.InlineKeyboard {
		for c := range vi.InlineKeyboard[r] {
			if *vi.InlineKeyboard[r][c].CallbackData != *en.InlineKeyboard[r][c].CallbackData {
				t.Errorf("callback data differs between languages at [%d][%d]", r, c)
			}
		}
	}
}

func TestBackCloseKeyboardLocalized(t *testing.T) {
	kb := backCloseKeyboard("en", "sess")
	row := kb.InlineKeyboard[0]
	if row[0].Text != "🔙 Back" || row[1].Text != "❌ Close" {
		t.Errorf("unexpected labels: %q %q", row[0].Text, row[1].Text)
	}
	if *row[0].CallbackData != "t_back:sess" || *row[1].CallbackData != "t_close:sess" {
		t.Errorf("unexpected callback data: %q %q", *row[0].CallbackData, *row[1].CallbackData)
	}
}
