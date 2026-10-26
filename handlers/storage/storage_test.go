package storage

import (
	"errors"
	"strings"
	"testing"

	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"
)

func testService() *service.BotService {
	return &service.BotService{TaskManager: &service.TaskManager{RcloneDest: "gdrive:/Mirror"}}
}

func TestStorageListMessageLocalized(t *testing.T) {
	providers := []Provider{{Name: "gdrive", RemoteName: "gdrive:", Type: "drive", Icon: "📁"}}

	text, _ := storageListMessage(testService(), "vi", providers)
	for _, want := range []string{"Các nhà cung cấp lưu trữ khả dụng", "*Hiện tại:*", "gdrive:/Mirror"} {
		if !strings.Contains(text, want) {
			t.Errorf("vi text missing %q in %q", want, text)
		}
	}

	text, kb := storageListMessage(testService(), "en", providers)
	if !strings.Contains(text, "Available storage providers") || !strings.Contains(text, "*Current:*") {
		t.Errorf("en text not localized: %q", text)
	}
	if strings.Contains(text, "lưu trữ") {
		t.Errorf("en text leaks Vietnamese: %q", text)
	}
	if len(kb.InlineKeyboard) != 2 {
		t.Fatalf("rows = %d, want provider + close", len(kb.InlineKeyboard))
	}
	if got := *kb.InlineKeyboard[0][0].CallbackData; got != "storage:select:gdrive" {
		t.Errorf("provider callback = %q", got)
	}
	if got := kb.InlineKeyboard[1][0]; got.Text != "✖️ Close" || *got.CallbackData != "storage:close" {
		t.Errorf("close button = %q / %q", got.Text, *got.CallbackData)
	}
}

func TestStorageBackKeyboardLocalized(t *testing.T) {
	kb := storageBackKeyboard("en", "gdrive")
	row := kb.InlineKeyboard[0]
	if row[0].Text != "🔙 Back" || *row[0].CallbackData != "storage:select:gdrive" {
		t.Errorf("back button = %q / %q", row[0].Text, *row[0].CallbackData)
	}
}

func TestListRemotesErrorIsTranslatable(t *testing.T) {
	cause := errors.New("exit status 1")
	err := i18n.WrapError(cause, "storage_err_list_remotes", cause)
	if got := i18n.Message("vi", err); got != "Không liệt kê được remote rclone: exit status 1" {
		t.Errorf("vi = %q", got)
	}
	if got := i18n.Message("en", err); got != "failed to list remotes: exit status 1" {
		t.Errorf("en = %q", got)
	}
}
