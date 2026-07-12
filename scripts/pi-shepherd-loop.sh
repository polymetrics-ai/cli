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
#   MAX_TURNS=1000                            # durable total-turn cap across resume
#   TURN_TIMEOUT_SECONDS=5400                 # hard orchestrator+validator turn deadline
#   TERM_GRACE_SECONDS=10                     # TERM grace before process-group KILL
#   CONTROL_HEARTBEAT_SECONDS=5               # durable authority heartbeat cadence
#   MODEL_PREFLIGHT_TIMEOUT_SECONDS=30         # bounded local validator-model discovery
#   COOLDOWN_SECONDS=5
#   PI_EXTRA_FLAGS=""                         # extra flags for every orchestrator invocation
#   LOOP_CMD=/pm-auto-loop                    # /pm-connector-loop for connector runs
#   SEARXNG_BASE=                             # research via the audited searxng connector (pm)
set -euo pipefail

# AUTO_LOOP_PHASE0_GUARD_BEGIN
# AUTO_LOOP_RUN_ENTRYPOINT: scripts/pi-shepherd-loop.sh
case "${1:-}" in
  help|-h|--help)
    printf '%s\n' 'Usage: scripts/pi-shepherd-loop.sh "<problem prompt>" | --resume'
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
if auto_loop_safety_guard_driver "scripts/pi-shepherd-loop.sh" json; then
  printf '%s\n' 'AUTO_LOOP_GUARD_UNEXPECTED_SUCCESS' >&2
  exit 78
else
  safety_rc=$?
  exit "$safety_rc"
fi
# AUTO_LOOP_PHASE0_GUARD_END

PI_BIN="${PI_BIN:-pi}"
ORCH_MODEL="${ORCH_MODEL:-openai-codex/gpt-5.5}"
PI_TOOLS="${PI_TOOLS:-read,bash,edit,write,grep,find,ls,subagent}"
VALIDATOR_BIN="${VALIDATOR_BIN:-pi}"
VALIDATOR_MODEL="openai-codex/gpt-5.6-sol"
VALIDATOR_ARGS_OVERRIDDEN=0
if [[ -n "${VALIDATOR_ARGS:-}" ]]; then
  VALIDATOR_ARGS_OVERRIDDEN=1
else
  VALIDATOR_ARGS="--model $VALIDATOR_MODEL --thinking high --tools read,bash,edit,write,grep,find,ls --approve"
  # VALIDATOR_ARGS may have arrived as an exported-but-empty variable. Do not let the lock
  # acquisition re-exec mistake this internally generated default for a caller override.
  export -n VALIDATOR_ARGS
fi
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
MAX_TURNS="${MAX_TURNS:-1000}"
TURN_TIMEOUT_SECONDS="${TURN_TIMEOUT_SECONDS:-5400}"
TERM_GRACE_SECONDS="${TERM_GRACE_SECONDS:-10}"
CONTROL_HEARTBEAT_SECONDS="${CONTROL_HEARTBEAT_SECONDS:-5}"
MODEL_PREFLIGHT_TIMEOUT_SECONDS="${MODEL_PREFLIGHT_TIMEOUT_SECONDS:-30}"

if [[ ! "$MODEL_PREFLIGHT_TIMEOUT_SECONDS" =~ ^[0-9]+$ ]] || \
   (( MODEL_PREFLIGHT_TIMEOUT_SECONDS < 1 || MODEL_PREFLIGHT_TIMEOUT_SECONDS > 300 )); then
  echo "MODEL_PREFLIGHT_TIMEOUT_SECONDS must be an integer from 1 through 300" >&2
  exit 2
fi

case "${1:-}" in
  "")
    echo "Usage: scripts/pi-shepherd-loop.sh \"<problem prompt>\" | --resume" >&2
    exit 2
    ;;
esac

SCRIPT_DIR="${BASH_SOURCE[0]%/*}"
[[ "$SCRIPT_DIR" != "${BASH_SOURCE[0]}" ]] || SCRIPT_DIR="."
SCRIPT_DIR="$(cd "$SCRIPT_DIR" && pwd -P)"
SCRIPT_SELF="$SCRIPT_DIR/${BASH_SOURCE[0]##*/}"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd -P)"
STATE_DIR="$REPO_ROOT/.planning/auto-loop"
mkdir -p "$STATE_DIR"
CONTROL_JSON="$STATE_DIR/CONTROL.json"
CONTROL_LOCK="$STATE_DIR/CONTROL.lock"

# Acquire one worktree-wide advisory lock and retain its open file description across re-exec and
# all descendants. A surviving registered child therefore keeps replacement controllers fenced.
if [[ -z "${AUTO_LOOP_CONTROL_FD:-}" ]]; then
  exec python3 - "$CONTROL_LOCK" "$SCRIPT_SELF" "$@" <<'PY'
import fcntl
import os
import stat
import sys

lock_path, script, *arguments = sys.argv[1:]
flags = os.O_RDWR | os.O_CREAT
if hasattr(os, "O_NOFOLLOW"):
    flags |= os.O_NOFOLLOW
try:
    fd = os.open(lock_path, flags, 0o600)
    info = os.fstat(fd)
    if not stat.S_ISREG(info.st_mode) or info.st_nlink != 1:
        raise OSError("unsafe controller lock")
    os.fchmod(fd, 0o600)
    fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
except BlockingIOError:
    print("CONTROLLER_HELD: another Shepherd controller owns this worktree", file=sys.stderr)
    raise SystemExit(75)
except OSError as exc:
    print(f"CONTROLLER_LOCK_UNSAFE: {exc}", file=sys.stderr)
    raise SystemExit(4)

os.set_inheritable(fd, True)
environment = os.environ.copy()
environment["AUTO_LOOP_CONTROL_FD"] = str(fd)
os.execve(script, [script, *arguments], environment)
PY
fi

# Reject forged or moved inherited descriptors before touching controller state.
if ! python3 - "$AUTO_LOOP_CONTROL_FD" "$CONTROL_LOCK" <<'PY'
import fcntl
import os
import stat
import sys

try:
    fd = int(sys.argv[1])
    held = os.fstat(fd)
    path = os.stat(sys.argv[2], follow_symlinks=False)
    if not stat.S_ISREG(held.st_mode) or held.st_nlink != 1:
        raise OSError("lock descriptor is not a private regular file")
    if (held.st_dev, held.st_ino) != (path.st_dev, path.st_ino):
        raise OSError("lock descriptor does not match canonical path")
    fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
except (OSError, ValueError) as exc:
    print(f"CONTROL_FENCE_INVALID: {exc}", file=sys.stderr)
    raise SystemExit(4)
PY
then
  exit 4
fi

# Subagent observability: locally-patched pi-sub-agent records child sessions here (opt-in).
PI_SUBAGENT_SESSION_DIR="${PI_SUBAGENT_SESSION_DIR:-$STATE_DIR/sessions}"; export PI_SUBAGENT_SESSION_DIR
CKPT_DIR="$STATE_DIR/checkpoints"
RUN_JSON="$STATE_DIR/RUN.json"
VERDICT_JSON="$STATE_DIR/VALIDATOR-VERDICT.json"
PROMPT_FILE="$STATE_DIR/PROMPT.txt"
LOG_FILE="$STATE_DIR/driver.log"
VAL_PROMPT="$REPO_ROOT/.agents/agentic-delivery/prompts/shepherd-validator-prompt.md"
mkdir -p "$CKPT_DIR"

log() { printf '%s %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*" | tee -a "$LOG_FILE" >&2; }

control_state() { # command [arguments...]
  local command="$1"
  shift
  python3 - "$CONTROL_JSON" "$command" "${CONTROL_FENCE:-}" "$@" <<'PY'
import datetime as dt
import json
import os
import pathlib
import stat
import sys
import tempfile
import uuid

path = pathlib.Path(sys.argv[1])
command = sys.argv[2]
fence_text = sys.argv[3]
args = sys.argv[4:]
UTC = dt.timezone.utc
source_present = False
source_bytes = None

def now():
    return dt.datetime.now(UTC)

def timestamp(value=None):
    return (value or now()).isoformat(timespec="seconds").replace("+00:00", "Z")

def fail(code, status=4):
    print(code, file=sys.stderr)
    raise SystemExit(status)

def read_state(required=True):
    global source_present, source_bytes
    if not os.path.lexists(path):
        if required:
            fail("CONTROL_STATE_MISSING")
        return None
    flags = os.O_RDONLY
    if hasattr(os, "O_NOFOLLOW"):
        flags |= os.O_NOFOLLOW
    try:
        fd = os.open(path, flags)
        info = os.fstat(fd)
        if not stat.S_ISREG(info.st_mode) or info.st_nlink != 1:
            fail("CONTROL_STATE_UNSAFE")
        if info.st_size > 131072:
            fail("CONTROL_STATE_OVERSIZED")
        with os.fdopen(fd, "rb") as handle:
            raw = handle.read()
        value = json.loads(raw.decode("utf-8"))
    except (OSError, UnicodeError, json.JSONDecodeError):
        fail("CONTROL_STATE_INVALID")
    if not isinstance(value, dict) or value.get("schema_version") != "1.0":
        fail("CONTROL_STATE_INVALID")
    source_present = True
    source_bytes = raw
    return value

def write_state(value):
    global source_present, source_bytes
    value["updated_at"] = timestamp()
    payload = (json.dumps(value, sort_keys=True, separators=(",", ":")) + "\n").encode()
    path.parent.mkdir(parents=True, exist_ok=True)
    if source_present:
        flags = os.O_RDONLY
        if hasattr(os, "O_NOFOLLOW"):
            flags |= os.O_NOFOLLOW
        try:
            current_fd = os.open(path, flags)
            with os.fdopen(current_fd, "rb") as current:
                if current.read() != source_bytes:
                    fail("CONTROL_STATE_MOVED")
        except OSError:
            fail("CONTROL_STATE_MOVED")
    elif os.path.lexists(path):
        fail("CONTROL_STATE_MOVED")
    fd, temporary = tempfile.mkstemp(prefix=".CONTROL.", dir=path.parent)
    try:
        os.fchmod(fd, 0o600)
        offset = 0
        while offset < len(payload):
            offset += os.write(fd, payload[offset:])
        os.fsync(fd)
        os.close(fd)
        fd = -1
        os.replace(temporary, path)
        directory = os.open(path.parent, os.O_RDONLY | getattr(os, "O_DIRECTORY", 0))
        try:
            os.fsync(directory)
        finally:
            os.close(directory)
        source_present = True
        source_bytes = payload
    except Exception:
        if fd >= 0:
            os.close(fd)
        try:
            os.unlink(temporary)
        except FileNotFoundError:
            pass
        fail("CONTROL_COMMIT_UNCERTAIN")

def parse_positive(name, value, minimum=1, maximum=86400, status=2, strict=False):
    if strict and (not isinstance(value, int) or isinstance(value, bool)):
        fail("CONTROL_STATE_INVALID", status)
    try:
        parsed = int(value)
    except (TypeError, ValueError):
        fail(f"CONTROL_CONFIG_INVALID:{name}", status)
    if parsed < minimum or parsed > maximum:
        fail(f"CONTROL_CONFIG_INVALID:{name}", status)
    return parsed

def state_integer(name, value, minimum, maximum):
    if not isinstance(value, int) or isinstance(value, bool) or value < minimum or value > maximum:
        fail(f"CONTROL_STATE_INVALID:{name}")
    return value

def validated_limits(value):
    limits = value.get("limits")
    counters = value.get("counters")
    if not isinstance(limits, dict) or not isinstance(counters, dict):
        fail("CONTROL_STATE_INVALID")
    result = {
        "max_turns": parse_positive("max_turns", limits.get("max_turns"), maximum=1000000, status=4, strict=True),
        "turn_timeout_seconds": parse_positive("turn_timeout_seconds", limits.get("turn_timeout_seconds"), maximum=604800, status=4, strict=True),
        "term_grace_seconds": parse_positive("term_grace_seconds", limits.get("term_grace_seconds"), maximum=300, status=4, strict=True),
        "heartbeat_seconds": parse_positive("heartbeat_seconds", limits.get("heartbeat_seconds"), maximum=300, status=4, strict=True),
        "max_reverts": parse_positive("max_reverts", limits.get("max_reverts"), maximum=1000000, status=4, strict=True),
        "max_no_verdict": parse_positive("max_no_verdict", limits.get("max_no_verdict"), maximum=1000000, status=4, strict=True),
        "max_minutes": parse_positive("max_minutes", limits.get("max_minutes"), minimum=0, maximum=525600, status=4, strict=True),
    }
    for key in ("reverts", "no_verdict", "active_seconds"):
        item = counters.get(key)
        if not isinstance(item, int) or isinstance(item, bool) or item < 0:
            fail("CONTROL_STATE_INVALID")
    return result, counters

def fence_of(value):
    return "|".join(str(value.get(key, "")) for key in (
        "run_id", "generation", "controller_id", "control_revision"
    ))

def exact_state(allowed=("active",)):
    value = read_state()
    if not fence_text or fence_of(value) != fence_text:
        fail("CONTROL_FENCE_MISMATCH")
    if value.get("phase") not in allowed:
        fail("CONTROL_PHASE_MISMATCH")
    return value

def refreshed_lease(value):
    seconds = int(value["limits"]["heartbeat_seconds"]) * 3
    current = now()
    value["lease"] = {
        "heartbeat_at": timestamp(current),
        "expires_at": timestamp(current + dt.timedelta(seconds=seconds)),
    }

if command == "init":
    (mode, max_turns_raw, timeout_raw, grace_raw, heartbeat_raw,
     max_reverts_raw, max_no_verdict_raw, max_minutes_raw) = args
    max_turns = parse_positive("max_turns", max_turns_raw, maximum=1000000)
    timeout = parse_positive("turn_timeout_seconds", timeout_raw, maximum=604800)
    grace = parse_positive("term_grace_seconds", grace_raw, maximum=300)
    heartbeat = parse_positive("heartbeat_seconds", heartbeat_raw, maximum=300)
    max_reverts = parse_positive("max_reverts", max_reverts_raw, maximum=1000000)
    max_no_verdict = parse_positive("max_no_verdict", max_no_verdict_raw, maximum=1000000)
    max_minutes = parse_positive("max_minutes", max_minutes_raw, minimum=0, maximum=525600)
    previous = read_state(required=False)

    if previous is None:
        if mode == "resume":
            fail("NO_RUN_TO_RESUME", 2)
        generation = revision = 1
        run_id = str(uuid.uuid4())
        ordinal = 0
        limits = {
            "max_turns": max_turns,
            "turn_timeout_seconds": timeout,
            "term_grace_seconds": grace,
            "heartbeat_seconds": heartbeat,
            "max_reverts": max_reverts,
            "max_no_verdict": max_no_verdict,
            "max_minutes": max_minutes,
        }
        counters = {"reverts": 0, "no_verdict": 0, "active_seconds": 0}
    else:
        try:
            phase = previous["phase"]
            generation = state_integer("generation", previous["generation"], 1, 9223372036854775807)
            revision = state_integer("control_revision", previous["control_revision"], 1, 9223372036854775807)
        except KeyError:
            fail("CONTROL_STATE_INVALID")
        limits, counters = validated_limits(previous)
        ordinal = state_integer("turn_ordinal", previous.get("turn_ordinal"), 0, limits["max_turns"])
        if mode == "resume":
            if phase == "halted":
                fail("HALT_LATCHED")
            if phase == "recovery_required":
                fail("RECOVERY_REQUIRED")
            if phase in ("active", "halting"):
                fail("RECOVERY_REQUIRED")
            if phase != "paused":
                fail("CONTROL_STATE_BLOCKED")
            if previous.get("active_turn") is not None or previous.get("children_quiescent") is not True or previous.get("halt") is not None:
                fail("RECOVERY_REQUIRED")
            if ordinal >= int(limits.get("max_turns", 0)):
                fail("TURN_CAP_REACHED", 3)
            run_id = previous["run_id"]
        else:
            if phase != "released":
                if phase == "halted":
                    fail("HALT_LATCHED")
                if phase == "recovery_required":
                    fail("RECOVERY_REQUIRED")
                if phase in ("active", "halting"):
                    fail("RECOVERY_REQUIRED")
                fail("CONTROL_STATE_BLOCKED")
            if previous.get("active_turn") is not None or previous.get("children_quiescent") is not True:
                fail("RECOVERY_REQUIRED")
            run_id = str(uuid.uuid4())
            ordinal = 0
            limits = {
                "max_turns": max_turns,
                "turn_timeout_seconds": timeout,
                "term_grace_seconds": grace,
                "heartbeat_seconds": heartbeat,
                "max_reverts": max_reverts,
                "max_no_verdict": max_no_verdict,
                "max_minutes": max_minutes,
            }
            counters = {"reverts": 0, "no_verdict": 0, "active_seconds": 0}
        generation += 1
        revision += 1
        if generation > 9223372036854775807 or revision > 9223372036854775807:
            fail("CONTROL_GENERATION_EXHAUSTED")

    value = {
        "schema_version": "1.0",
        "run_id": run_id,
        "generation": generation,
        "controller_id": str(uuid.uuid4()),
        "control_revision": revision,
        "phase": "active",
        "lease": {},
        "limits": limits,
        "turn_ordinal": ordinal,
        "counters": counters,
        "active_turn": None,
        "halt": None,
        "children_quiescent": True,
        "updated_at": timestamp(),
    }
    refreshed_lease(value)
    write_state(value)
    print("\t".join(str(item) for item in (
        fence_of(value),
        value["limits"]["max_turns"],
        value["limits"]["turn_timeout_seconds"],
        value["limits"]["term_grace_seconds"],
        value["limits"]["heartbeat_seconds"],
        value["limits"]["max_reverts"],
        value["limits"]["max_no_verdict"],
        value["limits"]["max_minutes"],
        value["counters"]["reverts"],
        value["counters"]["no_verdict"],
        value["counters"]["active_seconds"],
    )))
elif command == "assert":
    exact_state()
elif command == "heartbeat":
    value = exact_state()
    refreshed_lease(value)
    write_state(value)
elif command == "begin":
    value = exact_state()
    if value.get("active_turn") is not None:
        fail("CONTROL_TURN_ACTIVE")
    ordinal = int(value["turn_ordinal"])
    maximum = int(value["limits"]["max_turns"])
    if ordinal >= maximum:
        fail("TURN_CAP_REACHED", 3)
    ordinal += 1
    current = now()
    turn = {
        "turn_id": str(uuid.uuid4()),
        "ordinal": ordinal,
        "deadline_at": timestamp(current + dt.timedelta(seconds=int(value["limits"]["turn_timeout_seconds"]))),
        "orchestrator_session_id": str(uuid.uuid4()),
        "validator_session_id": str(uuid.uuid4()),
        "active_role": None,
        "leader_pid": None,
        "process_group_id": None,
    }
    value["turn_ordinal"] = ordinal
    value["active_turn"] = turn
    value["children_quiescent"] = True
    refreshed_lease(value)
    write_state(value)
    print("\t".join(str(turn[key]) for key in (
        "ordinal", "turn_id", "orchestrator_session_id", "validator_session_id"
    )))
elif command == "bind":
    role, pid_raw, pgid_raw = args
    if role not in ("orchestrator", "validator"):
        fail("CONTROL_ROLE_INVALID")
    pid = parse_positive("leader_pid", pid_raw, maximum=2147483647)
    pgid = parse_positive("process_group_id", pgid_raw, maximum=2147483647)
    value = exact_state()
    turn = value.get("active_turn")
    if not isinstance(turn, dict) or turn.get("active_role") is not None:
        fail("CONTROL_TURN_INVALID")
    turn.update({"active_role": role, "leader_pid": pid, "process_group_id": pgid})
    value["children_quiescent"] = False
    refreshed_lease(value)
    write_state(value)
elif command == "clear-role":
    role = args[0]
    value = exact_state()
    turn = value.get("active_turn")
    if not isinstance(turn, dict) or turn.get("active_role") != role:
        fail("CONTROL_ROLE_MISMATCH")
    turn.update({"active_role": None, "leader_pid": None, "process_group_id": None})
    value["children_quiescent"] = True
    refreshed_lease(value)
    write_state(value)
elif command == "complete":
    value = exact_state()
    if not value.get("children_quiescent") or not isinstance(value.get("active_turn"), dict):
        fail("CONTROL_TURN_INVALID")
    value["active_turn"] = None
    refreshed_lease(value)
    write_state(value)
elif command == "counter":
    name, raw = args
    if name not in ("reverts", "no_verdict"):
        fail("CONTROL_COUNTER_INVALID")
    count = parse_positive(name, raw, minimum=0, maximum=1000000)
    value = exact_state()
    _, counters = validated_limits(value)
    counters[name] = count
    refreshed_lease(value)
    write_state(value)
elif command == "latch":
    phase, code, reason = args
    if phase not in ("halting", "recovery_required"):
        fail("CONTROL_PHASE_INVALID")
    value = exact_state(("active", "halting", "recovery_required"))
    value["phase"] = phase
    value["halt"] = {
        "halt_id": str(uuid.uuid4()),
        "code": code,
        "reason": reason[:2048],
        "created_at": timestamp(),
    }
    value["children_quiescent"] = False
    write_state(value)
elif command == "finalize":
    phase, quiescent_raw = args
    if phase not in ("halted", "recovery_required"):
        fail("CONTROL_PHASE_INVALID")
    value = exact_state(("active", "halting", "recovery_required"))
    quiescent = quiescent_raw == "true"
    value["phase"] = phase
    value["children_quiescent"] = quiescent
    turn = value.get("active_turn")
    if quiescent and isinstance(turn, dict):
        turn.update({"active_role": None, "leader_pid": None, "process_group_id": None})
    write_state(value)
elif command in ("pause", "release"):
    active_seconds = parse_positive("active_seconds", args[0], minimum=0, maximum=315360000)
    value = exact_state()
    if not value.get("children_quiescent") or value.get("active_turn") is not None:
        fail("CONTROL_NOT_QUIESCENT")
    value["phase"] = "paused" if command == "pause" else "released"
    _, counters = validated_limits(value)
    counters["active_seconds"] = max(counters["active_seconds"], active_seconds)
    write_state(value)
else:
    fail("CONTROL_COMMAND_INVALID", 2)
PY
}

CONTROL_FENCE=""
CONTROL_CLOSED=0
AUTHORITY_ACQUIRED=0
ACTIVE_ROLE_PID=""
ACTIVE_ROLE_PGID=""
ACTIVE_ROLE_GROUP_READY=0
ACTIVE_ROLE_READY_FILE=""
ACTIVE_ROLE_START_FILE=""
ACTIVE_ROLE_OUTPUT_FILE=""
TURN_DEADLINE_MONO=0
ACTIVE_SECONDS_BASE=0
ACTIVE_START_SECONDS=$SECONDS

controller_startup_recovery() { # code reason
  local code="$1" reason="$2"
  if (( AUTHORITY_ACQUIRED == 1 && CONTROL_CLOSED == 0 )); then
    if control_state latch recovery_required "$code" "$reason" >/dev/null 2>&1; then
      control_state finalize recovery_required true >/dev/null 2>&1 || true
    fi
    CONTROL_CLOSED=1
  fi
}

controller_startup_signal() {
  trap - INT TERM HUP
  controller_startup_recovery CONTROLLER_STARTUP_SIGNAL "controller received a signal during startup"
  exit 4
}

controller_startup_exit() {
  local rc=$?
  trap - EXIT
  controller_startup_recovery CONTROLLER_STARTUP_EXIT "controller exited before supervision became ready"
  exit "$rc"
}

trap controller_startup_signal INT TERM HUP
trap controller_startup_exit EXIT

mode="start"
[[ "${1:-}" == "--resume" ]] && mode="resume"
if control_output="$(control_state init "$mode" "$MAX_TURNS" "$TURN_TIMEOUT_SECONDS" "$TERM_GRACE_SECONDS" "$CONTROL_HEARTBEAT_SECONDS" "$MAX_REVERTS" "$MAX_NO_VERDICT" "$MAX_MINUTES")"; then
  IFS=$'\t' read -r CONTROL_FENCE MAX_TURNS TURN_TIMEOUT_SECONDS TERM_GRACE_SECONDS CONTROL_HEARTBEAT_SECONDS \
    MAX_REVERTS MAX_NO_VERDICT MAX_MINUTES reverts no_verdict ACTIVE_SECONDS_BASE <<<"$control_output"
  AUTHORITY_ACQUIRED=1
  ACTIVE_START_SECONDS=$SECONDS
else
  control_rc=$?
  exit "$control_rc"
fi

process_group_alive() {
  local pgid="$1"
  [[ "$pgid" =~ ^[0-9]+$ && "$pgid" -gt 1 ]] || return 1
  kill -0 -- "-$pgid" 2>/dev/null
}

drain_role_group() {
  local pid="${ACTIVE_ROLE_PID:-}" pgid="${ACTIVE_ROLE_PGID:-}" started observed_pgid=""
  [[ "$pid" =~ ^[0-9]+$ ]] || return 0
  if kill -0 "$pid" 2>/dev/null; then
    observed_pgid="$(ps -o pgid= -p "$pid" 2>/dev/null | tr -d '[:space:]' || true)"
  fi
  if (( ACTIVE_ROLE_GROUP_READY == 1 )); then
    if [[ -n "$observed_pgid" && "$observed_pgid" != "$pgid" ]]; then
      kill -TERM "$pid" 2>/dev/null || true
      wait "$pid" 2>/dev/null || true
      return 1
    fi
  elif [[ "$observed_pgid" == "$pid" ]]; then
    pgid="$pid"
    ACTIVE_ROLE_PGID="$pgid"
    ACTIVE_ROLE_GROUP_READY=1
  else
    kill -TERM "$pid" 2>/dev/null || true
    started=$SECONDS
    while kill -0 "$pid" 2>/dev/null && (( SECONDS - started < TERM_GRACE_SECONDS )); do
      sleep 0.1
    done
    kill -KILL "$pid" 2>/dev/null || true
    wait "$pid" 2>/dev/null || true
    ACTIVE_ROLE_PID=""
    ACTIVE_ROLE_PGID=""
    ACTIVE_ROLE_GROUP_READY=0
    return 0
  fi
  [[ "$pgid" =~ ^[0-9]+$ && "$pgid" -gt 1 ]] || return 1
  if process_group_alive "$pgid"; then
    kill -TERM -- "-$pgid" 2>/dev/null || true
    started=$SECONDS
    while process_group_alive "$pgid" && (( SECONDS - started < TERM_GRACE_SECONDS )); do
      sleep 0.1
    done
  fi
  if process_group_alive "$pgid"; then
    kill -KILL -- "-$pgid" 2>/dev/null || true
  fi
  if [[ "$pid" =~ ^[0-9]+$ ]]; then
    wait "$pid" 2>/dev/null || true
  fi
  started=$SECONDS
  while process_group_alive "$pgid" && (( SECONDS - started < 2 )); do
    sleep 0.1
  done
  if process_group_alive "$pgid"; then
    return 1
  fi
  ACTIVE_ROLE_PID=""
  ACTIVE_ROLE_PGID=""
  ACTIVE_ROLE_GROUP_READY=0
  return 0
}

cleanup_role_artifacts() {
  [[ -z "$ACTIVE_ROLE_READY_FILE" ]] || rm -f "$ACTIVE_ROLE_READY_FILE" 2>/dev/null || true
  [[ -z "$ACTIVE_ROLE_START_FILE" ]] || rm -f "$ACTIVE_ROLE_START_FILE" 2>/dev/null || true
  [[ -z "$ACTIVE_ROLE_OUTPUT_FILE" ]] || rm -f "$ACTIVE_ROLE_OUTPUT_FILE" 2>/dev/null || true
  ACTIVE_ROLE_READY_FILE=""
  ACTIVE_ROLE_START_FILE=""
  ACTIVE_ROLE_OUTPUT_FILE=""
}

controller_emergency() { # code reason final-phase
  local code="$1" reason="$2" final_phase="$3" latch_phase="halting" latch_ok=0 drain_ok=0
  [[ "$final_phase" == "recovery_required" ]] && latch_phase="recovery_required"
  printf '%s: %s\n' "$code" "$reason" >&2
  if control_state latch "$latch_phase" "$code" "$reason" >/dev/null 2>&1; then
    latch_ok=1
  fi
  if drain_role_group; then
    drain_ok=1
  fi
  cleanup_role_artifacts
  if (( latch_ok == 1 && drain_ok == 1 )); then
    control_state finalize "$final_phase" true >/dev/null 2>&1 || true
  elif (( latch_ok == 1 )); then
    if control_state latch recovery_required TEARDOWN_NOT_QUIESCENT "registered process group could not be proven quiescent" >/dev/null 2>&1; then
      control_state finalize recovery_required false >/dev/null 2>&1 || true
    fi
  elif (( drain_ok == 1 )); then
    if control_state latch recovery_required HALT_PERSISTENCE_FAILED "the requested halt could not be durably latched" >/dev/null 2>&1; then
      control_state finalize recovery_required true >/dev/null 2>&1 || true
    fi
  fi
  CONTROL_CLOSED=1
}

controller_on_signal() {
  trap - INT TERM HUP
  controller_emergency CONTROLLER_SIGNAL "controller received a termination signal" recovery_required
  exit 4
}

controller_on_exit() {
  local rc=$?
  trap - EXIT
  if (( CONTROL_CLOSED == 0 )); then
    controller_emergency CONTROLLER_UNCLEAN_EXIT "controller exited without a safe terminal transition" recovery_required
  fi
  exit "$rc"
}

trap controller_on_signal INT TERM HUP
trap controller_on_exit EXIT

validator_model_available() {
  local ready_file output_file pid pgid attempt rc=0 deadline next_heartbeat found=1
  ready_file="$STATE_DIR/.model-ready.$$.$RANDOM"
  output_file="$STATE_DIR/.model-list.$$.$RANDOM"
  ACTIVE_ROLE_READY_FILE="$ready_file"
  ACTIVE_ROLE_OUTPUT_FILE="$output_file"
  python3 - "$ready_file" "$output_file" "$VALIDATOR_BIN" --offline --list-models "$VALIDATOR_MODEL" <<'PY' &
import os
import sys

ready, output, command, *arguments = sys.argv[1:]
os.setsid()
output_flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL
if hasattr(os, "O_NOFOLLOW"):
    output_flags |= os.O_NOFOLLOW
output_fd = os.open(output, output_flags, 0o600)
os.dup2(output_fd, 1)
os.close(output_fd)
null_fd = os.open(os.devnull, os.O_WRONLY)
os.dup2(null_fd, 2)
os.close(null_fd)
ready_fd = os.open(ready, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
try:
    os.write(ready_fd, f"{os.getpid()}\n".encode())
    os.fsync(ready_fd)
finally:
    os.close(ready_fd)
os.execvp(command, [command, *arguments])
PY
  pid=$!
  ACTIVE_ROLE_PID="$pid"
  ACTIVE_ROLE_PGID="$pid"
  ACTIVE_ROLE_GROUP_READY=0
  for ((attempt=0; attempt<1000; attempt++)); do
    [[ -s "$ready_file" ]] && break
    kill -0 "$pid" 2>/dev/null || break
    sleep 0.01
  done
  pgid="$(cat "$ready_file" 2>/dev/null || true)"
  rm -f "$ready_file"
  ACTIVE_ROLE_READY_FILE=""
  if [[ ! "$pgid" =~ ^[0-9]+$ || "$pgid" -ne "$pid" ]]; then
    controller_emergency VALIDATOR_MODEL_PREFLIGHT_FAILED "could not establish validator-model discovery group" recovery_required
    return 126
  fi
  ACTIVE_ROLE_PGID="$pgid"
  ACTIVE_ROLE_GROUP_READY=1
  deadline=$((SECONDS + MODEL_PREFLIGHT_TIMEOUT_SECONDS))
  next_heartbeat=$((SECONDS + CONTROL_HEARTBEAT_SECONDS))
  while process_group_alive "$pgid"; do
    if ! kill -0 "$pid" 2>/dev/null; then
      wait "$pid" 2>/dev/null || rc=$?
      if process_group_alive "$pgid"; then
        controller_emergency VALIDATOR_MODEL_PREFLIGHT_ORPHAN "validator-model discovery left live descendants" recovery_required
        return 125
      fi
      break
    fi
    if (( SECONDS >= deadline )); then
      controller_emergency VALIDATOR_MODEL_PREFLIGHT_TIMEOUT "validator-model discovery exceeded its deadline" recovery_required
      return 124
    fi
    if (( SECONDS >= next_heartbeat )); then
      if ! control_state heartbeat >/dev/null; then
        controller_emergency CONTROL_FENCE_MISMATCH "validator-model discovery lost controller authority" recovery_required
        return 126
      fi
      next_heartbeat=$((SECONDS + CONTROL_HEARTBEAT_SECONDS))
    fi
    sleep 0.1
  done
  if [[ -n "$ACTIVE_ROLE_PID" ]]; then
    wait "$ACTIVE_ROLE_PID" 2>/dev/null || rc=$?
  fi
  ACTIVE_ROLE_PID=""
  ACTIVE_ROLE_PGID=""
  ACTIVE_ROLE_GROUP_READY=0
  if (( rc == 0 )) && awk '$1 == "openai-codex" && $2 == "gpt-5.6-sol" { found=1 } END { exit !found }' "$output_file"; then
    found=0
  fi
  cleanup_role_artifacts
  return "$found"
}

# The default Shepherd model must exist after authority acquisition and startup-safe traps, but
# before prompt reads or the first orchestrator mutation. Explicit VALIDATOR_ARGS remain an expert
# override.
if (( VALIDATOR_ARGS_OVERRIDDEN == 0 )); then
  preflight_rc=0
  validator_model_available || preflight_rc=$?
  if (( preflight_rc != 0 )); then
    if (( preflight_rc >= 124 && preflight_rc <= 126 )); then
      exit 4
    fi
    printf 'FATAL: Shepherd requires %s with high reasoning; upgrade Pi to >=0.80.6 or provide a validated VALIDATOR_ARGS override.\n' \
      "$VALIDATOR_MODEL" >&2
    controller_emergency VALIDATOR_MODEL_UNAVAILABLE "required validator model unavailable" recovery_required
    exit 2
  fi
fi

controller_active_seconds() {
  printf '%s\n' "$((ACTIVE_SECONDS_BASE + SECONDS - ACTIVE_START_SECONDS))"
}

controller_pause() {
  control_state pause "$(controller_active_seconds)" >/dev/null
  CONTROL_CLOSED=1
}

controller_release() {
  control_state release "$(controller_active_seconds)" >/dev/null
  CONTROL_CLOSED=1
}

controller_set_counter() { # name value
  control_state counter "$1" "$2" >/dev/null
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

start_role_group() { # role command [arguments...]
  local role="$1" ready_file start_file pid pgid attempt
  shift
  ready_file="$STATE_DIR/.role-ready.$TURN_ID.$role"
  start_file="$STATE_DIR/.role-start.$TURN_ID.$role"
  ACTIVE_ROLE_READY_FILE="$ready_file"
  ACTIVE_ROLE_START_FILE="$start_file"
  python3 - "$ready_file" "$start_file" "$TURN_TIMEOUT_SECONDS" "$@" >>"$LOG_FILE" 2>&1 <<'PY' &
import os
import sys
import time

ready, start, timeout_raw, command, *arguments = sys.argv[1:]
parent_pid = os.getppid()
timeout = max(1, int(timeout_raw))
os.setsid()
fd = os.open(ready, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
try:
    os.write(fd, f"{os.getpid()}\n".encode())
    os.fsync(fd)
finally:
    os.close(fd)

# Stay inert until the controller has durably bound this exact PID/PGID. If the controller dies
# before authorization, exit without ever executing the mutating role command.
deadline = time.monotonic() + timeout
while True:
    flags = os.O_RDONLY
    if hasattr(os, "O_NOFOLLOW"):
        flags |= os.O_NOFOLLOW
    try:
        start_fd = os.open(start, flags)
    except FileNotFoundError:
        if os.getppid() != parent_pid or time.monotonic() >= deadline:
            raise SystemExit(125)
        time.sleep(0.01)
        continue
    try:
        token = os.read(start_fd, 128).decode("ascii", "strict").strip()
    finally:
        os.close(start_fd)
    if token != str(os.getpid()):
        raise SystemExit(125)
    os.unlink(start)
    break
os.execvp(command, [command, *arguments])
PY
  pid=$!
  ACTIVE_ROLE_PID="$pid"
  ACTIVE_ROLE_PGID="$pid"
  ACTIVE_ROLE_GROUP_READY=0
  for ((attempt=0; attempt<200; attempt++)); do
    [[ -s "$ready_file" ]] && break
    kill -0 "$pid" 2>/dev/null || break
    sleep 0.01
  done
  pgid="$(cat "$ready_file" 2>/dev/null || true)"
  rm -f "$ready_file"
  ACTIVE_ROLE_READY_FILE=""
  if [[ ! "$pgid" =~ ^[0-9]+$ || "$pgid" -ne "$pid" ]]; then
    controller_emergency ROLE_START_FAILED "could not establish $role process group" recovery_required
    return 126
  fi
  ACTIVE_ROLE_PGID="$pgid"
  ACTIVE_ROLE_GROUP_READY=1
  if ! control_state bind "$role" "$pid" "$pgid" >/dev/null; then
    controller_emergency CONTROL_FENCE_MISMATCH "could not bind $role process group" recovery_required
    return 126
  fi
  if (( SECONDS >= TURN_DEADLINE_MONO )); then
    controller_emergency TURN_DEADLINE "$role exhausted the persisted turn deadline before authorization" halted
    return 124
  fi
  if ! (umask 077; set -o noclobber; printf '%s\n' "$pid" >"$start_file") 2>/dev/null; then
    controller_emergency ROLE_START_FAILED "could not authorize bound $role process group" recovery_required
    return 126
  fi
  return 0
}

supervise_role_group() { # role
  local role="$1" role_code rc=0 next_heartbeat=$((SECONDS + CONTROL_HEARTBEAT_SECONDS))
  case "$role" in
    orchestrator) role_code="ORCHESTRATOR" ;;
    validator) role_code="VALIDATOR" ;;
    *) role_code="ROLE" ;;
  esac
  while process_group_alive "$ACTIVE_ROLE_PGID"; do
    if ! kill -0 "$ACTIVE_ROLE_PID" 2>/dev/null; then
      wait "$ACTIVE_ROLE_PID" 2>/dev/null || rc=$?
      if process_group_alive "$ACTIVE_ROLE_PGID"; then
        controller_emergency "${role_code}_ORPHAN" "$role leader exited with live descendants" halted
        return 125
      fi
      break
    fi
    if (( SECONDS >= TURN_DEADLINE_MONO )); then
      controller_emergency TURN_DEADLINE "$role exceeded the persisted turn deadline" halted
      return 124
    fi
    if (( SECONDS >= next_heartbeat )); then
      if ! control_state heartbeat >/dev/null; then
        controller_emergency CONTROL_FENCE_MISMATCH "$role lost controller authority" recovery_required
        return 126
      fi
      next_heartbeat=$((SECONDS + CONTROL_HEARTBEAT_SECONDS))
    fi
    sleep 0.1
  done
  if [[ -n "$ACTIVE_ROLE_PID" ]]; then
    wait "$ACTIVE_ROLE_PID" 2>/dev/null || rc=$?
  fi
  if process_group_alive "$ACTIVE_ROLE_PGID"; then
    controller_emergency "${role_code}_ORPHAN" "$role group remained live after leader completion" halted
    return 125
  fi
  ACTIVE_ROLE_PID=""
  ACTIVE_ROLE_PGID=""
  ACTIVE_ROLE_GROUP_READY=0
  if ! control_state clear-role "$role" >/dev/null; then
    controller_emergency CONTROL_FENCE_MISMATCH "could not clear $role authority" recovery_required
    return 126
  fi
  cleanup_role_artifacts
  return "$rc"
}

run_orchestrator() { # message
  local message="$1" session_dir="$STATE_DIR/sessions/turn-$GLOBAL_TURN-$TURN_ID/orchestrator-$ORCHESTRATOR_SESSION_ID"
  mkdir -p "$session_dir"
  # shellcheck disable=SC2086
  start_role_group orchestrator "$PI_BIN" -p --model "$ORCH_MODEL" --tools "$PI_TOOLS" --approve \
    --session-dir "$session_dir" $PI_EXTRA_FLAGS "$message" || return $?
  supervise_role_group orchestrator
}

run_validator() {
  local session_dir="$STATE_DIR/sessions/turn-$GLOBAL_TURN-$TURN_ID/validator-$VALIDATOR_SESSION_ID"
  mkdir -p "$session_dir"
  # shellcheck disable=SC2086
  start_role_group validator "$VALIDATOR_BIN" -p $VALIDATOR_ARGS --session-dir "$session_dir" \
    "$(cat "$VAL_PROMPT")" || return $?
  supervise_role_group validator
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

correction=""
for (( i=1; i<=MAX_ITERATIONS; i++ )); do
  if (( MAX_MINUTES > 0 )) && (( $(controller_active_seconds) >= MAX_MINUTES * 60 )); then
    controller_pause
    log "STOP: durable active-time cap ${MAX_MINUTES}m reached."; exit 3
  fi

  if turn_output="$(control_state begin)"; then
    IFS=$'\t' read -r GLOBAL_TURN TURN_ID ORCHESTRATOR_SESSION_ID VALIDATOR_SESSION_ID <<<"$turn_output"
  else
    turn_rc=$?
    if (( turn_rc == 3 )); then
      controller_pause
      log "STOP: persisted turn cap reached; a later human-authority phase must close or extend this run."
      exit 3
    fi
    controller_emergency CONTROL_BEGIN_FAILED "could not allocate a fenced turn" recovery_required
    exit 4
  fi
  TURN_DEADLINE_MONO=$((SECONDS + TURN_TIMEOUT_SECONDS))

  log "── turn $GLOBAL_TURN: ORCHESTRATOR ──${correction:+ (with correction)}"
  turn_msg="$LOOP_CMD $PROBLEM"
  if [[ -n "$correction" ]]; then
    turn_msg="$turn_msg

VALIDATOR CORRECTION (apply first): $correction"
  fi
  if run_orchestrator "$turn_msg"; then
    orchestrator_rc=0
  else
    orchestrator_rc=$?
    if (( orchestrator_rc >= 124 && orchestrator_rc <= 127 )); then
      exit 4
    fi
    log "turn $GLOBAL_TURN: orchestrator returned non-zero (validator will assess)"
  fi

  log "── turn $GLOBAL_TURN: VALIDATOR ──"
  # A validator must author a verdict for this turn. Never allow a missing/crashed validator to
  # replay the prior turn's shared-file result; full fence-bound verdict transactions follow in
  # the dedicated validator phase while the production fuse remains closed.
  if ! rm -f -- "$VERDICT_JSON" || [[ -e "$VERDICT_JSON" || -L "$VERDICT_JSON" ]]; then
    controller_emergency VERDICT_SLOT_UNSAFE "could not establish an empty validator result slot" recovery_required
    exit 4
  fi
  if run_validator; then
    validator_rc=0
  else
    validator_rc=$?
    if (( validator_rc >= 124 && validator_rc <= 127 )); then
      exit 4
    fi
    log "turn $GLOBAL_TURN: validator returned non-zero"
  fi
  "$REPO_ROOT/scripts/loop-trace.sh" distill >/dev/null 2>&1 && log "turn $GLOBAL_TURN: trace digest written (see .planning/auto-loop/trace/INDEX.md)" || true

  verdict="$(json_field "$VERDICT_JSON" verdict)"
  score="$(json_field "$VERDICT_JSON" step_score)"
  reason="$(json_field "$VERDICT_JSON" reason)"
  correction=""
  log "turn $GLOBAL_TURN: verdict=${verdict:-NONE} step_score=${score:-?} — ${reason:-}"

  case "$verdict" in
    PROCEED)
      no_verdict=0; controller_set_counter no_verdict "$no_verdict"
      control_state complete >/dev/null; checkpoint "$GLOBAL_TURN" ;;
    RETRY)
      no_verdict=0; controller_set_counter no_verdict "$no_verdict"
      control_state complete >/dev/null; correction="$(json_field "$VERDICT_JSON" correction)"; log "turn $GLOBAL_TURN: RETRY — $correction" ;;
    REVERT)
      no_verdict=0; reverts=$((reverts+1))
      controller_set_counter no_verdict "$no_verdict"
      controller_set_counter reverts "$reverts"
      if (( reverts > MAX_REVERTS )); then
        controller_emergency MAX_REVERTS "MAX_REVERTS=$MAX_REVERTS exceeded" halted
        log "HALT: MAX_REVERTS=$MAX_REVERTS exceeded"; exit 4
      fi
      control_state complete >/dev/null
      restore_checkpoint; correction="$(json_field "$VERDICT_JSON" correction)"; log "turn $GLOBAL_TURN: REVERT #$reverts — $correction" ;;
    HALT)
      controller_emergency VALIDATOR_HALT "${reason:-validator hard-stop}" halted
      log "HALT: validator hard-stop — ${reason:-}"; exit 4 ;;
    *)
      no_verdict=$((no_verdict+1))
      controller_set_counter no_verdict "$no_verdict"
      if (( no_verdict >= MAX_NO_VERDICT )); then
        controller_emergency NO_VALID_VERDICT "no validator verdict for $no_verdict consecutive turns" halted
        log "HALT: no VALIDATOR-VERDICT.json for $no_verdict consecutive turns. Check validator configuration."
        exit 4
      fi
      control_state complete >/dev/null
      log "turn $GLOBAL_TURN: no verdict ($no_verdict/$MAX_NO_VERDICT); retrying"
      correction="Emit a VALIDATOR-VERDICT.json with a verdict and cited evidence." ;;
  esac

  terminal="$(json_field "$RUN_JSON" terminal)"; stage="$(json_field "$RUN_JSON" stage)"
  log "turn $i: stage=${stage:-?} terminal=${terminal:-none}"
  case "$terminal" in
    blocked)
      controller_emergency RUN_BLOCKED "run ledger reported blocked" halted
      log "STOP: blocked (see ORCHESTRATION-STATE.json / VALIDATION.jsonl)."; exit 4 ;;
    budget)
      controller_pause
      log "STOP: budget ceiling; re-run --resume."; exit 3 ;;
    human_gate)
      if [[ "$verdict" == "PROCEED" ]]; then
        controller_release
        log "DONE: human-ready gate reached (human review before merge to main)."; exit 0
      fi
      log "turn $GLOBAL_TURN: terminal=$terminal ignored until the Shepherd ratifies this turn with PROCEED" ;;
    done)
      if [[ "$verdict" == "PROCEED" ]]; then
        controller_release
        log "DONE: all sub-issues complete and verified."; exit 0
      fi
      log "turn $GLOBAL_TURN: terminal=$terminal ignored until the Shepherd ratifies this turn with PROCEED" ;;
  esac

  sleep "$COOLDOWN_SECONDS"
done
controller_pause
log "STOP: MAX_ITERATIONS=$MAX_ITERATIONS for this invocation; resume is allowed while the durable MAX_TURNS=$MAX_TURNS cap remains."; exit 3
