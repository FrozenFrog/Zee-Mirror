package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"zee-mirror/internal/config"
	"zee-mirror/internal/domain"
)

func TestTorrentMenuKeyboard(t *testing.T) {
	tests := []struct {
		dashboard string
		wantWeb   string
	}{
		{"https://bot.example/", "https://bot.example/torrent-select/sess"},
		{"127.0.0.1", ""},
		{"http://localhost:5248", ""},
	}
	for _, tt := range tests {
		bs := &BotService{Config: &config.Config{DashboardURL: tt.dashboard}}
		kb := bs.torrentMenuKeyboard("vi", "sess")

		var data []string
		webURL := ""
		for _, row := range kb.InlineKeyboard {
			for _, b := range row {
				if b.URL != nil {
					webURL = *b.URL
				}
				if b.CallbackData != nil {
					data = append(data, *b.CallbackData)
				}
			}
		}
		assert.Equal(t, tt.wantWeb, webURL, tt.dashboard)
		assert.Equal(t, []string{"torrent_sel:all:sess", "torrent_sel:browse:sess:0", "torrent_sel:cancel:sess"}, data)
	}
}

func TestTorrentBrowseKeyboard(t *testing.T) {
	var files []domain.TorrentFile
	for i := 1; i <= 10; i++ {
		files = append(files, domain.TorrentFile{Name: fmt.Sprintf("f%d.iso", i), Index: i, Size: 1024})
	}
	files[0].Name = strings.Repeat("ệ", 60)

	kb := torrentBrowseKeyboard("vi", "sess", files, []int{2}, 0)
	rows := kb.InlineKeyboard
	// 8 files, nav row (next only), start, back
	assert.Len(t, rows, torrentBrowsePageSize+3)
	assert.Equal(t, "torrent_sel:toggle:sess:1:0", *rows[0][0].CallbackData)
	assert.True(t, strings.HasPrefix(rows[1][0].Text, "✅"))
	assert.True(t, strings.HasPrefix(rows[2][0].Text, "⬜"))
	assert.Contains(t, rows[0][0].Text, "…")
	assert.Len(t, rows[8], 1)
	assert.Equal(t, "torrent_sel:page:sess:8", *rows[8][0].CallbackData)
	assert.Equal(t, "torrent_sel:start_sel:sess", *rows[9][0].CallbackData)
	assert.Equal(t, "torrent_sel:back:sess", *rows[10][0].CallbackData)

	for _, row := range rows {
		for _, b := range row {
			assert.LessOrEqual(t, len(*b.CallbackData), 64, "Telegram callback_data limit")
		}
	}

	last := torrentBrowseKeyboard("vi", "sess", files, nil, 8).InlineKeyboard
	assert.Len(t, last, 2+3)
	assert.Equal(t, "torrent_sel:page:sess:0", *last[2][0].CallbackData)
}
