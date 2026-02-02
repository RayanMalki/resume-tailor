# Backend Verification (v1)

Prerequisites
- Docker + Docker Compose
- Go toolchain (per go.mod)
- jq (for smoke tests): `brew install jq`

Environment variables
- `DATABASE_URL` (required)
- `OPENAI_API_KEY` (required for worker)
- `OPENAI_MODEL` (optional)
- `API_BASE_URL` (optional, default `http://localhost:8080`)

Quick start
```
make verify
```

Smoke test only
```
make smoke
```
Note: `make smoke` assumes the API and worker are already running.

What verify does
- Runs gofmt check and `go test ./...`
- Starts a fresh Postgres via Docker Compose
- Applies migrations in order (001/002/003)
- Starts API + worker
- Runs an end-to-end smoke test (signup → resume → run → report)
- Tears everything down on completion or failure

Expected runtime
- Typically 30–90 seconds depending on Docker startup and model latency

Troubleshooting
- Docker not running: start Docker Desktop and retry
- Port conflicts: ensure 8080 and 5433 are available, or set `API_BASE_URL` and adjust `docker-compose.yml`
- Missing jq: install via `brew install jq`
- Worker fails: ensure `OPENAI_API_KEY` is set and valid
