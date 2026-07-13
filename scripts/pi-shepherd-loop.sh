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

# Create and validate each controller-state directory component without following aliases. The
# controller lock and all durable ledgers derive their authority from this root, so a symlinked or
# foreign-owned ancestor must fail before any lock, prompt, model, or provider side effect.
if ! python3 - "$REPO_ROOT" <<'PY'
import os
import stat
import sys

root = sys.argv[1]
directory_flags = os.O_RDONLY | getattr(os, "O_DIRECTORY", 0)
if hasattr(os, "O_NOFOLLOW"):
    directory_flags |= os.O_NOFOLLOW

opened = []
try:
    current = os.open(root, directory_flags)
    opened.append(current)
    for component in (".planning", "auto-loop"):
        try:
            os.mkdir(component, 0o700 if component == "auto-loop" else 0o755, dir_fd=current)
            os.fsync(current)
        except FileExistsError:
            pass
        child = os.open(component, directory_flags, dir_fd=current)
        info = os.fstat(child)
        if (
            not stat.S_ISDIR(info.st_mode)
            or info.st_uid != os.geteuid()
            or stat.S_IMODE(info.st_mode) & 0o022
        ):
            raise OSError(f"unsafe state directory component: {component}")
        opened.append(child)
        current = child
except OSError as exc:
    print(f"CONTROL_STATE_DIR_UNSAFE: {exc}", file=sys.stderr)
    raise SystemExit(4)
finally:
    for descriptor in reversed(opened):
        os.close(descriptor)
PY
then
  exit 4
fi
CONTROL_JSON="$STATE_DIR/CONTROL.json"

# Acquire one worktree-wide advisory lock and retain its open file description across re-exec and
# all descendants. A surviving registered child therefore keeps replacement controllers fenced.
if [[ -z "${AUTO_LOOP_CONTROL_FD:-}" && -z "${AUTO_LOOP_STATE_FD:-}" ]]; then
  exec python3 - "$REPO_ROOT" "$STATE_DIR" "$SCRIPT_SELF" "$@" <<'PY'
import fcntl
import os
import stat
import sys

root_path, _state_path, script, *arguments = sys.argv[1:]
directory_flags = os.O_RDONLY | getattr(os, "O_DIRECTORY", 0)
if hasattr(os, "O_NOFOLLOW"):
    directory_flags |= os.O_NOFOLLOW
try:
    fd = os.open(root_path, directory_flags)
    root_info = os.fstat(fd)
    if not stat.S_ISDIR(root_info.st_mode) or root_info.st_uid != os.geteuid():
        raise OSError("unsafe worktree root")
    fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
    planning_fd = os.open(".planning", directory_flags, dir_fd=fd)
    planning_info = os.fstat(planning_fd)
    if not stat.S_ISDIR(planning_info.st_mode) or planning_info.st_uid != os.geteuid() or \
       stat.S_IMODE(planning_info.st_mode) & 0o022:
        raise OSError("unsafe .planning directory")
    state_fd = os.open("auto-loop", directory_flags, dir_fd=planning_fd)
    os.close(planning_fd)
    state_info = os.fstat(state_fd)
    if not stat.S_ISDIR(state_info.st_mode) or state_info.st_uid != os.geteuid() or \
       stat.S_IMODE(state_info.st_mode) & 0o022:
        raise OSError("unsafe controller state root")
    anchor_payload = f"{state_info.st_dev}:{state_info.st_ino}\n".encode()
    read_flags = os.O_RDONLY | getattr(os, "O_NONBLOCK", 0)
    create_flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL
    if hasattr(os, "O_NOFOLLOW"):
        read_flags |= os.O_NOFOLLOW
        create_flags |= os.O_NOFOLLOW
    try:
        recovery_info = os.stat(".auto-loop-recovery", dir_fd=fd, follow_symlinks=False)
    except FileNotFoundError:
        recovery_info = None
    if recovery_info is not None:
        if not stat.S_ISREG(recovery_info.st_mode) or recovery_info.st_nlink != 1:
            raise OSError("unsafe recovery anchor")
        print("RECOVERY_REQUIRED: worktree state root requires human reconciliation", file=sys.stderr)
        raise SystemExit(4)
    try:
        anchor_fd = os.open(".auto-loop-anchor", read_flags, dir_fd=fd)
    except FileNotFoundError:
        anchor_fd = os.open(".auto-loop-anchor", create_flags, 0o600, dir_fd=fd)
        try:
            os.write(anchor_fd, anchor_payload)
            os.fsync(anchor_fd)
        finally:
            os.close(anchor_fd)
        os.fsync(fd)
    else:
        anchor_info = os.fstat(anchor_fd)
        anchor_bytes = os.read(anchor_fd, 256)
        os.close(anchor_fd)
        if not stat.S_ISREG(anchor_info.st_mode) or anchor_info.st_nlink != 1 or \
           anchor_bytes != anchor_payload:
            try:
                recovery_fd = os.open(".auto-loop-recovery", create_flags, 0o600, dir_fd=fd)
            except FileExistsError:
                pass
            else:
                os.write(recovery_fd, b"state-root-moved\n")
                os.fsync(recovery_fd)
                os.close(recovery_fd)
                os.fsync(fd)
            print("RECOVERY_REQUIRED: worktree state-root anchor changed", file=sys.stderr)
            raise SystemExit(4)
except BlockingIOError:
    print("CONTROLLER_HELD: another Shepherd controller owns this worktree", file=sys.stderr)
    raise SystemExit(75)
except OSError as exc:
    print(f"CONTROLLER_LOCK_UNSAFE: {exc}", file=sys.stderr)
    raise SystemExit(4)

os.set_inheritable(state_fd, True)
os.set_inheritable(fd, True)
environment = os.environ.copy()
environment["AUTO_LOOP_STATE_FD"] = str(state_fd)
environment["AUTO_LOOP_CONTROL_FD"] = str(fd)
os.execve(script, [script, *arguments], environment)
PY
elif [[ -z "${AUTO_LOOP_CONTROL_FD:-}" || -z "${AUTO_LOOP_STATE_FD:-}" ]]; then
  printf '%s\n' 'CONTROL_FENCE_INVALID: incomplete inherited controller authority' >&2
  exit 4
fi

# Reject forged or moved inherited descriptors before touching controller state.
if ! python3 - "$AUTO_LOOP_STATE_FD" "$STATE_DIR" "$AUTO_LOOP_CONTROL_FD" "$REPO_ROOT" <<'PY'
import fcntl
import os
import stat
import sys

try:
    state_fd = int(sys.argv[1])
    state_held = os.fstat(state_fd)
    fd = int(sys.argv[3])
    directory_flags = os.O_RDONLY | getattr(os, "O_DIRECTORY", 0)
    if hasattr(os, "O_NOFOLLOW"):
        directory_flags |= os.O_NOFOLLOW
    planning_fd = os.open(".planning", directory_flags, dir_fd=fd)
    canonical_state_fd = os.open("auto-loop", directory_flags, dir_fd=planning_fd)
    state_path = os.fstat(canonical_state_fd)
    os.close(canonical_state_fd)
    os.close(planning_fd)
    if not stat.S_ISDIR(state_held.st_mode) or (
        state_held.st_dev, state_held.st_ino
    ) != (state_path.st_dev, state_path.st_ino):
        raise OSError("state descriptor does not match canonical path")
    held = os.fstat(fd)
    path = os.stat(sys.argv[4], follow_symlinks=False)
    if not stat.S_ISDIR(held.st_mode) or (
        held.st_dev, held.st_ino
    ) != (path.st_dev, path.st_ino):
        raise OSError("lock descriptor does not match canonical worktree root")
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
RUN_JSON="$STATE_DIR/RUN.json"
VERDICT_JSON="$STATE_DIR/VALIDATOR-VERDICT.json"
PROMPT_FILE="$STATE_DIR/PROMPT.txt"
LOG_FILE="$STATE_DIR/driver.log"
VAL_PROMPT="$REPO_ROOT/.agents/agentic-delivery/prompts/shepherd-validator-prompt.md"

log() { printf '%s %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*" | tee -a "$LOG_FILE" >&2; }

control_state_exec() { # command [arguments...]
  local command="$1"
  shift
  # The small outer launcher makes every state helper a self-led process group. Turn-bounded
  # callers can therefore drain a wedged interpreter and all of its descendants without ever
  # signalling the controller's own process group.
  exec python3 -c 'import os,sys; os.setsid(); os.execvp(sys.argv[1], sys.argv[1:])' \
    python3 - "$AUTO_LOOP_CONTROL_FD" "$AUTO_LOOP_STATE_FD" "$CONTROL_JSON" \
    "$command" "${CONTROL_FENCE:-}" "$@" <<'PY'
import datetime as dt
import json
import os
import pathlib
import stat
import sys
import uuid

root_fd = int(sys.argv[1])
state_fd = int(sys.argv[2])
path = pathlib.Path(sys.argv[3])
state_name = path.name
terminal_guard_name = "CONTROL.transition"
command = sys.argv[4]
fence_text = sys.argv[5]
args = sys.argv[6:]
UTC = dt.timezone.utc
source_present = False
source_bytes = None

ROOT_KEYS = {
    "schema_version", "run_id", "generation", "controller_id", "control_revision",
    "phase", "lease", "limits", "turn_ordinal", "counters", "active_turn", "halt",
    "children_quiescent", "updated_at",
}
LIMIT_KEYS = {
    "max_turns", "turn_timeout_seconds", "term_grace_seconds", "heartbeat_seconds",
    "max_reverts", "max_no_verdict", "max_minutes",
}
COUNTER_KEYS = {"reverts", "no_verdict", "active_seconds"}
TURN_KEYS = {
    "turn_id", "ordinal", "deadline_at", "orchestrator_session_id",
    "validator_session_id", "active_role", "leader_pid", "process_group_id",
}
HALT_KEYS = {"halt_id", "code", "reason", "created_at"}
PHASES = {"active", "halting", "halted", "recovery_required", "paused", "released"}
ROLES = {"orchestrator", "validator", "trace"}

def now():
    return dt.datetime.now(UTC)

def timestamp(value=None):
    return (value or now()).isoformat(timespec="seconds").replace("+00:00", "Z")

def fail(code, status=4):
    print(code, file=sys.stderr)
    raise SystemExit(status)

def unique_object(pairs):
    value = {}
    for key, item in pairs:
        if key in value:
            fail("CONTROL_STATE_INVALID:DUPLICATE_KEY")
        value[key] = item
    return value

def reject_constant(value):
    fail(f"CONTROL_STATE_INVALID:constant:{value}")

def exact_keys(name, value, expected):
    if not isinstance(value, dict) or set(value) != expected:
        fail(f"CONTROL_STATE_INVALID:{name}")

def strict_integer(name, value, minimum=0, maximum=9223372036854775807):
    if not isinstance(value, int) or isinstance(value, bool) or not minimum <= value <= maximum:
        fail(f"CONTROL_STATE_INVALID:{name}")
    return value

def canonical_uuid(name, value):
    try:
        parsed = uuid.UUID(value)
    except (AttributeError, TypeError, ValueError):
        fail(f"CONTROL_STATE_INVALID:{name}")
    if parsed.version is None or str(parsed) != value:
        fail(f"CONTROL_STATE_INVALID:{name}")

def canonical_time(name, value):
    if not isinstance(value, str):
        fail(f"CONTROL_STATE_INVALID:{name}")
    try:
        parsed = dt.datetime.strptime(value, "%Y-%m-%dT%H:%M:%SZ").replace(tzinfo=UTC)
    except ValueError:
        fail(f"CONTROL_STATE_INVALID:{name}")
    if timestamp(parsed) != value:
        fail(f"CONTROL_STATE_INVALID:{name}")
    return parsed

def open_canonical_state(create=False):
    directory_flags = os.O_RDONLY | getattr(os, "O_DIRECTORY", 0)
    if hasattr(os, "O_NOFOLLOW"):
        directory_flags |= os.O_NOFOLLOW
    planning_fd = -1
    canonical_fd = -1
    try:
        planning_fd = os.open(".planning", directory_flags, dir_fd=root_fd)
        planning_info = os.fstat(planning_fd)
        if not stat.S_ISDIR(planning_info.st_mode) or planning_info.st_uid != os.geteuid() or \
           stat.S_IMODE(planning_info.st_mode) & 0o022:
            raise OSError("unsafe .planning directory")
        if create:
            try:
                os.mkdir("auto-loop", 0o700, dir_fd=planning_fd)
                os.fsync(planning_fd)
            except FileExistsError:
                pass
        canonical_fd = os.open("auto-loop", directory_flags, dir_fd=planning_fd)
        canonical_info = os.fstat(canonical_fd)
        if not stat.S_ISDIR(canonical_info.st_mode) or canonical_info.st_uid != os.geteuid() or \
           stat.S_IMODE(canonical_info.st_mode) & 0o022:
            raise OSError("unsafe auto-loop directory")
        return planning_fd, canonical_fd
    except OSError:
        if canonical_fd >= 0:
            os.close(canonical_fd)
        if planning_fd >= 0:
            os.close(planning_fd)
        raise

def write_root_block():
    flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL
    if hasattr(os, "O_NOFOLLOW"):
        flags |= os.O_NOFOLLOW
    try:
        fd = os.open(".auto-loop-recovery", flags, 0o600, dir_fd=root_fd)
    except FileExistsError:
        return
    try:
        os.write(fd, b"state-root-moved\n")
        os.fsync(fd)
    finally:
        os.close(fd)
    os.fsync(root_fd)

def validate_state_root():
    planning_fd = canonical_fd = -1
    try:
        try:
            blocked = os.stat(".auto-loop-recovery", dir_fd=root_fd, follow_symlinks=False)
        except FileNotFoundError:
            blocked = None
        if blocked is not None:
            if not stat.S_ISREG(blocked.st_mode) or blocked.st_nlink != 1:
                fail("CONTROL_STATE_ROOT_UNSAFE")
            fail("RECOVERY_REQUIRED")
        planning_fd, canonical_fd = open_canonical_state(create=True)
        held = os.fstat(state_fd)
        canonical = os.fstat(canonical_fd)
        if (held.st_dev, held.st_ino) != (canonical.st_dev, canonical.st_ino):
            write_root_block()
            fail("CONTROL_STATE_ROOT_MOVED")
    except OSError:
        write_root_block()
        fail("CONTROL_STATE_ROOT_UNSAFE")
    finally:
        if canonical_fd >= 0:
            os.close(canonical_fd)
        if planning_fd >= 0:
            os.close(planning_fd)

validate_state_root()

def validate_state(value):
    exact_keys("root", value, ROOT_KEYS)
    if value.get("schema_version") != "1.0":
        fail("CONTROL_STATE_INVALID:schema_version")
    canonical_uuid("run_id", value.get("run_id"))
    canonical_uuid("controller_id", value.get("controller_id"))
    strict_integer("generation", value.get("generation"), 1)
    strict_integer("control_revision", value.get("control_revision"), 1)
    phase = value.get("phase")
    if phase not in PHASES:
        fail("CONTROL_STATE_INVALID:phase")
    if not isinstance(value.get("children_quiescent"), bool):
        fail("CONTROL_STATE_INVALID:children_quiescent")
    canonical_time("updated_at", value.get("updated_at"))

    exact_keys("lease", value.get("lease"), {"heartbeat_at", "expires_at"})
    heartbeat_at = canonical_time("lease.heartbeat_at", value["lease"]["heartbeat_at"])
    expires_at = canonical_time("lease.expires_at", value["lease"]["expires_at"])
    if expires_at < heartbeat_at:
        fail("CONTROL_STATE_INVALID:lease_order")

    exact_keys("limits", value.get("limits"), LIMIT_KEYS)
    limits = value["limits"]
    strict_integer("limits.max_turns", limits["max_turns"], 1, 1000000)
    strict_integer("limits.turn_timeout_seconds", limits["turn_timeout_seconds"], 1, 604800)
    strict_integer("limits.term_grace_seconds", limits["term_grace_seconds"], 1, 300)
    strict_integer("limits.heartbeat_seconds", limits["heartbeat_seconds"], 1, 300)
    strict_integer("limits.max_reverts", limits["max_reverts"], 1, 1000000)
    strict_integer("limits.max_no_verdict", limits["max_no_verdict"], 1, 1000000)
    strict_integer("limits.max_minutes", limits["max_minutes"], 0, 525600)
    if expires_at - heartbeat_at != dt.timedelta(seconds=3 * limits["heartbeat_seconds"]):
        fail("CONTROL_STATE_INVALID:lease_interval")

    exact_keys("counters", value.get("counters"), COUNTER_KEYS)
    strict_integer("counters.reverts", value["counters"]["reverts"], 0, 1000000)
    strict_integer("counters.no_verdict", value["counters"]["no_verdict"], 0, 1000000)
    strict_integer("counters.active_seconds", value["counters"]["active_seconds"], 0, 315360000)
    ordinal = strict_integer("turn_ordinal", value.get("turn_ordinal"), 0, limits["max_turns"])

    turn = value.get("active_turn")
    if turn is not None:
        exact_keys("active_turn", turn, TURN_KEYS)
        canonical_uuid("active_turn.turn_id", turn.get("turn_id"))
        canonical_uuid("active_turn.orchestrator_session_id", turn.get("orchestrator_session_id"))
        canonical_uuid("active_turn.validator_session_id", turn.get("validator_session_id"))
        if strict_integer("active_turn.ordinal", turn.get("ordinal"), 1, limits["max_turns"]) != ordinal:
            fail("CONTROL_STATE_INVALID:active_turn.ordinal")
        canonical_time("active_turn.deadline_at", turn.get("deadline_at"))
        role = turn.get("active_role")
        leader = turn.get("leader_pid")
        process_group = turn.get("process_group_id")
        if role is None:
            if leader is not None or process_group is not None:
                fail("CONTROL_STATE_INVALID:active_turn.handles")
        else:
            if role not in ROLES:
                fail("CONTROL_STATE_INVALID:active_turn.active_role")
            leader = strict_integer("active_turn.leader_pid", leader, 2, 2147483647)
            process_group = strict_integer(
                "active_turn.process_group_id", process_group, 2, 2147483647
            )
            if leader != process_group or value["children_quiescent"]:
                fail("CONTROL_STATE_INVALID:active_turn.binding")

    halt = value.get("halt")
    if halt is not None:
        exact_keys("halt", halt, HALT_KEYS)
        canonical_uuid("halt.halt_id", halt.get("halt_id"))
        if not isinstance(halt.get("code"), str) or not halt["code"]:
            fail("CONTROL_STATE_INVALID:halt.code")
        if not isinstance(halt.get("reason"), str) or len(halt["reason"]) > 2048:
            fail("CONTROL_STATE_INVALID:halt.reason")
        canonical_time("halt.created_at", halt.get("created_at"))

    if value["children_quiescent"] and turn is not None and turn.get("active_role") is not None:
        fail("CONTROL_STATE_INVALID:quiescence")
    if phase == "active":
        if halt is not None or (
            (turn is None or turn.get("active_role") is None) and not value["children_quiescent"]
        ):
            fail("CONTROL_STATE_INVALID:active")
    elif phase in {"paused", "released"}:
        if turn is not None or not value["children_quiescent"] or halt is not None:
            fail("CONTROL_STATE_INVALID:terminal")
        if value["counters"]["reverts"] > limits["max_reverts"] or \
           value["counters"]["no_verdict"] >= limits["max_no_verdict"]:
            fail("CONTROL_STATE_INVALID:terminal_counters")
    elif phase == "halted":
        if halt is None or not value["children_quiescent"]:
            fail("CONTROL_STATE_INVALID:halted")
    elif phase == "halting":
        if halt is None or value["children_quiescent"]:
            fail("CONTROL_STATE_INVALID:halting")
    elif phase == "recovery_required":
        if halt is None:
            fail("CONTROL_STATE_INVALID:recovery")

def read_state(required=True):
    global source_present, source_bytes
    try:
        entry = os.stat(state_name, dir_fd=state_fd, follow_symlinks=False)
        if not stat.S_ISREG(entry.st_mode) or entry.st_nlink != 1:
            fail("CONTROL_STATE_UNSAFE")
        present = True
    except FileNotFoundError:
        present = False
    except OSError:
        fail("CONTROL_STATE_UNSAFE")
    if not present:
        if required:
            fail("CONTROL_STATE_MISSING")
        return None
    flags = os.O_RDONLY | getattr(os, "O_NONBLOCK", 0)
    if hasattr(os, "O_NOFOLLOW"):
        flags |= os.O_NOFOLLOW
    try:
        fd = os.open(state_name, flags, dir_fd=state_fd)
        info = os.fstat(fd)
        if not stat.S_ISREG(info.st_mode) or info.st_nlink != 1 or \
           (info.st_dev, info.st_ino) != (entry.st_dev, entry.st_ino):
            fail("CONTROL_STATE_UNSAFE")
        if info.st_size > 131072:
            fail("CONTROL_STATE_OVERSIZED")
        with os.fdopen(fd, "rb") as handle:
            raw = handle.read()
        value = json.loads(
            raw.decode("utf-8"), object_pairs_hook=unique_object, parse_constant=reject_constant
        )
    except (OSError, UnicodeError, ValueError):
        fail("CONTROL_STATE_INVALID")
    validate_state(value)
    source_present = True
    source_bytes = raw
    return value

def write_state(value):
    global source_present, source_bytes
    value["updated_at"] = timestamp()
    validate_state(value)
    payload = (json.dumps(value, sort_keys=True, separators=(",", ":")) + "\n").encode()
    if source_present:
        flags = os.O_RDONLY | getattr(os, "O_NONBLOCK", 0)
        if hasattr(os, "O_NOFOLLOW"):
            flags |= os.O_NOFOLLOW
        try:
            current_fd = os.open(state_name, flags, dir_fd=state_fd)
            current_info = os.fstat(current_fd)
            if not stat.S_ISREG(current_info.st_mode) or current_info.st_nlink != 1:
                fail("CONTROL_STATE_MOVED")
            with os.fdopen(current_fd, "rb") as current:
                if current.read() != source_bytes:
                    fail("CONTROL_STATE_MOVED")
        except OSError:
            fail("CONTROL_STATE_MOVED")
    else:
        try:
            os.stat(state_name, dir_fd=state_fd, follow_symlinks=False)
        except FileNotFoundError:
            pass
        except OSError:
            fail("CONTROL_STATE_MOVED")
        else:
            fail("CONTROL_STATE_MOVED")
    temporary = f".CONTROL.{os.getpid()}.{uuid.uuid4().hex}"
    temporary_flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL
    if hasattr(os, "O_NOFOLLOW"):
        temporary_flags |= os.O_NOFOLLOW
    fd = os.open(temporary, temporary_flags, 0o600, dir_fd=state_fd)
    try:
        os.fchmod(fd, 0o600)
        offset = 0
        while offset < len(payload):
            offset += os.write(fd, payload[offset:])
        os.fsync(fd)
        os.close(fd)
        fd = -1
        os.replace(temporary, state_name, src_dir_fd=state_fd, dst_dir_fd=state_fd)
        os.fsync(state_fd)
        source_present = True
        source_bytes = payload
    except Exception:
        if fd >= 0:
            os.close(fd)
        try:
            os.unlink(temporary, dir_fd=state_fd)
        except FileNotFoundError:
            pass
        fail("CONTROL_COMMIT_UNCERTAIN")

def read_terminal_guard(required=False):
    flags = os.O_RDONLY | getattr(os, "O_NONBLOCK", 0)
    if hasattr(os, "O_NOFOLLOW"):
        flags |= os.O_NOFOLLOW
    try:
        fd = os.open(terminal_guard_name, flags, dir_fd=state_fd)
    except FileNotFoundError:
        if required:
            fail("CONTROL_TERMINAL_GUARD_MISSING")
        return None
    except OSError:
        fail("CONTROL_TERMINAL_GUARD_UNSAFE")
    try:
        info = os.fstat(fd)
        if not stat.S_ISREG(info.st_mode) or info.st_nlink != 1 or info.st_size > 4096:
            fail("CONTROL_TERMINAL_GUARD_UNSAFE")
        with os.fdopen(fd, "rb") as handle:
            raw = handle.read()
        guard = json.loads(
            raw.decode("utf-8"), object_pairs_hook=unique_object, parse_constant=reject_constant
        )
    except (OSError, UnicodeError, ValueError):
        fail("CONTROL_TERMINAL_GUARD_INVALID")
    exact_keys(
        "terminal_guard", guard,
        {"schema_version", "run_id", "generation", "controller_id", "control_revision", "transition"},
    )
    if guard["schema_version"] != "1.0" or guard["transition"] not in {"pause", "release"}:
        fail("CONTROL_TERMINAL_GUARD_INVALID")
    canonical_uuid("terminal_guard.run_id", guard["run_id"])
    canonical_uuid("terminal_guard.controller_id", guard["controller_id"])
    strict_integer("terminal_guard.generation", guard["generation"], 1)
    strict_integer("terminal_guard.control_revision", guard["control_revision"], 1)
    return guard

def expected_terminal_guard(value, transition):
    return {
        "schema_version": "1.0",
        "run_id": value["run_id"],
        "generation": value["generation"],
        "controller_id": value["controller_id"],
        "control_revision": value["control_revision"],
        "transition": transition,
    }

def create_terminal_guard(value, transition):
    if read_terminal_guard(required=False) is not None:
        fail("CONTROL_TERMINAL_GUARD_HELD")
    payload = (
        json.dumps(expected_terminal_guard(value, transition), sort_keys=True, separators=(",", ":"))
        + "\n"
    ).encode()
    flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL
    if hasattr(os, "O_NOFOLLOW"):
        flags |= os.O_NOFOLLOW
    fd = -1
    try:
        fd = os.open(terminal_guard_name, flags, 0o600, dir_fd=state_fd)
        offset = 0
        while offset < len(payload):
            written = os.write(fd, payload[offset:])
            if written <= 0:
                raise OSError("short terminal guard write")
            offset += written
        os.fsync(fd)
        os.close(fd)
        fd = -1
        os.fsync(state_fd)
    except OSError:
        if fd >= 0:
            os.close(fd)
        fail("CONTROL_COMMIT_UNCERTAIN")

def require_terminal_guard(value, transition):
    guard = read_terminal_guard(required=True)
    if guard != expected_terminal_guard(value, transition):
        fail("CONTROL_TERMINAL_GUARD_MISMATCH")

def clear_terminal_guard(value, transition):
    require_terminal_guard(value, transition)
    try:
        os.unlink(terminal_guard_name, dir_fd=state_fd)
        os.fsync(state_fd)
    except OSError:
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

def exact_turn_state(allowed=("active",)):
    value = exact_state(allowed)
    turn = value.get("active_turn")
    if not isinstance(turn, dict):
        fail("CONTROL_TURN_INVALID")
    if canonical_time("active_turn.deadline_at", turn.get("deadline_at")) <= now():
        fail("TURN_DEADLINE", 124)
    return value

def refreshed_lease(value):
    seconds = int(value["limits"]["heartbeat_seconds"]) * 3
    current = now()
    value["lease"] = {
        "heartbeat_at": timestamp(current),
        "expires_at": timestamp(current + dt.timedelta(seconds=seconds)),
    }

if command == "phase":
    value = read_state()
    if not fence_text or fence_of(value) != fence_text:
        fail("CONTROL_FENCE_MISMATCH")
    print(value["phase"])
elif command == "init":
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
    terminal_guard = read_terminal_guard(required=False)

    if terminal_guard is not None:
        if previous is None or terminal_guard != expected_terminal_guard(
            previous, terminal_guard["transition"]
        ):
            fail("CONTROL_TERMINAL_GUARD_MISMATCH")
        if previous["phase"] not in {"halted", "recovery_required"}:
            previous["phase"] = "recovery_required"
            if previous.get("halt") is None:
                previous["halt"] = {
                    "halt_id": str(uuid.uuid4()),
                    "code": "CONTROL_COMMIT_UNCERTAIN",
                    "reason": "a prior terminal transition was not durably acknowledged",
                    "created_at": timestamp(),
                }
            refreshed_lease(previous)
            write_state(previous)
        fail("RECOVERY_REQUIRED")

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
    value = exact_turn_state()
    refreshed_lease(value)
    write_state(value)
elif command == "heartbeat":
    value = exact_state()
    turn = value.get("active_turn")
    if isinstance(turn, dict) and canonical_time(
        "active_turn.deadline_at", turn.get("deadline_at")
    ) <= now():
        fail("TURN_DEADLINE", 124)
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
    print("\t".join((
        str(turn["ordinal"]), turn["turn_id"], turn["orchestrator_session_id"],
        turn["validator_session_id"], str(int((current + dt.timedelta(
            seconds=int(value["limits"]["turn_timeout_seconds"])
        )).timestamp())),
    )))
elif command == "bind":
    role, pid_raw, pgid_raw = args
    if role not in ("orchestrator", "validator", "trace"):
        fail("CONTROL_ROLE_INVALID")
    pid = parse_positive("leader_pid", pid_raw, maximum=2147483647)
    pgid = parse_positive("process_group_id", pgid_raw, maximum=2147483647)
    value = exact_turn_state()
    turn = value.get("active_turn")
    if not isinstance(turn, dict) or turn.get("active_role") is not None:
        fail("CONTROL_TURN_INVALID")
    turn.update({"active_role": role, "leader_pid": pid, "process_group_id": pgid})
    value["children_quiescent"] = False
    refreshed_lease(value)
    write_state(value)
elif command == "clear-role":
    role = args[0]
    value = exact_turn_state()
    turn = value.get("active_turn")
    if not isinstance(turn, dict) or turn.get("active_role") != role:
        fail("CONTROL_ROLE_MISMATCH")
    turn.update({"active_role": None, "leader_pid": None, "process_group_id": None})
    value["children_quiescent"] = True
    refreshed_lease(value)
    write_state(value)
elif command == "complete":
    value = exact_turn_state()
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
    value = exact_turn_state()
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
elif command == "prepare-terminal":
    transition = args[0]
    if transition not in ("pause", "release"):
        fail("CONTROL_PHASE_INVALID")
    value = exact_state()
    if not value.get("children_quiescent") or value.get("active_turn") is not None:
        fail("CONTROL_NOT_QUIESCENT")
    create_terminal_guard(value, transition)
elif command in ("pause", "release"):
    active_seconds = parse_positive("active_seconds", args[0], minimum=0, maximum=315360000)
    value = exact_state()
    if not value.get("children_quiescent") or value.get("active_turn") is not None:
        fail("CONTROL_NOT_QUIESCENT")
    require_terminal_guard(value, command)
    value["phase"] = "paused" if command == "pause" else "released"
    _, counters = validated_limits(value)
    counters["active_seconds"] = max(counters["active_seconds"], active_seconds)
    write_state(value)
elif command == "clear-terminal":
    transition = args[0]
    if transition not in ("pause", "release"):
        fail("CONTROL_PHASE_INVALID")
    expected_phase = "paused" if transition == "pause" else "released"
    value = exact_state((expected_phase,))
    clear_terminal_guard(value, transition)
elif command == "recover-uncertain":
    reason = args[0]
    value = exact_state(("active", "paused", "released", "halting", "recovery_required", "halted"))
    if value["phase"] == "halted":
        raise SystemExit(0)
    value["phase"] = "recovery_required"
    if value.get("halt") is None:
        value["halt"] = {
            "halt_id": str(uuid.uuid4()),
            "code": "CONTROL_COMMIT_UNCERTAIN",
            "reason": reason[:2048],
            "created_at": timestamp(),
        }
    refreshed_lease(value)
    write_state(value)
else:
    fail("CONTROL_COMMAND_INVALID", 2)
PY
}

control_state() { # command [arguments...]
  (control_state_exec "$@")
}

CONTROL_FENCE=""
RUN_ID=""
CONTROL_CLOSED=0
EMERGENCY_ACTIVE=0
AUTHORITY_ACQUIRED=0
ACTIVE_ROLE_PID=""
ACTIVE_ROLE_PGID=""
ACTIVE_ROLE_GROUP_READY=0
ACTIVE_ROLE_READY_FILE=""
ACTIVE_ROLE_AUTH_FILE=""
ROLE_GO_FD=19
ROLE_TOKEN_FD=18
ACTIVE_ROLE_GO_FIFO=""
ACTIVE_ROLE_GO_FD_OPEN=0
ACTIVE_ROLE_TOKEN_FD_OPEN=0
ACTIVE_ROLE_OUTPUT_FILE=""
ACTIVE_CONTROL_PID=""
ACTIVE_CONTROL_PGID=""
ACTIVE_CONTROL_GROUP_READY=0
LAST_ROLE_EXIT_STATUS=0
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
  RUN_ID="${CONTROL_FENCE%%|*}"
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

process_group_has_live_descendant() { # pgid leader-pid
  local pgid="$1" leader="$2" listing
  listing="$(ps -axo pgid=,pid=,stat= 2>/dev/null)" || return 2
  awk -v pgid="$pgid" -v leader="$leader" '
    $1 == pgid && $2 != leader && $3 !~ /^Z/ { found=1; exit }
    END { exit !found }
  ' <<<"$listing"
}

# kill -0 reports zombies as present. Treat an exited or zombie child leader as reapable so a
# short-lived role cannot consume its whole deadline while bash still has an unreaped child entry.
leader_process_exited() {
  local pid="$1" state
  [[ "$pid" =~ ^[0-9]+$ ]] || return 0
  kill -0 "$pid" 2>/dev/null || return 0
  if state="$(ps -o stat= -p "$pid" 2>/dev/null)"; then
    state="$(tr -d '[:space:]' <<<"$state")"
    if [[ "$state" == Z* ]]; then
      return 0
    fi
    # An empty/unknown successful probe is not proof of exit; preserve supervision until a
    # subsequent kill/ps observation is conclusive or the persisted deadline expires.
    return 1
  fi
  kill -0 "$pid" 2>/dev/null || return 0
  return 1
}

drain_control_helper() {
  local pid="${ACTIVE_CONTROL_PID:-}" pgid="${ACTIVE_CONTROL_PGID:-}" started observed="" \
    descendant_rc=0 quiescent=0
  [[ "$pid" =~ ^[0-9]+$ ]] || return 0
  if ! leader_process_exited "$pid"; then
    observed="$(ps -o pgid= -p "$pid" 2>/dev/null | tr -d '[:space:]' || true)"
  fi
  if (( ACTIVE_CONTROL_GROUP_READY == 1 )); then
    [[ "$observed" == "$pgid" || -z "$observed" ]] || return 1
    kill -TERM -- "-$pgid" 2>/dev/null || true
  else
    kill -TERM "$pid" 2>/dev/null || true
    started=$SECONDS
    while ! leader_process_exited "$pid" && (( SECONDS - started < TERM_GRACE_SECONDS )); do
      sleep 0.05
    done
    leader_process_exited "$pid" || kill -KILL "$pid" 2>/dev/null || true
  fi
  if (( ACTIVE_CONTROL_GROUP_READY == 1 )); then
    started=$SECONDS
    while (( SECONDS - started < TERM_GRACE_SECONDS )); do
      descendant_rc=0
      process_group_has_live_descendant "$pgid" "$pid" || descendant_rc=$?
      if leader_process_exited "$pid" && (( descendant_rc == 1 )); then
        quiescent=1
        break
      fi
      sleep 0.05
    done
    if (( quiescent == 0 )); then
      kill -KILL -- "-$pgid" 2>/dev/null || true
    fi
  fi
  started=$SECONDS
  while (( SECONDS - started < 2 )); do
    descendant_rc=1
    if (( ACTIVE_CONTROL_GROUP_READY == 1 )); then
      descendant_rc=0
      process_group_has_live_descendant "$pgid" "$pid" || descendant_rc=$?
    fi
    if leader_process_exited "$pid" && (( descendant_rc == 1 )); then
      quiescent=1
      break
    fi
    sleep 0.05
  done
  (( quiescent == 1 )) || return 1
  wait "$pid" 2>/dev/null || true
  ACTIVE_CONTROL_PID=""
  ACTIVE_CONTROL_PGID=""
  ACTIVE_CONTROL_GROUP_READY=0
  return 0
}

exec_before_monotonic_deadline() { # monotonic-deadline exec-function [arguments...]
  local deadline="$1" runner="$2" pid observed="" rc=0 descendant_rc=0
  shift 2
  (( SECONDS < deadline )) || return 124
  "$runner" "$@" &
  pid=$!
  ACTIVE_CONTROL_PID="$pid"
  ACTIVE_CONTROL_PGID="$pid"
  ACTIVE_CONTROL_GROUP_READY=0

  while ! leader_process_exited "$pid"; do
    observed="$(ps -o pgid= -p "$pid" 2>/dev/null | tr -d '[:space:]' || true)"
    if [[ "$observed" == "$pid" ]]; then
      ACTIVE_CONTROL_GROUP_READY=1
    fi
    if (( SECONDS >= deadline )); then
      drain_control_helper || true
      return 124
    fi
    sleep 0.02
  done
  process_group_has_live_descendant "$pid" "$pid" || descendant_rc=$?
  if (( descendant_rc == 0 )); then
    ACTIVE_CONTROL_GROUP_READY=1
    drain_control_helper || true
    return 125
  elif (( descendant_rc != 1 )); then
    ACTIVE_CONTROL_GROUP_READY=1
    drain_control_helper || true
    return 126
  fi
  wait "$pid" 2>/dev/null || rc=$?
  ACTIVE_CONTROL_PID=""
  ACTIVE_CONTROL_PGID=""
  ACTIVE_CONTROL_GROUP_READY=0
  (( SECONDS < deadline )) || return 124
  return "$rc"
}

control_state_before_monotonic_deadline() { # monotonic-deadline command [arguments...]
  local deadline="$1"
  shift
  exec_before_monotonic_deadline "$deadline" control_state_exec "$@"
}

control_state_before_turn_deadline() { # command [arguments...]
  control_state_before_monotonic_deadline "$TURN_DEADLINE_MONO" "$@"
}

drain_role_group() {
  local pid="${ACTIVE_ROLE_PID:-}" pgid="${ACTIVE_ROLE_PGID:-}" started observed_pgid=""
  [[ "$pid" =~ ^[0-9]+$ ]] || return 0
  if kill -0 "$pid" 2>/dev/null; then
    observed_pgid="$(ps -o pgid= -p "$pid" 2>/dev/null | tr -d '[:space:]' || true)"
  fi
  if (( ACTIVE_ROLE_GROUP_READY == 1 )); then
    if [[ -n "$observed_pgid" && "$observed_pgid" != "$pgid" ]]; then
      # A PID whose observed group no longer matches the durable binding is untrusted. Never signal
      # that PID (or the possibly reused PGID); retain the handles for authenticated recovery.
      return 1
    fi
  elif [[ "$observed_pgid" == "$pid" ]]; then
    pgid="$pid"
    ACTIVE_ROLE_PGID="$pgid"
    ACTIVE_ROLE_GROUP_READY=1
  else
    kill -TERM "$pid" 2>/dev/null || true
    started=$SECONDS
    while ! leader_process_exited "$pid" && (( SECONDS - started < TERM_GRACE_SECONDS )); do
      sleep 0.1
    done
    leader_process_exited "$pid" || kill -KILL "$pid" 2>/dev/null || true
    started=$SECONDS
    while ! leader_process_exited "$pid" && (( SECONDS - started < 2 )); do
      sleep 0.1
    done
    leader_process_exited "$pid" || return 1
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
  started=$SECONDS
  while { process_group_alive "$pgid" || ! leader_process_exited "$pid"; } && \
        (( SECONDS - started < 2 )); do
    sleep 0.1
  done
  if process_group_alive "$pgid" || ! leader_process_exited "$pid"; then
    return 1
  fi
  wait "$pid" 2>/dev/null || true
  ACTIVE_ROLE_PID=""
  ACTIVE_ROLE_PGID=""
  ACTIVE_ROLE_GROUP_READY=0
  return 0
}

cleanup_role_artifacts() {
  if (( ACTIVE_ROLE_TOKEN_FD_OPEN == 1 )); then
    exec 18<&-
    ACTIVE_ROLE_TOKEN_FD_OPEN=0
  fi
  if (( ACTIVE_ROLE_GO_FD_OPEN == 1 )); then
    exec 19>&-
    ACTIVE_ROLE_GO_FD_OPEN=0
  fi
  [[ -z "$ACTIVE_ROLE_READY_FILE" ]] || rm -f "$ACTIVE_ROLE_READY_FILE" 2>/dev/null || true
  [[ -z "$ACTIVE_ROLE_AUTH_FILE" ]] || rm -f "$ACTIVE_ROLE_AUTH_FILE" 2>/dev/null || true
  [[ -z "$ACTIVE_ROLE_GO_FIFO" ]] || rm -f "$ACTIVE_ROLE_GO_FIFO" 2>/dev/null || true
  [[ -z "$ACTIVE_ROLE_OUTPUT_FILE" ]] || rm -f "$ACTIVE_ROLE_OUTPUT_FILE" 2>/dev/null || true
  ACTIVE_ROLE_READY_FILE=""
  ACTIVE_ROLE_AUTH_FILE=""
  ACTIVE_ROLE_GO_FIFO=""
  ACTIVE_ROLE_OUTPUT_FILE=""
}

controller_emergency() { # code reason final-phase
  local code="$1" reason="$2" final_phase="$3" latch_phase="halting" latch_ok=0 \
    drain_ok=0 helper_drain_ok=0 role_drain_ok=0 emergency_budget emergency_deadline
  if (( EMERGENCY_ACTIVE == 1 )); then
    return 4
  fi
  EMERGENCY_ACTIVE=1
  trap '' INT TERM HUP
  [[ "$final_phase" == "recovery_required" ]] && latch_phase="recovery_required"
  emergency_budget=$((TERM_GRACE_SECONDS + 3))
  (( emergency_budget > 15 )) && emergency_budget=15
  emergency_deadline=$((SECONDS + emergency_budget))
  printf '%s: %s\n' "$code" "$reason" >&2
  if drain_control_helper; then
    helper_drain_ok=1
  fi
  if (( helper_drain_ok == 1 )); then
    if control_state_before_monotonic_deadline "$emergency_deadline" \
        latch "$latch_phase" "$code" "$reason" >/dev/null 2>&1; then
      latch_ok=1
    fi
  fi
  if drain_role_group; then
    role_drain_ok=1
  fi
  (( helper_drain_ok == 1 && role_drain_ok == 1 )) && drain_ok=1
  cleanup_role_artifacts
  if (( latch_ok == 1 && drain_ok == 1 )); then
    control_state_before_monotonic_deadline "$emergency_deadline" \
      finalize "$final_phase" true >/dev/null 2>&1 || true
  elif (( latch_ok == 1 )); then
    if control_state_before_monotonic_deadline "$emergency_deadline" \
        latch recovery_required TEARDOWN_NOT_QUIESCENT \
        "registered process group could not be proven quiescent" >/dev/null 2>&1; then
      control_state_before_monotonic_deadline "$emergency_deadline" \
        finalize recovery_required false >/dev/null 2>&1 || true
    fi
  elif (( drain_ok == 1 )); then
    if control_state_before_monotonic_deadline "$emergency_deadline" \
        latch recovery_required HALT_PERSISTENCE_FAILED \
        "the requested halt could not be durably latched" >/dev/null 2>&1; then
      control_state_before_monotonic_deadline "$emergency_deadline" \
        finalize recovery_required true >/dev/null 2>&1 || true
    fi
  fi
  CONTROL_CLOSED=1
  EMERGENCY_ACTIVE=0
}

controller_on_signal() {
  trap '' INT TERM HUP
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
  local ready_file output_file pid pgid attempt rc=0 deadline next_heartbeat found=1 leader_reaped=0
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
state_descriptor = os.environ.pop("AUTO_LOOP_STATE_FD", "")
if state_descriptor:
    os.close(int(state_descriptor))
os.execvp(command, [command, *arguments])
PY
  pid=$!
  ACTIVE_ROLE_PID="$pid"
  ACTIVE_ROLE_PGID="$pid"
  ACTIVE_ROLE_GROUP_READY=0
  for ((attempt=0; attempt<1000; attempt++)); do
    [[ -s "$ready_file" ]] && break
    leader_process_exited "$pid" && break
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
    if leader_process_exited "$pid"; then
      wait "$pid" 2>/dev/null || rc=$?
      leader_reaped=1
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
      if ! control_state_before_monotonic_deadline "$deadline" heartbeat >/dev/null; then
        controller_emergency CONTROL_FENCE_MISMATCH "validator-model discovery lost controller authority" recovery_required
        return 126
      fi
      next_heartbeat=$((SECONDS + CONTROL_HEARTBEAT_SECONDS))
    fi
    sleep 0.1
  done
  if [[ -n "$ACTIVE_ROLE_PID" && "$leader_reaped" -eq 0 ]]; then
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

controller_terminal_transition() { # pause|release
  local transition="$1" rc=0 terminal_deadline=$((SECONDS + 15))
  if ! control_state_before_monotonic_deadline "$terminal_deadline" \
      prepare-terminal "$transition" >/dev/null; then
    controller_emergency CONTROL_TERMINAL_GUARD_FAILED \
      "could not persist the $transition transition guard" recovery_required
    return 4
  fi
  if control_state_before_monotonic_deadline "$terminal_deadline" \
      "$transition" "$(controller_active_seconds)" >/dev/null; then
    if control_state_before_monotonic_deadline "$terminal_deadline" \
        clear-terminal "$transition" >/dev/null; then
      CONTROL_CLOSED=1
      return 0
    else
      rc=$?
    fi
  else
    rc=$?
  fi
  printf 'CONTROL_COMMIT_UNCERTAIN: %s returned %s; forcing non-resumable recovery\n' \
    "$transition" "$rc" >&2
  if control_state_before_monotonic_deadline "$terminal_deadline" recover-uncertain \
      "$transition outcome was not acknowledged by the controller" >/dev/null 2>&1; then
    CONTROL_CLOSED=1
  fi
  return 4
}

controller_pause() {
  controller_terminal_transition pause
}

controller_release() {
  controller_terminal_transition release
}

controller_turn_state_command() { # command [arguments...]
  local command="$1" rc=0
  shift
  control_state_before_turn_deadline "$command" "$@" >/dev/null || rc=$?
  if (( rc == 124 )); then
    controller_emergency TURN_DEADLINE \
      "$command exhausted the persisted turn deadline" halted
    return 4
  elif (( rc != 0 )); then
    controller_emergency CONTROL_FENCE_MISMATCH \
      "$command could not update the active turn" recovery_required
    return 4
  fi
}

controller_set_counter() { # name value
  controller_turn_state_command counter "$1" "$2"
}

controller_complete_turn() {
  controller_turn_state_command complete
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

json_field_exec() { # $1=file $2=key
  [[ -f "$1" ]] || { echo ""; return 0; }
  exec python3 -c 'import os,sys; os.setsid(); os.execvp(sys.argv[1], sys.argv[1:])' \
    python3 - "$1" "$2" <<'PY'
import json,sys
import os,stat
try:
    flags=os.O_RDONLY|getattr(os,"O_NONBLOCK",0)
    if hasattr(os,"O_NOFOLLOW"): flags|=os.O_NOFOLLOW
    fd=os.open(sys.argv[1],flags)
    info=os.fstat(fd)
    if not stat.S_ISREG(info.st_mode) or info.st_nlink != 1 or info.st_size > 131072:
        raise ValueError("unsafe JSON input")
    with os.fdopen(fd,"rb") as handle: d=json.loads(handle.read().decode("utf-8"))
    v=d.get(sys.argv[2])
    if isinstance(v,dict): v=v.get("type","")
    print("" if v is None else v)
except Exception: print("")
PY
}

json_field() { # $1=file $2=key
  (json_field_exec "$@") 2>/dev/null || echo ""
}

json_field_before_turn_deadline() { # $1=file $2=key
  exec_before_monotonic_deadline "$TURN_DEADLINE_MONO" json_field_exec "$@"
}

verdict_snapshot_exec() { # $1=retained state-dir fd $2=current run id
  exec python3 -c 'import os,sys; os.setsid(); os.execvp(sys.argv[1], sys.argv[1:])' \
    python3 - "$1" "$2" <<'PY'
import json
import math
import os
import stat
import sys
import uuid

state_fd = int(sys.argv[1])
run_id = str(uuid.UUID(sys.argv[2]))
directory_flags = os.O_RDONLY | getattr(os, "O_DIRECTORY", 0)
if hasattr(os, "O_NOFOLLOW"):
    directory_flags |= os.O_NOFOLLOW

def read_regular_at(parent_fd, name, maximum):
    flags = os.O_RDONLY | getattr(os, "O_NONBLOCK", 0)
    if hasattr(os, "O_NOFOLLOW"):
        flags |= os.O_NOFOLLOW
    fd = os.open(name, flags, dir_fd=parent_fd)
    try:
        before = os.fstat(fd)
        if (
            not stat.S_ISREG(before.st_mode) or before.st_nlink != 1
            or before.st_uid != os.geteuid() or before.st_size > maximum
        ):
            raise ValueError("unsafe verdict artifact")
        chunks = []
        remaining = maximum + 1
        while remaining:
            chunk = os.read(fd, remaining)
            if not chunk:
                break
            chunks.append(chunk)
            remaining -= len(chunk)
        payload = b"".join(chunks)
        after = os.fstat(fd)
        if (
            len(payload) > maximum or len(payload) != after.st_size
            or (before.st_dev, before.st_ino, before.st_size, before.st_mtime_ns)
            != (after.st_dev, after.st_ino, after.st_size, after.st_mtime_ns)
        ):
            raise ValueError("verdict artifact changed during snapshot")
        return payload
    finally:
        os.close(fd)

def open_private_dir(parent_fd, name):
    fd = os.open(name, directory_flags, dir_fd=parent_fd)
    info = os.fstat(fd)
    if (
        not stat.S_ISDIR(info.st_mode) or info.st_uid != os.geteuid()
        or stat.S_IMODE(info.st_mode) & 0o022
    ):
        os.close(fd)
        raise ValueError("unsafe checkpoint directory")
    return fd

def unique_object(pairs):
    result = {}
    for key, item in pairs:
        if key in result:
            raise ValueError("duplicate key")
        result[key] = item
    return result

try:
    value = json.loads(
        read_regular_at(state_fd, "VALIDATOR-VERDICT.json", 131072).decode("utf-8"),
        object_pairs_hook=unique_object,
        parse_constant=lambda item: (_ for _ in ()).throw(ValueError(item)),
    )
except (OSError, UnicodeError, ValueError):
    value = {}

expected_keys = {
    "verdict", "step_score", "trajectory_geomean", "reason", "correction",
    "revert_to_checkpoint",
}
valid = isinstance(value, dict) and set(value) == expected_keys
verdict = value.get("verdict") if valid else ""
valid = valid and verdict in {"PROCEED", "RETRY", "REVERT", "HALT"}

score_value = value.get("step_score") if valid else None
trajectory_value = value.get("trajectory_geomean") if valid else None
numeric = lambda item: (
    isinstance(item, (int, float)) and not isinstance(item, bool) and math.isfinite(item)
)
valid = valid and numeric(score_value) and 1 <= score_value <= 5
valid = valid and numeric(trajectory_value) and 1 <= trajectory_value <= 5

reason = value.get("reason") if valid else ""
valid = (
    valid and isinstance(reason, str) and bool(reason.strip()) and len(reason.encode()) <= 8192
    and not any(ord(character) < 32 or ord(character) == 127 for character in reason)
)
correction_value = value.get("correction") if valid else None
checkpoint_value = value.get("revert_to_checkpoint") if valid else None

def valid_text(item):
    return (
        isinstance(item, str) and bool(item.strip()) and len(item.encode()) <= 8192
        and not any(ord(character) < 32 or ord(character) == 127 for character in item)
    )

def checkpoint_bundle_is_current(target):
    if not isinstance(target, str) or not target.isdecimal() or target == "0" or str(int(target)) != target:
        return False
    opened = []
    try:
        checkpoints_fd = open_private_dir(state_fd, "checkpoints")
        opened.append(checkpoints_fd)
        run_fd = open_private_dir(checkpoints_fd, f"run-{run_id}")
        opened.append(run_fd)
        marker = read_regular_at(run_fd, "LAST_GOOD", 32).decode("ascii").strip()
        if marker != target:
            return False
        target_fd = open_private_dir(run_fd, target)
        opened.append(target_fd)
        run_bytes = read_regular_at(target_fd, "RUN.json", 1048576)
        head = read_regular_at(target_fd, "HEAD.sha", 128).decode("ascii").strip()
        run_value = json.loads(
            run_bytes.decode("utf-8"), object_pairs_hook=unique_object,
            parse_constant=lambda item: (_ for _ in ()).throw(ValueError(item)),
        )
        return (
            isinstance(run_value, dict) and len(head) in (40, 64)
            and all(character in "0123456789abcdef" for character in head)
        )
    except (OSError, UnicodeError, ValueError):
        return False
    finally:
        for descriptor in reversed(opened):
            os.close(descriptor)

if valid and verdict == "PROCEED":
    valid = score_value >= 4 and correction_value is None and checkpoint_value is None
elif valid and verdict == "RETRY":
    valid = (
        2 <= score_value < 4 and valid_text(correction_value) and checkpoint_value is None
    )
elif valid and verdict == "REVERT":
    valid = (
        score_value < 2 and valid_text(correction_value)
        and checkpoint_bundle_is_current(checkpoint_value)
    )
elif valid and verdict == "HALT":
    valid = correction_value is None and checkpoint_value is None

if not valid:
    verdict, score, reason, correction, checkpoint = "", "", "", "", ""
else:
    score = str(score_value)
    correction = correction_value if isinstance(correction_value, str) else ""
    checkpoint = checkpoint_value if isinstance(checkpoint_value, str) else ""
print("|".join((verdict, score, reason.encode().hex(), correction.encode().hex(), checkpoint)))
PY
}

decode_hex_text() {
  python3 - "$1" <<'PY'
import sys
try:
    print(bytes.fromhex(sys.argv[1]).decode("utf-8"), end="")
except (UnicodeError, ValueError):
    raise SystemExit(1)
PY
}

start_role_group() { # role command [arguments...]
  local role="$1" ready_file auth_file go_fifo pid pgid observed_pgid remaining role_token bind_rc=0 assert_rc=0
  shift
  LAST_ROLE_EXIT_STATUS=0
  ready_file="$STATE_DIR/.role-ready.$TURN_ID.$role"
  auth_file="$STATE_DIR/.role-authorized.$TURN_ID.$role"
  go_fifo="$STATE_DIR/.role-go.$TURN_ID.$role"
  ACTIVE_ROLE_READY_FILE="$ready_file"
  ACTIVE_ROLE_AUTH_FILE="$auth_file"
  ACTIVE_ROLE_GO_FIFO="$go_fifo"
  remaining=$((TURN_DEADLINE_MONO - SECONDS))
  if (( remaining <= 0 )); then
    controller_emergency TURN_DEADLINE "$role exhausted the persisted turn deadline before startup" halted
    return 124
  fi
  if [[ "${AUTO_LOOP_CONTROL_FD:-}" == "$ROLE_GO_FD" || \
        "${AUTO_LOOP_CONTROL_FD:-}" == "$ROLE_TOKEN_FD" || \
        "${AUTO_LOOP_STATE_FD:-}" == "$ROLE_GO_FD" || \
        "${AUTO_LOOP_STATE_FD:-}" == "$ROLE_TOKEN_FD" || \
        "$ROLE_GO_FD" == "$ROLE_TOKEN_FD" ]]; then
    controller_emergency ROLE_START_FAILED "role authorization descriptor collides with the controller lock" recovery_required
    return 126
  fi
  role_token="$(python3 -c 'import secrets; print(secrets.token_hex(32))')"
  if [[ ! "$role_token" =~ ^[0-9a-f]{64}$ ]] || \
     ! exec 18< <(printf '%s\n' "$role_token"); then
    controller_emergency ROLE_START_FAILED "could not create private $role authorization token" recovery_required
    return 126
  fi
  ACTIVE_ROLE_TOKEN_FD_OPEN=1
  # The FIFO is only a rendezvous: a same-UID opener can deny service or inject bytes. Authority is
  # the fresh token carried to the inert child on a separate anonymous descriptor and published by
  # the controller only after the exact bind and final assertion are durable.
  if ! (umask 077; mkfifo -m 600 "$go_fifo") || [[ ! -p "$go_fifo" || -L "$go_fifo" ]]; then
    controller_emergency ROLE_START_FAILED "could not create private $role authorization channel" recovery_required
    return 126
  fi
  if ! exec 19<>"$go_fifo"; then
    controller_emergency ROLE_START_FAILED "could not open private $role authorization channel" recovery_required
    return 126
  fi
  ACTIVE_ROLE_GO_FD_OPEN=1
  if ! python3 - "$ROLE_GO_FD" <<'PY'
import fcntl
import os
import sys

descriptor = int(sys.argv[1])
fcntl.fcntl(descriptor, fcntl.F_SETFL, fcntl.fcntl(descriptor, fcntl.F_GETFL) | os.O_NONBLOCK)
PY
  then
    controller_emergency ROLE_START_FAILED "could not bound $role authorization channel" recovery_required
    return 126
  fi
  if ! rm -f "$go_fifo"; then
    controller_emergency ROLE_START_FAILED "could not unlink private $role authorization channel" recovery_required
    return 126
  fi
  ACTIVE_ROLE_GO_FIFO=""
  python3 -c 'import os,sys; os.setsid(); os.execvp(sys.argv[1], sys.argv[1:])' \
    python3 - "$ready_file" "${auth_file##*/}" "$ROLE_GO_FD" "$ROLE_TOKEN_FD" "$AUTO_LOOP_STATE_FD" \
    "$AUTO_LOOP_CONTROL_FD" "${CONTROL_JSON##*/}" "$CONTROL_FENCE" "$TURN_ID" "$role" "$remaining" \
    "$TURN_DEADLINE_EPOCH" "$@" >>"$LOG_FILE" 2>&1 <<'PY' &
import datetime as dt
import hmac
import json
import os
import re
import select
import stat
import sys
import time
import uuid

(
    ready, authorized_name, go_fd_raw, token_fd_raw, state_fd_raw, root_fd_raw, control_name,
    fence_text, expected_turn, role, remaining_raw, deadline_epoch_raw, command, *arguments
) = sys.argv[1:]
parent_pid = os.getppid()
go_fd = int(go_fd_raw)
token_fd = int(token_fd_raw)
state_fd = int(state_fd_raw)
root_fd = int(root_fd_raw)
remaining = int(remaining_raw)
deadline_epoch = int(deadline_epoch_raw)
if os.getpgrp() != os.getpid():
    raise SystemExit(125)

try:
    go_info = os.fstat(go_fd)
    token_info = os.fstat(token_fd)
    state_info = os.fstat(state_fd)
except OSError:
    raise SystemExit(125)
if not stat.S_ISFIFO(go_info.st_mode) or not stat.S_ISFIFO(token_info.st_mode) or \
   not stat.S_ISDIR(state_info.st_mode):
    raise SystemExit(125)

secret = bytearray()
while len(secret) <= 65:
    chunk = os.read(token_fd, 66 - len(secret))
    if not chunk:
        break
    secret.extend(chunk)
os.close(token_fd)
if len(secret) != 65 or secret[-1:] != b"\n" or not re.fullmatch(b"[0-9a-f]{64}\n", secret):
    raise SystemExit(125)
authorization_token = bytes(secret)
for index in range(len(secret)):
    secret[index] = 0

def write_all(fd, payload):
    written = 0
    while written < len(payload):
        count = os.write(fd, payload[written:])
        if count <= 0:
            raise OSError("short role handshake write")
        written += count

def read_control():
    directory_flags = os.O_RDONLY | getattr(os, "O_DIRECTORY", 0)
    if hasattr(os, "O_NOFOLLOW"):
        directory_flags |= os.O_NOFOLLOW
    try:
        try:
            os.stat(".auto-loop-recovery", dir_fd=root_fd, follow_symlinks=False)
        except FileNotFoundError:
            pass
        else:
            raise OSError("controller state is blocked")
        planning_fd = os.open(".planning", directory_flags, dir_fd=root_fd)
        canonical_fd = os.open("auto-loop", directory_flags, dir_fd=planning_fd)
        held_state = os.fstat(state_fd)
        canonical_state = os.fstat(canonical_fd)
        os.close(canonical_fd)
        os.close(planning_fd)
        if (held_state.st_dev, held_state.st_ino) != (canonical_state.st_dev, canonical_state.st_ino):
            raise OSError("controller state root moved")
    except OSError:
        raise SystemExit(125)
    flags = os.O_RDONLY | getattr(os, "O_NONBLOCK", 0)
    if hasattr(os, "O_NOFOLLOW"):
        flags |= os.O_NOFOLLOW
    try:
        fd = os.open(control_name, flags, dir_fd=state_fd)
        info = os.fstat(fd)
        if not stat.S_ISREG(info.st_mode) or info.st_nlink != 1 or info.st_size > 131072:
            raise OSError("unsafe control state")
        with os.fdopen(fd, "rb") as handle:
            def unique_object(pairs):
                result = {}
                for key, item in pairs:
                    if key in result:
                        raise ValueError("duplicate key")
                    result[key] = item
                return result
            value = json.loads(
                handle.read().decode("utf-8"),
                object_pairs_hook=unique_object,
                parse_constant=lambda item: (_ for _ in ()).throw(ValueError(item)),
            )
    except (OSError, UnicodeError, ValueError):
        raise SystemExit(125)
    root_keys = {
        "schema_version", "run_id", "generation", "controller_id", "control_revision",
        "phase", "lease", "limits", "turn_ordinal", "counters", "active_turn", "halt",
        "children_quiescent", "updated_at",
    }
    if not isinstance(value, dict) or set(value) != root_keys or value.get("schema_version") != "1.0":
        raise SystemExit(125)
    try:
        if str(uuid.UUID(value["run_id"])) != value["run_id"] or \
           str(uuid.UUID(value["controller_id"])) != value["controller_id"]:
            raise ValueError("noncanonical UUID")
        for key in ("generation", "control_revision"):
            if not isinstance(value[key], int) or isinstance(value[key], bool) or value[key] < 1:
                raise ValueError("invalid integer")
        if not isinstance(value["turn_ordinal"], int) or isinstance(value["turn_ordinal"], bool) or \
           value["turn_ordinal"] < 0 or not isinstance(value["children_quiescent"], bool):
            raise ValueError("invalid root scalar")
        if set(value["lease"]) != {"heartbeat_at", "expires_at"} or \
           set(value["limits"]) != {
               "max_turns", "turn_timeout_seconds", "term_grace_seconds", "heartbeat_seconds",
               "max_reverts", "max_no_verdict", "max_minutes",
           } or set(value["counters"]) != {"reverts", "no_verdict", "active_seconds"}:
            raise ValueError("invalid nested shape")
        limits = value["limits"]
        bounds = {
            "max_turns": (1, 1000000), "turn_timeout_seconds": (1, 604800),
            "term_grace_seconds": (1, 300), "heartbeat_seconds": (1, 300),
            "max_reverts": (1, 1000000), "max_no_verdict": (1, 1000000),
            "max_minutes": (0, 525600),
        }
        for key, (minimum, maximum) in bounds.items():
            item = limits[key]
            if not isinstance(item, int) or isinstance(item, bool) or not minimum <= item <= maximum:
                raise ValueError("invalid limit")
        counter_bounds = {"reverts": 1000000, "no_verdict": 1000000, "active_seconds": 315360000}
        for key, maximum in counter_bounds.items():
            item = value["counters"][key]
            if not isinstance(item, int) or isinstance(item, bool) or not 0 <= item <= maximum:
                raise ValueError("invalid counter")
        parse_time = lambda item: dt.datetime.strptime(item, "%Y-%m-%dT%H:%M:%SZ").replace(
            tzinfo=dt.timezone.utc
        )
        heartbeat_at = parse_time(value["lease"]["heartbeat_at"])
        expires_at = parse_time(value["lease"]["expires_at"])
        updated_at = parse_time(value["updated_at"])
        for raw, parsed in (
            (value["lease"]["heartbeat_at"], heartbeat_at),
            (value["lease"]["expires_at"], expires_at),
            (value["updated_at"], updated_at),
        ):
            if parsed.strftime("%Y-%m-%dT%H:%M:%SZ") != raw:
                raise ValueError("noncanonical timestamp")
        if expires_at - heartbeat_at != dt.timedelta(seconds=3 * limits["heartbeat_seconds"]):
            raise ValueError("invalid lease interval")
        if value.get("phase") == "active" and value.get("halt") is not None:
            raise ValueError("active state has a halt")
    except (KeyError, TypeError, ValueError, AttributeError):
        raise SystemExit(125)
    return value

def fence_of(value):
    return "|".join(str(value.get(key, "")) for key in (
        "run_id", "generation", "controller_id", "control_revision"
    ))

def is_exact_binding(value):
    if fence_of(value) != fence_text or value.get("phase") != "active":
        raise SystemExit(125)
    lease_expiry = dt.datetime.strptime(
        value["lease"]["expires_at"], "%Y-%m-%dT%H:%M:%SZ"
    ).replace(tzinfo=dt.timezone.utc)
    if lease_expiry <= dt.datetime.now(dt.timezone.utc):
        return False
    turn = value.get("active_turn")
    turn_keys = {
        "turn_id", "ordinal", "deadline_at", "orchestrator_session_id",
        "validator_session_id", "active_role", "leader_pid", "process_group_id",
    }
    if not isinstance(turn, dict) or set(turn) != turn_keys or turn.get("turn_id") != expected_turn:
        raise SystemExit(125)
    try:
        persisted_deadline = dt.datetime.strptime(
            turn["deadline_at"], "%Y-%m-%dT%H:%M:%SZ"
        ).replace(tzinfo=dt.timezone.utc)
        if int(persisted_deadline.timestamp()) != deadline_epoch or \
           turn.get("ordinal") != value.get("turn_ordinal"):
            raise ValueError("turn mismatch")
        for key in ("turn_id", "orchestrator_session_id", "validator_session_id"):
            if str(uuid.UUID(turn[key])) != turn[key]:
                raise ValueError("noncanonical turn UUID")
    except (KeyError, TypeError, ValueError):
        raise SystemExit(125)
    active_role = turn.get("active_role")
    if active_role is None:
        if turn.get("leader_pid") is not None or turn.get("process_group_id") is not None:
            raise SystemExit(125)
        return False
    if (
        active_role != role
        or turn.get("leader_pid") != os.getpid()
        or turn.get("process_group_id") != os.getpid()
        or value.get("children_quiescent") is not False
    ):
        raise SystemExit(125)
    return True

fd = os.open(ready, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
try:
    write_all(fd, f"{os.getpid()}\n".encode())
    os.fsync(fd)
finally:
    os.close(fd)

# Stay inert until the canonical control record contains this exact fence, turn, role, PID, and
# PGID. Authorization also requires the exact per-role token on the inherited FIFO. Static input
# can never authorize, while malformed/flooded input fails within a bounded frame budget.
wall_remaining = deadline_epoch - int(time.time())
effective_remaining = min(remaining, wall_remaining)
if effective_remaining <= 0:
    raise SystemExit(125)
deadline = time.monotonic() + effective_remaining
frame = bytearray()
total = 0
frames = 0
authorized = False
while True:
    now = time.monotonic()
    if os.getppid() != parent_pid or now >= deadline:
        raise SystemExit(125)
    if not is_exact_binding(read_control()):
        time.sleep(min(0.01, max(0.0, deadline - now)))
        continue
    readable, _, _ = select.select([go_fd], [], [], min(0.05, max(0.0, deadline - now)))
    if not readable:
        continue
    try:
        chunk = os.read(go_fd, min(256, 4096 - total))
    except BlockingIOError:
        continue
    if not chunk:
        raise SystemExit(125)
    total += len(chunk)
    for byte in chunk:
        frame.append(byte)
        if len(frame) > 65:
            raise SystemExit(125)
        if byte == 10:
            frames += 1
            if hmac.compare_digest(bytes(frame), authorization_token):
                authorized = True
                break
            frame.clear()
            if frames >= 64:
                raise SystemExit(125)
    if authorized:
        break
    if total >= 4096:
        raise SystemExit(125)

# Re-derive every authority predicate after consuming the capability. Stale/moved state, parent
# death, malformed input, or an elapsed shared deadline must fail before provider execution.
if os.getppid() != parent_pid or time.monotonic() >= deadline or int(time.time()) >= deadline_epoch \
   or not is_exact_binding(read_control()):
    raise SystemExit(125)
authorization_flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL
if hasattr(os, "O_NOFOLLOW"):
    authorization_flags |= os.O_NOFOLLOW
authorization_fd = os.open(authorized_name, authorization_flags, 0o600, dir_fd=state_fd)
try:
    write_all(authorization_fd, f"{os.getpid()}\n".encode())
    os.fsync(authorization_fd)
finally:
    os.close(authorization_fd)
if os.getppid() != parent_pid or time.monotonic() >= deadline or int(time.time()) >= deadline_epoch \
   or not is_exact_binding(read_control()):
    raise SystemExit(125)
os.close(go_fd)
os.close(state_fd)
os.environ.pop("AUTO_LOOP_STATE_FD", None)
os.execvp(command, [command, *arguments])
PY
  pid=$!
  exec 18<&-
  ACTIVE_ROLE_TOKEN_FD_OPEN=0
  ACTIVE_ROLE_PID="$pid"
  ACTIVE_ROLE_PGID="$pid"
  ACTIVE_ROLE_GROUP_READY=0
  while true; do
    [[ -s "$ready_file" ]] && break
    leader_process_exited "$pid" && break
    if (( SECONDS >= TURN_DEADLINE_MONO )); then
      controller_emergency TURN_DEADLINE "$role exhausted the persisted turn deadline during startup" halted
      return 124
    fi
    sleep 0.01
  done
  pgid="$(cat "$ready_file" 2>/dev/null || true)"
  rm -f "$ready_file"
  ACTIVE_ROLE_READY_FILE=""
  observed_pgid="$(ps -o pgid= -p "$pid" 2>/dev/null | tr -d '[:space:]' || true)"
  if [[ ! "$pgid" =~ ^[0-9]+$ || "$pgid" -ne "$pid" || "$observed_pgid" != "$pid" ]] || \
     leader_process_exited "$pid"; then
    controller_emergency ROLE_START_FAILED "could not establish $role process group" recovery_required
    return 126
  fi
  ACTIVE_ROLE_PGID="$pgid"
  ACTIVE_ROLE_GROUP_READY=1
  control_state_before_turn_deadline bind "$role" "$pid" "$pgid" >/dev/null || bind_rc=$?
  if (( bind_rc == 124 )); then
    controller_emergency TURN_DEADLINE "$role bind exhausted the persisted turn deadline" halted
    return 124
  elif (( bind_rc != 0 )); then
    controller_emergency CONTROL_FENCE_MISMATCH "could not bind $role process group" recovery_required
    return 126
  fi
  if (( SECONDS >= TURN_DEADLINE_MONO )); then
    controller_emergency TURN_DEADLINE "$role exhausted the persisted turn deadline before authorization" halted
    return 124
  fi
  control_state_before_turn_deadline assert >/dev/null || assert_rc=$?
  if (( assert_rc == 124 || SECONDS >= TURN_DEADLINE_MONO )); then
    controller_emergency TURN_DEADLINE "$role assertion exhausted the persisted turn deadline" halted
    return 124
  elif (( assert_rc != 0 )); then
    controller_emergency CONTROL_FENCE_MISMATCH "could not reassert $role authority" recovery_required
    return 126
  fi
  if ! printf '%s\n' "$role_token" >&19; then
    controller_emergency ROLE_START_FAILED "could not authorize durably bound $role process group" recovery_required
    return 126
  fi
  role_token=""
  exec 19>&-
  ACTIVE_ROLE_GO_FD_OPEN=0
  return 0
}

role_authorization_acknowledged() {
  python3 - "$AUTO_LOOP_STATE_FD" "${ACTIVE_ROLE_AUTH_FILE##*/}" "$ACTIVE_ROLE_PID" <<'PY'
import os
import stat
import sys

state_fd = int(sys.argv[1])
name = sys.argv[2]
expected = (sys.argv[3] + "\n").encode()
flags = os.O_RDONLY | getattr(os, "O_NONBLOCK", 0)
if hasattr(os, "O_NOFOLLOW"):
    flags |= os.O_NOFOLLOW
try:
    fd = os.open(name, flags, dir_fd=state_fd)
    info = os.fstat(fd)
    if not stat.S_ISREG(info.st_mode) or info.st_nlink != 1 or info.st_size != len(expected):
        raise OSError("unsafe authorization acknowledgement")
    with os.fdopen(fd, "rb") as handle:
        observed = handle.read(len(expected) + 1)
except OSError:
    raise SystemExit(1)
raise SystemExit(0 if observed == expected else 1)
PY
}

supervise_role_group() { # role
  local role="$1" role_code rc=0 heartbeat_rc=0 clear_rc=0 \
    next_heartbeat=$((SECONDS + CONTROL_HEARTBEAT_SECONDS)) leader_reaped=0
  case "$role" in
    orchestrator) role_code="ORCHESTRATOR" ;;
    validator) role_code="VALIDATOR" ;;
    trace) role_code="TRACE" ;;
    *) role_code="ROLE" ;;
  esac
  while true; do
    if leader_process_exited "$ACTIVE_ROLE_PID"; then
      if ! role_authorization_acknowledged; then
        controller_emergency ROLE_AUTH_FAILED "$role exited before consuming controller authorization" recovery_required
        return 126
      fi
      if process_group_alive "$ACTIVE_ROLE_PGID"; then
        controller_emergency "${role_code}_ORPHAN" "$role leader exited with live descendants" halted
        return 125
      fi
      wait "$ACTIVE_ROLE_PID" 2>/dev/null || rc=$?
      leader_reaped=1
      break
    fi
    if (( SECONDS >= TURN_DEADLINE_MONO )); then
      controller_emergency TURN_DEADLINE "$role exceeded the persisted turn deadline" halted
      return 124
    fi
    if (( SECONDS >= next_heartbeat )); then
      heartbeat_rc=0
      control_state_before_turn_deadline heartbeat >/dev/null || heartbeat_rc=$?
      if (( heartbeat_rc == 124 )); then
        controller_emergency TURN_DEADLINE "$role heartbeat exhausted the persisted turn deadline" halted
        return 124
      elif (( heartbeat_rc != 0 )); then
        controller_emergency CONTROL_FENCE_MISMATCH "$role lost controller authority" recovery_required
        return 126
      fi
      next_heartbeat=$((SECONDS + CONTROL_HEARTBEAT_SECONDS))
    fi
    sleep 0.1
  done
  if process_group_alive "$ACTIVE_ROLE_PGID"; then
    controller_emergency "${role_code}_ORPHAN" "$role group remained live after leader completion" halted
    return 125
  fi
  clear_rc=0
  control_state_before_turn_deadline clear-role "$role" >/dev/null || clear_rc=$?
  if (( clear_rc == 124 )); then
    controller_emergency TURN_DEADLINE "$role cleanup exhausted the persisted turn deadline" halted
    return 124
  elif (( clear_rc != 0 )); then
    controller_emergency CONTROL_FENCE_MISMATCH "could not clear $role authority" recovery_required
    return 126
  fi
  ACTIVE_ROLE_PID=""
  ACTIVE_ROLE_PGID=""
  ACTIVE_ROLE_GROUP_READY=0
  cleanup_role_artifacts
  LAST_ROLE_EXIT_STATUS="$rc"
  return 0
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

run_trace() {
  start_role_group trace "$REPO_ROOT/scripts/loop-trace.sh" distill || return $?
  supervise_role_group trace
}

checkpoint_artifact_exec() { # publish|restore run-id checkpoint-ordinal
  exec python3 -c 'import os,sys; os.setsid(); os.execvp(sys.argv[1], sys.argv[1:])' \
    python3 - "$AUTO_LOOP_STATE_FD" "$REPO_ROOT" "$1" "$2" "$3" <<'PY'
import json
import os
import stat
import subprocess
import sys
import uuid

state_fd = int(sys.argv[1])
repo_root = sys.argv[2]
command = sys.argv[3]
run_id = str(uuid.UUID(sys.argv[4]))
target = sys.argv[5]
if not target.isdecimal() or target == "0" or str(int(target)) != target:
    raise SystemExit("CHECKPOINT_TARGET_INVALID")

directory_flags = os.O_RDONLY | getattr(os, "O_DIRECTORY", 0)
if hasattr(os, "O_NOFOLLOW"):
    directory_flags |= os.O_NOFOLLOW

def unique_object(pairs):
    result = {}
    for key, item in pairs:
        if key in result:
            raise ValueError("duplicate key")
        result[key] = item
    return result

def decode_run(payload):
    value = json.loads(
        payload.decode("utf-8"), object_pairs_hook=unique_object,
        parse_constant=lambda item: (_ for _ in ()).throw(ValueError(item)),
    )
    if not isinstance(value, dict):
        raise ValueError("RUN.json must contain an object")
    return value

def read_regular_at(parent_fd, name, maximum):
    flags = os.O_RDONLY | getattr(os, "O_NONBLOCK", 0)
    if hasattr(os, "O_NOFOLLOW"):
        flags |= os.O_NOFOLLOW
    fd = os.open(name, flags, dir_fd=parent_fd)
    try:
        before = os.fstat(fd)
        if (
            not stat.S_ISREG(before.st_mode) or before.st_nlink != 1
            or before.st_uid != os.geteuid() or before.st_size > maximum
        ):
            raise ValueError(f"unsafe checkpoint artifact: {name}")
        chunks = []
        remaining = maximum + 1
        while remaining:
            chunk = os.read(fd, remaining)
            if not chunk:
                break
            chunks.append(chunk)
            remaining -= len(chunk)
        payload = b"".join(chunks)
        after = os.fstat(fd)
        if (
            len(payload) > maximum or len(payload) != after.st_size
            or (before.st_dev, before.st_ino, before.st_size, before.st_mtime_ns)
            != (after.st_dev, after.st_ino, after.st_size, after.st_mtime_ns)
        ):
            raise ValueError(f"checkpoint artifact changed during snapshot: {name}")
        return payload
    finally:
        os.close(fd)

def open_private_dir(parent_fd, name, create=False):
    if create:
        try:
            os.mkdir(name, 0o700, dir_fd=parent_fd)
            os.fsync(parent_fd)
        except FileExistsError:
            pass
    fd = os.open(name, directory_flags, dir_fd=parent_fd)
    info = os.fstat(fd)
    if (
        not stat.S_ISDIR(info.st_mode) or info.st_uid != os.geteuid()
        or stat.S_IMODE(info.st_mode) & 0o022
    ):
        os.close(fd)
        raise ValueError(f"unsafe checkpoint directory: {name}")
    return fd

def write_all(fd, payload):
    offset = 0
    while offset < len(payload):
        offset += os.write(fd, payload[offset:])

def write_exclusive(parent_fd, name, payload):
    flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL
    if hasattr(os, "O_NOFOLLOW"):
        flags |= os.O_NOFOLLOW
    fd = os.open(name, flags, 0o600, dir_fd=parent_fd)
    try:
        write_all(fd, payload)
        os.fsync(fd)
    finally:
        os.close(fd)

def atomic_write(parent_fd, name, payload):
    temporary = f".{name}.tmp.{os.getpid()}.{uuid.uuid4().hex}"
    flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL
    if hasattr(os, "O_NOFOLLOW"):
        flags |= os.O_NOFOLLOW
    fd = os.open(temporary, flags, 0o600, dir_fd=parent_fd)
    try:
        write_all(fd, payload)
        os.fsync(fd)
    finally:
        os.close(fd)
    try:
        os.replace(temporary, name, src_dir_fd=parent_fd, dst_dir_fd=parent_fd)
        os.fsync(parent_fd)
    except Exception:
        try:
            os.unlink(temporary, dir_fd=parent_fd)
        except OSError:
            pass
        raise

def git_head():
    result = subprocess.run(
        ["git", "-C", repo_root, "rev-parse", "HEAD"],
        stdin=subprocess.DEVNULL, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL,
        check=True, timeout=10,
    )
    head = result.stdout.decode("ascii").strip()
    if len(head) not in (40, 64) or any(character not in "0123456789abcdef" for character in head):
        raise ValueError("invalid git HEAD")
    return head

opened = []
try:
    checkpoints_fd = open_private_dir(state_fd, "checkpoints", create=command == "publish")
    opened.append(checkpoints_fd)
    run_fd = open_private_dir(checkpoints_fd, f"run-{run_id}", create=command == "publish")
    opened.append(run_fd)

    if command == "publish":
        run_bytes = read_regular_at(state_fd, "RUN.json", 1048576)
        decode_run(run_bytes)
        head = git_head()
        os.mkdir(target, 0o700, dir_fd=run_fd)
        os.fsync(run_fd)
        target_fd = open_private_dir(run_fd, target)
        opened.append(target_fd)
        write_exclusive(target_fd, "RUN.json", run_bytes)
        write_exclusive(target_fd, "HEAD.sha", (head + "\n").encode("ascii"))
        os.fsync(target_fd)
        atomic_write(run_fd, "LAST_GOOD", (target + "\n").encode("ascii"))
        print(target)
    elif command == "restore":
        marker = read_regular_at(run_fd, "LAST_GOOD", 32).decode("ascii").strip()
        if marker != target:
            raise ValueError("checkpoint marker changed")
        target_fd = open_private_dir(run_fd, target)
        opened.append(target_fd)
        run_bytes = read_regular_at(target_fd, "RUN.json", 1048576)
        decode_run(run_bytes)
        good_head = read_regular_at(target_fd, "HEAD.sha", 128).decode("ascii").strip()
        if len(good_head) not in (40, 64) or any(
            character not in "0123456789abcdef" for character in good_head
        ):
            raise ValueError("invalid checkpoint HEAD")
        current_head = git_head()
        cleanup = json.dumps({
            "good_fork_sha": good_head,
            "diverged_head_sha": current_head,
            "checkpoint": target,
            "run_id": run_id,
            "instruction": "REVERT: reset local-only commits after good_fork_sha, or revert-forward pushed commits per your gates; never force-push. Then replay the stage.",
        }, sort_keys=True, indent=2).encode("utf-8") + b"\n"
        atomic_write(state_fd, "RUN.json", run_bytes)
        atomic_write(state_fd, "REVERT-CLEANUP.json", cleanup)
        print(good_head)
    else:
        raise ValueError("unknown checkpoint command")
except (OSError, UnicodeError, ValueError, subprocess.SubprocessError) as exc:
    print(f"CHECKPOINT_{command.upper()}_FAILED: {exc}", file=sys.stderr)
    raise SystemExit(4)
finally:
    for descriptor in reversed(opened):
        os.close(descriptor)
PY
}

checkpoint_artifact_before_turn_deadline() { # publish|restore checkpoint-ordinal
  exec_before_monotonic_deadline "$TURN_DEADLINE_MONO" checkpoint_artifact_exec \
    "$1" "$RUN_ID" "$2"
}

controller_publish_checkpoint() {
  local target="$1" rc=0
  checkpoint_artifact_before_turn_deadline publish "$target" >/dev/null || rc=$?
  if (( rc == 124 )); then
    controller_emergency TURN_DEADLINE "checkpoint publication exhausted the persisted turn deadline" halted
    return 4
  elif (( rc != 0 )); then
    controller_emergency CHECKPOINT_PUBLISH_FAILED "checkpoint bundle could not be atomically published" recovery_required
    return 4
  fi
}

controller_restore_checkpoint() {
  local target="$1" rc=0 good_sha
  good_sha="$(checkpoint_artifact_before_turn_deadline restore "$target")" || rc=$?
  if (( rc == 124 )); then
    controller_emergency TURN_DEADLINE "checkpoint restoration exhausted the persisted turn deadline" halted
    return 4
  elif (( rc != 0 )); then
    controller_emergency CHECKPOINT_RESTORE_FAILED "checkpoint bundle could not be atomically restored" recovery_required
    return 4
  fi
  log "restored checkpoint $target for run $RUN_ID (fork ${good_sha:0:8}); wrote REVERT-CLEANUP.json"
}

correction=""
for (( i=1; i<=MAX_ITERATIONS; i++ )); do
  if (( MAX_MINUTES > 0 )) && (( $(controller_active_seconds) >= MAX_MINUTES * 60 )); then
    controller_pause
    log "STOP: durable active-time cap ${MAX_MINUTES}m reached."; exit 3
  fi

  TURN_DEADLINE_MONO=$((SECONDS + TURN_TIMEOUT_SECONDS))
  turn_output_path="$(mktemp "$STATE_DIR/.begin-output.XXXXXX")"
  if [[ "${AUTO_LOOP_CONTROL_FD:-}" == "16" || "${AUTO_LOOP_CONTROL_FD:-}" == "17" || \
        "${AUTO_LOOP_STATE_FD:-}" == "16" || "${AUTO_LOOP_STATE_FD:-}" == "17" ]] || \
     ! exec 17>"$turn_output_path" || ! exec 16<"$turn_output_path" || \
     ! rm -f "$turn_output_path"; then
    controller_emergency CONTROL_BEGIN_FAILED "could not create a private turn-allocation result channel" recovery_required
    exit 4
  fi
  if control_state_before_turn_deadline begin >&17; then
    exec 17>&-
    IFS=$'\t' read -r GLOBAL_TURN TURN_ID ORCHESTRATOR_SESSION_ID VALIDATOR_SESSION_ID \
      TURN_DEADLINE_EPOCH <&16
    exec 16<&-
  else
    turn_rc=$?
    exec 17>&-
    exec 16<&-
    if (( turn_rc == 3 )); then
      controller_pause
      log "STOP: persisted turn cap reached; a later human-authority phase must close or extend this run."
      exit 3
    fi
    if (( turn_rc == 124 )); then
      controller_emergency TURN_DEADLINE "turn allocation exhausted the persisted turn deadline" halted
      exit 4
    fi
    controller_emergency CONTROL_BEGIN_FAILED "could not allocate a fenced turn" recovery_required
    exit 4
  fi
  if ! python3 - "$GLOBAL_TURN" "$TURN_ID" "$ORCHESTRATOR_SESSION_ID" \
      "$VALIDATOR_SESSION_ID" <<'PY'
import sys
import uuid

try:
    ordinal = int(sys.argv[1])
    if str(ordinal) != sys.argv[1] or ordinal < 1:
        raise ValueError("invalid ordinal")
    for raw in sys.argv[2:]:
        if str(uuid.UUID(raw)) != raw:
            raise ValueError("noncanonical UUID")
except (TypeError, ValueError):
    raise SystemExit(1)
PY
  then
    controller_emergency CONTROL_BEGIN_FAILED "turn allocation returned malformed identity fields" recovery_required
    exit 4
  fi
  TURN_NOW_EPOCH="$(date -u +%s)"
  if [[ ! "$TURN_DEADLINE_EPOCH" =~ ^[0-9]+$ || ! "$TURN_NOW_EPOCH" =~ ^[0-9]+$ ]]; then
    controller_emergency CONTROL_BEGIN_FAILED "turn allocation returned an invalid persisted deadline" recovery_required
    exit 4
  fi
  TURN_REMAINING=$((TURN_DEADLINE_EPOCH - TURN_NOW_EPOCH))
  if (( TURN_REMAINING > TURN_TIMEOUT_SECONDS )); then
    controller_emergency CONTROL_CLOCK_ROLLBACK "wall clock moved behind the persisted turn deadline anchor" recovery_required
    exit 4
  fi
  if (( SECONDS + TURN_REMAINING < TURN_DEADLINE_MONO )); then
    TURN_DEADLINE_MONO=$((SECONDS + TURN_REMAINING))
  fi
  if (( TURN_REMAINING <= 0 || SECONDS >= TURN_DEADLINE_MONO )); then
    controller_emergency TURN_DEADLINE "turn allocation returned after its persisted deadline" halted
    exit 4
  fi

  log "── turn $GLOBAL_TURN: ORCHESTRATOR ──${correction:+ (with correction)}"
  turn_msg="$LOOP_CMD $PROBLEM"
  if [[ -n "$correction" ]]; then
    turn_msg="$turn_msg

VALIDATOR CORRECTION (apply first): $correction"
  fi
  if run_orchestrator "$turn_msg"; then
    orchestrator_rc="$LAST_ROLE_EXIT_STATUS"
    if (( orchestrator_rc != 0 )); then
      log "turn $GLOBAL_TURN: orchestrator returned $orchestrator_rc (validator will assess)"
    fi
  else
    orchestrator_rc=$?
    exit 4
  fi

  # Distill the orchestrator session before validation so the Shepherd judges the complete turn.
  # No mutable worktree script runs between validator quiescence and the private verdict snapshot.
  if run_trace; then
    trace_rc="$LAST_ROLE_EXIT_STATUS"
    if (( trace_rc == 0 )); then
      log "turn $GLOBAL_TURN: trace digest written (see .planning/auto-loop/trace/INDEX.md)"
    else
      log "turn $GLOBAL_TURN: trace digest returned $trace_rc (continuing after supervised quiescence)"
    fi
  else
    trace_rc=$?
    exit 4
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
    validator_rc="$LAST_ROLE_EXIT_STATUS"
    if (( validator_rc != 0 )); then
      log "turn $GLOBAL_TURN: validator returned $validator_rc"
    fi
  else
    validator_rc=$?
    exit 4
  fi
  if (( SECONDS >= TURN_DEADLINE_MONO )); then
    controller_emergency TURN_DEADLINE "verdict parsing started after the persisted turn deadline" halted
    exit 4
  fi
  verdict_output_path="$(mktemp "$STATE_DIR/.verdict-output.XXXXXX")"
  if [[ "${AUTO_LOOP_CONTROL_FD:-}" == "16" || "${AUTO_LOOP_CONTROL_FD:-}" == "17" || \
        "${AUTO_LOOP_STATE_FD:-}" == "16" || "${AUTO_LOOP_STATE_FD:-}" == "17" ]] || \
     ! exec 17>"$verdict_output_path" || ! exec 16<"$verdict_output_path" || \
     ! rm -f "$verdict_output_path"; then
    controller_emergency VERDICT_SLOT_UNSAFE "could not create a private verdict snapshot channel" recovery_required
    exit 4
  fi
  verdict_read_rc=0
  if (( validator_rc == 0 )); then
    exec_before_monotonic_deadline "$TURN_DEADLINE_MONO" verdict_snapshot_exec \
      "$AUTO_LOOP_STATE_FD" "$RUN_ID" >&17 || verdict_read_rc=$?
  else
    printf '||||\n' >&17 || verdict_read_rc=$?
  fi
  exec 17>&-
  if (( verdict_read_rc == 0 )); then
    IFS='|' read -r verdict score reason_hex correction_hex revert_checkpoint <&16 || verdict_read_rc=$?
  fi
  exec 16<&-
  if (( verdict_read_rc != 0 )); then
    if (( verdict_read_rc == 124 )); then
      controller_emergency TURN_DEADLINE "verdict parsing could not finish inside the persisted turn deadline" halted
    else
      controller_emergency VERDICT_SNAPSHOT_FAILED "verdict artifact could not be safely snapshotted" recovery_required
    fi
    exit 4
  fi
  reason="$(decode_hex_text "$reason_hex")"
  verdict_correction="$(decode_hex_text "$correction_hex")"
  if (( SECONDS >= TURN_DEADLINE_MONO )); then
    controller_emergency TURN_DEADLINE "verdict parsing exceeded the persisted turn deadline" halted
    exit 4
  fi
  correction=""
  log "turn $GLOBAL_TURN: verdict=${verdict:-NONE} step_score=${score:-?} — ${reason:-}"

  case "$verdict" in
    PROCEED)
      no_verdict=0; controller_set_counter no_verdict "$no_verdict"
      controller_publish_checkpoint "$GLOBAL_TURN"
      controller_complete_turn ;;
    RETRY)
      no_verdict=0; controller_set_counter no_verdict "$no_verdict"
      controller_complete_turn; correction="$verdict_correction"; log "turn $GLOBAL_TURN: RETRY — $correction" ;;
    REVERT)
      no_verdict=0; reverts=$((reverts+1))
      if (( reverts > MAX_REVERTS )); then
        controller_emergency MAX_REVERTS "MAX_REVERTS=$MAX_REVERTS exceeded" halted
        log "HALT: MAX_REVERTS=$MAX_REVERTS exceeded"; exit 4
      fi
      controller_set_counter no_verdict "$no_verdict"
      controller_set_counter reverts "$reverts"
      controller_restore_checkpoint "$revert_checkpoint"
      controller_complete_turn
      correction="$verdict_correction"; log "turn $GLOBAL_TURN: REVERT #$reverts to checkpoint $revert_checkpoint — $correction" ;;
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
      controller_complete_turn
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
