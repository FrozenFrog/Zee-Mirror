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
	empty := formatDriveFileList("/root", nil)
	if !strings.Contains(empty, "trống") {
		t.Error("empty folder should report trống")
	}

	files := []service.DriveFile{
		{Name: "sub", IsDir: true},
		{Name: "movie.mkv", Size: 1024},
	}
	out := formatDriveFileList("/root", files)
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

	kb := buildDriveNavigationKeyboard(s, files, "Movies")
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
