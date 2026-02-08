package latex

import (
	"strings"

	"resume-tailor/internal/ai"
)

type RenderLimits struct {
	MaxExperience     int
	MaxExpBullets     int
	MaxProjBullets    int
	MaxEducation      int
	MaxSkillGroups    int
	MaxSkillsPerGroup int
	MaxSkillsFallback int
}

func RenderResume(spec ai.ResumeSpec) string {
	limits := RenderLimits{
		MaxExperience:     5,
		MaxExpBullets:     4,
		MaxProjBullets:    4,
		MaxEducation:      2,
		MaxSkillGroups:    6,
		MaxSkillsPerGroup: 20,
		MaxSkillsFallback: 24,
	}

	name := escapeLatex(fallback(spec.Name, "Candidate"))
	contact := joinAndEscape(spec.Contact, " \\quad ")
	education := clampEducation(spec.Education, limits.MaxEducation)
	experience := clampExperience(spec.Experience, limits.MaxExperience, limits.MaxExpBullets)
	projects := clampProjects(spec.Projects, limits.MaxProjBullets)
	skillGroups := clampSkillGroups(spec.SkillGroups, limits.MaxSkillGroups, limits.MaxSkillsPerGroup)
	skillsFallback := clampAndEscape(spec.Skills, limits.MaxSkillsFallback)
	sections := sectionLabels(spec.Language)

	var b strings.Builder
	b.WriteString("\\documentclass[letterpaper,11pt]{article}\n\n")
	b.WriteString("\\usepackage{latexsym}\n")
	b.WriteString("\\usepackage[empty]{fullpage}\n")
	b.WriteString("\\usepackage{titlesec}\n")
	b.WriteString("\\usepackage{marvosym}\n")
	b.WriteString("\\usepackage[usenames,dvipsnames]{color}\n")
	b.WriteString("\\usepackage{verbatim}\n")
	b.WriteString("\\usepackage{enumitem}\n")
	b.WriteString("\\usepackage[hidelinks]{hyperref}\n")
	b.WriteString("\\usepackage{fancyhdr}\n")
	b.WriteString("\\usepackage[english]{babel}\n")
	b.WriteString("\\usepackage{tabularx}\n")
	b.WriteString("\\usepackage[default]{lato}\n\n")

	b.WriteString("\\pagestyle{fancy}\n")
	b.WriteString("\\fancyhf{}\n")
	b.WriteString("\\fancyfoot{}\n")
	b.WriteString("\\renewcommand{\\headrulewidth}{0pt}\n")
	b.WriteString("\\renewcommand{\\footrulewidth}{0pt}\n\n")
	b.WriteString("\\addtolength{\\oddsidemargin}{-0.5in}\n")
	b.WriteString("\\addtolength{\\evensidemargin}{-0.5in}\n")
	b.WriteString("\\addtolength{\\textwidth}{1in}\n")
	b.WriteString("\\addtolength{\\topmargin}{-.5in}\n")
	b.WriteString("\\addtolength{\\textheight}{1.0in}\n\n")
	b.WriteString("\\urlstyle{same}\n")
	b.WriteString("\\raggedbottom\n")
	b.WriteString("\\raggedright\n")
	b.WriteString("\\setlength{\\tabcolsep}{0in}\n\n")
	b.WriteString("\\titleformat{\\section}{\n")
	b.WriteString("  \\vspace{-4pt}\\scshape\\raggedright\\large\n")
	b.WriteString("}{}{0em}{}[\\color{black}\\titlerule\\vspace{-5pt}]\n")
	b.WriteString("\\ifdefined\\pdfgentounicode\n")
	b.WriteString("  \\pdfgentounicode=1\n")
	b.WriteString("\\fi\n\n")

	b.WriteString("\\newcommand{\\resumeItem}[1]{\n")
	b.WriteString("  \\item\\small{{#1 \\vspace{-2pt}}}\n")
	b.WriteString("}\n\n")
	b.WriteString("\\newcommand{\\resumeSubheading}[4]{\n")
	b.WriteString("  \\vspace{-2pt}\\item\n")
	b.WriteString("    \\begin{tabular*}{0.97\\textwidth}[t]{l@{\\extracolsep{\\fill}}r}\n")
	b.WriteString("      \\textbf{#1} & #2 \\\\\n")
	b.WriteString("      \\textit{\\small#3} & \\textit{\\small #4} \\\\\n")
	b.WriteString("    \\end{tabular*}\\vspace{-7pt}\n")
	b.WriteString("}\n\n")
	b.WriteString("\\newcommand{\\resumeProjectHeading}[2]{\n")
	b.WriteString("    \\item\n")
	b.WriteString("    \\begin{tabular*}{0.97\\textwidth}{l@{\\extracolsep{\\fill}}r}\n")
	b.WriteString("      \\small#1 & #2 \\\\\n")
	b.WriteString("    \\end{tabular*}\\vspace{-7pt}\n")
	b.WriteString("}\n\n")
	b.WriteString("\\newcommand{\\resumeSubHeadingListStart}{\\begin{itemize}[leftmargin=0.15in, label={}]}\n")
	b.WriteString("\\newcommand{\\resumeSubHeadingListEnd}{\\end{itemize}}\n")
	b.WriteString("\\newcommand{\\resumeItemListStart}{\\begin{itemize}}\n")
	b.WriteString("\\newcommand{\\resumeItemListEnd}{\\end{itemize}\\vspace{-5pt}}\n\n")

	b.WriteString("\\begin{document}\n\n")
	b.WriteString("\\begin{center}\n")
	b.WriteString("    \\textbf{\\Huge \\scshape ")
	b.WriteString(name)
	b.WriteString("} \\\\\\vspace{1pt}\n")
	if contact != "" {
		b.WriteString("    \\small ")
		b.WriteString(contact)
		b.WriteString("\n")
	}
	b.WriteString("\\end{center}\n\n")

	if len(education) > 0 {
		b.WriteString("\\section{")
		b.WriteString(escapeLatex(sections.Education))
		b.WriteString("}\n")
		b.WriteString("\\resumeSubHeadingListStart\n")
		for _, edu := range education {
			b.WriteString("    \\resumeSubheading\n")
			b.WriteString("      {")
			b.WriteString(edu.Degree)
			b.WriteString("}{")
			b.WriteString(edu.Dates)
			b.WriteString("}\n")
			b.WriteString("      {")
			b.WriteString(edu.School)
			b.WriteString("}{")
			b.WriteString(edu.Location)
			b.WriteString("}\n")
			if len(edu.Details) > 0 {
				b.WriteString("      \\resumeItemListStart\n")
				for _, detail := range edu.Details {
					b.WriteString("        \\resumeItem{")
					b.WriteString(detail)
					b.WriteString("}\n")
				}
				b.WriteString("      \\resumeItemListEnd\n")
			}
			b.WriteString("\n")
		}
		b.WriteString("\\resumeSubHeadingListEnd\n\n")
	}

	if len(skillGroups) > 0 || len(skillsFallback) > 0 {
		b.WriteString("\\section{")
		b.WriteString(escapeLatex(sections.Skills))
		b.WriteString("}\n")
		b.WriteString("    \\begin{itemize}[leftmargin=0.15in, label={}]\n")
		b.WriteString("        \\small{\\item{\n")
		if len(skillGroups) > 0 {
			for i, group := range skillGroups {
				if i > 0 {
					b.WriteString(" \\\\\n")
				}
				b.WriteString("            \\textbf{")
				b.WriteString(group.Name)
				b.WriteString("}{: ")
				b.WriteString(strings.Join(group.Items, ", "))
				b.WriteString("}")
			}
		} else {
			b.WriteString("            \\textbf{Skills}{: ")
			b.WriteString(strings.Join(skillsFallback, ", "))
			b.WriteString("}")
		}
		b.WriteString("\n")
		b.WriteString("        }}\n")
		b.WriteString("    \\end{itemize}\n\n")
	}

	if len(projects) > 0 {
		b.WriteString("\\section{")
		b.WriteString(escapeLatex(sections.Projects))
		b.WriteString("}\n")
		b.WriteString("\\resumeSubHeadingListStart\n")
		for _, proj := range projects {
			b.WriteString("    \\resumeProjectHeading\n")
			b.WriteString("      {\\textbf{")
			b.WriteString(proj.Name)
			if proj.Stack != "" {
				b.WriteString("} $|$ \\emph{")
				b.WriteString(proj.Stack)
			}
			b.WriteString("}}{")
			b.WriteString(proj.Dates)
			b.WriteString("}\n")
			if len(proj.Bullets) > 0 {
				b.WriteString("      \\resumeItemListStart\n")
				for _, bullet := range proj.Bullets {
					b.WriteString("        \\resumeItem{")
					b.WriteString(bullet)
					b.WriteString("}\n")
				}
				b.WriteString("      \\resumeItemListEnd\n")
			}
			b.WriteString("\n")
		}
		b.WriteString("\\resumeSubHeadingListEnd\n\n")
	}

	if len(experience) > 0 {
		b.WriteString("\\section{")
		b.WriteString(escapeLatex(sections.Experience))
		b.WriteString("}\n")
		b.WriteString("\\resumeSubHeadingListStart\n")
		for _, exp := range experience {
			b.WriteString("    \\resumeSubheading\n")
			b.WriteString("      {")
			b.WriteString(exp.Company)
			b.WriteString("}{")
			b.WriteString(exp.Location)
			b.WriteString("}\n")
			b.WriteString("      {")
			b.WriteString(exp.Role)
			b.WriteString("}{")
			b.WriteString(exp.Dates)
			b.WriteString("}\n")
			if len(exp.Bullets) > 0 {
				b.WriteString("      \\resumeItemListStart\n")
				for _, bullet := range exp.Bullets {
					b.WriteString("        \\resumeItem{")
					b.WriteString(bullet)
					b.WriteString("}\n")
				}
				b.WriteString("      \\resumeItemListEnd\n")
			}
			b.WriteString("\n")
		}
		b.WriteString("\\resumeSubHeadingListEnd\n\n")
	}

	b.WriteString("\\end{document}\n")
	return b.String()
}

type sectionSet struct {
	Education  string
	Skills     string
	Projects   string
	Experience string
}

func sectionLabels(language string) sectionSet {
	if isFrenchLanguage(language) {
		return sectionSet{
			Education:  "Formation",
			Skills:     "Connaissances techniques",
			Projects:   "Projets",
			Experience: "Experience de travail",
		}
	}
	return sectionSet{
		Education:  "Education",
		Skills:     "Technical Skills",
		Projects:   "Projects",
		Experience: "Professional Experience",
	}
}

func isFrenchLanguage(language string) bool {
	lang := strings.ToLower(strings.TrimSpace(language))
	return strings.HasPrefix(lang, "fr")
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

func clampProjects(items []ai.ResumeProject, maxBullets int) []ai.ResumeProject {
	out := make([]ai.ResumeProject, 0, len(items))
	for _, item := range items {
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

func clampSkillGroups(items []ai.ResumeSkillGroup, maxGroups, maxSkills int) []ai.ResumeSkillGroup {
	out := make([]ai.ResumeSkillGroup, 0, maxGroups)
	for _, item := range items {
		if len(out) >= maxGroups {
			break
		}
		name := escapeLatex(item.Name)
		if name == "" {
			continue
		}
		group := ai.ResumeSkillGroup{
			Name:  name,
			Items: clampAndEscape(item.Items, maxSkills),
		}
		if len(group.Items) == 0 {
			continue
		}
		out = append(out, group)
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
