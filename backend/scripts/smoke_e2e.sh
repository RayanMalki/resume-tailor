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
    if [[ "$name" == "jq" ]]; then
      echo "ERROR: missing required binary: jq (install with: brew install jq)" >&2
    else
      echo "ERROR: missing required binary: $name" >&2
    fi
    exit 1
  fi
}

require_env DATABASE_URL
require_env OPENAI_API_KEY
require_bin curl
require_bin jq

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"

WORK_DIR="$(mktemp -d)"
COOKIE_JAR="$WORK_DIR/cookies.txt"

cleanup() {
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

rand() {
  date +%s%N | tail -c 6
}

email="test+$(rand)@example.com"
password="TestPass123!"

signup() {
  curl -fsS -c "$COOKIE_JAR" \
    -X POST "$API_BASE_URL/v1/auth/signup" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$email\",\"password\":\"$password\",\"displayName\":\"Smoke Test\"}" >/dev/null
}

create_resume() {
  local payload
  payload=$(curl -fsS -b "$COOKIE_JAR" \
    -X POST "$API_BASE_URL/v1/resumes" \
    -H "Content-Type: application/json" \
    -d '{"title":"Smoke Resume","contentText":"Go developer with Docker and Kubernetes"}')

  echo "$payload" | jq -e -r '.resumeId' >/dev/null
  echo "$payload" | jq -r '.resumeId'
}

create_run() {
  local resume_id="$1"
  local payload
  payload=$(curl -fsS -b "$COOKIE_JAR" \
    -X POST "$API_BASE_URL/v1/runs" \
    -H "Content-Type: application/json" \
    -d "{\"resumeId\":\"$resume_id\",\"jobText\":\"Go developer with Docker Kubernetes CI/CD\"}")

  echo "$payload" | jq -e -r '.runId' >/dev/null
  echo "$payload" | jq -r '.runId'
}

poll_run() {
  local run_id="$1"
  local timeout_s=120
  local interval_s=1
  local elapsed=0

  while [[ $elapsed -lt $timeout_s ]]; do
    local payload
    payload=$(curl -fsS -b "$COOKIE_JAR" "$API_BASE_URL/v1/runs/$run_id")
    local status
    status=$(echo "$payload" | jq -r '.status')

    if [[ "$status" == "succeeded" ]]; then
      echo "$payload" >/dev/null
      return 0
    fi

    if [[ "$status" == "failed" ]]; then
      echo "ERROR: run failed" >&2
      echo "$payload" >&2
      return 1
    fi

    sleep "$interval_s"
    elapsed=$((elapsed + interval_s))
  done

  echo "ERROR: run did not complete within ${timeout_s}s" >&2
  return 1
}

fetch_report() {
  local run_id="$1"
  local payload
  payload=$(curl -fsS -b "$COOKIE_JAR" "$API_BASE_URL/v1/runs/$run_id/report")

  echo "$payload" | jq -e '.report_version == 1' >/dev/null
  echo "$payload" | jq -e '.bm25_signals' >/dev/null
  echo "$payload" | jq -e '.generated_at' >/dev/null

  echo "$payload"
}

main() {
  echo "==> signup"
  signup

  echo "==> create resume"
  resume_id=$(create_resume)

  echo "==> create run"
  run_id=$(create_run "$resume_id")

  echo "==> poll run"
  poll_run "$run_id"

  echo "==> fetch report"
  report=$(fetch_report "$run_id")

  echo "==> success"
  echo "run_id=$run_id"
  echo "bm25.score=$(echo "$report" | jq -r '.bm25_signals.score')"
  echo "bm25.top_terms=$(echo "$report" | jq -c '.bm25_signals.top_job_terms[0:3]')"
}

main
