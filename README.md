# Resume Tailor (WIP)

Resume Tailor is a backend-first project that helps generate **ATS-style feedback** and a **prioritized change plan** for a resume based on a job posting. The goal is to provide actionable improvements while keeping the system **reliable, testable, and end-to-end**.

> **Status:** 🚧 Under active development — not deployed yet.

---

## What it does (current vision)
- Accepts **resume text** + **job description text**
- Computes relevance signals using a **custom BM25 implementation**
- Produces:
  - an **ATS Report** (gaps, strengths, keyword coverage, clarity issues)
  - a **Change Plan** (what to edit first and why)
- Stores results per **run**, so every analysis is traceable and retrievable

---

## What I’ve built so far
### Backend foundations
- **Run/report design:** a run lifecycle where outputs are saved in a `run_reports` store once processing completes.
- **Protected report endpoint:** `GET /v1/runs/{runID}/report` (requires authentication).
- **Auth plan (v1):** session-based authentication using an **HttpOnly cookie** (SameSite=Lax, Path=/, with expiration).
- **Backend structure:** separation between **handlers → services → repositories**, with clear boundaries to keep the code maintainable.

### Processing approach
- Defined the worker pipeline direction:
  - Extract signals (BM25 keywords/score)
  - Call an LLM to generate an ATS report + change plan
  - Persist results into run_reports / artifacts
  - Update run + job statuses end-to-end

---

## What’s still in progress (next steps)
### Core worker logic
- Implement the real `process_run` logic (replace placeholders):
  - Parse and normalize resume/job text
  - Compute **BM25** signals and keyword coverage
  - Generate structured outputs (ATSReport + ChangePlan) via LLM
  - Save outputs in `run_reports` and later `run_artifacts`
  - Mark runs/jobs as completed/failed with correct status transitions

### Auth + middleware completion
- Finish session repo/service logic:
  - `GetSessionByTokenHash`, `DeleteSessionByTokenHash`
  - `Authenticate`, `Logout`
- Add `AuthRequired` middleware and `GET /v1/me`
- Validate behavior with `curl` (401/200/logout + cookie jar)

### Deployment (planned)
- Pick deployment target and ship:
  - API + DB (and worker) with environment variables
  - Logging/monitoring basics
  - Production-safe config (CORS, cookie settings, secrets, rate limits)

---

## Roadmap
- ✅ Strong backend structure (handlers/services/repos, run report endpoint, run/report concept)
- 🔄 Worker `process_run` real logic (BM25 → LLM → persistence)
- 🔄 Complete auth middleware + `/v1/me` and finalize cookie-based sessions
- ⏭ Add `run_artifacts` (PDF/LaTeX later, handled in a separate phase)
- ⏭ Deploy (API + worker + DB) once v1 is stable

---

## Why this project
Resume Tailor is designed to be more than generic advice: it combines **transparent scoring signals (BM25)** with **LLM-generated guidance**, and persists results in a run-based system so it stays debuggable and production-oriented.
