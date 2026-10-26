package basic

import (
	"sort"
	"strings"
	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"
)

type HelpCommandEntry struct {
	Name string
	// Description is the DefaultLang text, used when the catalog has no
	// i18n.CommandKey(Name) entry.
	Description string
	Category    string
	Emoji       string
	Aliases     []string
}

var registeredCommands []HelpCommandEntry

func SetRegisteredCommands(commands []HelpCommandEntry) {
	registeredCommands = commands
}

// helpPage renders the help page whose texts live under the catalog keys
// helptext_<page>_{title,desc,usage,example,extra}. example, when non-empty,
// is language-neutral command syntax used instead of the catalog example.
func helpPage(lang, page, example string) string {
	text := func(part string) string {
		s, _ := i18n.Lookup(lang, "helptext_"+page+"_"+part)
		return s
	}
	if example == "" {
		example = text("example")
	}
	return service.HelpDetailMessageLang(lang, text("title"), text("desc"), text("usage"), example, text("extra"))
}

// localizedDescription is the command description in lang, falling back to
// the registered (DefaultLang) description.
func localizedDescription(lang string, cmd HelpCommandEntry) string {
	if s, ok := i18n.Lookup(lang, i18n.CommandKey(cmd.Name)); ok {
		return s
	}
	return cmd.Description
}

func GetHelpMirror(lang string) string {
	return helpPage(lang, "mirror", "`/mirror https://example\\.com/file\\.zip`\n`/mirror \\-n \"My File\\.zip\" https://example\\.com/file\\.zip`\n`/mirror \\-z https://example\\.com/folder`\n`/mirror \\-uz \\-p secret https://file\\.rar`")
}

func GetHelpLeech(lang string) string {
	return helpPage(lang, "leech", "`/leech https://example\\.com/video\\.mp4`\n`/leech \\-n \"My Video\\.mp4\" https://example\\.com/video\\.mp4`\n`/leech \\-uz https://archive\\.rar`")
}

func GetHelpYTDLP(lang string) string {
	return helpPage(lang, "ytdlp", "`/ytdlp https://youtube\\.com/watch?v=xxxxx`\n`/ytdlp \\-q 1080p https://tiktok\\.com/@user/video/123`")
}

func GetHelpYTDLPLeech(lang string) string {
	return helpPage(lang, "ytdlpleech", "`/ytdlpleech https://youtube\\.com/watch?v=xxxxx`")
}

func GetHelpViking(lang string) string {
	return helpPage(lang, "viking", "`/viking https://example\\.com/file\\.zip`")
}

func GetHelpTorrent(lang string) string {
	return helpPage(lang, "torrent", "`/torrent magnet:?xt=urn:btih:xxxxx`\n`/torrent \\-z magnet:?xt=urn:btih:xxxxx`")
}

func GetHelpClone(lang string) string {
	return helpPage(lang, "clone", "`/clone https://drive\\.google\\.com/file/d/abc123`")
}

func GetHelpBatch(lang string) string {
	return helpPage(lang, "batch", "/batch\nhttps://file1\\.com/a\\.zip\nhttps://file2\\.com/b\\.zip\n\n`/batch \\-name DuLich \\-z`\n`https://photo\\.com/a\\.jpg`\n`https://photo\\.com/b\\.jpg`")
}

func GetHelpSearch(lang string) string {
	return helpPage(lang, "search", "`/search ubuntu 22\\.04`\n`/search archlinux iso`")
}

func GetHelpStatus(lang string) string {
	return helpPage(lang, "status", "`/status`")
}

func GetHelpStats(lang string) string {
	return helpPage(lang, "stats", "`/stats`")
}

func GetHelpSystem(lang string) string {
	return helpPage(lang, "system", "`/system`")
}

func GetHelpHealth(lang string) string {
	return helpPage(lang, "health", "`/health`")
}

func GetHelpLogs(lang string) string {
	return helpPage(lang, "logs", "")
}

func GetHelpPing(lang string) string {
	return helpPage(lang, "ping", "`/ping`")
}

func GetHelpSpeed(lang string) string {
	return helpPage(lang, "speed", "`/speed`")
}

func GetHelpLs(lang string) string {
	return helpPage(lang, "ls", "")
}

func GetHelpMkdir(lang string) string {
	return helpPage(lang, "mkdir", "`/mkdir Movies`\n`/mkdir Backup/2024`")
}

func GetHelpRm(lang string) string {
	return helpPage(lang, "rm", "`/rm old_file\\.zip`\n`/rm Temp/cache`")
}

func GetHelpMv(lang string) string {
	return helpPage(lang, "mv", "`/mv file\\.zip Backup/`\n`/mv old\\.mp4 new\\.mp4`")
}

func GetHelpShare(lang string) string {
	return helpPage(lang, "share", "`/share movie\\.mp4`\n`/share SharedFolder`")
}

func GetHelpFind(lang string) string {
	return helpPage(lang, "find", "")
}

func GetHelpExtractAudio(lang string) string {
	return helpPage(lang, "extractaudio", "")
}

func GetHelpCompress(lang string) string {
	return helpPage(lang, "compress", "")
}

func GetHelpThumbnail(lang string) string {
	return helpPage(lang, "thumbnail", "")
}

func GetHelpScreenshots(lang string) string {
	return helpPage(lang, "screenshots", "")
}

func GetHelpSubtitle(lang string) string {
	return helpPage(lang, "subtitle", "")
}

func GetHelpHardsub(lang string) string {
	return helpPage(lang, "hardsub", "")
}

func GetHelpRescale(lang string) string {
	return helpPage(lang, "rescale", "")
}

func GetHelpConvert(lang string) string {
	return helpPage(lang, "convert", "`/convert mp4`\n`/convert mp3`")
}

func GetHelpMediaInfo(lang string) string {
	return helpPage(lang, "mediainfo", "")
}

func GetHelpCancel(lang string) string {
	return helpPage(lang, "cancel", "`/cancel abc123`")
}

func GetHelpCancelAll(lang string) string {
	return helpPage(lang, "cancelall", "`/cancelall`")
}

func GetHelpStorages(lang string) string {
	return helpPage(lang, "storages", "`/storages`")
}

func GetHelpSetStorage(lang string) string {
	return helpPage(lang, "setstorage", "`/setstorage gdrive`")
}

func GetHelpAuthorize(lang string) string {
	return helpPage(lang, "authorize", "`/authorize 123456789 Zee`")
}

func GetHelpUnauthorize(lang string) string {
	return helpPage(lang, "unauthorize", "`/unauthorize 123456789`")
}

func GetHelpUsers(lang string) string {
	return helpPage(lang, "users", "`/users`")
}

func GetHelpSetAlertChannel(lang string) string {
	return helpPage(lang, "setalertchannel", "`/setalertchannel \\-1001234567890`")
}

func GetHelpSetLogChannel(lang string) string {
	return helpPage(lang, "setlogchannel", "`/setlogchannel -1001234567890`")
}

func GetHelpRecover(lang string) string {
	return helpPage(lang, "recover", "`/recover`")
}

func GetHelpRecoveryStatus(lang string) string {
	return helpPage(lang, "recoverystatus", "`/recoverystatus`")
}

func GetHelpSettings(lang string) string {
	return helpPage(lang, "settings", "`/settings`")
}

func getHelpAllCommands(lang string) string {
	return GenerateAllCommandsHelp(lang, registeredCommands)
}

func escapeMarkdownV2(text string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	result := text
	for _, ch := range specialChars {
		result = strings.ReplaceAll(result, ch, "\\"+ch)
	}
	return result
}

var helpCategoryOrder = []string{"download", "monitor", "files", "media", "task", "storage", "admin", "recovery", "general"}

// GenerateAllCommandsHelp lists commands by category with their descriptions in lang.
func GenerateAllCommandsHelp(lang string, commands []HelpCommandEntry) string {
	title := i18n.T(lang, "helptext_all_title")
	heading := i18n.T(lang, "helptext_all_heading")
	if len(commands) == 0 {
		return service.ProfessionalMessage(title, heading+"\n\n"+i18n.T(lang, "helptext_all_empty"))
	}

	byCategory := make(map[string][]HelpCommandEntry)
	for _, cmd := range commands {
		byCategory[cmd.Category] = append(byCategory[cmd.Category], cmd)
	}

	for cat := range byCategory {
		sort.Slice(byCategory[cat], func(i, j int) bool {
			return byCategory[cat][i].Name < byCategory[cat][j].Name
		})
	}

	var sb strings.Builder
	sb.WriteString(heading)
	sb.WriteString("\n")

	for _, cat := range helpCategoryOrder {
		cmds, ok := byCategory[cat]
		if !ok || len(cmds) == 0 {
			continue
		}
		sb.WriteString("\n")
		sb.WriteString(i18n.T(lang, "helptext_cat_"+cat))
		sb.WriteString("\n")
		for _, cmd := range cmds {
			sb.WriteString("• `/")
			sb.WriteString(cmd.Name)
			sb.WriteString("`")
			if len(cmd.Aliases) > 0 {
				sb.WriteString(" \\(")
				for i, alias := range cmd.Aliases {
					if i > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString("/")
					sb.WriteString(escapeMarkdownV2(alias))
				}
				sb.WriteString("\\)")
			}
			sb.WriteString(" ─ ")
			sb.WriteString(escapeMarkdownV2(localizedDescription(lang, cmd)))
			sb.WriteString("\n")
		}
	}

	return service.ProfessionalMessage(title, sb.String())
}

func GetHelpNotify(lang string) string {
	return helpPage(lang, "notify", "")
}
