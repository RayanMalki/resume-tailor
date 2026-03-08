package ai

import (
	"strings"
	"testing"
)

func TestBuildCoverLetterPromptBodyOnlyAndSingleLanguage(t *testing.T) {
	prompt := buildCoverLetterPrompt("resume", "job", nil, "French", DisciplineContext{})

	if !strings.Contains(prompt, "LANGUAGE: All output must be in French. Use exactly one language for the entire letter.") {
		t.Fatalf("expected strict single-language guidance")
	}
	if !strings.Contains(prompt, "Output ONLY the body paragraphs. No addressee block.") {
		t.Fatalf("expected body-only instruction")
	}
	if !strings.Contains(prompt, "Do NOT include salutations") {
		t.Fatalf("expected no-salutation instruction")
	}
	if !strings.Contains(prompt, "Do NOT include complimentary closes/signatures") {
		t.Fatalf("expected no-closing instruction")
	}
}
