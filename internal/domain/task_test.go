package domain

import "testing"

func TestSnapshotUsesOriginalURL(t *testing.T) {
	task := &Task{URL: "https://example.com/a.iso"}
	if got := task.GetSnapshot().URL; got != "https://example.com/a.iso" {
		t.Errorf("snapshot URL = %q, want the task URL", got)
	}

	task = &Task{
		URL:       "https://api.telegram.org/file/bot123:secret/documents/a.iso",
		SourceURL: "tgfileid://abc",
	}
	if got := task.GetSnapshot().URL; got != "tgfileid://abc" {
		t.Errorf("snapshot URL = %q, want the source URL", got)
	}
}
