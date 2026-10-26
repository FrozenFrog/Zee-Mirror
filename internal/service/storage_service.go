package service

import (
	"context"
	"strings"

	"zee-mirror/internal/domain"
	"zee-mirror/pkg/i18n"
)

type StorageProvider struct {
	Name string
	Type string
	Icon string
}

func (s *BotService) GetAvailableStorages() ([]StorageProvider, error) {
	cmd := execCommand(context.Background(), "rclone", "listremotes")
	output, err := cmd.Output()
	if err != nil {
		return nil, i18n.WrapError(domain.ErrExternal, "storage_err_list_remotes", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var providers []StorageProvider
	for _, line := range lines {
		name := strings.TrimSuffix(line, ":")
		if name == "" {
			continue
		}

		providers = append(providers, StorageProvider{
			Name: name,
			Type: "rclone",
			Icon: "📁",
		})
	}

	return providers, nil
}
