# Resume Tailor

Resume Tailor analyzes a resume against a job description and generates a full set of tailored application assets — ATS report, improved resume, and cover letter.

It combines profile-aware BM25 keyword scoring with LLM generation and deterministic rendering in a production-style architecture.

---

## What It Generates

- **ATS report** — score (0–100%), keyword gaps, AI-written summary, and interview questions
- **Programmatic change plan** — keyword diff used to drive resume improvements
- **Tailored resume** — LaTeX source, DOCX, and optional PDF
- **Cover letter** — AI-generated text, DOCX, and optional PDF
- **Project relevance reasons** — when project controls are provided

---

## Architecture at a Glance

```
Browser → Next.js Frontend → Go API → PostgreSQL ← Go Worker → OpenAI
                                    ↖ jobs table ↗
```

Three services, one database:

| Service | Role |
|---------|------|
| **Go API** (`cmd/api`) | Auth, resumes, run creation, artifact downloads |
| **Go Worker** (`cmd/worker`) | BM25 scoring, OpenAI calls, LaTeX/DOCX rendering |
| **Next.js** (`web/`) | UI, proxies `/api/*` to the Go API |

The API writes jobs to a PostgreSQL queue and returns immediately. The worker polls the queue and processes runs asynchronously.

---

## Technology Stack

| Layer | Technology |
|-------|-----------|
| Backend language | Go 1.24 |
| Frontend framework | Next.js 14+ (App Router, TypeScript, Tailwind) |
| Database | PostgreSQL 15+ |
| Job queue | PostgreSQL (SQL-backed, no external broker) |
| LLM | OpenAI (gpt-4o-mini by default, configurable) |
| PDF compilation | Tectonic (optional) |
| Deployment | Render.com (3-service blueprint in `render.yaml`) |

---

## Documentation

| Document | Description |
|----------|-------------|
| [docs/user-journey.md](docs/user-journey.md) | End-to-end user flow: landing → upload → results → export |
| [docs/system-design.md](docs/system-design.md) | Architecture, DB schema, request lifecycle, security, deployment |
| [docs/scoring.md](docs/scoring.md) | BM25 algorithm, discipline profiles, IDF table, phrase detection |
| [docs/api-reference.md](docs/api-reference.md) | All HTTP endpoints with request/response details |
| [docs/local-development.md](docs/local-development.md) | Full local dev setup guide |
| [backend/README.md](backend/README.md) | Backend directory layout, key packages, how to run/test |
| [backend/web/README.md](backend/web/README.md) | Frontend pages, components, env vars, how to run |

---

## Quick Start

See [docs/local-development.md](docs/local-development.md) for the full guide.

**Fastest path:**
```bash
cd backend
cp .env.example .env
# Set DATABASE_URL and OPENAI_API_KEY in .env
make dev
```

`make dev` starts Postgres, runs migrations, starts the API + worker, and launches the Next.js dev server.

---

## Key Capabilities

- Email/password auth with HttpOnly session cookies
- Google OAuth login
- Resume upload (`.pdf` and `.docx`) with server-side text extraction
- Profile-aware BM25 scoring across 5 engineering disciplines
- Discipline auto-detection from resume content
- Async run queue with retries and configurable timeout
- BYOK (Bring Your Own OpenAI Key), encrypted at rest with AES-256-GCM
- PDF compilation via Tectonic (optional)
- CSRF protection, per-IP and per-user rate limiting, suspicious scan blocking
- Optional Sentry error monitoring and Discord webhook notifications

### Rate and Usage Limits

| Limit | Value |
|-------|-------|
| Global IP rate limit | 60 req/min |
| Login | 5/min per IP |
| Signup | 3/hour per IP |
| Resume upload | 10/min per IP |
| Run creation (burst) | 1/min per user |
| Run creation (daily) | 10/day per user |
| Free-tier (no BYOK) | 3/day per user, 6/day per creator IP |

---

## Repository Layout

```
/
├── README.md              This file
├── docs/                  Topic-specific documentation
│   ├── user-journey.md
│   ├── system-design.md
│   ├── scoring.md
│   ├── api-reference.md
│   └── local-development.md
├── backend/               Go backend + Next.js frontend
│   ├── cmd/               Entry points (api, worker, migrate)
│   ├── internal/          All Go library code
│   ├── migrations/        SQL migrations
│   ├── web/               Next.js app
│   ├── scripts/           Dev/CI scripts
│   └── docs/              Internal backend reference docs
└── render.yaml            Render.com deployment blueprint
```

---

## Configuration

Critical environment variables (set in `backend/.env`):

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `OPENAI_API_KEY` | Yes | Global OpenAI key (fallback when user has no BYOK key) |
| `API_KEY_ENCRYPTION_SECRET` | No* | Enables BYOK — AES-256-GCM encryption secret |
| `OPENAI_MODEL` | No | Model name (default `gpt-4o-mini`) |
| `RESUME_PDF_ENABLED` | No | Set `1` to enable PDF compilation via Tectonic |

Full variable reference: [docs/system-design.md#environment-variables](docs/system-design.md#environment-variables)

---

## Deployment

The `render.yaml` blueprint defines:
- `resume-tailor-api` — Go web service (free plan)
- `resume-tailor-worker` — Go worker (starter plan, needs persistent CPU for Tectonic)
- `resume-tailor-web` — Next.js web service (free plan)

See [docs/system-design.md#deployment](docs/system-design.md#deployment-rendercom) for details.

---

## Testing

From `backend/`:

```bash
make test     # go test ./...
make smoke    # e2e smoke flow (API + worker must be running)
make verify   # full check: fmt + tests + build + smoke
```

---

## Additional Internal Docs

- `backend/docs/discipline-profiles.md` — profile governance and update process
- `backend/docs/testing.md` — testing strategy notes
- `backend/docs/bm25.md` — legacy BM25 signal reference
