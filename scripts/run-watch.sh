#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ -z "${TMUX:-}" ]]; then
  if [[ "${TMUXWATCH_FORCE_TMUX:-0}" == "1" ]]; then
    echo "tmuxwatch must run inside tmux. Launch via: tmux new-session -s watch './scripts/run-watch.sh'" >&2
    exit 1
  fi
  echo "[tmuxwatch] TMUX not detected; watch mode works best inside tmux. Set TMUXWATCH_FORCE_TMUX=1 to require tmux." >&2
fi

if ! command -v poltergeist >/dev/null 2>&1 || ! command -v polter >/dev/null 2>&1; then
  echo "[tmuxwatch] poltergeist CLI not found. Install via 'pnpm install' in ~/Projects/poltergeist or npm install -g @steipete/poltergeist." >&2
  exit 1
fi

mkdir -p "$ROOT/dist/dev"

cd "$ROOT"

echo "[tmuxwatch] Ensuring poltergeist daemon is running (target: tmuxwatch-cli)..."
if ! poltergeist haunt --target tmuxwatch-cli >/dev/null 2>&1; then
  echo "[tmuxwatch] poltergeist daemon already running or failed to launch; continuing..." >&2
fi

echo "[tmuxwatch] Starting tmuxwatch in watch mode (auto-restart on successful builds)..."
exec polter tmuxwatch-cli --watch "$@"
