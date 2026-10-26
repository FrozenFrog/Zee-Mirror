package i18n

import (
	"strings"
	"testing"
)

var filesPrefixes = []string{"drive_", "fm_", "storage_", "search_"}

func isFilesKey(key string) bool {
	for _, p := range filesPrefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}

// files.go marks plain-text keys (buttons, toasts, errors, messages sent
// without a parse mode) in their name; everything else is MarkdownV2.
func isFilesPlainKey(key string) bool {
	for _, marker := range []string{"_btn_", "_toast_", "_err_", "_plain_"} {
		if strings.Contains(key, marker) {
			return true
		}
	}
	return false
}

// The generic TestMarkdownV2Escaping only checks keys whose Vietnamese text
// already contains a backslash. Every MarkdownV2 key in files.go is sent with
// ParseMode MarkdownV2, so check all of them in every language.
func TestFilesMarkdownV2Keys(t *testing.T) {
	checked := 0
	for key := range translations[DefaultLang] {
		if !isFilesKey(key) || isFilesPlainKey(key) {
			continue
		}
		checked++
		for lang, msgs := range translations {
			if c := mdv2Unescaped(msgs[key]); c != 0 {
				t.Errorf("%s/%s: unescaped %q in %q", lang, key, c, msgs[key])
			}
		}
	}
	if checked == 0 {
		t.Fatal("no MarkdownV2 keys found for files catalog")
	}
}

// Plain-text keys are sent without a parse mode, so a backslash would show up
// literally.
func TestFilesPlainKeysHaveNoEscapes(t *testing.T) {
	for key := range translations[DefaultLang] {
		if !isFilesKey(key) || !isFilesPlainKey(key) {
			continue
		}
		for lang, msgs := range translations {
			if strings.Contains(msgs[key], `\`) {
				t.Errorf("%s/%s: plain text contains a backslash: %q", lang, key, msgs[key])
			}
		}
	}
}

func TestFilesCatalogSample(t *testing.T) {
	if got := T("vi", "fm_count_files", 3); got != "📄 File: `3`" {
		t.Errorf("vi = %q", got)
	}
	if got := T("en", "fm_count_files", 3); got != "📄 Files: `3`" {
		t.Errorf("en = %q", got)
	}
	err := WrapError(nil, "storage_err_list_remotes", "boom")
	if got := Message("vi", err); got != "Không liệt kê được remote rclone: boom" {
		t.Errorf("Message(vi) = %q", got)
	}
	if got := err.Error(); got != "failed to list remotes: boom" {
		t.Errorf("Error() = %q", got)
	}
}
