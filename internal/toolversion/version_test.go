package toolversion

import "testing"

func TestVersionParsing(t *testing.T) {
	tests := []struct {
		name, output, want string
	}{
		{"aria2", "aria2 version 1.37.0\nCopyright", "1.37.0"},
		{"rclone", "rclone v1.75.2\n- os/version: alpine", "1.75.2"},
		{"yt-dlp", "2026.10.26\n", "2026.10.26"},
	}
	for _, test := range tests {
		if got := parseCurrent(test.name, test.output); got != test.want {
			t.Errorf("%s: got %q, want %q", test.name, got, test.want)
		}
	}
	if got := normalizeReleaseTag("release-1.37.0"); got != "1.37.0" {
		t.Errorf("aria2 release tag: got %q", got)
	}
	if got := normalizeReleaseTag("v1.75.2"); got != "1.75.2" {
		t.Errorf("rclone release tag: got %q", got)
	}
}
