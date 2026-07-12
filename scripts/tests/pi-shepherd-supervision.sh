#!/usr/bin/env bash
set -u

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REAL_PYTHON_BIN="$(command -v python3)"
REAL_PS_BIN="$(command -v ps)"
TEST_TMP="$(mktemp -d)"
failures=0
executed_tests=0
SUITE_PID=$$
WATCHDOG_PID=""
SUITE_TIMEOUT_SECONDS="${SUITE_TIMEOUT_SECONDS:-120}"

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
  done < <(find "$TEST_TMP" \( -name child-pids -o -name controller-pids -o -name model-pids \) -type f 2>/dev/null)
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
  printf 'shepherd-test-%s-%s-%s\n' "$$" "$RANDOM" "${root##*/}" >"$root/nonce"

  cat >"$root/scripts/loop-trace.sh" <<'SH'
#!/usr/bin/env bash
set -u
if [[ "${FAKE_MODE:-}" == "halt-latch-failure" ]]; then
  mv "$TEST_REPO/.planning/auto-loop/CONTROL.json" "$TEST_REPO/.planning/auto-loop/CONTROL.backup"
  ln -s CONTROL.backup "$TEST_REPO/.planning/auto-loop/CONTROL.json"
fi
exit 0
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
  "$REAL_PYTHON_BIN" - "$AUTO_LOOP_CONTROL_FD" "$TEST_REPO/.planning/auto-loop/CONTROL.lock" \
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
if not stat.S_ISREG(held.st_mode) or (held.st_dev, held.st_ino) != (canonical.st_dev, canonical.st_ino):
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
if [[ "${2:-}" == *'.role-ready.'* && "${FAKE_PRESEED_ROLE_GO:-0}" == "1" ]]; then
  printf '%s\n' "$$" >"${3:?missing role go path}"
  printf 'preseeded\n' >"$PRESEED_PROBE_FILE"
fi
if [[ "${2:-}" == *'.role-ready.'* && "${FAKE_LAUNCH_DELAY:-0}" != "0" ]]; then
  printf 'launching\n' >"$LAUNCH_PROBE_FILE"
  /bin/sleep "$FAKE_LAUNCH_DELAY"
fi
if [[ "${3:-}" == "bind" && "${FAKE_BIND_DELAY:-0}" != "0" ]]; then
  printf 'binding\n' >"$BIND_PROBE_FILE"
  /bin/sleep "$FAKE_BIND_DELAY"
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
if [[ "${1:-}" == "-o" && "${2:-}" == "stat=" && "${3:-}" == "-p" && \
      "${FAKE_GO_PROBE_DELAY:-0}" != "0" && ! -e "$GO_PROBE_ONCE_FILE" ]] && \
   find "$TEST_REPO/.planning/auto-loop" -name '.role-go-ready.*' -type f -size +0c -print -quit 2>/dev/null | grep -q .; then
  printf '%s %s\n' "$$" "$TEST_REPO/bin/ps" >>"$CONTROLLER_PID_FILE"
  : >"$GO_PROBE_ONCE_FILE"
  printf 'probing\n' >"$GO_PROBE_FILE"
  /bin/sleep "$FAKE_GO_PROBE_DELAY"
fi
exec "$REAL_PS_BIN" "$@"
SH
  chmod +x "$root/bin/ps"

  cat >"$root/bin/pi" <<'SH'
#!/usr/bin/env bash
set -u

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
  printf '{"verdict":"%s","step_score":5,"reason":"synthetic","correction":null}\n' "$verdict" \
    >"$TEST_REPO/.planning/auto-loop/VALIDATOR-VERDICT.json"
}

if [[ "$role" == "validator" ]]; then
  if [[ "$FAKE_MODE" == "validator-hang" ]]; then
    child_nonce="$TEST_NONCE-validator-$$"
    "$TEST_REPO/bin/test-child" "$child_nonce" 10 &
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
  else
    write_verdict PROCEED
  fi
  exit 0
fi

case "$FAKE_MODE" in
  retry-human-gate|no-verdict-human-gate)
    printf '{"stage":"FINALIZE","terminal":"human_gate"}\n' >"$TEST_REPO/.planning/auto-loop/RUN.json"
    ;;
  no-verdict-budget)
    printf '{"stage":"EXECUTE","terminal":"budget"}\n' >"$TEST_REPO/.planning/auto-loop/RUN.json"
    ;;
esac

case "$FAKE_MODE" in
  validator-hang)
    /bin/sleep 3
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
    "NATURAL_EXIT_LOG=$root/natural-exits"
    "TEST_NONCE=$(cat "$root/nonce")"
    "TEST_REPO=$root"
    "FAKE_MODE=${FAKE_MODE:-normal}"
    "FAKE_MODEL_DELAY=${FAKE_MODEL_DELAY:-0}"
    "FAKE_MODEL_MISSING=${FAKE_MODEL_MISSING:-0}"
    "FAKE_LAUNCH_DELAY=${FAKE_LAUNCH_DELAY:-0}"
    "FAKE_BIND_DELAY=${FAKE_BIND_DELAY:-0}"
    "FAKE_PRESEED_ROLE_GO=${FAKE_PRESEED_ROLE_GO:-0}"
    "FAKE_GO_PROBE_DELAY=${FAKE_GO_PROBE_DELAY:-0}"
    "DESCENDANT_LOCK_READY_FILE=${DESCENDANT_LOCK_READY_FILE:-}"
    "MODEL_PROBE_FILE=$root/model-probe"
    "MODEL_COMPLETION_FILE=$root/model-completion"
    "LAUNCH_PROBE_FILE=$root/launch-probe"
    "BIND_PROBE_FILE=$root/bind-probe"
    "PRESEED_PROBE_FILE=$root/preseed-probe"
    "PGID_MISMATCH_ARM_FILE=$root/pgid-mismatch-arm"
    "GO_PROBE_FILE=$root/go-probe"
    "GO_PROBE_ONCE_FILE=$root/go-probe-once"
    "REAL_PYTHON_BIN=$REAL_PYTHON_BIN"
    "REAL_PS_BIN=$REAL_PS_BIN"
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

monotonic_now() {
  python3 - <<'PY'
import time
print(time.monotonic())
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
  local root driver child child_nonce rc=0 start end elapsed_fast
  root="$(mktemp -d "$TEST_TMP/validator-deadline.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_MODE=validator-hang TURN_TIMEOUT_SECONDS=8 \
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
  start="$(monotonic_now)"
  wait "$driver" || rc=$?
  end="$(monotonic_now)"
  elapsed_fast="$(python3 - "$start" "$end" <<'PY'
import sys
print("yes" if float(sys.argv[2]) - float(sys.argv[1]) < 7.5 else "no")
PY
)"

  if [[ "$rc" -ne 4 || "$elapsed_fast" != "yes" ]] || \
     [[ "$(event_count "$root" orchestrator)" -ne 1 || "$(event_count "$root" validator)" -ne 1 ]] || \
     [[ "$(control_field "$root" phase)" != "halted" ]] || \
     [[ "$(control_field "$root" halt.code)" != "TURN_DEADLINE" ]] || \
     [[ -f "$root/.planning/auto-loop/checkpoints/LAST_GOOD" ]]; then
    fail "validator did not share the persisted hard turn deadline"
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
  local root rc=0
  root="$(mktemp -d "$TEST_TMP/role-bind-deadline.XXXXXX")"
  prepare_fixture "$root"

  FAKE_BIND_DELAY=3 TURN_TIMEOUT_SECONDS=1 \
    driver_env "$root" "synthetic expired role bind" >"$root/stdout" 2>"$root/stderr" || rc=$?
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
  local root driver leader probe_pid state
  root="$(mktemp -d "$TEST_TMP/pre-go-sigkill.XXXXXX")"
  prepare_fixture "$root"

  (
    DRIVER_EXEC=1 FAKE_GO_PROBE_DELAY=5 TURN_TIMEOUT_SECONDS=15 \
      driver_env "$root" "synthetic pre-go sigkill" >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_file "$root/go-probe"; then
    fail "pre-go SIGKILL fixture did not reach the stopped role handshake"
    kill_owned_process "$driver" "$root/scripts/pi-shepherd-loop.sh"
    wait "$driver" 2>/dev/null || true
    return
  fi
  leader="$(control_field "$root" active_turn.leader_pid)"
  printf '%s %s\n' "$leader" "$REAL_PYTHON_BIN" >>"$root/controller-pids"
  signal_owned_process KILL "$driver" "$root/scripts/pi-shepherd-loop.sh"
  wait "$driver" 2>/dev/null || true
  probe_pid="$(awk '$2 ~ /\/bin\/ps$/ { pid=$1 } END { print pid }' "$root/controller-pids")"
  kill_owned_process "$probe_pid" "$root/bin/ps"
  state="$($REAL_PS_BIN -o stat= -p "$leader" 2>/dev/null | tr -d '[:space:]' || true)"
  if [[ -s "$root/events" || "$state" != T* ]]; then
    fail "controller SIGKILL released or lost the pre-go role: state=$state control=$(control_snapshot "$root")"
  fi
  kill_owned_process "$leader" "$REAL_PYTHON_BIN"
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

  for case_name in active_turn children halt malformed_limit negative_ordinal boolean_generation; do
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
tmp = path.with_suffix(".dirty")
tmp.write_text(json.dumps(value, sort_keys=True, separators=(",", ":")) + "\n")
os.replace(tmp, path)
PY
    before="$(shasum -a 256 "$root/.planning/auto-loop/CONTROL.json" | awk '{print $1}')"
    : >"$root/events"
    rc=0
    FAKE_MODE=normal driver_env "$root" --resume >"$root/stdout" 2>"$root/stderr" || rc=$?
    after="$(shasum -a 256 "$root/.planning/auto-loop/CONTROL.json" | awk '{print $1}')"
    if [[ "$rc" -eq 0 || -s "$root/events" || "$before" != "$after" ]]; then
      fail "resume accepted or rewrote isolated dirty-paused invariant $case_name"
    fi
  done
}

test_control_path_aliases_fail_closed() {
  local root mode rc source inode_source inode_control links
  for mode in dangling_symlink live_symlink hardlink; do
    root="$(mktemp -d "$TEST_TMP/control-$mode.XXXXXX")"
    prepare_fixture "$root"
    source="$root/.planning/auto-loop/CONTROL.source"
    case "$mode" in
      dangling_symlink)
        ln -s CONTROL.missing "$root/.planning/auto-loop/CONTROL.json"
        ;;
      live_symlink)
        printf '%s\n' '{"active_turn":null,"children_quiescent":true,"control_revision":1,"controller_id":"synthetic-controller","counters":{"active_seconds":0,"no_verdict":0,"reverts":0},"generation":1,"halt":null,"lease":{"expires_at":"2099-01-01T00:00:00Z","heartbeat_at":"2099-01-01T00:00:00Z"},"limits":{"heartbeat_seconds":1,"max_minutes":0,"max_no_verdict":3,"max_reverts":6,"max_turns":20,"term_grace_seconds":1,"turn_timeout_seconds":10},"phase":"released","run_id":"synthetic-run","schema_version":"1.0","turn_ordinal":0,"updated_at":"2099-01-01T00:00:00Z"}' >"$source"
        ln -s CONTROL.source "$root/.planning/auto-loop/CONTROL.json"
        ;;
      hardlink)
        printf '%s\n' '{"active_turn":null,"children_quiescent":true,"control_revision":1,"controller_id":"synthetic-controller","counters":{"active_seconds":0,"no_verdict":0,"reverts":0},"generation":1,"halt":null,"lease":{"expires_at":"2099-01-01T00:00:00Z","heartbeat_at":"2099-01-01T00:00:00Z"},"limits":{"heartbeat_seconds":1,"max_minutes":0,"max_no_verdict":3,"max_reverts":6,"max_turns":20,"term_grace_seconds":1,"turn_timeout_seconds":10},"phase":"released","run_id":"synthetic-run","schema_version":"1.0","turn_ordinal":0,"updated_at":"2099-01-01T00:00:00Z"}' >"$source"
        ln "$source" "$root/.planning/auto-loop/CONTROL.json"
        read -r inode_source _ < <(python3 - "$source" <<'PY'
import os, sys
value = os.stat(sys.argv[1])
print(value.st_ino, value.st_nlink)
PY
)
        ;;
    esac
    rc=0
    FAKE_MODE=normal driver_env "$root" "synthetic unsafe control path" \
      >"$root/stdout" 2>"$root/stderr" || rc=$?
    if [[ "$rc" -eq 0 || -s "$root/events" || -e "$root/model-probe" ]]; then
      fail "$mode control path did not fail before prompt/provider access"
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

test_terminal_requires_shepherd_ratification() {
  local root rc

  root="$(mktemp -d "$TEST_TMP/unratified-human-gate.XXXXXX")"
  prepare_fixture "$root"
  rc=0
  FAKE_MODE=retry-human-gate driver_env "$root" "synthetic unratified human gate" \
    >"$root/stdout" 2>"$root/stderr" || rc=$?
  if [[ "$rc" -ne 3 || "$(control_field "$root" phase)" != "paused" ]] || \
     grep -q 'DONE: human-ready' "$root/stderr"; then
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

test_run_counters_survive_pause_resume() {
  local root rc active_before active_after
  root="$(mktemp -d "$TEST_TMP/counters.XXXXXX")"
  prepare_fixture "$root"
  FAKE_MODE=revert FAKE_MODEL_DELAY=1 MAX_TURNS=5 driver_env "$root" "synthetic counter task" \
    >"$root/stdout.1" 2>"$root/stderr.1" || true
  if [[ "$(control_field "$root" counters.reverts)" != "1" ]] || \
     [[ "$(control_field "$root" counters.no_verdict)" != "0" ]]; then
    fail "revert counter was not persisted before pause"
    return
  fi
  active_before="$(control_field "$root" counters.active_seconds)"
  FAKE_MODE=no-verdict FAKE_MODEL_DELAY=1 MAX_TURNS=5 driver_env "$root" --resume >"$root/stdout.2" 2>"$root/stderr.2"
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
  test_hard_deadline_drains_process_tree
  test_leader_exit_with_live_descendant_halts
  test_validator_shares_the_hard_turn_deadline
  test_missing_validator_model_fails_before_work
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
  test_terminal_requires_shepherd_ratification
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
if (( filter_valid == 1 )); then
  for test_name in "${TEST_NAMES[@]}"; do
    run_test "$test_name"
  done
  if (( executed_tests == 0 )); then
    fail "Shepherd test filter selected zero tests"
  fi
fi

if (( failures > 0 )); then
  printf 'pi-shepherd-supervision: %d failure(s)\n' "$failures" >&2
  exit 1
fi

printf 'pi-shepherd-supervision: ok\n'
