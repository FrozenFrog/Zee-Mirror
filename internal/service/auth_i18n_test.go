package service_test

import (
	"database/sql"
	"testing"
	"time"

	"zee-mirror/internal/config"
	"zee-mirror/internal/domain"
	"zee-mirror/internal/errors"
	"zee-mirror/internal/repository/mocks"
	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckQuotaErrorsAreTranslated(t *testing.T) {
	cfg := &config.Config{OwnerID: 12345}

	t.Run("task limit", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		svc := service.NewAuthService(cfg, mockRepo, mockRepo)
		mockRepo.On("GetByID", mock.Anything, int64(666)).Return(&domain.User{
			ID: 666, IsActive: true, MaxDailyTasks: 5, MaxDailyBandwidth: -1,
		}, nil).Once()
		mockRepo.On("GetUserTodayStats", mock.Anything, int64(666)).Return(&domain.DailyStats{TotalTasks: 5}, nil).Once()

		err := svc.CheckQuota(666)
		assert.ErrorIs(t, err, errors.ErrQuotaExceeded)
		assert.Equal(t, "Đã đạt giới hạn task mỗi ngày (5/5)", i18n.Message("vi", err))
		assert.Equal(t, "Daily task limit reached (5/5)", i18n.Message("en", err))
		assert.Equal(t, "Daily task limit reached (5/5)", err.Error())
	})

	t.Run("expired account", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		svc := service.NewAuthService(cfg, mockRepo, mockRepo)
		expired := time.Date(2024, time.March, 5, 0, 0, 0, 0, time.UTC)
		mockRepo.On("GetByID", mock.Anything, int64(888)).Return(&domain.User{
			ID: 888, ExpiresAt: sql.NullTime{Time: expired, Valid: true},
		}, nil).Once()

		err := svc.CheckQuota(888)
		assert.ErrorIs(t, err, errors.ErrAccountInactive)
		assert.Equal(t, "Your access expired on 05 Mar 2024", i18n.Message("en", err))
		assert.Contains(t, i18n.Message("vi", err), "05 Mar 2024")
	})

	t.Run("unknown user", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		svc := service.NewAuthService(cfg, mockRepo, mockRepo)
		mockRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, errors.ErrUserNotFound).Once()

		err := svc.CheckQuota(999)
		assert.ErrorIs(t, err, errors.ErrUserNotFound)
		assert.Equal(t, "Không tìm thấy người dùng", i18n.Message("vi", err))
	})
}

func TestQuotaErrorMessageEscapesMarkdown(t *testing.T) {
	err := i18n.WrapError(errors.ErrQuotaExceeded, "auth_quota_tasks", 5, 5)
	assert.Equal(t, "⚠️ *Vượt hạn mức*\n\nĐã đạt giới hạn task mỗi ngày \\(5/5\\)", service.QuotaErrorMessage("vi", err))
	assert.Contains(t, service.QuotaErrorMessage("en", err), "Daily task limit reached \\(5/5\\)")
}

func TestSenderLanguage(t *testing.T) {
	assert.Equal(t, "en", service.SenderLanguage(&tgbotapi.User{LanguageCode: "en-US"}))
	assert.Equal(t, "zh", service.SenderLanguage(&tgbotapi.User{LanguageCode: "zh-hans"}))
	assert.Equal(t, i18n.DefaultLang, service.SenderLanguage(&tgbotapi.User{LanguageCode: "pt-br"}))
	assert.Equal(t, i18n.DefaultLang, service.SenderLanguage(&tgbotapi.User{}))
	assert.Equal(t, i18n.DefaultLang, service.SenderLanguage(nil))
}

func TestAccessDeniedText(t *testing.T) {
	assert.Equal(t,
		"❌ *TRUY CẬP BỊ TỪ CHỐI*\n\nBạn chưa được cấp quyền sử dụng bot này\\.\nVui lòng liên hệ Owner để được cấp quyền truy cập\\.",
		service.AccessDeniedMessage("vi"))
	assert.Equal(t, "🚫 Truy cập bị từ chối", service.AccessDeniedToast("vi"))
	assert.Equal(t, "🚫 Access denied", service.AccessDeniedToast("en"))
	for _, l := range i18n.Languages {
		assert.NotEqual(t, "auth_toast_denied", service.AccessDeniedToast(l.Code), l.Code)
		assert.NotContains(t, service.AccessDeniedMessage(l.Code), "auth_", l.Code)
	}
}
