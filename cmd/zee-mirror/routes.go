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
	"zee-mirror/pkg/i18n"

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
// command menu so users get autocomplete without BotFather setup. The default
// list uses DefaultLang descriptions; every other supported language gets its
// own list, which Telegram shows to users whose app language matches.
func registerBotMenu(bot *tgbotapi.BotAPI, r *router.Router) {
	if _, err := bot.Request(tgbotapi.NewSetMyCommands(botMenuCommands(r, i18n.DefaultLang)...)); err != nil {
		slog.Warn("Failed to register Telegram command menu", "error", err)
	}
	for _, l := range i18n.Languages {
		if l.Code == i18n.DefaultLang {
			continue
		}
		cfg := tgbotapi.NewSetMyCommandsWithScopeAndLanguage(tgbotapi.NewBotCommandScopeDefault(), l.Code, botMenuCommands(r, l.Code)...)
		if _, err := bot.Request(cfg); err != nil {
			slog.Warn("Failed to register Telegram command menu", "lang", l.Code, "error", err)
		}
	}
}

// botMenuCommands lists the non-admin commands with descriptions in lang,
// capped at Telegram's limit of 100.
func botMenuCommands(r *router.Router, lang string) []tgbotapi.BotCommand {
	var cmds []tgbotapi.BotCommand
	for _, c := range r.GetAllCommandsFlat() {
		if c.Category == "admin" {
			continue
		}
		cmds = append(cmds, tgbotapi.BotCommand{Command: c.Name, Description: c.LocalizedDescription(lang)})
	}
	if len(cmds) > 100 {
		cmds = cmds[:100]
	}
	return cmds
}

func setupBasicRoutes(r *router.Router) {
	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "start",
		Category: "general", Emoji: "🤖",
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.StartHandler(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "help",
		Category: "general", Emoji: "❓",
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.HelpHandler(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "settings",
		Category: "general", Emoji: "⚙️",
		DetailedFn: basic.GetHelpSettings,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleSettings(m)
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "ping",
		Category: "monitor", Emoji: "🏓",
		DetailedFn: basic.GetHelpPing,
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.HandlePing(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "speed",
		Category: "monitor", Emoji: "🚀",
		DetailedFn: basic.GetHelpSpeed,
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.HandleSpeed(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "stats",
		Category: "monitor", Emoji: "📈",
		DetailedFn: basic.GetHelpStats,
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.HandleStats(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "lang", Aliases: []string{"language"},
		Category: "general", Emoji: "🌐",
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleLanguage(m)
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "schedule",
		Category: "general", Emoji: "📅",
	}, func(s *service.BotService, m *tgbotapi.Message) {
		basic.HandleSchedule(s, m, m.CommandArguments())
	})
}

func setupDownloadRoutes(r *router.Router) {
	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "mirror", Aliases: []string{"m"},
		Category: "download", Emoji: "📥",
		DetailedFn: basic.GetHelpMirror,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		download.HandleMirror(s, m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "leech", Aliases: []string{"l"},
		Category: "download", Emoji: "📤",
		DetailedFn: basic.GetHelpLeech,
	}, func(s *service.BotService, m *tgbotapi.Message) { download.HandleLeech(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "viking", Aliases: []string{"v"},
		Category: "download", Emoji: "⚔️",
		DetailedFn: basic.GetHelpViking,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleViking(m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "ytdlp", Aliases: []string{"y", "yt"},
		Category: "download", Emoji: "🎬",
		DetailedFn: basic.GetHelpYTDLP,
	}, func(s *service.BotService, m *tgbotapi.Message) { download.HandleYTDLP(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "ytdlpleech", Aliases: []string{"yl"},
		Category: "download", Emoji: "🎬",
		DetailedFn: basic.GetHelpYTDLPLeech,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		download.HandleYTDLPLeech(s, m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "torrent", Aliases: []string{"t"},
		Category: "download", Emoji: "🧲",
		DetailedFn: basic.GetHelpTorrent,
	}, func(s *service.BotService, m *tgbotapi.Message) { download.HandleTorrent(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "clone", Aliases: []string{"cl"},
		Category: "download", Emoji: "📋",
		DetailedFn: basic.GetHelpClone,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleClone(m, m.CommandArguments())
	})
}

func setupAdminRoutes(r *router.Router) {
	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "status", Aliases: []string{"st"},
		Category: "monitor", Emoji: "📊",
		DetailedFn: basic.GetHelpStatus,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleStatus(m)
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "cancel", Aliases: []string{"c"},
		Category: "task", Emoji: "❌",
		DetailedFn: basic.GetHelpCancel,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleCancel(m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "cancelall",
		Category: "task", Emoji: "🚫",
		DetailedFn: basic.GetHelpCancelAll,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleCancelAll(m)
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "notify",
		Category: "task", Emoji: "🔔",
		DetailedFn: basic.GetHelpNotify,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		basic.HandleNotify(s, m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name: "search", Aliases: []string{"searchall"},
		Category: "download", Emoji: "🔍",
		DetailedFn: basic.GetHelpSearch,
	}, func(s *service.BotService, m *tgbotapi.Message) { search.HandleSearch(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "batch",
		Category: "download", Emoji: "📦",
		DetailedFn: basic.GetHelpBatch,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleBatch(m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "authorize",
		Category: "admin", Emoji: "✅",
		DetailedFn: basic.GetHelpAuthorize,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		admin.HandleAuth(s, m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "unauthorize",
		Category: "admin", Emoji: "❌",
		DetailedFn: basic.GetHelpUnauthorize,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		admin.HandleUnauth(s, m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "removeuser",
		Category: "admin", Emoji: "🗑️",
	}, func(s *service.BotService, m *tgbotapi.Message) { admin.RemoveUserHandler(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "setrole",
		Category: "admin", Emoji: "👤",
	}, func(s *service.BotService, m *tgbotapi.Message) { admin.SetRoleHandler(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "setlimit",
		Category: "admin", Emoji: "📏",
	}, func(s *service.BotService, m *tgbotapi.Message) { admin.SetLimitHandler(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "setexpire",
		Category: "admin", Emoji: "⏰",
	}, func(s *service.BotService, m *tgbotapi.Message) { admin.SetExpireHandler(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "users",
		Category: "admin", Emoji: "👥",
		DetailedFn: basic.GetHelpUsers,
	}, func(s *service.BotService, m *tgbotapi.Message) { admin.HandleUserList(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "setalertchannel",
		Category: "admin", Emoji: "🚨",
		DetailedFn: basic.GetHelpSetAlertChannel,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleSetAlertChannel(m, m.CommandArguments())
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     service.CmdSystem,
		Category: "monitor", Emoji: "🖥️",
		DetailedFn: basic.GetHelpSystem,
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.HandleSystem(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     service.CmdHealth,
		Category: "monitor", Emoji: "🏥",
		DetailedFn: basic.GetHelpHealth,
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.HandleHealth(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     service.CmdLogs,
		Category: "monitor", Emoji: "📜",
		DetailedFn: basic.GetHelpLogs,
	}, func(s *service.BotService, m *tgbotapi.Message) { basic.HandleLogs(s, m, m.CommandArguments()) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "recover",
		Category: "recovery", Emoji: "🔄",
		DetailedFn: basic.GetHelpRecover,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleRecover(m)
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "recoverystatus",
		Category: "recovery", Emoji: "📊",
		DetailedFn: basic.GetHelpRecoveryStatus,
	}, func(s *service.BotService, m *tgbotapi.Message) {
		s.HandleRecoveryStatus(m)
	})

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "join",
		Category: "general", Emoji: "📢",
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
		Category: "files", Emoji: "📂",
		DetailedFn: basic.GetHelpLs,
	}, fmHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "mkdir",
		Category: "files", Emoji: "📁",
		DetailedFn: basic.GetHelpMkdir,
	}, fmHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "rm",
		Category: "files", Emoji: "🗑️",
		DetailedFn: basic.GetHelpRm,
	}, fmHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "mv",
		Category: "files", Emoji: "📦",
		DetailedFn: basic.GetHelpMv,
	}, fmHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "share",
		Category: "files", Emoji: "🔗",
		DetailedFn: basic.GetHelpShare,
	}, fmHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "find",
		Category: "files", Emoji: "🔍",
		DetailedFn: basic.GetHelpFind,
	}, fmHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "storages",
		Category: "storage", Emoji: "📋",
		DetailedFn: basic.GetHelpStorages,
	}, func(s *service.BotService, m *tgbotapi.Message) { storage.HandleStorages(s, m) })

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "setstorage",
		Category: "storage", Emoji: "⚙️",
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
		Name:     "extractaudio",
		Category: "media", Emoji: "🎵",
		DetailedFn: basic.GetHelpExtractAudio,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "compress",
		Category: "media", Emoji: "🗜️",
		DetailedFn: basic.GetHelpCompress,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "thumbnail",
		Category: "media", Emoji: "🖼️",
		DetailedFn: basic.GetHelpThumbnail,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "screenshots",
		Category: "media", Emoji: "📸",
		DetailedFn: basic.GetHelpScreenshots,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "subtitle",
		Category: "media", Emoji: "💬",
		DetailedFn: basic.GetHelpSubtitle,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "hardsub",
		Category: "media", Emoji: "🔥",
		DetailedFn: basic.GetHelpHardsub,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "rescale",
		Category: "media", Emoji: "📐",
		DetailedFn: basic.GetHelpRescale,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "convert",
		Category: "media", Emoji: "🔄",
		DetailedFn: basic.GetHelpConvert,
	}, mediaHandler)

	r.RegisterCommandWithInfo(router.CommandInfo{
		Name:     "mediainfo",
		Category: "media", Emoji: "ℹ️",
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
			lang := i18n.DefaultLang
			if cb.From != nil {
				lang = s.GetUserLanguage(cb.From.ID)
			}
			_, _ = s.Bot.Request(tgbotapi.NewCallback(cb.ID, i18n.T(lang, "menu_unsupported_here")))
		}
		return false
	}
	return true
}
