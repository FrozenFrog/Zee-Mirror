package basic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"zee-mirror/internal/service"
	"zee-mirror/pkg/chart"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// errorCard renders service.GetErrorMessage with a translated title and body.
// The body key holds pre-escaped MarkdownV2; the title is escaped here because
// GetErrorMessage inserts it verbatim.
func errorCard(lang, titleKey, bodyKey string, args ...any) string {
	return service.GetErrorMessage(utils.EscapeMarkdownV2(i18n.T(lang, titleKey)), i18n.T(lang, bodyKey, args...))
}

func HandlePing(s *service.BotService, message *tgbotapi.Message) {
	lang := senderLang(s, message)
	start := time.Now()

	msg := tgbotapi.NewMessage(message.Chat.ID, i18n.T(lang, "basic_pinging"))
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	sentMsg, err := s.Bot.Send(msg)
	if err != nil {
		return
	}

	elapsed := time.Since(start)

	text := i18n.T(lang, "basic_pong", elapsed.Round(time.Millisecond))
	editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sentMsg.MessageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)
}

func HandleSpeed(s *service.BotService, message *tgbotapi.Message) {
	lang := senderLang(s, message)
	msg := tgbotapi.NewMessage(message.Chat.ID, i18n.T(lang, "basic_speedtest_running"))
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	sentMsg, err := s.Bot.Send(msg)
	if err != nil {
		return
	}

	go func() {
		cmd := execCommand(context.Background(), "speedtest-cli", "--simple")
		output, err := cmd.CombinedOutput()
		if err != nil {
			text := i18n.T(lang, "basic_speedtest_error", utils.EscapeMarkdownV2Code(err.Error()))
			editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sentMsg.MessageID, text)
			editMsg.ParseMode = tgbotapi.ModeMarkdownV2
			_, _ = s.Bot.Send(editMsg)
			return
		}

		outStr := string(output)
		lines := strings.Split(strings.TrimSpace(outStr), "\n")
		var result strings.Builder
		result.WriteString(i18n.T(lang, "basic_speedtest_result") + "\n\n")
		for _, line := range lines {
			fmt.Fprintf(&result, "• `%s`\n", utils.EscapeMarkdownV2Code(line))
		}

		dlRe := regexp.MustCompile(`Download:\s+([0-9.]+)\s+Mbit/s`)
		ulRe := regexp.MustCompile(`Upload:\s+([0-9.]+)\s+Mbit/s`)

		var dlVal, ulVal float64
		if match := dlRe.FindStringSubmatch(outStr); len(match) > 1 {
			dlVal, _ = strconv.ParseFloat(match[1], 64)
		}
		if match := ulRe.FindStringSubmatch(outStr); len(match) > 1 {
			ulVal, _ = strconv.ParseFloat(match[1], 64)
		}

		chartBytes, chartErr := chart.GenerateSpeedtestChart(dlVal, ulVal)

		if chartErr == nil && len(chartBytes) > 0 {
			_, _ = s.Bot.Request(tgbotapi.NewDeleteMessage(message.Chat.ID, sentMsg.MessageID))

			msg := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FileBytes{Name: "speedtest.png", Bytes: chartBytes})
			msg.Caption = result.String()
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			_, _ = s.Bot.Send(msg)
		} else {
			editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sentMsg.MessageID, result.String())
			editMsg.ParseMode = tgbotapi.ModeMarkdownV2
			_, _ = s.Bot.Send(editMsg)
		}
	}()
}

func HandleSystem(s *service.BotService, message *tgbotapi.Message) {
	if !s.CanUse(message) {
		return
	}

	lang := senderLang(s, message)
	stats := s.GetSystemStats()
	text := s.FormatSystemStatsLang(lang, stats)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.ReplyMarkup = systemKeyboard(lang)
	_, _ = s.Bot.Send(msg)
}

func HandleHealth(s *service.BotService, message *tgbotapi.Message) {
	if !s.CanUse(message) {
		return
	}

	lang := senderLang(s, message)
	checks := s.PerformHealthChecksLang(lang)
	text := s.FormatHealthCheckLang(lang, checks)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(msg)
}

func HandleLogs(s *service.BotService, message *tgbotapi.Message, _ string) {
	lang := senderLang(s, message)
	if !s.IsAdmin(message.From.ID) {
		s.Reply(message, errorCard(lang, "sys_logs_denied_hdr", "sys_logs_denied"))
		return
	}

	logPath := filepath.Join(s.Config.ConfigDir, "zee-mirror.log")

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		s.Reply(message, errorCard(lang, "sys_logs_missing_hdr", "sys_logs_missing"))
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	tempFileName := fmt.Sprintf("zee-mirror_logs_%s.log", timestamp)
	tempPath := filepath.Join(os.TempDir(), tempFileName)

	input, err := os.ReadFile(filepath.Clean(logPath))
	if err != nil {
		s.Reply(message, errorCard(lang, "sys_logs_read_hdr", "sys_logs_read_err", utils.EscapeMarkdownV2(err.Error())))
		return
	}

	// #nosec G703 -- temp file name is constants + timestamp under os.TempDir()
	err = os.WriteFile(tempPath, input, 0600)
	if err != nil {
		s.Reply(message, errorCard(lang, "sys_logs_write_hdr", "sys_logs_write_err", utils.EscapeMarkdownV2(err.Error())))
		return
	}
	defer func() { _ = os.Remove(tempPath) }()

	file := tgbotapi.NewDocument(message.Chat.ID, tgbotapi.FilePath(tempPath))
	file.Caption = i18n.T(lang, "sys_logs_caption", utils.EscapeMarkdownV2(tempFileName))
	file.ParseMode = tgbotapi.ModeMarkdownV2

	_, err = s.Bot.Send(file)
	if err != nil {
		s.Reply(message, errorCard(lang, "sys_logs_send_hdr", "sys_logs_send_err", utils.EscapeMarkdownV2(err.Error())))
	}
}

func systemKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "basic_btn_refresh"), "system:"+CmdRefresh),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "sys_btn_detail"), "system:detailed"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "sys_btn_logs"), "system:logs"),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "sys_btn_cleanup"), "system:cleanup"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_close"), "system:"+CmdClose),
		),
	)
}

func HandleSystemCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) < 2 {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		return
	}

	action := parts[1]
	lang := s.GetUserLanguage(callback.From.ID)

	switch action {
	case CmdRefresh:
		stats := s.GetSystemStats()
		text := s.FormatSystemStatsLang(lang, stats)

		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
		kb := systemKeyboard(lang)
		editMsg.ReplyMarkup = &kb
		_, _ = s.Bot.Send(editMsg)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "sys_toast_refreshed")))

	case "detailed":
		checks := s.PerformHealthChecksLang(lang)
		text := s.FormatHealthCheckLang(lang, checks)

		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
		kb := systemKeyboard(lang)
		editMsg.ReplyMarkup = &kb
		_, _ = s.Bot.Send(editMsg)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "sys_toast_detail")))

	case "logs":
		msg := &tgbotapi.Message{
			Chat: callback.Message.Chat,
			From: callback.From,
		}
		HandleLogs(s, msg, "")
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	case "logs_file":
		msg := &tgbotapi.Message{
			Chat: callback.Message.Chat,
			From: callback.From,
		}
		HandleLogs(s, msg, "file")
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "sys_toast_sending")))

	case "cleanup":
		if callback.From.ID != s.Config.OwnerID {
			_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "sys_toast_owner_only")))
			return
		}
		go s.PerformDiskCleanup()
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "sys_toast_cleanup")))

	case CmdClose:
		deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
		_, _ = s.Bot.Request(deleteMsg)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "basic_toast_closed")))
		return
	}
}
