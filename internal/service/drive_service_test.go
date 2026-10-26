package service

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"

	"zee-mirror/internal/domain"
	"zee-mirror/pkg/i18n"
)

func TestFormatDriveFileListLangLocalized(t *testing.T) {
	bs := &BotService{}
	files := []DriveFile{
		{Name: "Movies", IsDir: true},
		{Name: "file1.zip", Size: 1024},
	}

	vi := bs.FormatDriveFileListLang("vi", "/data", files)
	if !strings.Contains(vi, "Thư mục: `1`") || !strings.Contains(vi, "THƯ MỤC") {
		t.Errorf("vi output not localized: %q", vi)
	}
	if got := bs.FormatDriveFileList("/data", files); got != vi {
		t.Error("FormatDriveFileList must render the default language")
	}

	en := bs.FormatDriveFileListLang("en", "/data", files)
	for _, want := range []string{"FILE MANAGER", "FOLDERS", "Folders: `1`", "Files: `1`", "Movies", "file1.zip"} {
		if !strings.Contains(en, want) {
			t.Errorf("en output missing %q in %q", want, en)
		}
	}
	if empty := bs.FormatDriveFileListLang("en", "/data", nil); !strings.Contains(empty, "empty") {
		t.Errorf("en empty list = %q", empty)
	}
}

func TestGetAvailableStoragesErrorIsTranslatable(t *testing.T) {
	orig := execCommand
	t.Cleanup(func() { execCommand = orig })
	execCommand = func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "false")
	}

	_, err := (&BotService{}).GetAvailableStorages()
	if err == nil {
		t.Fatal("expected an error")
	}
	if !errors.Is(err, domain.ErrExternal) {
		t.Error("error must still match domain.ErrExternal")
	}
	if got := i18n.Message("vi", err); !strings.HasPrefix(got, "Không liệt kê được remote rclone") {
		t.Errorf("vi message = %q", got)
	}
	if got := i18n.Message("en", err); !strings.HasPrefix(got, "failed to list remotes") {
		t.Errorf("en message = %q", got)
	}
}
