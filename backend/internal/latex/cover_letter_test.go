package latex

import (
	"strings"
	"testing"
)

func TestRenderCoverLetterFrenchTemplate(t *testing.T) {
	out := RenderCoverLetter(
		"Rayan Malki",
		[]string{"rayan@example.com", "github.com/RayanMalki"},
		"Paragraphe un.\n\nParagraphe deux.",
		"French",
	)

	if !strings.Contains(out, "\\selectlanguage{french}") {
		t.Fatalf("expected french document language")
	}
	if !strings.Contains(out, "Madame, Monsieur,") {
		t.Fatalf("expected french salutation")
	}
	if !strings.Contains(out, "Cordialement,") {
		t.Fatalf("expected french closing")
	}
	if strings.Contains(out, "Hiring Manager") || strings.Contains(out, "Company Name") {
		t.Fatalf("expected addressee block to be removed")
	}
}

func TestRenderCoverLetterEnglishTemplate(t *testing.T) {
	out := RenderCoverLetter(
		"Rayan Malki",
		[]string{"rayan@example.com"},
		"Paragraph one.\n\nParagraph two.",
		"English",
	)

	if !strings.Contains(out, "\\selectlanguage{english}") {
		t.Fatalf("expected english document language")
	}
	if !strings.Contains(out, "Dear Hiring Team,") {
		t.Fatalf("expected english salutation")
	}
	if !strings.Contains(out, "Sincerely,") {
		t.Fatalf("expected english closing")
	}
}
