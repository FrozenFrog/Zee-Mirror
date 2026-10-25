package ytdlp

import "testing"

func TestYoutubeReferer(t *testing.T) {
	youtube := []string{"https://youtu.be/DKqoqfoLqNQ?si=x", "https://www.youtube.com/watch?v=1", "https://m.youtube.com/shorts/1", "https://music.youtube.com/watch?v=1"}
	other := []string{"https://x.com/yess_bi/status/2108211995975766207", "https://www.tiktok.com/@a/video/1", "https://notyoutube.com/a", "::bad"}
	for _, u := range youtube {
		if len(youtubeReferer(u)) == 0 {
			t.Errorf("%s: expected the YouTube Referer", u)
		}
	}
	for _, u := range other {
		if got := youtubeReferer(u); got != nil {
			t.Errorf("%s: unexpected headers %v", u, got)
		}
	}
}
