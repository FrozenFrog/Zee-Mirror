package service

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"zee-mirror/internal/domain"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

func (s *BotService) ShowTorrentSelectionMenu(message *tgbotapi.Message, url, name string, zip, unzip bool, password string, replyID int) {
	sessionID := s.CreateTorrentSession(url, name, zip, unzip, password, message.Chat.ID, message.MessageID, replyID, message.From.ID)

	lang := s.GetUserLanguage(message.From.ID)
	text := i18n.T(lang, "torrent_menu_text")

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = MarkdownV2
	msg.ReplyMarkup = s.torrentMenuKeyboard(lang, sessionID)

	sentMsg, err := s.Bot.Send(msg)
	if err != nil {
		slog.Error("Failed to send torrent selection menu", "error", err)
		return
	}

	s.TaskManager.Mu.Lock()
	if session, exists := s.TaskManager.TorrentSessions[sessionID]; exists {
		session.MessageID = sentMsg.MessageID
	}
	s.TaskManager.Mu.Unlock()

	slog.Info("Torrent selection menu shown", "sessionID", sessionID, "url", url)
}

const torrentBrowsePageSize = 8

// torrentWebSelectURL returns the dashboard page for picking files, or "" when
// WEB_DASHBOARD_URL is not a public http(s) URL (Telegram rejects such buttons).
func (s *BotService) torrentWebSelectURL(sessionID string) string {
	base := strings.TrimRight(s.Config.DashboardURL, "/")
	u, err := url.Parse(base)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return ""
	}
	switch u.Hostname() {
	case "localhost", "127.0.0.1", "::1":
		return ""
	}
	return base + "/torrent-select/" + sessionID
}

func (s *BotService) torrentMenuKeyboard(lang, sessionID string) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "torrent_select_all"), "torrent_sel:all:"+sessionID),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "torrent_browse_files"), "torrent_sel:browse:"+sessionID+":0"),
		),
	}
	if webURL := s.torrentWebSelectURL(sessionID); webURL != "" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(i18n.T(lang, "torrent_select_files"), webURL),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "torrent_cancel"), "torrent_sel:cancel:"+sessionID),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func torrentBrowseKeyboard(lang, sessionID string, files []domain.TorrentFile, selected []int, offset int) tgbotapi.InlineKeyboardMarkup {
	chosen := make(map[int]bool, len(selected))
	for _, idx := range selected {
		chosen[idx] = true
	}

	end := min(offset+torrentBrowsePageSize, len(files))
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, f := range files[offset:end] {
		icon := "⬜"
		if chosen[f.Index] {
			icon = "✅"
		}
		name := []rune(f.Name)
		if len(name) > 40 {
			name = append(name[:39], '…')
		}
		label := fmt.Sprintf("%s %s (%s)", icon, string(name), utils.FormatBytes(f.Size))
		data := fmt.Sprintf("torrent_sel:toggle:%s:%d:%d", sessionID, f.Index, offset)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(label, data)))
	}

	var nav []tgbotapi.InlineKeyboardButton
	if offset > 0 {
		prev := max(offset-torrentBrowsePageSize, 0)
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData("⬅️", fmt.Sprintf("torrent_sel:page:%s:%d", sessionID, prev)))
	}
	if end < len(files) {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData("➡️", fmt.Sprintf("torrent_sel:page:%s:%d", sessionID, end)))
	}
	if len(nav) > 0 {
		rows = append(rows, nav)
	}

	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s (%d)", i18n.T(lang, "torrent_start_selected"), len(selected)), "torrent_sel:start_sel:"+sessionID),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "help_back"), "torrent_sel:back:"+sessionID),
		),
	)
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (s *BotService) CreateTorrentSession(url, name string, zip, unzip bool, password string, chatID int64, msgID, replyID int, userID int64) string {
	sessionID := uuid.New().String()[:12]

	s.TaskManager.Mu.Lock()
	defer s.TaskManager.Mu.Unlock()

	s.TaskManager.TorrentSessions[sessionID] = &TorrentSession{
		URL:           url,
		FileName:      name,
		Zip:           zip,
		Unzip:         unzip,
		Password:      password,
		SelectedFiles: nil,
		ChatID:        chatID,
		MessageID:     msgID,
		ReplyID:       replyID,
		UserID:        userID,
		IsFetching:    true,
	}

	go s.FetchTorrentMetadataBackground(sessionID)

	go func() {
		time.Sleep(1 * time.Hour)
		s.TaskManager.Mu.Lock()
		delete(s.TaskManager.TorrentSessions, sessionID)
		s.TaskManager.Mu.Unlock()
	}()

	return sessionID
}

func (s *BotService) HandleTorrentSelectionCallback(callback *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) < 3 {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "❌ Phiên không hợp lệ"))
		return
	}

	action := parts[1]
	sessionID := parts[2]

	s.TaskManager.Mu.Lock()
	session, exists := s.TaskManager.TorrentSessions[sessionID]
	if !exists {
		s.TaskManager.Mu.Unlock()
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "❌ Không tìm thấy phiên hoặc phiên đã hết hạn"))
		return
	}

	switch action {
	case "all":
		delete(s.TaskManager.TorrentSessions, sessionID)
		s.TaskManager.Mu.Unlock()

		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "✅ Bắt đầu tải tất cả file..."))

		_, _ = s.Bot.Request(tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID))

		confirmMsg := tgbotapi.NewMessage(callback.Message.Chat.ID, "✅ *Đã bắt đầu tải*\n\nĐang tải tất cả file trong torrent\\.\\.\\.")
		confirmMsg.ParseMode = MarkdownV2
		sentMsg, _ := s.Bot.Send(confirmMsg)
		statusMsgID := sentMsg.MessageID

		fileName := session.FileName
		if fileName == "" {
			fileName = utils.GetFileNameFromURL(session.URL)
			if fileName == "unknown_file" {
				fileName = "torrent_download"
			}
		}

		task, err := s.TaskManager.CreateTask(TypeTorrent, session.URL, fileName, session.ChatID, statusMsgID, session.ReplyID, session.UserID, session.Zip, session.Unzip, session.Password, "", 0, "", false)
		if err != nil {
			s.HandleCreateTaskError(session.ChatID, statusMsgID, err)
			return
		}
		s.UpdateSharedDashboard(session.ChatID, true)
		s.HandleAutoDelete(task)
		slog.Info("Torrent task created (all files)", "taskID", task.ID, "url", session.URL)

	case "browse", "page":
		offset := 0
		if len(parts) >= 4 {
			_, _ = fmt.Sscanf(parts[3], "%d", &offset)
		}

		if session.IsFetching {
			s.TaskManager.Mu.Unlock()
			_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "⏳ Vẫn đang lấy metadata..."))
			return
		}

		if len(session.Files) == 0 {
			s.TaskManager.Mu.Unlock()
			_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "❌ Không lấy được danh sách file"))
			return
		}

		s.TaskManager.Mu.Unlock()
		s.ShowTorrentBrowseMenu(callback, sessionID, offset)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	case "toggle":
		if len(parts) < 5 {
			s.TaskManager.Mu.Unlock()
			return
		}

		fileIdx := 0
		offset := 0
		_, _ = fmt.Sscanf(parts[3], "%d", &fileIdx)
		_, _ = fmt.Sscanf(parts[4], "%d", &offset)

		found := false
		for i, idx := range session.SelectedFiles {
			if idx == fileIdx {
				session.SelectedFiles = append(session.SelectedFiles[:i], session.SelectedFiles[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			session.SelectedFiles = append(session.SelectedFiles, fileIdx)
		}

		s.TaskManager.Mu.Unlock()
		s.ShowTorrentBrowseMenu(callback, sessionID, offset)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "🔄 Đã cập nhật lựa chọn"))

	case "start_sel":
		selected := session.SelectedFiles
		if len(selected) == 0 {
			s.TaskManager.Mu.Unlock()
			_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "⚠️ Hãy chọn ít nhất một file!"))
			return
		}
		s.TaskManager.Mu.Unlock()
		_ = s.StartTorrentWithSelectedFiles(sessionID, selected)

	case "back":
		s.TaskManager.Mu.Unlock()

		lang := s.GetUserLanguage(callback.From.ID)
		text := i18n.T(lang, "torrent_menu_text")

		edit := tgbotapi.NewEditMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, text, s.torrentMenuKeyboard(lang, sessionID))
		edit.ParseMode = MarkdownV2
		_, _ = s.Bot.Send(edit)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	case "cancel":
		delete(s.TaskManager.TorrentSessions, sessionID)
		s.TaskManager.Mu.Unlock()

		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "❌ Đã hủy"))

		deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
		_, _ = s.Bot.Request(deleteMsg)

		slog.Info("Torrent session cancelled", "sessionID", sessionID)

	default:
		s.TaskManager.Mu.Unlock()
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "❌ Thao tác không hợp lệ"))
	}
}

func (s *BotService) ShowTorrentBrowseMenu(callback *tgbotapi.CallbackQuery, sessionID string, offset int) {
	s.TaskManager.Mu.RLock()
	session, exists := s.TaskManager.TorrentSessions[sessionID]
	if !exists {
		s.TaskManager.Mu.RUnlock()
		return
	}

	files := session.Files
	selected := append([]int(nil), session.SelectedFiles...)
	s.TaskManager.Mu.RUnlock()

	if offset < 0 || offset >= len(files) {
		offset = 0
	}

	lang := s.GetUserLanguage(callback.From.ID)
	text := i18n.T(lang, "torrent_browse_title", len(files))

	edit := tgbotapi.NewEditMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, text,
		torrentBrowseKeyboard(lang, sessionID, files, selected, offset))
	edit.ParseMode = MarkdownV2

	_, _ = s.Bot.Request(edit)
}

func (s *BotService) StartTorrentWithSelectedFiles(sessionID string, selectedFiles []int) error {
	s.TaskManager.Mu.Lock()
	session, exists := s.TaskManager.TorrentSessions[sessionID]
	if !exists {
		s.TaskManager.Mu.Unlock()
		return fmt.Errorf("%w: session not found", domain.ErrNotFound)
	}

	delete(s.TaskManager.TorrentSessions, sessionID)
	s.TaskManager.Mu.Unlock()

	url := session.URL
	if len(selectedFiles) > 0 {
		selectFileStr := ""
		for i, idx := range selectedFiles {
			if i > 0 {
				selectFileStr += ","
			}
			selectFileStr += fmt.Sprintf("%d", idx)
		}
		url = url + "#select=" + selectFileStr
	}

	fileName := session.FileName
	if fileName == "" {
		fileName = utils.GetFileNameFromURL(session.URL)
		if fileName == "unknown_file" {
			fileName = "torrent_download"
		}
	}

	_, _ = s.Bot.Request(tgbotapi.NewDeleteMessage(session.ChatID, session.MessageID))

	confirmText := fmt.Sprintf("✅ *Đã bắt đầu tải*\n\nĐang tải %d file đã chọn\\.\\.\\.", len(selectedFiles))
	msg := tgbotapi.NewMessage(session.ChatID, confirmText)
	msg.ParseMode = MarkdownV2
	sentMsg, _ := s.Bot.Send(msg)

	task, err := s.TaskManager.CreateTask(TypeTorrent, url, fileName, session.ChatID, sentMsg.MessageID, session.ReplyID, session.UserID, session.Zip, session.Unzip, session.Password, "", 0, "", false)
	if err != nil {
		s.HandleCreateTaskError(session.ChatID, sentMsg.MessageID, err)
		return nil
	}
	s.UpdateSharedDashboard(session.ChatID, true)
	s.HandleAutoDelete(task)
	slog.Info("Torrent task created (selected files)", "taskID", task.ID, "selectedFiles", selectedFiles)

	return nil
}
