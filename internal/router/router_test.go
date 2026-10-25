package router

import (
	"testing"

	"zee-mirror/handlers/download"
	"zee-mirror/internal/service"

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

func TestIsWizardPromptReply(t *testing.T) {
	if !isWizardPromptReply(stripMarkdownV2(download.PromptMirrorURL)) {
		t.Error("plain-text mirror URL prompt was not recognised")
	}
	if !isWizardPromptReply(stripMarkdownV2(download.PromptRename)) {
		t.Error("plain-text rename prompt was not recognised")
	}
	if isWizardPromptReply("some unrelated bot message") {
		t.Error("unrelated text was recognised as a wizard prompt")
	}
	if got := stripMarkdownV2("🔗 Gửi *URL* \\(http\\)\\."); got != "🔗 Gửi URL (http)." {
		t.Errorf("stripMarkdownV2 = %q", got)
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
