package llm

import (
	"strings"
	"testing"
)

func TestBuildPrompt_ContainsSubjectAndBody(t *testing.T) {
	prompt := buildPrompt("Weekly Tech Digest", "Here are this week's top stories...")

	if !strings.Contains(prompt, "Subject: Weekly Tech Digest") {
		t.Error("prompt should contain the email subject")
	}

	if !strings.Contains(prompt, "Here are this week's top stories...") {
		t.Error("prompt should contain the email body")
	}
}

func TestBuildPrompt_ContainsNewsletterCheck(t *testing.T) {
	prompt := buildPrompt("subject", "body")

	if !strings.Contains(prompt, "CONTENT NEWSLETTER") {
		t.Error("prompt should instruct LLM to check whether email is a content newsletter")
	}

	if !strings.Contains(prompt, "Marketing emails") {
		t.Error("prompt should list marketing emails as non-newsletters")
	}
}

func TestBuildPrompt_ContainsContentTypePrefix(t *testing.T) {
	prompt := buildPrompt("subject", "body")

	if !strings.Contains(prompt, `"Article."`) {
		t.Error("prompt should include Article as example content type prefix")
	}

	if !strings.Contains(prompt, `"Blog post."`) {
		t.Error("prompt should include Blog post as example content type prefix")
	}

	if !strings.Contains(prompt, `"Podcast."`) {
		t.Error("prompt should include Podcast as example content type prefix")
	}
}

func TestBuildPrompt_ContainsBoilerplateExclusion(t *testing.T) {
	prompt := buildPrompt("subject", "body")

	for _, marker := range []string{
		"Werbung abbestellen",
		"Datenschutzinformationen",
		"Impressum",
		"Unsubscribe",
		"Privacy Policy",
	} {
		if !strings.Contains(prompt, marker) {
			t.Errorf("prompt should list %q as boilerplate to exclude", marker)
		}
	}
}

func TestBuildPrompt_ContainsSponsorExclusion(t *testing.T) {
	prompt := buildPrompt("subject", "body")

	if !strings.Contains(prompt, "(Sponsor)") {
		t.Error("prompt should list (Sponsor) as a sponsor marker")
	}
}

func TestBuildPrompt_ReuseSummaryTakesPrecedence(t *testing.T) {
	prompt := buildPrompt("subject", "body")

	reuseIdx := strings.Index(prompt, "reuse that summary word-for-word")
	otherwiseIdx := strings.Index(prompt, "Otherwise, write teasers")

	if reuseIdx == -1 {
		t.Fatal("prompt should contain reuse-summary rule")
	}
	if otherwiseIdx == -1 {
		t.Fatal("prompt should contain fallback length rule")
	}
	if reuseIdx > otherwiseIdx {
		t.Error("reuse-summary rule should appear before the fallback length rule")
	}
}
