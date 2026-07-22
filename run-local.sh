#!/usr/bin/env bash
# Run librarr locally with the same config as the server.
# Loads .env, ensures the local-data dirs exist, then starts the server.
set -euo pipefail

cd "$(dirname "$0")"

if ! command -v go >/dev/null 2>&1; then
  echo "Go is not installed. Install it with: brew install go (needs 1.25+)" >&2
  exit 1
fi

if [[ ! -f .env ]]; then
  echo ".env not found. Copy it from the server first." >&2
  exit 1
fi

# Load .env into the environment (the Go code reads os.Getenv, it does not parse .env itself).
set -a
# shellcheck disable=SC1091
source .env
set +a

# Make sure the repo-local data dirs exist.
mkdir -p \
  local-data/books/ebooks \
  local-data/books/audiobooks \
  local-data/books/manga \
  local-data/incoming \
  local-data/manga-incoming

echo "Starting librarr on http://localhost:${LIBRARR_PORT:-5050}"
exec go run ./cmd/librarr/
