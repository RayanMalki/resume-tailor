# Scoring Algorithm

Resume Tailor uses a profile-aware BM25 variant to score how well a resume matches a job description. This score serves as an ATS (Applicant Tracking System) compatibility proxy.

Source: `backend/internal/scoring/bm25/`

---

## Why BM25?

BM25 measures **keyword overlap** between two documents — the resume and the job description. It is a strong proxy for how ATS systems filter resumes: most real-world ATS tools rank candidates by term frequency against the job posting, not by semantic similarity.

The implementation uses:
- A **static IDF table** built from ~50,000 job descriptions (not per-document IDF)
- **Discipline profiles** that assign per-category bucket weights
- **Phrase detection** for multi-word tech terms
- A **sqrt normalization** of the raw coverage ratio to produce an intuitive 0–100% score

---

## Pipeline

```
Input text (resume + job description)
        │
        ▼
1. normalizeForTokenization
   ├─ lowercase + stripAccents (NFD decomposition)
   └─ techTokenReplacer: "C#"→"csharp", "ci/cd"→"cicd", "C++"→"cpp", etc.
        │
        ▼
2. tokenizeWithProfile
   ├─ split on non-letter/non-digit characters
   ├─ drop tokens shorter than 2 characters
   ├─ filter stopwords (English + French ISO lists + job-posting boilerplate)
   ├─ depluralize: strip trailing "-s" (4+ char tokens, skips known synonyms)
   └─ canonicalize: apply profile synonyms then global synonyms
        │
        ▼
3. extractPhraseCounts
   └─ scan normalized text for entries in phraseIDF dictionary
      (dictionary-based, not bigrams — e.g. "machine learning", "ci cd")
        │
        ▼
4. Per-term scoring loop (over job terms)
   ├─ lookupIDF(term) → corpus IDF value
   ├─ importance = IDF × jobFreq × profileWeight (if in CanonicalTerms)
   ├─ bucketForTerm → assign to discipline category
   ├─ weightedImportance = importance × BucketWeight(profile, category)
   ├─ isLowSignalTerm check → skip generic/noisy terms from coverage math
   └─ accumulate: totalWeight, matchedWeight, per-bucket totals
        │
        ▼
5. Score computation
   ├─ rawCoverage = matchedWeight / totalWeight
   └─ score = sqrt(rawCoverage)   ← NormalizeCoverageScore
```

---

## Tokenization Details

### Accent Stripping
Unicode NFD decomposition removes combining accents:
- `"développement"` → `"developpement"`
- `"résumé"` → `"resume"`

This allows French-language resumes to match the English IDF table.

PDF extractors sometimes emit modifier letter accents (U+02B0–U+02FF) instead of proper combining marks — these are stripped as well.

### Tech Punctuation Normalization
Applied before splitting so multi-character tech tokens survive:

| Input | Normalized |
|-------|-----------|
| `C#` | `csharp` |
| `C++` | `cpp` |
| `F#` | `fsharp` |
| `.NET` / `ASP.NET` | `dotnet` / `aspnet` |
| `Next.js` | `nextjs` |
| `Node.js` | `nodejs` |
| `ci/cd` | `cicd` |

### Depluralization
Strips trailing `-s` from tokens of 4+ characters. Exceptions:
- Tokens that end in `-ss` (e.g. `process`, `class`) — not stripped
- Tokens in the synonym table (e.g. `jenkins`, `redis`, `pandas`) — preserved as-is

### Synonym Canonicalization
Applied after depluralization. Examples:

| Input | Canonical |
|-------|----------|
| `go`, `golang` | `golang` |
| `js`, `javascript` | `javascript` |
| `ts`, `typescript` | `typescript` |
| `postgres`, `postgresql` | `postgresql` |
| `k8s`, `kubernetes` | `kubernetes` |
| `reactjs`, `react` | `react` |
| `node`, `nodejs` | `nodejs` |
| `rest`, `restful` | `rest` |
| `microservices` | `microservice` |
| `apis` | `api` |
| `infonuagique` | `cloud` |

Profile-level synonyms are applied first (discipline-specific aliases), then global synonyms.

### Stopword Filtering
A comprehensive list (English ISO 639 + French ISO 639 + job-posting boilerplate) is applied before and after depluralization. Profiles can mark terms as `StopwordOverrides` to keep them (e.g. discipline-specific short acronyms).

---

## IDF Table

Defined in `idf_table.go`. Key constants:

| Constant | Value | Meaning |
|----------|-------|---------|
| `defaultCorpusIDF` | 5.0 | Unknown terms — treated as moderately rare |
| `shortUnknownIDF` | 4.2 | Short unknown acronyms |
| `noisyUnknownIDF` | 2.5 | Likely OCR/typo garbage |

IDF formula: `log((N - df + 0.5) / (df + 0.5) + 1)` where N = 50,000.

Terms appearing in almost every document receive low IDF (~0.0). Rare specialized terms receive high IDF (~6–10). The table covers 500+ common technical and professional vocabulary terms across software, data science, PM, design, and business roles.

---

## Phrase Detection

Multi-word phrases (e.g. `"machine learning"`, `"natural language processing"`) are detected via a dictionary lookup against `phraseIDF` — a separate map of known multi-word terms and their IDF values.

Implementation (`extractPhraseCounts`):
- Scans the normalized (lowercased + accent-stripped) text
- Uses `strings.Count` for each known phrase in the dictionary
- Returns occurrence counts merged into the term frequency map

This is **dictionary-based**, not n-gram enumeration. Only pre-defined phrases are recognized. Phrases have their own IDF values and are never suppressed by the low-signal filter.

Phrase categories are tracked in `phraseCategories` — a parallel map from phrase to its bucket label.

---

## Discipline Profiles

Five discipline profiles are defined in `backend/internal/scoring/profiles/data/profiles.go`:

| Discipline | Use case |
|-----------|----------|
| `it_software` | Software engineering, data, DevOps |
| `mechanical` | Mechanical engineering |
| `electrical` | Electrical / electronics engineering |
| `aerospace` | Aerospace and defense |
| `industrial_logistics` | Manufacturing, supply chain, operations |

Each profile contains:

| Field | Purpose |
|-------|---------|
| `CanonicalTerms` | term → `{bucket, weight}` map for discipline-specific terms |
| `Synonyms` | discipline-specific aliases (e.g. `plc` → `programmable logic controller`) |
| `StopwordOverrides` | terms to preserve that would otherwise be filtered |
| `LowSignalTerms` | noisy terms suppressed from top/missing and coverage calculation |
| `Buckets` | named category labels for reporting |
| `BucketWeights` | per-bucket multipliers applied to term importance |
| `PromptHints` | context strings injected into AI prompts |

### Bucket Assignment (`bucketForTerm`)

1. Multi-word phrases → look up in `phraseCategories`; otherwise `other`
2. Single tokens → check `profiles.BucketForTerm` (profile's `CanonicalTerms`)
3. For `it_software` profile only: fallback to `classifyTerm` using hardcoded category maps (languages, cloud_devops_db, practices, soft_skills)
4. Default: `other`

### Coverage Calculation

For each bucket:
```
categoryCoverage[bucket] = bucketMatched[bucket] / bucketTotals[bucket]
```

Overall:
```
rawCoverage = matchedWeight / totalWeight
score       = sqrt(rawCoverage)    // lifts mid-range values; 0→0, 1→1
```

---

## Low-Signal Filter

Terms are excluded from `TopJobTerms`, `MissingJobTerms`, and coverage math if:
- They appear in the profile's `LowSignalTerms`
- They are digit-only strings
- They appear in the global `lowSignalTerms` map (e.g. `experience`, `required`, `skills`, `team`, `work`)
- They are in the `other` bucket with IDF < 6.2 AND are ≤3 chars, or end in `-ing`/`-tion`/`-ment`
- They contain underscores (parser noise)

Curated phrases in `phraseIDF` bypass this filter entirely.

---

## Output Signals

The `Signals` struct returned by `ComputeWithProfile`:

| Field | Type | Description |
|-------|------|-------------|
| `Score` | float64 | Normalized coverage score in [0,1] (sqrt of raw coverage) |
| `RawCoverage` | float64 | Raw `matchedWeight / totalWeight` ratio |
| `BM25Score` | float64 | Raw 2-document BM25 score (diagnostic only, not used as ATS score) |
| `TopJobTerms` | `[]TermScore` | Top 10 highest-importance job terms (whether or not matched) |
| `MissingJobTerms` | `[]TermScore` | Important job terms absent from resume, sorted by importance |
| `OverlapTerms` | `[]string` | All terms appearing in both resume and job |
| `BucketedTopTerms` | `map[string][]TermScore` | Top 10 terms per category bucket |
| `CategoryCoverage` | `map[string]float64` | Per-bucket coverage ratio |
| `LowSignalTerms` | `[]TermScore` | Filtered-out terms (diagnostic) |
| `Discipline` | string | Active discipline name |
| `DisciplineEvidence` | `[]TermScore` | Terms that drove discipline detection |
| `ProfileVersion` | string | Version string of the profile dictionary used |

---

## Discipline Detection

The `classifier/` package scores each resume against all five discipline profiles and returns the one with the highest weighted term overlap. A confidence threshold must be met for the result to be considered reliable.

The detected discipline is used to:
1. Select the discipline profile for BM25 scoring
2. Choose bucket weights for coverage calculation
3. Inject `PromptHints` into the OpenAI system prompt
