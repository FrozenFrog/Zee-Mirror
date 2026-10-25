package userbot

import "testing"

func TestParseMessageLink(t *testing.T) {
	valid := map[string]messageRef{
		"https://t.me/burning_high/81822":     {Username: "burning_high", MsgID: 81822},
		"t.me/burning_high/81822?single":      {Username: "burning_high", MsgID: 81822},
		"https://t.me/s/burning_high/81822":   {Username: "burning_high", MsgID: 81822},
		"https://t.me/forumname/12/345":       {Username: "forumname", MsgID: 345},
		"https://t.me/c/3205657826/6231/7000": {ChannelID: 3205657826, MsgID: 7000},
		"https://t.me/c/3205657826/99/":       {ChannelID: 3205657826, MsgID: 99},
	}
	for link, want := range valid {
		got, err := parseMessageLink(link)
		if err != nil || got != want {
			t.Errorf("parseMessageLink(%q) = %+v, %v; want %+v", link, got, err, want)
		}
	}

	for _, link := range []string{"https://t.me/burning_high", "https://t.me/c/abc/5", "https://example.com/a/1", "https://t.me/name/abc"} {
		if _, err := parseMessageLink(link); err == nil {
			t.Errorf("parseMessageLink(%q): expected an error", link)
		}
	}
}
