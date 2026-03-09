# User Journey

This document walks through the end-to-end experience of using Resume Tailor, from landing page to downloading tailored application assets.

---

## Overview

```
Landing → Sign Up/Login → Upload Resume → Describe Job → Processing → Results → Export
```

---

## Step 1 — Landing Page (`/`)

The root page introduces the product. Users can:
- Start the sign-up flow
- Log in with email/password or Google OAuth
- Read about what the tool generates

**Feature callout:** Users without an OpenAI API key can still use the service (subject to free-tier rate limits). Power users can bring their own key (BYOK) for unlimited usage.

---

## Step 2 — Sign Up / Login

**Email/password flow:**
1. `POST /v1/auth/signup` — creates account and sends a verification email
2. `POST /v1/auth/login` — authenticates and sets an HttpOnly session cookie
3. Optional: `GET /v1/auth/verify?token=...` — verifies email address

**Google OAuth flow:**
1. `GET /v1/auth/google/start` — redirects to Google's consent screen
2. `GET /v1/auth/google/callback` — exchanges code, creates/finds user, sets session cookie

**Password recovery:**
- `POST /v1/auth/forgot-password` → email with reset link
- `POST /v1/auth/reset-password` — sets new password via token

After login, the user lands on `/dashboard`.

---

## Step 3 — Dashboard (`/dashboard`)

The dashboard shows:
- A button to start a new run
- A history table of past runs with status and ATS score
- An ATS score trend chart across runs

**Feature callout:** Each row links to the full result for that run.

---

## Step 4 — Upload Resume (`/resume`)

Users provide their base resume in one of two ways:

| Method | Endpoint | Notes |
|--------|----------|-------|
| Paste text | `POST /v1/resumes` | Plain text body |
| Upload file | `POST /v1/resumes/upload` | `.pdf` or `.docx`, max 2 MB |

The server extracts text from uploaded files server-side. Saved resumes are reusable across future runs.

---

## Step 5 — Describe the Job (`/job`)

Users enter:
- **Job description** — the full text of the role they're applying for
- **Discipline** — optionally pre-selected; auto-detected if left blank

The discipline selector picks from: `it_software`, `mechanical`, `electrical`, `aerospace`, `industrial_logistics`.

**Feature callout:** Users can trigger discipline auto-detection via `POST /v1/disciplines/detect` before creating the run. The detected discipline affects which BM25 bucket weights apply.

Optional fields during run creation:
- Project controls: `pinned`, `auto`, or `exclude` per project in the resume
- Job title (used in cover letter generation)
- Company name (used in cover letter generation)

Run is created via `POST /v1/runs`.

---

## Step 6 — Processing (async)

After run creation:
1. The API inserts a row into the `jobs` table (SQL-backed queue)
2. The API returns `{ run_id, status: "pending" }` immediately
3. The frontend polls `GET /v1/runs/{runID}` until `status` changes from `pending` → `processing` → `done` (or `error`)
4. The worker picks up the job, runs BM25 scoring, calls OpenAI, compiles LaTeX/DOCX artifacts
5. Results are stored in `run_reports` and `run_artifacts_items`

Typical processing time: 20–60 seconds depending on OpenAI latency.

---

## Step 7 — Results (`/result/{runId}`)

The results page has several tabs:

### ATS Report Tab
- **ATS Score** (0–100%) — normalized BM25 coverage score
- **Score breakdown by category** — e.g. languages, cloud/devops, practices
- **Missing keywords** — high-IDF terms in the job not found in the resume
- **AI-generated summary** — narrative assessment of fit
- **Interview questions** — tailored to the job description

### BM25 Analysis Tab
- **Top job terms** — highest-scoring terms from the job description
- **Overlap terms** — terms appearing in both resume and job
- **Bucketed term breakdown** — terms grouped by discipline category

### Resume Preview Tab
- Rendered view of the tailored resume
- Incorporates keyword improvements from the BM25 diff

### Cover Letter Tab
- AI-generated cover letter text personalized to the job

---

## Step 8 — Export

From the results page, users can download:

| Artifact | Endpoint | Format |
|----------|----------|--------|
| Resume (LaTeX source) | `GET /v1/runs/{runID}/artifacts/resume-latex` | `.tex` |
| Resume (PDF) | `GET /v1/runs/{runID}/artifacts/resume-pdf` | `.pdf` (if Tectonic enabled) |
| Resume (Word) | `GET /v1/runs/{runID}/artifacts/resume-docx` | `.docx` |
| Cover letter (text) | `GET /v1/runs/{runID}/artifacts/cover-letter` | `.txt` |
| Cover letter (PDF) | `GET /v1/runs/{runID}/artifacts/cover-letter-pdf` | `.pdf` |
| Cover letter (Word) | `GET /v1/runs/{runID}/artifacts/cover-letter-docx` | `.docx` |
| Project reasons | `GET /v1/runs/{runID}/artifacts/project-reasons` | JSON |

---

## Settings (`/settings`)

Users can manage:
- Password change (`PUT /v1/me/password`)
- OpenAI API key (BYOK) — stored AES-256-GCM encrypted (`PUT /v1/me/api-key`)
- Account deletion (`DELETE /v1/me`)

---

## Rate Limits Summary

| Limit | Value |
|-------|-------|
| Global IP rate limit | 60 req/min |
| Login | 5/min per IP |
| Signup | 3/hour per IP |
| Resume upload | 10/min per IP |
| Run creation (burst) | 1/min per user |
| Run creation (daily) | 10/day per user |
| Free-tier daily runs (no BYOK) | 3/day per user, 6/day per creator IP |
