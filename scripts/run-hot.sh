#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_PATH="${TMUXWATCH_DAEMON_LOG:-/tmp/tmuxwatch-daemon.log}"

# Allow running outside tmux unless explicitly forced
export TMUXWATCH_FORCE_TMUX="${TMUXWATCH_FORCE_TMUX:-0}"

cd "$ROOT"

if ! poltergeist status --target tmuxwatch-cli >/dev/null 2>&1; then
  echo "[tmuxwatch] Starting poltergeist daemon (logs -> $LOG_PATH)..."
  nohup poltergeist haunt --target tmuxwatch-cli >/dev/null 2>>"$LOG_PATH" &
  sleep 1
else
  echo "[tmuxwatch] poltergeist daemon already running."
fi

echo "[tmuxwatch] Launching tmuxwatch with hot reload..."
exec polter tmuxwatch-cli --watch "$@"
