#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT"

echo "Rebuilding and running tmuxwatch with cache disabled..."
exec env GOCACHE=off go run ./cmd/tmuxwatch "$@"
