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
	Score float64  `json:"score"`
	Notes []string `json:"notes"`
}

// ChangePlan represents the recommended changes
type ChangePlan struct {
	Changes []string `json:"changes"`
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

// GenerateRunReport generates an ATS report and change plan using OpenAI
func (c *Client) GenerateRunReport(ctx context.Context, resumeText, jobText string, bm25Signals any) (ATSReport, ChangePlan, error) {
	// Build the prompt
	prompt := c.buildPrompt(resumeText, jobText, bm25Signals)

	// Call OpenAI
	req := openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are an expert ATS (Applicant Tracking System) analyzer. You analyze resumes against job descriptions and provide structured JSON responses."),
			openai.UserMessage(prompt),
		},
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

	// Parse JSON response
	var reportResp ReportResponse
	if err := json.Unmarshal([]byte(content), &reportResp); err != nil {
		return ATSReport{}, ChangePlan{}, fmt.Errorf("failed to parse OpenAI JSON response: %w", err)
	}

	// Validate the response
	if reportResp.ATSReport.Score < 0 || reportResp.ATSReport.Score > 1 {
		return ATSReport{}, ChangePlan{}, fmt.Errorf("invalid ATS score: must be between 0 and 1")
	}

	return reportResp.ATSReport, reportResp.ChangePlan, nil
}

func (c *Client) buildPrompt(resumeText, jobText string, bm25Signals any) string {
	var b strings.Builder

	b.WriteString("Analyze the following resume against the job description and provide:\n")
	b.WriteString("1. An ATS compatibility score (0.0 to 1.0)\n")
	b.WriteString("2. Notes explaining the score\n")
	b.WriteString("3. A change plan with specific recommendations\n\n")

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

	b.WriteString("Respond with a JSON object in this exact format:\n")
	b.WriteString(`{
  "ats_report": {
    "score": <number between 0.0 and 1.0>,
    "notes": ["<string>", ...]
  },
  "change_plan": {
    "changes": ["<string>", ...]
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
