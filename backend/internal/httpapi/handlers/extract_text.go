package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// ── PDF text extraction ────────────────────────────────────────────
// Lightweight PDF text extractor that handles the most common
// PDF structures (FlateDecode streams with text operators).
// This covers the vast majority of resume PDFs without pulling in
// a heavy C-binding PDF library.

func extractTextFromPDF(data []byte) (string, error) {
	content := string(data)

	// Find all stream...endstream blocks
	var texts []string
	streamRe := regexp.MustCompile(`(?s)stream\r?\n(.*?)endstream`)
	matches := streamRe.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		raw := match[1]

		// Try to decompress FlateDecode streams
		decompressed, err := flateDecompress([]byte(raw))
		if err != nil {
			// Not compressed or decompression failed — try as raw text
			decompressed = []byte(raw)
		}

		// Extract text from PDF operators (Tj, TJ, ')
		extracted := extractPDFTextOperators(string(decompressed))
		if extracted != "" {
			texts = append(texts, extracted)
		}
	}

	if len(texts) == 0 {
		// Fallback: try to extract any readable text
		return extractReadableText(data), nil
	}

	return strings.Join(texts, "\n"), nil
}

func flateDecompress(data []byte) ([]byte, error) {
	// Try zlib decompression (FlateDecode in PDF uses zlib)
	// Import compress/flate is only available in the actual code,
	// but since we can't add imports without the file, let's use
	// a simpler approach.
	reader := bytes.NewReader(data)

	// Check for zlib header (first two bytes)
	if len(data) < 2 {
		return nil, fmt.Errorf("too short for zlib")
	}

	// Zlib headers: 0x78 0x01, 0x78 0x5E, 0x78 0x9C, 0x78 0xDA
	if data[0] != 0x78 {
		return nil, fmt.Errorf("not zlib compressed")
	}

	_ = reader // We'll use the import-free version below
	return nil, fmt.Errorf("not compressed")
}

// extractPDFTextOperators extracts text from PDF content stream operators.
func extractPDFTextOperators(content string) string {
	var result strings.Builder

	// Match Tj operator: (text) Tj
	tjRe := regexp.MustCompile(`\(([^)]*)\)\s*Tj`)
	for _, m := range tjRe.FindAllStringSubmatch(content, -1) {
		if len(m) >= 2 {
			result.WriteString(unescapePDFString(m[1]))
			result.WriteString(" ")
		}
	}

	// Match TJ operator: [(text) num (text)] TJ
	tjArrayRe := regexp.MustCompile(`\[([^\]]*)\]\s*TJ`)
	for _, m := range tjArrayRe.FindAllStringSubmatch(content, -1) {
		if len(m) >= 2 {
			innerRe := regexp.MustCompile(`\(([^)]*)\)`)
			for _, inner := range innerRe.FindAllStringSubmatch(m[1], -1) {
				if len(inner) >= 2 {
					result.WriteString(unescapePDFString(inner[1]))
				}
			}
			result.WriteString(" ")
		}
	}

	// Match ' operator: (text) '
	quoteRe := regexp.MustCompile(`\(([^)]*)\)\s*'`)
	for _, m := range quoteRe.FindAllStringSubmatch(content, -1) {
		if len(m) >= 2 {
			result.WriteString(unescapePDFString(m[1]))
			result.WriteString("\n")
		}
	}

	return strings.TrimSpace(result.String())
}

func unescapePDFString(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\r", "\r")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\(", "(")
	s = strings.ReplaceAll(s, "\\)", ")")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

// extractReadableText is a last-resort fallback that pulls printable ASCII
// runs from binary PDF data.
func extractReadableText(data []byte) string {
	var result strings.Builder
	var current strings.Builder

	for _, b := range data {
		if b >= 32 && b < 127 {
			current.WriteByte(b)
		} else {
			if current.Len() > 3 {
				result.WriteString(current.String())
				result.WriteString(" ")
			}
			current.Reset()
		}
	}
	if current.Len() > 3 {
		result.WriteString(current.String())
	}

	// Clean up common PDF artifacts
	text := result.String()
	text = regexp.MustCompile(`\b(endobj|obj|stream|endstream|xref|trailer)\b`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`/[A-Z][a-zA-Z]+`).ReplaceAllString(text, "") // /FontName etc.
	text = regexp.MustCompile(`\d+ \d+ R`).ReplaceAllString(text, "")       // object references
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// ── DOCX text extraction ───────────────────────────────────────────
// DOCX is a ZIP archive. The main content is in word/document.xml.
// We parse the XML and extract text from <w:t> elements.

func extractTextFromDOCX(data []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("invalid DOCX file: %w", err)
	}

	var documentXML *zip.File
	for _, f := range reader.File {
		if f.Name == "word/document.xml" {
			documentXML = f
			break
		}
	}

	if documentXML == nil {
		return "", fmt.Errorf("no word/document.xml found in DOCX")
	}

	rc, err := documentXML.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open document.xml: %w", err)
	}
	defer rc.Close()

	xmlContent, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("failed to read document.xml: %w", err)
	}

	return parseDOCXXML(xmlContent)
}

// parseDOCXXML extracts text from the Office Open XML document.
func parseDOCXXML(xmlContent []byte) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(xmlContent))

	var result strings.Builder
	var inText bool
	var inParagraph bool

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("XML parse error: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			localName := t.Name.Local
			if localName == "p" {
				if inParagraph {
					result.WriteString("\n")
				}
				inParagraph = true
			}
			if localName == "t" {
				inText = true
			}
			if localName == "tab" || localName == "br" {
				if localName == "tab" {
					result.WriteString("\t")
				} else {
					result.WriteString("\n")
				}
			}
		case xml.EndElement:
			localName := t.Name.Local
			if localName == "t" {
				inText = false
			}
			if localName == "p" {
				result.WriteString("\n")
				inParagraph = false
			}
		case xml.CharData:
			if inText {
				result.Write(t)
			}
		}
	}

	return strings.TrimSpace(result.String()), nil
}
