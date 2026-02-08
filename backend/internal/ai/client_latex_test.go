package ai

import (
	"strings"
	"testing"
)

func TestBuildResumeLatexPrompt(t *testing.T) {
	prompt := buildResumeLatexPrompt("resume", "job", nil)

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

func TestBuildResumeSpecPromptHasLanguageAndProjectRules(t *testing.T) {
	prompt := buildResumeSpecPrompt("resume", "job", nil)

	if !strings.Contains(prompt, "Use the resume's primary language for ALL section content.") {
		t.Fatalf("expected resume language instruction")
	}
	if !strings.Contains(prompt, "Do NOT switch language to match the job posting language.") {
		t.Fatalf("expected explicit language precedence instruction")
	}
	if !strings.Contains(prompt, "include EVERY project from the resume that is relevant to the job") {
		t.Fatalf("expected relevant project inclusion instruction")
	}
	if !strings.Contains(prompt, "\"skill_groups\"") {
		t.Fatalf("expected skill_groups schema")
	}
	if !strings.Contains(prompt, "\"language\"") {
		t.Fatalf("expected language field in schema")
	}
}
