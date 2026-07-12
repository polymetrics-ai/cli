#!/usr/bin/env bash
set -u

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REAL_PYTHON_BIN="$(command -v python3)"
REAL_PS_BIN="$(command -v ps)"
REAL_MKFIFO_BIN="$(command -v mkfifo)"
REAL_GIT_BIN="$(command -v git)"
TEST_TMP="$(mktemp -d)"
failures=0
executed_tests=0
SUITE_PID=$$
WATCHDOG_PID=""
SUITE_TIMEOUT_SECONDS="${SUITE_TIMEOUT_SECONDS:-600}"

owned_process_alive() {
  local pid="$1" nonce="$2" command state
  [[ "$pid" =~ ^[0-9]+$ ]] || return 1
  kill -0 "$pid" 2>/dev/null || return 1
  command="$(ps -p "$pid" -o command= 2>/dev/null || true)"
  state="$(ps -p "$pid" -o stat= 2>/dev/null || true)"
  [[ -n "$command" && "$command" == *"$nonce"* && "$state" != Z* ]]
}

kill_owned_process() {
  local pid="$1" nonce="$2"
  if owned_process_alive "$pid" "$nonce"; then
    kill -KILL "$pid" 2>/dev/null || true
    wait "$pid" 2>/dev/null || true
  fi
}

signal_owned_process() {
  local signal="$1" pid="$2" nonce="$3"
  if owned_process_alive "$pid" "$nonce"; then
    kill -"$signal" "$pid" 2>/dev/null || true
  fi
}

cleanup() {
  local pid_file pid nonce
  if [[ -n "$WATCHDOG_PID" ]]; then
    kill "$WATCHDOG_PID" 2>/dev/null || true
    wait "$WATCHDOG_PID" 2>/dev/null || true
  fi
  while IFS= read -r pid_file; do
    [[ -f "$pid_file" ]] || continue
    while read -r pid nonce; do
      [[ -n "${nonce:-}" ]] || continue
      kill_owned_process "$pid" "$nonce"
    done <"$pid_file"
  done < <(find "$TEST_TMP" \( -name child-pids -o -name controller-pids -o -name model-pids -o -name role-pids \) -type f 2>/dev/null)
  if [[ "${KEEP_TEST_TMP:-0}" == "1" ]]; then
    printf 'pi-shepherd-supervision: preserved fixtures at %s\n' "$TEST_TMP" >&2
  else
    rm -rf "$TEST_TMP"
  fi
}
trap cleanup EXIT
trap 'cleanup; exit 124' INT TERM

(
  /bin/sleep "$SUITE_TIMEOUT_SECONDS"
  kill -TERM "$SUITE_PID" 2>/dev/null || true
) &
WATCHDOG_PID=$!

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  failures=$((failures + 1))
}

run_test() {
  local name="$1" rc=0
  if [[ -n "${SHEPHERD_TEST_FILTER:-}" && ",${SHEPHERD_TEST_FILTER}," != *",${name},"* ]]; then
    return
  fi
  if ! declare -F "$name" >/dev/null; then
    fail "missing test function: $name"
    return
  fi
  executed_tests=$((executed_tests + 1))
  "$name" || rc=$?
  if [[ "$rc" -ne 0 ]]; then
    fail "$name returned unexpected status $rc"
  fi
}

copy_launcher_without_phase0_guard() {
  local destination="$1"
  # The production fuse has no enable route. Tests exercise the exact post-guard launcher body by
  # removing only the explicitly delimited guard block from a temporary copy; no production flag
  # or environment bypass exists.
  if [[ "$(grep -c '^# AUTO_LOOP_PHASE0_GUARD_BEGIN$' "$REPO_ROOT/scripts/pi-shepherd-loop.sh")" -ne 1 ]] ||
     [[ "$(grep -c '^# AUTO_LOOP_PHASE0_GUARD_END$' "$REPO_ROOT/scripts/pi-shepherd-loop.sh")" -ne 1 ]]; then
    fail "canonical launcher does not have exactly one Phase 0 guard sentinel pair"
    return 1
  fi
  awk '
    /^# AUTO_LOOP_PHASE0_GUARD_BEGIN$/ { skipping=1; seen_begin++ }
    /^# AUTO_LOOP_PHASE0_GUARD_END$/ { skipping=0; seen_end++; next }
    !skipping { print }
    END { if (seen_begin != 1 || seen_end != 1) exit 1 }
  ' "$REPO_ROOT/scripts/pi-shepherd-loop.sh" >"$destination"
  chmod +x "$destination"
}

prepare_fixture() {
  local root="$1"
  mkdir -p \
    "$root/scripts/tests" \
    "$root/bin" \
    "$root/.pi/extensions/pi-sub-agent" \
    "$root/.agents/agentic-delivery/prompts" \
    "$root/.planning/auto-loop" \
    "$root/home"

  copy_launcher_without_phase0_guard "$root/scripts/pi-shepherd-loop.sh"
  : >"$root/.pi/extensions/pi-sub-agent/index.ts"
  printf 'synthetic validator prompt\n' >"$root/.agents/agentic-delivery/prompts/shepherd-validator-prompt.md"
  printf '{"stage":"PLAN","terminal":null}\n' >"$root/.planning/auto-loop/RUN.json"
  : >"$root/events"
  : >"$root/child-pids"
  : >"$root/model-pids"
  : >"$root/controller-pids"
  : >"$root/role-pids"
  printf 'shepherd-test-%s-%s-%s\n' "$$" "$RANDOM" "${root##*/}" >"$root/nonce"

  cat >"$root/scripts/loop-trace.sh" <<'SH'
#!/usr/bin/env bash
set -u
if [[ "${FAKE_MODE:-}" == "halt-latch-failure" ]]; then
  mv "$TEST_REPO/.planning/auto-loop/CONTROL.json" "$TEST_REPO/.planning/auto-loop/CONTROL.backup"
  ln -s CONTROL.backup "$TEST_REPO/.planning/auto-loop/CONTROL.json"
fi
if [[ "${FAKE_TRACE_OVERWRITE:-0}" == "1" ]]; then
  printf '%s\n' '{"verdict":"PROCEED","step_score":5,"trajectory_geomean":5,"reason":"trace-forged evidence","correction":null,"revert_to_checkpoint":null}' \
    >"$TEST_REPO/.planning/auto-loop/VALIDATOR-VERDICT.json"
fi
if [[ "${FAKE_TRACE_DELAY:-0}" != "0" ]]; then
  printf 'tracing\n' >"$TRACE_PROBE_FILE"
  child_nonce="$TEST_NONCE-trace-$$"
  "$TEST_REPO/bin/test-child" "$child_nonce" "$FAKE_TRACE_DELAY" &
  child=$!
  printf '%s %s\n' "$child" "$child_nonce" >>"$CHILD_PID_FILE"
  wait "$child"
fi
exit "${FAKE_TRACE_EXIT:-0}"
SH
  chmod +x "$root/scripts/loop-trace.sh"

  cat >"$root/bin/sleep" <<'SH'
#!/usr/bin/env bash
exec /bin/sleep 0.02
SH
  chmod +x "$root/bin/sleep"

  cat >"$root/bin/test-child" <<'SH'
#!/usr/bin/env bash
set -u
nonce="$1"
duration="$2"
if [[ -n "${DESCENDANT_LOCK_READY_FILE:-}" ]]; then
  "$REAL_PYTHON_BIN" - "$AUTO_LOOP_CONTROL_FD" "$TEST_REPO" \
    "$DESCENDANT_LOCK_READY_FILE" "$nonce" <<'PY'
import os
import pathlib
import stat
import sys

fd = int(sys.argv[1])
lock_path = sys.argv[2]
marker = pathlib.Path(sys.argv[3])
nonce = sys.argv[4]
held = os.fstat(fd)
canonical = os.stat(lock_path, follow_symlinks=False)
if not stat.S_ISDIR(held.st_mode) or (held.st_dev, held.st_ino) != (canonical.st_dev, canonical.st_ino):
    raise SystemExit("descendant inherited the wrong controller lock")
marker_fd = os.open(marker, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
try:
    os.write(marker_fd, (nonce + "\n").encode("utf-8"))
    os.fsync(marker_fd)
finally:
    os.close(marker_fd)
PY
fi
exec "$REAL_PYTHON_BIN" -c '
import pathlib, signal, sys, time
signal.signal(signal.SIGTERM, signal.SIG_IGN)
time.sleep(float(sys.argv[2]))
if sys.argv[3]:
    with pathlib.Path(sys.argv[3]).open("a") as handle:
        handle.write(sys.argv[1] + "\n")
' "$nonce" "$duration" "${NATURAL_EXIT_LOG:-}"
SH
  chmod +x "$root/bin/test-child"

  cat >"$root/bin/test-bystander" <<'SH'
#!/usr/bin/env bash
set -u
nonce="$1"
duration="$2"
signal_log="$3"
trap 'printf "%s\n" "$nonce" >>"$signal_log"' TERM
deadline=$((SECONDS + duration))
while (( SECONDS < deadline )); do
  /bin/sleep 0.2 || true
done
SH
  chmod +x "$root/bin/test-bystander"

  cat >"$root/bin/python3" <<'SH'
#!/usr/bin/env bash
set -u
control_command=""
wrapper_arguments=("$@")
if [[ "${1:-}" == "-" ]]; then
  for ((wrapper_index=0; wrapper_index<${#wrapper_arguments[@]}-1; wrapper_index++)); do
    if [[ "${wrapper_arguments[$wrapper_index]}" == *'/CONTROL.json' ]]; then
      control_command="${wrapper_arguments[$((wrapper_index + 1))]}"
      break
    fi
  done
fi
if [[ "${2:-}" == *'.role-ready.'* ]]; then
  printf '%s %s\n' "$$" "$2" >>"$ROLE_PID_FILE"
fi
if [[ "${2:-}" == *'.role-ready.'* && "${FAKE_EARLY_READY_DELAY:-0}" != "0" ]]; then
  printf '%s\n' "$$" >"$2"
  printf 'early-ready\n' >"$EARLY_READY_PROBE_FILE"
  child_nonce="$TEST_NONCE-false-ready-$$"
  "$TEST_REPO/bin/test-child" "$child_nonce" "$FAKE_EARLY_READY_DELAY" &
  child=$!
  printf '%s %s\n' "$child" "$child_nonce" >>"$CHILD_PID_FILE"
  wait "$child"
fi
if [[ "${2:-}" == *'.role-ready.'* && "${FAKE_LAUNCH_DELAY:-0}" != "0" ]]; then
  printf 'launching\n' >"$LAUNCH_PROBE_FILE"
  /bin/sleep "$FAKE_LAUNCH_DELAY"
fi
if [[ "$control_command" == "begin" && "${FAKE_BEGIN_RETURN_DELAY:-0}" != "0" ]]; then
  printf '%s %s\n' "$$" "$TEST_REPO/bin/python3" >>"$CONTROLLER_PID_FILE"
  output="$("$REAL_PYTHON_BIN" "$@")"
  rc=$?
  printf 'committed\n' >"$BEGIN_COMMIT_PROBE_FILE"
  child_nonce="$TEST_NONCE-control-begin-$$"
  "$TEST_REPO/bin/test-child" "$child_nonce" "$FAKE_BEGIN_RETURN_DELAY" &
  child=$!
  printf '%s %s\n' "$child" "$child_nonce" >>"$CHILD_PID_FILE"
  wait "$child"
  printf '%s\n' "$output"
  exit "$rc"
fi
if [[ "$control_command" == "bind" && "${FAKE_CONTROL_ORPHAN:-0}" == "1" ]]; then
  child_nonce="$TEST_NONCE-control-orphan-$$"
  "$TEST_REPO/bin/test-child" "$child_nonce" 8 &
  child=$!
  printf '%s %s\n' "$child" "$child_nonce" >>"$CHILD_PID_FILE"
  exec "$REAL_PYTHON_BIN" "$@"
fi
if [[ "$control_command" == "bind" && "${FAKE_BIND_DELAY:-0}" != "0" ]]; then
  printf '%s %s\n' "$$" "$TEST_REPO/bin/python3" >>"$CONTROLLER_PID_FILE"
  printf 'binding\n' >"$BIND_PROBE_FILE"
  /bin/sleep "$FAKE_BIND_DELAY"
  "$REAL_PYTHON_BIN" "$@"
  exit $?
fi
if [[ "$control_command" == "assert" && "${FAKE_ASSERT_DELAY:-0}" != "0" ]]; then
  printf '%s %s\n' "$$" "$TEST_REPO/bin/python3" >>"$CONTROLLER_PID_FILE"
  printf 'asserting\n' >"$ASSERT_PROBE_FILE"
  /bin/sleep "$FAKE_ASSERT_DELAY"
  "$REAL_PYTHON_BIN" "$@"
  rc=$?
  printf 'completed\n' >"$ASSERT_COMPLETION_FILE"
  exit "$rc"
fi
if [[ "$control_command" == "assert" && "${FAKE_ASSERT_BARRIER:-0}" == "1" ]]; then
  printf '%s %s\n' "$$" "$TEST_REPO/bin/python3" >>"$CONTROLLER_PID_FILE"
  printf 'asserting\n' >"$ASSERT_PROBE_FILE"
  while [[ ! -e "$ASSERT_RELEASE_FILE" ]]; do
    /bin/sleep 0.01
  done
  "$REAL_PYTHON_BIN" "$@"
  rc=$?
  printf 'completed\n' >"$ASSERT_COMPLETION_FILE"
  exit "$rc"
fi
if [[ "$control_command" == "recover-uncertain" && \
      "${FAKE_RECOVER_UNCERTAIN_FAIL:-0}" == "1" ]]; then
  printf 'RECOVERY_WRITE_FAILED\n' >&2
  exit 4
fi
if [[ "$control_command" == "${FAKE_TERMINAL_UNCERTAIN:-disabled}" && \
      ( "$control_command" == "pause" || "$control_command" == "release" ) ]]; then
  "$REAL_PYTHON_BIN" "$@"
  rc=$?
  if [[ "$rc" -eq 0 ]]; then
    printf 'CONTROL_COMMIT_UNCERTAIN\n' >&2
    exit 4
  fi
  exit "$rc"
fi
exec "$REAL_PYTHON_BIN" "$@"
SH
  chmod +x "$root/bin/python3"

  cat >"$root/bin/ps" <<'SH'
#!/usr/bin/env bash
set -u
if [[ "${1:-}" == "-o" && "${2:-}" == "pgid=" && "${3:-}" == "-p" && \
      -e "$PGID_MISMATCH_ARM_FILE" ]]; then
  printf '%s\n' "$(( ${4:?missing pid} + 100000 ))"
  exit 0
fi
if [[ "${1:-}" == "-axo" && "${FAKE_PS_LIST_FAIL:-0}" == "1" ]]; then
  exit 1
fi
exec "$REAL_PS_BIN" "$@"
SH
  chmod +x "$root/bin/ps"

  cat >"$root/bin/mkfifo" <<'SH'
#!/usr/bin/env bash
set -u
target=""
for argument in "$@"; do
  target="$argument"
done
if [[ "${FAKE_PRESEED_ROLE_GO:-0}" == "1" ]]; then
  printf 'preseeded\n' >"$target"
  printf 'preseeded\n' >"$PRESEED_PROBE_FILE"
  exit 1
fi
if [[ "${FAKE_FLOOD_ROLE_GO:-0}" == "1" ]]; then
  "$REAL_MKFIFO_BIN" "$@"
  "$REAL_PYTHON_BIN" - "$target" "$FLOOD_PROBE_FILE" <<'PY' &
import os
import pathlib
import sys
import time

fd = os.open(sys.argv[1], os.O_RDWR | os.O_NONBLOCK)
deadline = time.monotonic() + 10
written = 0
blocked = False
while time.monotonic() < deadline:
    try:
        written += os.write(fd, b"X" * 4096)
    except BlockingIOError:
        blocked = True
        break
pathlib.Path(sys.argv[2]).write_text(f"{written} {'blocked' if blocked else 'unblocked'}\n")
while time.monotonic() < deadline:
    time.sleep(0.01)
os.close(fd)
PY
  flooder=$!
  printf '%s %s\n' "$flooder" "$target" >>"$ROLE_PID_FILE"
  for ((attempt=0; attempt<500; attempt++)); do
    [[ -s "$FLOOD_PROBE_FILE" ]] && break
    /bin/sleep 0.01
  done
  exit 0
fi
if [[ "${FAKE_INTRUDE_ROLE_GO:-0}" == "1" ]]; then
  if ! mkdir "$INTRUDER_ONCE_DIR" 2>/dev/null; then
    exec "$REAL_MKFIFO_BIN" "$@"
  fi
  "$REAL_MKFIFO_BIN" "$@"
  "$REAL_PYTHON_BIN" - "$target" "$INTRUDER_PROBE_FILE" <<'PY' &
import os
import pathlib
import sys
import time

target = sys.argv[1]
marker = pathlib.Path(sys.argv[2])
fd = os.open(target, os.O_RDWR)
marker.write_text("opened\n")
os.write(fd, b"GO\n" + (b"0" * 64) + b"\n")
time.sleep(5)
os.close(fd)
PY
  intruder=$!
  printf '%s %s\n' "$intruder" "$target" >>"$ROLE_PID_FILE"
  for ((attempt=0; attempt<500; attempt++)); do
    [[ -s "$INTRUDER_PROBE_FILE" ]] && break
    /bin/sleep 0.01
  done
  exit 0
fi
exec "$REAL_MKFIFO_BIN" "$@"
SH
  chmod +x "$root/bin/mkfifo"

  cat >"$root/bin/git" <<'SH'
#!/usr/bin/env bash
set -u
if [[ " $* " == *" rev-parse "* && " $* " == *" HEAD "* ]]; then
  [[ "${FAKE_GIT_HEAD_FAIL:-0}" == "0" ]] || exit 1
  printf '%040d\n' 1
  exit 0
fi
exec "$REAL_GIT_BIN" "$@"
SH
  chmod +x "$root/bin/git"

  cat >"$root/bin/pi" <<'SH'
#!/usr/bin/env bash
set -u

if [[ -n "${AUTO_LOOP_STATE_FD:-}" ]]; then
  printf '%s %s\n' "$$" "$AUTO_LOOP_STATE_FD" >>"$STATE_FD_LEAK_FILE"
fi

if [[ " $* " == *" --offline --list-models "* ]]; then
  printf '%s %s\n' "$$" "$TEST_REPO/bin/pi" >>"$MODEL_PID_FILE"
  if [[ "${FAKE_MODEL_DELAY:-0}" != "0" ]]; then
    printf 'probing\n' >"$MODEL_PROBE_FILE"
    /bin/sleep "$FAKE_MODEL_DELAY"
    printf 'completed\n' >"$MODEL_COMPLETION_FILE"
  fi
  printf '%s\n' 'provider      model          context  max-out  thinking  images'
  if [[ "${FAKE_MODEL_MISSING:-0}" == "1" ]]; then
    printf '%s\n' 'openai-codex  gpt-5.5        272K     128K     yes       yes'
  else
    printf '%s\n' 'openai-codex  gpt-5.6-sol    372K     128K     yes       yes'
  fi
  exit 0
fi

model=""
thinking="default"
previous=""
for argument in "$@"; do
  if [[ "$previous" == "--model" ]]; then
    model="$argument"
  elif [[ "$previous" == "--thinking" ]]; then
    thinking="$argument"
  fi
  previous="$argument"
done

role="orchestrator"
if [[ "$model" == "openai-codex/gpt-5.6-sol" ]]; then
  role="validator"
fi
role_pgid="$(ps -o pgid= -p "$$" | tr -d '[:space:]')"
printf '%s %s %s %s %s\n' "$role" "$$" "${model:-default}" "$thinking" "$role_pgid" >>"$EVENT_LOG"

write_verdict() {
  local verdict="$1"
  local last_good
  last_good="$(find "$TEST_REPO/.planning/auto-loop/checkpoints" \
    -name LAST_GOOD -type f -exec cat {} \; 2>/dev/null | tail -n 1)"
  case "$verdict" in
    PROCEED)
      printf '{"verdict":"PROCEED","step_score":5,"trajectory_geomean":5,"reason":"synthetic trace evidence","correction":null,"revert_to_checkpoint":null}\n' ;;
    RETRY)
      printf '{"verdict":"RETRY","step_score":3,"trajectory_geomean":4,"reason":"synthetic trace evidence","correction":"repeat the synthetic stage","revert_to_checkpoint":null}\n' ;;
    REVERT)
      printf '{"verdict":"REVERT","step_score":1,"trajectory_geomean":3,"reason":"synthetic trace evidence","correction":"restore and repeat the synthetic stage","revert_to_checkpoint":"%s"}\n' "$last_good" ;;
    HALT)
      printf '{"verdict":"HALT","step_score":1,"trajectory_geomean":3,"reason":"synthetic hard-gate evidence","correction":null,"revert_to_checkpoint":null}\n' ;;
  esac >"$TEST_REPO/.planning/auto-loop/VALIDATOR-VERDICT.json"
}

write_malformed_verdict() {
  case "$FAKE_VERDICT_CASE" in
    missing-trajectory)
      printf '%s\n' '{"verdict":"PROCEED","step_score":5,"reason":"synthetic trace evidence","correction":null,"revert_to_checkpoint":null}' ;;
    low-proceed)
      printf '%s\n' '{"verdict":"PROCEED","step_score":3,"trajectory_geomean":5,"reason":"synthetic trace evidence","correction":null,"revert_to_checkpoint":null}' ;;
    empty-reason)
      printf '%s\n' '{"verdict":"PROCEED","step_score":5,"trajectory_geomean":5,"reason":"","correction":null,"revert_to_checkpoint":null}' ;;
    extra-key)
      printf '%s\n' '{"verdict":"PROCEED","step_score":5,"trajectory_geomean":5,"reason":"synthetic trace evidence","correction":null,"revert_to_checkpoint":null,"unexpected":true}' ;;
    proceed-correction)
      printf '%s\n' '{"verdict":"PROCEED","step_score":5,"trajectory_geomean":5,"reason":"synthetic trace evidence","correction":"must be null","revert_to_checkpoint":null}' ;;
    revert-target-mismatch)
      printf '%s\n' '{"verdict":"REVERT","step_score":1,"trajectory_geomean":3,"reason":"synthetic trace evidence","correction":"restore it","revert_to_checkpoint":"999999"}' ;;
    high-score-revert)
      printf '%s\n' '{"verdict":"REVERT","step_score":5,"trajectory_geomean":5,"reason":"synthetic trace evidence","correction":"restore it","revert_to_checkpoint":"1"}' ;;
  esac >"$TEST_REPO/.planning/auto-loop/VALIDATOR-VERDICT.json"
}

if [[ "$role" == "validator" ]]; then
  if [[ "$FAKE_MODE" == "validator-hang" ]]; then
    child_nonce="$TEST_NONCE-validator-$$"
    "$TEST_REPO/bin/test-child" "$child_nonce" 30 &
    child=$!
    printf '%s %s\n' "$child" "$child_nonce" >>"$CHILD_PID_FILE"
    wait "$child"
  elif [[ "$FAKE_MODE" == "halt" || "$FAKE_MODE" == "halt-latch-failure" ]]; then
    write_verdict HALT
  elif [[ "$FAKE_MODE" == "revert" ]]; then
    write_verdict REVERT
  elif [[ "$FAKE_MODE" == "no-verdict" ]]; then
    : # Deliberately emit nothing; any pre-existing verdict must be retired by the controller.
  elif [[ "$FAKE_MODE" == "retry-human-gate" ]]; then
    write_verdict RETRY
  elif [[ "$FAKE_MODE" == "no-verdict-human-gate" || "$FAKE_MODE" == "no-verdict-budget" ]]; then
    :
  elif [[ "$FAKE_MODE" == "malformed-human-gate" ]]; then
    write_malformed_verdict
  elif [[ "$FAKE_MODE" == "validator-nonzero-proceed" ]]; then
    write_verdict PROCEED
    exit 1
  elif [[ "$FAKE_MODE" == "trace-overwrite" ]]; then
    write_verdict RETRY
  else
    write_verdict PROCEED
  fi
  exit 0
fi

case "$FAKE_MODE" in
  retry-human-gate|no-verdict-human-gate|proceed-human-gate|malformed-human-gate|validator-nonzero-proceed|trace-overwrite)
    printf '{"stage":"FINALIZE","terminal":"human_gate"}\n' >"$TEST_REPO/.planning/auto-loop/RUN.json"
    ;;
  no-verdict-budget)
    printf '{"stage":"EXECUTE","terminal":"budget"}\n' >"$TEST_REPO/.planning/auto-loop/RUN.json"
    ;;
  checkpoint-invalid-run)
    printf '%s\n' 'not-json' >"$TEST_REPO/.planning/auto-loop/RUN.json"
    ;;
  checkpoint-alias)
    mkdir -p "$TEST_REPO/outside-checkpoints"
    printf 'outside-sentinel\n' >"$TEST_REPO/outside-checkpoints/sentinel"
    ln -s "$TEST_REPO/outside-checkpoints" "$TEST_REPO/.planning/auto-loop/checkpoints"
    ;;
  revert)
    printf '{"stage":"DIVERGED","terminal":null}\n' >"$TEST_REPO/.planning/auto-loop/RUN.json"
    ;;
esac

case "$FAKE_MODE" in
  validator-hang)
    /bin/sleep 8
    ;;
  concurrent)
    /bin/sleep 1
    ;;
  hang-tree)
    child_nonce="$TEST_NONCE-role-$$"
    "$TEST_REPO/bin/test-child" "$child_nonce" 8 &
    child=$!
    printf '%s %s\n' "$child" "$child_nonce" >>"$CHILD_PID_FILE"
    trap '' TERM
    wait "$child"
    ;;
  orphan)
    child_nonce="$TEST_NONCE-role-$$"
    "$TEST_REPO/bin/test-child" "$child_nonce" 6 &
    child=$!
    printf '%s %s\n' "$child" "$child_nonce" >>"$CHILD_PID_FILE"
    exit 0
    ;;
  signal)
    child_nonce="$TEST_NONCE-role-$$"
    "$TEST_REPO/bin/test-child" "$child_nonce" 8 &
    child=$!
    printf '%s %s\n' "$child" "$child_nonce" >>"$CHILD_PID_FILE"
    wait "$child"
    ;;
  reserved-exit)
    exit 124
    ;;
  *)
    exit 0
    ;;
esac
SH
  chmod +x "$root/bin/pi"
}

driver_env() {
  local root="$1"
  local -a clean_env
  shift
  clean_env=(
    "PATH=$root/bin:$PATH"
    "HOME=$root/home"
    "PI_BIN=$root/bin/pi"
    "VALIDATOR_BIN=$root/bin/pi"
    "VALIDATOR_ARGS="
    "PI_EXTRA_FLAGS="
    "EVENT_LOG=$root/events"
    "CHILD_PID_FILE=$root/child-pids"
    "MODEL_PID_FILE=$root/model-pids"
    "CONTROLLER_PID_FILE=$root/controller-pids"
    "ROLE_PID_FILE=$root/role-pids"
    "STATE_FD_LEAK_FILE=$root/state-fd-leaks"
    "NATURAL_EXIT_LOG=$root/natural-exits"
    "TEST_NONCE=$(cat "$root/nonce")"
    "TEST_REPO=$root"
    "FAKE_MODE=${FAKE_MODE:-normal}"
    "FAKE_VERDICT_CASE=${FAKE_VERDICT_CASE:-}"
    "FAKE_MODEL_DELAY=${FAKE_MODEL_DELAY:-0}"
    "FAKE_MODEL_MISSING=${FAKE_MODEL_MISSING:-0}"
    "FAKE_LAUNCH_DELAY=${FAKE_LAUNCH_DELAY:-0}"
    "FAKE_EARLY_READY_DELAY=${FAKE_EARLY_READY_DELAY:-0}"
    "FAKE_BEGIN_RETURN_DELAY=${FAKE_BEGIN_RETURN_DELAY:-0}"
    "FAKE_CONTROL_ORPHAN=${FAKE_CONTROL_ORPHAN:-0}"
    "FAKE_BIND_DELAY=${FAKE_BIND_DELAY:-0}"
    "FAKE_ASSERT_DELAY=${FAKE_ASSERT_DELAY:-0}"
    "FAKE_ASSERT_BARRIER=${FAKE_ASSERT_BARRIER:-0}"
    "FAKE_PRESEED_ROLE_GO=${FAKE_PRESEED_ROLE_GO:-0}"
    "FAKE_FLOOD_ROLE_GO=${FAKE_FLOOD_ROLE_GO:-0}"
    "FAKE_INTRUDE_ROLE_GO=${FAKE_INTRUDE_ROLE_GO:-0}"
    "FAKE_TERMINAL_UNCERTAIN=${FAKE_TERMINAL_UNCERTAIN:-}"
    "FAKE_RECOVER_UNCERTAIN_FAIL=${FAKE_RECOVER_UNCERTAIN_FAIL:-0}"
    "FAKE_TRACE_DELAY=${FAKE_TRACE_DELAY:-0}"
    "FAKE_TRACE_EXIT=${FAKE_TRACE_EXIT:-0}"
    "FAKE_TRACE_OVERWRITE=${FAKE_TRACE_OVERWRITE:-0}"
    "FAKE_GIT_HEAD_FAIL=${FAKE_GIT_HEAD_FAIL:-0}"
    "FAKE_PS_LIST_FAIL=${FAKE_PS_LIST_FAIL:-0}"
    "DESCENDANT_LOCK_READY_FILE=${DESCENDANT_LOCK_READY_FILE:-}"
    "MODEL_PROBE_FILE=$root/model-probe"
    "MODEL_COMPLETION_FILE=$root/model-completion"
    "LAUNCH_PROBE_FILE=$root/launch-probe"
    "EARLY_READY_PROBE_FILE=$root/early-ready-probe"
    "BEGIN_COMMIT_PROBE_FILE=$root/begin-commit-probe"
    "BIND_PROBE_FILE=$root/bind-probe"
    "ASSERT_PROBE_FILE=$root/assert-probe"
    "ASSERT_COMPLETION_FILE=$root/assert-completion"
    "ASSERT_RELEASE_FILE=$root/assert-release"
    "PRESEED_PROBE_FILE=$root/preseed-probe"
    "FLOOD_PROBE_FILE=$root/flood-probe"
    "INTRUDER_PROBE_FILE=$root/intruder-probe"
    "INTRUDER_ONCE_DIR=$root/intruder-once"
    "TRACE_PROBE_FILE=$root/trace-probe"
    "PGID_MISMATCH_ARM_FILE=$root/pgid-mismatch-arm"
    "REAL_PYTHON_BIN=$REAL_PYTHON_BIN"
    "REAL_PS_BIN=$REAL_PS_BIN"
    "REAL_MKFIFO_BIN=$REAL_MKFIFO_BIN"
    "REAL_GIT_BIN=$REAL_GIT_BIN"
    "MAX_ITERATIONS=${MAX_ITERATIONS:-1}"
    "MAX_TURNS=${MAX_TURNS:-20}"
    "TURN_TIMEOUT_SECONDS=${TURN_TIMEOUT_SECONDS:-30}"
    "TERM_GRACE_SECONDS=${TERM_GRACE_SECONDS:-1}"
    "CONTROL_HEARTBEAT_SECONDS=${CONTROL_HEARTBEAT_SECONDS:-1}"
    "COOLDOWN_SECONDS=0"
    "STALL_MINUTES=1"
  )
  if [[ "${DRIVER_EXEC:-0}" == "1" ]]; then
    exec /usr/bin/env -i "${clean_env[@]}" "$root/scripts/pi-shepherd-loop.sh" "$@"
  fi
  /usr/bin/env -i "${clean_env[@]}" "$root/scripts/pi-shepherd-loop.sh" "$@"
}

register_controller() {
  local root="$1" pid="$2"
  printf '%s %s\n' "$pid" "$root/scripts/pi-shepherd-loop.sh" >>"$root/controller-pids"
}

start_bystander() {
  local root="$1" nonce pid
  nonce="$(cat "$root/nonce")-bystander-$RANDOM"
  "$root/bin/test-bystander" "$nonce" 30 "$root/bystander-signals" >/dev/null 2>&1 &
  pid=$!
  printf '%s %s\n' "$pid" "$nonce" >>"$root/child-pids"
  BYSTANDER_IDENTITY="$pid $nonce $root/bystander-signals"
}

assert_bystander_alive() {
  local identity="$1" context="$2" pid nonce signal_log
  read -r pid nonce signal_log <<<"$identity"
  if ! owned_process_alive "$pid" "$nonce"; then
    fail "$context killed unrelated bystander $pid"
  fi
  if grep -Fxq "$nonce" "$signal_log" 2>/dev/null; then
    fail "$context signalled unrelated bystander $pid"
  fi
}

wait_owned_gone() {
  local pid="$1" nonce="$2" attempts="${3:-200}" i
  for ((i=0; i<attempts; i++)); do
    owned_process_alive "$pid" "$nonce" || return 0
    /bin/sleep 0.01
  done
  return 1
}

wait_driver_bounded() {
  local root="$1" pid="$2" attempts="${3:-2000}"
  if wait_owned_gone "$pid" "$root/scripts/pi-shepherd-loop.sh" "$attempts"; then
    return 0
  fi
  signal_owned_process KILL "$pid" "$root/scripts/pi-shepherd-loop.sh"
  return 1
}

monotonic_now() {
  python3 - <<'PY'
import time
print(time.monotonic())
PY
}

duration_less_than() {
  python3 - "$1" "$2" "$3" <<'PY'
import sys
raise SystemExit(0 if float(sys.argv[2]) - float(sys.argv[1]) < float(sys.argv[3]) else 1)
PY
}

event_count() {
  local root="$1" role="$2"
  awk -v role="$role" '$1 == role { count++ } END { print count + 0 }' "$root/events"
}

control_field() {
  local root="$1" expression="$2"
  python3 - "$root/.planning/auto-loop/CONTROL.json" "$expression" <<'PY' 2>/dev/null
import json, pathlib, sys
path = pathlib.Path(sys.argv[1])
if not path.exists():
    print("")
    raise SystemExit(0)
value = json.loads(path.read_text())
for part in sys.argv[2].split("."):
    if not part:
        continue
    if not isinstance(value, dict) or part not in value:
        print("")
        raise SystemExit(0)
    value = value[part]
if isinstance(value, bool):
    print(str(value).lower())
elif value is None:
    print("")
else:
    print(value)
PY
}

control_snapshot() {
  local root="$1"
  python3 - "$root/.planning/auto-loop/CONTROL.json" <<'PY' 2>/dev/null
import json, pathlib, sys
path = pathlib.Path(sys.argv[1])
if not path.exists():
    print("missing\t\t\t")
    raise SystemExit(0)
value = json.loads(path.read_text())
print("\t".join((
    str(value.get("phase", "")),
    str(value.get("turn_ordinal", "")),
    "null" if value.get("active_turn") is None else "present",
    str(value.get("children_quiescent", "")).lower(),
)))
PY
}

write_valid_released_control() {
  python3 - "$1" <<'PY'
import datetime as dt
import json
import pathlib
import sys
import uuid

now = dt.datetime.now(dt.timezone.utc).replace(microsecond=0)
stamp = lambda value: value.isoformat(timespec="seconds").replace("+00:00", "Z")
value = {
    "schema_version": "1.0",
    "run_id": str(uuid.uuid4()),
    "generation": 1,
    "controller_id": str(uuid.uuid4()),
    "control_revision": 1,
    "phase": "released",
    "lease": {"heartbeat_at": stamp(now), "expires_at": stamp(now + dt.timedelta(seconds=3))},
    "limits": {
        "max_turns": 20,
        "turn_timeout_seconds": 10,
        "term_grace_seconds": 1,
        "heartbeat_seconds": 1,
        "max_reverts": 6,
        "max_no_verdict": 3,
        "max_minutes": 0,
    },
    "turn_ordinal": 0,
    "counters": {"reverts": 0, "no_verdict": 0, "active_seconds": 0},
    "active_turn": None,
    "halt": None,
    "children_quiescent": True,
    "updated_at": stamp(now),
}
pathlib.Path(sys.argv[1]).write_text(
    json.dumps(value, sort_keys=True, separators=(",", ":")) + "\n"
)
PY
}

wait_for_control_value() {
  local root="$1" expression="$2" expected="$3" attempts="${4:-1000}" i
  for ((i=0; i<attempts; i++)); do
    [[ "$(control_field "$root" "$expression")" == "$expected" ]] && return 0
    /bin/sleep 0.01
  done
  return 1
}

wait_for_file() {
  local path="$1" attempts="${2:-1000}"
  local i
  for ((i=0; i<attempts; i++)); do
    [[ -s "$path" ]] && return 0
    /bin/sleep 0.01
  done
  return 1
}

checkpoint_marker_exists() {
  local root="$1"
  [[ -n "$(find "$root/.planning/auto-loop/checkpoints" -name LAST_GOOD -type f -print -quit 2>/dev/null)" ]]
}

wait_for_role_child() {
  local path="$1" attempts="${2:-1000}" i
  for ((i=0; i<attempts; i++)); do
    if [[ -s "$path" ]] && grep -qv 'bystander' "$path"; then
      return 0
    fi
    /bin/sleep 0.01
  done
  return 1
}

test_concurrent_controllers_have_one_winner() {
  local root pid rc winners=0 held=0 other=0 i
  local -a pids
  root="$(mktemp -d "$TEST_TMP/concurrent.XXXXXX")"
  prepare_fixture "$root"

  for ((i=0; i<32; i++)); do
    (
      DRIVER_EXEC=1 FAKE_MODE=concurrent TURN_TIMEOUT_SECONDS=10 driver_env "$root" "synthetic concurrent task" \
        >"$root/stdout.$i" 2>"$root/stderr.$i"
    ) &
    pids[$i]=$!
    register_controller "$root" "${pids[$i]}"
  done

  for ((i=0; i<32; i++)); do
    pid="${pids[$i]}"
    wait "$pid"
    rc=$?
    case "$rc" in
      3) winners=$((winners + 1)) ;;
      75)
        held=$((held + 1))
        if ! grep -q 'CONTROLLER_HELD' "$root/stderr.$i"; then
          fail "controller contender $i exited 75 without CONTROLLER_HELD"
        fi
        ;;
      *) other=$((other + 1)) ;;
    esac
  done

  if [[ "$(event_count "$root" orchestrator)" -ne 1 || "$winners" -ne 1 || "$held" -ne 31 || "$other" -ne 0 ]]; then
    fail "32 concurrent controllers: orchestrators=$(event_count "$root" orchestrator) winners=$winners held=$held other=$other; want 1/1/31/0"
  fi
}

test_state_directory_replacement_cannot_split_controller_lock() {
  local root level driver child child_nonce before rc first_rc
  for level in auto-loop planning; do
    rc=0
    first_rc=0
    root="$(mktemp -d "$TEST_TMP/state-root-replacement-$level.XXXXXX")"
    prepare_fixture "$root"

    (
      DRIVER_EXEC=1 FAKE_MODE=signal TURN_TIMEOUT_SECONDS=20 \
        driver_env "$root" "synthetic $level replacement" \
        >"$root/stdout.$level.1" 2>"$root/stderr.$level.1"
    ) &
    driver=$!
    register_controller "$root" "$driver"
    if ! wait_for_control_value "$root" active_turn.active_role orchestrator || \
       ! wait_for_role_child "$root/child-pids"; then
      fail "$level replacement fixture did not establish an active controller tree"
      kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
      wait "$driver" 2>/dev/null || true
      continue
    fi
    read -r child child_nonce < <(awk 'END { print $1, $2 }' "$root/child-pids")
    before="$(event_count "$root" orchestrator)"
    if [[ "$level" == "auto-loop" ]]; then
      mv "$root/.planning/auto-loop" "$root/.planning/auto-loop.original"
    else
      mv "$root/.planning" "$root/.planning.original"
      mkdir "$root/.planning"
    fi
    mkdir "$root/.planning/auto-loop"

    FAKE_MODE=normal driver_env "$root" "synthetic replacement contender" \
      >"$root/stdout.$level.2" 2>"$root/stderr.$level.2" || rc=$?
    if [[ "$rc" -ne 75 ]] || ! grep -q 'CONTROLLER_HELD' "$root/stderr.$level.2" || \
       [[ "$(event_count "$root" orchestrator)" -ne "$before" ]] || \
       [[ -e "$root/.planning/auto-loop/CONTROL.json" ]]; then
      fail "replaced $level directory admitted a second controller lock namespace: rc=$rc"
    fi

    signal_owned_process TERM "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || first_rc=$?
    if [[ "$first_rc" -ne 4 ]] || ! wait_owned_gone "$child" "$child_nonce" 200; then
      fail "original controller did not drain after $level replacement test"
    fi
    rc=0
    FAKE_MODE=normal driver_env "$root" "synthetic post-replacement contender" \
      >"$root/stdout.$level.3" 2>"$root/stderr.$level.3" || rc=$?
    if [[ "$rc" -ne 4 || "$(event_count "$root" orchestrator)" -ne "$before" ]] || \
       [[ ! -f "$root/.auto-loop-recovery" ]] || \
       ! grep -q 'RECOVERY_REQUIRED' "$root/stderr.$level.3"; then
      fail "$level replacement admitted work after the original lock released: rc=$rc"
    fi
  done
}

test_hard_deadline_drains_process_tree() {
  local root driver child child_nonce rc=0 bystander start end elapsed_fast
  root="$(mktemp -d "$TEST_TMP/deadline.XXXXXX")"
  prepare_fixture "$root"
  start_bystander "$root"
  bystander="$BYSTANDER_IDENTITY"

  (
    DRIVER_EXEC=1 FAKE_MODE=hang-tree TURN_TIMEOUT_SECONDS=3 \
      driver_env "$root" "synthetic deadline task" >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_role_child "$root/child-pids"; then
    fail "hard deadline test did not start its supervised child"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  read -r child child_nonce < <(grep -v 'bystander' "$root/child-pids" | tail -n 1)
  start="$(monotonic_now)"
  wait "$driver" || rc=$?
  end="$(monotonic_now)"
  elapsed_fast="$(python3 - "$start" "$end" <<'PY'
import sys
print("yes" if float(sys.argv[2]) - float(sys.argv[1]) < 6.0 else "no")
PY
)"

  if [[ "$rc" -ne 4 ]]; then
    fail "hard deadline exit=$rc, want 4"
  fi
  if [[ "$(event_count "$root" validator)" -ne 0 ]]; then
    fail "validator started after hard deadline"
  fi
  if [[ "$elapsed_fast" != "yes" ]]; then
    fail "hard deadline waited for natural child exit"
  fi
  if grep -Fxq "$child_nonce" "$root/natural-exits" 2>/dev/null; then
    fail "hard deadline allowed its child to exit naturally"
  fi
  if owned_process_alive "$child" "$child_nonce"; then
    fail "hard deadline left descendant $child alive"
  fi
  if [[ "$(control_field "$root" phase)" != "halted" ]] || \
     [[ "$(control_field "$root" children_quiescent)" != "true" ]]; then
    fail "hard deadline did not persist halted/quiescent control state"
  fi
  if [[ "$(control_field "$root" halt.code)" != "TURN_DEADLINE" ]]; then
    fail "hard deadline did not persist TURN_DEADLINE"
  fi
  assert_bystander_alive "$bystander" "hard deadline teardown"
}

test_leader_exit_with_live_descendant_halts() {
  local root child child_nonce rc bystander
  root="$(mktemp -d "$TEST_TMP/orphan.XXXXXX")"
  prepare_fixture "$root"
  start_bystander "$root"
  bystander="$BYSTANDER_IDENTITY"

  FAKE_MODE=orphan driver_env "$root" "synthetic orphan task" >"$root/stdout" 2>"$root/stderr"
  rc=$?
  read -r child child_nonce < <(grep -v 'bystander' "$root/child-pids" | tail -n 1)

  if [[ "$rc" -ne 4 || "$(event_count "$root" validator)" -ne 0 ]]; then
    fail "leader-exit orphan was accepted: exit=$rc validators=$(event_count "$root" validator)"
  fi
  if owned_process_alive "$child" "$child_nonce"; then
    fail "leader-exit orphan left descendant $child alive"
  fi
  if [[ "$(control_field "$root" halt.code)" != "ORCHESTRATOR_ORPHAN" ]]; then
    fail "leader-exit orphan did not persist ORCHESTRATOR_ORPHAN"
  fi
  assert_bystander_alive "$bystander" "orphan teardown"
}

test_validator_shares_the_hard_turn_deadline() {
  local root driver child child_nonce rc=0 start end deadline_at deadline_epoch now_epoch remaining allowed
  root="$(mktemp -d "$TEST_TMP/validator-deadline.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_MODE=validator-hang TURN_TIMEOUT_SECONDS=15 \
      driver_env "$root" "synthetic validator deadline" >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_role_child "$root/child-pids"; then
    fail "validator deadline test did not start its supervised child"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  read -r child child_nonce < <(grep -v 'bystander' "$root/child-pids" | tail -n 1)
  deadline_at="$(control_field "$root" active_turn.deadline_at)"
  deadline_epoch="$(python3 - "$deadline_at" <<'PY'
import datetime as dt
import sys
try:
    print(int(dt.datetime.fromisoformat(sys.argv[1].replace("Z", "+00:00")).timestamp()))
except (IndexError, ValueError):
    print(0)
PY
)"
  now_epoch="$(date -u +%s)"
  remaining=$((deadline_epoch - now_epoch))
  allowed=$((remaining + 5))
  start="$(monotonic_now)"
  wait "$driver" || rc=$?
  end="$(monotonic_now)"

  if [[ "$rc" -ne 4 ]] || \
     [[ "$(event_count "$root" orchestrator)" -ne 1 || "$(event_count "$root" validator)" -ne 1 ]] || \
     [[ "$(control_field "$root" phase)" != "halted" ]] || \
     [[ "$(control_field "$root" halt.code)" != "TURN_DEADLINE" ]] || \
     (( remaining <= 0 || allowed >= 15 )) || \
     ! duration_less_than "$start" "$end" "$allowed" || \
     checkpoint_marker_exists "$root"; then
    fail "validator did not share the persisted hard turn deadline: rc=$rc remaining=$remaining allowed=$allowed start=$start end=$end control=$(control_snapshot "$root")"
  fi
  if owned_process_alive "$child" "$child_nonce" || \
     grep -Fxq "$child_nonce" "$root/natural-exits" 2>/dev/null; then
    fail "validator deadline left or naturally completed its descendant"
  fi
}

test_missing_validator_model_fails_before_work() {
  local root rc=0 model_pid
  root="$(mktemp -d "$TEST_TMP/model-missing.XXXXXX")"
  prepare_fixture "$root"

  FAKE_MODEL_MISSING=1 driver_env "$root" "synthetic missing validator model" \
    >"$root/stdout" 2>"$root/stderr" || rc=$?
  model_pid="$(awk 'END { print $1 }' "$root/model-pids")"
  if [[ "$rc" -ne 2 || -s "$root/events" ]] || \
     [[ -e "$root/.planning/auto-loop/PROMPT.txt" ]] || \
     [[ "$(control_field "$root" phase)" != "recovery_required" ]] || \
     [[ "$(control_field "$root" halt.code)" != "VALIDATOR_MODEL_UNAVAILABLE" ]] || \
     ! grep -q 'FATAL: Shepherd requires openai-codex/gpt-5.6-sol' "$root/stderr"; then
    fail "missing exact Shepherd model did not fail closed before work"
  fi
  if owned_process_alive "$model_pid" "$root/bin/pi"; then
    fail "missing-model preflight left model discovery alive"
  fi
}

test_slow_model_preflight_renews_epoch_lease() {
  local root rc=0
  root="$(mktemp -d "$TEST_TMP/model-heartbeat.XXXXXX")"
  prepare_fixture "$root"

  FAKE_MODEL_DELAY=3 CONTROL_HEARTBEAT_SECONDS=1 TURN_TIMEOUT_SECONDS=20 \
    driver_env "$root" "synthetic slow model preflight" >"$root/stdout" 2>"$root/stderr" || rc=$?
  if [[ "$rc" -ne 3 || ! -s "$root/model-completion" ]] || \
     [[ "$(event_count "$root" orchestrator)" -ne 1 || "$(event_count "$root" validator)" -ne 1 ]] || \
     [[ "$(control_field "$root" phase)" != "paused" ]]; then
    fail "slow validator-model discovery lost epoch authority: rc=$rc control=$(control_snapshot "$root")"
  fi
}

test_halt_is_durable_and_blocks_resume() {
  local root rc before after snapshot diagnostic
  root="$(mktemp -d "$TEST_TMP/halt.XXXXXX")"
  prepare_fixture "$root"

  FAKE_MODE=halt driver_env "$root" "synthetic halt task" >"$root/stdout.1" 2>"$root/stderr.1"
  rc=$?
  if [[ "$rc" -ne 4 || "$(control_field "$root" phase)" != "halted" ]]; then
    snapshot="$(control_snapshot "$root")"
    diagnostic="$(tail -n 6 "$root/stderr.1" 2>/dev/null | tr '\n' '|')"
    fail "validator HALT observed rc=$rc control=$snapshot stderr=$diagnostic"
    return
  fi
  before="$(shasum -a 256 "$root/.planning/auto-loop/CONTROL.json" | awk '{print $1}')"
  : >"$root/events"
  mv "$root/.planning/auto-loop/PROMPT.txt" "$root/PROMPT.blocked"

  FAKE_MODE=normal driver_env "$root" --resume >"$root/stdout.2" 2>"$root/stderr.2"
  rc=$?
  mv "$root/PROMPT.blocked" "$root/.planning/auto-loop/PROMPT.txt"
  after="$(shasum -a 256 "$root/.planning/auto-loop/CONTROL.json" | awk '{print $1}')"

  if [[ "$rc" -ne 4 || -s "$root/events" || "$before" != "$after" ]] || \
     ! grep -q 'HALT_LATCHED' "$root/stderr.2"; then
    fail "halted resume was not rejected before prompt/provider access"
  fi

  FAKE_MODE=normal driver_env "$root" "synthetic fresh start after halt" >"$root/stdout.3" 2>"$root/stderr.3"
  rc=$?
  after="$(shasum -a 256 "$root/.planning/auto-loop/CONTROL.json" | awk '{print $1}')"
  if [[ "$rc" -ne 4 || -s "$root/events" || "$before" != "$after" ]] || \
     ! grep -q 'HALT_LATCHED' "$root/stderr.3"; then
    fail "halted fresh start was not rejected before provider access"
  fi
}

test_signal_during_startup_is_durable() {
  local root driver model_pid rc=0 start end elapsed_fast
  root="$(mktemp -d "$TEST_TMP/startup-signal.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_MODEL_DELAY=5 driver_env "$root" "synthetic startup signal" \
      >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_file "$root/model-probe"; then
    fail "startup signal test did not enter validator-model preflight"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  model_pid="$(awk 'END { print $1 }' "$root/model-pids")"
  start="$(monotonic_now)"
  signal_owned_process TERM "$driver" "$root/scripts/pi-shepherd-loop.sh"
  wait "$driver" || rc=$?
  end="$(monotonic_now)"
  elapsed_fast="$(python3 - "$start" "$end" <<'PY'
import sys
print("yes" if float(sys.argv[2]) - float(sys.argv[1]) < 3.0 else "no")
PY
)"
  if [[ "$rc" -ne 4 || "$(control_field "$root" phase)" != "recovery_required" ]] || \
     [[ "$(control_field "$root" children_quiescent)" != "true" ]] || [[ -s "$root/events" ]] || \
     [[ "$elapsed_fast" != "yes" || -e "$root/model-completion" ]] || \
     [[ "$(control_field "$root" halt.code)" != "CONTROLLER_SIGNAL" ]]; then
    fail "signal during model preflight did not durably enter quiescent recovery"
  fi
  if owned_process_alive "$model_pid" "$root/bin/pi"; then
    fail "signal during model preflight left the model-discovery process alive"
  fi
}

test_signal_during_role_startup_does_not_orphan() {
  local root driver rc=0
  root="$(mktemp -d "$TEST_TMP/role-start-signal.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_LAUNCH_DELAY=5 driver_env "$root" "synthetic role-start signal" \
      >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_file "$root/launch-probe"; then
    fail "role-start signal test did not enter launch-before-bind window"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  signal_owned_process TERM "$driver" "$root/scripts/pi-shepherd-loop.sh"
  wait "$driver" || rc=$?
  if [[ "$rc" -ne 4 || "$(control_field "$root" phase)" != "recovery_required" ]] || \
     [[ "$(control_field "$root" children_quiescent)" != "true" ]] || [[ -s "$root/events" ]]; then
    fail "signal during role startup left unproven authority or launched Pi"
  fi
}

test_role_stays_inert_until_durable_bind() {
  local root driver rc=0
  root="$(mktemp -d "$TEST_TMP/role-bind-signal.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_BIND_DELAY=3 TURN_TIMEOUT_SECONDS=10 \
      driver_env "$root" "synthetic role-bind signal" >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_file "$root/bind-probe"; then
    fail "role-bind test did not enter the pre-bind persistence window"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  if [[ -s "$root/events" ]]; then
    fail "role command executed before its control bind was durable"
  fi
  signal_owned_process TERM "$driver" "$root/scripts/pi-shepherd-loop.sh"
  wait "$driver" || rc=$?
  if [[ "$rc" -ne 4 || "$(control_field "$root" phase)" != "recovery_required" ]] || \
     [[ "$(control_field "$root" children_quiescent)" != "true" || -s "$root/events" ]]; then
    fail "signal during durable role bind launched work or left unproven quiescence"
  fi
}

test_role_is_not_authorized_after_deadline() {
  local root driver rc=0
  root="$(mktemp -d "$TEST_TMP/role-bind-deadline.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_BIND_DELAY=12 TURN_TIMEOUT_SECONDS=6 \
      driver_env "$root" "synthetic expired role bind" >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_file "$root/bind-probe"; then
    fail "expired-bind fixture did not reach the inert durable-bind delay"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  wait "$driver" || rc=$?
  if [[ "$rc" -ne 4 || -s "$root/events" ]] || \
     [[ "$(control_field "$root" phase)" != "halted" ]] || \
     [[ "$(control_field "$root" halt.code)" != "TURN_DEADLINE" ]] || \
     [[ "$(control_field "$root" children_quiescent)" != "true" ]]; then
    fail "role was authorized after its persisted turn deadline"
  fi
}

test_preseeded_role_go_cannot_launch_provider() {
  local root rc=0
  root="$(mktemp -d "$TEST_TMP/preseed-role-go.XXXXXX")"
  prepare_fixture "$root"

  FAKE_PRESEED_ROLE_GO=1 driver_env "$root" "synthetic preseeded role go" \
    >"$root/stdout" 2>"$root/stderr" || rc=$?
  if [[ ! -s "$root/preseed-probe" ]]; then
    fail "role-go preseed fixture did not arm"
  fi
  if [[ "$rc" -ne 4 || -s "$root/events" ]] || \
     [[ "$(control_field "$root" phase)" != "recovery_required" ]]; then
    fail "preseeded role-go marker launched a provider or escaped recovery: rc=$rc control=$(control_snapshot "$root")"
  fi
}

test_controller_sigkill_before_role_go_never_launches_provider() {
  local root driver leader assert_pid
  root="$(mktemp -d "$TEST_TMP/pre-go-sigkill.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_ASSERT_DELAY=3 TURN_TIMEOUT_SECONDS=15 \
      driver_env "$root" "synthetic pre-go sigkill" >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_file "$root/assert-probe"; then
    fail "pre-go SIGKILL fixture did not reach the post-bind authorization window"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  leader="$(control_field "$root" active_turn.leader_pid)"
  # The former SIGSTOP/SIGCONT handshake could be released by any same-UID process. A stray or
  # hostile CONT must now be irrelevant because only the inherited FIFO capability authorizes.
  signal_owned_process CONT "$leader" "$REAL_PYTHON_BIN"
  /bin/sleep 0.1
  if [[ -s "$root/events" ]]; then
    fail "external SIGCONT authorized the pre-go role"
  fi
  signal_owned_process KILL "$driver" "$root/scripts/pi-shepherd-loop.sh"
  wait "$driver" 2>/dev/null || true
  assert_pid="$(awk '$2 ~ /\/bin\/python3$/ { pid=$1 } END { print pid }' "$root/controller-pids")"
  if ! wait_owned_gone "$assert_pid" "$root/bin/python3" 500; then
    fail "post-bind assertion helper did not terminate after controller SIGKILL"
  fi
  if ! wait_owned_gone "$leader" "$REAL_PYTHON_BIN" 500; then
    fail "private-GO role did not terminate after controller SIGKILL"
  fi
  if [[ -s "$root/events" ]] || owned_process_alive "$leader" "$REAL_PYTHON_BIN"; then
    fail "controller SIGKILL authorized or orphaned the pre-go role: control=$(control_snapshot "$root")"
  fi
}

test_signal_drains_group_and_requires_recovery() {
  local root driver child child_nonce phase bystander rc=0
  root="$(mktemp -d "$TEST_TMP/signal.XXXXXX")"
  prepare_fixture "$root"
  start_bystander "$root"
  bystander="$BYSTANDER_IDENTITY"

  (
    DRIVER_EXEC=1 FAKE_MODE=signal TURN_TIMEOUT_SECONDS=10 driver_env "$root" "synthetic signal task" \
      >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_role_child "$root/child-pids"; then
    fail "signal test did not start descendant"
    signal_owned_process KILL "$driver" "$root/scripts/pi-shepherd-loop.sh"
    return
  fi
  read -r child child_nonce < <(grep -v 'bystander' "$root/child-pids" | tail -n 1)
  signal_owned_process TERM "$driver" "$root/scripts/pi-shepherd-loop.sh"
  wait "$driver" 2>/dev/null || rc=$?
  /bin/sleep 0.1

  if owned_process_alive "$child" "$child_nonce"; then
    fail "signal left descendant $child alive"
  fi
  phase="$(control_field "$root" phase)"
  if [[ "$rc" -ne 4 || "$phase" != "recovery_required" ]] || \
     [[ "$(control_field "$root" halt.code)" != "CONTROLLER_SIGNAL" ]] || \
     [[ "$(control_field "$root" children_quiescent)" != "true" ]] || \
     [[ -n "$(control_field "$root" active_turn.leader_pid)" ]] || \
     [[ -n "$(control_field "$root" active_turn.process_group_id)" ]] || \
     [[ "$(event_count "$root" validator)" -ne 0 ]]; then
    fail "signal persisted phase=$phase, want recovery_required"
  fi
  assert_bystander_alive "$bystander" "signal teardown"
}

test_pgid_mismatch_never_signals_untrusted_pid() {
  local root driver child child_nonce leader rc=0
  root="$(mktemp -d "$TEST_TMP/pgid-mismatch.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_MODE=signal TURN_TIMEOUT_SECONDS=15 \
      driver_env "$root" "synthetic pgid mismatch" >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_control_value "$root" active_turn.active_role orchestrator || \
     ! wait_for_role_child "$root/child-pids"; then
    fail "PGID mismatch fixture did not reach a bound role tree"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  read -r child child_nonce < <(grep -v 'bystander' "$root/child-pids" | tail -n 1)
  leader="$(control_field "$root" active_turn.leader_pid)"
  printf '%s %s\n' "$leader" "$root/bin/pi" >>"$root/controller-pids"
  : >"$root/pgid-mismatch-arm"
  signal_owned_process TERM "$driver" "$root/scripts/pi-shepherd-loop.sh"
  wait "$driver" 2>/dev/null || rc=$?

  if [[ "$rc" -ne 4 || "$(control_field "$root" phase)" != "recovery_required" ]] || \
     [[ "$(control_field "$root" children_quiescent)" == "true" ]] || \
     [[ -z "$(control_field "$root" active_turn.process_group_id)" ]]; then
    fail "PGID mismatch did not preserve unresolved recovery evidence"
  fi
  if ! owned_process_alive "$leader" "$root/bin/pi"; then
    fail "PGID mismatch signalled the now-untrusted leader PID"
  fi
  if ! owned_process_alive "$child" "$child_nonce"; then
    fail "PGID mismatch unexpectedly signalled the registered descendant"
  fi
  kill_owned_process "$leader" "$root/bin/pi"
  kill_owned_process "$child" "$child_nonce"
}

test_sigkill_controller_does_not_admit_replacement() {
  local root driver child child_nonce leader before rc=0 attempt
  root="$(mktemp -d "$TEST_TMP/sigkill.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_MODE=signal TURN_TIMEOUT_SECONDS=10 \
      DESCENDANT_LOCK_READY_FILE="$root/descendant-lock-ready" \
      driver_env "$root" "synthetic sigkill task" \
      >"$root/stdout.1" 2>"$root/stderr.1"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_control_value "$root" active_turn.active_role orchestrator; then
    fail "SIGKILL role never reached durable orchestrator binding: control=$(control_snapshot "$root")"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  if ! wait_for_role_child "$root/child-pids"; then
    fail "SIGKILL bound role did not start its descendant: control=$(control_snapshot "$root")"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  read -r child child_nonce < <(grep -v 'bystander' "$root/child-pids" | tail -n 1)
  if ! wait_for_file "$root/descendant-lock-ready"; then
    fail "SIGKILL descendant did not prove inherited lock: child_alive=$(owned_process_alive "$child" "$child_nonce" && printf yes || printf no) control=$(control_snapshot "$root")"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  if [[ "$(cat "$root/descendant-lock-ready")" != "$child_nonce" ]]; then
    fail "SIGKILL descendant lock proof did not match the owned child"
    return
  fi
  leader="$(awk '$1 == "orchestrator" { print $2; exit }' "$root/events")"
  printf '%s %s\n' "$leader" "$root/bin/pi" >>"$root/controller-pids"
  before="$(event_count "$root" orchestrator)"

  signal_owned_process KILL "$driver" "$root/scripts/pi-shepherd-loop.sh"
  wait "$driver" 2>/dev/null || true
  if ! wait_owned_gone "$driver" "$root/scripts/pi-shepherd-loop.sh"; then
    fail "SIGKILL controller identity remained live"
    return
  fi
  if ! owned_process_alive "$child" "$child_nonce"; then
    fail "SIGKILL fixture lost its surviving child before replacement check"
    return
  fi
  signal_owned_process KILL "$leader" "$root/bin/pi"
  if ! wait_owned_gone "$leader" "$root/bin/pi"; then
    fail "SIGKILL role leader identity remained live"
    return
  fi
  if ! owned_process_alive "$child" "$child_nonce"; then
    fail "SIGKILL fixture lost its descendant when only the role leader was removed"
    return
  fi

  FAKE_MODE=normal driver_env "$root" --resume >"$root/stdout.2" 2>"$root/stderr.2" || rc=$?
  if [[ "$rc" -ne 75 ]] || ! grep -q 'CONTROLLER_HELD' "$root/stderr.2"; then
    fail "replacement after controller SIGKILL was not rejected by inherited lock"
  fi
  if [[ "$(event_count "$root" orchestrator)" -ne "$before" ]]; then
    fail "replacement launched a second orchestrator after controller SIGKILL"
  fi

  signal_owned_process KILL "$child" "$child_nonce"
  if ! wait_owned_gone "$child" "$child_nonce"; then
    fail "SIGKILL descendant identity remained live"
    return
  fi
  : >"$root/events"
  for ((attempt=0; attempt<200; attempt++)); do
    rc=0
    FAKE_MODE=normal driver_env "$root" --resume >"$root/stdout.3" 2>"$root/stderr.3" || rc=$?
    [[ "$rc" -ne 75 ]] && break
    /bin/sleep 0.01
  done
  if [[ "$rc" -ne 4 || -s "$root/events" ]] || ! grep -q 'RECOVERY_REQUIRED' "$root/stderr.3"; then
    fail "post-crash quiescence did not require an explicit recovery decision"
  fi
}

test_fence_movement_fails_closed() {
  local root driver child child_nonce rc=0 attempt
  root="$(mktemp -d "$TEST_TMP/fence.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_MODE=signal TURN_TIMEOUT_SECONDS=10 driver_env "$root" "synthetic fence task" \
      >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_file "$root/.planning/auto-loop/CONTROL.json" || ! wait_for_file "$root/child-pids"; then
    fail "fence test did not create active control state"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  for ((attempt=0; attempt<200; attempt++)); do
    [[ "$(control_field "$root" active_turn.active_role)" == "orchestrator" ]] && break
    /bin/sleep 0.01
  done
  if [[ "$(control_field "$root" active_turn.active_role)" != "orchestrator" ]]; then
    fail "fence test did not durably bind the orchestrator process group"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  read -r child child_nonce < <(grep -v 'bystander' "$root/child-pids" | tail -n 1)
  python3 - "$root/.planning/auto-loop/CONTROL.json" <<'PY'
import json, os, pathlib, uuid
path = pathlib.Path(os.sys.argv[1])
value = json.loads(path.read_text())
value["controller_id"] = str(uuid.uuid4())
tmp = path.with_suffix(".moved")
tmp.write_text(json.dumps(value, sort_keys=True) + "\n")
os.replace(tmp, path)
PY
  wait "$driver" || rc=$?

  if [[ "$rc" -ne 4 ]] || ! grep -q 'CONTROL_FENCE_MISMATCH' "$root/stderr"; then
    fail "fence movement did not stop controller with CONTROL_FENCE_MISMATCH"
  fi
  if [[ "$(event_count "$root" validator)" -ne 0 ]]; then
    fail "validator started after controller fence movement"
  fi
  if owned_process_alive "$child" "$child_nonce"; then
    fail "fence movement left descendant $child alive"
  fi
  if [[ "$(control_field "$root" children_quiescent)" == "true" ]] || \
     [[ -z "$(control_field "$root" active_turn.process_group_id)" ]]; then
    fail "fence movement falsely claimed quiescence or erased recovery handles"
  fi
}

test_failed_halt_persistence_never_claims_halted() {
  local root rc snapshot diagnostic
  root="$(mktemp -d "$TEST_TMP/halt-persistence.XXXXXX")"
  prepare_fixture "$root"

  FAKE_MODE=halt-latch-failure driver_env "$root" "synthetic halt persistence failure" \
    >"$root/stdout" 2>"$root/stderr"
  rc=$?
  if [[ "$rc" -ne 4 ]]; then
    fail "halt persistence failure exit=$rc, want 4"
  fi
  if [[ "$(control_field "$root" phase)" == "halted" ]] || \
     [[ -n "$(control_field "$root" halt.code)" ]]; then
    snapshot="$(control_snapshot "$root")"
    diagnostic="$(tail -n 12 "$root/.planning/auto-loop/driver.log" 2>/dev/null | tr '\n' '|')"
    fail "failed HALT persistence observed control=$snapshot halt=$(control_field "$root" halt.code) stderr=$diagnostic"
  fi
}

test_resume_requires_clean_paused_state() {
  local baseline root case_name rc before after
  baseline="$(mktemp -d "$TEST_TMP/clean-pause.XXXXXX")"
  prepare_fixture "$baseline"
  rc=0
  FAKE_MODE=normal MAX_TURNS=5 driver_env "$baseline" "synthetic clean pause" \
    >"$baseline/stdout" 2>"$baseline/stderr" || rc=$?
  if [[ "$rc" -ne 3 || "$(control_field "$baseline" phase)" != "paused" ]] || \
     [[ "$(control_field "$baseline" children_quiescent)" != "true" ]] || \
     [[ -n "$(control_field "$baseline" active_turn.turn_id)" ]] || \
     [[ -n "$(control_field "$baseline" halt.code)" ]] || \
     [[ ! -s "$baseline/.planning/auto-loop/PROMPT.txt" ]]; then
    fail "dirty-resume matrix could not establish a clean paused baseline"
    return
  fi

  for case_name in \
    active_turn children halt malformed_limit negative_ordinal boolean_generation \
    lease_null lease_missing_expiry lease_missing_heartbeat invalid_heartbeat invalid_expiry \
    reversed_lease equal_lease wrong_lease_interval invalid_run_id invalid_controller_id \
    invalid_updated_at extra_root_field duplicate_key nan_value infinity_value \
    huge_active_seconds reverts_over_limit no_verdict_at_limit
  do
    root="$(mktemp -d "$TEST_TMP/dirty-$case_name.XXXXXX")"
    prepare_fixture "$root"
    cp "$baseline/.planning/auto-loop/CONTROL.json" "$root/.planning/auto-loop/CONTROL.json"
    cp "$baseline/.planning/auto-loop/PROMPT.txt" "$root/.planning/auto-loop/PROMPT.txt"
    python3 - "$root/.planning/auto-loop/CONTROL.json" "$case_name" <<'PY'
import json, os, pathlib
path = pathlib.Path(os.sys.argv[1])
case_name = os.sys.argv[2]
value = json.loads(path.read_text())
if case_name == "active_turn":
    value["active_turn"] = {"turn_id": "stale"}
elif case_name == "children":
    value["children_quiescent"] = False
elif case_name == "halt":
    value["halt"] = {"code": "STALE"}
elif case_name == "malformed_limit":
    value["limits"]["max_turns"] = "invalid"
elif case_name == "negative_ordinal":
    value["turn_ordinal"] = -1
elif case_name == "boolean_generation":
    value["generation"] = True
elif case_name == "lease_null":
    value["lease"] = None
elif case_name == "lease_missing_expiry":
    value["lease"].pop("expires_at", None)
elif case_name == "lease_missing_heartbeat":
    value["lease"].pop("heartbeat_at", None)
elif case_name == "invalid_heartbeat":
    value["lease"]["heartbeat_at"] = "tomorrow"
elif case_name == "invalid_expiry":
    value["lease"]["expires_at"] = "tomorrow"
elif case_name == "reversed_lease":
    value["lease"]["heartbeat_at"] = "2099-01-01T00:00:00Z"
    value["lease"]["expires_at"] = "2000-01-01T00:00:00Z"
elif case_name == "equal_lease":
    value["lease"]["expires_at"] = value["lease"]["heartbeat_at"]
elif case_name == "wrong_lease_interval":
    import datetime as dt
    heartbeat = dt.datetime.strptime(value["lease"]["heartbeat_at"], "%Y-%m-%dT%H:%M:%SZ")
    value["lease"]["expires_at"] = (heartbeat + dt.timedelta(seconds=4)).strftime("%Y-%m-%dT%H:%M:%SZ")
elif case_name == "invalid_run_id":
    value["run_id"] = "not-a-uuid"
elif case_name == "invalid_controller_id":
    value["controller_id"] = "not-a-uuid"
elif case_name == "invalid_updated_at":
    value["updated_at"] = "tomorrow"
elif case_name == "extra_root_field":
    value["unexpected"] = "accepted-by-a-lax-reader"
elif case_name == "duplicate_key":
    raw = json.dumps(value, sort_keys=True, separators=(",", ":"))
    path.write_text(raw[:-1] + ',"run_id":' + json.dumps(value["run_id"]) + "}\n")
    raise SystemExit(0)
elif case_name == "nan_value":
    value["counters"]["active_seconds"] = float("nan")
elif case_name == "infinity_value":
    value["counters"]["active_seconds"] = float("inf")
elif case_name == "huge_active_seconds":
    value["counters"]["active_seconds"] = 315360001
elif case_name == "reverts_over_limit":
    value["counters"]["reverts"] = value["limits"]["max_reverts"] + 1
elif case_name == "no_verdict_at_limit":
    value["counters"]["no_verdict"] = value["limits"]["max_no_verdict"]
tmp = path.with_suffix(".dirty")
tmp.write_text(json.dumps(value, sort_keys=True, separators=(",", ":")) + "\n")
os.replace(tmp, path)
PY
    before="$(shasum -a 256 "$root/.planning/auto-loop/CONTROL.json" | awk '{print $1}')"
    : >"$root/events"
    rc=0
    FAKE_MODE=normal driver_env "$root" --resume >"$root/stdout" 2>"$root/stderr" || rc=$?
    after="$(shasum -a 256 "$root/.planning/auto-loop/CONTROL.json" | awk '{print $1}')"
    if [[ "$rc" -eq 0 || -s "$root/events" || -s "$root/model-pids" || "$before" != "$after" ]] || \
       ! grep -q 'CONTROL_STATE_INVALID' "$root/stderr"; then
      fail "resume accepted or rewrote isolated dirty-paused invariant $case_name"
    fi
  done
}

test_control_path_aliases_fail_closed() {
  local root mode rc source source_hash after_hash inode_source inode_control links
  for mode in dangling_symlink live_symlink hardlink; do
    root="$(mktemp -d "$TEST_TMP/control-$mode.XXXXXX")"
    prepare_fixture "$root"
    source="$root/.planning/auto-loop/CONTROL.source"
    case "$mode" in
      dangling_symlink)
        ln -s CONTROL.missing "$root/.planning/auto-loop/CONTROL.json"
        ;;
      live_symlink)
        write_valid_released_control "$source"
        ln -s CONTROL.source "$root/.planning/auto-loop/CONTROL.json"
        ;;
      hardlink)
        write_valid_released_control "$source"
        ln "$source" "$root/.planning/auto-loop/CONTROL.json"
        read -r inode_source _ < <(python3 - "$source" <<'PY'
import os, sys
value = os.stat(sys.argv[1])
print(value.st_ino, value.st_nlink)
PY
)
        ;;
    esac
    source_hash=""
    [[ -f "$source" ]] && source_hash="$(shasum -a 256 "$source" | awk '{print $1}')"
    rc=0
    FAKE_MODE=normal driver_env "$root" "synthetic unsafe control path" \
      >"$root/stdout" 2>"$root/stderr" || rc=$?
    if [[ "$rc" -ne 4 || -s "$root/events" || -s "$root/model-pids" ]] || \
       ! grep -q 'CONTROL_STATE_UNSAFE' "$root/stderr"; then
      fail "$mode control path did not fail before prompt/provider access"
    fi
    if [[ -n "$source_hash" ]]; then
      after_hash="$(shasum -a 256 "$source" | awk '{print $1}')"
      [[ "$source_hash" == "$after_hash" ]] || fail "$mode control source bytes changed"
    fi
    if [[ "$mode" == *symlink && ! -L "$root/.planning/auto-loop/CONTROL.json" ]]; then
      fail "$mode control path was replaced instead of rejected"
    elif [[ "$mode" == "hardlink" ]]; then
      read -r inode_control links < <(python3 - "$root/.planning/auto-loop/CONTROL.json" <<'PY'
import os, sys
value = os.stat(sys.argv[1])
print(value.st_ino, value.st_nlink)
PY
)
      if [[ "$inode_source" != "$inode_control" || "$links" -ne 2 ]]; then
        fail "hardlinked control path was replaced or rewritten"
      fi
    fi
  done
}

test_state_directory_alias_fails_before_effects() {
  local root external level alias rc
  for level in planning auto-loop; do
    rc=0
    root="$(mktemp -d "$TEST_TMP/state-root-$level.XXXXXX")"
    external="$(mktemp -d "$TEST_TMP/external-state-$level.XXXXXX")"
    prepare_fixture "$root"
    if [[ "$level" == "planning" ]]; then
      mv "$root/.planning" "$root/.planning.original"
      alias="$root/.planning"
    else
      mv "$root/.planning/auto-loop" "$root/.planning/auto-loop.original"
      alias="$root/.planning/auto-loop"
    fi
    ln -s "$external" "$alias"

    FAKE_MODE=normal driver_env "$root" "synthetic aliased $level state root" \
      >"$root/stdout" 2>"$root/stderr" || rc=$?
    if [[ "$rc" -ne 4 || -s "$root/events" || -s "$root/model-pids" ]] || \
       [[ -n "$(find "$external" -mindepth 1 -maxdepth 1 -print -quit 2>/dev/null)" ]] || \
       ! grep -q 'CONTROL_STATE_DIR_UNSAFE' "$root/stderr"; then
      fail "aliased $level directory was not rejected before controller/provider effects: rc=$rc"
    fi
    if [[ ! -L "$alias" ]]; then
      fail "unsafe $level alias was replaced instead of rejected"
    fi
  done
}

test_persisted_deadline_survives_delayed_begin_return() {
  local root driver helper helper_nonce rc=0 start end
  root="$(mktemp -d "$TEST_TMP/begin-deadline.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_BEGIN_RETURN_DELAY=4 TURN_TIMEOUT_SECONDS=2 \
      driver_env "$root" "synthetic delayed begin return" >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_file "$root/begin-commit-probe"; then
    fail "delayed-begin deadline fixture did not commit the canonical deadline"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  start="$(monotonic_now)"
  wait_driver_bounded "$root" "$driver" 800 || fail "delayed-begin controller exceeded its bounded teardown"
  wait "$driver" || rc=$?
  end="$(monotonic_now)"
  read -r helper helper_nonce < <(awk 'END { print $1, $2 }' "$root/controller-pids")
  wait_owned_gone "$helper" "$helper_nonce" 200 || fail "delayed-begin helper survived controller teardown"
  if [[ "$rc" -ne 4 || -s "$root/events" ]] || \
     [[ "$(control_field "$root" phase)" != "halted" ]] || \
     [[ "$(control_field "$root" halt.code)" != "TURN_DEADLINE" ]] || \
     ! duration_less_than "$start" "$end" 4.0; then
    fail "controller re-anchored or ignored the persisted deadline after begin returned: rc=$rc control=$(control_snapshot "$root")"
  fi
}

test_signal_during_delayed_begin_drains_helper() {
  local root driver helper helper_nonce child child_nonce rc=0 start end
  root="$(mktemp -d "$TEST_TMP/begin-signal.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_BEGIN_RETURN_DELAY=8 TURN_TIMEOUT_SECONDS=10 \
      driver_env "$root" "synthetic delayed begin signal" >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_file "$root/begin-commit-probe" || ! wait_for_role_child "$root/child-pids"; then
    fail "delayed-begin signal fixture did not commit its turn"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  read -r helper helper_nonce < <(awk 'END { print $1, $2 }' "$root/controller-pids")
  read -r child child_nonce < <(awk '$2 ~ /control-begin/ { print $1, $2; exit }' "$root/child-pids")
  start="$(monotonic_now)"
  signal_owned_process TERM "$driver" "$root/scripts/pi-shepherd-loop.sh"
  wait "$driver" 2>/dev/null || rc=$?
  end="$(monotonic_now)"
  if [[ "$rc" -ne 4 || -s "$root/events" ]] || \
     [[ "$(control_field "$root" phase)" != "recovery_required" ]] || \
     [[ "$(control_field "$root" halt.code)" != "CONTROLLER_SIGNAL" ]] || \
     ! duration_less_than "$start" "$end" 5.0 || \
     ! wait_owned_gone "$helper" "$helper_nonce" 200 || \
     ! wait_owned_gone "$child" "$child_nonce" 200; then
    fail "signal during delayed begin did not drain its state-helper group: rc=$rc control=$(control_snapshot "$root")"
  fi
}

test_role_readiness_is_bounded_by_persisted_deadline() {
  local root driver helper helper_nonce rc=0 start end
  root="$(mktemp -d "$TEST_TMP/readiness-deadline.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_LAUNCH_DELAY=6 TURN_TIMEOUT_SECONDS=3 \
      driver_env "$root" "synthetic delayed role readiness" >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_file "$root/launch-probe"; then
    fail "role-readiness deadline fixture did not enter launcher delay"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  start="$(monotonic_now)"
  wait_driver_bounded "$root" "$driver" 1000 || fail "role-readiness controller exceeded its bounded teardown"
  wait "$driver" || rc=$?
  end="$(monotonic_now)"
  read -r helper helper_nonce < <(awk 'END { print $1, $2 }' "$root/role-pids")
  wait_owned_gone "$helper" "$helper_nonce" 200 || fail "delayed role launcher survived deadline teardown"
  if [[ "$rc" -ne 4 || -s "$root/events" ]] || \
     [[ "$(control_field "$root" phase)" != "halted" ]] || \
     [[ "$(control_field "$root" halt.code)" != "TURN_DEADLINE" ]] || \
     ! duration_less_than "$start" "$end" 6.0; then
    fail "role readiness outlived the persisted turn deadline: rc=$rc control=$(control_snapshot "$root")"
  fi
}

test_false_ready_cannot_create_an_unbounded_wait() {
  local root driver helper helper_nonce child child_nonce rc=0 start end
  root="$(mktemp -d "$TEST_TMP/false-ready.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_EARLY_READY_DELAY=6 TURN_TIMEOUT_SECONDS=3 \
      driver_env "$root" "synthetic false role readiness" >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_file "$root/early-ready-probe" || ! wait_for_role_child "$root/child-pids"; then
    fail "false-ready fixture did not publish its premature readiness tree"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  read -r child child_nonce < <(awk '$2 ~ /false-ready/ { print $1, $2; exit }' "$root/child-pids")
  start="$(monotonic_now)"
  wait_driver_bounded "$root" "$driver" 1000 || fail "false-ready controller exceeded its bounded teardown"
  wait "$driver" || rc=$?
  end="$(monotonic_now)"
  read -r helper helper_nonce < <(awk 'END { print $1, $2 }' "$root/role-pids")
  wait_owned_gone "$helper" "$helper_nonce" 200 || fail "false-ready launcher survived deadline teardown"
  wait_owned_gone "$child" "$child_nonce" 200 || fail "false-ready descendant survived group teardown"
  if [[ "$rc" -ne 4 || -s "$root/events" ]] || \
     [[ "$(control_field "$root" phase)" != "halted" ]] || \
     [[ "$(control_field "$root" halt.code)" != "TURN_DEADLINE" ]] || \
     [[ "$(control_field "$root" children_quiescent)" != "true" ]] || \
     ! duration_less_than "$start" "$end" 6.0; then
    fail "premature readiness caused a post-deadline blocking wait: rc=$rc control=$(control_snapshot "$root")"
  fi
}

test_delayed_assert_cannot_authorize_after_deadline() {
  local root driver helper helper_nonce rc=0 start end
  root="$(mktemp -d "$TEST_TMP/assert-deadline.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_ASSERT_DELAY=8 TURN_TIMEOUT_SECONDS=4 \
      driver_env "$root" "synthetic delayed authorization assert" >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_file "$root/assert-probe"; then
    fail "delayed-assert fixture did not reach the authorization assertion"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  start="$(monotonic_now)"
  wait_driver_bounded "$root" "$driver" 1200 || fail "delayed-assert controller exceeded its bounded teardown"
  wait "$driver" || rc=$?
  end="$(monotonic_now)"
  read -r helper helper_nonce < <(awk 'END { print $1, $2 }' "$root/controller-pids")
  wait_owned_gone "$helper" "$helper_nonce" 200 || fail "delayed assertion helper survived deadline teardown"
  if [[ "$rc" -ne 4 || -s "$root/events" ]] || \
     [[ "$(control_field "$root" phase)" != "halted" ]] || \
     [[ "$(control_field "$root" halt.code)" != "TURN_DEADLINE" ]] || \
     ! duration_less_than "$start" "$end" 7.0; then
    fail "delayed authority assertion authorized work after deadline: rc=$rc control=$(control_snapshot "$root")"
  fi
}

test_control_helper_orphans_are_drained() {
  local root child child_nonce rc=0
  root="$(mktemp -d "$TEST_TMP/control-helper-orphan.XXXXXX")"
  prepare_fixture "$root"

  FAKE_CONTROL_ORPHAN=1 TURN_TIMEOUT_SECONDS=10 \
    driver_env "$root" "synthetic control-helper orphan" >"$root/stdout" 2>"$root/stderr" || rc=$?
  read -r child child_nonce < <(awk '$2 ~ /control-orphan/ { print $1, $2; exit }' "$root/child-pids")
  if [[ -z "$child" || "$rc" -ne 4 || -s "$root/events" ]] || \
     [[ "$(control_field "$root" phase)" != "recovery_required" ]] || \
     ! wait_owned_gone "$child" "$child_nonce" 200; then
    fail "orphaned control helper escaped registered group teardown: rc=$rc control=$(control_snapshot "$root")"
  fi
}

test_control_helper_observation_failure_is_conservative() {
  local root rc=0 before after
  root="$(mktemp -d "$TEST_TMP/control-helper-observation.XXXXXX")"
  prepare_fixture "$root"

  FAKE_PS_LIST_FAIL=1 TURN_TIMEOUT_SECONDS=10 \
    driver_env "$root" "synthetic helper observation failure" >"$root/stdout.1" 2>"$root/stderr.1" || rc=$?
  if [[ "$rc" -ne 4 || -s "$root/events" ]]; then
    fail "failed helper-group observation was treated as quiescent: rc=$rc"
    return
  fi
  before="$(shasum -a 256 "$root/.planning/auto-loop/CONTROL.json" | awk '{print $1}')"
  rc=0
  FAKE_MODE=normal driver_env "$root" --resume >"$root/stdout.2" 2>"$root/stderr.2" || rc=$?
  after="$(shasum -a 256 "$root/.planning/auto-loop/CONTROL.json" | awk '{print $1}')"
  if [[ "$rc" -ne 4 || -s "$root/events" || "$before" != "$after" ]]; then
    fail "helper observation failure admitted a replacement controller"
  fi
}

test_static_fifo_injection_cannot_authorize_work() {
  local root driver helper helper_nonce rc=0 premature=0 attempt
  root="$(mktemp -d "$TEST_TMP/fifo-injection.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_INTRUDE_ROLE_GO=1 FAKE_ASSERT_BARRIER=1 TURN_TIMEOUT_SECONDS=20 \
      driver_env "$root" "synthetic authorization injection" >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_file "$root/intruder-probe" || ! wait_for_file "$root/assert-probe"; then
    fail "FIFO-injection fixture did not reach the pre-authorization race"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  for ((attempt=0; attempt<500; attempt++)); do
    if [[ -s "$root/events" ]] || \
       compgen -G "$root/.planning/auto-loop/.role-authorized.*" >/dev/null; then
      premature=1
      break
    fi
    /bin/sleep 0.01
  done
  : >"$root/assert-release"
  if ! wait_owned_gone "$driver" "$root/scripts/pi-shepherd-loop.sh" 2000; then
    fail "authorization-channel fixture did not finish after assertion release"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
  fi
  wait "$driver" || rc=$?
  read -r helper helper_nonce < <(awk 'END { print $1, $2 }' "$root/controller-pids")
  wait_owned_gone "$helper" "$helper_nonce" 200 || fail "authorization assertion helper survived release"
  if [[ "$premature" -ne 0 ]]; then
    fail "static same-UID FIFO input authorized a role before the controller assertion completed"
  fi
  if [[ "$rc" -ne 3 ]] || \
     [[ "$(event_count "$root" orchestrator)" -ne 1 || "$(event_count "$root" validator)" -ne 1 ]]; then
    fail "authorization-channel defense denied the legitimate controller: rc=$rc control=$(control_snapshot "$root")"
  fi
}

test_flooded_authorization_channel_is_bounded() {
  local root flooder flooder_nonce rc=0 start end
  root="$(mktemp -d "$TEST_TMP/fifo-flood.XXXXXX")"
  prepare_fixture "$root"

  start="$(monotonic_now)"
  FAKE_FLOOD_ROLE_GO=1 TURN_TIMEOUT_SECONDS=4 \
    driver_env "$root" "synthetic authorization flood" >"$root/stdout" 2>"$root/stderr" || rc=$?
  end="$(monotonic_now)"
  if ! awk '$1 > 0 && $2 == "blocked" { found=1 } END { exit !found }' "$root/flood-probe" 2>/dev/null; then
    fail "authorization-flood fixture did not prove successful writes through EAGAIN"
  fi
  if [[ "$rc" -ne 4 || -s "$root/events" ]] || \
     [[ "$(control_field "$root" phase)" != "recovery_required" && \
        "$(control_field "$root" phase)" != "halted" ]] || \
     ! duration_less_than "$start" "$end" 8.0; then
    fail "flooded authorization channel escaped bounded fail-closed teardown: rc=$rc control=$(control_snapshot "$root")"
  fi
  read -r flooder flooder_nonce < <(awk 'NR == 1 { print $1, $2 }' "$root/role-pids")
  kill_owned_process "$flooder" "$flooder_nonce"
}

test_uncertain_terminal_commit_blocks_reentry() {
  local root transition mode rc before after prompt_before prompt_after
  for transition in pause release; do
    root="$(mktemp -d "$TEST_TMP/uncertain-$transition.XXXXXX")"
    prepare_fixture "$root"
    mode=normal
    [[ "$transition" == "release" ]] && mode=proceed-human-gate
    rc=0
    FAKE_MODE="$mode" FAKE_TERMINAL_UNCERTAIN="$transition" MAX_TURNS=5 \
      driver_env "$root" "synthetic uncertain $transition" >"$root/stdout.1" 2>"$root/stderr.1" || rc=$?
    if [[ "$rc" -ne 4 || "$(control_field "$root" phase)" != "recovery_required" ]] || \
       [[ "$(control_field "$root" halt.code)" != "CONTROL_COMMIT_UNCERTAIN" ]] || \
       [[ -n "$(control_field "$root" active_turn.turn_id)" ]] || \
       [[ "$(control_field "$root" children_quiescent)" != "true" ]]; then
      fail "uncertain $transition commit did not latch recovery: rc=$rc control=$(control_snapshot "$root")"
      continue
    fi
    before="$(shasum -a 256 "$root/.planning/auto-loop/CONTROL.json" | awk '{print $1}')"
    prompt_before="$(shasum -a 256 "$root/.planning/auto-loop/PROMPT.txt" | awk '{print $1}')"
    : >"$root/events"
    : >"$root/model-pids"
    rc=0
    if [[ "$transition" == "pause" ]]; then
      FAKE_MODE=normal driver_env "$root" --resume >"$root/stdout.2" 2>"$root/stderr.2" || rc=$?
    else
      FAKE_MODE=normal driver_env "$root" "synthetic reentry after uncertain release" \
        >"$root/stdout.2" 2>"$root/stderr.2" || rc=$?
    fi
    after="$(shasum -a 256 "$root/.planning/auto-loop/CONTROL.json" | awk '{print $1}')"
    prompt_after="$(shasum -a 256 "$root/.planning/auto-loop/PROMPT.txt" | awk '{print $1}')"
    if [[ "$rc" -ne 4 || -s "$root/events" || -s "$root/model-pids" ]] || \
       [[ "$before" != "$after" || "$prompt_before" != "$prompt_after" ]] || \
       ! grep -q 'RECOVERY_REQUIRED' "$root/stderr.2"; then
      fail "controller admitted or rewrote reentry after uncertain $transition"
    fi
  done
}

test_terminal_guard_survives_failed_recovery_write() {
  local root transition mode expected_phase rc prompt_before prompt_after snapshot
  for transition in pause release; do
    root="$(mktemp -d "$TEST_TMP/guarded-$transition.XXXXXX")"
    prepare_fixture "$root"
    mode=normal
    expected_phase=paused
    [[ "$transition" == "release" ]] && mode=proceed-human-gate
    [[ "$transition" == "release" ]] && expected_phase=released
    rc=0
    FAKE_MODE="$mode" FAKE_TERMINAL_UNCERTAIN="$transition" \
      FAKE_RECOVER_UNCERTAIN_FAIL=1 MAX_TURNS=5 \
      driver_env "$root" "synthetic guarded $transition" >"$root/stdout.1" 2>"$root/stderr.1" || rc=$?
    if [[ "$rc" -ne 4 || "$(control_field "$root" phase)" != "$expected_phase" ]] || \
       [[ ! -f "$root/.planning/auto-loop/CONTROL.transition" ]]; then
      fail "failed $transition recovery did not retain its durable transition guard: rc=$rc control=$(control_snapshot "$root")"
      continue
    fi
    prompt_before="$(shasum -a 256 "$root/.planning/auto-loop/PROMPT.txt" | awk '{print $1}')"
    : >"$root/events"
    : >"$root/model-pids"
    rc=0
    if [[ "$transition" == "pause" ]]; then
      FAKE_MODE=normal driver_env "$root" --resume >"$root/stdout.2" 2>"$root/stderr.2" || rc=$?
    else
      FAKE_MODE=normal driver_env "$root" "synthetic guarded release reentry" \
        >"$root/stdout.2" 2>"$root/stderr.2" || rc=$?
    fi
    prompt_after="$(shasum -a 256 "$root/.planning/auto-loop/PROMPT.txt" | awk '{print $1}')"
    if [[ "$rc" -ne 4 || -s "$root/events" || -s "$root/model-pids" ]] || \
       [[ "$prompt_before" != "$prompt_after" ]] || \
       [[ "$(control_field "$root" phase)" != "recovery_required" ]] || \
       [[ "$(control_field "$root" halt.code)" != "CONTROL_COMMIT_UNCERTAIN" ]] || \
       [[ ! -f "$root/.planning/auto-loop/CONTROL.transition" ]]; then
      fail "terminal guard admitted reentry after failed $transition recovery"
      continue
    fi
    snapshot="$(shasum -a 256 "$root/.planning/auto-loop/CONTROL.json" | awk '{print $1}')"
    rc=0
    FAKE_MODE=normal driver_env "$root" --resume >"$root/stdout.3" 2>"$root/stderr.3" || rc=$?
    if [[ "$rc" -ne 4 || "$snapshot" != \
          "$(shasum -a 256 "$root/.planning/auto-loop/CONTROL.json" | awk '{print $1}')" ]]; then
      fail "recovery-latched transition guard was not idempotently non-resumable"
    fi
  done
}

test_trace_distill_is_supervised_by_turn_deadline() {
  local root driver child child_nonce rc=0 start end
  root="$(mktemp -d "$TEST_TMP/trace-deadline.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_MODE=proceed-human-gate FAKE_TRACE_DELAY=20 TURN_TIMEOUT_SECONDS=12 \
      driver_env "$root" "synthetic hanging trace" >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_file "$root/trace-probe" || ! wait_for_role_child "$root/child-pids"; then
    fail "trace supervision fixture did not start digest work"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  read -r child child_nonce < <(awk '$2 ~ /-trace-/ { print $1, $2; exit }' "$root/child-pids")
  start="$(monotonic_now)"
  wait_driver_bounded "$root" "$driver" 2500 || fail "trace controller exceeded its bounded teardown"
  wait "$driver" || rc=$?
  end="$(monotonic_now)"
  wait_owned_gone "$child" "$child_nonce" 200 || fail "trace deadline left its descendant alive"
  if [[ "$rc" -ne 4 ]] || \
     [[ "$(control_field "$root" phase)" != "halted" ]] || \
     [[ "$(control_field "$root" halt.code)" != "TURN_DEADLINE" ]] || \
     checkpoint_marker_exists "$root" || \
     ! duration_less_than "$start" "$end" 15.0; then
    fail "trace digest escaped the persisted turn deadline: rc=$rc control=$(control_snapshot "$root")"
  fi
}

test_terminal_requires_shepherd_ratification() {
  local root rc

  root="$(mktemp -d "$TEST_TMP/unratified-human-gate.XXXXXX")"
  prepare_fixture "$root"
  rc=0
  FAKE_MODE=retry-human-gate driver_env "$root" "synthetic unratified human gate" \
    >"$root/stdout" 2>"$root/stderr" || rc=$?
  if [[ "$rc" -ne 3 || "$(control_field "$root" phase)" != "paused" ]] || \
     grep -q 'DONE: human-ready' "$root/stderr" || \
     ! grep -q 'verdict=RETRY' "$root/stderr" || \
     ! grep -q 'RETRY — repeat the synthetic stage' "$root/stderr"; then
    fail "RETRY incorrectly ratified a human-gate terminal"
  fi

  root="$(mktemp -d "$TEST_TMP/unratified-budget.XXXXXX")"
  prepare_fixture "$root"
  rc=0
  FAKE_MODE=no-verdict-budget driver_env "$root" "synthetic budget stop" \
    >"$root/stdout" 2>"$root/stderr" || rc=$?
  if [[ "$rc" -ne 3 || "$(control_field "$root" phase)" != "paused" ]] || \
     ! grep -q 'STOP: budget ceiling' "$root/stderr"; then
    fail "budget stop depended on an unavailable validator verdict"
  fi
}

test_nonzero_validator_cannot_ratify_or_checkpoint() {
  local root rc=0
  root="$(mktemp -d "$TEST_TMP/nonzero-validator.XXXXXX")"
  prepare_fixture "$root"
  FAKE_MODE=validator-nonzero-proceed driver_env "$root" "synthetic nonzero validator" \
    >"$root/stdout" 2>"$root/stderr" || rc=$?
  if [[ "$rc" -ne 3 || "$(control_field "$root" phase)" != "paused" ]] || \
     checkpoint_marker_exists "$root" || grep -q 'DONE: human-ready' "$root/stderr"; then
    fail "nonzero validator ratified or checkpointed its PROCEED artifact: rc=$rc control=$(control_snapshot "$root")"
  fi
}

test_trace_cannot_rewrite_snapshotted_verdict() {
  local root rc=0
  root="$(mktemp -d "$TEST_TMP/trace-verdict-rewrite.XXXXXX")"
  prepare_fixture "$root"
  FAKE_MODE=trace-overwrite FAKE_TRACE_OVERWRITE=1 \
    driver_env "$root" "synthetic trace verdict rewrite" >"$root/stdout" 2>"$root/stderr" || rc=$?
  if [[ "$rc" -ne 3 || "$(control_field "$root" phase)" != "paused" ]] || \
     checkpoint_marker_exists "$root" || grep -q 'DONE: human-ready' "$root/stderr" || \
     ! grep -q 'verdict=RETRY' "$root/stderr"; then
    fail "trace mutated the validator result before controller action: rc=$rc control=$(control_snapshot "$root")"
  fi
}

test_checkpoint_publication_is_fail_closed() {
  local root case_name rc
  for case_name in invalid-run missing-head aliased-root; do
    root="$(mktemp -d "$TEST_TMP/checkpoint-failure.XXXXXX")"
    prepare_fixture "$root"
    rc=0
    if [[ "$case_name" == "invalid-run" ]]; then
      FAKE_MODE=checkpoint-invalid-run driver_env "$root" "synthetic invalid checkpoint RUN" \
        >"$root/stdout" 2>"$root/stderr" || rc=$?
    elif [[ "$case_name" == "missing-head" ]]; then
      FAKE_GIT_HEAD_FAIL=1 driver_env "$root" "synthetic missing checkpoint HEAD" \
        >"$root/stdout" 2>"$root/stderr" || rc=$?
    else
      FAKE_MODE=checkpoint-alias driver_env "$root" "synthetic aliased checkpoint root" \
        >"$root/stdout" 2>"$root/stderr" || rc=$?
    fi
    if [[ "$rc" -ne 4 ]] || checkpoint_marker_exists "$root" || \
       [[ "$(control_field "$root" phase)" != "recovery_required" && \
          "$(control_field "$root" phase)" != "halted" ]]; then
      fail "checkpoint $case_name failure published success: rc=$rc control=$(control_snapshot "$root")"
    fi
    if [[ "$case_name" == "aliased-root" ]] && \
       { [[ ! -L "$root/.planning/auto-loop/checkpoints" ]] || \
         [[ "$(cat "$root/outside-checkpoints/sentinel" 2>/dev/null)" != "outside-sentinel" ]] || \
         [[ -e "$root/outside-checkpoints/LAST_GOOD" ]] || \
         [[ -d "$root/outside-checkpoints/1" ]]; }; then
      fail "checkpoint publication followed its aliased root outside controller state"
    fi
  done
}

test_checkpoint_bundle_aliases_fail_closed() {
  local root alias_type bundle outside before after rc
  for alias_type in symlink hardlink; do
    root="$(mktemp -d "$TEST_TMP/checkpoint-bundle-alias.XXXXXX")"
    prepare_fixture "$root"
    FAKE_MODE=normal driver_env "$root" "synthetic checkpoint alias seed" \
      >"$root/stdout.1" 2>"$root/stderr.1" || true
    bundle="$(find "$root/.planning/auto-loop/checkpoints" -type d -path '*/run-*/*' -name 1 -print -quit)"
    if [[ -z "$bundle" || ! -f "$bundle/RUN.json" ]]; then
      fail "checkpoint $alias_type fixture did not publish a seed bundle"
      continue
    fi
    outside="$root/outside-$alias_type-RUN.json"
    cp "$bundle/RUN.json" "$outside"
    rm "$bundle/RUN.json"
    if [[ "$alias_type" == "symlink" ]]; then
      ln -s "$outside" "$bundle/RUN.json"
    else
      ln "$outside" "$bundle/RUN.json"
    fi
    before="$(shasum -a 256 "$outside" | awk '{print $1}')"
    rc=0
    FAKE_MODE=revert driver_env "$root" --resume >"$root/stdout.2" 2>"$root/stderr.2" || rc=$?
    after="$(shasum -a 256 "$outside" | awk '{print $1}')"
    if [[ "$rc" -ne 3 || "$before" != "$after" ]] || \
       [[ "$(control_field "$root" counters.reverts)" != "0" ]] || \
       [[ -e "$root/.planning/auto-loop/REVERT-CLEANUP.json" ]] || \
       ! grep -q 'verdict=NONE' "$root/stderr.2"; then
      fail "checkpoint $alias_type bundle was trusted or mutated: rc=$rc control=$(control_snapshot "$root")"
    fi
  done
}

test_new_run_cannot_restore_previous_run_checkpoint() {
  local root first_run_id second_run_id rc=0
  root="$(mktemp -d "$TEST_TMP/cross-run-checkpoint.XXXXXX")"
  prepare_fixture "$root"
  FAKE_MODE=proceed-human-gate driver_env "$root" "synthetic first released run" \
    >"$root/stdout.1" 2>"$root/stderr.1"
  first_run_id="$(control_field "$root" run_id)"
  if ! checkpoint_marker_exists "$root"; then
    fail "cross-run fixture did not publish its first checkpoint"
    return
  fi
  FAKE_MODE=revert driver_env "$root" "synthetic second run revert" \
    >"$root/stdout.2" 2>"$root/stderr.2" || rc=$?
  second_run_id="$(control_field "$root" run_id)"
  if [[ "$rc" -ne 3 || -z "$first_run_id" || -z "$second_run_id" || "$first_run_id" == "$second_run_id" ]] || \
     [[ "$(control_field "$root" counters.reverts)" != "0" ]] || \
     [[ -e "$root/.planning/auto-loop/REVERT-CLEANUP.json" ]] || \
     ! grep -q 'verdict=NONE' "$root/stderr.2"; then
    fail "new run restored or accepted a previous run checkpoint: rc=$rc control=$(control_snapshot "$root")"
  fi
}

test_malformed_verdicts_cannot_ratify_or_checkpoint() {
  local root case_name rc
  for case_name in missing-trajectory low-proceed empty-reason extra-key proceed-correction high-score-revert revert-target-mismatch; do
    root="$(mktemp -d "$TEST_TMP/malformed-verdict.XXXXXX")"
    prepare_fixture "$root"
    rc=0
    if [[ "$case_name" == "revert-target-mismatch" || "$case_name" == "high-score-revert" ]]; then
      FAKE_MODE=normal driver_env "$root" "synthetic checkpoint seed" \
        >"$root/stdout.seed" 2>"$root/stderr.seed" || true
      : >"$root/events"
    fi
    if [[ "$case_name" == "revert-target-mismatch" || "$case_name" == "high-score-revert" ]]; then
      FAKE_MODE=malformed-human-gate FAKE_VERDICT_CASE="$case_name" \
        driver_env "$root" --resume >"$root/stdout" 2>"$root/stderr" || rc=$?
    else
      FAKE_MODE=malformed-human-gate FAKE_VERDICT_CASE="$case_name" \
        driver_env "$root" "synthetic malformed verdict $case_name" \
        >"$root/stdout" 2>"$root/stderr" || rc=$?
    fi
    if [[ "$rc" -ne 3 || "$(control_field "$root" phase)" != "paused" ]] || \
       grep -q 'DONE: human-ready' "$root/stderr"; then
      fail "malformed verdict $case_name ratified or checkpointed a turn: rc=$rc control=$(control_snapshot "$root")"
    fi
    if [[ "$case_name" == "revert-target-mismatch" || "$case_name" == "high-score-revert" ]] && \
       { [[ "$(control_field "$root" counters.reverts)" != "0" ]] || \
         [[ -e "$root/.planning/auto-loop/REVERT-CLEANUP.json" ]] || \
         ! grep -q 'verdict=NONE' "$root/stderr"; }; then
      fail "mismatched REVERT target was accepted after a real checkpoint"
    elif [[ "$case_name" != "revert-target-mismatch" && "$case_name" != "high-score-revert" ]] && checkpoint_marker_exists "$root"; then
      fail "malformed verdict $case_name published a checkpoint"
    fi
  done
}

test_natural_reserved_exit_codes_are_not_supervisor_failures() {
  local root rc=0
  root="$(mktemp -d "$TEST_TMP/natural-provider-124.XXXXXX")"
  prepare_fixture "$root"
  FAKE_MODE=reserved-exit TURN_TIMEOUT_SECONDS=20 \
    driver_env "$root" "synthetic natural provider 124" >"$root/stdout.1" 2>"$root/stderr.1" || rc=$?
  if [[ "$rc" -ne 3 || "$(control_field "$root" phase)" != "paused" ]] || \
     [[ "$(event_count "$root" orchestrator)" -ne 1 || "$(event_count "$root" validator)" -ne 1 ]]; then
    fail "natural provider exit 124 was conflated with a supervisor deadline: rc=$rc"
  fi

  root="$(mktemp -d "$TEST_TMP/natural-trace-127.XXXXXX")"
  prepare_fixture "$root"
  rc=0
  FAKE_MODE=proceed-human-gate FAKE_TRACE_EXIT=127 TURN_TIMEOUT_SECONDS=20 \
    driver_env "$root" "synthetic natural trace 127" >"$root/stdout.2" 2>"$root/stderr.2" || rc=$?
  if [[ "$rc" -ne 0 || "$(control_field "$root" phase)" != "released" ]] || \
     ! grep -q 'trace digest returned 127' "$root/stderr.2"; then
    fail "natural trace exit 127 was conflated with a supervisor failure: rc=$rc"
  fi
}

test_run_counters_survive_pause_resume() {
  local root rc active_before active_after
  root="$(mktemp -d "$TEST_TMP/counters.XXXXXX")"
  prepare_fixture "$root"
  FAKE_MODE=normal FAKE_MODEL_DELAY=1 MAX_TURNS=5 driver_env "$root" "synthetic counter task" \
    >"$root/stdout.1" 2>"$root/stderr.1" || true
  FAKE_MODE=revert FAKE_MODEL_DELAY=1 MAX_TURNS=5 driver_env "$root" --resume \
    >"$root/stdout.2" 2>"$root/stderr.2" || true
  if [[ "$(control_field "$root" counters.reverts)" != "1" ]] || \
     [[ "$(control_field "$root" counters.no_verdict)" != "0" ]]; then
    fail "revert counter was not persisted before pause"
    return
  fi
  if ! grep -q '"stage":"PLAN"' "$root/.planning/auto-loop/RUN.json" || \
     ! grep -q '"checkpoint": "1"' "$root/.planning/auto-loop/REVERT-CLEANUP.json" || \
     ! grep -q 'restored checkpoint 1' "$root/stderr.2"; then
    fail "valid REVERT did not restore and identify its exact run-bound checkpoint"
    return
  fi
  active_before="$(control_field "$root" counters.active_seconds)"
  FAKE_MODE=no-verdict FAKE_MODEL_DELAY=1 MAX_TURNS=5 driver_env "$root" --resume >"$root/stdout.3" 2>"$root/stderr.3"
  rc=$?
  if [[ "$rc" -ne 3 || "$(control_field "$root" counters.reverts)" != "1" ]] || \
     [[ "$(control_field "$root" counters.no_verdict)" != "1" ]]; then
    fail "run counters reset or drifted across pause/resume"
  fi
  active_after="$(control_field "$root" counters.active_seconds)"
  if [[ ! "$active_before" =~ ^[0-9]+$ || ! "$active_after" =~ ^[0-9]+$ ]] || \
     (( active_after <= active_before )); then
    fail "durable active-time counter did not advance across resume"
  fi
}

test_turn_limit_survives_pause_resume() {
  local root first_count rc orchestrator_pgid validator_pgid snapshot phase ordinal active_turn quiescent diagnostic
  root="$(mktemp -d "$TEST_TMP/turn-limit.XXXXXX")"
  prepare_fixture "$root"

  FAKE_MODE=normal MAX_TURNS=1 driver_env "$root" "synthetic capped task" \
    >"$root/stdout.1" 2>"$root/stderr.1"
  rc=$?
  first_count="$(event_count "$root" orchestrator)"
  snapshot="$(control_snapshot "$root")"
  IFS=$'\t' read -r phase ordinal active_turn quiescent <<<"$snapshot"
  if [[ "$rc" -ne 3 || "$first_count" -ne 1 || "$phase" != "paused" || "$ordinal" != "1" ]] || \
     [[ "$active_turn" != "null" || "$quiescent" != "true" ]]; then
    diagnostic="$(tail -n 6 "$root/stderr.1" 2>/dev/null | tr '\n' '|')"
    fail "turn limit observed rc=$rc orchestrators=$first_count control=$snapshot stderr=$diagnostic; want 3/1/paused/1/null/true"
    return
  fi
  if [[ -s "$root/state-fd-leaks" ]]; then
    fail "provider inherited the controller state-directory descriptor"
  fi
  if ! awk '$1 == "orchestrator" && $3 == "openai-codex/gpt-5.5" && $4 == "default" { found=1 } END { exit !found }' "$root/events"; then
    fail "orchestrator model policy drifted while updating the Shepherd"
  fi
  if ! awk '$1 == "validator" && $3 == "openai-codex/gpt-5.6-sol" && $4 == "high" { found=1 } END { exit !found }' "$root/events"; then
    fail "Shepherd did not use exact gpt-5.6-sol high policy"
  fi
  orchestrator_pgid="$(awk '$1 == "orchestrator" { print $5; exit }' "$root/events")"
  validator_pgid="$(awk '$1 == "validator" { print $5; exit }' "$root/events")"
  if [[ -z "$orchestrator_pgid" || -z "$validator_pgid" || "$orchestrator_pgid" == "$validator_pgid" ]] || \
     ! awk '$1 == "orchestrator" || $1 == "validator" { if ($2 != $5) bad=1; seen[$1]=1 } END { exit bad || !seen["orchestrator"] || !seen["validator"] }' "$root/events"; then
    fail "orchestrator and Shepherd did not run in distinct self-led process groups"
  fi

  FAKE_MODE=normal MAX_TURNS=999 driver_env "$root" --resume >"$root/stdout.2" 2>"$root/stderr.2"
  rc=$?
  snapshot="$(control_snapshot "$root")"
  IFS=$'\t' read -r phase ordinal active_turn quiescent <<<"$snapshot"
  if [[ "$rc" -ne 3 || "$(event_count "$root" orchestrator)" -ne "$first_count" ]] || \
     [[ "$phase" != "paused" || "$ordinal" != "1" || "$active_turn" != "null" || "$quiescent" != "true" ]]; then
    fail "resume cap observed rc=$rc orchestrators=$(event_count "$root" orchestrator) control=$snapshot"
  fi
}

TEST_NAMES=(
  test_concurrent_controllers_have_one_winner
  test_state_directory_replacement_cannot_split_controller_lock
  test_hard_deadline_drains_process_tree
  test_leader_exit_with_live_descendant_halts
  test_validator_shares_the_hard_turn_deadline
  test_missing_validator_model_fails_before_work
  test_slow_model_preflight_renews_epoch_lease
  test_halt_is_durable_and_blocks_resume
  test_signal_during_startup_is_durable
  test_signal_during_role_startup_does_not_orphan
  test_role_stays_inert_until_durable_bind
  test_role_is_not_authorized_after_deadline
  test_preseeded_role_go_cannot_launch_provider
  test_controller_sigkill_before_role_go_never_launches_provider
  test_signal_drains_group_and_requires_recovery
  test_pgid_mismatch_never_signals_untrusted_pid
  test_sigkill_controller_does_not_admit_replacement
  test_fence_movement_fails_closed
  test_failed_halt_persistence_never_claims_halted
  test_resume_requires_clean_paused_state
  test_control_path_aliases_fail_closed
  test_state_directory_alias_fails_before_effects
  test_persisted_deadline_survives_delayed_begin_return
  test_signal_during_delayed_begin_drains_helper
  test_role_readiness_is_bounded_by_persisted_deadline
  test_false_ready_cannot_create_an_unbounded_wait
  test_delayed_assert_cannot_authorize_after_deadline
  test_control_helper_orphans_are_drained
  test_control_helper_observation_failure_is_conservative
  test_static_fifo_injection_cannot_authorize_work
  test_flooded_authorization_channel_is_bounded
  test_uncertain_terminal_commit_blocks_reentry
  test_terminal_guard_survives_failed_recovery_write
  test_trace_distill_is_supervised_by_turn_deadline
  test_terminal_requires_shepherd_ratification
  test_nonzero_validator_cannot_ratify_or_checkpoint
  test_trace_cannot_rewrite_snapshotted_verdict
  test_checkpoint_publication_is_fail_closed
  test_checkpoint_bundle_aliases_fail_closed
  test_new_run_cannot_restore_previous_run_checkpoint
  test_malformed_verdicts_cannot_ratify_or_checkpoint
  test_natural_reserved_exit_codes_are_not_supervisor_failures
  test_run_counters_survive_pause_resume
  test_turn_limit_survives_pause_resume
)

validate_test_filter() {
  local request known seen="," found
  local -a requests
  [[ -n "${SHEPHERD_TEST_FILTER:-}" ]] || return 0
  if [[ ",${SHEPHERD_TEST_FILTER}," == *",,"* ]]; then
    fail "invalid Shepherd test filter: empty test name"
    return 1
  fi
  IFS=',' read -r -a requests <<<"$SHEPHERD_TEST_FILTER"
  for request in "${requests[@]}"; do
    if [[ -z "$request" || "$seen" == *",${request},"* ]]; then
      fail "invalid or duplicate Shepherd test filter: ${request:-<empty>}"
      return 1
    fi
    found=0
    for known in "${TEST_NAMES[@]}"; do
      [[ "$request" == "$known" ]] && found=$((found + 1))
    done
    if [[ "$found" -ne 1 ]]; then
      fail "unknown Shepherd test filter: $request"
      return 1
    fi
    seen="${seen}${request},"
  done
}

filter_valid=1
validate_test_filter || filter_valid=0
if [[ "${SHEPHERD_REQUIRE_FULL:-0}" == "1" && -n "${SHEPHERD_TEST_FILTER:-}" ]]; then
  fail "full Shepherd gate refuses SHEPHERD_TEST_FILTER"
  filter_valid=0
fi
if (( filter_valid == 1 )); then
  for test_name in "${TEST_NAMES[@]}"; do
    run_test "$test_name"
  done
  if (( executed_tests == 0 )); then
    fail "Shepherd test filter selected zero tests"
  fi
  if [[ "${SHEPHERD_REQUIRE_FULL:-0}" == "1" && "$executed_tests" -ne "${#TEST_NAMES[@]}" ]]; then
    fail "full Shepherd gate executed $executed_tests/${#TEST_NAMES[@]} tests"
  fi
fi

if (( failures > 0 )); then
  printf 'pi-shepherd-supervision: %d failure(s)\n' "$failures" >&2
  exit 1
fi

printf 'pi-shepherd-supervision: ok\n'
