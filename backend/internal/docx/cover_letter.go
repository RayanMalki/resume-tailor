package docx

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
)

// RenderCoverLetter builds a minimal .docx for a cover letter.
// name is the applicant's full name, contact is a joined contact line,
// and coverLetterText is the plain-text body (paragraphs separated by blank lines).
func RenderCoverLetter(name, contact, coverLetterText string) ([]byte, error) {
	return RenderCoverLetterWithLanguage(name, contact, coverLetterText, "")
}

// RenderCoverLetterWithLanguage builds a minimal .docx for a cover letter and
// localizes salutation/closing for supported languages.
func RenderCoverLetterWithLanguage(name, contact, coverLetterText, language string) ([]byte, error) {
	documentXML := buildCoverLetterDocumentXML(name, contact, coverLetterText, language)

	files := map[string]string{
		"[Content_Types].xml": contentTypesXML,
		"_rels/.rels":         relsXML,
		"word/document.xml":   documentXML,
		"word/_rels/document.xml.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`,
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			return nil, fmt.Errorf("create zip entry %s: %w", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			return nil, fmt.Errorf("write zip entry %s: %w", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("close zip: %w", err)
	}
	return buf.Bytes(), nil
}

func buildCoverLetterDocumentXML(name, contact, coverLetterText, language string) string {
	salutation, closing := docxCoverLetterPhrases(language)

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<w:document xmlns:wpc="http://schemas.microsoft.com/office/word/2010/wordprocessingCanvas" xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006" xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:m="http://schemas.openxmlformats.org/officeDocument/2006/math" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:wp14="http://schemas.microsoft.com/office/word/2010/wordprocessingDrawing" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:w10="urn:schemas-microsoft-com:office:word" xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" mc:Ignorable="w14">`)
	b.WriteString(`<w:body>`)

	// Name as bold heading
	if strings.TrimSpace(name) != "" {
		b.WriteString(`<w:p><w:pPr><w:jc w:val="left"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="28"/><w:szCs w:val="28"/></w:rPr><w:t xml:space="preserve">`)
		b.WriteString(xmlEscape(strings.TrimSpace(name)))
		b.WriteString(`</w:t></w:r></w:p>`)
	}

	// Contact line
	if strings.TrimSpace(contact) != "" {
		b.WriteString(`<w:p><w:r><w:rPr><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr><w:t xml:space="preserve">`)
		b.WriteString(xmlEscape(strings.TrimSpace(contact)))
		b.WriteString(`</w:t></w:r></w:p>`)
	}

	// Horizontal rule via paragraph bottom border
	b.WriteString(`<w:p><w:pPr><w:pBdr><w:bottom w:val="single" w:sz="6" w:space="1" w:color="000000"/></w:pBdr></w:pPr></w:p>`)

	// Empty spacer paragraph
	b.WriteString(`<w:p></w:p>`)

	// Salutation
	b.WriteString(`<w:p><w:r><w:t xml:space="preserve">`)
	b.WriteString(xmlEscape(salutation))
	b.WriteString(`</w:t></w:r></w:p>`)

	// Body paragraphs split on blank lines
	rawParas := strings.Split(coverLetterText, "\n\n")
	for _, para := range rawParas {
		trimmed := strings.TrimSpace(para)
		if trimmed == "" {
			continue
		}
		b.WriteString(`<w:p><w:r><w:t xml:space="preserve">`)
		b.WriteString(xmlEscape(trimmed))
		b.WriteString(`</w:t></w:r></w:p>`)
	}

	// Closing + signature
	b.WriteString(`<w:p></w:p>`)
	b.WriteString(`<w:p><w:r><w:t xml:space="preserve">`)
	b.WriteString(xmlEscape(closing))
	b.WriteString(`</w:t></w:r></w:p>`)
	b.WriteString(`<w:p></w:p>`)
	if strings.TrimSpace(name) != "" {
		b.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">`)
		b.WriteString(xmlEscape(strings.TrimSpace(name)))
		b.WriteString(`</w:t></w:r></w:p>`)
	}

	b.WriteString(`<w:sectPr><w:pgSz w:w="12240" w:h="15840"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="708" w:footer="708" w:gutter="0"/></w:sectPr>`)
	b.WriteString(`</w:body></w:document>`)
	return b.String()
}

func docxCoverLetterPhrases(language string) (salutation, closing string) {
	lang := strings.ToLower(strings.TrimSpace(language))
	if strings.HasPrefix(lang, "fr") || strings.Contains(lang, "french") {
		return "Madame, Monsieur,", "Cordialement,"
	}
	return "Dear Hiring Team,", "Sincerely,"
}
