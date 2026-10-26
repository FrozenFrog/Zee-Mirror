package domain

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"zee-mirror/pkg/i18n"
)

var (
	ErrUnauthorized   = errors.New("unauthorized: access denied")
	ErrNotFound       = errors.New("resource not found")
	ErrDuplicateTask  = errors.New("active task already exists")
	ErrLimitExceeded  = errors.New("daily usage limit exceeded")
	ErrInvalidInput   = errors.New("invalid input provided")
	ErrInternal       = errors.New("internal server error")
	ErrStorageFull    = errors.New("storage is full")
	ErrInvalidQuality = errors.New("invalid media quality")
	ErrTaskCancelled  = errors.New("task was cancelled")
	ErrTaskFailed     = errors.New("task execution failed")
	ErrLinkExpired    = errors.New("link expired or invalid")

	ErrNetwork  = errors.New("network error")
	ErrStorage  = errors.New("storage error")
	ErrAuth     = errors.New("authentication error")
	ErrQuota    = errors.New("quota exceeded")
	ErrExternal = errors.New("external service error")
)

type NetworkError struct {
	Err error
	URL string
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error accessing %s: %v", e.URL, e.Err)
}

func (e *NetworkError) Unwrap() error {
	return ErrNetwork
}

type StorageError struct {
	Err  error
	Path string
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("storage error at %s: %v", e.Path, e.Err)
}

func (e *StorageError) Unwrap() error {
	return ErrStorage
}

type QuotaError struct {
	Err    error
	UserID int64
	Limit  int
}

func (e *QuotaError) Error() string {
	return fmt.Sprintf("quota exceeded for user %d (limit: %d): %v", e.UserID, e.Limit, e.Err)
}

func (e *QuotaError) Unwrap() error {
	return ErrQuota
}

type ExternalError struct {
	Err      error
	Tool     string
	Output   string
	ExitCode int
}

func (e *ExternalError) Error() string {
	if e.Output != "" {
		return fmt.Sprintf("%s error (exit %d): %v | output: %s", e.Tool, e.ExitCode, e.Err, e.Output)
	}
	return fmt.Sprintf("%s error (exit %d): %v", e.Tool, e.ExitCode, e.Err)
}

func (e *ExternalError) Unwrap() error {
	return ErrExternal
}

func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrLinkExpired) || errors.Is(err, ErrNotFound) || errors.Is(err, ErrAuth) || errors.Is(err, ErrInvalidInput) {
		return false
	}
	return true
}

func CategorizeError(err error) error {
	if err == nil {
		return nil
	}

	var netErr *NetworkError
	var storErr *StorageError
	var quotaErr *QuotaError
	var extErr *ExternalError
	if errors.As(err, &netErr) || errors.As(err, &storErr) || errors.As(err, &quotaErr) || errors.As(err, &extErr) {
		return err
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		addr := ""
		if opErr.Addr != nil {
			addr = opErr.Addr.String()
		}
		return &NetworkError{Err: err, URL: addr}
	}

	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "status=618"), strings.Contains(errMsg, "link expired"), strings.Contains(errMsg, "token expired"):
		return fmt.Errorf("%w: %w", ErrLinkExpired, err)
	case strings.Contains(errMsg, "404"), strings.Contains(errMsg, "not found"):
		return fmt.Errorf("%w: %w", ErrNotFound, err)
	case strings.Contains(errMsg, "403"), strings.Contains(errMsg, "unauthorized"), strings.Contains(errMsg, "forbidden"):
		return fmt.Errorf("%w: %w", ErrAuth, err)
	case strings.Contains(errMsg, "quota"), strings.Contains(errMsg, "rate limit"), strings.Contains(errMsg, "too many requests"):
		return fmt.Errorf("%w: %w", ErrQuota, err)
	case strings.Contains(errMsg, "connection"), strings.Contains(errMsg, "timeout"), strings.Contains(errMsg, "dial"), strings.Contains(errMsg, "no route"):
		return fmt.Errorf("%w: %w", ErrNetwork, err)
	case strings.Contains(errMsg, "disk"), strings.Contains(errMsg, "space"), strings.Contains(errMsg, "no space"), strings.Contains(errMsg, "permission denied"):
		return fmt.Errorf("%w: %w", ErrStorage, err)
	case strings.Contains(errMsg, "aria2"), strings.Contains(errMsg, "yt-dlp"), strings.Contains(errMsg, "rclone"):
		return fmt.Errorf("%w: %w", ErrExternal, err)
	default:
		return err
	}
}

// GetUserMessage is GetUserMessageLang in i18n.DefaultLang.
func GetUserMessage(err error) string {
	return GetUserMessageLang(i18n.DefaultLang, err)
}

// GetUserMessageLang explains err to a user in lang (plain text), by the
// sentinel it wraps.
func GetUserMessageLang(lang string, err error) string {
	if err == nil {
		return ""
	}

	if errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrAuth) {
		return i18n.T(lang, "core_err_unauthorized")
	}
	if errors.Is(err, ErrDuplicateTask) {
		return i18n.T(lang, "core_err_duplicate")
	}
	if errors.Is(err, ErrLimitExceeded) || errors.Is(err, ErrQuota) {
		return i18n.T(lang, "core_err_limit")
	}
	if errors.Is(err, ErrNotFound) {
		return i18n.T(lang, "core_err_not_found")
	}
	if errors.Is(err, ErrLinkExpired) {
		return i18n.T(lang, "core_err_link_expired")
	}
	if errors.Is(err, ErrInvalidInput) {
		return i18n.T(lang, "core_err_invalid_input")
	}
	if errors.Is(err, ErrNetwork) {
		return i18n.T(lang, "core_err_network")
	}
	if errors.Is(err, ErrStorage) {
		return i18n.T(lang, "core_err_storage")
	}
	if errors.Is(err, ErrExternal) {
		return i18n.T(lang, "core_err_external")
	}

	return i18n.T(lang, "core_err_internal")
}
