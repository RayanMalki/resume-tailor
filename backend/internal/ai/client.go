package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

// ATSReport represents the ATS scoring report
type ATSReport struct {
	Score              float64             `json:"score"`
	Notes              []string            `json:"notes"`
	Summary            string              `json:"summary"`
	InterviewQuestions []InterviewQuestion `json:"interview_questions,omitempty"`
}

// ChangePlan represents the recommended changes
type ChangePlan struct {
	Changes []string `json:"changes"`
}

type InterviewQuestion struct {
	Question   string   `json:"question"`
	AnswerSTAR []string `json:"answer_star"`
}

type ResumeSpec struct {
	Name        string             `json:"name"`
	Title       string             `json:"title"`
	Language    string             `json:"language"`
	Contact     []string           `json:"contact"`
	Summary     []string           `json:"summary"`
	Experience  []ResumeExperience `json:"experience"`
	Projects    []ResumeProject    `json:"projects"`
	Education   []ResumeEducation  `json:"education"`
	SkillGroups []ResumeSkillGroup `json:"skill_groups"`
	Skills      []string           `json:"skills"`
}

type ResumeSkillGroup struct {
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

type ProjectControl struct {
	Name string `json:"name"`
	Mode string `json:"mode"`
}

type ProjectReason struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type ResumeExperience struct {
	Company  string   `json:"company"`
	Role     string   `json:"role"`
	Location string   `json:"location"`
	Dates    string   `json:"dates"`
	Bullets  []string `json:"bullets"`
}

type ResumeProject struct {
	Name    string   `json:"name"`
	Stack   string   `json:"stack"`
	Dates   string   `json:"dates"`
	Bullets []string `json:"bullets"`
}

type ResumeEducation struct {
	School   string   `json:"school"`
	Degree   string   `json:"degree"`
	Location string   `json:"location"`
	Dates    string   `json:"dates"`
	Details  []string `json:"details"`
}

// ResumeSpecToText converts a ResumeSpec to plain text for BM25 analysis.
// Includes section headers (Formation, Expérience, etc.) so they can match job keywords.
func ResumeSpecToText(spec ResumeSpec) string {
	var b strings.Builder
	b.WriteString(spec.Name + "\n")
	b.WriteString(spec.Title + "\n")

	// Contact info (contains github, linkedin, email, etc.)
	for _, c := range spec.Contact {
		b.WriteString(c + "\n")
	}

	for _, s := range spec.Summary {
		b.WriteString(s + "\n")
	}

	// Section header "Formation" / "Education"
	b.WriteString("Formation Education\n")
	for _, ed := range spec.Education {
		b.WriteString(ed.School + " " + ed.Degree + " " + ed.Location + "\n")
		for _, d := range ed.Details {
			b.WriteString(d + "\n")
		}
	}

	// Section header "Compétences" / "Skills"
	b.WriteString("Connaissances techniques Skills\n")
	for _, sg := range spec.SkillGroups {
		b.WriteString(sg.Name + ": " + strings.Join(sg.Items, ", ") + "\n")
	}
	for _, s := range spec.Skills {
		b.WriteString(s + " ")
	}
	b.WriteString("\n")

	// Section header "Projets" / "Projects"
	b.WriteString("Projets Projects\n")
	for _, p := range spec.Projects {
		b.WriteString(p.Name + " " + p.Stack + "\n")
		for _, bullet := range p.Bullets {
			b.WriteString(bullet + "\n")
		}
	}

	// Section header "Expérience" / "Experience"
	b.WriteString("Expérience de travail Experience\n")
	for _, e := range spec.Experience {
		b.WriteString(e.Company + " " + e.Role + " " + e.Location + "\n")
		for _, bullet := range e.Bullets {
			b.WriteString(bullet + "\n")
		}
	}

	return b.String()
}

// ReportResponse is the expected JSON structure from OpenAI
type ReportResponse struct {
	ATSReport  ATSReport  `json:"ats_report"`
	ChangePlan ChangePlan `json:"change_plan"`
}

// Client wraps the OpenAI client
type Client struct {
	client openai.Client
	model  string
}

// NewClientFromEnv creates a new OpenAI client from environment variables
func NewClientFromEnv(apiKey, model string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}

	oc := openai.NewClient(option.WithAPIKey(apiKey))

	return &Client{
		client: oc,
		model:  model,
	}, nil
}

// GenerateRunReport generates a short ATS summary and notes using OpenAI.
// It does NOT generate a change plan — that's computed programmatically from BM25 diffs.
// addedKeywords = terms that moved from missing to matched (real changes).
// missingTerms = terms still missing from the tailored resume.
func (c *Client) GenerateRunReport(ctx context.Context, addedKeywords []string, missingTerms []string, resumeLang string) (ATSReport, ChangePlan, error) {
	prompt := buildReportPrompt(addedKeywords, missingTerms, resumeLang)

	req := openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(
				"You write short ATS compatibility summaries. " +
					"You MUST write ALL output in the language specified. No exceptions."),
			openai.UserMessage(prompt),
		},
		Temperature:         openai.Float(0.3),
		MaxCompletionTokens: openai.Int(500),
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: func() *shared.ResponseFormatJSONObjectParam {
				p := shared.NewResponseFormatJSONObjectParam()
				return &p
			}(),
		},
	}

	resp, err := c.client.Chat.Completions.New(ctx, req)
	if err != nil {
		return ATSReport{}, ChangePlan{}, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return ATSReport{}, ChangePlan{}, fmt.Errorf("no choices in OpenAI response")
	}

	content := resp.Choices[0].Message.Content
	if content == "" {
		return ATSReport{}, ChangePlan{}, fmt.Errorf("empty content in OpenAI response")
	}

	var reportResp ReportResponse
	if err := json.Unmarshal([]byte(content), &reportResp); err != nil {
		return ATSReport{}, ChangePlan{}, fmt.Errorf("failed to parse OpenAI JSON response: %w", err)
	}

	return reportResp.ATSReport, reportResp.ChangePlan, nil
}

func buildReportPrompt(addedKeywords, missingTerms []string, lang string) string {
	var b strings.Builder

	b.WriteString("LANGUAGE: Write ALL text in " + lang + ". Every string MUST be in " + lang + ".\n\n")

	b.WriteString("Based on the keyword analysis below, write a short ATS summary.\n\n")

	if len(addedKeywords) > 0 {
		b.WriteString("KEYWORDS ADDED by tailoring: " + strings.Join(addedKeywords, ", ") + "\n")
	} else {
		b.WriteString("KEYWORDS ADDED by tailoring: none (resume was already well-matched)\n")
	}
	if len(missingTerms) > 0 {
		b.WriteString("KEYWORDS STILL MISSING: " + strings.Join(missingTerms, ", ") + "\n\n")
	} else {
		b.WriteString("KEYWORDS STILL MISSING: none\n\n")
	}

	b.WriteString("RULES:\n")
	b.WriteString("- 'summary': 2-3 sentences. If no keywords were added, say the resume was already well-suited and explain what's still missing.\n")
	b.WriteString("- 'notes': 2-3 short observations about ATS compatibility.\n")
	b.WriteString("- 'interview_questions': exactly 5 likely interview questions for this role.\n")
	b.WriteString("- For each interview question, provide STAR bullet answers in 4 bullets: Situation, Task, Action, Result.\n")
	b.WriteString("- STAR bullets must be grounded in resume evidence and remain truthful.\n")
	b.WriteString("- Do NOT describe formatting changes, reorganization, or title changes unless keywords tell you so.\n")
	b.WriteString("- Be factual. Only mention what the keyword data shows.\n\n")

	b.WriteString("Respond with JSON:\n")
	b.WriteString(`{
  "ats_report": {
    "score": 0.0,
    "notes": ["<string>", ...],
    "summary": "<string>",
    "interview_questions": [
      {
        "question": "<string>",
        "answer_star": [
          "Situation: <string>",
          "Task: <string>",
          "Action: <string>",
          "Result: <string>"
        ]
      }
    ]
  },
  "change_plan": {
    "changes": []
  }
}`)

	return b.String()
}

// GenerateResumeLatex generates a Jake's Resume-style LaTeX output tailored to the job.
func (c *Client) GenerateResumeLatex(ctx context.Context, resumeText, jobText string, bm25Signals any) (string, error) {
	prompt := buildResumeLatexPrompt(resumeText, jobText, bm25Signals)

	req := openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a resume writer. Return ONLY LaTeX code, no commentary."),
			openai.UserMessage(prompt),
		},
	}

	resp, err := c.client.Chat.Completions.New(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI response")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("empty content in OpenAI response")
	}

	return content, nil
}

func buildResumeLatexPrompt(resumeText, jobText string, bm25Signals any) string {
	var b strings.Builder
	b.WriteString("Return ONLY LaTeX. No ``` fences. No explanations.\\n")
	b.WriteString("Use Jake's Resume-style one-column layout with compact sections and bullets.\\n")
	b.WriteString("Sections in this order: Education, Experience, Projects, Skills.\\n")
	b.WriteString("Keep bullet points concise and impact-focused.\\n")
	b.WriteString("The resume MUST fit on ONE page. If needed, reduce bullets, shorten phrasing, or drop least-relevant items to stay on one page.\\n")
	b.WriteString("Use the pattern: \"Did X using Y resulting in Z\" for experience and project bullets.\\n")
	b.WriteString("Bold technical skills/keywords (e.g., languages, frameworks, tools, platforms) using \\\\textbf{...}, especially in project/experience bullets. Do not bold non-technical words.\\n")
	b.WriteString("Tailor to the job description. Use the resume content as the source.\\n\\n")

	b.WriteString("RESUME:\\n")
	b.WriteString(resumeText)
	b.WriteString("\\n\\n")

	b.WriteString("JOB DESCRIPTION:\\n")
	b.WriteString(jobText)
	b.WriteString("\\n\\n")

	if bm25Signals != nil {
		b.WriteString("BM25 SIGNALS:\\n")
		serialized, err := json.MarshalIndent(bm25Signals, "", "  ")
		if err != nil {
			b.WriteString("(BM25 analysis available, failed to serialize)\\n\\n")
		} else {
			b.WriteString(string(serialized))
			b.WriteString("\\n\\n")
		}
	}

	b.WriteString("Return LaTeX only.")
	return b.String()
}

// GenerateResumeSpec generates a strict JSON resume spec for a fixed template.
func (c *Client) GenerateResumeSpec(ctx context.Context, resumeText, jobText string, bm25Signals any, projectControls []ProjectControl) (ResumeSpec, error) {
	prompt := buildResumeSpecPrompt(resumeText, jobText, bm25Signals, projectControls)

	req := openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(
				"You are an expert resume writer who specializes in creating ATS-optimized resumes. " +
					"You understand how applicant tracking systems parse resumes and which keywords matter most. " +
					"You preserve the candidate's real experience and achievements while tailoring language " +
					"and emphasis to match the target role. Return ONLY JSON in the requested format."),
			openai.UserMessage(prompt),
		},
		Temperature:         openai.Float(0.3),
		MaxCompletionTokens: openai.Int(4000),
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: func() *shared.ResponseFormatJSONObjectParam {
				p := shared.NewResponseFormatJSONObjectParam()
				return &p
			}(),
		},
	}

	resp, err := c.client.Chat.Completions.New(ctx, req)
	if err != nil {
		return ResumeSpec{}, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return ResumeSpec{}, fmt.Errorf("no choices in OpenAI response")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	if content == "" {
		return ResumeSpec{}, fmt.Errorf("empty content in OpenAI response")
	}

	var spec ResumeSpec
	if err := json.Unmarshal([]byte(content), &spec); err != nil {
		return ResumeSpec{}, fmt.Errorf("failed to parse resume JSON: %w", err)
	}

	return spec, nil
}

func buildResumeSpecPrompt(resumeText, jobText string, bm25Signals any, projectControls []ProjectControl) string {
	var b strings.Builder
	b.WriteString("Return ONLY JSON. No LaTeX. No commentary. Use ASCII text only.\n")
	b.WriteString("Create a one-page resume spec tailored to the job.\n")
	b.WriteString("Style target: Jake's resume look, similar to the provided candidate template.\n")
	b.WriteString("Section order MUST be: Education, Skills, Projects, Experience.\n")
	b.WriteString("Use the resume's primary language for ALL section content.\n")
	b.WriteString("Do NOT switch language to match the job posting language.\n")
	b.WriteString("Constraints:\n")
	b.WriteString("- Experience: include all relevant roles, up to 8 roles, typically 4-6 bullets each\n")
	b.WriteString("- Projects: include EVERY project from the resume that is relevant to the job, typically 4-8 bullets each\n")
	b.WriteString("- Education: up to 3 entries\n")
	b.WriteString("- Skills: group skills into categories (for example: Languages, Technologies, Concepts)\n")
	b.WriteString("Preserve the resume's level of detail and density; do not output a sparse resume.\n")
	b.WriteString("Do not drop strong achievements, technical depth, or quantified impact from relevant items.\n")
	b.WriteString("Keep bullets concise but substantive (about 18-32 words) and impact-oriented.\n")
	b.WriteString("Use a pattern similar to: action + technology + result/impact.\n")
	b.WriteString("Bold technical keywords in bullets (languages, frameworks, protocols, tools, platforms).\n")
	b.WriteString("Use the resume content as the ONLY source of truth. Do not invent companies, projects, degrees, dates, achievements, or skills.\n")
	b.WriteString("NEVER add a technology, tool, platform, or skill that is not already present in the original resume.\n")
	b.WriteString("You may only add synonym forms of existing skills (e.g. 'Golang' if 'Go' is present, 'Cloud' if 'Infonuagique' is present).\n")
	b.WriteString("Target one page, but prioritize preserving relevant content quality over aggressive trimming.\n")
	b.WriteString("If space is tight, shorten wording before removing relevant projects or experiences.\n\n")

	b.WriteString("RESUME:\n")
	b.WriteString(resumeText)
	b.WriteString("\n\n")

	b.WriteString("JOB DESCRIPTION:\n")
	b.WriteString(jobText)
	b.WriteString("\n\n")

	if bm25Signals != nil {
		b.WriteString("BM25 KEYWORD ANALYSIS:\n")
		b.WriteString("These signals rank keywords by importance (higher score = more important to the job).\n")
		b.WriteString("- missing_job_terms: keywords in the job that are ABSENT from the resume — YOU MUST incorporate these\n")
		b.WriteString("- overlap_terms: keywords already present in both — make sure these stay prominent\n")
		b.WriteString("- top_job_terms: the most important job keywords overall\n\n")

		b.WriteString("CRITICAL KEYWORD INSTRUCTIONS:\n")
		b.WriteString("- NEVER invent skills, technologies, or experiences that are NOT in the original resume.\n")
		b.WriteString("- If a missing keyword is a SYNONYM or alternate name for something already in the resume, add it.\n")
		b.WriteString("  Examples: 'Go' in resume → add 'Golang'; 'Méthodologies agiles' → add 'Scrum' if context fits; 'Infonuagique' → add 'Cloud'.\n")
		b.WriteString("- If a missing keyword is genuinely NOT in the candidate's background, do NOT add it. For example, if the resume has no Kubernetes experience, do NOT add Kubernetes just because the job mentions it.\n")
		b.WriteString("- For overlap_terms: make sure these stay prominent and well-placed.\n")
		b.WriteString("- Rephrase existing bullets to naturally highlight relevant keywords already present in the resume.\n")
		b.WriteString("- The goal is to maximize keyword overlap for ATS parsing while staying 100% truthful to the candidate's actual experience.\n\n")

		serialized, err := json.MarshalIndent(bm25Signals, "", "  ")
		if err != nil {
			b.WriteString("(BM25 analysis available, failed to serialize)\n\n")
		} else {
			b.WriteString(string(serialized))
			b.WriteString("\n\n")
		}
	}

	if len(projectControls) > 0 {
		b.WriteString("PROJECT CONTROLS:\n")
		serialized, err := json.MarshalIndent(projectControls, "", "  ")
		if err != nil {
			b.WriteString("(project controls provided, failed to serialize)\n\n")
		} else {
			b.WriteString(string(serialized))
			b.WriteString("\n\n")
		}
		b.WriteString("Control rules:\n")
		b.WriteString("- mode=pinned: project MUST be included if found in source resume\n")
		b.WriteString("- mode=exclude: project MUST NOT appear in output\n")
		b.WriteString("- mode=auto: include when relevant\n\n")
	}

	b.WriteString("Return JSON in this exact format:\n")
	b.WriteString(`{
  "name": "",
  "title": "",
  "language": "",
  "contact": ["email", "phone", "linkedin", "github", "website"],
  "summary": [],
  "education": [
    {
      "school": "",
      "degree": "",
      "location": "",
		"dates": "",
      "details": [""]
    }
  ],
  "skill_groups": [
    {
      "name": "",
      "items": ["", ""]
    }
  ],
  "experience": [
    {
      "company": "",
      "role": "",
      "location": "",
      "dates": "",
      "bullets": ["", ""]
    }
  ],
  "projects": [
    {
      "name": "",
      "stack": "",
      "dates": "",
      "bullets": ["", ""]
    }
  ],
  "skills": []
}`)
	return b.String()
}

func (c *Client) GenerateProjectReasons(ctx context.Context, resumeText, jobText, latex string, bm25Signals any, projectControls []ProjectControl) ([]ProjectReason, error) {
	prompt := buildProjectReasonsPrompt(resumeText, jobText, latex, bm25Signals, projectControls)

	req := openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a resume analyst who explains project selection decisions. Return ONLY JSON in the requested format."),
			openai.UserMessage(prompt),
		},
		Temperature:         openai.Float(0.3),
		MaxCompletionTokens: openai.Int(1500),
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: func() *shared.ResponseFormatJSONObjectParam {
				p := shared.NewResponseFormatJSONObjectParam()
				return &p
			}(),
		},
	}

	resp, err := c.client.Chat.Completions.New(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}
	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	if content == "" {
		return nil, fmt.Errorf("empty content in OpenAI response")
	}

	var parsed struct {
		Projects []ProjectReason `json:"projects"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse project reasons JSON: %w", err)
	}
	return parsed.Projects, nil
}

func buildProjectReasonsPrompt(resumeText, jobText, latex string, bm25Signals any, projectControls []ProjectControl) string {
	var b strings.Builder
	b.WriteString("Return ONLY JSON. No LaTeX. No commentary.\n")
	b.WriteString("Explain ONLY accepted/included projects and why each is relevant to the job.\n")
	b.WriteString("Use the job description, resume source, BM25 signals, project controls, and final LaTeX.\n")
	b.WriteString("Rules:\n")
	b.WriteString("- Include only projects that are present in final LaTeX.\n")
	b.WriteString("- One concise but specific reason per project (18-45 words).\n")
	b.WriteString("- Mention concrete alignment: stack, domain, impact, keywords, responsibilities.\n")
	b.WriteString("- If a project was pinned and included, mention that it was user-pinned while still stating technical relevance.\n\n")

	b.WriteString("PROJECT CONTROLS:\n")
	controls, err := json.MarshalIndent(projectControls, "", "  ")
	if err != nil {
		b.WriteString("[]\n\n")
	} else {
		b.WriteString(string(controls))
		b.WriteString("\n\n")
	}

	b.WriteString("RESUME:\n")
	b.WriteString(resumeText)
	b.WriteString("\n\n")
	b.WriteString("JOB DESCRIPTION:\n")
	b.WriteString(jobText)
	b.WriteString("\n\n")
	if bm25Signals != nil {
		b.WriteString("BM25 SIGNALS:\n")
		serialized, err := json.MarshalIndent(bm25Signals, "", "  ")
		if err != nil {
			b.WriteString("(BM25 analysis available, failed to serialize)\n\n")
		} else {
			b.WriteString(string(serialized))
			b.WriteString("\n\n")
		}
	}
	b.WriteString("FINAL LATEX:\n")
	b.WriteString(latex)
	b.WriteString("\n\n")

	b.WriteString("Return JSON in this format:\n")
	b.WriteString(`{
  "projects": [
    {
      "name": "",
      "reason": ""
    }
  ]
}`)
	return b.String()
}

func (c *Client) GenerateCoverLetter(ctx context.Context, resumeText, jobText string, bm25Signals any, lang string) (string, error) {
	prompt := buildCoverLetterPrompt(resumeText, jobText, bm25Signals, lang)

	req := openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(
				"You write concise, truthful, role-targeted cover letters. " +
					"Return ONLY JSON in the requested format."),
			openai.UserMessage(prompt),
		},
		Temperature:         openai.Float(0.35),
		MaxCompletionTokens: openai.Int(1400),
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: func() *shared.ResponseFormatJSONObjectParam {
				p := shared.NewResponseFormatJSONObjectParam()
				return &p
			}(),
		},
	}

	resp, err := c.client.Chat.Completions.New(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI response")
	}
	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("empty content in OpenAI response")
	}

	var parsed struct {
		CoverLetter string `json:"cover_letter"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return "", fmt.Errorf("failed to parse cover letter JSON: %w", err)
	}
	if strings.TrimSpace(parsed.CoverLetter) == "" {
		return "", fmt.Errorf("cover letter generation returned empty content")
	}
	return strings.TrimSpace(parsed.CoverLetter), nil
}

func buildCoverLetterPrompt(resumeText, jobText string, bm25Signals any, lang string) string {
	var b strings.Builder
	b.WriteString("Return ONLY JSON. No markdown. No commentary.\n")
	b.WriteString("Write a one-page cover letter grounded strictly in the candidate resume and target job.\n")
	b.WriteString("LANGUAGE: All output must be in " + lang + ".\n")
	b.WriteString("Tone: professional, specific, concise. Avoid generic buzzwords.\n")
	b.WriteString("Rules:\n")
	b.WriteString("- 4 to 6 paragraphs, each 2-4 sentences.\n")
	b.WriteString("- Mention concrete resume evidence (projects, stack, outcomes) that matches the role.\n")
	b.WriteString("- Do not invent achievements, metrics, or technologies not in the resume.\n")
	b.WriteString("- Include a short close paragraph with interview interest.\n\n")

	b.WriteString("RESUME:\n")
	b.WriteString(resumeText)
	b.WriteString("\n\n")
	b.WriteString("JOB DESCRIPTION:\n")
	b.WriteString(jobText)
	b.WriteString("\n\n")
	if bm25Signals != nil {
		b.WriteString("BM25 SIGNALS:\n")
		serialized, err := json.MarshalIndent(bm25Signals, "", "  ")
		if err != nil {
			b.WriteString("(BM25 analysis available, failed to serialize)\n\n")
		} else {
			b.WriteString(string(serialized))
			b.WriteString("\n\n")
		}
	}

	b.WriteString("Return JSON in this format:\n")
	b.WriteString(`{
  "cover_letter": ""
}`)
	return b.String()
}
