# Local Development

This guide covers setting up the full Resume Tailor stack locally: Go API, Go worker, PostgreSQL, and the Next.js frontend.

---

## Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| Go | 1.24+ | [golang.org/dl](https://golang.org/dl) |
| Node.js | 18+ | [nodejs.org](https://nodejs.org) |
| Docker | Any recent | Used to run PostgreSQL via `docker compose` |
| OpenAI API key | — | Required to run the worker (LLM calls) |
| Tectonic | Optional | Required only if you want PDF export |

---

## Quick Start (Recommended)

From the `backend/` directory:

```bash
cd backend
cp .env.example .env
# Edit .env — set DATABASE_URL and OPENAI_API_KEY at minimum
make dev
```

`make dev` starts Postgres via Docker Compose, applies migrations, runs the API, runs the worker, and starts the Next.js frontend — all in one command.

---

## Manual Setup

### 1. Clone and configure

```bash
git clone <repo-url>
cd resume-tailor/backend
cp .env.example .env
```

Open `.env` and set at minimum:
```bash
DATABASE_URL=postgres://app:app_password@localhost:5433/resume_tailor?sslmode=disable
OPENAI_API_KEY=sk-...
```

Optional but recommended for BYOK support:
```bash
API_KEY_ENCRYPTION_SECRET=<random-32-byte-hex>
```

### 2. Start PostgreSQL

```bash
docker compose up -d
```

This starts PostgreSQL on port `5433` (not the default 5432, to avoid conflicts).

### 3. Run migrations

Migrations run automatically when the API starts if `RUN_MIGRATIONS=1` (the default). To run them manually:

```bash
go run ./cmd/migrate
```

Or see `scripts/verify.sh` for the simple migration runner approach.

### 4. Start the API server

```bash
go run ./cmd/api
```

The API listens on `:8080` by default (`HTTP_ADDR` env var).

### 5. Start the worker

In a separate terminal:

```bash
go run ./cmd/worker
```

The worker polls the `jobs` table every few seconds and processes pending runs.

### 6. Start the frontend

```bash
cd web
cp .env.local.example .env.local   # if it exists, otherwise create manually
# Set:
# NEXT_PUBLIC_API_BASE_URL=/api
# API_PROXY_TARGET=http://localhost:8080
# NEXT_PUBLIC_SITE_URL=http://localhost:3000

npm install
npm run dev
```

The frontend is available at `http://localhost:3000`. All `/api/*` requests are proxied to the Go API at `http://localhost:8080`.

---

## Environment Variables Reference

See [system-design.md](./system-design.md#environment-variables) for the full table.

Key values for local development:

```bash
# .env (in backend/)
DATABASE_URL=postgres://app:app_password@localhost:5433/resume_tailor?sslmode=disable
OPENAI_API_KEY=sk-...
HTTP_ADDR=":8080"
FRONTEND_ORIGIN="http://localhost:3000"
COOKIE_SECURE="0"
COOKIE_SAMESITE="lax"
COOKIE_DOMAIN=""
```

```bash
# .env.local (in backend/web/)
NEXT_PUBLIC_API_BASE_URL=/api
API_PROXY_TARGET=http://localhost:8080
NEXT_PUBLIC_SITE_URL=http://localhost:3000
```

---

## Running Tests

From `backend/`:

```bash
# Run all Go tests
make test
# or
go test ./...

# Run smoke tests (requires API + worker to be running)
make smoke

# Run full verification (format checks + tests + integration boot + smoke)
make verify
```

Notes:
- `make smoke` requires the API and worker to be running and `OPENAI_API_KEY` to be set
- `make verify` is the closest thing to the CI check — run it before opening a PR

---

## PDF Export (Optional)

PDF compilation requires [Tectonic](https://tectonic-typesetting.github.io/), a TeX engine.

**Install Tectonic:**
```bash
# macOS
brew install tectonic

# Linux (see tectonic docs for your distro)
```

**Enable in `.env`:**
```bash
RESUME_PDF_ENABLED=1
TECTONIC_BIN=tectonic   # or the full path if not in PATH
```

When enabled, the worker will compile the generated LaTeX resume to PDF and store it as the `resume-pdf` artifact.

---

## CORS and Cookies

The API enforces CORS via `FRONTEND_ORIGIN`. For local development:
- Set `FRONTEND_ORIGIN=http://localhost:3000` in the backend `.env`
- All frontend requests use `credentials: "include"` so the session cookie is sent

The Next.js proxy (`/api/*` → `http://localhost:8080`) avoids cross-origin cookie issues in local dev, matching the production setup where the frontend and API share the same origin.

---

## Common Issues

### Port 5432 already in use
The `docker-compose.yml` uses port `5433` externally to avoid conflicts with a local Postgres instance. If 5433 is also taken, edit `docker-compose.yml` to use a different host port.

### Migrations fail to connect
Ensure Docker Compose is running (`docker compose ps`) and `DATABASE_URL` in `.env` matches the Compose service credentials.

### Worker doesn't pick up jobs
- Confirm the worker is running (`go run ./cmd/worker`)
- Check that `DATABASE_URL` is set in the environment where the worker runs
- Look for errors in the worker log output

### PDF endpoint returns 404
PDF generation is opt-in. Set `RESUME_PDF_ENABLED=1` and ensure `tectonic` is installed and in PATH (or `TECTONIC_BIN` points to the binary).
