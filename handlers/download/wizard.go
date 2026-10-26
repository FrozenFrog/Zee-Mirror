package download

import (
	"strings"
	"zee-mirror/internal/service"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// PromptMirrorURL and PromptRename are the wizard's ForceReply prompts in
// DefaultLang. Use IsWizardPromptReply to recognise a reply in any language.
var (
	PromptMirrorURL = i18n.T(i18n.DefaultLang, "wizard_prompt_url")
	PromptRename    = i18n.T(i18n.DefaultLang, "wizard_prompt_rename")
)

// Markers of prompts sent before the wizard was translated (Indonesian), so
// replies to old messages still route.
var (
	legacyURLPromptMarkers    = []string{"Silakan kirim URL"}
	legacyRenamePromptMarkers = []string{"Silakan kirim Nama Baru"}
)

// IsWizardPromptReply reports whether replyText (a prompt as Telegram returns
// it: plain text without MarkdownV2 markup) is one of the wizard's ForceReply
// prompts, in any supported language.
func IsWizardPromptReply(replyText string) bool {
	return isWizardPrompt(replyText, "wizard_prompt_url", legacyURLPromptMarkers) ||
		isWizardPrompt(replyText, "wizard_prompt_rename", legacyRenamePromptMarkers)
}

func isWizardPrompt(replyText, key string, legacy []string) bool {
	for _, l := range i18n.Languages {
		if plain := plainMarkdownV2(i18n.T(l.Code, key)); plain != "" && strings.Contains(replyText, plain) {
			return true
		}
	}
	for _, marker := range legacy {
		if strings.Contains(replyText, marker) {
			return true
		}
	}
	return false
}

// plainMarkdownV2 drops MarkdownV2 markup and escapes, giving the text
// Telegram reports for a message sent with that markup.
func plainMarkdownV2(text string) string {
	var sb strings.Builder
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		switch runes[i] {
		case '\\':
			if i+1 < len(runes) {
				i++
				sb.WriteRune(runes[i])
			}
		case '*', '_', '~', '`', '|':
		default:
			sb.WriteRune(runes[i])
		}
	}
	return strings.TrimSpace(sb.String())
}

func HandleMirrorWizard(s *service.BotService, message *tgbotapi.Message) {
	lang := s.GetUserLanguage(message.From.ID)
	sendMirrorWizard(s, message.Chat.ID, lang)
}

func sendMirrorWizard(s *service.BotService, chatID int64, lang string) {
	msg := tgbotapi.NewMessage(chatID, wizardWelcomeText(lang))
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.ReplyMarkup = wizardStartKeyboard(lang)
	_, _ = s.Bot.Send(msg)
}

func wizardWelcomeText(lang string) string {
	return service.ProfessionalMessage("MIRROR WIZARD", i18n.T(lang, "wizard_welcome"))
}

func wizardStartKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "wizard_btn_input_url"), "wizard:input_url")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "wizard_btn_cancel"), "dashboard:close")),
	)
}

func wizardOptionsKeyboard(lang string, isZip bool) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📦 Zip: "+utils.BoolToEmoji(isZip), "wizard:toggle_zip"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "wizard_btn_start"), "wizard:start"),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "wizard_btn_cancel"), "dashboard:close"),
		),
	)
}

// wizardOptionsText renders the options card. The "URL:" and "Zip:" labels
// are identical in every language: toggle/start read the state back from
// the message text.
func wizardOptionsText(lang, url string, isZip bool) string {
	return i18n.T(lang, "wizard_options", utils.EscapeMarkdownV2(url), utils.BoolToEmoji(isZip))
}

func HandleMirrorWizardCallback(s *service.BotService, callback *tgbotapi.CallbackQuery, parts []string) {
	action := parts[1]
	lang := s.GetUserLanguage(callback.From.ID)

	switch action {
	case "input_url":
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, i18n.T(lang, "wizard_prompt_url"))
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}
		_, _ = s.Bot.Send(msg)

		_, _ = s.Bot.Request(tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID))

	case "back_main":
		sendMirrorWizard(s, callback.Message.Chat.ID, lang)

		edit := tgbotapi.NewEditMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, wizardWelcomeText(lang), wizardStartKeyboard(lang))
		edit.ParseMode = tgbotapi.ModeMarkdownV2
		_, _ = s.Bot.Send(edit)

	case "toggle_zip":
		toggleWizardOption(s, callback, lang, "zip")

	case "opt_rename":
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, i18n.T(lang, "wizard_prompt_rename"))
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}

		_, _ = s.Bot.Send(msg)

	case "start":
		executeWizardTask(s, callback, lang)
	}
}

func HandleWizardInput(s *service.BotService, message *tgbotapi.Message) {
	replyText := message.ReplyToMessage.Text

	if isWizardPrompt(replyText, "wizard_prompt_url", legacyURLPromptMarkers) {
		lang := s.GetUserLanguage(message.From.ID)
		url := message.Text
		if !strings.HasPrefix(url, "http") && !strings.HasPrefix(url, "magnet") {
			s.Reply(message, i18n.T(lang, "wizard_invalid_url"))
			return
		}

		showWizardOptionsMenu(s, message.Chat.ID, lang, url, false)
	}
}

func showWizardOptionsMenu(s *service.BotService, chatID int64, lang, url string, isZip bool) {
	msg := tgbotapi.NewMessage(chatID, wizardOptionsText(lang, url, isZip))
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.LinkPreviewOptions = tgbotapi.LinkPreviewOptions{IsDisabled: true}
	msg.ReplyMarkup = wizardOptionsKeyboard(lang, isZip)
	_, _ = s.Bot.Send(msg)
}

// parseWizardOptions reads the URL and zip flag back from an options card.
func parseWizardOptions(text string) (url string, isZip bool) {
	for _, line := range strings.Split(text, "\n") {
		if idx := strings.Index(line, "URL:"); idx != -1 {
			url = strings.TrimSpace(line[idx+len("URL:"):])
		}
		if strings.Contains(line, "Zip:") && strings.Contains(line, "✅") {
			isZip = true
		}
	}
	return url, isZip
}

func toggleWizardOption(s *service.BotService, callback *tgbotapi.CallbackQuery, lang, option string) {
	url, isZip := parseWizardOptions(callback.Message.Text)

	if option == "zip" {
		isZip = !isZip
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, wizardOptionsText(lang, url, isZip), wizardOptionsKeyboard(lang, isZip))
	edit.ParseMode = tgbotapi.ModeMarkdownV2
	edit.LinkPreviewOptions = tgbotapi.LinkPreviewOptions{IsDisabled: true}
	_, _ = s.Bot.Send(edit)
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
}

func executeWizardTask(s *service.BotService, callback *tgbotapi.CallbackQuery, lang string) {
	url, isZip := parseWizardOptions(callback.Message.Text)

	if url == "" {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "wizard_toast_no_url")))
		return
	}

	fileName := utils.GetFileNameFromURL(url)
	task, err := s.TaskManager.CreateTask(service.TypeMirror, url, fileName, callback.Message.Chat.ID, callback.Message.MessageID, 0, callback.From.ID, isZip, false, "", "", 0, "", false)

	if err != nil {
		_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "wizard_toast_start_failed")))
		return
	}

	s.UpdateSharedDashboard(callback.Message.Chat.ID, true)
	s.HandleAutoDelete(task)

	_, _ = s.Bot.Request(tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID))
	_, _ = s.Bot.Request(tgbotapi.NewCallback(callback.ID, i18n.T(lang, "wizard_toast_started")))
}
