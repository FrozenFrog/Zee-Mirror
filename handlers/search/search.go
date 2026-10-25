package search

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"zee-mirror/internal/service"
	"zee-mirror/pkg/utils"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type Result struct {
	Title   string
	Size    string
	Seeders string
	Magnet  string
	Source  string
}

type Session struct {
	CreatedAt time.Time
	Query     string
	Provider  string
	Results   []Result
	Page      int
}

var (
	SearchSessions = make(map[string]*Session)
	SearchMu       sync.RWMutex
)

const (
	searchSessionCleanupInterval = 30 * time.Minute
	searchSessionMaxAge          = 1 * time.Hour
)

func StartSearchSessionCleanup(shutdown <-chan struct{}) {
	ticker := time.NewTicker(searchSessionCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-shutdown:
			return
		case <-ticker.C:
			cleanupSearchSessions()
		}
	}
}

func cleanupSearchSessions() {
	cutoff := time.Now().Add(-searchSessionMaxAge)

	SearchMu.Lock()
	var toRemove []string
	for id, session := range SearchSessions {
		if session.CreatedAt.Before(cutoff) {
			toRemove = append(toRemove, id)
		}
	}
	for _, id := range toRemove {
		delete(SearchSessions, id)
	}
	SearchMu.Unlock()

	if len(toRemove) > 0 {
		slog.Debug("Search session cleanup completed", "removed", len(toRemove))
	}
}

func HandleSearch(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.CanUse(message) {
		return
	}

	if args == "" {
		s.Reply(message, "⚠️ *Sai định dạng*\n\nCách dùng: `/search <từ khóa>` hoặc `/searchall <từ khóa>`")
		return
	}

	query := args
	isSearchAll := message.Command() == "searchall"

	if !isSearchAll {
		sessionID := uuid.New().String()[:12]
		SearchMu.Lock()
		SearchSessions[sessionID] = &Session{CreatedAt: time.Now(), Query: query}
		SearchMu.Unlock()

		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("🔍 *Tìm kiếm:* `%s`\n\nChọn nguồn Torrent để bắt đầu tìm kiếm:", utils.EscapeMarkdownV2(query)))
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		msg.ReplyMarkup = providerKeyboard(sessionID)
		_, _ = s.Bot.Send(msg)
		return
	}

	sessionID := uuid.New().String()

	SearchMu.Lock()
	SearchSessions[sessionID] = &Session{
		CreatedAt: time.Now(),
		Query:     query,
		Provider:  "All",
		Results:   []Result{},
		Page:      0,
	}
	SearchMu.Unlock()

	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("🔍 *Đang tìm* `%s` trên tất cả các nguồn\\.\\.\\.", utils.EscapeMarkdownV2(query)))
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	sent, err := s.Bot.Send(msg)
	if err != nil {
		slog.Error("Failed to send initial search message", "error", err)
		return
	}

	go func() {
		nyResults := scrapeNyaa(query)
		pbResults := searchPirateBay(query)
		slResults := searchSolidTorrents(query)

		var all []Result
		all = append(all, nyResults...)
		all = append(all, pbResults...)
		all = append(all, slResults...)

		SortResultsBySeeders(all)

		SearchMu.Lock()
		if sess, ok := SearchSessions[sessionID]; ok {
			sess.Results = all
		}
		SearchMu.Unlock()

		showSearchResults(s, sent.Chat.ID, sent.MessageID, sessionID)
	}()
}

func SortResultsBySeeders(results []Result) {
	sort.Slice(results, func(i, j int) bool {
		s1 := parseSeeders(results[i].Seeders)
		s2 := parseSeeders(results[j].Seeders)
		return s1 > s2
	})
}

func parseSeeders(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	val, _ := strconv.Atoi(s)
	return val
}

func HandleSearchCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) < 3 {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "❌ Từ khóa không hợp lệ"))
		return
	}

	provider := parts[1]
	query := strings.Join(parts[2:], ":")
	if sid, ok := strings.CutPrefix(query, sessionQueryPrefix); ok {
		SearchMu.RLock()
		session, exists := SearchSessions[sid]
		if exists {
			query = session.Query
		}
		SearchMu.RUnlock()
		if !exists {
			_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "❌ Phiên đã hết hạn"))
			return
		}
	}

	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "🔍 Đang tìm kiếm..."))

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, fmt.Sprintf("🔍 Đang tìm `%s` • *%s*\\.\\.\\.", utils.EscapeMarkdownV2(query), provider))
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, _ = s.Bot.Send(editMsg)

	var results []Result
	slog.Info("Torrent search initiated", "provider", provider, "query", query)
	switch provider {
	case "solid":
		results = searchSolidTorrents(query)
	case "nyaa":
		results = scrapeNyaa(query)
	case "apibay":
		results = searchPirateBay(query)
	case "all":
		nyResults := scrapeNyaa(query)
		pbResults := searchPirateBay(query)
		slResults := searchSolidTorrents(query)
		results = append(results, nyResults...)
		results = append(results, pbResults...)
		results = append(results, slResults...)
		provider = "All"
	}

	SortResultsBySeeders(results)

	if len(results) == 0 {
		sessionID := uuid.New().String()[:12]
		session := &Session{
			Query:     query,
			Provider:  provider,
			Results:   results,
			Page:      0,
			CreatedAt: time.Now(),
		}
		SearchMu.Lock()
		SearchSessions[sessionID] = session
		SearchMu.Unlock()

		text := fmt.Sprintf("📭 *Không tìm thấy kết quả nào trên %s*\n\n_Hãy thử từ khóa khác hoặc nguồn khác\\._", provider)

		editMsg := tgbotapi.NewEditMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, text, backCloseKeyboard(sessionID))
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
		_, _ = s.Bot.Send(editMsg)
		return
	}

	sessionID := uuid.New().String()[:12]
	session := &Session{
		Query:     query,
		Provider:  provider,
		Results:   results,
		Page:      0,
		CreatedAt: time.Now(),
	}

	SearchMu.Lock()
	SearchSessions[sessionID] = session
	SearchMu.Unlock()

	showSearchResults(s, callback.Message.Chat.ID, callback.Message.MessageID, sessionID)
}

func showSearchResults(s *service.BotService, chatID int64, messageID int, sessionID string) {
	SearchMu.Lock()
	session, exists := SearchSessions[sessionID]

	if !exists {
		SearchMu.Unlock()
		s.SendOrEditMessage(chatID, messageID, "❌ Phiên tìm kiếm đã hết hạn\\. Hãy tìm kiếm lại\\.", nil)
		return
	}

	perPage := 5
	totalItems := len(session.Results)
	totalPages := (totalItems + perPage - 1) / perPage

	if session.Page >= totalPages {
		session.Page = 0
	}
	if session.Page < 0 {
		session.Page = totalPages - 1
	}

	page := session.Page
	start := page * perPage
	end := start + perPage
	if end > totalItems {
		end = totalItems
	}

	visibleItems := make([]Result, end-start)
	copy(visibleItems, session.Results[start:end])
	provider := session.Provider
	query := session.Query
	SearchMu.Unlock()

	text := fmt.Sprintf("🔍 *Kết quả tìm kiếm \\(%s\\)*\n`%s`\n\n", utils.EscapeMarkdownV2(provider), utils.EscapeMarkdownV2Code(query))
	text += fmt.Sprintf("📄 Trang %d/%d\n\n", page+1, totalPages)

	for i, item := range visibleItems {
		idx := start + i + 1
		text += fmt.Sprintf("%d\\. *%s*\n📦 %s • 👤 %s • 🌐 %s\n\n",
			idx,
			utils.EscapeMarkdownV2(utils.TruncateString(item.Title, 40)),
			utils.EscapeMarkdownV2(item.Size),
			utils.EscapeMarkdownV2(item.Seeders),
			utils.EscapeMarkdownV2(item.Source),
		)
	}

	var numberRow []tgbotapi.InlineKeyboardButton
	for i := range visibleItems {
		numberRow = append(numberRow, tgbotapi.NewInlineKeyboardButtonData(
			strconv.Itoa(start+i+1), fmt.Sprintf("t_item:%d:%s", start+i, sessionID)))
	}
	rows := [][]tgbotapi.InlineKeyboardButton{numberRow}
	if totalPages > 1 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️", fmt.Sprintf("t_page:%d:%s", page-1, sessionID)),
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d/%d", page+1, totalPages), fmt.Sprintf("t_page:%d:%s", page, sessionID)),
			tgbotapi.NewInlineKeyboardButtonData("➡️", fmt.Sprintf("t_page:%d:%s", page+1, sessionID)),
		))
	}
	rows = append(rows, backCloseKeyboard(sessionID).InlineKeyboard...)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	editMsg.ReplyMarkup = &keyboard

	if _, err := s.Bot.Send(editMsg); err != nil {
		slog.Warn("Search results markdown error, using fallback", "error", err)
		fallbackText := fmt.Sprintf("🔍 Kết quả tìm kiếm (%s): %s\n\nTrang %d/%d\n\n", provider, query, page+1, totalPages)
		for i, item := range visibleItems {
			fallbackText += fmt.Sprintf("%d. %s\nDung lượng: %s | Seed: %s\n\n", start+i+1, item.Title, item.Size, item.Seeders)
		}
		fallbackEdit := tgbotapi.NewEditMessageText(chatID, messageID, fallbackText)
		fallbackEdit.ReplyMarkup = &keyboard
		_, _ = s.Bot.Send(fallbackEdit)
	}
}

// sessionQueryPrefix marks a t_search callback whose query lives in SearchSessions,
// keeping callback_data under Telegram's 64-byte limit for long queries.
const sessionQueryPrefix = "@"

func providerKeyboard(sessionID string) tgbotapi.InlineKeyboardMarkup {
	data := func(provider string) string {
		return "t_search:" + provider + ":" + sessionQueryPrefix + sessionID
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌟 Solid", data("solid")),
			tgbotapi.NewInlineKeyboardButtonData("🏴‍☠️ PirateBay", data("apibay")),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌸 Nyaa (Anime)", data("nyaa")),
			tgbotapi.NewInlineKeyboardButtonData("🌍 Tất cả", data("all")),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Đóng", "t_close:"+sessionID),
		),
	)
}

func backCloseKeyboard(sessionID string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 Quay lại", "t_back:"+sessionID),
		tgbotapi.NewInlineKeyboardButtonData("❌ Đóng", "t_close:"+sessionID),
	))
}

func HandleSearchNavCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, parts []string) {
	action := parts[0]

	if action == "t_back" {
		if len(parts) < 2 {
			return
		}
		sessionID := parts[1]

		SearchMu.RLock()
		session, exists := SearchSessions[sessionID]
		if !exists {
			SearchMu.RUnlock()
			_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "❌ Phiên đã hết hạn"))
			return
		}
		query := session.Query
		SearchMu.RUnlock()

		textRaw := fmt.Sprintf("🔍 Tìm kiếm: %s\n\nChọn nguồn tìm kiếm ổn định hơn:", query)
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, textRaw, providerKeyboard(sessionID))

		if _, err := s.Bot.Send(editMsg); err != nil {
			slog.Warn("t_back edit error", "error", err)
		}

		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		return
	}

	if action == "t_close" {
		sessionID := parts[1]
		SearchMu.Lock()
		delete(SearchSessions, sessionID)
		SearchMu.Unlock()
		_, _ = s.Bot.Request(tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID))
		return
	}

	if len(parts) < 3 {
		return
	}

	if action == "t_page" {
		page, _ := strconv.Atoi(parts[1])
		sessionID := parts[2]

		SearchMu.Lock()
		if session, exists := SearchSessions[sessionID]; exists {
			session.Page = page
		}
		SearchMu.Unlock()

		showSearchResults(s, callback.Message.Chat.ID, callback.Message.MessageID, sessionID)
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		return
	}

	if action == "t_item" {
		index, _ := strconv.Atoi(parts[1])
		sessionID := parts[2]

		SearchMu.RLock()
		session, exists := SearchSessions[sessionID]
		if !exists || index < 0 || index >= len(session.Results) {
			SearchMu.RUnlock()
			if !exists {
				_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, "❌ Phiên đã hết hạn"))
			}
			return
		}

		item := session.Results[index]
		cleanMagnet := CleanMagnetLink(item.Magnet)
		title := item.Title
		size := item.Size
		seeders := item.Seeders
		source := item.Source
		SearchMu.RUnlock()

		text := fmt.Sprintf("📄 *Chi tiết Torrent*\n\n"+
			"*Tiêu đề:* `%s`\n"+
			"*Dung lượng:* `%s`\n"+
			"*Seeders:* `%s`\n"+
			"*Nguồn:* `%s`\n\n"+
			"🧲 *Link Magnet:*\n`%s`\n\n"+
			"Nhấn vào link trên để sao chép hoặc reply bằng /mirror",
			utils.EscapeMarkdownV2Code(title),
			utils.EscapeMarkdownV2Code(size),
			utils.EscapeMarkdownV2Code(seeders),
			utils.EscapeMarkdownV2Code(source),
			utils.EscapeMarkdownV2Code(cleanMagnet),
		)

		page := 0
		if index >= 0 {
			page = index / 5
		}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Quay lại", fmt.Sprintf("t_page:%d:%s", page, sessionID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Đóng", "t_close:"+sessionID),
		))

		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
		editMsg.ReplyMarkup = &keyboard

		if _, err := s.Bot.Send(editMsg); err != nil {
			slog.Warn("Search item detail markdown error, using fallback", "error", err)
			fallbackText := fmt.Sprintf("Chi tiết Torrent:\n\n%s\nDung lượng: %s\nMagnet:\n%s", title, size, cleanMagnet)
			fallbackEdit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, fallbackText)
			fallbackEdit.ReplyMarkup = &keyboard
			_, _ = s.Bot.Send(fallbackEdit)
		}
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
	}
}

func searchSolidTorrents(query string) []Result {
	var results []Result
	apiURL := fmt.Sprintf("https://solidtorrents.to/api/v1/search?q=%s&sort=seeders", url.QueryEscape(query))
	slog.Debug("Requesting SolidTorrents", "url", apiURL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		slog.Error("SolidTorrents request failed", "error", err)
		return results
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("Error closing SolidTorrents response body", "error", err)
		}
	}()

	slog.Debug("SolidTorrents response", "status", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return results
	}

	var data struct {
		Results []struct {
			Title    string `json:"title"`
			Magnet   string `json:"magnet"`
			InfoHash string `json:"infohash"`
			Size     int64  `json:"size"`
			Seeders  int    `json:"seeders"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		slog.Error("SolidTorrents decode error", "error", err)
		return results
	}

	seen := make(map[string]bool)

	for _, item := range data.Results {
		if item.Size < 10*1024*1024 {
			continue
		}

		cleanTitle := strings.ReplaceAll(item.Title, ".", " ")
		cleanTitle = strings.TrimSpace(cleanTitle)

		if seen[cleanTitle] {
			continue
		}
		seen[cleanTitle] = true

		magnet := item.Magnet
		if magnet == "" && item.InfoHash != "" {
			magnet = fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s", item.InfoHash, url.QueryEscape(item.Title))
		}

		results = append(results, Result{
			Title:   cleanTitle,
			Size:    utils.FormatBytes(item.Size),
			Seeders: fmt.Sprintf("%d", item.Seeders),
			Magnet:  magnet,
			Source:  "Solid",
		})
	}
	return results
}

func scrapeNyaa(query string) []Result {
	var results []Result
	searchURL := fmt.Sprintf("https://nyaa.si/?f=0&c=0_0&q=%s", url.QueryEscape(query))
	slog.Debug("Requesting Nyaa", "url", searchURL)

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Nyaa request failed", "error", err)
		return results
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			slog.Warn("Error closing Nyaa response body", "error", errClose)
		}
	}()

	slog.Debug("Nyaa response", "status", resp.StatusCode)

	if resp.StatusCode != 200 {
		return results
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		slog.Error("Nyaa parse error", "error", err)
		return results
	}

	doc.Find("table.torrent-list tbody tr").Each(func(_ int, s *goquery.Selection) {
		if len(results) >= 20 {
			return
		}

		title := s.Find("td[colspan='2'] a").Last().AttrOr("title", "")
		if title == "" {
			title = s.Find("td[colspan='2'] a").Last().Text()
		}

		magnet, _ := s.Find("td.text-center a[href^='magnet:']").Attr("href")
		size := s.Find("td.text-center").Eq(1).Text()
		seeders := s.Find("td.text-center").Eq(3).Text()

		if title != "" && magnet != "" {
			results = append(results, Result{
				Title:   title,
				Size:    size,
				Seeders: seeders,
				Magnet:  magnet,
				Source:  "Nyaa",
			})
		}
	})

	return results
}

func searchPirateBay(query string) []Result {
	var results []Result
	apiURL := fmt.Sprintf("https://apibay.org/q.php?q=%s&cat=0", url.QueryEscape(query))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return results
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("Error closing PirateBay response body", "error", err)
		}
	}()

	var data []struct {
		Name     string `json:"name"`
		InfoHash string `json:"info_hash"`
		Seeders  string `json:"seeders"`
		Size     string `json:"size"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return results
	}

	for _, item := range data {
		if len(results) >= 20 {
			break
		}
		if item.InfoHash == "0000000000000000000000000000000000000000" {
			continue
		}

		sizeInt, _ := strconv.ParseInt(item.Size, 10, 64)

		magnet := fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s", item.InfoHash, url.QueryEscape(item.Name))

		results = append(results, Result{
			Title:   item.Name,
			Size:    utils.FormatBytes(sizeInt),
			Seeders: item.Seeders,
			Magnet:  magnet,
			Source:  "TPB",
		})
	}
	return results
}

func CleanMagnetLink(magnet string) string {
	if magnet == "" {
		return "Không có Magnet"
	}
	if strings.HasPrefix(magnet, "magnet:?") {
		return magnet
	}
	if len(magnet) >= 32 {
		return fmt.Sprintf("magnet:?xt=urn:btih:%s", magnet)
	}
	return magnet
}
