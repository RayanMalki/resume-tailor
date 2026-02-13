package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ── PDF text extraction ────────────────────────────────────────────
// Uses github.com/ledongthuc/pdf for proper PDF parsing including
// FlateDecode decompression, CMap font mapping, and text operators.

func extractTextFromPDF(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to parse PDF: %w", err)
	}

	var buf strings.Builder
	numPages := reader.NumPage()
	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue // skip pages that fail
		}
		if buf.Len() > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(text)
	}

	return strings.TrimSpace(buf.String()), nil
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
