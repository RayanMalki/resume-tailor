package docx

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"

	"resume-tailor/internal/ai"
)

// RenderResume builds a minimal .docx file from the tailored resume spec.
func RenderResume(spec ai.ResumeSpec) ([]byte, error) {
	paragraphs := resumeParagraphs(spec)
	documentXML := buildDocumentXML(paragraphs)

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

func resumeParagraphs(spec ai.ResumeSpec) []string {
	var out []string
	appendIf := func(s string) {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}

	appendIf(spec.Name)
	appendIf(spec.Title)
	appendIf(strings.Join(spec.Contact, " | "))
	out = append(out, "")

	if len(spec.Summary) > 0 {
		appendIf("Summary")
		for _, s := range spec.Summary {
			appendIf("- " + s)
		}
		out = append(out, "")
	}

	if len(spec.Education) > 0 {
		appendIf("Education")
		for _, e := range spec.Education {
			appendIf(strings.TrimSpace(e.Degree + " - " + e.School))
			appendIf(strings.TrimSpace(e.Location + " | " + e.Dates))
			for _, d := range e.Details {
				appendIf("  - " + d)
			}
		}
		out = append(out, "")
	}

	if len(spec.SkillGroups) > 0 || len(spec.Skills) > 0 {
		appendIf("Skills")
		for _, g := range spec.SkillGroups {
			if strings.TrimSpace(g.Name) == "" {
				continue
			}
			appendIf(g.Name + ": " + strings.Join(g.Items, ", "))
		}
		if len(spec.Skills) > 0 {
			appendIf("General: " + strings.Join(spec.Skills, ", "))
		}
		out = append(out, "")
	}

	if len(spec.Projects) > 0 {
		appendIf("Projects")
		for _, p := range spec.Projects {
			appendIf(strings.TrimSpace(p.Name + " | " + p.Stack + " | " + p.Dates))
			for _, b := range p.Bullets {
				appendIf("  - " + b)
			}
		}
		out = append(out, "")
	}

	if len(spec.Experience) > 0 {
		appendIf("Experience")
		for _, e := range spec.Experience {
			appendIf(strings.TrimSpace(e.Role + " - " + e.Company))
			appendIf(strings.TrimSpace(e.Location + " | " + e.Dates))
			for _, b := range e.Bullets {
				appendIf("  - " + b)
			}
		}
	}

	return out
}

func buildDocumentXML(paragraphs []string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<w:document xmlns:wpc="http://schemas.microsoft.com/office/word/2010/wordprocessingCanvas" xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006" xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:m="http://schemas.openxmlformats.org/officeDocument/2006/math" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:wp14="http://schemas.microsoft.com/office/word/2010/wordprocessingDrawing" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:w10="urn:schemas-microsoft-com:office:word" xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" mc:Ignorable="w14">`)
	b.WriteString(`<w:body>`)
	for _, p := range paragraphs {
		b.WriteString(`<w:p><w:r><w:t xml:space="preserve">`)
		b.WriteString(xmlEscape(p))
		b.WriteString(`</w:t></w:r></w:p>`)
	}
	b.WriteString(`<w:sectPr><w:pgSz w:w="12240" w:h="15840"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="708" w:footer="708" w:gutter="0"/></w:sectPr>`)
	b.WriteString(`</w:body></w:document>`)
	return b.String()
}

func xmlEscape(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(s)
}

const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

const relsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
