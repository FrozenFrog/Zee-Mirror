package userbot

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"zee-mirror/internal/config"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
)

type UserBot struct {
	Client  *telegram.Client
	Config  *config.Config
	Manager *peers.Manager
	Sender  *message.Sender
	Context context.Context
	Cancel  context.CancelFunc
	Mu      sync.RWMutex
	Started bool
}

var (
	instance *UserBot
	once     sync.Once
)

func GetInstance(cfg *config.Config) *UserBot {
	once.Do(func() {
		instance = &UserBot{
			Config: cfg,
		}
	})
	return instance
}

func (u *UserBot) Start() error {
	u.Mu.Lock()
	defer u.Mu.Unlock()

	if u.Started {
		return nil
	}

	if u.Config.AppID == 0 || u.Config.AppHash == "" || u.Config.UserSessionString == "" {
		slog.Warn("🚫 Userbot credentials missing (APP_ID, APP_HASH, or USER_SESSION_STRING). Userbot disabled.")
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	u.Context = ctx
	u.Cancel = cancel

	storage := &session.StorageMemory{}
	if u.Config.UserSessionString != "" {
		data, decodeErr := base64.StdEncoding.DecodeString(u.Config.UserSessionString)
		if decodeErr == nil {
			if storeErr := storage.StoreSession(context.Background(), data); storeErr != nil {
				slog.Warn("Failed to store session to memory", "error", storeErr)
			}
		} else {
			slog.Warn("Invalid base64 session string", "error", decodeErr)
		}
	}

	client := telegram.NewClient(u.Config.AppID, u.Config.AppHash, telegram.Options{
		SessionStorage: storage,
		Logger:         nil,
	})

	u.Client = client

	go func() {
		if err := client.Run(ctx, func(ctx context.Context) error {
			status, authErr := client.Auth().Status(ctx)
			if authErr != nil {
				return authErr
			}
			if !status.Authorized {
				slog.Error("❌ Userbot session is invalid or expired. Please generate a new session string using gotd.")
				return fmt.Errorf("userbot unauthorized")
			}

			u.Manager = peers.Options{}.Build(client.API())

			slog.Info("🔄 Userbot syncing dialogs...")

			if _, err := client.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
				Limit:      100,
				OffsetPeer: &tg.InputPeerEmpty{},
			}); err != nil {
				slog.Warn("Failed to sync dialogs", "error", err)
			}

			self, _ := client.Self(ctx)
			slog.Info("✅ Userbot Connected", "user", self.Username, "id", self.ID)

			u.Mu.Lock()
			u.Started = true
			u.Mu.Unlock()

			<-ctx.Done()
			return ctx.Err()
		}); err != nil {
			slog.Error("Userbot stopped", "error", err)
			u.Mu.Lock()
			u.Started = false
			u.Mu.Unlock()
		}
	}()

	return nil
}

func (u *UserBot) Stop() {
	u.Mu.Lock()
	defer u.Mu.Unlock()
	if u.Started && u.Cancel != nil {
		slog.Info("Stopping Userbot session gracefully...")
		u.Cancel()
		u.Started = false
	}
}

func (u *UserBot) DownloadFile(link string, outputDir string, onProgress func(downloaded, total int64)) (string, error) {
	u.Mu.RLock()
	if !u.Started {
		u.Mu.RUnlock()
		return "", fmt.Errorf("userbot chưa được khởi động")
	}
	u.Mu.RUnlock()

	ref, err := parseMessageLink(link)
	if err != nil {
		return "", err
	}
	msgID := ref.MsgID

	ctx, cancel := context.WithTimeout(u.Context, 10*time.Minute)
	defer cancel()

	msg, err := u.fetchMessage(ctx, ref)
	if err != nil {
		return "", err
	}
	media := msg.Media

	if media == nil {
		return "", fmt.Errorf("không tìm thấy media trong tin nhắn")
	}

	dl := downloader.NewDownloader()

	outPath := filepath.Join(outputDir, fmt.Sprintf("userbot_dl_%d", msgID))
	var totalSize int64

	if doc, ok := media.(*tg.MessageMediaDocument); ok {
		if d, ok := doc.Document.(*tg.Document); ok {
			totalSize = d.Size
			for _, attr := range d.Attributes {
				if fn, ok := attr.(*tg.DocumentAttributeFilename); ok {
					outPath = filepath.Join(outputDir, fn.FileName)
					break
				}
			}
		}
	}

	f, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("không tạo được file: %v", err)
	}
	defer f.Close()

	location, err := getFileLocation(media)
	if err != nil {
		return "", err
	}

	dlCtx := context.Background()

	pw := &ProgressWriter{
		w:          f,
		onProgress: onProgress,
		total:      totalSize,
	}

	if _, err := dl.Download(u.Client.API(), location).Stream(dlCtx, pw); err != nil {
		return "", fmt.Errorf("tải xuống thất bại: %v", err)
	}

	return outPath, nil
}

// messageRef is a parsed t.me message link: either a public username or the
// numeric id of a private channel (t.me/c/<id>/...), plus the message id.
type messageRef struct {
	Username  string
	ChannelID int64
	MsgID     int
}

// parseMessageLink accepts t.me/<username>/<msg>, t.me/c/<id>/<msg> and their
// forum topic forms (t.me/<username>/<topic>/<msg>, t.me/c/<id>/<topic>/<msg>).
// The message id is always the last path segment.
func parseMessageLink(link string) (messageRef, error) {
	var ref messageRef
	link = strings.TrimSpace(link)
	if i := strings.IndexAny(link, "?#"); i >= 0 {
		link = link[:i]
	}
	idx := strings.Index(link, "t.me/")
	if idx < 0 {
		return ref, fmt.Errorf("định dạng link không hợp lệ")
	}
	parts := strings.Split(strings.Trim(link[idx+len("t.me/"):], "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return ref, fmt.Errorf("định dạng link không hợp lệ")
	}
	if parts[0] == "s" {
		parts = parts[1:]
	}

	switch {
	case len(parts) >= 3 && parts[0] == "c":
		id, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return ref, fmt.Errorf("ID channel không hợp lệ")
		}
		ref.ChannelID = id
	case len(parts) >= 2:
		ref.Username = parts[0]
	default:
		return ref, fmt.Errorf("link thiếu ID tin nhắn")
	}

	msgID, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil || msgID <= 0 {
		return ref, fmt.Errorf("ID tin nhắn không hợp lệ")
	}
	ref.MsgID = msgID
	return ref, nil
}

var errDialogFound = errors.New("dialog found")

// resolvePrivateChannel finds a channel by its numeric id. The peer cache is
// in memory, so after a restart the access hash is only known once the
// channel is seen again; the dialog list provides it for joined channels.
func (u *UserBot) resolvePrivateChannel(ctx context.Context, id int64) (peers.Channel, error) {
	if ch, err := u.Manager.ResolveChannelID(ctx, id); err == nil {
		return ch, nil
	}

	var input *tg.InputChannel
	err := query.GetDialogs(u.Client.API()).BatchSize(100).ForEach(ctx, func(_ context.Context, e dialogs.Elem) error {
		if p, ok := e.Peer.(*tg.InputPeerChannel); ok && p.ChannelID == id {
			input = &tg.InputChannel{ChannelID: p.ChannelID, AccessHash: p.AccessHash}
			return errDialogFound
		}
		return nil
	})
	if err != nil && !errors.Is(err, errDialogFound) {
		return peers.Channel{}, err
	}
	if input == nil {
		return peers.Channel{}, fmt.Errorf("userbot chưa tham gia channel/nhóm này")
	}
	return u.Manager.GetChannel(ctx, input)
}

// fetchMessage loads the message a link points to. Channel and supergroup
// messages must go through channels.getMessages: messages.getMessages only
// reads the account's own message box and returns an empty message for
// channel ids instead of an error.
func (u *UserBot) fetchMessage(ctx context.Context, ref messageRef) (*tg.Message, error) {
	var peer peers.Peer
	if ref.ChannelID != 0 {
		ch, err := u.resolvePrivateChannel(ctx, ref.ChannelID)
		if err != nil {
			return nil, fmt.Errorf("không xác định được channel riêng tư: %v (hãy chắc chắn userbot đã tham gia)", err)
		}
		peer = ch
	} else {
		p, err := u.Manager.ResolveDomain(ctx, ref.Username)
		if err != nil {
			return nil, fmt.Errorf("không xác định được username %s: %v", ref.Username, err)
		}
		peer = p
	}

	ids := []tg.InputMessageClass{&tg.InputMessageID{ID: ref.MsgID}}
	var (
		res tg.MessagesMessagesClass
		err error
	)
	if ch, ok := peer.InputPeer().(*tg.InputPeerChannel); ok {
		res, err = u.Client.API().ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{
			Channel: &tg.InputChannel{ChannelID: ch.ChannelID, AccessHash: ch.AccessHash},
			ID:      ids,
		})
	} else {
		res, err = u.Client.API().MessagesGetMessages(ctx, ids)
	}
	if err != nil {
		return nil, fmt.Errorf("không lấy được tin nhắn: %v", err)
	}

	modified, ok := res.AsModified()
	if !ok {
		return nil, fmt.Errorf("không lấy được tin nhắn")
	}
	for _, m := range modified.GetMessages() {
		if msg, ok := m.(*tg.Message); ok {
			return msg, nil
		}
	}
	return nil, fmt.Errorf("tin nhắn %d không tồn tại hoặc userbot không xem được", ref.MsgID)
}

type ProgressWriter struct {
	w          io.Writer
	onProgress func(downloaded, total int64)
	total      int64
	current    int64
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.w.Write(p)
	if n > 0 {
		pw.current += int64(n)
		if pw.onProgress != nil {
			pw.onProgress(pw.current, pw.total)
		}
	}
	return n, err
}

func getFileLocation(media tg.MessageMediaClass) (tg.InputFileLocationClass, error) {
	switch m := media.(type) {
	case *tg.MessageMediaDocument:
		if doc, ok := m.Document.(*tg.Document); ok {
			return &tg.InputDocumentFileLocation{
				ID:            doc.ID,
				AccessHash:    doc.AccessHash,
				FileReference: doc.FileReference,
				ThumbSize:     "",
			}, nil
		}
	case *tg.MessageMediaPhoto:
		if photo, ok := m.Photo.(*tg.Photo); ok {
			return &tg.InputPhotoFileLocation{
				ID:            photo.ID,
				AccessHash:    photo.AccessHash,
				FileReference: photo.FileReference,
				ThumbSize:     "y",
			}, nil
		}
	}
	return nil, fmt.Errorf("loại media này không hỗ trợ tải xuống")
}

func (u *UserBot) JoinChat(hash string) (string, error) {
	u.Mu.RLock()
	if !u.Started {
		u.Mu.RUnlock()
		return "", fmt.Errorf("userbot chưa được khởi động")
	}
	u.Mu.RUnlock()

	ctx, cancel := context.WithTimeout(u.Context, 30*time.Second)
	defer cancel()

	res, err := u.Client.API().MessagesImportChatInvite(ctx, hash)
	if err != nil {
		if strings.Contains(err.Error(), "USER_ALREADY_PARTICIPANT") {
			return "Already joined", nil
		}
		return "", fmt.Errorf("không tham gia được chat: %w", err)
	}

	var title string
	if result, ok := res.(*tg.MessagesChatInviteJoinResultOk); ok {
		if u, ok := result.Updates.(*tg.Updates); ok {
			for _, chat := range u.Chats {
				switch c := chat.(type) {
				case *tg.Chat:
					title = c.Title
				case *tg.Channel:
					title = c.Title
				}
			}
		}
	}

	if title == "" {
		title = "Chat không rõ tên"
	}

	return fmt.Sprintf("Đã tham gia thành công: %s", title), nil
}
