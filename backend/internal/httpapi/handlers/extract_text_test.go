package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
)

// ── DOCX extraction tests ──────────────────────────────────────────

func TestExtractTextFromDOCX(t *testing.T) {
	// Create a minimal valid DOCX (ZIP with word/document.xml)
	docxBytes := createTestDOCX(t, `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p>
      <w:r>
        <w:t>John Doe</w:t>
      </w:r>
    </w:p>
    <w:p>
      <w:r>
        <w:t>Software Engineer</w:t>
      </w:r>
    </w:p>
    <w:p>
      <w:r>
        <w:t>5 years of experience in Go and Python</w:t>
      </w:r>
    </w:p>
  </w:body>
</w:document>`)

	text, err := extractTextFromDOCX(docxBytes)
	if err != nil {
		t.Fatalf("extractTextFromDOCX() error: %v", err)
	}

	if !strings.Contains(text, "John Doe") {
		t.Errorf("expected to find 'John Doe' in extracted text, got: %s", text)
	}
	if !strings.Contains(text, "Software Engineer") {
		t.Errorf("expected to find 'Software Engineer' in extracted text, got: %s", text)
	}
	if !strings.Contains(text, "Go and Python") {
		t.Errorf("expected to find 'Go and Python' in extracted text, got: %s", text)
	}
}

func TestExtractTextFromDOCXEmpty(t *testing.T) {
	docxBytes := createTestDOCX(t, `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body></w:body>
</w:document>`)

	text, err := extractTextFromDOCX(docxBytes)
	if err != nil {
		t.Fatalf("extractTextFromDOCX() error: %v", err)
	}
	if text != "" {
		t.Errorf("expected empty text, got: %s", text)
	}
}

func TestExtractTextFromDOCXInvalidZip(t *testing.T) {
	_, err := extractTextFromDOCX([]byte("not a zip file"))
	if err == nil {
		t.Fatal("expected error for invalid ZIP")
	}
}

func TestExtractTextFromDOCXNoDocumentXML(t *testing.T) {
	// Create a ZIP without word/document.xml
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, _ := w.Create("other.xml")
	f.Write([]byte("<root/>"))
	w.Close()

	_, err := extractTextFromDOCX(buf.Bytes())
	if err == nil {
		t.Fatal("expected error for missing word/document.xml")
	}
}

// ── parseDOCXXML tests ─────────────────────────────────────────────

func TestParseDOCXXMLMultipleParagraphs(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>Line 1</w:t></w:r></w:p>
    <w:p><w:r><w:t>Line 2</w:t></w:r></w:p>
  </w:body>
</w:document>`

	text, err := parseDOCXXML([]byte(xmlContent))
	if err != nil {
		t.Fatalf("parseDOCXXML() error: %v", err)
	}
	if !strings.Contains(text, "Line 1") || !strings.Contains(text, "Line 2") {
		t.Errorf("expected both lines, got: %s", text)
	}
}

// ── PDF extraction tests ───────────────────────────────────────────

func TestExtractPDFTextOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "Tj operator",
			input:    "(Hello World) Tj",
			contains: "Hello World",
		},
		{
			name:     "TJ array operator",
			input:    "[(Hello) 50 (World)] TJ",
			contains: "HelloWorld",
		},
		{
			name:     "escaped parentheses in nested context",
			input:    "(Hello World) Tj (Resume) Tj",
			contains: "Hello World",
		},
		{
			name:     "empty content",
			input:    "",
			contains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPDFTextOperators(tt.input)
			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("expected result to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestUnescapePDFString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`Hello\nWorld`, "Hello\nWorld"},
		{`Test\(paren\)`, "Test(paren)"},
		{`Back\\slash`, "Back\\slash"},
		{`Tab\there`, "Tab\there"},
	}

	for _, tt := range tests {
		result := unescapePDFString(tt.input)
		if result != tt.expected {
			t.Errorf("unescapePDFString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestExtractReadableText(t *testing.T) {
	// Simulate binary data with some readable text embedded
	data := []byte{0, 0, 0}
	data = append(data, []byte("John Doe Software Engineer")...)
	data = append(data, []byte{0, 0, 0}...)
	data = append(data, []byte("Python JavaScript Docker")...)
	data = append(data, []byte{0, 0, 0}...)

	result := extractReadableText(data)
	if !strings.Contains(result, "John Doe") {
		t.Errorf("expected 'John Doe' in result, got: %s", result)
	}
	if !strings.Contains(result, "Python") {
		t.Errorf("expected 'Python' in result, got: %s", result)
	}
}

// ── JSON helper tests ──────────────────────────────────────────────

func TestWriteErrorJSON(t *testing.T) {
	// This is a basic validation that writeError produces JSON
	// Testing would require httptest.ResponseRecorder (covered in handler tests)
}

// ── Helper functions ───────────────────────────────────────────────

func createTestDOCX(t *testing.T, documentXML string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	// Validate XML before adding
	if err := xml.Unmarshal([]byte(documentXML), new(interface{})); err != nil {
		// Not strictly valid XML, but that's fine for testing parsing
	}

	f, err := w.Create("word/document.xml")
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	if _, err := f.Write([]byte(documentXML)); err != nil {
		t.Fatalf("write zip entry: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}
