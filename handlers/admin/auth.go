package admin

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"zee-mirror/internal/domain"
	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// UnknownSize is the placeholder username when /authorize gets only an ID.
const UnknownSize = "unknown"

func HandleAuth(s *service.BotService, message *tgbotapi.Message, args string) {
	lang := s.GetUserLanguage(message.From.ID)
	if message.From.ID != s.Config.OwnerID {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "auth_title_access_denied"), i18n.T(lang, "admin_need_owner")))
		return
	}

	targetID, username := parseUserArgs(message, args)
	if targetID == 0 {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_bad_format"), i18n.T(lang, "admin_usage_authorize")))
		return
	}

	if username == "" || username == UnknownSize {
		username = "User"
	}

	ctx := context.Background()
	user := domain.User{
		ID:                targetID,
		Username:          username,
		Role:              "authorized",
		CreatedAt:         time.Now(),
		MaxDailyTasks:     s.Config.DefaultMaxDailyTasks,
		MaxDailyBandwidth: s.Config.DefaultMaxDailyBandwidth,
	}
	// Upsert overwrites the language column; keep the target's own choice.
	if existing, getErr := s.DB.GetByID(ctx, targetID); getErr == nil && existing != nil {
		user.Language = existing.Language
	}

	err := s.DB.Upsert(ctx, user)
	if err != nil {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_db_error"), i18n.T(lang, "admin_save_failed", utils.EscapeMarkdownV2(err.Error()))))
		return
	}

	unlimited := i18n.T(lang, "admin_unlimited")
	quota := unlimited
	if s.Config.DefaultMaxDailyTasks != -1 || s.Config.DefaultMaxDailyBandwidth != -1 {
		tasks := unlimited
		if s.Config.DefaultMaxDailyTasks != -1 {
			tasks = fmt.Sprintf("%d", s.Config.DefaultMaxDailyTasks)
		}
		bw := unlimited
		if s.Config.DefaultMaxDailyBandwidth != -1 {
			bw = utils.FormatBytes(s.Config.DefaultMaxDailyBandwidth)
		}
		quota = i18n.T(lang, "admin_quota_format", tasks, bw)
	}

	content := i18n.T(lang, "admin_authorize_info", utils.EscapeMarkdownV2(username), targetID, utils.EscapeMarkdownV2(quota))
	s.Reply(message, service.GetSuccessMessage(i18n.T(lang, "admin_title_authorized"), content))
}

func HandleUnauth(s *service.BotService, message *tgbotapi.Message, args string) {
	lang := s.GetUserLanguage(message.From.ID)
	if message.From.ID != s.Config.OwnerID {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "auth_title_access_denied"), i18n.T(lang, "admin_need_owner")))
		return
	}

	if args == "" {
		s.Reply(message, i18n.T(lang, "admin_unauth_usage"))
		return
	}

	userID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		s.Reply(message, i18n.T(lang, "admin_invalid_user_id"))
		return
	}

	ctx := context.Background()
	if err := s.DB.SetRole(ctx, userID, "user"); err != nil {
		s.Reply(message, i18n.T(lang, "admin_unauth_failed", utils.EscapeMarkdownV2(err.Error())))
		return
	}

	s.Reply(message, i18n.T(lang, "admin_unauth_done", userID))
}

func HandleUserList(s *service.BotService, message *tgbotapi.Message) {
	if !s.IsAuthorized(message.From.ID) {
		return
	}
	lang := s.GetUserLanguage(message.From.ID)

	ctx := context.Background()
	usersCount, _ := s.DB.GetCount(ctx)
	users, err := s.DB.GetAll(ctx)
	if err != nil {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_db_error"), i18n.T(lang, "admin_user_list_failed")))
		return
	}

	var content strings.Builder
	content.WriteString(i18n.T(lang, "admin_user_count", usersCount))
	for _, u := range users {
		status := "✅"
		if !u.IsActive {
			status = "❌"
		}

		role := u.Role

		limits := ""
		if u.MaxDailyTasks != -1 {
			limits += fmt.Sprintf("T:%d ", u.MaxDailyTasks)
		}
		if u.MaxDailyBandwidth != -1 {
			limits += fmt.Sprintf("B:%s ", utils.FormatBytes(u.MaxDailyBandwidth))
		}
		if limits == "" {
			limits = "♾️"
		}

		_, _ = fmt.Fprintf(&content, "%s `%d` \\| %s \\(*%s*\\) \\[%s\\]\n",
			status, u.ID, utils.EscapeMarkdownV2(u.Username), strings.ToUpper(role), utils.EscapeMarkdownV2(limits))
	}

	s.Reply(message, service.ProfessionalMessage(i18n.T(lang, "admin_title_user_list"), content.String()))
}

func RemoveUserHandler(s *service.BotService, message *tgbotapi.Message, args string) {
	lang := s.GetUserLanguage(message.From.ID)
	if !s.IsAdmin(message.From.ID) {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "auth_title_access_denied"), i18n.T(lang, "admin_need_admin")))
		return
	}

	targetID, _ := parseUserArgs(message, args)
	if targetID == 0 {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_bad_format"), i18n.T(lang, "admin_usage_removeuser")))
		return
	}

	if s.IsOwner(targetID) {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_error"), i18n.T(lang, "admin_cannot_remove_owner")))
		return
	}

	ctx := context.Background()
	err := s.DB.Delete(ctx, targetID)
	if err != nil {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_db_error"), i18n.T(lang, "admin_remove_failed", utils.EscapeMarkdownV2(err.Error()))))
		return
	}

	s.Reply(message, service.GetSuccessMessage(i18n.T(lang, "admin_title_user_removed"), i18n.T(lang, "admin_remove_done", targetID)))
}

func SetRoleHandler(s *service.BotService, message *tgbotapi.Message, args string) {
	lang := s.GetUserLanguage(message.From.ID)
	if !s.IsAdmin(message.From.ID) {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "auth_title_access_denied"), i18n.T(lang, "admin_need_admin")))
		return
	}

	targetID, _ := parseUserArgs(message, args)
	if targetID == 0 {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_bad_format"), i18n.T(lang, "admin_usage_setrole")))
		return
	}

	if s.IsOwner(targetID) {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_error"), i18n.T(lang, "admin_cannot_change_owner_role")))
		return
	}

	parts := strings.Fields(args)
	role := "authorized"
	if len(parts) > 1 {
		role = parts[1]
	} else if message.ReplyToMessage != nil && len(parts) > 0 {
		role = parts[0]
	}

	ctx := context.Background()
	err := s.DB.SetRole(ctx, targetID, role)
	if err != nil {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_db_error"), i18n.T(lang, "admin_setrole_failed", utils.EscapeMarkdownV2(err.Error()))))
		return
	}

	s.Reply(message, service.GetSuccessMessage(i18n.T(lang, "admin_title_role_updated"), i18n.T(lang, "admin_setrole_done", targetID, strings.ToUpper(role))))
}

func SetLimitHandler(s *service.BotService, message *tgbotapi.Message, args string) {
	lang := s.GetUserLanguage(message.From.ID)
	if !s.IsAdmin(message.From.ID) {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "auth_title_access_denied"), i18n.T(lang, "admin_need_admin")))
		return
	}

	targetID, _ := parseUserArgs(message, args)
	if targetID == 0 {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_bad_format"), i18n.T(lang, "admin_usage_setlimit")))
		return
	}

	parts := strings.Fields(args)
	remainingArgs := parts
	if len(parts) > 0 {
		var maybeID int64
		_, err := fmt.Sscanf(parts[0], "%d", &maybeID)
		if err == nil && maybeID == targetID {
			remainingArgs = parts[1:]
		}
	}

	if len(remainingArgs) < 2 {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_bad_format"), i18n.T(lang, "admin_setlimit_missing_args")))
		return
	}

	maxTasks, err := strconv.Atoi(remainingArgs[0])
	if err != nil {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_invalid_value"), i18n.T(lang, "admin_tasks_nan")))
		return
	}

	maxBandwidth := utils.ParseBytesString(remainingArgs[1])
	if maxBandwidth == 0 && remainingArgs[1] != "0" && remainingArgs[1] != "-1" {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_invalid_value"), i18n.T(lang, "admin_bandwidth_invalid")))
		return
	}

	if remainingArgs[1] == "-1" {
		maxBandwidth = -1
	}

	ctx := context.Background()
	err = s.DB.SetLimits(ctx, targetID, maxTasks, maxBandwidth)
	if err != nil {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_db_error"), i18n.T(lang, "admin_setlimit_failed", utils.EscapeMarkdownV2(err.Error()))))
		return
	}

	unlimited := i18n.T(lang, "admin_unlimited")
	bwStr := unlimited
	if maxBandwidth != -1 {
		bwStr = utils.FormatBytes(maxBandwidth)
	}
	taskStr := unlimited
	if maxTasks != -1 {
		taskStr = strconv.Itoa(maxTasks)
	}

	content := i18n.T(lang, "admin_setlimit_done", targetID, taskStr, bwStr)
	s.Reply(message, service.GetSuccessMessage(i18n.T(lang, "admin_title_limit_updated"), content))
}

func SetExpireHandler(s *service.BotService, message *tgbotapi.Message, args string) {
	lang := s.GetUserLanguage(message.From.ID)
	if !s.IsAdmin(message.From.ID) {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "auth_title_access_denied"), i18n.T(lang, "admin_need_admin")))
		return
	}

	targetID, _ := parseUserArgs(message, args)
	if targetID == 0 {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_bad_format"), i18n.T(lang, "admin_usage_setexpire")))
		return
	}

	parts := strings.Fields(args)
	remainingArgs := parts
	if len(parts) > 0 {
		var maybeID int64
		_, err := fmt.Sscanf(parts[0], "%d", &maybeID)
		if err == nil && maybeID == targetID {
			remainingArgs = parts[1:]
		}
	}

	if len(remainingArgs) < 1 {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_bad_format"), i18n.T(lang, "admin_expire_missing_days")))
		return
	}

	days, err := strconv.Atoi(remainingArgs[0])
	if err != nil {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_invalid_value"), i18n.T(lang, "admin_days_nan")))
		return
	}

	expiresAt := time.Now().AddDate(0, 0, days)
	ctx := context.Background()
	err = s.DB.SetExpiration(ctx, targetID, expiresAt)
	if err != nil {
		s.Reply(message, service.GetErrorMessage(i18n.T(lang, "admin_title_db_error"), i18n.T(lang, "admin_setexpire_failed", utils.EscapeMarkdownV2(err.Error()))))
		return
	}

	s.Reply(message, service.GetSuccessMessage(i18n.T(lang, "admin_title_expiry_updated"),
		i18n.T(lang, "admin_setexpire_done", targetID, expiresAt.Format("02 Jan 2006"), days)))
}

func parseUserArgs(message *tgbotapi.Message, args string) (int64, string) {
	if message.ReplyToMessage != nil {
		username := message.ReplyToMessage.From.UserName
		if username == "" {
			username = message.ReplyToMessage.From.FirstName
		}
		return message.ReplyToMessage.From.ID, username
	}

	if args != "" {
		parts := strings.Fields(args)
		if len(parts) > 0 {
			var id int64
			_, err := fmt.Sscanf(parts[0], "%d", &id)
			if err != nil {
				return 0, ""
			}
			username := UnknownSize
			if len(parts) > 1 {
				username = parts[1]
			}
			return id, username
		}
	}

	return 0, ""
}
