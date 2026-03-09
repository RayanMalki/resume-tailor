# System Design

## Architecture Overview

Resume Tailor is a three-service application: a stateless Go API, an asynchronous Go worker, and a Next.js frontend. PostgreSQL serves as both the primary database and the job queue.

```
┌─────────────────────────────────────────────────────────────┐
│                         Browser                             │
└───────────────────────┬─────────────────────────────────────┘
                        │ HTTPS (cookie auth)
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                  Next.js Frontend                           │
│  (SSR + Client Components, /api/* proxied to Go API)        │
└───────────────────────┬─────────────────────────────────────┘
                        │ HTTP (proxied)
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                    Go API Server                            │
│  Auth · Resumes · Runs · Reports · Artifacts · Middleware   │
└───────────────┬────────────────────┬────────────────────────┘
                │                    │
          (reads/writes)       (inserts jobs)
                │                    │
                ▼                    ▼
┌──────────────────────────────────────────────────────┐
│                    PostgreSQL                         │
│  users · sessions · resumes · runs · jobs            │
│  run_reports · run_artifacts_items                   │
└──────────────────────────┬───────────────────────────┘
                           │ (polls for pending jobs)
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Go Worker                                │
│  BM25 Scoring · OpenAI LLM calls · LaTeX/DOCX rendering     │
└────────────────────────────────────────────────────────────┘
                           │
                           ▼ (external)
                    ┌──────────────┐
                    │   OpenAI API  │
                    └──────────────┘
```

---

## Service Responsibilities

### Go API (`cmd/api`)
- Handles all HTTP traffic: authentication, resume management, run creation, artifact downloads
- Validates requests, enforces rate limits, checks session cookies
- Writes run requests to the `jobs` table and returns immediately (non-blocking)
- Does **not** call OpenAI or perform heavy computation

### Go Worker (`cmd/worker`)
- Polls the `jobs` table for pending work using a SQL-backed queue with row-level locking
- Runs BM25 scoring against the discipline profile
- Calls OpenAI to generate resume improvements, ATS report, and cover letter
- Compiles LaTeX → PDF via Tectonic (if enabled)
- Converts to DOCX via internal renderer
- Stores results in `run_reports` and `run_artifacts_items`

**Why separate?** LLM calls can take 20–60 seconds. Keeping them out of the API process prevents request timeouts and allows independent scaling.

### Next.js Frontend (`web/`)
- App Router with server and client components
- Proxies all `/api/*` requests to the Go API (avoids cross-origin cookie issues)
- Polls `GET /v1/runs/{runID}` until run status is `done`
- Renders the multi-tab results page (ATS report, BM25 analysis, resume preview, cover letter)

---

## Database Schema

### Core Tables

| Table | Description |
|-------|-------------|
| `users` | Account records: email, hashed password, encrypted API key, flags |
| `sessions` | HttpOnly session tokens tied to a user |
| `resumes` | Saved resume records: raw text, file metadata |
| `runs` | One run per job application attempt: resume ID, job text, discipline, status |
| `jobs` | SQL-backed work queue: one row per pending/in-progress/done job |
| `run_reports` | BM25 signals + AI-generated ATS report JSON for a completed run |
| `run_artifacts_items` | Binary/text blobs per artifact type (resume-latex, resume-pdf, cover-letter, etc.) |

### Relationships

```
users ──< sessions
users ──< resumes
users ──< runs
runs  ──< jobs               (one active job per run)
runs  ──  run_reports        (one report per completed run)
runs  ──< run_artifacts_items (multiple artifacts per run)
```

---

## Request Lifecycle — Creating a Run

```
1. POST /v1/runs
   ├─ API validates session, checks rate limits
   ├─ Creates run row (status=pending)
   ├─ Inserts job row into jobs table
   └─ Returns { run_id, status: "pending" }

2. Frontend polls GET /v1/runs/{runID} every 2–3s

3. Worker loop (every few seconds):
   ├─ SELECT ... FOR UPDATE SKIP LOCKED on jobs table
   ├─ Claims job, sets status=processing
   ├─ Runs BM25 scoring with discipline profile
   ├─ Calls OpenAI for resume tailoring + ATS report + cover letter
   ├─ Compiles LaTeX → PDF (if RESUME_PDF_ENABLED=1)
   ├─ Renders DOCX artifacts
   ├─ Writes run_reports row
   ├─ Writes run_artifacts_items rows
   └─ Updates run status=done (or error on failure)

4. Frontend receives status=done → renders results
```

---

## Job Queue Mechanics

- **Storage:** PostgreSQL `jobs` table — no external queue dependency
- **Locking:** `SELECT FOR UPDATE SKIP LOCKED` ensures each job is claimed by exactly one worker
- **Retries:** Failed jobs can be retried up to a configurable max attempt count
- **Timeout:** `WORKER_JOB_TIMEOUT` (default `15m`) — jobs exceeding this are marked failed
- **Worker ID:** `WORKER_ID` env var labels which worker processed a job (useful in multi-worker deployments)

---

## Security Model

| Mechanism | Implementation |
|-----------|---------------|
| Session auth | HttpOnly, Secure, SameSite session cookies |
| CSRF protection | Custom `X-CSRF-Protection` header check on mutating requests |
| Rate limiting | Per-IP and per-user limits via in-memory limiter |
| Body size limits | `middleware.Limits` caps request body size |
| Suspicious scan blocking | `middleware.SuspiciousScanBlocker` rejects known scanner paths |
| BYOK encryption | AES-256-GCM via `API_KEY_ENCRYPTION_SECRET` |
| Password storage | bcrypt hashed |
| CORS | `FRONTEND_ORIGIN` allowlist via `middleware.CORS` |

---

## Deployment (Render.com)

The `render.yaml` blueprint defines three services:

| Service | Type | Runtime | Plan |
|---------|------|---------|------|
| `resume-tailor-api` | Web | Go | Free |
| `resume-tailor-worker` | Worker | Go | Starter |
| `resume-tailor-web` | Web | Node | Free |

The worker uses a `starter` plan (not free) because it needs persistent CPU for Tectonic PDF compilation. The API and web frontend use free plans.

Auto-deploy on push to the connected branch is enabled by default on Render.

### Build Commands
- **API:** `go build -trimpath -ldflags "-s -w" -o bin/api ./cmd/api`
- **Worker:** `./scripts/install_tectonic.sh && go build -o bin/worker ./cmd/worker`
- **Web:** `npm ci --include=dev && npm run build`

---

## Environment Variables

| Variable | Required | Default | Purpose |
|----------|----------|---------|---------|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `OPENAI_API_KEY` | Yes | — | Global fallback OpenAI key |
| `OPENAI_MODEL` | No | `gpt-4o-mini` | Model for all LLM calls |
| `API_KEY_ENCRYPTION_SECRET` | No* | — | AES secret for BYOK key storage (*required for BYOK) |
| `HTTP_ADDR` | No | `:8080` | API listen address |
| `FRONTEND_ORIGIN` | No | — | CORS allowlist (comma-separated origins) |
| `WORKER_ID` | No | `worker-1` | Labels worker in logs/DB |
| `WORKER_JOB_TIMEOUT` | No | `15m` | Per-job timeout |
| `DISCIPLINE_MODE` | No | `enforce` | `off`, `observe`, or `enforce` |
| `RESUME_PDF_ENABLED` | No | `0` | Set `1` to enable PDF compilation |
| `TECTONIC_BIN` | No | `tectonic` | Path to Tectonic binary |
| `COOKIE_SECURE` | No | `0` | Set `1` for HTTPS-only cookies |
| `COOKIE_SAMESITE` | No | `lax` | `lax`, `strict`, or `none` |
| `COOKIE_DOMAIN` | No | — | Cookie domain (e.g. `.resumetailor.live`) |
| `GOOGLE_CLIENT_ID` | No | — | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | No | — | Google OAuth client secret |
| `GOOGLE_REDIRECT_URL` | No | — | Google OAuth callback URL |
| `EMAIL_PROVIDER` | No | — | `resend` or `sendgrid` (empty = log-only) |
| `EMAIL_API_KEY` | No | — | API key for email provider |
| `EMAIL_FROM` | No | — | Sender address for transactional email |
| `SENTRY_DSN` | No | — | Sentry error monitoring DSN |
| `DISCORD_WEBHOOK_URL` | No | — | Discord channel webhook for notifications |
| `RUN_MIGRATIONS` | No | `1` | Auto-run migrations on API startup |
| `API_BASE_URL` | No | — | Base URL used in generated links/emails |
