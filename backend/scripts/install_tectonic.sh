#!/usr/bin/env bash
set -euo pipefail

VERSION="0.13.1"
TARGET="x86_64-unknown-linux-gnu"
ARCHIVE="tectonic-${VERSION}-${TARGET}.tar.gz"
URL="https://github.com/tectonic-typesetting/tectonic/releases/download/tectonic%40${VERSION}/${ARCHIVE}"

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
