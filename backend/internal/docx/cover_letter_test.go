package docx

import (
	"strings"
	"testing"
)

func TestBuildCoverLetterDocumentXMLFrenchTemplate(t *testing.T) {
	out := buildCoverLetterDocumentXML(
		"Rayan Malki",
		"rayan@example.com",
		"Paragraphe un.\n\nParagraphe deux.",
		"French",
	)

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

func TestBuildCoverLetterDocumentXMLEnglishTemplate(t *testing.T) {
	out := buildCoverLetterDocumentXML(
		"Rayan Malki",
		"rayan@example.com",
		"Paragraph one.\n\nParagraph two.",
		"English",
	)

	if !strings.Contains(out, "Dear Hiring Team,") {
		t.Fatalf("expected english salutation")
	}
	if !strings.Contains(out, "Sincerely,") {
		t.Fatalf("expected english closing")
	}
}
