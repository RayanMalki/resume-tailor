package bm25

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"

	libbm25 "github.com/crawlab-team/bm25"
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
func Compute(resumeText, jobText string) (Signals, error) {
	resumeTokens := tokenize(resumeText)
	jobTokens := tokenize(jobText)

	if len(resumeTokens) == 0 || len(jobTokens) == 0 {
		return Signals{}, nil
	}

	bm25Instance, err := libbm25.NewBM25Okapi([]string{resumeText}, tokenize, defaultK1, defaultB, nil)
	if err != nil {
		return Signals{}, fmt.Errorf("bm25 init: %w", err)
	}

	scores, err := bm25Instance.GetScores(jobTokens)
	if err != nil {
		return Signals{}, fmt.Errorf("bm25 score: %w", err)
	}

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
		if tf > 0 {
			overlapTerms = append(overlapTerms, term)
		} else {
			missingTerms = append(missingTerms, TermScore{
				Term:  term,
				Score: idf(1, 0) * float64(qtf),
			})
		}

		topTerms = append(topTerms, TermScore{
			Term:  term,
			Score: bm25TermScore(tf, docLen, avgDocLen, idf(1, boolToDF(tf > 0))),
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

func tokenize(text string) []string {
	if text == "" {
		return nil
	}

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

func idf(totalDocs, docFreq int) float64 {
	n := float64(totalDocs)
	df := float64(docFreq)
	return math.Log((n-df+0.5)/(df+0.5) + 1.0)
}

func boolToDF(hasTerm bool) int {
	if hasTerm {
		return 1
	}
	return 0
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

var stopwords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "be": {}, "by": {},
	"for": {}, "from": {}, "in": {}, "is": {}, "of": {}, "on": {}, "or": {}, "that": {},
	"the": {}, "this": {}, "to": {}, "with": {},
}

func isStopword(token string) bool {
	_, ok := stopwords[token]
	return ok
}
