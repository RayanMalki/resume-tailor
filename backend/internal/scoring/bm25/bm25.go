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

// synonyms maps technology aliases to a canonical form so BM25 treats them as the same token.
// Both directions must be defined (e.g. "go"→"golang" AND "golang"→"go") — the canonical
// form is the FIRST entry so both occurrences end up as the same token.
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
	"c#":     "csharp",
	// Continuous Integration / Continuous Deployment
	"ci":   "cicd",
	"cd":   "cicd",
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
	"aws":    "aws",
	"amazon": "aws",
	// Google Cloud Platform
	"gcp": "gcp",
	// Infonuagique (French for cloud computing)
	"infonuagique": "cloud",
	"cloud":        "cloud",
	// Agile / Scrum
	"scrum": "agile",
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
	"jenkins":  "jenkins",
	"redis":    "redis",
	"travis":   "travis",
	"atlas":    "atlas",
	"pandas":   "pandas",
	"keras":    "keras",
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

	// ── Additional English generic words common in job postings ────────
	"knowledge": {}, "understanding": {}, "familiar": {}, "familiarity": {},
	"support": {}, "supporting": {}, "supported": {},
	"allow": {}, "allowing": {}, "allowed": {},
	"another": {}, "area": {}, "areas": {},
	"asset": {}, "assets": {},
	"available": {}, "availability": {},
	"based": {}, "basis": {},
	"best": {}, "better": {},
	"build": {}, "building": {},
	"business": {}, "career": {},
	"challenge": {}, "challenges": {}, "challenging": {},
	"change": {}, "changes": {},
	"complete": {}, "completing": {}, "completion": {},
	"create": {}, "creating": {},
	"current": {}, "currently": {},
	"day": {}, "days": {},
	"deliver": {}, "delivering": {},
	"description": {},
	"different": {}, "diverse": {}, "diversity": {},
	"effort": {}, "efforts": {},
	"ensure": {}, "ensuring": {},
	"every": {}, "everyone": {},
	"first": {}, "follow": {}, "following": {},
	"full": {}, "fully": {},
	"given": {}, "great": {},
	"grow": {}, "growing": {},
	"help": {}, "helping": {},
	"high": {}, "highly": {},
	"include": {}, "included": {}, "includes": {},
	"key": {}, "keep": {},
	"know": {}, "known": {},
	"large": {}, "learn": {}, "learning": {},
	"like": {}, "long": {},
	"make": {}, "making": {},
	"manage": {}, "managing": {},
	"many": {}, "much": {},
	"need": {}, "needed": {}, "needs": {},
	"new": {}, "next": {},
	"offer": {}, "offering": {}, "offers": {},
	"open": {}, "order": {},
	"part": {}, "people": {}, "person": {},
	"place": {}, "please": {},
	"provide": {}, "providing": {}, "provided": {},
	"range": {},
	"related": {}, "relevant": {},
	"responsible": {}, "responsibility": {},
	"right": {},
	"set": {}, "several": {},
	"share": {}, "sharing": {},
	"show": {}, "significant": {},
	"similar": {}, "since": {},
	"skill": {}, "skills": {},
	"start": {}, "starting": {},
	"success": {}, "successful": {}, "successfully": {},
	"take": {}, "taking": {},
	"think": {}, "thinking": {},
	"time": {}, "today": {},
	"together": {},
	"top": {}, "toward": {}, "towards": {},
	"true": {}, "turn": {},
	"understand": {},
	"use": {}, "used": {}, "using": {}, "utilize": {},
	"value": {}, "values": {},
	"want": {}, "way": {}, "ways": {},
	"world": {},
	"able": {}, "along": {}, "always": {}, "become": {},
	// "being" already in standard English stopwords
	"bring": {}, "brought": {},
	"come": {}, "comes": {},
	"consider": {}, "continue": {},
	"directly": {},
	"even": {}, "expect": {}, "expected": {},
	"find": {},
	"good": {},
	"important": {}, "improve": {},
	"information": {},
	"involve": {}, "involved": {},
	"look": {},
	// "looking" already in job-posting boilerplate
	"move": {}, "moving": {},
	"often": {},
	"plan": {}, "planning": {},
	"play": {}, "possible": {},
	"process": {},
	"project": {}, "projects": {},
	"put": {},
	"really": {},
	"report": {}, "reporting": {},
	"run": {}, "running": {},
	"see": {},
	"serve": {}, "serving": {},
	"specific": {},
	"still": {},
	// "such" already in standard English stopwords
	"thing": {}, "things": {},
	"already": {},
	"focus": {}, "focused": {},
	"level": {},
	"enable": {}, "enabling": {},
	"act": {}, "acting": {},
	"positive": {},
	"various": {},
	"thrive": {},
	"respect": {},
	"inspire": {}, "inspiring": {},
	"confirm": {},
	"via": {},
	"maximum": {},

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
	"tu": {}, "te": {}, "toi": {},
	"ne": {}, "se": {}, "si": {}, "ya": {}, "ca": {}, "cet": {}, "cette": {},
	"ici": {}, "entre": {}, "comme": {}, "plus": {}, "moins": {}, "tres": {},
	"bien": {}, "aussi": {}, "meme": {}, "autre": {}, "autres": {},
	"peut": {}, "fait": {}, "faire": {}, "etre": {}, "avoir": {},
	"sera": {}, "seront": {}, "etait": {}, "etaient": {},
	"chez": {}, "dont": {}, "depuis": {}, "vers": {}, "sans": {},
	"alors": {}, "donc": {}, "encore": {}, "deja": {}, "apres": {},
	"avant": {}, "sous": {}, "sommes": {}, "etes": {}, "suis": {},
	"quand": {}, "lorsque": {}, "pendant": {}, "chaque": {},
	"peu": {}, "beaucoup": {}, "trop": {}, "assez": {},
	"ans": {}, "mois": {}, "jour": {}, "jours": {},

	// French job-posting boilerplate (accent-stripped)
	// Note: "candidate" and "experience" already defined in English section.
	"poste": {}, "entreprise": {}, "equipe": {}, "recherche": {}, "recherchons": {},
	"responsabilites": {}, "competences": {}, "requises": {}, "souhaitees": {},
	"profil": {}, "candidat": {}, "postuler": {},
	"salaire": {}, "avantages": {}, "environnement": {},
	"annees": {}, "niveau": {},
	"recrutement": {}, "recruter": {}, "agence": {},
	"talents": {}, "talent": {}, "collegues": {}, "collegue": {},
	"diversite": {}, "inclusion": {}, "inclusif": {}, "inclusive": {},
	"ensemble": {}, "participer": {}, "participation": {},
	"anglais": {}, "francais": {}, "bilingue": {},
	"offre": {}, "offrons": {}, "proposons": {},
	"rejoindre": {}, "rejoignez": {},
	"passionnee": {}, "passionne": {},
	"dynamique": {}, "motivee": {}, "motive": {},
	"carriere": {}, "emploi": {}, "stage": {}, "stagiaire": {},
	"connaissances": {}, "connaissance": {},
	"apprentissage": {}, "apprendre": {},
	"contenu": {}, "contenus": {},
	"personnalises": {}, "personnalise": {},
	"favoriser": {}, "enrichir": {},
	// "continue" already in English section
	"disponibles": {}, "disponible": {},
	"possedant": {}, "differentes": {},
	// "different" already in English section
	"expertises": {}, "expertise": {}, "experiences": {},
	"profils": {}, "diversifies": {},
	"points": {}, "vue": {},
	"titre": {}, "positif": {}, "organisation": {},
	"grace": {}, "permettent": {}, "permet": {},
	"maitriser": {}, "metier": {},
	"mode": {}, "bases": {},
	"atout": {}, "comprehension": {},
	"prerequis": {}, "prealables": {},
	"secteur": {}, "activite": {},
	"curiosite": {}, "fort": {}, "esprit": {},
	"rigueur": {}, "travail": {},
	"completement": {},
	// "completion" already in English section
	"connexe": {}, "etudes": {},
	"relever": {}, "defis": {},
	"supporter": {}, "croissance": {},
	"confirmer": {}, "livrables": {},
	// "via" already in English section
	"integrer": {}, "inspirante": {}, "respecte": {},
	"meilleures": {}, "pratiques": {},
	"innovantes": {}, "innovante": {},
	// "maximum" already in English section
	"valeur": {},
	"divers": {}, "partenaires": {}, "affaires": {},

	// ── French generic verbs/nouns that are not ATS-relevant ──────────
	"acces": {}, "acceder": {},
	"agir": {}, "action": {}, "actions": {},
	"assurant": {}, "assurer": {}, "assure": {},
	"creer": {}, "creation": {}, "cree": {},
	"cours": {}, // "en cours de"
	"banque": {}, "bancaire": {}, "nationale": {}, "national": {},
	"back": {}, "end": {}, // "back-end" splits into "back" + "end"
	"front": {}, // "front-end"
	// "ton", "tes", "toi" already in French function words above
	"developper": {}, "developpe": {}, "developpee": {},
	"deployer": {}, "deploye": {},
	"fonctions": {}, "fonction": {},
	"solutions": {}, "solution": {},
	"technologiques": {}, "technologique": {},
	"programmes": {}, "programme": {},
	"basees": {}, "basee": {},
	"impact": {}, "impactant": {},
	"baccalaureat": {},
	"diplome": {}, "diplomes": {},
	"type": {}, "types": {},
	// "role" already in English job-posting boilerplate above
	"developpeurs": {}, "developpeuse": {},
	"assurance": {},
	"resultats": {}, "resultat": {},
	"processus": {}, "procedure": {}, "procedures": {},
	"gerer": {}, "gerant": {},
	"tant": {}, // "en tant que"
	"vient": {}, "venir": {},
	"supportent": {}, "supporte": {},
	// "supporter" already defined above
	// "description" already in English section
	"concernant": {}, "concerne": {},
	"permettre": {}, "permettant": {},
	"doit": {}, "doivent": {},
	"presente": {}, "presenter": {},
	// "assurer", "assurant" already defined above
	"souhaite": {}, "souhaiter": {},
	"capable": {}, "capacite": {},
	"necessaire": {}, "necessaires": {},
	// "important" already in English section
	"importante": {},
	"essentiels": {}, "essentiel": {}, "essentielle": {},
	"specifique": {}, "specifiques": {},
	"pertinent": {}, "pertinente": {}, "pertinents": {},
	"contribuer": {},
	"repondre": {},
	"ameliorer": {},

	// ── City names (not useful for ATS keyword matching) ──────────────
	"montreal": {}, "toronto": {}, "vancouver": {}, "ottawa": {}, "quebec": {},
	"paris": {}, "lyon": {}, "marseille": {}, "toulouse": {}, "bordeaux": {},
	// "new" already in English section
	"york": {}, "san": {}, "francisco": {}, "london": {},
	"berlin": {}, "remote": {},
}

func isStopword(token string) bool {
	_, ok := stopwords[token]
	return ok
}
