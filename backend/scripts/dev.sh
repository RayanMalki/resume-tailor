#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ -f "$ROOT_DIR/.env" ]]; then
  set -a
  source "$ROOT_DIR/.env"
  set +a
fi

require_env() {
  local name="$1"
  if [[ -z "${!name:-}" ]]; then
    echo "ERROR: $name is required" >&2
    exit 1
  fi
}

require_bin() {
  local name="$1"
  if ! command -v "$name" >/dev/null 2>&1; then
    echo "ERROR: missing required binary: $name" >&2
    exit 1
  fi
}

require_env DATABASE_URL
require_env OPENAI_API_KEY
require_bin docker
require_bin go
require_bin npm

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"

export GOCACHE="${GOCACHE:-/tmp/go-build}"

cleanup() {
  local exit_code=$?
  set +e
  if [[ -n "${API_PID:-}" ]]; then
    kill "$API_PID" >/dev/null 2>&1 || true
  fi
  if [[ -n "${WORKER_PID:-}" ]]; then
    kill "$WORKER_PID" >/dev/null 2>&1 || true
  fi
  if [[ -n "${WEB_PID:-}" ]]; then
    kill "$WEB_PID" >/dev/null 2>&1 || true
  fi
  docker compose -f "$ROOT_DIR/docker-compose.yml" down >/dev/null 2>&1 || true
  exit "$exit_code"
}
trap cleanup EXIT

wait_for_db() {
  local retries=60
  local sleep_s=1

  for _ in $(seq 1 "$retries"); do
    if docker compose -f "$ROOT_DIR/docker-compose.yml" exec -T postgres pg_isready -U app -d resume_tailor >/dev/null 2>&1; then
      return 0
    fi
    sleep "$sleep_s"
  done

  echo "ERROR: Postgres did not become ready" >&2
  exit 1
}

apply_migrations() {
  if [[ ! "$DATABASE_URL" =~ ^postgres://([^:]+):([^@]+)@([^:/]+):?([0-9]+)?/([^?]+) ]]; then
    echo "ERROR: DATABASE_URL must be in postgres://user:pass@host:port/db format" >&2
    exit 1
  fi

  local db_user="${BASH_REMATCH[1]}"
  local db_pass="${BASH_REMATCH[2]}"
  local db_name="${BASH_REMATCH[5]}"
  local mig_dir="$ROOT_DIR/migrations"

  for file in $(ls "$mig_dir"/*.sql | sort); do
    awk 'BEGIN{up=0} $0=="-- +goose Up"{up=1; next} $0=="-- +goose Down"{up=0} up{print}' "$file" | \
      docker compose -f "$ROOT_DIR/docker-compose.yml" exec -T postgres env PGPASSWORD="$db_pass" \
        psql -h 127.0.0.1 -p 5432 -U "$db_user" -d "$db_name" >/dev/null
  done
}

start_api_worker() {
  DATABASE_URL="$DATABASE_URL" OPENAI_API_KEY="$OPENAI_API_KEY" OPENAI_MODEL="${OPENAI_MODEL:-}" \
    go run ./cmd/api > /tmp/rt_api.log 2>&1 &
  API_PID=$!

  DATABASE_URL="$DATABASE_URL" OPENAI_API_KEY="$OPENAI_API_KEY" OPENAI_MODEL="${OPENAI_MODEL:-}" \
    go run ./cmd/worker > /tmp/rt_worker.log 2>&1 &
  WORKER_PID=$!

  local retries=60
  local sleep_s=1

  for _ in $(seq 1 "$retries"); do
    if curl -fsS "$API_BASE_URL/v1/health" >/dev/null 2>&1; then
      return 0
    fi
    sleep "$sleep_s"
  done

  echo "ERROR: API did not become healthy" >&2
  tail -n 200 /tmp/rt_api.log >&2 || true
  exit 1
}

start_web() {
  if [[ ! -d "$ROOT_DIR/web" ]]; then
    echo "ERROR: web directory not found" >&2
    exit 1
  fi

  if [[ ! -d "$ROOT_DIR/web/node_modules" ]]; then
    (cd "$ROOT_DIR/web" && npm install)
  fi

  (cd "$ROOT_DIR/web" && NEXT_PUBLIC_API_BASE_URL="$API_BASE_URL" npm run dev) > /tmp/rt_web.log 2>&1 &
  WEB_PID=$!
}

main() {
  echo "==> starting database"
  docker compose -f "$ROOT_DIR/docker-compose.yml" up -d
  wait_for_db

  echo "==> applying migrations"
  apply_migrations

  echo "==> starting api + worker"
  start_api_worker

  echo "==> starting web"
  start_web

  echo "==> ready"
  echo "API:  $API_BASE_URL"
  echo "WEB:  http://localhost:3000"
  echo "Logs: /tmp/rt_api.log /tmp/rt_worker.log /tmp/rt_web.log"

  wait
}

main
