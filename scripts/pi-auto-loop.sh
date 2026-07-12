#!/usr/bin/env bash
# Autonomous, resumable Pi orchestration driver.
#
# Runs the pm-auto-loop orchestrator (Claude Opus 4.8) headlessly, advancing the
# stage machine one turn per invocation and re-launching until the run reaches a
# terminal state. All progress is durable (ORCHESTRATION-STATE.json + GSD
# artifacts + git + GitHub), so a run killed at any point — including token
# exhaustion — resumes exactly where it stopped: each turn RECONCILES from disk
# before acting.
#
# Usage:
#   scripts/pi-auto-loop.sh "Add full CLI parity for the Freshservice connector"
#   scripts/pi-auto-loop.sh --resume        # continue the current run, no new prompt
#
# Config (env; defaults shown). Set the model IDs once here to whatever your
# `/model` list in Pi shows for your logged-in Claude Max + Codex subscriptions.
#   PI_BIN=pi
#   ORCH_MODEL=anthropic/claude-opus-4-8      # orchestrator (main session) = Claude Opus 4.8
#   PI_TOOLS=read,bash,edit,write,grep,find,ls,subagent
#   MAX_ITERATIONS=200                        # hard backstop on orchestrator turns
#   MAX_MINUTES=0                             # wall-clock cap (0 = no cap)
#   CONTINUE_SESSION=1                        # 1 = `pi -c` (cheaper), fall back to fresh reconcile
#   COOLDOWN_SECONDS=5
#   PI_EXTRA_FLAGS=""                         # extra flags passed to every pi invocation
#   LOOP_CMD=/pm-auto-loop                    # set to /pm-connector-loop for connector runs
#
# For connector work, also export SEARXNG_BASE (and its token if the instance is proxied) so the
# pm-web-researcher subagent can query the audited searxng connector via `pm`. This driver passes
# the ambient environment through to pi, so exporting SEARXNG_BASE before launch is sufficient.
set -euo pipefail

# AUTO_LOOP_RUN_ENTRYPOINT: scripts/pi-auto-loop.sh
case "${1:-}" in
  help|-h|--help)
    printf '%s\n' 'Usage: scripts/pi-auto-loop.sh "<problem prompt>" | --resume'
    exit 0
    ;;
esac

AUTO_LOOP_SCRIPT_DIR="${BASH_SOURCE[0]%/*}"
[[ "$AUTO_LOOP_SCRIPT_DIR" != "${BASH_SOURCE[0]}" ]] || AUTO_LOOP_SCRIPT_DIR="."
AUTO_LOOP_SAFETY_LIB="$AUTO_LOOP_SCRIPT_DIR/auto-loop-safety.sh"
if [[ ! -r "$AUTO_LOOP_SAFETY_LIB" ]]; then
  printf '%s\n' '{"schema_version":"1.0","state":"closed","run_enabled":false,"resume_enabled":false,"code":"AUTO_LOOP_DISABLED_PHASE_0","exit_class":"safety_disabled"}' >&2
  exit 78
fi
# shellcheck source=scripts/auto-loop-safety.sh
source "$AUTO_LOOP_SAFETY_LIB"
if auto_loop_safety_guard_driver "scripts/pi-auto-loop.sh" json; then
  printf '%s\n' 'AUTO_LOOP_GUARD_UNEXPECTED_SUCCESS' >&2
  exit 78
else
  safety_rc=$?
  exit "$safety_rc"
fi

PI_BIN="${PI_BIN:-pi}"
ORCH_MODEL="${ORCH_MODEL:-anthropic/claude-opus-4-8}"
PI_TOOLS="${PI_TOOLS:-read,bash,edit,write,grep,find,ls,subagent}"
MAX_ITERATIONS="${MAX_ITERATIONS:-200}"
MAX_MINUTES="${MAX_MINUTES:-0}"
CONTINUE_SESSION="${CONTINUE_SESSION:-1}"
COOLDOWN_SECONDS="${COOLDOWN_SECONDS:-5}"
PI_EXTRA_FLAGS="${PI_EXTRA_FLAGS:-}"
LOOP_CMD="${LOOP_CMD:-/pm-auto-loop}"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATE_DIR="$REPO_ROOT/.planning/auto-loop"
RUN_JSON="$STATE_DIR/RUN.json"
PROMPT_FILE="$STATE_DIR/PROMPT.txt"
LOG_FILE="$STATE_DIR/driver.log"
mkdir -p "$STATE_DIR"

log() { printf '%s %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*" | tee -a "$LOG_FILE" >&2; }

# --- resolve the problem prompt -------------------------------------------------
if [[ "${1:-}" == "--resume" ]]; then
  [[ -f "$PROMPT_FILE" ]] || { echo "No run to resume (missing $PROMPT_FILE). Pass a prompt." >&2; exit 2; }
  PROBLEM="$(cat "$PROMPT_FILE")"
  log "RESUME run: $PROBLEM"
elif [[ -n "${1:-}" ]]; then
  PROBLEM="$*"
  printf '%s' "$PROBLEM" > "$PROMPT_FILE"
  log "START run: $PROBLEM"
else
  echo "Usage: scripts/pi-auto-loop.sh \"<problem prompt>\"   |   --resume" >&2
  exit 2
fi

# --- read terminal/stage from RUN.json (python3; no jq dependency) --------------
run_field() { # $1 = json key
  [[ -f "$RUN_JSON" ]] || { echo ""; return 0; }
  python3 - "$RUN_JSON" "$1" <<'PY' 2>/dev/null || echo ""
import json,sys
try:
    d=json.load(open(sys.argv[1]))
    v=d.get(sys.argv[2])
    print("" if v is None else v)
except Exception:
    print("")
PY
}

run_pi() { # $1 = "fresh" | "continue"
  local mode="$1" ; local rc=0
  if [[ "$mode" == "continue" ]]; then
    # shellcheck disable=SC2086
    "$PI_BIN" -c -p --tools "$PI_TOOLS" --approve $PI_EXTRA_FLAGS \
      "Continue the autonomous loop: RECONCILE from durable state, then advance exactly one stage. Follow .pi/prompts/pm-auto-loop.md." \
      >>"$LOG_FILE" 2>&1 || rc=$?
  else
    # shellcheck disable=SC2086
    "$PI_BIN" -p --model "$ORCH_MODEL" --tools "$PI_TOOLS" --approve $PI_EXTRA_FLAGS \
      "$LOOP_CMD $PROBLEM" \
      >>"$LOG_FILE" 2>&1 || rc=$?
  fi
  return $rc
}

START_EPOCH="$(date +%s)"
started_session=0

for (( i=1; i<=MAX_ITERATIONS; i++ )); do
  # wall-clock guard
  if (( MAX_MINUTES > 0 )); then
    elapsed_min=$(( ( $(date +%s) - START_EPOCH ) / 60 ))
    if (( elapsed_min >= MAX_MINUTES )); then
      log "STOP: wall-clock cap ${MAX_MINUTES}m reached at turn $i (resumable via --resume)"; exit 3
    fi
  fi

  if (( started_session == 1 )) && [[ "$CONTINUE_SESSION" == "1" ]]; then
    log "turn $i: pi continue"
    if ! run_pi continue; then
      log "turn $i: continue failed; falling back to fresh reconcile"
      run_pi fresh || log "turn $i: fresh run returned non-zero (reconciler will recover next turn)"
    fi
  else
    log "turn $i: pi fresh (reconcile from durable state)"
    run_pi fresh || log "turn $i: fresh run returned non-zero (reconciler will recover next turn)"
    started_session=1
  fi

  terminal="$(run_field terminal)"
  stage="$(run_field stage)"
  log "turn $i: stage=${stage:-?} terminal=${terminal:-none}"

  case "$terminal" in
    human_gate) log "DONE: reached human-ready gate. Human review required before merge to main."; exit 0 ;;
    done)       log "DONE: all sub-issues complete and verified."; exit 0 ;;
    blocked)    log "STOP: run blocked (see ORCHESTRATION-STATE.json for the blocker)."; exit 4 ;;
    budget)     log "STOP: budget ceiling hit. Re-run 'scripts/pi-auto-loop.sh --resume' to continue."; exit 3 ;;
  esac

  sleep "$COOLDOWN_SECONDS"
done

log "STOP: MAX_ITERATIONS=$MAX_ITERATIONS reached without terminal state (resumable via --resume)"
exit 3
