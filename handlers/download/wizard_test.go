package download

import (
	"testing"

	"zee-mirror/pkg/i18n"
)

func TestIsWizardPromptReply(t *testing.T) {
	for _, l := range i18n.Languages {
		for _, key := range []string{"wizard_prompt_url", "wizard_prompt_rename"} {
			if plain := plainMarkdownV2(i18n.T(l.Code, key)); !IsWizardPromptReply(plain) {
				t.Errorf("%s/%s prompt %q not recognised", l.Code, key, plain)
			}
		}
		if !isWizardPrompt(plainMarkdownV2(i18n.T(l.Code, "wizard_prompt_url")), "wizard_prompt_url", legacyURLPromptMarkers) {
			t.Errorf("%s URL prompt not recognised as URL prompt", l.Code)
		}
		if isWizardPrompt(plainMarkdownV2(i18n.T(l.Code, "wizard_prompt_rename")), "wizard_prompt_url", legacyURLPromptMarkers) {
			t.Errorf("%s rename prompt mistaken for URL prompt", l.Code)
		}
	}
	if !IsWizardPromptReply("🔗 Silakan kirim URL yang ingin Anda Mirror:") {
		t.Error("legacy Indonesian prompt not recognised")
	}
	if IsWizardPromptReply("some unrelated bot message") {
		t.Error("unrelated text recognised as a wizard prompt")
	}
	if PromptMirrorURL != i18n.T("vi", "wizard_prompt_url") {
		t.Errorf("PromptMirrorURL = %q", PromptMirrorURL)
	}
}

// The wizard keeps its state in the options card, so every translation must
// round-trip the URL and zip flag.
func TestWizardOptionsRoundTrip(t *testing.T) {
	const url = "https://example.com/a-b.zip"
	for _, l := range i18n.Languages {
		for _, zip := range []bool{false, true} {
			plain := plainMarkdownV2(wizardOptionsText(l.Code, url, zip))
			gotURL, gotZip := parseWizardOptions(plain)
			if gotURL != url || gotZip != zip {
				t.Errorf("%s zip=%v: parsed (%q, %v) from %q", l.Code, zip, gotURL, gotZip, plain)
			}
		}
	}
}
