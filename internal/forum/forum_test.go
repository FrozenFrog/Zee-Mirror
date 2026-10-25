package forum

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"testing"

	"zee-mirror/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	chatID   = int64(-1003205657826)
	botTopic = 6231
)

type recorder struct{ params []url.Values }

func (r *recorder) Do(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	var values url.Values
	mediaType, ct, _ := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if mediaType == "multipart/form-data" {
		form, err := multipart.NewReader(bytes.NewReader(body), ct["boundary"]).ReadForm(1 << 20)
		if err != nil {
			return nil, err
		}
		values = form.Value
	} else {
		values, _ = url.ParseQuery(string(body))
	}
	if path.Base(req.URL.Path) != "getMe" {
		r.params = append(r.params, values)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1,"chat":{"id":1}}}`)),
	}, nil
}

func newBot(t *testing.T) (*tgbotapi.BotAPI, *Topics, *recorder) {
	t.Helper()
	topics := New([]config.AuthorizedChat{{ID: chatID}, {ID: chatID, TopicID: botTopic}})
	rec := &recorder{}
	bot, err := tgbotapi.NewBotAPIWithClient("token", tgbotapi.APIEndpoint, topics.Client(rec))
	if err != nil {
		t.Fatal(err)
	}
	return bot, topics, rec
}

func topicMessage(id, topic int) *tgbotapi.Message {
	return &tgbotapi.Message{
		MessageID:       id,
		MessageThreadID: topic,
		IsTopicMessage:  true,
		Chat:            tgbotapi.Chat{ID: chatID, IsForum: true},
	}
}

func TestSendGoesToConfiguredTopic(t *testing.T) {
	bot, _, rec := newBot(t)
	if _, err := bot.Send(tgbotapi.NewMessage(chatID, "status")); err != nil {
		t.Fatal(err)
	}
	if got := rec.params[0].Get("message_thread_id"); got != "6231" {
		t.Fatalf("message_thread_id = %q, want 6231", got)
	}
}

func TestReplyFollowsRepliedMessage(t *testing.T) {
	bot, topics, rec := newBot(t)
	topics.Remember(topicMessage(42, 77))

	known := tgbotapi.NewMessage(chatID, "reply")
	known.ReplyParameters.MessageID = 42
	unknown := tgbotapi.NewMessage(chatID, "reply")
	unknown.ReplyParameters.MessageID = 43
	for _, m := range []tgbotapi.MessageConfig{known, unknown} {
		if _, err := bot.Send(m); err != nil {
			t.Fatal(err)
		}
	}
	if got := rec.params[0].Get("message_thread_id"); got != "77" {
		t.Fatalf("known reply: message_thread_id = %q, want 77", got)
	}
	if got := rec.params[1].Get("message_thread_id"); got != "" {
		t.Fatalf("unknown reply: message_thread_id = %q, want none", got)
	}
}

func TestOtherChatsAndExplicitThreadUntouched(t *testing.T) {
	bot, _, rec := newBot(t)
	explicit := tgbotapi.NewMessage(chatID, "x")
	explicit.MessageThreadID = 5
	for _, m := range []tgbotapi.MessageConfig{tgbotapi.NewMessage(12345, "dm"), explicit} {
		if _, err := bot.Send(m); err != nil {
			t.Fatal(err)
		}
	}
	if got := rec.params[0].Get("message_thread_id"); got != "" {
		t.Fatalf("private chat: message_thread_id = %q, want none", got)
	}
	if got := rec.params[1]["message_thread_id"]; len(got) != 1 || got[0] != "5" {
		t.Fatalf("explicit thread: message_thread_id = %v, want [5]", got)
	}
}

func TestUploadGoesToConfiguredTopic(t *testing.T) {
	bot, _, rec := newBot(t)
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{Name: "a.txt", Bytes: bytes.Repeat([]byte("x"), 200<<10)})
	if _, err := bot.Send(doc); err != nil {
		t.Fatal(err)
	}
	if got := rec.params[0]["message_thread_id"]; len(got) != 1 || got[0] != "6231" {
		t.Fatalf("message_thread_id = %v, want [6231]", got)
	}
}

func TestGeneralIsNotATopic(t *testing.T) {
	general := &tgbotapi.Message{MessageID: 1, Chat: tgbotapi.Chat{ID: chatID, IsForum: true}}
	if got := TopicOf(general); got != GeneralTopic {
		t.Fatalf("TopicOf(General) = %d, want %d", got, GeneralTopic)
	}
	topics := New([]config.AuthorizedChat{{ID: chatID, TopicID: GeneralTopic}})
	if got := topics.topicFor(chatID, 0); got != 0 {
		t.Fatalf("General as default topic = %d, want 0", got)
	}
}
