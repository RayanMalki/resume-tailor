package ai

import (
	"strings"
	"testing"
)

func TestBuildResumeLatexPrompt(t *testing.T) {
	prompt := buildResumeLatexPrompt("resume", "job")

	if !strings.Contains(prompt, "Return ONLY LaTeX") {
		t.Fatalf("expected strict latex instruction")
	}
	if !strings.Contains(prompt, "Jake's Resume") {
		t.Fatalf("expected jake resume instruction")
	}
	if !strings.Contains(prompt, "Education") || !strings.Contains(prompt, "Experience") || !strings.Contains(prompt, "Projects") || !strings.Contains(prompt, "Skills") {
		t.Fatalf("expected section order guidance")
	}
}
