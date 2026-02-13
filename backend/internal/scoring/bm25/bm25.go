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
	Score           float64     `json:"score"`
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
	overallScore := 0.0
	if len(scores) > 0 {
		overallScore = scores[0]
	}

	jobFreq := termFreq(jobTokens)
	resumeFreq := termFreq(resumeTokens)

	overlapTerms := make([]string, 0)
	missingTerms := make([]TermScore, 0)
	topTerms := make([]TermScore, 0, len(jobFreq))

	avgDocLen := float64(len(resumeTokens))
	docLen := float64(len(resumeTokens))

	for term, qtf := range jobFreq {
		tf := resumeFreq[term]

		// Use static corpus IDF instead of single-document IDF.
		termIDF := lookupIDF(term)

		if tf > 0 {
			overlapTerms = append(overlapTerms, term)
		} else {
			missingTerms = append(missingTerms, TermScore{
				Term:  term,
				Score: termIDF * float64(qtf),
			})
		}

		topTerms = append(topTerms, TermScore{
			Term:  term,
			Score: bm25TermScore(tf, docLen, avgDocLen, termIDF),
		})
	}

	sortStrings(overlapTerms)
	sortTermScores(topTerms)
	sortTermScores(missingTerms)

	if len(topTerms) > defaultTopN {
		topTerms = topTerms[:defaultTopN]
	}

	return Signals{
		TopJobTerms:     topTerms,
		MissingJobTerms: missingTerms,
		OverlapTerms:    overlapTerms,
		Score:           overallScore,
	}, nil
}

// stripAccents removes diacritics/accents from text using Unicode NFD
// decomposition (e.g. "développement" → "developpement", "résumé" → "resume").
// This allows French and other accented text to match English IDF table terms.
func stripAccents(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range norm.NFD.String(s) {
		// After NFD decomposition, accents become separate combining marks.
		// Keep only non-combining characters (base letters/digits).
		if unicode.Is(unicode.Mn, r) { // Mn = Mark, Nonspacing (combining accents)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func tokenize(text string) []string {
	if text == "" {
		return nil
	}

	// Normalize accented characters before tokenizing so that e.g. "développement"
	// and "developpement" produce the same token.
	text = stripAccents(text)

	var tokens []string
	var b strings.Builder

	flush := func() {
		if b.Len() < minTokenLength {
			b.Reset()
			return
		}
		token := b.String()
		if !isStopword(token) {
			tokens = append(tokens, token)
		}
		b.Reset()
	}

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
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

func bm25TermScore(tf int, docLen, avgDocLen, idfVal float64) float64 {
	if tf == 0 {
		return 0
	}
	numerator := float64(tf) * (defaultK1 + 1)
	denominator := float64(tf) + defaultK1*(1-defaultB+defaultB*(docLen/avgDocLen))
	return idfVal * (numerator / denominator)
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

// stopwords contains common English words plus job-posting boilerplate that add
// noise to BM25 signals without providing meaningful differentiation.
var stopwords = map[string]struct{}{
	// Standard English stopwords
	"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "be": {}, "been": {},
	"being": {}, "but": {}, "by": {}, "can": {}, "could": {}, "did": {}, "do": {},
	"does": {}, "doing": {}, "done": {}, "each": {}, "few": {}, "for": {}, "from": {},
	"get": {}, "got": {}, "had": {}, "has": {}, "have": {}, "having": {}, "he": {},
	"her": {}, "here": {}, "hers": {}, "him": {}, "his": {}, "how": {}, "if": {},
	"in": {}, "into": {}, "is": {}, "it": {}, "its": {}, "just": {}, "may": {},
	"me": {}, "might": {}, "more": {}, "most": {}, "must": {}, "my": {}, "no": {},
	"nor": {}, "not": {}, "now": {}, "of": {}, "on": {}, "only": {}, "or": {},
	"other": {}, "our": {}, "out": {}, "own": {}, "same": {}, "she": {}, "should": {},
	"so": {}, "some": {}, "such": {}, "than": {}, "that": {}, "the": {}, "their": {},
	"them": {}, "then": {}, "there": {}, "these": {}, "they": {}, "this": {}, "those": {},
	"through": {}, "to": {}, "too": {}, "under": {}, "up": {}, "us": {}, "very": {},
	"was": {}, "we": {}, "were": {}, "what": {}, "when": {}, "where": {}, "which": {},
	"while": {}, "who": {}, "whom": {}, "why": {}, "will": {}, "with": {}, "would": {},
	"you": {}, "your": {}, "yours": {}, "about": {}, "above": {}, "after": {}, "again": {},
	"against": {}, "all": {}, "also": {}, "am": {}, "any": {}, "because": {}, "before": {},
	"below": {}, "between": {}, "both": {}, "during": {}, "further": {}, "itself": {},
	"off": {}, "once": {}, "over": {}, "shall": {}, "until": {}, "upon": {},

	// Job-posting boilerplate — these appear in almost every listing
	// and don't help distinguish one role from another.
	"role": {}, "position": {}, "company": {}, "team": {}, "looking": {}, "seeking": {},
	"responsibilities": {}, "requirements": {}, "qualifications": {}, "preferred": {},
	"required": {}, "ability": {}, "including": {}, "within": {}, "across": {},
	"strong": {}, "excellent": {}, "proven": {}, "experience": {}, "work": {},
	"working": {}, "well": {}, "environment": {}, "opportunity": {}, "join": {},
	"ideal": {}, "candidate": {}, "applicant": {}, "apply": {}, "equal": {},
	"employer": {}, "benefits": {}, "salary": {}, "competitive": {},

	// ── French stopwords ──────────────────────────────────────────────
	// Common French function words that add noise to BM25 signals.
	// Note: accented forms are stripped by tokenizer, so "é"→"e", "à"→"a", etc.
	"le": {}, "la": {}, "les": {}, "un": {}, "une": {}, "des": {}, "du": {},
	"de": {}, "et": {}, "en": {}, "au": {}, "aux": {}, "ce": {}, "ces": {},
	"est": {}, "sont": {}, "ete": {}, "ont": {}, "sur": {}, "par": {},
	"pour": {}, "pas": {}, "que": {}, "qui": {}, "dans": {}, "avec": {},
	"tout": {}, "tous": {}, "toute": {}, "toutes": {}, "mais": {}, "ou": {},
	"ses": {}, "son": {}, "sa": {}, "nos": {}, "notre": {}, "vos": {}, "votre": {},
	"leur": {}, "leurs": {}, "ils": {}, "elles": {}, "nous": {}, "vous": {},
	"mon": {}, "ma": {}, "mes": {}, "ton": {}, "ta": {}, "tes": {},
	"ne": {}, "se": {}, "si": {}, "ya": {}, "ca": {}, "cet": {}, "cette": {},
	"ici": {}, "entre": {}, "comme": {}, "plus": {}, "moins": {}, "tres": {},
	"bien": {}, "aussi": {}, "meme": {}, "autre": {}, "autres": {},
	"peut": {}, "fait": {}, "faire": {}, "etre": {}, "avoir": {},
	"sera": {}, "seront": {}, "etait": {}, "etaient": {},
	"chez": {}, "dont": {}, "depuis": {}, "vers": {}, "sans": {},
	"alors": {}, "donc": {}, "encore": {}, "deja": {}, "apres": {},
	"avant": {}, "sous": {},

	// French job-posting boilerplate (accent-stripped)
	// Note: "candidate" and "experience" already defined in English section.
	"poste": {}, "entreprise": {}, "equipe": {}, "recherche": {}, "recherchons": {},
	"responsabilites": {}, "competences": {}, "requises": {}, "souhaitees": {},
	"profil": {}, "candidat": {}, "postuler": {},
	"salaire": {}, "avantages": {}, "environnement": {},
	"annees": {}, "niveau": {},
}

func isStopword(token string) bool {
	_, ok := stopwords[token]
	return ok
}
