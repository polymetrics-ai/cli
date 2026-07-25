#!/usr/bin/env bash
# Compatibility entry point for the canonical Pi orchestration driver.
#
# Terminal state must never be interpreted by an unsupervised driver. Delegate with exec so
# scripts/pi-shepherd-loop.sh owns clean-synthesis authentication, exact identity binding,
# independent validation, watchdogs, human gates, pm-terminal-classifier.sh handling for every
# blocked human decision, and the final process status. The Shepherd
# validator itself is invoked only after an authenticated clean exact-head synthesis exists.
# Neither driver merges; final merge authority remains human-only.
#
# Usage:
#   scripts/pi-auto-loop.sh "<problem prompt>"
#   scripts/pi-auto-loop.sh --resume
set -euo pipefail

ORCH_MODEL="${ORCH_MODEL:-openai-codex/gpt-5.6-sol}"
ORCH_THINKING="${ORCH_THINKING:-xhigh}"
PI_TOOLS="${PI_TOOLS:-read,bash,edit,write,grep,find,ls,subagent}"
LOOP_CMD="${LOOP_CMD:-/pm-auto-loop}"
export ORCH_MODEL ORCH_THINKING PI_TOOLS LOOP_CMD

# Routing parity is implemented by the delegated driver with: --thinking "$ORCH_THINKING".
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SHEPHERD_DRIVER="${SHEPHERD_DRIVER:-$REPO_ROOT/scripts/pi-shepherd-loop.sh}"
if [[ ! -x "$SHEPHERD_DRIVER" ]]; then
  printf 'FATAL: canonical Shepherd driver is not executable: %s\n' "$SHEPHERD_DRIVER" >&2
  exit 2
fi

exec "$SHEPHERD_DRIVER" "$@"
