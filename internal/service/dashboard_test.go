package service

import (
	"testing"
	"time"
)

func TestCompactDuration(t *testing.T) {
	cases := map[time.Duration]string{
		0:                         "0s",
		45 * time.Second:          "45s",
		time.Hour + 5*time.Second: "1h5s",
		32*time.Hour + 56*time.Minute + 23*time.Second: "1d8h56m23s",
	}
	for d, want := range cases {
		if got := compactDuration(d); got != want {
			t.Errorf("compactDuration(%v) = %q, want %q", d, got, want)
		}
	}
}

func TestBuildStatusKeyboard(t *testing.T) {
	kb := buildStatusKeyboard(0, 1)
	if len(kb.InlineKeyboard) != 1 || len(kb.InlineKeyboard[0]) != 2 {
		t.Fatalf("single page: want one row of 📜 ♻️, got %+v", kb.InlineKeyboard)
	}

	kb = buildStatusKeyboard(0, 3)
	if len(kb.InlineKeyboard) != 2 {
		t.Fatalf("paged: want navigation row plus action row, got %d rows", len(kb.InlineKeyboard))
	}
	nav := kb.InlineKeyboard[0]
	if *nav[0].CallbackData != "dashboard:page:2" || *nav[2].CallbackData != "dashboard:page:1" {
		t.Errorf("navigation wraps wrong: prev=%s next=%s", *nav[0].CallbackData, *nav[2].CallbackData)
	}
}
