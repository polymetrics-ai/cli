#!/usr/bin/env bash
# Codex-only autonomous Pi orchestration driver WITH the Shepherd validator layer.
#
# Best-of-both-worlds merge of scripts/pi-auto-loop.sh (pi-native orchestrator with the
# `subagent` tool and the .pi/agents/pm-* roster) and scripts/claude-auto-loop.sh (Shepherd
# supervision: an independent validator judges every orchestrator turn, with checkpoints,
# RETRY corrections, REVERT-to-checkpoint, and HALT).
#
# Each iteration:
#   1. ORCHESTRATOR turn  — `pi -p` runs the loop prompt (default /pm-auto-loop; use
#        LOOP_CMD=/pm-connector-loop for connector runs). It RECONCILES from durable state,
#        advances exactly ONE stage, and dispatches implementation via the pi `subagent` tool
#        (or the detached-worker recipe for long EXECUTE stages).
#   2. VALIDATOR turn     — an independent Shepherd (default: Codex via pi) re-derives ground
#        truth (git/gh/disk/ps), scores the turn on the six-dimension rubric (geometric mean),
#        and writes VALIDATOR-VERDICT.json + appends VALIDATION.jsonl.
#   3. ACT on the verdict — PROCEED (checkpoint + continue) / RETRY (replay stage with the
#        correction) / REVERT (restore last checkpoint, write REVERT-CLEANUP.json, replay) /
#        HALT (stop for a human).
#
# All progress is durable (git + GitHub + RUN.json/ORCHESTRATION-STATE.json), so a run killed
# anywhere resumes by reconciling: scripts/pi-shepherd-loop.sh --resume
#
# Model policy: Codex on the ChatGPT plan via the `openai-codex/*` provider ONLY. Never route
# any role through OpenRouter or another pay-per-token gateway.
#
# Requires the `subagent` tool package once per machine:  pi install npm:pi-sub-agent
# (project agents in .pi/agents/ are auto-discovered when running with --approve).
#
# Usage:
#   scripts/pi-shepherd-loop.sh "Add full CLI parity for the <name> connector"
#   scripts/pi-shepherd-loop.sh --resume
#
# Config (env; defaults shown):
#   PI_BIN=pi
#   ORCH_MODEL=openai-codex/gpt-5.5           # orchestrator model (thinking level via ":<level>")
#   PI_TOOLS=read,bash,edit,write,grep,find,ls,subagent
#   VALIDATOR_BIN=pi                          # Shepherd CLI (cross-model judging is a feature)
#   VALIDATOR_ARGS="--model openai-codex/gpt-5.6-sol --thinking high --tools read,bash,edit,write,grep,find,ls --approve"
#   MAX_ITERATIONS=200                        # hard backstop on orchestrator turns
#   MAX_REVERTS=6                             # total revert budget per run before HALT
#   MAX_NO_VERDICT=3                          # consecutive no-verdict turns before HALT
#   MAX_MINUTES=0                             # wall-clock cap (0 = none)
#   COOLDOWN_SECONDS=5
#   PI_EXTRA_FLAGS=""                         # extra flags for every orchestrator invocation
#   LOOP_CMD=/pm-auto-loop                    # /pm-connector-loop for connector runs
#   AUTO_LOOP_STATE_DIR=.planning/auto-loop   # set to isolate separate Shepherd runs
#   SEARXNG_BASE=                             # research via the audited searxng connector (pm)
set -euo pipefail

PI_BIN="${PI_BIN:-pi}"
ORCH_MODEL="${ORCH_MODEL:-openai-codex/gpt-5.5}"
PI_TOOLS="${PI_TOOLS:-read,bash,edit,write,grep,find,ls,subagent}"
VALIDATOR_BIN="${VALIDATOR_BIN:-pi}"
VALIDATOR_ARGS="${VALIDATOR_ARGS:---model openai-codex/gpt-5.6-sol --thinking high --tools read,bash,edit,write,grep,find,ls --approve}"
MAX_ITERATIONS="${MAX_ITERATIONS:-200}"
MAX_REVERTS="${MAX_REVERTS:-6}"
MAX_NO_VERDICT="${MAX_NO_VERDICT:-3}"
MAX_MINUTES="${MAX_MINUTES:-0}"
COOLDOWN_SECONDS="${COOLDOWN_SECONDS:-5}"
PI_EXTRA_FLAGS="${PI_EXTRA_FLAGS:-}"
LOOP_CMD="${LOOP_CMD:-/pm-auto-loop}"
# Research: default SEARXNG_BASE from the shell's SEARXNG_URL (name mismatch guard) and export.
SEARXNG_BASE="${SEARXNG_BASE:-${SEARXNG_URL:-}}"; export SEARXNG_BASE
STALL_MINUTES="${STALL_MINUTES:-20}"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATE_DIR="${AUTO_LOOP_STATE_DIR:-$REPO_ROOT/.planning/auto-loop}"
[[ "$STATE_DIR" = /* ]] || STATE_DIR="$REPO_ROOT/$STATE_DIR"
# Subagent observability: locally-patched pi-sub-agent records child sessions here (opt-in).
PI_SUBAGENT_SESSION_DIR="${PI_SUBAGENT_SESSION_DIR:-$STATE_DIR/sessions}"; export PI_SUBAGENT_SESSION_DIR
CKPT_DIR="$STATE_DIR/checkpoints"
RUN_JSON="$STATE_DIR/RUN.json"
VERDICT_JSON="$STATE_DIR/VALIDATOR-VERDICT.json"
PROMPT_FILE="$STATE_DIR/PROMPT.txt"
LOG_FILE="$STATE_DIR/driver.log"
VAL_PROMPT="$REPO_ROOT/.agents/agentic-delivery/prompts/shepherd-validator-prompt.md"
mkdir -p "$STATE_DIR" "$CKPT_DIR"

log() { printf '%s %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*" | tee -a "$LOG_FILE" >&2; }

state_root_instruction() {
  cat <<EOF
SHEPHERD STATE ROOT FOR THIS RUN:
- Use $STATE_DIR as the run ledger root.
- Read/write RUN.json, ORCHESTRATION-STATE.json, VALIDATION.jsonl, VALIDATOR-VERDICT.json, checkpoints, tasks, trace, sessions, and RESEARCH under $STATE_DIR.
- If any prompt, workflow, or reference says .planning/auto-loop, interpret that as $STATE_DIR for this run.
- Do not read or write $REPO_ROOT/.planning/auto-loop for this run unless explicitly using it as historical evidence and clearly labeling it historical.
EOF
}

# --- preflight: the subagent tool must be available (vendored extension OR installed package) ---
# We vendor pi-sub-agent under .pi/extensions/ (records child sessions via PI_SUBAGENT_SESSION_DIR),
# loaded through .pi/settings.json. Accept either the vendored extension or the npm package; fail
# only if neither is present (subagent tool silently absent → .pi/agents/* cannot be spawned).
if [[ ! -f "$REPO_ROOT/.pi/extensions/pi-sub-agent/index.ts" ]] \
   && ! "$PI_BIN" list 2>/dev/null | grep -q "pi-sub-agent"; then
  echo "FATAL: the pi 'subagent' tool is unavailable — no vendored .pi/extensions/pi-sub-agent and" >&2
  echo "no installed package, so .pi/agents/* cannot be spawned. Restore the vendored extension or" >&2
  echo "run:  $PI_BIN install npm:pi-sub-agent" >&2
  exit 2
fi

# --- resolve the problem prompt --------------------------------------------------------------
if [[ "${1:-}" == "--resume" ]]; then
  [[ -f "$PROMPT_FILE" ]] || { echo "No run to resume (missing $PROMPT_FILE)." >&2; exit 2; }
  PROBLEM="$(cat "$PROMPT_FILE")"; log "RESUME: $PROBLEM"
elif [[ -n "${1:-}" ]]; then
  PROBLEM="$*"; printf '%s' "$PROBLEM" > "$PROMPT_FILE"; log "START: $PROBLEM"
else
  echo "Usage: scripts/pi-shepherd-loop.sh \"<problem prompt>\" | --resume" >&2; exit 2
fi

json_field() { # $1=file $2=key
  [[ -f "$1" ]] || { echo ""; return 0; }
  python3 - "$1" "$2" <<'PY' 2>/dev/null || echo ""
import json,sys
try:
    d=json.load(open(sys.argv[1])); v=d.get(sys.argv[2])
    if isinstance(v,dict): v=v.get("type","")
    print("" if v is None else v)
except Exception: print("")
PY
}

latest_session_file() {
  python3 - "$STATE_DIR/sessions" <<'PY'
import os,sys
root=sys.argv[1]
best=None
if os.path.isdir(root):
    for cur, subdirs, names in os.walk(root):
        subdirs[:] = [d for d in subdirs if d not in {".git","node_modules","vendor"}]
        for name in names:
            if not name.endswith(".jsonl"):
                continue
            path=os.path.join(cur,name)
            try:
                item=(os.path.getmtime(path),path)
            except OSError:
                continue
            if best is None or item > best:
                best=item
if best:
    print(best[1])
PY
}

run_orchestrator() { # $1=turn-message — with stall watchdog (session mtime + child liveness)
  local msg="$1" rc=0 pid sess
  # shellcheck disable=SC2086
  "$PI_BIN" -p --model "$ORCH_MODEL" --tools "$PI_TOOLS" --approve --session-dir "$STATE_DIR/sessions" $PI_EXTRA_FLAGS \
    "$msg" >>"$LOG_FILE" 2>&1 & pid=$!
  while kill -0 "$pid" 2>/dev/null; do
    sleep 60
    sess="$(latest_session_file)"
    if [[ -n "$sess" ]]; then
      local age=$(( $(date +%s) - $(stat -f %m "$sess" 2>/dev/null || echo 0) ))
      if (( age > STALL_MINUTES * 60 )) && ! pgrep -P "$pid" >/dev/null 2>&1; then
        log "STALL GUARD: no session event ${age}s and no live children — killing turn pid $pid"
        kill -TERM "$pid" 2>/dev/null; sleep 5; kill -KILL "$pid" 2>/dev/null || true
        return 124
      fi
    fi
  done
  wait "$pid" 2>/dev/null || rc=$?
  return $rc
}

run_validator() {
  local rc=0
  # shellcheck disable=SC2086
  "$VALIDATOR_BIN" -p $VALIDATOR_ARGS --session-dir "$STATE_DIR/sessions" "$(state_root_instruction)

$(cat "$VAL_PROMPT")" >>"$LOG_FILE" 2>&1 || rc=$?
  return $rc
}

checkpoint() { # $1=turn — snapshot the run ledger + the worktree HEAD SHA (the fork point).
  local d="$CKPT_DIR/$1"; mkdir -p "$d"
  [[ -f "$RUN_JSON" ]] && cp "$RUN_JSON" "$d/RUN.json" 2>/dev/null || true
  git -C "$REPO_ROOT" rev-parse HEAD > "$d/HEAD.sha" 2>/dev/null || true
  echo "$1" > "$CKPT_DIR/LAST_GOOD"
}

restore_checkpoint() { # Restore the ledger to the last PROCEED checkpoint; hand the orchestrator
  # an explicit cleanup task for diverged commits (never rewrite pushed history here).
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
    log "reverted to checkpoint $last (fork ${good_sha:0:8}); wrote REVERT-CLEANUP.json for orchestrator cleanup of ${cur_sha:0:8}"
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
  turn_msg="$LOOP_CMD $PROBLEM

$(state_root_instruction)"
  if [[ -n "$correction" ]]; then
    turn_msg="$turn_msg

VALIDATOR CORRECTION (apply first): $correction"
  fi
  run_orchestrator "$turn_msg" || log "turn $i: orchestrator returned non-zero (validator will assess)"

  log "── turn $i: VALIDATOR ──"
  run_validator || log "turn $i: validator returned non-zero"
  AUTO_LOOP_STATE_DIR="$STATE_DIR" "$REPO_ROOT/scripts/loop-trace.sh" distill >/dev/null 2>&1 && log "turn $i: trace digest written (see $STATE_DIR/trace/INDEX.md)" || true

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
        log "HALT: no VALIDATOR-VERDICT.json for $no_verdict consecutive turns. Check that VALIDATOR_ARGS grants writable tools (--tools ...,edit,write --approve) and that '$VALIDATOR_BIN' is logged in, then --resume."
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
