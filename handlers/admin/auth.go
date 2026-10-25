package admin

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"zee-mirror/internal/domain"
	"zee-mirror/internal/service"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	UnknownSize  = "unknown"
	UnlimitedStr = "Không giới hạn"
)

func HandleAuth(s *service.BotService, message *tgbotapi.Message, args string) {
	if message.From.ID != s.Config.OwnerID {
		s.Reply(message, service.GetErrorMessage("TRUY CẬP BỊ TỪ CHỐI", "Chỉ Owner mới có thể dùng lệnh này\\."))
		return
	}

	targetID, username := parseUserArgs(message, args)
	if targetID == 0 {
		s.Reply(message, service.GetErrorMessage("SAI ĐỊNH DẠNG", "Cách dùng: /authorize ID \\[username\\] hoặc reply vào người dùng\\."))
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

	err := s.DB.Upsert(ctx, user)
	if err != nil {
		s.Reply(message, service.GetErrorMessage("LỖI CƠ SỞ DỮ LIỆU", "Lưu người dùng thất bại: "+utils.EscapeMarkdownV2(err.Error())))
		return
	}

	quota := UnlimitedStr
	if s.Config.DefaultMaxDailyTasks != -1 || s.Config.DefaultMaxDailyBandwidth != -1 {
		tasks := UnlimitedStr
		if s.Config.DefaultMaxDailyTasks != -1 {
			tasks = fmt.Sprintf("%d", s.Config.DefaultMaxDailyTasks)
		}
		bw := UnlimitedStr
		if s.Config.DefaultMaxDailyBandwidth != -1 {
			bw = utils.FormatBytes(s.Config.DefaultMaxDailyBandwidth)
		}
		quota = fmt.Sprintf("Task: %s | BW: %s", tasks, bw)
	}

	content := fmt.Sprintf("👤 *Người dùng:* %s\n🏷️ *ID:* `%d`\n🔰 *Vai trò:* `Được cấp quyền`\n♾️ *Hạn mức:* `%s`",
		utils.EscapeMarkdownV2(username), targetID, utils.EscapeMarkdownV2(quota))
	s.Reply(message, service.GetSuccessMessage("ĐÃ CẤP QUYỀN", content))
}

func HandleUnauth(s *service.BotService, message *tgbotapi.Message, args string) {
	if message.From.ID != s.Config.OwnerID {
		s.Reply(message, service.GetErrorMessage("TRUY CẬP BỊ TỪ CHỐI", "Chỉ Owner mới có thể dùng lệnh này\\."))
		return
	}

	if args == "" {
		s.Reply(message, "⚠️ *Sai định dạng*\n\nCách dùng: `/unauth <user_id>`")
		return
	}

	userID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		s.Reply(message, "⚠️ *ID người dùng không hợp lệ*")
		return
	}

	ctx := context.Background()
	if err := s.DB.SetRole(ctx, userID, "user"); err != nil {
		s.Reply(message, fmt.Sprintf("❌ *Xóa người dùng thất bại:* %s", utils.EscapeMarkdownV2(err.Error())))
		return
	}

	s.Reply(message, fmt.Sprintf("✅ *Đã xóa người dùng %d khỏi danh sách được cấp quyền*", userID))
}

func HandleUserList(s *service.BotService, message *tgbotapi.Message) {
	if !s.IsAuthorized(message.From.ID) {
		return
	}

	ctx := context.Background()
	usersCount, _ := s.DB.GetCount(ctx)
	users, err := s.DB.GetAll(ctx)
	if err != nil {
		s.Reply(message, service.GetErrorMessage("LỖI CƠ SỞ DỮ LIỆU", "Không thể lấy danh sách người dùng\\."))
		return
	}

	var content strings.Builder
	_, _ = fmt.Fprintf(&content, "Tổng số người dùng: %d\n\n", usersCount)
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

	s.Reply(message, service.ProfessionalMessage("DANH SÁCH NGƯỜI DÙNG", content.String()))
}

func RemoveUserHandler(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.IsAdmin(message.From.ID) {
		s.Reply(message, service.GetErrorMessage("TRUY CẬP BỊ TỪ CHỐI", "Chỉ Admin/Owner mới có thể dùng lệnh này\\."))
		return
	}

	targetID, _ := parseUserArgs(message, args)
	if targetID == 0 {
		s.Reply(message, service.GetErrorMessage("SAI ĐỊNH DẠNG", "Cách dùng: /removeuser ID hoặc reply vào người dùng\\."))
		return
	}

	if s.IsOwner(targetID) {
		s.Reply(message, service.GetErrorMessage("LỖI", "Không thể xóa Owner\\."))
		return
	}

	ctx := context.Background()
	err := s.DB.Delete(ctx, targetID)
	if err != nil {
		s.Reply(message, service.GetErrorMessage("LỖI CƠ SỞ DỮ LIỆU", "Xóa người dùng thất bại: "+utils.EscapeMarkdownV2(err.Error())))
		return
	}

	s.Reply(message, service.GetSuccessMessage("ĐÃ XÓA NGƯỜI DÙNG", fmt.Sprintf("Người dùng `%d` đã bị xóa khỏi cơ sở dữ liệu\\.", targetID)))
}

func SetRoleHandler(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.IsAdmin(message.From.ID) {
		s.Reply(message, service.GetErrorMessage("TRUY CẬP BỊ TỪ CHỐI", "Chỉ Admin/Owner mới có thể dùng lệnh này\\."))
		return
	}

	targetID, _ := parseUserArgs(message, args)
	if targetID == 0 {
		s.Reply(message, service.GetErrorMessage("SAI ĐỊNH DẠNG", "Cách dùng: /setrole ID <role> hoặc reply vào người dùng kèm <role>\\."))
		return
	}

	if s.IsOwner(targetID) {
		s.Reply(message, service.GetErrorMessage("LỖI", "Không thể đổi vai trò của Owner\\."))
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
		s.Reply(message, service.GetErrorMessage("LỖI CƠ SỞ DỮ LIỆU", "Đổi vai trò thất bại: "+utils.EscapeMarkdownV2(err.Error())))
		return
	}

	s.Reply(message, service.GetSuccessMessage("ĐÃ CẬP NHẬT VAI TRÒ", fmt.Sprintf("Vai trò của người dùng `%d` đã được đổi thành `%s`\\.", targetID, strings.ToUpper(role))))
}

func SetLimitHandler(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.IsAdmin(message.From.ID) {
		s.Reply(message, service.GetErrorMessage("TRUY CẬP BỊ TỪ CHỐI", "Chỉ Admin/Owner mới có thể dùng lệnh này\\."))
		return
	}

	targetID, _ := parseUserArgs(message, args)
	if targetID == 0 {
		s.Reply(message, service.GetErrorMessage("SAI ĐỊNH DẠNG", "Cách dùng: /setlimit ID <tasks> <bandwidth> hoặc reply kèm <tasks> <bandwidth>\\.\nDùng \\-1 để không giới hạn\\."))
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
		s.Reply(message, service.GetErrorMessage("SAI ĐỊNH DẠNG", "Hãy nhập số Task và bandwidth \\(ví dụ: 10 50GB\\)\\."))
		return
	}

	maxTasks, err := strconv.Atoi(remainingArgs[0])
	if err != nil {
		s.Reply(message, service.GetErrorMessage("GIÁ TRỊ KHÔNG HỢP LỆ", "Số Task phải là số\\."))
		return
	}

	maxBandwidth := utils.ParseBytesString(remainingArgs[1])
	if maxBandwidth == 0 && remainingArgs[1] != "0" && remainingArgs[1] != "-1" {
		s.Reply(message, service.GetErrorMessage("GIÁ TRỊ KHÔNG HỢP LỆ", "Định dạng bandwidth không hợp lệ \\(ví dụ: 10GB, 100MB, \\-1\\)\\."))
		return
	}

	if remainingArgs[1] == "-1" {
		maxBandwidth = -1
	}

	ctx := context.Background()
	err = s.DB.SetLimits(ctx, targetID, maxTasks, maxBandwidth)
	if err != nil {
		s.Reply(message, service.GetErrorMessage("LỖI CƠ SỞ DỮ LIỆU", "Lưu giới hạn thất bại: "+utils.EscapeMarkdownV2(err.Error())))
		return
	}

	bwStr := UnlimitedStr
	if maxBandwidth != -1 {
		bwStr = utils.FormatBytes(maxBandwidth)
	}
	taskStr := UnlimitedStr
	if maxTasks != -1 {
		taskStr = strconv.Itoa(maxTasks)
	}

	content := fmt.Sprintf("👤 *ID người dùng:* `%d`\n📋 *Task mỗi ngày:* `%s`\n📊 *Bandwidth mỗi ngày:* `%s`",
		targetID, taskStr, bwStr)
	s.Reply(message, service.GetSuccessMessage("ĐÃ CẬP NHẬT GIỚI HẠN", content))
}

func SetExpireHandler(s *service.BotService, message *tgbotapi.Message, args string) {
	if !s.IsAdmin(message.From.ID) {
		s.Reply(message, service.GetErrorMessage("TRUY CẬP BỊ TỪ CHỐI", "Chỉ Admin/Owner mới có thể dùng lệnh này\\."))
		return
	}

	targetID, _ := parseUserArgs(message, args)
	if targetID == 0 {
		s.Reply(message, service.GetErrorMessage("SAI ĐỊNH DẠNG", "Cách dùng: /setexpire ID <days> hoặc reply kèm <days>\\."))
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
		s.Reply(message, service.GetErrorMessage("SAI ĐỊNH DẠNG", "Hãy nhập số ngày hiệu lực\\."))
		return
	}

	days, err := strconv.Atoi(remainingArgs[0])
	if err != nil {
		s.Reply(message, service.GetErrorMessage("GIÁ TRỊ KHÔNG HỢP LỆ", "Số ngày phải là số\\."))
		return
	}

	expiresAt := time.Now().AddDate(0, 0, days)
	ctx := context.Background()
	err = s.DB.SetExpiration(ctx, targetID, expiresAt)
	if err != nil {
		s.Reply(message, service.GetErrorMessage("LỖI CƠ SỞ DỮ LIỆU", "Lưu thời hạn thất bại: "+utils.EscapeMarkdownV2(err.Error())))
		return
	}

	s.Reply(message, service.GetSuccessMessage("ĐÃ CẬP NHẬT THỜI HẠN", fmt.Sprintf("Thời hạn của người dùng `%d` đã được cập nhật đến *%s* \\(%d ngày\\)\\.",
		targetID, expiresAt.Format("02 Jan 2006"), days)))
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
