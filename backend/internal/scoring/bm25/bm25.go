package bm25

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	libbm25 "github.com/crawlab-team/bm25"
	"golang.org/x/text/unicode/norm"
)

const (
	defaultK1      = 1.5
	defaultB       = 0.75
	defaultTopN    = 10
	minTokenLength = 2
)

// TermScore represents a term with its BM25-derived score.
type TermScore struct {
	Term  string  `json:"term"`
	Score float64 `json:"score"`
}

// Signals exposes BM25-derived signals for reporting.
type Signals struct {
	TopJobTerms     []TermScore `json:"top_job_terms"`
	MissingJobTerms []TermScore `json:"missing_job_terms"`
	OverlapTerms    []string    `json:"overlap_terms"`
	// Score is IDF-weighted keyword coverage in [0,1]:
	// matched job-term weight / total job-term weight.
	Score float64 `json:"score"`
	// BM25Score is the raw 2-document BM25 score (resume scored against job query).
	// It is exposed for diagnostics only and is not used as ATS compatibility score.
	BM25Score float64 `json:"bm25_score"`
}

// Compute calculates BM25 signals for resume and job text matching.
//
// IDF values come from a pre-computed static table built from a large corpus
// of job descriptions, replacing the previous single-document IDF calculation
// which produced meaningless binary scores.
func Compute(resumeText, jobText string) (Signals, error) {
	resumeTokens := tokenize(resumeText)
	jobTokens := tokenize(jobText)

	if len(resumeTokens) == 0 || len(jobTokens) == 0 {
		return Signals{}, nil
	}

	// We still use the library for the overall document score, but now pass
	// both resume and job as separate corpus documents so the library's
	// internal IDF has at least two documents to work with.
	bm25Instance, err := libbm25.NewBM25Okapi(
		[]string{resumeText, jobText},
		tokenize, defaultK1, defaultB, nil,
	)
	if err != nil {
		return Signals{}, fmt.Errorf("bm25 init: %w", err)
	}

	scores, err := bm25Instance.GetScores(jobTokens)
	if err != nil {
		return Signals{}, fmt.Errorf("bm25 score: %w", err)
	}

	// The resume is document 0 in the corpus.
	bm25Score := 0.0
	if len(scores) > 0 {
		bm25Score = scores[0]
	}

	jobFreq := termFreq(jobTokens)
	resumeFreq := termFreq(resumeTokens)

	overlapTerms := make([]string, 0)
	missingTerms := make([]TermScore, 0)
	topTerms := make([]TermScore, 0, len(jobFreq))
	totalWeight := 0.0
	matchedWeight := 0.0

	for term, qtf := range jobFreq {
		tf := resumeFreq[term]

		// Job-side importance is based on static corpus IDF and query frequency.
		termIDF := lookupIDF(term)
		importance := termIDF * float64(qtf)
		totalWeight += importance

		if tf > 0 {
			overlapTerms = append(overlapTerms, term)
			matchedWeight += importance
		} else {
			missingTerms = append(missingTerms, TermScore{
				Term:  term,
				Score: importance,
			})
		}

		topTerms = append(topTerms, TermScore{
			Term:  term,
			Score: importance,
		})
	}

	sortStrings(overlapTerms)
	sortTermScores(topTerms)
	sortTermScores(missingTerms)

	if len(topTerms) > defaultTopN {
		topTerms = topTerms[:defaultTopN]
	}

	coverageScore := 0.0
	if totalWeight > 0 {
		coverageScore = matchedWeight / totalWeight
	}

	return Signals{
		TopJobTerms:     topTerms,
		MissingJobTerms: missingTerms,
		OverlapTerms:    overlapTerms,
		Score:           coverageScore,
		BM25Score:       bm25Score,
	}, nil
}

// stripAccents removes diacritics/accents from text using Unicode NFD
// decomposition (e.g. "développement" → "developpement", "résumé" → "resume").
// This allows French and other accented text to match English IDF table terms.
// It also strips modifier letter accents (ˊ U+02CA, ˋ U+02CB, etc.) which
// some PDF extractors produce instead of proper combining accents.
func stripAccents(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range norm.NFD.String(s) {
		// After NFD decomposition, accents become separate combining marks.
		if unicode.Is(unicode.Mn, r) { // Mn = Mark, Nonspacing (combining accents)
			continue
		}
		// Also strip modifier letter accents (U+02B0–U+02FF) which some
		// PDF extractors emit instead of proper combining marks.
		// Includes ˊ (U+02CA), ˋ (U+02CB), ˆ (U+02C6), etc.
		if r >= 0x02B0 && r <= 0x02FF {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// synonyms maps safe technology aliases to canonical forms so matching treats
// equivalent tokens as the same concept.
var synonyms = map[string]string{
	// Go / Golang
	"go":     "golang",
	"golang": "golang",
	// JavaScript / JS
	"js":         "javascript",
	"javascript": "javascript",
	// TypeScript / TS
	"ts":         "typescript",
	"typescript": "typescript",
	// PostgreSQL variants
	"postgres":   "postgresql",
	"postgresql": "postgresql",
	// Kubernetes / K8s
	"k8s":        "kubernetes",
	"kubernetes": "kubernetes",
	// C# / CSharp
	"csharp": "csharp",
	"dotnet": "dotnet",
	"aspnet": "aspnet",
	"cpp":    "cpp",
	"fsharp": "fsharp",
	// Continuous Integration / Continuous Deployment
	"ci":   "ci",
	"cd":   "cd",
	"cicd": "cicd",
	// ReactJS / React
	"reactjs": "react",
	"react":   "react",
	// NodeJS / Node
	"nodejs": "nodejs",
	"node":   "nodejs",
	// REST / RESTful
	"rest":    "rest",
	"restful": "rest",
	// Amazon Web Services
	"aws": "aws",
	// Google Cloud Platform
	"gcp": "gcp",
	// Infonuagique (French for cloud computing)
	"infonuagique": "cloud",
	"cloud":        "cloud",
	// Agile / Scrum
	"scrum": "scrum",
	"agile": "agile",
	// DevOps
	"devops": "devops",
	// Microservices
	"microservices": "microservice",
	"microservice":  "microservice",
	// API
	"api":  "api",
	"apis": "api",
	// Tests automatisés / automated tests
	"automatise":   "automatise",
	"automatises":  "automatise",
	"automatisee":  "automatise",
	"automatisees": "automatise",
	// Protect tech terms that naturally end in 's' from depluralization
	"jenkins": "jenkins",
	"redis":   "redis",
	"travis":  "travis",
	"atlas":   "atlas",
	"pandas":  "pandas",
	"keras":   "keras",
	"express": "express",
	// "postgres" already defined above
}

func canonicalize(token string) string {
	if canon, ok := synonyms[token]; ok {
		return canon
	}
	return token
}

// depluralize applies simple plural→singular normalization for English and French.
// Only strips trailing "-s". We never strip "-es" (2 chars) because French singulars
// keep the "e" (e.g. "techniques"→"technique", NOT "techniqu").
func depluralize(token string) string {
	n := len(token)
	if n < 4 {
		return token // too short to safely strip
	}

	// Don't touch known tech terms / synonyms — they're already canonical
	if _, ok := synonyms[token]; ok {
		return token
	}

	// Only strip trailing -s
	if !strings.HasSuffix(token, "s") {
		return token
	}
	// Don't strip if it ends in "ss" (e.g. "process", "class")
	if strings.HasSuffix(token, "ss") {
		return token
	}

	candidate := token[:n-1]
	if len(candidate) >= 3 {
		return candidate
	}
	return token
}

func tokenize(text string) []string {
	if text == "" {
		return nil
	}

	// Normalize accents and punctuation-heavy tech forms before tokenizing.
	text = normalizeForTokenization(text)

	var tokens []string
	var b strings.Builder

	flush := func() {
		if b.Len() < minTokenLength {
			b.Reset()
			return
		}
		token := b.String()
		if isStopword(token) {
			b.Reset()
			return
		}
		// Normalize plural → singular, then check stopwords again
		// (e.g. "fonctions" → "fonction" which might be a stopword)
		token = depluralize(token)
		if isStopword(token) {
			b.Reset()
			return
		}
		tokens = append(tokens, canonicalize(token))
		b.Reset()
	}

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			continue
		}
		flush()
	}
	flush()

	return tokens
}

func termFreq(tokens []string) map[string]int {
	freq := make(map[string]int, len(tokens))
	for _, t := range tokens {
		freq[t]++
	}
	return freq
}

var techTokenReplacer = strings.NewReplacer(
	"asp.net", " aspnet ",
	"next.js", " nextjs ",
	"node.js", " nodejs ",
	"nuxt.js", " nuxtjs ",
	"c++", " cpp ",
	"c#", " csharp ",
	"f#", " fsharp ",
	".net", " dotnet ",
	"ci/cd", " cicd ",
	"ci-cd", " cicd ",
)

func normalizeForTokenization(text string) string {
	text = strings.ToLower(stripAccents(text))
	return techTokenReplacer.Replace(text)
}

func sortStrings(values []string) {
	sort.Strings(values)
}

func sortTermScores(items []TermScore) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Score == items[j].Score {
			return items[i].Term < items[j].Term
		}
		return items[i].Score > items[j].Score
	})
}

// stopwords and isStopword are defined in stopwords.go using comprehensive
// ISO 639 stopword lists for English and French, plus custom job-posting boilerplate.
