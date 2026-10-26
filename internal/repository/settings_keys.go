package repository

import "fmt"

// Keys of settings rows the bot writes at runtime. Like the config keys
// (BOT_TOKEN, RCLONE_DEST, ...) they are uppercase.
const (
	SettingDefaultMode        = "DEFAULT_MODE"
	SettingAutoDeleteMessages = "AUTO_DELETE_MESSAGES"
	SettingYtdlpQuality       = "YTDLP_QUALITY"
	SettingAlertChannelID     = "ALERT_CHANNEL_ID"
	SettingRcloneDest         = "RCLONE_DEST"
	SettingRcloneConfig       = "RCLONE_CONFIG"
	SettingCookies            = "COOKIES"
)

// DashboardMessageKey is the row holding the pinned dashboard message ID of a chat.
func DashboardMessageKey(chatID int64) string {
	return fmt.Sprintf("DASHBOARD_MSG_%d", chatID)
}
