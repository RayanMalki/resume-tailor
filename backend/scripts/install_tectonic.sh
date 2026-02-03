#!/usr/bin/env bash
set -euo pipefail

VERSION="0.13.1"
TARGET="x86_64-unknown-linux-gnu"
ARCHIVE="tectonic-${VERSION}-${TARGET}.tar.gz"
URL="https://github.com/tectonic-typesetting/tectonic/releases/download/tectonic%40${VERSION}/${ARCHIVE}"

mkdir -p bin

curl -L "$URL" | tar -xz -C bin --strip-components=1 "tectonic-${VERSION}-${TARGET}/tectonic"
chmod +x bin/tectonic
