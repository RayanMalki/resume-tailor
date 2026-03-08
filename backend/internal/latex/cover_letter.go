package latex

import (
	"strings"
)

// parseContactFields extracts phone, email, linkedin handle, and github handle
// from a slice of contact strings.
func parseContactFields(contact []string) (phone, email, linkedin, github string) {
	for _, item := range contact {
		lower := strings.ToLower(item)
		if email == "" && strings.Contains(item, "@") {
			email = strings.TrimSpace(item)
			continue
		}
		if linkedin == "" && strings.Contains(lower, "linkedin") {
			// Extract handle after linkedin.com/in/
			idx := strings.Index(lower, "linkedin.com/in/")
			if idx >= 0 {
				handle := item[idx+len("linkedin.com/in/"):]
				// Strip trailing slashes or spaces
				handle = strings.TrimRight(handle, "/ ")
				linkedin = handle
			}
			continue
		}
		if github == "" && strings.Contains(lower, "github") {
			// Extract handle after github.com/
			idx := strings.Index(lower, "github.com/")
			if idx >= 0 {
				handle := item[idx+len("github.com/"):]
				handle = strings.TrimRight(handle, "/ ")
				github = handle
			}
			continue
		}
		if phone == "" {
			trimmed := strings.TrimSpace(item)
			if strings.HasPrefix(trimmed, "+") || (len(trimmed) > 0 && trimmed[0] == '(') {
				phone = trimmed
			}
		}
	}
	return
}

// RenderCoverLetter builds a LaTeX document for a professional cover letter.
// name is the applicant's full name, contact is a slice of contact items,
// and coverLetterText is the plain-text body (paragraphs separated by blank lines).
func RenderCoverLetter(name string, contact []string, coverLetterText string) string {
	escapedName := escapeLatex(fallback(name, "Applicant"))

	phone, email, linkedin, github := parseContactFields(contact)

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

	b.WriteString("\\documentclass[11pt, letterpaper]{article}\n")
	b.WriteString("\\usepackage[top=0.75in, bottom=0.75in, left=0.85in, right=0.85in]{geometry}\n")
	b.WriteString("\\usepackage{hyperref}\n")
	b.WriteString("\\usepackage{xcolor}\n")
	b.WriteString("\\usepackage{parskip}\n")
	b.WriteString("\\usepackage[T1]{fontenc}\n")
	b.WriteString("\\definecolor{linkcolor}{HTML}{0366d6}\n")
	b.WriteString("\\hypersetup{colorlinks=true, urlcolor=linkcolor, linkcolor=linkcolor}\n")
	b.WriteString("\\pagestyle{empty}\n\n")

	b.WriteString("\\begin{document}\n\n")

	// Header: name on first line, contact details on second
	b.WriteString("{\\LARGE \\textbf{")
	b.WriteString(escapedName)
	b.WriteString("}} \\\\\n")
	b.WriteString("[4pt]\n")
	b.WriteString("\\small\n")

	// Build contact line with separators
	var contactParts []string
	if phone != "" {
		contactParts = append(contactParts, escapeLatex(phone))
	}
	if email != "" {
		contactParts = append(contactParts, "\\href{mailto:"+escapeLatex(email)+"}{"+escapeLatex(email)+"}")
	}
	if linkedin != "" {
		contactParts = append(contactParts, "\\href{https://linkedin.com/in/"+escapeLatex(linkedin)+"}{linkedin.com/in/"+escapeLatex(linkedin)+"}")
	}
	if github != "" {
		contactParts = append(contactParts, "\\href{https://github.com/"+escapeLatex(github)+"}{github.com/"+escapeLatex(github)+"}")
	}
	if len(contactParts) > 0 {
		b.WriteString(strings.Join(contactParts, " \\quad $\\cdot$ \\quad "))
		b.WriteString("\n")
	}

	b.WriteString("\\vspace{4pt}\n")
	b.WriteString("\\noindent\\rule{\\linewidth}{0.4pt}\n")
	b.WriteString("\\vspace{6pt}\n\n")

	// Date
	b.WriteString("\\today\n\n")

	// Addressee
	b.WriteString("\\vspace{10pt}\n")
	b.WriteString("\\textbf{Hiring Manager} \\\\\n")
	b.WriteString("\\textbf{Company Name}\n\n")

	// Salutation
	b.WriteString("\\vspace{14pt}\n")
	b.WriteString("Dear Hiring Manager,\n\n")

	// Body paragraphs
	b.WriteString("\\vspace{6pt}\n")
	for i, para := range paras {
		b.WriteString(para)
		b.WriteString("\n")
		if i < len(paras)-1 {
			b.WriteString("\n\\vspace{8pt}\n")
		}
	}

	// Closing
	b.WriteString("\n\\vspace{14pt}\n")
	b.WriteString("Sincerely,\n\n")
	b.WriteString("\\vspace{28pt}\n")
	b.WriteString("\\textbf{")
	b.WriteString(escapedName)
	b.WriteString("}\n\n")

	b.WriteString("\\end{document}\n")
	return b.String()
}
