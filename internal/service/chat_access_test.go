package service_test

import (
	"errors"
	"strings"
	"testing"

	"zee-mirror/internal/config"
	"zee-mirror/internal/repository/mocks"
	"zee-mirror/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	groupID  = int64(-1003205657826)
	ownerID  = int64(1)
	memberID = int64(555)
)

func newAccessService() *service.BotService {
	repo := new(mocks.MockRepository)
	repo.On("GetByID", mock.Anything, memberID).Return(nil, errors.New("not found"))
	cfg := &config.Config{
		OwnerID:         ownerID,
		AuthorizedChats: config.ParseAuthorizedChats("-1003205657826, -1003205657826:6231"),
	}
	return &service.BotService{
		Config: cfg,
		Auth:   service.NewAuthService(cfg, repo, repo),
		Bot:    &tgbotapi.BotAPI{Self: tgbotapi.User{ID: 99, UserName: "ZeeBot"}},
	}
}

// groupMessage builds a message in groupID; topic 0 means General.
func groupMessage(from int64, topic int, text string) *tgbotapi.Message {
	m := &tgbotapi.Message{
		MessageID:       10,
		From:            &tgbotapi.User{ID: from},
		Chat:            tgbotapi.Chat{ID: groupID, Type: "supergroup", IsForum: true},
		Text:            text,
		MessageThreadID: topic,
		IsTopicMessage:  topic != 0,
	}
	if cmd, _, _ := strings.Cut(text, " "); strings.HasPrefix(text, "/") {
		m.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(cmd)}}
	}
	return m
}

func TestAdmitMessageAuthorizedChats(t *testing.T) {
	svc := newAccessService()
	tests := []struct {
		name string
		msg  *tgbotapi.Message
		want service.Access
	}{
		{"topic needs no mention", groupMessage(memberID, 6231, "/mirror https://x"), service.AccessGranted},
		{"general without mention is ignored", groupMessage(ownerID, 0, "/mirror https://x"), service.AccessIgnore},
		{"general with mention", groupMessage(memberID, 0, "/mirror@ZeeBot https://x"), service.AccessGranted},
		{"mention is case-insensitive", groupMessage(memberID, 0, "/mirror@zeebot https://x"), service.AccessGranted},
		{"other topic with mention", groupMessage(memberID, 7, "/mirror@ZeeBot https://x"), service.AccessGranted},
		{"command for another bot", groupMessage(ownerID, 6231, "/mirror@OtherBot https://x"), service.AccessIgnore},
		{"plain chatter in general", groupMessage(memberID, 0, "hello"), service.AccessIgnore},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, svc.AdmitMessage(tt.msg))
		})
	}
}

func TestAdmitMessageStripsMention(t *testing.T) {
	svc := newAccessService()
	m := groupMessage(memberID, 0, "@ZeeBot magnet:?xt=urn:btih:abc")
	assert.Equal(t, service.AccessGranted, svc.AdmitMessage(m))
	assert.Equal(t, "magnet:?xt=urn:btih:abc", m.Text)
}

func TestAdmitMessageOutsideAuthorizedChats(t *testing.T) {
	svc := newAccessService()
	dm := func(from int64, text string) *tgbotapi.Message {
		m := groupMessage(from, 0, text)
		m.Chat = tgbotapi.Chat{ID: from, Type: "private"}
		return m
	}
	assert.Equal(t, service.AccessGranted, svc.AdmitMessage(dm(ownerID, "/mirror x")))
	assert.Equal(t, service.AccessDenied, svc.AdmitMessage(dm(memberID, "/mirror x")))
	assert.Equal(t, service.AccessGranted, svc.AdmitMessage(dm(memberID, "/start")))
	assert.False(t, svc.CanUse(dm(memberID, "/stats")))
	assert.True(t, svc.CanUse(groupMessage(memberID, 6231, "/stats")))
}

func TestParseAuthorizedChats(t *testing.T) {
	got := config.ParseAuthorizedChats("-100123 -100123:6231,bad, -100456:0 -100789:x")
	assert.Equal(t, []config.AuthorizedChat{{ID: -100123}, {ID: -100123, TopicID: 6231}}, got)
}
