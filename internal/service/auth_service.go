package service

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"zee-mirror/internal/config"
	"zee-mirror/internal/errors"
	"zee-mirror/internal/forum"
	"zee-mirror/internal/repository"
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

func (s *AuthService) CheckQuota(userID int64) error {
	if s.IsOwner(userID) {
		return nil
	}

	ctx := context.Background()
	user, err := s.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	if !user.IsActive {
		if user.ExpiresAt.Valid {
			return fmt.Errorf("%w: akses anda telah berakhir pada %s", errors.ErrAccountInactive, user.ExpiresAt.Time.Format("02 Jan 2006"))
		}
		return errors.ErrAccountInactive
	}

	if user.MaxDailyTasks == -1 && user.MaxDailyBandwidth == -1 {
		return nil
	}

	stats, err := s.DB.GetUserTodayStats(ctx, userID)
	if err != nil {
		return nil
	}

	if user.MaxDailyTasks != -1 && stats.TotalTasks >= user.MaxDailyTasks {
		return fmt.Errorf("%w: (%d/%d)", errors.ErrQuotaExceeded, stats.TotalTasks, user.MaxDailyTasks)
	}

	if user.MaxDailyBandwidth != -1 && stats.TotalBandwidth >= user.MaxDailyBandwidth {
		return fmt.Errorf("%w: bandwidth (%s/%s)", errors.ErrQuotaExceeded, utils.FormatBytes(stats.TotalBandwidth), utils.FormatBytes(user.MaxDailyBandwidth))
	}

	return nil
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
		if c.TopicID == 0 {
			scope = scopeChat
		} else if c.TopicID == topic {
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
