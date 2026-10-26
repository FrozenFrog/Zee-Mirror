package service

import (
	"context"
	"slices"
	"strings"

	"zee-mirror/internal/config"
	"zee-mirror/internal/errors"
	"zee-mirror/internal/forum"
	"zee-mirror/internal/repository"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Authorizer is the single seam for access policy: bot commands, queue
// priority, and dedup bypass all go through this interface.
type Authorizer interface {
	IsAuthorized(userID int64) bool
	IsAdmin(userID int64) bool
	IsOwner(userID int64) bool
	// IsPrivileged preserves the legacy env-based privilege (owner or
	// AUTHORIZED_USERS) used for queue priority and duplicate bypass.
	IsPrivileged(userID int64) bool
}

const (
	RoleAdmin      = "admin"
	RoleAuthorized = "authorized"
	RoleOwner      = "owner"
)

type AuthService struct {
	Config   *config.Config
	UserRepo repository.UserRepository
	DB       repository.FullRepository
}

func NewAuthService(cfg *config.Config, userRepo repository.UserRepository, db repository.FullRepository) *AuthService {
	return &AuthService{
		Config:   cfg,
		UserRepo: userRepo,
		DB:       db,
	}
}

func (s *AuthService) IsAuthorized(userID int64) bool {
	if userID == s.Config.OwnerID {
		return true
	}

	ctx := context.Background()
	user, err := s.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return false
	}

	if !user.IsActive {
		return false
	}

	return user.Role == RoleAdmin || user.Role == RoleAuthorized || user.Role == RoleOwner
}

func (s *AuthService) IsOwner(userID int64) bool {
	return userID == s.Config.OwnerID
}

func (s *AuthService) IsPrivileged(userID int64) bool {
	return isEnvPrivileged(s.Config, userID)
}

func isEnvPrivileged(cfg *config.Config, userID int64) bool {
	if cfg == nil {
		return false
	}
	return userID == cfg.OwnerID || slices.Contains(cfg.AuthorizedUsers, userID)
}

func (s *AuthService) IsAdmin(userID int64) bool {
	if userID == s.Config.OwnerID {
		return true
	}

	ctx := context.Background()
	user, err := s.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return false
	}

	if !user.IsActive {
		return false
	}

	return user.Role == RoleAdmin || user.Role == RoleOwner
}

// CheckQuota returns nil when userID may start another task. Failures are
// *i18n.Error values wrapping the sentinel (errors.ErrUserNotFound,
// ErrAccountInactive, ErrQuotaExceeded), so errors.Is still matches while
// i18n.Message(lang, err) renders them in the user's language; err.Error()
// is English. Display them with QuotaErrorMessage.
func (s *AuthService) CheckQuota(userID int64) error {
	if s.IsOwner(userID) {
		return nil
	}

	ctx := context.Background()
	user, err := s.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return i18n.WrapError(errors.ErrUserNotFound, "auth_user_not_found")
	}

	if !user.IsActive {
		if user.ExpiresAt.Valid {
			return i18n.WrapError(errors.ErrAccountInactive, "auth_account_expired", user.ExpiresAt.Time.Format("02 Jan 2006"))
		}
		return i18n.WrapError(errors.ErrAccountInactive, "auth_account_inactive")
	}

	if user.MaxDailyTasks == -1 && user.MaxDailyBandwidth == -1 {
		return nil
	}

	stats, err := s.DB.GetUserTodayStats(ctx, userID)
	if err != nil {
		return nil
	}

	if user.MaxDailyTasks != -1 && stats.TotalTasks >= user.MaxDailyTasks {
		return i18n.WrapError(errors.ErrQuotaExceeded, "auth_quota_tasks", stats.TotalTasks, user.MaxDailyTasks)
	}

	if user.MaxDailyBandwidth != -1 && stats.TotalBandwidth >= user.MaxDailyBandwidth {
		return i18n.WrapError(errors.ErrQuotaExceeded, "auth_quota_bandwidth", utils.FormatBytes(stats.TotalBandwidth), utils.FormatBytes(user.MaxDailyBandwidth))
	}

	return nil
}

// QuotaErrorMessage renders an error from CheckQuota as a MarkdownV2 message
// in lang.
func QuotaErrorMessage(lang string, err error) string {
	return i18n.T(lang, "quota_exceeded", utils.EscapeMarkdownV2(i18n.Message(lang, err)))
}

// SenderLanguage is the bot language matching a Telegram client's language
// code, or i18n.DefaultLang. Use it for people who may not be in the
// database (e.g. denied users), where GetUserLanguage has nothing to read.
func SenderLanguage(u *tgbotapi.User) string {
	if u != nil {
		if lang := i18n.Normalize(u.LanguageCode); lang != "" {
			return lang
		}
	}
	return i18n.DefaultLang
}

// AccessDeniedMessage is the MarkdownV2 reply for a user who may not use the bot.
func AccessDeniedMessage(lang string) string {
	return GetErrorMessage(i18n.T(lang, "auth_title_access_denied"), i18n.T(lang, "auth_denied_body"))
}

// AccessDeniedToast is the plain-text callback answer for the same case.
func AccessDeniedToast(lang string) string {
	return i18n.T(lang, "auth_toast_denied")
}

func (s *BotService) IsAuthorized(userID int64) bool {
	return s.Auth.IsAuthorized(userID)
}

// Access is the verdict for an incoming update.
type Access int

const (
	// AccessIgnore means the message isn't addressed to this bot.
	AccessIgnore Access = iota
	AccessDenied
	AccessGranted
)

type chatScope int

const (
	scopeNone  chatScope = iota
	scopeChat            // listed as "chat": the bot must be addressed with @username
	scopeTopic           // listed as "chat:topic": every message is for the bot
)

func (s *BotService) chatScope(m *tgbotapi.Message) chatScope {
	if m == nil {
		return scopeNone
	}
	topic := forum.TopicOf(m)
	scope := scopeNone
	for _, c := range s.Config.AuthorizedChats {
		if c.ID != m.Chat.ID {
			continue
		}
		switch c.TopicID {
		case 0:
			scope = scopeChat
		case topic:
			return scopeTopic
		}
	}
	return scope
}

// AdmitMessage applies AUTHORIZED_CHATS to an incoming message. Everyone in a
// listed chat may use the bot: directly inside a listed topic, and by
// addressing it with @username anywhere else in the chat (the mention is
// stripped from plain text so handlers see the bare input). Other chats fall
// back to per-user authorization.
func (s *BotService) AdmitMessage(m *tgbotapi.Message) Access {
	s.Topics.Remember(m)

	if cmd := m.CommandWithAt(); strings.Contains(cmd, "@") && !s.mentionsMe(cmd) {
		return AccessIgnore
	}

	switch s.chatScope(m) {
	case scopeTopic:
		return AccessGranted
	case scopeChat:
		if !s.addressedToMe(m) {
			return AccessIgnore
		}
		return AccessGranted
	}

	if (m.From != nil && s.IsAuthorized(m.From.ID)) || (m.IsCommand() && m.Command() == "start") {
		return AccessGranted
	}
	return AccessDenied
}

// AdmitCallback reports whether a button press may be handled. Buttons need
// no mention, so any member of a listed chat may press them.
func (s *BotService) AdmitCallback(cb *tgbotapi.CallbackQuery) bool {
	s.Topics.Remember(cb.Message)
	return s.chatScope(cb.Message) != scopeNone || s.IsAuthorized(cb.From.ID)
}

// CanUse reports whether the sender of m may run regular (non-admin) commands.
func (s *BotService) CanUse(m *tgbotapi.Message) bool {
	return s.chatScope(m) != scopeNone || (m.From != nil && s.IsAuthorized(m.From.ID))
}

func (s *BotService) mentionsMe(text string) bool {
	return strings.Contains(strings.ToLower(text), "@"+strings.ToLower(s.Bot.Self.UserName))
}

func (s *BotService) addressedToMe(m *tgbotapi.Message) bool {
	if m.IsCommand() {
		return strings.Contains(m.CommandWithAt(), "@")
	}
	if m.ReplyToMessage != nil && m.ReplyToMessage.From != nil && m.ReplyToMessage.From.ID == s.Bot.Self.ID {
		return true
	}
	if m.Text != "" && s.mentionsMe(m.Text) {
		m.Text = stripMention(m.Text, s.Bot.Self.UserName)
		return true
	}
	if m.Caption != "" && s.mentionsMe(m.Caption) {
		m.Caption = stripMention(m.Caption, s.Bot.Self.UserName)
		return true
	}
	return false
}

func stripMention(text, username string) string {
	i := strings.Index(strings.ToLower(text), "@"+strings.ToLower(username))
	return strings.TrimSpace(text[:i] + text[i+len(username)+1:])
}

func (s *BotService) IsOwner(userID int64) bool {
	return s.Auth.IsOwner(userID)
}

func (s *BotService) IsAdmin(userID int64) bool {
	return s.Auth.IsAdmin(userID)
}

func (s *BotService) CheckQuota(userID int64) error {
	return s.Auth.CheckQuota(userID)
}
