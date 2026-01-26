# Resume Tailor (WIP)

Resume Tailor is a backend-first project that generates an **ATS-style report** and a **prioritized change plan** by comparing a resume to a job description. The focus is on building something **production-oriented**: authenticated access, run-based processing, and reproducible outputs.

> **Status:** 🚧 Under active development — not deployed yet.

---

## What it does (today)
- Accepts **resume text** + **job posting text**
- Creates a **run** and processes it with a **worker**
- Produces and stores a **run report** that can be fetched via a protected API endpoint

---

## What I’ve already built
### ✅ Authentication (done)
- Session-based auth using an **HttpOnly cookie** (SameSite + expiration)
- Auth routes (signup/login/logout)
- Middleware to protect private endpoints (e.g., fetching reports)

### ✅ Run pipeline + worker processing (done)
- A run lifecycle (created → processing → completed/failed)
- Worker processes runs end-to-end:
  - loads inputs
  - generates/stores report output
  - updates run status
- Persistent storage for run outputs (e.g., `run_reports`)
- Protected endpoint to retrieve a completed report:
  - `GET /v1/runs/{runID}/report`

---

## What’s next (remaining work)
### 🔄 BM25 algorithm (in progress / next)
- Implement a **custom BM25** scoring layer to produce transparent keyword/coverage signals
- Use BM25 results to drive better, more explainable recommendations (not just generic LLM advice)

### 🔄 LLM path (in progress / next)
- Integrate the LLM step into the worker (or refine it if it’s currently a placeholder):
  - generate structured outputs (ATS report + change plan)
  - ensure predictable formatting + schema validation
  - add tests and prompt/versioning so outputs are stable over time

### 🔄 Server-side LaTeX compilation (planned)
- Generate LaTeX from the change plan / improved resume
- Compile on the server to produce a downloadable artifact (PDF)
- Store generated artifacts per run (e.g., `run_artifacts`)

### 🔄 Frontend (planned)
- UI to:
  - submit resume + job text
  - track run status
  - view report + change plan
  - download compiled PDF artifact

---

## Deployment goal
Once BM25 + LLM + LaTeX compilation are stable, the plan is to deploy the system as:
- **API server**
- **worker**
- **database + storage** (for reports/artifacts)
with proper environment-based config, logging, and production-safe auth/cookie settings.

---

## Why this project
Resume Tailor is built to be **useful and explainable**: combining ranking signals (BM25) with LLM-generated guidance, wrapped in a real backend workflow (auth + runs + persistence) so it behaves like a product, not a script.
