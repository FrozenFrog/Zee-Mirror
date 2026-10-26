package service

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"zee-mirror/internal/config"
	"zee-mirror/internal/domain"
	"zee-mirror/internal/repository/mocks"
	"zee-mirror/pkg/i18n"
)

func TestLanguageKeyboardListsLanguages(t *testing.T) {
	var got []string
	for _, row := range languageKeyboard().InlineKeyboard {
		assert.LessOrEqual(t, len(row), 2)
		for _, b := range row {
			got = append(got, *b.CallbackData)
		}
	}
	var want []string
	for _, l := range i18n.Languages {
		want = append(want, "settings:lang_"+l.Code)
	}
	assert.Equal(t, want, got)

	for _, l := range i18n.Languages {
		action := strings.TrimPrefix("settings:lang_"+l.Code, "settings:")
		found, ok := languageFromAction(action)
		assert.True(t, ok, action)
		assert.Equal(t, l, found)
	}
	_, ok := languageFromAction("lang_xx")
	assert.False(t, ok)
	_, ok = languageFromAction("quality_720")
	assert.False(t, ok)
}

func TestLanguageSetToastUsesNewLanguage(t *testing.T) {
	en, _ := languageFromAction("lang_en")
	assert.Equal(t, "🇺🇸 Language: English", languageSetToast(en))
	vi, _ := languageFromAction("lang_vi")
	assert.Equal(t, "🇻🇳 Ngôn ngữ: Tiếng Việt", languageSetToast(vi))
	assert.Equal(t, i18n.DefaultLang, languageInfo("xx").Code)
}

func TestInitialLanguage(t *testing.T) {
	assert.Equal(t, "en", initialLanguage("en-US"))
	assert.Equal(t, "zh", initialLanguage("zh-hans"))
	assert.Equal(t, i18n.DefaultLang, initialLanguage("pt-br"))
	assert.Equal(t, i18n.DefaultLang, initialLanguage(""))
}

func TestSyncUserLanguage(t *testing.T) {
	tests := []struct {
		name     string
		existing *domain.User
		code     string
		want     string
	}{
		{"new user takes Telegram language", nil, "ja", "ja"},
		{"new user with unsupported language", nil, "pt-br", i18n.DefaultLang},
		{"stored choice is kept", &domain.User{Language: "id"}, "en", "id"},
		{"seeded user without language", &domain.User{Role: "owner"}, "en-GB", "en"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.MockRepository)
			if tt.existing != nil {
				repo.On("GetByID", mock.Anything, int64(7)).Return(tt.existing, nil)
			} else {
				repo.On("GetByID", mock.Anything, int64(7)).Return(nil, domain.ErrNotFound)
			}
			var saved domain.User
			repo.On("Upsert", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				saved = args.Get(1).(domain.User)
			}).Return(nil)

			cfg := &config.Config{}
			s := &BotService{DB: repo, Config: cfg, Auth: NewAuthService(cfg, repo, repo)}
			s.SyncUser(&tgbotapi.User{ID: 7, LanguageCode: tt.code})
			assert.Equal(t, tt.want, saved.Language)
		})
	}
}

func TestStatusLabelAndKeyOrText(t *testing.T) {
	assert.Equal(t, "Queued", StatusLabel("en", StatusQueued))
	assert.Equal(t, "Đang nén", StatusLabel("vi", domain.StatusZipping))
	assert.Equal(t, "Weird", StatusLabel("en", "weird"))

	assert.Equal(t, "Zipping", keyOrText("en", "task_state_zipping"))
	assert.Equal(t, "Đang ghép video/audio...", keyOrText("en", "Đang ghép video/audio..."))
	assert.Equal(t, "", keyOrText("en", ""))
}

func TestTaskFailureUsesOwnerLanguage(t *testing.T) {
	repo := new(mocks.MockRepository)
	repo.On("GetByID", mock.Anything, int64(42)).Return(&domain.User{Language: "en"}, nil)
	s := &BotService{DB: repo}

	task := newTestTask("t1", "g1", StatusDownloading, time.Now())
	task.UserID = 42
	cause := fmt.Errorf("plugin: %w", i18n.NewError("task_err_no_file"))
	s.coreFailTask(task, "task_err_upload", cause)

	assert.Equal(t, StatusFailed, task.Status)
	assert.Equal(t, "Upload failed: Downloaded file not found", task.Error)

	text := buildTaskStatusText("vi", task.GetSnapshot(), "")
	assert.Contains(t, text, "Lỗi: `Upload failed")

	// Plain errors keep their text.
	task2 := newTestTask("t2", "g2", StatusDownloading, time.Now())
	task2.UserID = 42
	s.coreFailTask(task2, "", errors.New("boom"))
	assert.Equal(t, "boom", task2.Error)
}

func TestCategorizeErrorKeepsTranslatableCause(t *testing.T) {
	err := domain.CategorizeError(fmt.Errorf("connection reset: %w", i18n.NewError("task_err_tg_timeout", "abc")))
	assert.Equal(t, "hết thời gian chờ lấy file Telegram: abc", i18n.Message("vi", err))
	assert.True(t, errors.Is(err, domain.ErrNetwork))
}

func TestGetErrorMessageLang(t *testing.T) {
	msg := GetErrorMessageLang("en", "T", i18n.NewError("task_err_rate_limit"))
	assert.Equal(t, "❌ *T*\n\nRate limit exceeded: 5 tasks/minute", msg)
	assert.Contains(t, GetErrorMessage("T", i18n.NewError("task_err_rate_limit")), "Vượt giới hạn")
	assert.Equal(t, "📉 You have used up today's quota.", domain.GetUserMessageLang("en", domain.ErrQuota))
}

func TestStatusHeaderLang(t *testing.T) {
	assert.Contains(t, GetStatusHeaderLang("en"), "ACTIVE TASKS")
	assert.Equal(t, GetStatusHeaderLang(i18n.DefaultLang), GetStatusHeader())
}
