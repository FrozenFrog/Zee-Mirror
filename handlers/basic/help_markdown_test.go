package basic

import (
	"fmt"
	"strings"
	"testing"
)

// checkMarkdownV2 reports the first construct Telegram would reject in a
// MarkdownV2 text: an unescaped reserved character or an unclosed entity.
// It covers the subset the help pages use (bold, italic, strikethrough,
// inline code), so links and pre blocks are reported as errors.
func checkMarkdownV2(text string) error {
	const reserved = "_*[]()~`>#+-=|{}.!"
	runes := []rune(text)
	open := map[rune]int{}
	inCode := false
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		if c == '\\' {
			if i+1 >= len(runes) || runes[i+1] > 126 {
				return fmt.Errorf("dangling backslash at %d", i)
			}
			i++
			continue
		}
		if inCode {
			if c == '`' {
				inCode = false
			}
			continue
		}
		switch c {
		case '`':
			inCode = true
		case '*', '_', '~':
			open[c] ^= 1
		default:
			if strings.ContainsRune(reserved, c) {
				return fmt.Errorf("unescaped %q at %d: %q", c, i, around(runes, i))
			}
		}
	}
	if inCode {
		return fmt.Errorf("unclosed inline code")
	}
	for c, n := range open {
		if n != 0 {
			return fmt.Errorf("unclosed %q entity", c)
		}
	}
	return nil
}

func around(runes []rune, i int) string {
	return string(runes[max(0, i-20):min(len(runes), i+20)])
}

func TestCheckMarkdownV2(t *testing.T) {
	valid := []string{"*a* `x.y` \\. _b_", "`a\\`b`", "~s~"}
	invalid := []string{"a.b", "*a", "`a", "(x)", "a\\"}
	for _, s := range valid {
		if err := checkMarkdownV2(s); err != nil {
			t.Errorf("%q: unexpected error %v", s, err)
		}
	}
	for _, s := range invalid {
		if checkMarkdownV2(s) == nil {
			t.Errorf("%q: expected an error", s)
		}
	}
}

func TestHelpPagesAreValidMarkdownV2(t *testing.T) {
	pages := map[string]func() string{
		"mirror": GetHelpMirror, "leech": GetHelpLeech, "ytdlp": GetHelpYTDLP,
		"ytdlpleech": GetHelpYTDLPLeech, "viking": GetHelpViking, "torrent": GetHelpTorrent,
		"clone": GetHelpClone, "batch": GetHelpBatch, "search": GetHelpSearch,
		"status": GetHelpStatus, "stats": GetHelpStats, "system": GetHelpSystem,
		"health": GetHelpHealth, "logs": GetHelpLogs, "ping": GetHelpPing,
		"speed": GetHelpSpeed, "ls": GetHelpLs, "mkdir": GetHelpMkdir, "rm": GetHelpRm,
		"mv": GetHelpMv, "share": GetHelpShare, "find": GetHelpFind,
		"extractaudio": GetHelpExtractAudio, "compress": GetHelpCompress,
		"thumbnail": GetHelpThumbnail, "screenshots": GetHelpScreenshots,
		"subtitle": GetHelpSubtitle, "hardsub": GetHelpHardsub, "rescale": GetHelpRescale,
		"convert": GetHelpConvert, "mediainfo": GetHelpMediaInfo, "cancel": GetHelpCancel,
		"cancelall": GetHelpCancelAll, "storages": GetHelpStorages,
		"setstorage": GetHelpSetStorage, "authorize": GetHelpAuthorize,
		"unauthorize": GetHelpUnauthorize, "users": GetHelpUsers,
		"setalertchannel": GetHelpSetAlertChannel, "setlogchannel": GetHelpSetLogChannel,
		"recover": GetHelpRecover, "recoverystatus": GetHelpRecoveryStatus,
		"settings": GetHelpSettings, "notify": GetHelpNotify,
		"all_empty": func() string { return GenerateAllCommandsHelp(nil) },
		"all": func() string {
			return GenerateAllCommandsHelp([]HelpCommandEntry{
				{Name: "ytdlp_leech", Description: "Tải (YT-DLP) về Telegram.", Category: "download", Aliases: []string{"yl", "y_l"}},
				{Name: "stats", Description: "Thống kê", Category: "monitor"},
			})
		},
	}
	for name, page := range pages {
		if err := checkMarkdownV2(page()); err != nil {
			t.Errorf("help page %s: %v", name, err)
		}
	}
}
