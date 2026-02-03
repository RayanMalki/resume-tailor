#!/usr/bin/env bash
set -euo pipefail

VERSION="${TECTONIC_VERSION:-0.15.0}"
MUSL_TARGET="x86_64-unknown-linux-musl"
GNU_TARGET="x86_64-unknown-linux-gnu"

BASE_URL="https://github.com/tectonic-typesetting/tectonic/releases/download/tectonic%40${VERSION}"
MUSL_ARCHIVE="tectonic-${VERSION}-${MUSL_TARGET}.tar.gz"
GNU_ARCHIVE="tectonic-${VERSION}-${GNU_TARGET}.tar.gz"

URL="${BASE_URL}/${MUSL_ARCHIVE}"
ARCHIVE="${MUSL_ARCHIVE}"

if ! curl -fsI "$URL" >/dev/null 2>&1; then
  URL="${BASE_URL}/${GNU_ARCHIVE}"
  ARCHIVE="${GNU_ARCHIVE}"
fi

mkdir -p bin

TMP_DIR="$(mktemp -d)"
ARCHIVE_PATH="${TMP_DIR}/${ARCHIVE}"

curl -L "$URL" -o "$ARCHIVE_PATH"
tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"

TECTONIC_PATH="$(find "$TMP_DIR" -type f -name "tectonic" | head -n 1)"
if [[ -z "$TECTONIC_PATH" ]]; then
  echo "tectonic binary not found in archive" >&2
  exit 1
fi

cp "$TECTONIC_PATH" bin/tectonic
chmod +x bin/tectonic

rm -rf "$TMP_DIR"
