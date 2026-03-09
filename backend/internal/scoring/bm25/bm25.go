package bm25

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"resume-tailor/internal/scoring/profiles"

	libbm25 "github.com/crawlab-team/bm25"
	"golang.org/x/text/unicode/norm"
)

const (
	defaultK1      = 1.5
	defaultB       = 0.75
	defaultTopN    = 10
	minTokenLength = 2
	// Terms that fall into "other" are hidden unless they are truly distinctive.
	highIDFOtherThreshold = 6.2
)

// TermScore represents a term with its BM25-derived score.
type TermScore struct {
	Term     string  `json:"term"`
	Score    float64 `json:"score"`
	Category string  `json:"category,omitempty"`
}

// Signals exposes BM25-derived signals for reporting.
type Signals struct {
	TopJobTerms     []TermScore `json:"top_job_terms"`
	MissingJobTerms []TermScore `json:"missing_job_terms"`
	OverlapTerms    []string    `json:"overlap_terms"`
	// BucketedTopTerms groups high-signal terms into deterministic categories.
	BucketedTopTerms map[string][]TermScore `json:"bucketed_top_terms,omitempty"`
	// LowSignalTerms are filtered from top/missing displays to reduce noise.
	LowSignalTerms []TermScore `json:"low_signal_terms,omitempty"`
	// CategoryCoverage reports matched/total weighted coverage per bucket.
	CategoryCoverage map[string]float64 `json:"category_coverage,omitempty"`
	// Discipline is the selected discipline for this scoring pass.
	Discipline string `json:"discipline,omitempty"`
	// DisciplineEvidence are the matched terms that drove discipline selection.
	DisciplineEvidence []TermScore `json:"discipline_evidence,omitempty"`
	// ProfileVersion identifies the profile dictionary version used for scoring.
	ProfileVersion string `json:"profile_version,omitempty"`
	// Score is normalized IDF-weighted keyword coverage in [0,1].
	// Derived from RawCoverage via sqrt transformation for intuitive scaling.
	Score float64 `json:"score"`
	// RawCoverage is the un-normalized matched/total weight ratio for diagnostics.
	RawCoverage float64 `json:"raw_coverage"`
	// BM25Score is the raw 2-document BM25 score (resume scored against job query).
	// It is exposed for diagnostics only and is not used as ATS compatibility score.
	BM25Score float64 `json:"bm25_score"`
}

// NormalizeCoverageScore applies a sqrt curve to lift mid-range coverage
// values into a more intuitive range while preserving 0→0 and 1→1 boundaries.
func NormalizeCoverageScore(raw float64) float64 {
	if raw <= 0 {
		return 0
	}
	if raw >= 1 {
		return 1
	}
	return math.Sqrt(raw)
}

// Compute calculates BM25 signals for resume and job text matching.
//
// IDF values come from a pre-computed static table built from a large corpus
// of job descriptions, replacing the previous single-document IDF calculation
// which produced meaningless binary scores.
func Compute(resumeText, jobText string) (Signals, error) {
	return ComputeWithProfile(resumeText, jobText, profiles.Get(profiles.DefaultDiscipline()))
}

func ComputeWithProfile(resumeText, jobText string, profile profiles.Profile) (Signals, error) {
	if profile.Discipline == "" {
		profile = profiles.Get(profiles.DefaultDiscipline())
	}
	tokenizer := func(text string) []string {
		return tokenizeWithProfile(text, profile)
	}

	resumeTokens := tokenizer(resumeText)
	jobTokens := tokenizer(jobText)

	if len(resumeTokens) == 0 || len(jobTokens) == 0 {
		return Signals{
			Discipline:     string(profile.Discipline),
			ProfileVersion: profile.Version,
		}, nil
	}

	// We still use the library for the overall document score, but now pass
	// both resume and job as separate corpus documents so the library's
	// internal IDF has at least two documents to work with.
	bm25Instance, err := libbm25.NewBM25Okapi(
		[]string{resumeText, jobText},
		tokenizer, defaultK1, defaultB, nil,
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

	normalizedJob := normalizeForTokenization(jobText)
	normalizedResume := normalizeForTokenization(resumeText)
	for phrase, count := range extractPhraseCounts(normalizedJob) {
		jobFreq[phrase] += count
	}
	for phrase, count := range extractPhraseCounts(normalizedResume) {
		resumeFreq[phrase] += count
	}

	overlapTerms := make([]string, 0)
	missingTerms := make([]TermScore, 0)
	topTerms := make([]TermScore, 0, len(jobFreq))
	lowSignalTerms := make([]TermScore, 0)
	buckets := make(map[string][]TermScore, 5)
	bucketTotals := make(map[string]float64, len(profile.Buckets))
	bucketMatched := make(map[string]float64, len(profile.Buckets))
	totalWeight := 0.0
	matchedWeight := 0.0

	for term, qtf := range jobFreq {
		tf := resumeFreq[term]

		// Job-side importance is based on static corpus IDF and query frequency.
		termIDF := lookupIDF(term)
		importance := termIDF * float64(qtf)
		if meta, ok := profile.CanonicalTerms[term]; ok && meta.Weight > 0 {
			importance *= meta.Weight
		}
		category := bucketForTerm(profile, term)
		weightedImportance := importance * profiles.BucketWeight(profile, category)
		scored := TermScore{
			Term:     term,
			Score:    weightedImportance,
			Category: category,
		}

		// Filter generic/noisy terms out of top/missing and coverage math.
		if isLowSignalTerm(term, category, termIDF, profile) {
			lowSignalTerms = append(lowSignalTerms, scored)
			continue
		}

		totalWeight += weightedImportance
		bucketTotals[category] += weightedImportance

		if tf > 0 {
			overlapTerms = append(overlapTerms, term)
			matchedWeight += weightedImportance
			bucketMatched[category] += weightedImportance
		} else {
			missingTerms = append(missingTerms, scored)
		}

		topTerms = append(topTerms, scored)
		buckets[category] = append(buckets[category], scored)
	}

	sortStrings(overlapTerms)
	sortTermScores(topTerms)
	sortTermScores(missingTerms)
	sortTermScores(lowSignalTerms)
	for category, terms := range buckets {
		sortTermScores(terms)
		if len(terms) > defaultTopN {
			terms = terms[:defaultTopN]
		}
		buckets[category] = terms
	}

	if len(topTerms) > defaultTopN {
		topTerms = topTerms[:defaultTopN]
	}

	rawCoverage := 0.0
	coverageScore := 0.0
	if totalWeight > 0 {
		rawCoverage = matchedWeight / totalWeight
		coverageScore = NormalizeCoverageScore(rawCoverage)
	}
	categoryCoverage := make(map[string]float64, len(bucketTotals))
	for bucket, total := range bucketTotals {
		if total <= 0 {
			continue
		}
		categoryCoverage[bucket] = bucketMatched[bucket] / total
	}

	return Signals{
		TopJobTerms:      topTerms,
		MissingJobTerms:  missingTerms,
		OverlapTerms:     overlapTerms,
		BucketedTopTerms: buckets,
		LowSignalTerms:   lowSignalTerms,
		CategoryCoverage: categoryCoverage,
		Discipline:       string(profile.Discipline),
		ProfileVersion:   profile.Version,
		Score:            coverageScore,
		RawCoverage:      rawCoverage,
		BM25Score:        bm25Score,
	}, nil
}

const (
	categoryLanguages   = "languages"
	categoryCloudDevOps = "cloud_devops_db"
	categoryPractices   = "practices"
	categorySoftSkills  = "soft_skills"
	categoryOther       = "other"
)

var (
	languageTerms = map[string]struct{}{
		"python": {}, "java": {}, "javascript": {}, "typescript": {}, "golang": {}, "csharp": {},
		"cpp": {}, "ruby": {}, "php": {}, "rust": {}, "kotlin": {}, "swift": {}, "scala": {},
		"sql": {}, "bash": {}, "powershell": {}, "r": {}, "perl": {}, "matlab": {},
	}
	cloudDevOpsDBTerms = map[string]struct{}{
		"aws": {}, "azure": {}, "gcp": {}, "cloud": {}, "devops": {}, "docker": {}, "kubernetes": {},
		"terraform": {}, "ansible": {}, "jenkins": {}, "helm": {}, "linux": {}, "nginx": {},
		"postgresql": {}, "mysql": {}, "mongodb": {}, "redis": {}, "dynamodb": {}, "snowflake": {},
		"bigquery": {}, "databricks": {}, "kafka": {}, "rabbitmq": {}, "prometheus": {}, "grafana": {},
		"nodejs": {}, "nextjs": {},
	}
	practiceTerms = map[string]struct{}{
		"agile": {}, "scrum": {}, "kanban": {}, "tdd": {}, "ddd": {}, "sre": {}, "ci": {},
		"cd": {}, "cicd": {}, "microservice": {}, "api": {}, "rest": {}, "graphql": {},
		"testing": {}, "test": {}, "automation": {}, "architecture": {}, "observability": {},
		"monitoring": {}, "reliability": {}, "security": {}, "performance": {},
	}
	softSkillTerms = map[string]struct{}{
		"communication": {}, "leadership": {}, "mentoring": {}, "collaboration": {}, "stakeholder": {},
		"ownership": {}, "initiative": {}, "teamwork": {}, "presentation": {}, "adaptability": {},
		"problem": {}, "problemsolving": {}, "problem-solving": {}, "coaching": {},
	}
	lowSignalTerms = map[string]struct{}{
		"experience": {}, "years": {}, "year": {}, "preferred": {}, "required": {}, "requirements": {},
		"ability": {}, "strong": {}, "excellent": {}, "good": {}, "role": {}, "position": {},
		"responsibility": {}, "responsibilities": {}, "candidate": {}, "skills": {}, "skill": {},
		"knowledge": {}, "understanding": {}, "familiarity": {}, "using": {}, "plus": {}, "must": {},
		"nice": {}, "seeking": {}, "join": {}, "company": {}, "business": {}, "customer": {},
		"customers": {}, "team": {}, "teams": {}, "support": {}, "supporting": {}, "work": {},
		"working": {}, "develop": {}, "development": {}, "design": {}, "implement": {},
		"implementation": {}, "maintain": {}, "maintenance": {}, "build": {}, "building": {},
		"solutions": {}, "solution": {}, "environments": {}, "environment": {},
	}
	reSQLFamily = regexp.MustCompile(`.+sql$`)
)

func classifyTerm(term string) string {
	if _, ok := languageTerms[term]; ok {
		return categoryLanguages
	}
	if _, ok := cloudDevOpsDBTerms[term]; ok || reSQLFamily.MatchString(term) {
		return categoryCloudDevOps
	}
	if _, ok := practiceTerms[term]; ok {
		return categoryPractices
	}
	if _, ok := softSkillTerms[term]; ok {
		return categorySoftSkills
	}
	return categoryOther
}

// extractPhraseCounts scans normalizedText for all known phrases and returns
// their occurrence counts. normalizedText must already be lowercased/accent-stripped.
func extractPhraseCounts(normalizedText string) map[string]int {
	counts := make(map[string]int, len(phraseIDF))
	for phrase := range phraseIDF {
		if n := strings.Count(normalizedText, phrase); n > 0 {
			counts[phrase] = n
		}
	}
	return counts
}

func bucketForTerm(profile profiles.Profile, term string) string {
	if strings.Contains(term, " ") {
		if cat, ok := phraseCategories[term]; ok {
			return cat
		}
		return categoryOther
	}
	bucket := profiles.BucketForTerm(profile, term)
	if bucket != "" && bucket != categoryOther {
		return bucket
	}
	if profile.Discipline == profiles.DisciplineITSoftware {
		return classifyTerm(term)
	}
	return categoryOther
}

func isLowSignalTerm(term, category string, idf float64, profile profiles.Profile) bool {
	if profiles.IsLowSignal(profile, term) {
		return true
	}
	if isDigitsOnly(term) {
		return true
	}
	if _, ok := lowSignalTerms[term]; ok {
		return true
	}
	// Curated phrases are always surfaced regardless of category or IDF.
	if _, ok := phraseIDF[term]; ok {
		return false
	}
	// "Other" terms are shown only if they look highly distinctive.
	if category == categoryOther && idf < highIDFOtherThreshold {
		if len(term) <= 3 {
			return true
		}
		if strings.HasSuffix(term, "ing") || strings.HasSuffix(term, "tion") || strings.HasSuffix(term, "ment") {
			return true
		}
		if _, ok := lowSignalTerms[term]; ok {
			return true
		}
	}
	// Non-standard tokens are often parser noise.
	if strings.Contains(term, "_") {
		return true
	}
	return false
}

func isDigitsOnly(term string) bool {
	if term == "" {
		return false
	}
	for _, r := range term {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
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
	return tokenizeWithProfile(text, profiles.Get(profiles.DefaultDiscipline()))
}

func tokenizeWithProfile(text string, profile profiles.Profile) []string {
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
		if isStopword(token) && !profile.StopwordOverrides[token] {
			b.Reset()
			return
		}
		// Normalize plural → singular only when token is not explicitly canonical.
		// This avoids breaking tools like "solidworks" / "ansys".
		if _, isCanonical := profile.CanonicalTerms[token]; !isCanonical {
			token = depluralize(token)
		}
		if isStopword(token) && !profile.StopwordOverrides[token] {
			b.Reset()
			return
		}
		tokens = append(tokens, canonicalizeWithProfile(token, profile))
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

func canonicalizeWithProfile(token string, profile profiles.Profile) string {
	token = profiles.Canonicalize(profile, token)
	token = canonicalize(token)
	token = profiles.Canonicalize(profile, token)
	return token
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
