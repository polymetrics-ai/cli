#!/usr/bin/env bash
# Claude-orchestrated autonomous loop with a Shepherd validator layer.
#
# Each iteration:
#   1. ORCHESTRATOR turn  — `claude -p <claude-orchestrator.md> <problem> [correction]`
#        Claude Code reconciles from durable state and advances ONE stage (it does the
#        Claude-role work itself and dispatches Codex via `pi` for implementation).
#   2. VALIDATOR turn     — `claude -p <shepherd-validator-prompt.md>`
#        A Shepherd supervisor meta-agent scores that (state,action) step 1-5 (Anthropic
#        rubric, geometric mean) and writes a verdict.
#   3. ACT on the verdict — PROCEED (checkpoint + continue) / RETRY (re-run with correction)
#        / REVERT (restore last checkpoint + replay the stage) / HALT (stop for human).
#
# All progress is durable (git commits + RUN.json/ORCHESTRATION-STATE.json + GitHub), so a run
# killed anywhere resumes by reconciling. Claude runs on your Max plan via the first-party CLI
# (no Pi third-party "extra usage" gate); Codex implements via `pi --model openai-codex/gpt-5.5`.
#
# Usage:
#   scripts/claude-auto-loop.sh "twenty (Twenty CRM, https://twenty.com) full all-ops CLI parity"
#   scripts/claude-auto-loop.sh --resume
#
# Config (env; defaults shown):
#   CLAUDE_BIN=claude
#   CLAUDE_ARGS="--output-format text"        # headless autonomy also needs a permission posture,
#                                             # e.g. append: --permission-mode acceptEdits  (safer)
#                                             # or (full unattended): --dangerously-skip-permissions
#   MAX_ITERATIONS=300                        # hard backstop on orchestrator turns
#   MAX_REVERTS=6                             # total revert budget per run before HALT
#   MAX_MINUTES=0                             # wall-clock cap (0 = none)
#   COOLDOWN_SECONDS=4
#   SEARXNG_BASE=                             # your self-hosted SearXNG for research (optional)
set -euo pipefail

CLAUDE_BIN="${CLAUDE_BIN:-claude}"
# Default includes a permission posture so a headless `claude -p` can actually write its
# verdict/state files. Without one, every turn produces no verdict and the loop no-ops. For a
# fully unattended run that also needs bash (gh/make/pi), export
# CLAUDE_ARGS="--output-format text --dangerously-skip-permissions".
CLAUDE_ARGS="${CLAUDE_ARGS:---output-format text --permission-mode acceptEdits}"
MAX_ITERATIONS="${MAX_ITERATIONS:-300}"
MAX_REVERTS="${MAX_REVERTS:-6}"
MAX_NO_VERDICT="${MAX_NO_VERDICT:-3}"          # consecutive no-verdict turns before HALT (footgun guard)
MAX_MINUTES="${MAX_MINUTES:-0}"
COOLDOWN_SECONDS="${COOLDOWN_SECONDS:-4}"

# Fail fast on the most common footgun: no permission posture => headless claude can't write files.
case " $CLAUDE_ARGS " in
  *" --permission-mode "*|*" --dangerously-skip-permissions "*) : ;;
  *) echo "WARN: CLAUDE_ARGS has no permission posture; headless 'claude -p' can't write files. Add --permission-mode acceptEdits (or --dangerously-skip-permissions for fully unattended)." >&2 ;;
esac

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATE_DIR="$REPO_ROOT/.planning/auto-loop"
CKPT_DIR="$STATE_DIR/checkpoints"
RUN_JSON="$STATE_DIR/RUN.json"
VERDICT_JSON="$STATE_DIR/VALIDATOR-VERDICT.json"
PROMPT_FILE="$STATE_DIR/PROMPT.txt"
LOG_FILE="$STATE_DIR/driver.log"
ORCH_PROMPT="$REPO_ROOT/.agents/agentic-delivery/prompts/claude-orchestrator.md"
VAL_PROMPT="$REPO_ROOT/.agents/agentic-delivery/prompts/shepherd-validator-prompt.md"
mkdir -p "$STATE_DIR" "$CKPT_DIR"

log() { printf '%s %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*" | tee -a "$LOG_FILE" >&2; }

if [[ "${1:-}" == "--resume" ]]; then
  [[ -f "$PROMPT_FILE" ]] || { echo "No run to resume (missing $PROMPT_FILE)." >&2; exit 2; }
  PROBLEM="$(cat "$PROMPT_FILE")"; log "RESUME: $PROBLEM"
elif [[ -n "${1:-}" ]]; then
  PROBLEM="$*"; printf '%s' "$PROBLEM" > "$PROMPT_FILE"; log "START: $PROBLEM"
else
  echo "Usage: scripts/claude-auto-loop.sh \"<problem prompt>\" | --resume" >&2; exit 2
fi

json_field() { # $1=file $2=key
  [[ -f "$1" ]] || { echo ""; return 0; }
  python3 - "$1" "$2" <<'PY' 2>/dev/null || echo ""
import json,sys
try:
    d=json.load(open(sys.argv[1])); v=d.get(sys.argv[2]); print("" if v is None else v)
except Exception: print("")
PY
}

run_claude() { # $1=prompt-file $2=extra-inline-text  -> stdout to driver.log
  local pf="$1" extra="${2:-}" rc=0
  # shellcheck disable=SC2086
  "$CLAUDE_BIN" -p $CLAUDE_ARGS "$(cat "$pf")

$extra" >>"$LOG_FILE" 2>&1 || rc=$?
  return $rc
}

checkpoint() { # $1=turn — snapshot the run ledger + the worktree HEAD SHA (the fork point).
  local d="$CKPT_DIR/$1"; mkdir -p "$d"
  [[ -f "$RUN_JSON" ]] && cp "$RUN_JSON" "$d/RUN.json" 2>/dev/null || true
  git -C "$REPO_ROOT" rev-parse HEAD > "$d/HEAD.sha" 2>/dev/null || true
  echo "$1" > "$CKPT_DIR/LAST_GOOD"
}
restore_checkpoint() { # Restore the run ledger to the last PROCEED checkpoint AND hand the
  # orchestrator an explicit cleanup task for the diverged commits. The validator never rewrites
  # pushed git history itself (shepherd-validator.md); it records the good fork-point SHA and the
  # current (bad) SHA in REVERT-CLEANUP.json, and the next RECONCILE resets local-only divergence
  # or reverts-forward pushed commits per its own gates. Reverting RUN.json alone is not enough —
  # this is what makes REVERT actually undo the bad step rather than re-derive forward from it.
  local last; last="$(cat "$CKPT_DIR/LAST_GOOD" 2>/dev/null || echo "")"
  local good_sha cur_sha; cur_sha="$(git -C "$REPO_ROOT" rev-parse HEAD 2>/dev/null || echo "")"
  if [[ -n "$last" && -f "$CKPT_DIR/$last/RUN.json" ]]; then
    cp "$CKPT_DIR/$last/RUN.json" "$RUN_JSON"
    good_sha="$(cat "$CKPT_DIR/$last/HEAD.sha" 2>/dev/null || echo "")"
    python3 - "$STATE_DIR/REVERT-CLEANUP.json" "$good_sha" "$cur_sha" "$last" <<'PY' 2>/dev/null || true
import json,sys
json.dump({"good_fork_sha":sys.argv[2],"diverged_head_sha":sys.argv[3],"checkpoint":sys.argv[4],
           "instruction":"REVERT: reset local-only commits after good_fork_sha, or revert-forward pushed commits per your gates; never force-push. Then replay the stage."},
          open(sys.argv[1],"w"),indent=2)
PY
    log "reverted to checkpoint $last (fork ${good_sha:0:8}); wrote REVERT-CLEANUP.json for orchestrator to undo diverged commits ${cur_sha:0:8}"
  else
    log "no checkpoint to revert to"
  fi
}

START_EPOCH="$(date +%s)"; reverts=0; no_verdict=0; correction=""
for (( i=1; i<=MAX_ITERATIONS; i++ )); do
  if (( MAX_MINUTES > 0 )) && (( ( $(date +%s) - START_EPOCH ) / 60 >= MAX_MINUTES )); then
    log "STOP: wall-clock cap ${MAX_MINUTES}m (resumable via --resume)"; exit 3
  fi

  log "── turn $i: ORCHESTRATOR ──${correction:+ (with correction)}"
  run_claude "$ORCH_PROMPT" "PROBLEM: $PROBLEM${correction:+

VALIDATOR CORRECTION (apply first): $correction}" \
    || log "turn $i: orchestrator returned non-zero (validator will assess)"

  log "── turn $i: VALIDATOR ──"
  run_claude "$VAL_PROMPT" "" || log "turn $i: validator returned non-zero"

  verdict="$(json_field "$VERDICT_JSON" verdict)"
  score="$(json_field "$VERDICT_JSON" step_score)"
  reason="$(json_field "$VERDICT_JSON" reason)"
  correction=""
  log "turn $i: verdict=${verdict:-NONE} step_score=${score:-?} — ${reason:-}"

  case "$verdict" in
    PROCEED) no_verdict=0; checkpoint "$i" ;;
    RETRY)   no_verdict=0; correction="$(json_field "$VERDICT_JSON" correction)"; log "turn $i: RETRY — $correction" ;;
    REVERT)
      no_verdict=0; reverts=$((reverts+1))
      if (( reverts > MAX_REVERTS )); then log "HALT: MAX_REVERTS=$MAX_REVERTS exceeded"; exit 4; fi
      restore_checkpoint; correction="$(json_field "$VERDICT_JSON" correction)"; log "turn $i: REVERT #$reverts — $correction" ;;
    HALT)    log "HALT: validator hard-stop — ${reason:-}"; exit 4 ;;
    *)
      no_verdict=$((no_verdict+1))
      if (( no_verdict >= MAX_NO_VERDICT )); then
        log "HALT: validator produced no VALIDATOR-VERDICT.json for $no_verdict consecutive turns. Most likely CLAUDE_ARGS lacks a permission posture so headless 'claude -p' can't write files — set --permission-mode acceptEdits (or --dangerously-skip-permissions), verify 'claude -p' is logged in, then --resume."
        exit 4
      fi
      log "turn $i: no verdict ($no_verdict/$MAX_NO_VERDICT); retrying"
      correction="Emit a VALIDATOR-VERDICT.json with a verdict and cited evidence." ;;
  esac

  terminal="$(json_field "$RUN_JSON" terminal)"; stage="$(json_field "$RUN_JSON" stage)"
  log "turn $i: stage=${stage:-?} terminal=${terminal:-none}"
  case "$terminal" in
    human_gate) log "DONE: human-ready gate reached (human review before merge to main)."; exit 0 ;;
    done)       log "DONE: all sub-issues complete and verified."; exit 0 ;;
    blocked)    log "STOP: blocked (see ORCHESTRATION-STATE.json / VALIDATION.jsonl)."; exit 4 ;;
    budget)     log "STOP: budget ceiling; re-run --resume."; exit 3 ;;
  esac

  sleep "$COOLDOWN_SECONDS"
done
log "STOP: MAX_ITERATIONS=$MAX_ITERATIONS without terminal (resumable via --resume)"; exit 3
