package latex

import (
	"strings"

	"resume-tailor/internal/ai"
)

type RenderLimits struct {
	MaxSummary     int
	MaxExperience  int
	MaxExpBullets  int
	MaxProjects    int
	MaxProjBullets int
	MaxEducation   int
	MaxSkills      int
}

func RenderResume(spec ai.ResumeSpec) string {
	limits := RenderLimits{
		MaxSummary:     3,
		MaxExperience:  4,
		MaxExpBullets:  3,
		MaxProjects:    3,
		MaxProjBullets: 3,
		MaxEducation:   2,
		MaxSkills:      14,
	}

	name := escapeLatex(fallback(spec.Name, "Candidate"))
	title := escapeLatex(spec.Title)
	contact := joinAndEscape(spec.Contact, " \\textbullet{} ")
	summary := clampAndEscape(spec.Summary, limits.MaxSummary)

	experience := clampExperience(spec.Experience, limits.MaxExperience, limits.MaxExpBullets)
	projects := clampProjects(spec.Projects, limits.MaxProjects, limits.MaxProjBullets)
	education := clampEducation(spec.Education, limits.MaxEducation)
	skills := clampAndEscape(spec.Skills, limits.MaxSkills)

	var b strings.Builder
	b.WriteString("\\documentclass[10pt]{article}\n")
	b.WriteString("\\usepackage[margin=0.6in]{geometry}\n")
	b.WriteString("\\usepackage[T1]{fontenc}\n")
	b.WriteString("\\usepackage[utf8]{inputenc}\n")
	b.WriteString("\\usepackage{enumitem}\n")
	b.WriteString("\\usepackage[hidelinks]{hyperref}\n")
	b.WriteString("\\usepackage{titlesec}\n")
	b.WriteString("\\setlength{\\parindent}{0pt}\n")
	b.WriteString("\\setlist[itemize]{noitemsep, topsep=2pt, leftmargin=*}\n")
	b.WriteString("\\titleformat{\\section}{\\bfseries\\small}{}{0pt}{}\n")
	b.WriteString("\\pagenumbering{gobble}\n")
	b.WriteString("\\begin{document}\n")
	b.WriteString("\\begin{center}\n")
	b.WriteString("{\\LARGE \\textbf{")
	b.WriteString(name)
	b.WriteString("}}\\\\\n")
	if title != "" {
		b.WriteString("{\\small ")
		b.WriteString(title)
		b.WriteString("}\\\\\n")
	}
	if contact != "" {
		b.WriteString("{\\small ")
		b.WriteString(contact)
		b.WriteString("}\\\\\n")
	}
	b.WriteString("\\end{center}\n")

	if len(summary) > 0 {
		section(&b, "Summary")
		bulletList(&b, summary)
	}

	if len(experience) > 0 {
		section(&b, "Experience")
		for _, exp := range experience {
			entryHeader(&b, exp.Role, exp.Company, exp.Location, exp.Dates)
			bulletList(&b, exp.Bullets)
		}
	}

	if len(projects) > 0 {
		section(&b, "Projects")
		for _, proj := range projects {
			projectHeader(&b, proj.Name, proj.Stack, proj.Dates)
			bulletList(&b, proj.Bullets)
		}
	}

	if len(education) > 0 {
		section(&b, "Education")
		for _, edu := range education {
			entryHeader(&b, edu.Degree, edu.School, edu.Location, edu.Dates)
			if len(edu.Details) > 0 {
				bulletList(&b, edu.Details)
			}
		}
	}

	if len(skills) > 0 {
		section(&b, "Skills")
		b.WriteString("\\small ")
		for idx, skill := range skills {
			if idx > 0 {
				b.WriteString(", ")
			}
			b.WriteString("\\textbf{")
			b.WriteString(skill)
			b.WriteString("}")
		}
		b.WriteString("\n")
	}

	b.WriteString("\\end{document}\n")
	return b.String()
}

func section(b *strings.Builder, title string) {
	b.WriteString("\\section{")
	b.WriteString(title)
	b.WriteString("}\n")
}

func entryHeader(b *strings.Builder, title, org, location, dates string) {
	line := strings.TrimSpace(strings.Join([]string{title, org}, " - "))
	b.WriteString("\\textbf{")
	b.WriteString(escapeLatex(line))
	b.WriteString("}")
	if dates != "" {
		b.WriteString(" \\hfill ")
		b.WriteString(escapeLatex(dates))
	}
	b.WriteString("\\\\\n")
	if location != "" {
		b.WriteString("{\\small ")
		b.WriteString(escapeLatex(location))
		b.WriteString("}\\\\\n")
	}
}

func projectHeader(b *strings.Builder, name, stack, dates string) {
	label := escapeLatex(name)
	if stack != "" {
		label = label + " \\textbar{} " + escapeLatex(stack)
	}
	b.WriteString("\\textbf{")
	b.WriteString(label)
	b.WriteString("}")
	if dates != "" {
		b.WriteString(" \\hfill ")
		b.WriteString(escapeLatex(dates))
	}
	b.WriteString("\\\\\n")
}

func bulletList(b *strings.Builder, items []string) {
	if len(items) == 0 {
		return
	}
	b.WriteString("\\begin{itemize}\n")
	for _, item := range items {
		b.WriteString("\\item ")
		b.WriteString(item)
		b.WriteString("\n")
	}
	b.WriteString("\\end{itemize}\n")
}

func clampAndEscape(items []string, max int) []string {
	clean := make([]string, 0, max)
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		clean = append(clean, escapeLatex(trimmed))
		if len(clean) >= max {
			break
		}
	}
	return clean
}

func clampExperience(items []ai.ResumeExperience, maxEntries, maxBullets int) []ai.ResumeExperience {
	out := make([]ai.ResumeExperience, 0, maxEntries)
	for _, item := range items {
		if len(out) >= maxEntries {
			break
		}
		exp := ai.ResumeExperience{
			Company:  escapeLatex(item.Company),
			Role:     escapeLatex(item.Role),
			Location: escapeLatex(item.Location),
			Dates:    escapeLatex(item.Dates),
			Bullets:  clampAndEscape(item.Bullets, maxBullets),
		}
		if exp.Company == "" && exp.Role == "" {
			continue
		}
		out = append(out, exp)
	}
	return out
}

func clampProjects(items []ai.ResumeProject, maxEntries, maxBullets int) []ai.ResumeProject {
	out := make([]ai.ResumeProject, 0, maxEntries)
	for _, item := range items {
		if len(out) >= maxEntries {
			break
		}
		proj := ai.ResumeProject{
			Name:    escapeLatex(item.Name),
			Stack:   escapeLatex(item.Stack),
			Dates:   escapeLatex(item.Dates),
			Bullets: clampAndEscape(item.Bullets, maxBullets),
		}
		if proj.Name == "" {
			continue
		}
		out = append(out, proj)
	}
	return out
}

func clampEducation(items []ai.ResumeEducation, maxEntries int) []ai.ResumeEducation {
	out := make([]ai.ResumeEducation, 0, maxEntries)
	for _, item := range items {
		if len(out) >= maxEntries {
			break
		}
		edu := ai.ResumeEducation{
			School:   escapeLatex(item.School),
			Degree:   escapeLatex(item.Degree),
			Location: escapeLatex(item.Location),
			Dates:    escapeLatex(item.Dates),
			Details:  clampAndEscape(item.Details, 2),
		}
		if edu.School == "" && edu.Degree == "" {
			continue
		}
		out = append(out, edu)
	}
	return out
}

func joinAndEscape(items []string, sep string) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		parts = append(parts, escapeLatex(trimmed))
	}
	return strings.Join(parts, sep)
}

func escapeLatex(input string) string {
	replacer := strings.NewReplacer(
		"\\", "\\textbackslash{}",
		"&", "\\&",
		"%", "\\%",
		"$", "\\$",
		"#", "\\#",
		"_", "\\_",
		"{", "\\{",
		"}", "\\}",
		"~", "\\textasciitilde{}",
		"^", "\\textasciicircum{}",
	)
	out := replacer.Replace(input)
	out = strings.ReplaceAll(out, "\n", " ")
	out = strings.ReplaceAll(out, "\r", " ")
	return strings.TrimSpace(out)
}

func fallback(value, alt string) string {
	if strings.TrimSpace(value) == "" {
		return alt
	}
	return value
}
