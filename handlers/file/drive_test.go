package file

import (
	"strings"
	"testing"
	"unicode/utf8"

	"zee-mirror/internal/service"
)

func TestGetFileIcon(t *testing.T) {
	if got := getFileIcon("movie.MP4"); got == "" {
		t.Fatal("expected non-empty icon")
	}
	if getFileIcon("a.mkv") != getFileIcon("b.mp4") {
		t.Error("same category extensions should share icon")
	}
	if getFileIcon("a.mp4") == getFileIcon("b.mp3") {
		t.Error("video and audio icons should differ")
	}
	if got := getFileIcon("notes.txt"); got != IconFile {
		t.Errorf("unknown extension should fall back to IconFile, got %q", got)
	}
}

func TestFormatDriveFileList(t *testing.T) {
	empty := formatDriveFileList("vi", "/root", nil)
	if !strings.Contains(empty, "trống") {
		t.Error("empty folder should report trống")
	}

	files := []service.DriveFile{
		{Name: "sub", IsDir: true},
		{Name: "movie.mkv", Size: 1024},
	}
	out := formatDriveFileList("vi", "/root", files)
	for _, want := range []string{"THƯ MỤC", "FILE", "sub", "movie.mkv", "Thư mục: `1`", "File: `1`"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestBuildDriveNavigationKeyboard(t *testing.T) {
	s := &service.BotService{}
	long := strings.Repeat("thư-mục-rất-dài-", 6)
	files := []service.DriveFile{
		{Name: long, IsDir: true},
		{Name: "a.mkv"},
	}

	kb := buildDriveNavigationKeyboard(s, "vi", files, "Movies")
	if len(kb.InlineKeyboard) != 3 {
		t.Fatalf("rows = %d, want folder, file and nav rows", len(kb.InlineKeyboard))
	}
	if got := *kb.InlineKeyboard[1][0].CallbackData; got != "dr:i:Movies/a.mkv" {
		t.Errorf("file callback = %q", got)
	}
	for _, row := range kb.InlineKeyboard {
		for _, b := range row {
			if len(*b.CallbackData) > 64 {
				t.Errorf("callback_data %q exceeds 64 bytes", *b.CallbackData)
			}
			if !utf8.ValidString(b.Text) {
				t.Errorf("button text %q is not valid UTF-8", b.Text)
			}
		}
	}
	if !strings.HasPrefix(*kb.InlineKeyboard[0][0].CallbackData, "dr:c:id:") {
		t.Errorf("long folder path should use a cached id, got %q", *kb.InlineKeyboard[0][0].CallbackData)
	}
}

func TestFormatDriveFileListEnglish(t *testing.T) {
	files := []service.DriveFile{
		{Name: "sub", IsDir: true},
		{Name: "movie.mkv", Size: 1024},
	}
	out := formatDriveFileList("en", "/root", files)
	for _, want := range []string{"FILE MANAGER", "FOLDERS", "FILES", "Folders: `1`", "Files: `1`"} {
		if !strings.Contains(out, want) {
			t.Errorf("en output missing %q", want)
		}
	}
	if strings.Contains(out, "THƯ MỤC") {
		t.Error("en output must not contain Vietnamese labels")
	}
	if empty := formatDriveFileList("en", "/root", nil); !strings.Contains(empty, "empty") {
		t.Error("empty folder should report empty in English")
	}
}

func TestBuildDriveNavigationKeyboardLocalizedLabels(t *testing.T) {
	s := &service.BotService{}
	kb := buildDriveNavigationKeyboard(s, "en", nil, "")
	nav := kb.InlineKeyboard[len(kb.InlineKeyboard)-1]
	if len(nav) != 3 || nav[0].Text != "🔄 Refresh" || nav[1].Text != "🏠 Root" || nav[2].Text != "✖️ Close" {
		t.Errorf("unexpected en nav row: %+v", nav)
	}
	if got := *nav[1].CallbackData; got != "dr:h" {
		t.Errorf("callback data must stay language independent, got %q", got)
	}
}

func TestFormatSearchResultsLocalized(t *testing.T) {
	files := []service.DriveFile{{Name: "a.mkv", Path: "dir/a.mkv", Size: 10}}
	vi := formatSearchResults("vi", "a", files)
	en := formatSearchResults("en", "a", files)
	if !strings.Contains(vi, "KẾT QUẢ TÌM KIẾM") || !strings.Contains(en, "SEARCH RESULTS") {
		t.Errorf("titles not localized:\nvi=%q\nen=%q", vi, en)
	}
	if !strings.Contains(en, "dir/a\\.mkv") {
		t.Errorf("path must stay MarkdownV2-escaped, got %q", en)
	}
}
