#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_PATH="${TMUXWATCH_DAEMON_LOG:-/tmp/tmuxwatch-daemon.log}"

export TMUXWATCH_FORCE_TMUX="${TMUXWATCH_FORCE_TMUX:-0}"

cd "$ROOT"

POLTERGEIST_BIN="${POLTERGEIST_BIN:-}"
if [[ -z "$POLTERGEIST_BIN" ]]; then
	if [[ -x "$ROOT/../poltergeist/dist/cli.js" ]]; then
		POLTERGEIST_BIN="$ROOT/../poltergeist/dist/cli.js"
	else
		POLTERGEIST_BIN="poltergeist"
	fi
fi

POLTER_BIN="${POLTER_BIN:-}"
USING_DIST_POLTER=0
if [[ -z "$POLTER_BIN" ]]; then
	if [[ -x "$ROOT/../poltergeist/dist/polter.js" ]]; then
		POLTER_BIN="$ROOT/../poltergeist/dist/polter.js"
		USING_DIST_POLTER=1
	else
		POLTER_BIN="polter"
	fi
fi

if [[ $USING_DIST_POLTER -eq 0 ]]; then
	if ! "$POLTER_BIN" --help 2>&1 | grep -q -- '--watch'; then
		echo "[tmuxwatch] This helper requires a polter build with --watch support."
		echo "            Either run 'pnpm build' in ../poltergeist and retry, or install the latest release."
		exit 1
	fi
elif [[ ! -x "$POLTER_BIN" ]]; then
	echo "[tmuxwatch] Unable to find executable polter binary (expected at $POLTER_BIN)."
	exit 1
fi

if ! "$POLTERGEIST_BIN" status --target tmuxwatch-cli >/dev/null 2>&1; then
	echo "[tmuxwatch] Starting poltergeist daemon (logs -> $LOG_PATH)..."
	nohup "$POLTERGEIST_BIN" haunt --target tmuxwatch-cli >/dev/null 2>>"$LOG_PATH" &
	sleep 1
else
	echo "[tmuxwatch] poltergeist daemon already running."
fi

echo "[tmuxwatch] Launching tmuxwatch with hot reload..."
exec "$POLTER_BIN" tmuxwatch-cli --watch "$@"
