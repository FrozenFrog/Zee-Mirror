package main

import (
	"log/slog"
	"strings"
	"zee-mirror/handlers/admin"
	"zee-mirror/handlers/basic"
	"zee-mirror/handlers/download"
	"zee-mirror/handlers/file"
	"zee-mirror/handlers/media"
	"zee-mirror/handlers/search"
	"zee-mirror/handlers/storage"
	"zee-mirror/internal/router"
	"zee-mirror/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func setupRoutes(r *router.Router) {
	setupBasicRoutes(r)
	setupDownloadRoutes(r)
	setupAdminRoutes(r)
	setupFileManagerRoutes(r)
	setupMediaRoutes(r)
	setupCallbackRoutes(r)

	var helpCommands []basic.HelpCommandEntry
	for _, cmd := range r.GetAllCommandsFlat() {
		helpCommands = append(helpCommands, basic.HelpCommandEntry{
			Name:        cmd.Name,
			Aliases:     cmd.Aliases,
			Description: cmd.Description,
			Category:    cmd.Category,
			Emoji:       cmd.Emoji,
		})
	}
	basic.SetRegisteredCommands(helpCommands)
}

// registerBotMenu pushes the router's command list into Telegram's native
// command menu so users get autocomplete without BotFather setup.
func registerBotMenu(bot *tgbotapi.BotAPI, r *router.Router) {
	var cmds []tgbotapi.BotCommand
	for _, c := range r.GetAllCommandsFlat() {
		if c.Category == "admin" {
			continue
		}
		cmds = append(cmds, tgbotapi.BotCommand{Command: c.Name, Description: c.Description})
	}
	if len(cmds) > 100 {
		cmds = cmds[:100]
	}
	if _, err := bot.Request(tgbotapi.NewSetMyCommands(cmds...)); err != nil {
		slog.Warn("Failed to register Telegram command menu", "error", err)
	}
}

func setupBasicRoutes(r *router.Router) {
	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "start", Description: "Tin nhắn chào mừng",
		Category: "general", Emoji: "🤖",
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.StartHandler(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "help", Description: "Hiển thị menu trợ giúp",
		Category: "general", Emoji: "❓",
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.HelpHandler(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "settings", Description: "Menu cài đặt bot",
		Category: "general", Emoji: "⚙️",
		DetailedFn: basic.GetHelpSettings,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleSettings(m)
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "ping", Description: "Kiểm tra độ trễ của bot",
		Category: "monitor", Emoji: "🏓",
		DetailedFn: basic.GetHelpPing,
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.HandlePing(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "speed", Description: "Kiểm tra tốc độ internet",
		Category: "monitor", Emoji: "🚀",
		DetailedFn: basic.GetHelpSpeed,
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.HandleSpeed(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "stats", Description: "Thống kê sử dụng bot",
		Category: "monitor", Emoji: "📈",
		DetailedFn: basic.GetHelpStats,
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.HandleStats(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "lang", Aliases: []string{"language"},
		Description: "Chọn ngôn ngữ",
		Category:    "general", Emoji: "🌐",
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleLanguage(m)
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "schedule", Description: "Lên lịch task",
		Category: "general", Emoji: "📅",
	}, func(s *service.BotService, m *tgbotapi.Message) {
		basic.HandleSchedule(s, m, m.CommandArguments())
	})
}

func setupDownloadRoutes(r *router.Router) {
	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "mirror", Aliases: []string{"m"},
		Description: "Upload file từ URL lên Google Drive",
		Category:    "download", Emoji: "📥",
		DetailedFn: basic.GetHelpMirror,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		download.HandleMirror(s, m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "leech", Aliases: []string{"l"},
		Description: "Tải file từ URL về Telegram",
		Category:    "download", Emoji: "📤",
		DetailedFn: basic.GetHelpLeech,
	}, func(s *service.BotService, m *tgbotapi.Message) { download.HandleLeech(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "viking", Aliases: []string{"v"},
		Description: "Upload lên Viking File",
		Category:    "download", Emoji: "⚔️",
		DetailedFn: basic.GetHelpViking,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleViking(m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "ytdlp", Aliases: []string{"y", "yt"},
		Description: "Tải video qua YT-DLP lên Drive",
		Category:    "download", Emoji: "🎬",
		DetailedFn: basic.GetHelpYTDLP,
	}, func(s *service.BotService, m *tgbotapi.Message) { download.HandleYTDLP(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "ytdlpleech", Aliases: []string{"yl"},
		Description: "Tải video qua YT-DLP về Telegram",
		Category:    "download", Emoji: "🎬",
		DetailedFn: basic.GetHelpYTDLPLeech,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		download.HandleYTDLPLeech(s, m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "torrent", Aliases: []string{"t"},
		Description: "Tải qua Magnet link/file torrent",
		Category:    "download", Emoji: "🧲",
		DetailedFn: basic.GetHelpTorrent,
	}, func(s *service.BotService, m *tgbotapi.Message) { download.HandleTorrent(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "clone", Aliases: []string{"cl"},
		Description: "Clone file/thư mục Google Drive",
		Category:    "download", Emoji: "📋",
		DetailedFn: basic.GetHelpClone,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleClone(m, m.CommandArguments())
	})
}

func setupAdminRoutes(r *router.Router) {
	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "status", Aliases: []string{"st"},
		Description: "Xem các task đang chạy",
		Category:    "monitor", Emoji: "📊",
		DetailedFn: basic.GetHelpStatus,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleStatus(m)
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "cancel", Aliases: []string{"c"},
		Description: "Hủy một task cụ thể",
		Category:    "task", Emoji: "❌",
		DetailedFn: basic.GetHelpCancel,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleCancel(m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "cancelall",
		Description: "Hủy tất cả task đang chạy",
		Category:    "task", Emoji: "🚫",
		DetailedFn: basic.GetHelpCancelAll,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleCancelAll(m)
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "notify",
		Description: "Đặt URL webhook nhận thông báo khi task hoàn tất",
		Category:    "task", Emoji: "🔔",
		DetailedFn: basic.GetHelpNotify,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		basic.HandleNotify(s, m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "search", Aliases: []string{"searchall"},
		Description: "Tìm torrent từ các tracker",
		Category:    "download", Emoji: "🔍",
		DetailedFn: basic.GetHelpSearch,
	}, func(s *service.BotService, m *tgbotapi.Message) { search.HandleSearch(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "batch",
		Description: "Tải nhiều URL cùng lúc (Batch)",
		Category:    "download", Emoji: "📦",
		DetailedFn: basic.GetHelpBatch,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleBatch(m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "authorize",
		Description: "Cấp quyền dùng bot cho người dùng",
		Category:    "admin", Emoji: "✅",
		DetailedFn: basic.GetHelpAuthorize,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		admin.HandleAuth(s, m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "unauthorize",
		Description: "Thu hồi quyền truy cập của người dùng",
		Category:    "admin", Emoji: "❌",
		DetailedFn: basic.GetHelpUnauthorize,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		admin.HandleUnauth(s, m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "removeuser",
		Description: "Xóa người dùng khỏi database",
		Category:    "admin", Emoji: "🗑️",
	}, func(s *service.BotService, m *tgbotapi.Message) { admin.RemoveUserHandler(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "setrole",
		Description: "Đặt vai trò người dùng",
		Category:    "admin", Emoji: "👤",
	}, func(s *service.BotService, m *tgbotapi.Message) { admin.SetRoleHandler(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "setlimit",
		Description: "Đặt giới hạn tải của người dùng",
		Category:    "admin", Emoji: "📏",
	}, func(s *service.BotService, m *tgbotapi.Message) { admin.SetLimitHandler(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "setexpire",
		Description: "Đặt thời hạn truy cập của người dùng",
		Category:    "admin", Emoji: "⏰",
	}, func(s *service.BotService, m *tgbotapi.Message) { admin.SetExpireHandler(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "users",
		Description: "Danh sách người dùng đã đăng ký",
		Category:    "admin", Emoji: "👥",
		DetailedFn: basic.GetHelpUsers,
	}, func(s *service.BotService, m *tgbotapi.Message) { admin.HandleUserList(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "setalertchannel",
		Description: "Đặt kênh cảnh báo lỗi",
		Category:    "admin", Emoji: "🚨",
		DetailedFn: basic.GetHelpSetAlertChannel,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleSetAlertChannel(m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        service.CmdSystem,
		Description: "Thông tin hệ thống (CPU, RAM, ổ đĩa)",
		Category:    "monitor", Emoji: "🖥️",
		DetailedFn: basic.GetHelpSystem,
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.HandleSystem(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        service.CmdHealth,
		Description: "Kiểm tra tình trạng các thành phần",
		Category:    "monitor", Emoji: "🏥",
		DetailedFn: basic.GetHelpHealth,
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.HandleHealth(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        service.CmdLogs,
		Description: "Xem log của bot",
		Category:    "monitor", Emoji: "📜",
		DetailedFn: basic.GetHelpLogs,
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.HandleLogs(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "recover",
		Description: "Khôi phục các task chưa hoàn tất",
		Category:    "recovery", Emoji: "🔄",
		DetailedFn: basic.GetHelpRecover,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleRecover(m)
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "recoverystatus",
		Description: "Xem thống kê khôi phục",
		Category:    "recovery", Emoji: "📊",
		DetailedFn: basic.GetHelpRecoveryStatus,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleRecoveryStatus(m)
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "join",
		Description: "Tham gia kênh Telegram qua userbot",
		Category:    "general", Emoji: "📢",
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.HandleJoin(s, m, m.CommandArguments()) })
}

func setupFileManagerRoutes(r *router.Router) {
	fmHandler := func(s *service.BotService, m *tgbotapi.Message) {
		cmd := m.Command()
		args := m.CommandArguments()
		switch cmd {
		case "ls", "dir":
			file.HandleDriveList(s, m, args, 0)
		case "mkdir":
			file.HandleDriveMkdir(s, m, args)
		case "rm":
			file.HandleDriveDelete(s, m, args)
		case "mv":
			file.HandleDriveMove(s, m, args)
		case "share":
			file.HandleDriveShare(s, m, args)
		case "find":
			file.HandleDriveSearch(s, m, args)
		}
	}

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "ls", Aliases: []string{"dir"},
		Description: "Liệt kê file trên cloud storage",
		Category:    "files", Emoji: "📂",
		DetailedFn: basic.GetHelpLs,
	}, fmHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "mkdir",
		Description: "Tạo thư mục trên cloud storage",
		Category:    "files", Emoji: "📁",
		DetailedFn: basic.GetHelpMkdir,
	}, fmHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "rm",
		Description: "Xóa file/thư mục khỏi storage",
		Category:    "files", Emoji: "🗑️",
		DetailedFn: basic.GetHelpRm,
	}, fmHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "mv",
		Description: "Di chuyển/đổi tên file trong storage",
		Category:    "files", Emoji: "📦",
		DetailedFn: basic.GetHelpMv,
	}, fmHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "share",
		Description: "Tạo link chia sẻ cho file",
		Category:    "files", Emoji: "🔗",
		DetailedFn: basic.GetHelpShare,
	}, fmHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "find",
		Description: "Tìm file trên cloud storage",
		Category:    "files", Emoji: "🔍",
		DetailedFn: basic.GetHelpFind,
	}, fmHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "storages",
		Description: "Danh sách storage hiện có",
		Category:    "storage", Emoji: "📋",
		DetailedFn: basic.GetHelpStorages,
	}, func(s *service.BotService, m *tgbotapi.Message) { storage.HandleStorages(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "setstorage",
		Description: "Chọn storage đang dùng",
		Category:    "storage", Emoji: "⚙️",
		DetailedFn: basic.GetHelpSetStorage,
	}, func(s *service.BotService, m *tgbotapi.Message) { storage.HandleSetStorage(s, m, m.CommandArguments()) })
}

func setupMediaRoutes(r *router.Router) {
	mediaHandler := func(s *service.BotService, m *tgbotapi.Message) {
		cmd := m.Command()
		args := m.CommandArguments()
		switch cmd {
		case "extractaudio":
			media.HandleExtractAudio(s, m, args)
		case "compress":
			media.HandleCompressVideo(s, m, args)
		case "thumbnail":
			media.HandleGenerateThumbnail(s, m, args)
		case "screenshots":
			media.HandleScreenshots(s, m, args)
		case "subtitle":
			media.HandleEmbedSubtitle(s, m, args)
		case "hardsub":
			media.HandleHardsub(s, m, args)
		case "rescale":
			media.HandleRescale(s, m, args)
		case "convert":
			media.HandleConvertFormat(s, m, args)
		case "mediainfo":
			media.HandleMediaInfo(s, m, args)
		}
	}

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "extractaudio",
		Description: "Tách audio từ video (MP3)",
		Category:    "media", Emoji: "🎵",
		DetailedFn: basic.GetHelpExtractAudio,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "compress",
		Description: "Nén giảm dung lượng video",
		Category:    "media", Emoji: "🗜️",
		DetailedFn: basic.GetHelpCompress,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "thumbnail",
		Description: "Tạo thumbnail từ video",
		Category:    "media", Emoji: "🖼️",
		DetailedFn: basic.GetHelpThumbnail,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "screenshots",
		Description: "Chụp nhiều ảnh màn hình từ video",
		Category:    "media", Emoji: "📸",
		DetailedFn: basic.GetHelpScreenshots,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "subtitle",
		Description: "Nhúng phụ đề mềm (soft-sub)",
		Category:    "media", Emoji: "💬",
		DetailedFn: basic.GetHelpSubtitle,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "hardsub",
		Description: "Ghi cứng phụ đề vào video (hard-sub)",
		Category:    "media", Emoji: "🔥",
		DetailedFn: basic.GetHelpHardsub,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "rescale",
		Description: "Đổi độ phân giải video",
		Category:    "media", Emoji: "📐",
		DetailedFn: basic.GetHelpRescale,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "convert",
		Description: "Chuyển đổi định dạng file",
		Category:    "media", Emoji: "🔄",
		DetailedFn: basic.GetHelpConvert,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:        "mediainfo",
		Description: "Xem thông tin chi tiết file media",
		Category:    "media", Emoji: "ℹ️",
		DetailedFn: basic.GetHelpMediaInfo,
	}, mediaHandler)
}

func setupCallbackRoutes(r *router.Router) {
	r.RegisterCallback("dashboard", func(s *service.BotService, cb *tgbotapi.CallbackQuery) {
		if !ensureCallbackMessage(s, cb, "dashboard") {
			return
		}
		basic.HandleDashboardCallback(s, cb)
	})
	r.RegisterCallback("help", func(s *service.BotService, cb *tgbotapi.CallbackQuery) {
		if !ensureCallbackMessage(s, cb, "help") {
			return
		}
		basic.HandleHelpCallback(s, cb, "")
	})
	r.RegisterCallback("refresh_status", func(s *service.BotService, cb *tgbotapi.CallbackQuery) {
		if !ensureCallbackMessage(s, cb, "refresh_status") {
			return
		}
		s.HandleRefreshStatusCallback(cb)
	})
	r.RegisterCallback("status_overview", func(s *service.BotService, cb *tgbotapi.CallbackQuery) {
		if !ensureCallbackMessage(s, cb, "status_overview") {
			return
		}
		s.HandleStatusOverviewCallback(cb)
	})

	searchHandler := func(s *service.BotService, cb *tgbotapi.CallbackQuery) {
		if !ensureCallbackMessage(s, cb, "search") {
			return
		}
		parts := strings.Split(cb.Data, ":")
		if parts[0] == "t_search" {
			search.HandleSearchCallback(s, cb, parts)
		} else {
			search.HandleSearchNavCallback(s, cb, parts)
		}
	}
	r.RegisterCallback("t_search", searchHandler)
	r.RegisterCallback("t_page", searchHandler)
	r.RegisterCallback("t_item", searchHandler)
	r.RegisterCallback("t_back", searchHandler)
	r.RegisterCallback("t_close", searchHandler)

	taskActionHandler := func(s *service.BotService, cb *tgbotapi.CallbackQuery) {
		if !ensureCallbackMessage(s, cb, "task_action") {
			return
		}
		parts := strings.Split(cb.Data, ":")
		switch parts[0] {
		case "ytdlp_q":
			download.YTDLPQualityCallbackHandler(s, cb, parts)
		case "settings":
			s.HandleSettingsCallback(cb, parts)
		case "batch":
			s.HandleBatchCallback(cb, parts)
		case "confirm":
			s.HandleConfirmCallback(cb, parts)
		case "torrent_sel":
			download.HandleTorrentSelectionCallback(s, cb, parts)
		}
	}
	r.RegisterCallback("ytdlp_q", taskActionHandler)
	r.RegisterCallback("settings", taskActionHandler)
	r.RegisterCallback("batch", taskActionHandler)
	r.RegisterCallback("confirm", taskActionHandler)
	r.RegisterCallback("torrent_sel", taskActionHandler)

	systemHandler := func(s *service.BotService, cb *tgbotapi.CallbackQuery) {
		if !ensureCallbackMessage(s, cb, "system") {
			return
		}
		parts := strings.Split(cb.Data, ":")
		switch parts[0] {
		case "stats":
			basic.HandleStatsCallback(s, cb, parts)
		case service.CmdSystem:
			basic.HandleSystemCallback(s, cb, parts)
		case "storage":
			storage.HandleStorageCallback(s, cb, parts)
		case "drive", "dr":
			file.HandleDriveCallback(s, cb, parts)
		}
	}
	r.RegisterCallback("stats", systemHandler)
	r.RegisterCallback(service.CmdSystem, systemHandler)
	r.RegisterCallback("storage", systemHandler)
	r.RegisterCallback("drive", systemHandler)
	r.RegisterCallback("dr", systemHandler)

	r.RegisterCallback("media_m", func(s *service.BotService, cb *tgbotapi.CallbackQuery) {
		if !ensureCallbackMessage(s, cb, "media_m") {
			return
		}
		parts := strings.Split(cb.Data, ":")
		media.HandleMediaMirrorCallback(s, cb, parts)
	})
	r.RegisterCallback("media", func(s *service.BotService, cb *tgbotapi.CallbackQuery) {
		if !ensureCallbackMessage(s, cb, "media") {
			return
		}
		parts := strings.Split(cb.Data, ":")
		media.HandleMediaMenuCallback(s, cb, parts)
	})

	r.RegisterCallback("wizard", func(s *service.BotService, cb *tgbotapi.CallbackQuery) {
		if !ensureCallbackMessage(s, cb, "wizard") {
			return
		}
		parts := strings.Split(cb.Data, ":")
		download.HandleMirrorWizardCallback(s, cb, parts)
	})
}

func ensureCallbackMessage(s *service.BotService, cb *tgbotapi.CallbackQuery, route string) bool {
	if cb == nil {
		slog.Warn("Ignoring nil callback", "route", route)
		return false
	}
	if cb.Message == nil {
		slog.Warn("Ignoring callback without message", "route", route, "data", cb.Data)
		if cb.ID != "" {
			_, _ = s.Bot.Request(tgbotapi.NewCallback(cb.ID, "Thao tác này không được hỗ trợ ở đây"))
		}
		return false
	}
	return true
}
