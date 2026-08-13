#!/usr/bin/env bash
# Idempotent Cloud Agent setup for the Vent repository.
# Installs the CLI toolchain (templ, just, atlas), warms the Go module cache,
# regenerates code, and prepares the example app's SQLite database.
set -euo pipefail

# Pinned versions keep setup deterministic.
TEMPL_VERSION="v0.3.1020" # keep in sync with go.mod's github.com/a-h/templ
JUST_VERSION="1.58.0"
ATLAS_URL="https://atlasbinaries.com/atlas/atlas-community-linux-amd64-latest"

BIN_DIR="$(go env GOPATH)/bin"
mkdir -p "$BIN_DIR"
export PATH="$BIN_DIR:$PATH"

cd "$(dirname "$0")/.."

echo "==> Installing templ ${TEMPL_VERSION}"
if ! command -v templ >/dev/null 2>&1 || [ "$(templ --version 2>/dev/null)" != "$TEMPL_VERSION" ]; then
  go install "github.com/a-h/templ/cmd/templ@${TEMPL_VERSION}"
fi

echo "==> Installing just ${JUST_VERSION}"
if ! command -v just >/dev/null 2>&1 || [ "$(just --version 2>/dev/null)" != "just ${JUST_VERSION}" ]; then
  curl -fsSL https://just.systems/install.sh | bash -s -- --tag "${JUST_VERSION}" --to "$BIN_DIR"
fi

echo "==> Installing atlas CLI"
if ! command -v atlas >/dev/null 2>&1; then
  curl -fsSL "$ATLAS_URL" -o "$BIN_DIR/atlas"
  chmod +x "$BIN_DIR/atlas"
fi

echo "==> Downloading Go modules"
go mod download

echo "==> Generating templ + Vent admin code"
just gen

echo "==> Preparing example SQLite database"
mkdir -p tmp
just migrate

echo "==> Setup complete"
