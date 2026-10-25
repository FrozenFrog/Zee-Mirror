// Package forum keeps the bot's messages inside forum topics.
//
// Most send sites only know a chat ID (tasks outlive the command that started
// them), so instead of threading the topic through every one of them the HTTP
// client adds message_thread_id to outgoing send requests:
//   - a reply goes to the topic of the message it replies to, when known;
//   - anything else goes to the chat's configured topic (AUTHORIZED_CHATS
//     "chat:topic"), or stays in General when none is configured.
package forum

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"zee-mirror/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GeneralTopic is the ID t.me links use for a forum's General topic. The Bot
// API rejects it as message_thread_id; General is simply "no thread".
const GeneralTopic = 1

const (
	maxRemembered = 4096
	// tgbotapi writes every multipart field before the files, so the fields
	// we need always fit in this much of the body.
	multipartPeek = 64 << 10
)

type msgKey struct {
	chat int64
	msg  int
}

type Topics struct {
	defaults map[int64]int

	topics map[msgKey]int
	order  []msgKey
	next   int
	mu     sync.Mutex
}

// New uses the first "chat:topic" entry of each chat as its default topic.
func New(chats []config.AuthorizedChat) *Topics {
	t := &Topics{
		defaults: make(map[int64]int),
		topics:   make(map[msgKey]int, maxRemembered),
		order:    make([]msgKey, maxRemembered),
	}
	for _, c := range chats {
		if _, ok := t.defaults[c.ID]; !ok && c.TopicID > GeneralTopic {
			t.defaults[c.ID] = c.TopicID
		}
	}
	return t
}

// TopicOf returns the forum topic a message was posted in, GeneralTopic for
// General, or 0 outside forums.
func TopicOf(m *tgbotapi.Message) int {
	if m == nil || m.Chat.ID == 0 || !m.Chat.IsForum {
		return 0
	}
	if m.IsTopicMessage && m.MessageThreadID != 0 {
		return m.MessageThreadID
	}
	return GeneralTopic
}

// Remember records the topic of an incoming message so replies to it land in
// the same topic.
func (t *Topics) Remember(m *tgbotapi.Message) {
	topic := TopicOf(m)
	if t == nil || topic <= GeneralTopic {
		return
	}
	key := msgKey{m.Chat.ID, m.MessageID}

	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.topics[key]; !ok {
		delete(t.topics, t.order[t.next])
		t.order[t.next] = key
		t.next = (t.next + 1) % maxRemembered
	}
	t.topics[key] = topic
}

// topicFor picks the topic for a message sent to chatID, replying to replyTo
// (0 when not a reply).
func (t *Topics) topicFor(chatID int64, replyTo int) int {
	if replyTo != 0 {
		t.mu.Lock()
		defer t.mu.Unlock()
		// An unknown reply target is left to Telegram, which keeps a reply
		// next to the message it answers.
		return t.topics[msgKey{chatID, replyTo}]
	}
	return t.defaults[chatID]
}

// Client wraps next so send requests get a message_thread_id.
func (t *Topics) Client(next tgbotapi.HTTPClient) tgbotapi.HTTPClient {
	return &client{topics: t, next: next}
}

type client struct {
	topics *Topics
	next   tgbotapi.HTTPClient
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	if req.Body == nil || !takesThread(path.Base(req.URL.Path)) {
		return c.next.Do(req)
	}
	mediaType, ctParams, _ := mime.ParseMediaType(req.Header.Get("Content-Type"))
	var err error
	switch mediaType {
	case "application/x-www-form-urlencoded":
		err = c.routeForm(req)
	case "multipart/form-data":
		err = c.routeMultipart(req, ctParams["boundary"])
	}
	if err != nil {
		return nil, err
	}
	return c.next.Do(req)
}

func takesThread(method string) bool {
	return strings.HasPrefix(method, "send") ||
		strings.HasPrefix(method, "copyMessage") ||
		strings.HasPrefix(method, "forwardMessage")
}

func (c *client) routeForm(req *http.Request) error {
	body, err := io.ReadAll(req.Body)
	_ = req.Body.Close()
	if err != nil {
		return err
	}
	if values, perr := url.ParseQuery(string(body)); perr == nil && values.Get("message_thread_id") == "" {
		chatID, _ := strconv.ParseInt(values.Get("chat_id"), 10, 64)
		if topic := c.topics.topicFor(chatID, replyTarget(values.Get("reply_parameters"))); topic != 0 {
			values.Set("message_thread_id", strconv.Itoa(topic))
			body = []byte(values.Encode())
		}
	}
	setBody(req, io.NopCloser(bytes.NewReader(body)), int64(len(body)))
	return nil
}

func (c *client) routeMultipart(req *http.Request, boundary string) error {
	if boundary == "" {
		return nil
	}
	r := bufio.NewReaderSize(req.Body, multipartPeek)
	head, err := r.Peek(multipartPeek)
	if err != nil && err != io.EOF && err != bufio.ErrBufferFull {
		_ = req.Body.Close()
		return err
	}
	var body io.Reader = r
	if multipartField(head, "message_thread_id") == "" {
		chatID, _ := strconv.ParseInt(multipartField(head, "chat_id"), 10, 64)
		if topic := c.topics.topicFor(chatID, replyTarget(multipartField(head, "reply_parameters"))); topic != 0 {
			part := fmt.Sprintf("--%s\r\nContent-Disposition: form-data; name=\"message_thread_id\"\r\n\r\n%d\r\n", boundary, topic)
			body = io.MultiReader(strings.NewReader(part), r)
		}
	}
	// Close the original pipe so the tgbotapi writer goroutine is released
	// if the request fails before the body is drained.
	setBody(req, readCloser{body, req.Body}, -1)
	return nil
}

var fieldPatterns = func() map[string]*regexp.Regexp {
	patterns := make(map[string]*regexp.Regexp)
	for _, name := range []string{"chat_id", "message_thread_id", "reply_parameters"} {
		patterns[name] = regexp.MustCompile(`Content-Disposition: form-data; name="` + name + `"\r\n\r\n([^\r]*)\r\n`)
	}
	return patterns
}()

// multipartField extracts a plain form field written by mime/multipart.
func multipartField(head []byte, name string) string {
	if m := fieldPatterns[name].FindSubmatch(head); m != nil {
		return string(m[1])
	}
	return ""
}

func replyTarget(raw string) int {
	var p struct {
		MessageID int `json:"message_id"`
	}
	if raw == "" || json.Unmarshal([]byte(raw), &p) != nil {
		return 0
	}
	return p.MessageID
}

func setBody(req *http.Request, body io.ReadCloser, length int64) {
	req.Body = body
	req.ContentLength = length
	req.GetBody = nil
}

type readCloser struct {
	io.Reader
	closer io.Closer
}

func (rc readCloser) Close() error { return rc.closer.Close() }
