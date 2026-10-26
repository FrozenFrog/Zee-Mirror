package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"zee-mirror/internal/domain"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"
)

func (s *BotService) ListDriveFiles(path string) ([]DriveFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	configPath := s.TaskManager.ConfigDir + "/rclone.conf"
	args := []string{
		"lsjson",
		path,
		"--config", configPath,
		"--no-modtime",
	}

	cmd := execCommand(ctx, "rclone", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, i18n.WrapError(domain.ErrExternal, "fm_err_rclone", utils.RcloneError(err))
	}

	var files []DriveFile
	if err := json.Unmarshal(output, &files); err != nil {
		return nil, i18n.WrapError(domain.ErrExternal, "fm_err_parse", err)
	}

	return files, nil
}

// FormatDriveFileList renders the file list in the default language. Callers
// that know the user's language should use FormatDriveFileListLang.
func (s *BotService) FormatDriveFileList(path string, files []DriveFile) string {
	return s.FormatDriveFileListLang(i18n.DefaultLang, path, files)
}

// FormatDriveFileListLang renders the MarkdownV2 file list in lang.
func (s *BotService) FormatDriveFileListLang(lang, path string, files []DriveFile) string {
	var text strings.Builder

	text.WriteString(i18n.T(lang, "fm_title") + "\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")
	text.WriteString(i18n.T(lang, "fm_path", utils.EscapeMarkdownV2Code(path)) + "\n\n")

	if len(files) == 0 {
		text.WriteString(i18n.T(lang, "fm_empty") + "\n")
	} else {
		folderCount := 0
		fileCount := 0

		var folders []string
		for _, f := range files {
			if f.IsDir {
				folderCount++
				folders = append(folders, fmt.Sprintf("📁 `%s`", utils.EscapeMarkdownV2Code(utils.TruncateString(f.Name, 40))))
			}
		}

		if len(folders) > 0 {
			text.WriteString(i18n.T(lang, "fm_folders") + "\n")
			for _, f := range folders {
				text.WriteString(f + "\n")
			}
			text.WriteString("\n")
		}

		var fileList []string
		for _, f := range files {
			if !f.IsDir {
				fileCount++
				size := utils.FormatBytes(f.Size)
				fileList = append(fileList, fmt.Sprintf("📄 `%s` \\(%s\\)",
					utils.EscapeMarkdownV2Code(utils.TruncateString(f.Name, 35)),
					utils.EscapeMarkdownV2(size)))
			}
		}

		if len(fileList) > 0 {
			text.WriteString(i18n.T(lang, "fm_files") + "\n")
			for _, f := range fileList {
				text.WriteString(f + "\n")
			}
			text.WriteString("\n")
		}

		text.WriteString(i18n.T(lang, "fm_summary") + "\n")
		text.WriteString(i18n.T(lang, "fm_count_folders", folderCount) + "\n")
		text.WriteString(i18n.T(lang, "fm_count_files", fileCount) + "\n")
	}

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")
	return text.String()
}
