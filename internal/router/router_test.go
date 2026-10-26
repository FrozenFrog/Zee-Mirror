package router

import (
	"testing"

	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestRouter_HandleMessage(t *testing.T) {
	r := NewRouter(nil)
	called := false
	r.RegisterCommandWithInfo(CommandInfo{Name: "testcmd"}, func(_ *service.BotService, _ *tgbotapi.Message) {
		called = true
	})

	msg := &tgbotapi.Message{
		Text: "/testcmd",
		Entities: []tgbotapi.MessageEntity{
			{Type: "bot_command", Offset: 0, Length: 8},
		},
	}

	r.HandleMessage(msg)
	if !called {
		t.Error("Command handler for 'testcmd' was not called")
	}
}

func TestCommandDescriptions(t *testing.T) {
	if got := i18n.CommandKey("ytdlp-Leech.2"); got != "cmd_ytdlp_leech_2" {
		t.Errorf("CommandKey = %q", got)
	}
	r := NewRouter(nil)
	noop := func(_ *service.BotService, _ *tgbotapi.Message) {}
	r.RegisterCommandWithInfo(CommandInfo{Name: "mirror"}, noop)
	r.RegisterCommandWithInfo(CommandInfo{Name: "custom", Description: "Custom"}, noop)
	infos := map[string]CommandInfo{}
	for _, c := range r.GetAllCommandsFlat() {
		infos[c.Name] = c
	}
	if want := i18n.T(i18n.DefaultLang, "cmd_mirror"); infos["mirror"].Description != want {
		t.Errorf("mirror Description = %q, want %q", infos["mirror"].Description, want)
	}
	if got := infos["mirror"].LocalizedDescription("en"); got != i18n.T("en", "cmd_mirror") {
		t.Errorf("mirror en = %q", got)
	}
	if got := infos["custom"].LocalizedDescription("ja"); got != "Custom" {
		t.Errorf("custom fallback = %q", got)
	}
}

func TestRouter_HandleCallback(t *testing.T) {
	r := NewRouter(nil)
	called := false
	r.RegisterCallback("testcb", func(_ *service.BotService, _ *tgbotapi.CallbackQuery) {
		called = true
	})

	cb := &tgbotapi.CallbackQuery{
		Data: "testcb:action:id",
	}

	r.HandleCallback(cb)
	if !called {
		t.Error("Callback handler for 'testcb' was not called")
	}
}
