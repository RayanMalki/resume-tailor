package latex

import (
	"strings"
	"time"
)

// RenderCoverLetter builds a LaTeX document for a professional cover letter.
// name is the applicant's full name, contact is a pre-joined contact line,
// and coverLetterText is the plain-text body (paragraphs separated by blank lines).
func RenderCoverLetter(name, contact, coverLetterText string) string {
	escapedName := escapeLatex(fallback(name, "Applicant"))
	escapedContact := escapeLatex(contact)
	date := time.Now().Format("January 2, 2006")

	// Split body into paragraphs on blank lines
	rawParas := strings.Split(coverLetterText, "\n\n")
	var paras []string
	for _, p := range rawParas {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			paras = append(paras, escapeLatex(trimmed))
		}
	}

	var b strings.Builder

	b.WriteString("\\documentclass[letterpaper,11pt]{article}\n\n")
	b.WriteString("\\usepackage[margin=1in]{geometry}\n")
	b.WriteString("\\usepackage[hidelinks]{hyperref}\n")
	b.WriteString("\\usepackage{parskip}\n")
	b.WriteString("\\usepackage{lmodern}\n")
	b.WriteString("\\usepackage{microtype}\n")
	b.WriteString("\\usepackage{xcolor}\n\n")

	b.WriteString("\\pagestyle{empty}\n\n")
	b.WriteString("\\setlength{\\parskip}{0.8em}\n")
	b.WriteString("\\setlength{\\parindent}{0pt}\n\n")

	b.WriteString("\\begin{document}\n\n")

	// Header: name + contact
	b.WriteString("{\\Large \\textbf{")
	b.WriteString(escapedName)
	b.WriteString("}}\n\n")
	if escapedContact != "" {
		b.WriteString("{\\small ")
		b.WriteString(escapedContact)
		b.WriteString("}\n\n")
	}

	// Horizontal rule
	b.WriteString("\\noindent\\rule{\\textwidth}{0.4pt}\n\n")

	// Date right-aligned
	b.WriteString("\\begin{flushright}\n")
	b.WriteString(escapeLatex(date))
	b.WriteString("\n\\end{flushright}\n\n")

	// Body paragraphs
	for _, para := range paras {
		b.WriteString(para)
		b.WriteString("\n\n")
	}

	b.WriteString("\\end{document}\n")
	return b.String()
}
