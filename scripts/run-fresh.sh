#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT"

if [[ -z "${TMUX:-}" ]]; then
  echo "tmuxwatch must run inside tmux. Launch via tmux new-session -s watch './scripts/run-fresh.sh'" >&2
  exit 1
fi

echo "Cleaning build cache..."
go clean -cache

echo "Rebuilding and running tmuxwatch..."
exec go run ./cmd/tmuxwatch "$@"
