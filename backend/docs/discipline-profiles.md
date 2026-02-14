# Discipline Profiles Governance

This document describes how ResumeTailor maintains discipline-aware dictionaries used for:

- discipline classification
- BM25 canonicalization and filtering
- bucketed coverage scoring
- AI prompt guidance

## Data location

Profiles are versioned in:

- `backend/internal/scoring/profiles/data/profiles.go`

The active profile version is exported as `Version` and attached to run reports.

## Supported disciplines

- `mechanical`
- `electrical`
- `industrial_logistics`
- `aerospace`
- `it_software`

## Profile structure

Each profile contains:

- `CanonicalTerms`: term -> `{bucket, weight}`
- `Synonyms`: deterministic synonym/canonical map
- `StopwordOverrides`: terms to keep even if globally treated as stopwords
- `LowSignalTerms`: terms filtered out from top/missing displays
- `Buckets`: named category buckets for reporting
- `BucketWeights`: weighting used in discipline-aware ATS coverage score
- `PromptHints`: discipline context injected into AI prompts

## Update process

1. Add/update terms and synonyms in the target profile.
2. Keep mappings deterministic (no fuzzy logic, no model-generated dictionaries).
3. Add tests for classifier behavior and BM25 filtering where relevant.
4. Bump `Version` in `profiles/data/profiles.go` when changing semantics.
5. Validate backward compatibility by running:
   - `go test ./...`
   - `go build ./cmd/api ./cmd/worker`
6. Check report payload contains updated `profile_version`.

## Authoring rules

- Prefer canonical singular forms for terms.
- Avoid generic words in `CanonicalTerms`; place them in `LowSignalTerms`.
- Keep `other` bucket as fallback only; high-IDF terms may still surface there.
- Never add terms that imply fabricated experience in AI generation prompts.

