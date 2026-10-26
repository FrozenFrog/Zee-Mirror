package userbot

import (
	"errors"
	"testing"

	"zee-mirror/pkg/i18n"
)

func TestParseMessageLinkErrorsAreTranslated(t *testing.T) {
	tests := []struct {
		link, wantVi, wantEn string
	}{
		{"https://example.com/a/1", "định dạng link không hợp lệ", "invalid link format"},
		{"https://t.me/", "định dạng link không hợp lệ", "invalid link format"},
		{"https://t.me/burning_high", "link thiếu ID tin nhắn", "link is missing the message ID"},
		{"https://t.me/c/abc/5", "ID channel không hợp lệ", "invalid channel ID"},
		{"https://t.me/name/abc", "ID tin nhắn không hợp lệ", "invalid message ID"},
	}
	for _, tt := range tests {
		_, err := parseMessageLink(tt.link)
		if err == nil {
			t.Errorf("parseMessageLink(%q): expected an error", tt.link)
			continue
		}
		if got := i18n.Message("vi", err); got != tt.wantVi {
			t.Errorf("%q: vi = %q, want %q", tt.link, got, tt.wantVi)
		}
		if got := i18n.Message("en", err); got != tt.wantEn {
			t.Errorf("%q: en = %q, want %q", tt.link, got, tt.wantEn)
		}
		if got := err.Error(); got != tt.wantEn {
			t.Errorf("%q: Error() = %q, want English %q", tt.link, got, tt.wantEn)
		}
	}
}

func TestNotStartedErrors(t *testing.T) {
	u := &UserBot{}
	for name, call := range map[string]func() error{
		"DownloadFile": func() error { _, err := u.DownloadFile("https://t.me/a/1", t.TempDir(), nil); return err },
		"Join":         func() error { _, err := u.Join("hash"); return err },
	} {
		err := call()
		if err == nil {
			t.Fatalf("%s: expected an error", name)
		}
		if got := i18n.Message("vi", err); got != "userbot chưa được khởi động" {
			t.Errorf("%s: vi = %q", name, got)
		}
		if got := i18n.Message("ja", err); got != "userbot が起動していません" {
			t.Errorf("%s: ja = %q", name, got)
		}
	}
}

func TestWrappedCauseStaysReachable(t *testing.T) {
	cause := errors.New("RPC error")
	err := i18n.WrapError(cause, "userbot_download_failed", cause)
	if !errors.Is(err, cause) {
		t.Fatal("cause not reachable through errors.Is")
	}
	if got := i18n.Message("vi", err); got != "tải xuống thất bại: RPC error" {
		t.Errorf("vi = %q", got)
	}
}

func TestJoinResultMessage(t *testing.T) {
	tests := []struct {
		name   string
		res    JoinResult
		lang   string
		expect string
	}{
		{"titled vi", JoinResult{Title: "Zee Chat"}, "vi", "Đã tham gia thành công: Zee Chat"},
		{"titled en", JoinResult{Title: "Zee Chat"}, "en", "Joined successfully: Zee Chat"},
		{"untitled vi", JoinResult{}, "vi", "Đã tham gia thành công: Chat không rõ tên"},
		{"untitled en", JoinResult{}, "en", "Joined successfully: Unknown chat"},
		{"already joined", JoinResult{AlreadyJoined: true}, "vi", "Đã tham gia chat này rồi"},
	}
	for _, tt := range tests {
		if got := tt.res.Message(tt.lang); got != tt.expect {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.expect)
		}
	}
}
