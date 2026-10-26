package downloader

import (
	"context"
	"time"

	"zee-mirror/internal/domain"
	"zee-mirror/pkg/i18n"
)

type ProgressUpdate struct {
	FileName    string
	Error       string
	Message     string
	Downloaded  int64
	Total       int64
	Speed       int64
	Progress    float64
	ETA         time.Duration
	Connections int
	// LocalError and LocalMessage are the translatable forms of Error and
	// Message. When set, Error/Message hold their DefaultLang rendering so
	// consumers without a language keep working; ErrorIn/MessageIn render
	// them for the task owner.
	LocalError   *i18n.Error
	LocalMessage *i18n.Error
}

// SetMessage sets Message to the i18n key, rendered in DefaultLang and kept
// translatable in LocalMessage.
func (p *ProgressUpdate) SetMessage(key string, args ...any) {
	p.LocalMessage = i18n.NewError(key, args...)
	p.Message = i18n.T(i18n.DefaultLang, key, args...)
}

// SetError is SetMessage for Error.
func (p *ProgressUpdate) SetError(key string, args ...any) {
	p.LocalError = i18n.NewError(key, args...)
	p.Error = i18n.T(i18n.DefaultLang, key, args...)
}

// MessageIn returns Message rendered in lang.
func (p ProgressUpdate) MessageIn(lang string) string {
	if p.LocalMessage != nil {
		return i18n.Message(lang, p.LocalMessage)
	}
	return p.Message
}

// ErrorIn returns Error rendered in lang.
func (p ProgressUpdate) ErrorIn(lang string) string {
	if p.LocalError != nil {
		return i18n.Message(lang, p.LocalError)
	}
	return p.Error
}

func (p ProgressUpdate) Found() bool {
	return p.Downloaded > 0 || p.Total > 0 || p.Speed > 0 || p.Progress > 0 || p.FileName != ""
}

type PlaylistMetadata struct {
	Title   string          `json:"title"`
	Entries []PlaylistEntry `json:"entries"`
}

type PlaylistEntry struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Index int    `json:"playlist_index"`
}

type DownloadEngine interface {
	Download(ctx context.Context, task *domain.Task, outputDir string, onProgress func(ProgressUpdate)) error
}

type FormatInfo struct {
	FPS  float64
	Size int64
}

type FormatLister interface {
	GetFormats(ctx context.Context, url string) (map[int]FormatInfo, error)
	GetPlaylistMetadata(ctx context.Context, url string) (*PlaylistMetadata, error)
	IsPlaylist(url string) bool
}

type MediaDownloader interface {
	DownloadEngine
	FormatLister
}
