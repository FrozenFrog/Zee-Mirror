package i18n

import (
	"errors"
	"fmt"
	"strings"
)

// Language describes a language the bot can speak.
type Language struct {
	Code string
	Flag string
	Name string // native name
}

// Languages lists every supported language in menu order.
var Languages = []Language{
	{Code: "vi", Flag: "🇻🇳", Name: "Tiếng Việt"},
	{Code: "en", Flag: "🇺🇸", Name: "English"},
	{Code: "id", Flag: "🇮🇩", Name: "Bahasa Indonesia"},
	{Code: "ja", Flag: "🇯🇵", Name: "日本語"},
	{Code: "zh", Flag: "🇨🇳", Name: "简体中文"},
}

// IsSupported reports whether code is one of Languages.
func IsSupported(code string) bool {
	_, ok := translations[strings.ToLower(code)]
	return ok
}

// Normalize maps an IETF/Telegram language code ("en-US", "zh-hans", "pt-br")
// to a supported language code, or "" when none matches.
func Normalize(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if i := strings.IndexAny(code, "-_"); i >= 0 {
		code = code[:i]
	}
	if IsSupported(code) {
		return code
	}
	return ""
}

// register merges a domain's catalog into translations. Each domain file
// (pkg/i18n/<domain>.go) owns its keys and calls this from init; defining a
// key twice or using an unknown language is a programming error.
func register(catalog map[string]map[string]string) {
	for lang, msgs := range catalog {
		dst, ok := translations[lang]
		if !ok {
			panic(fmt.Sprintf("i18n: unknown language %q", lang))
		}
		for key, msg := range msgs {
			if _, dup := dst[key]; dup {
				panic(fmt.Sprintf("i18n: duplicate key %q (%s)", key, lang))
			}
			dst[key] = msg
		}
	}
}

// Lookup returns the translation of key in lang (falling back to
// DefaultLang) and whether the key exists at all.
func Lookup(lang, key string, args ...any) (string, bool) {
	if _, ok := translations[DefaultLang][key]; !ok {
		return "", false
	}
	return T(lang, key, args...), true
}

// Error is an error whose user-facing text is translated when shown, so code
// without a user context (plugins, uploaders, userbot) can still fail in the
// user's language. Error() renders English for logs.
type Error struct {
	Key  string
	Args []any
	Err  error // optional cause, exposed via Unwrap
}

// NewError returns an *Error for key formatted with args.
func NewError(key string, args ...any) *Error {
	return &Error{Key: key, Args: args}
}

// WrapError is NewError with a cause that errors.Is/As can reach.
func WrapError(cause error, key string, args ...any) *Error {
	return &Error{Key: key, Args: args, Err: cause}
}

func (e *Error) Error() string { return T("en", e.Key, e.Args...) }

func (e *Error) Unwrap() error { return e.Err }

// Message renders err for a user speaking lang: the first *Error in the chain
// is translated, any other error falls back to err.Error().
func Message(lang string, err error) string {
	if err == nil {
		return ""
	}
	var e *Error
	if errors.As(err, &e) {
		return T(lang, e.Key, e.Args...)
	}
	return err.Error()
}
