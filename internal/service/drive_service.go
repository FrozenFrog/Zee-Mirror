package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"zee-mirror/internal/domain"
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
		return nil, fmt.Errorf("%w: rclone lsjson failed: %v", domain.ErrExternal, err)
	}

	var files []DriveFile
	if err := json.Unmarshal(output, &files); err != nil {
		return nil, fmt.Errorf("%w: failed to parse response: %v", domain.ErrExternal, err)
	}

	return files, nil
}

func (s *BotService) FormatDriveFileList(path string, files []DriveFile) string {
	var text strings.Builder

	text.WriteString("📂 *QUẢN LÝ FILE \\- GDRIVE*\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")
	fmt.Fprintf(&text, "📍 *ĐƯỜNG DẪN:* `%s`\n\n", utils.EscapeMarkdownV2Code(path))

	if len(files) == 0 {
		text.WriteString("📭 _Thư mục này trống_\n")
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
			text.WriteString("📁 *THƯ MỤC*\n")
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
			text.WriteString("📄 *FILE*\n")
			for _, f := range fileList {
				text.WriteString(f + "\n")
			}
			text.WriteString("\n")
		}

		text.WriteString("📊 *TỔNG KẾT*\n")
		fmt.Fprintf(&text, "📂 Thư mục: `%d`\n", folderCount)
		fmt.Fprintf(&text, "📄 File: `%d`\n", fileCount)
	}

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")
	return text.String()
}
