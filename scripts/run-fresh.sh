#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT"

echo "Cleaning build cache..."
go clean -cache

echo "Rebuilding and running tmuxwatch..."
exec go run ./cmd/tmuxwatch "$@"
