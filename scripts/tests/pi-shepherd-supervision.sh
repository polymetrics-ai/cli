#!/usr/bin/env bash
set -u

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TEST_TMP="$(mktemp -d)"
failures=0
SUITE_PID=$$
WATCHDOG_PID=""

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
  done < <(find "$TEST_TMP" \( -name child-pids -o -name controller-pids \) -type f 2>/dev/null)
  rm -rf "$TEST_TMP"
}
trap cleanup EXIT
trap 'cleanup; exit 124' INT TERM

(
  /bin/sleep 60
  kill -TERM "$SUITE_PID" 2>/dev/null || true
) &
WATCHDOG_PID=$!

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  failures=$((failures + 1))
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
  : >"$root/controller-pids"
  printf 'shepherd-test-%s-%s-%s\n' "$$" "$RANDOM" "${root##*/}" >"$root/nonce"

  cat >"$root/scripts/loop-trace.sh" <<'SH'
#!/usr/bin/env bash
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
trap '' TERM
/bin/sleep "$duration"
SH
  chmod +x "$root/bin/test-child"

  cat >"$root/bin/pi" <<'SH'
#!/usr/bin/env bash
set -u

if [[ " $* " == *" --offline --list-models "* ]]; then
  printf '%s\n' \
    'provider      model          context  max-out  thinking  images' \
    'openai-codex  gpt-5.6-sol    372K     128K     yes       yes'
  exit 0
fi

model=""
previous=""
for argument in "$@"; do
  if [[ "$previous" == "--model" ]]; then
    model="$argument"
    break
  fi
  previous="$argument"
done

role="orchestrator"
if [[ "$model" == "openai-codex/gpt-5.6-sol" ]]; then
  role="validator"
fi
printf '%s %s\n' "$role" "$$" >>"$EVENT_LOG"

write_verdict() {
  local verdict="$1"
  printf '{"verdict":"%s","step_score":5,"reason":"synthetic","correction":null}\n' "$verdict" \
    >"$TEST_REPO/.planning/auto-loop/VALIDATOR-VERDICT.json"
}

if [[ "$role" == "validator" ]]; then
  if [[ "$FAKE_MODE" == "halt" ]]; then
    write_verdict HALT
  else
    write_verdict PROCEED
  fi
  exit 0
fi

case "$FAKE_MODE" in
  concurrent)
    /bin/sleep 1
    ;;
  hang-tree)
    child_nonce="$TEST_NONCE-role-$$"
    "$TEST_REPO/bin/test-child" "$child_nonce" 6 &
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
  shift
  env -i \
    PATH="$root/bin:$PATH" \
    HOME="$root/home" \
    PI_BIN="$root/bin/pi" \
    VALIDATOR_BIN="$root/bin/pi" \
    ORCH_MODEL="openai-codex/gpt-5.5" \
    VALIDATOR_ARGS="" \
    PI_EXTRA_FLAGS="" \
    EVENT_LOG="$root/events" \
    CHILD_PID_FILE="$root/child-pids" \
    TEST_NONCE="$(cat "$root/nonce")" \
    TEST_REPO="$root" \
    FAKE_MODE="${FAKE_MODE:-normal}" \
    MAX_ITERATIONS="${MAX_ITERATIONS:-1}" \
    MAX_TURNS="${MAX_TURNS:-20}" \
    TURN_TIMEOUT_SECONDS="${TURN_TIMEOUT_SECONDS:-2}" \
    TERM_GRACE_SECONDS="${TERM_GRACE_SECONDS:-1}" \
    CONTROL_HEARTBEAT_SECONDS="${CONTROL_HEARTBEAT_SECONDS:-1}" \
    COOLDOWN_SECONDS=0 \
    STALL_MINUTES=1 \
    "$root/scripts/pi-shepherd-loop.sh" "$@"
}

register_controller() {
  local root="$1" pid="$2"
  printf '%s %s\n' "$pid" "$root/scripts/pi-shepherd-loop.sh" >>"$root/controller-pids"
}

start_bystander() {
  local root="$1" nonce pid
  nonce="$(cat "$root/nonce")-bystander-$RANDOM"
  "$root/bin/test-child" "$nonce" 30 >/dev/null 2>&1 &
  pid=$!
  printf '%s %s\n' "$pid" "$nonce" >>"$root/child-pids"
  BYSTANDER_IDENTITY="$pid $nonce"
}

assert_bystander_alive() {
  local identity="$1" context="$2" pid nonce
  read -r pid nonce <<<"$identity"
  if ! owned_process_alive "$pid" "$nonce"; then
    fail "$context killed unrelated bystander $pid"
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

wait_for_file() {
  local path="$1" attempts="${2:-200}"
  local i
  for ((i=0; i<attempts; i++)); do
    [[ -s "$path" ]] && return 0
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
      FAKE_MODE=concurrent driver_env "$root" "synthetic concurrent task" \
        >"$root/stdout.$i" 2>"$root/stderr.$i"
    ) &
    pids[$i]=$!
    register_controller "$root" "${pids[$i]}"
  done

  for pid in "${pids[@]}"; do
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
  local root child child_nonce rc bystander start end elapsed_fast
  root="$(mktemp -d "$TEST_TMP/deadline.XXXXXX")"
  prepare_fixture "$root"
  start_bystander "$root"
  bystander="$BYSTANDER_IDENTITY"

  start="$(monotonic_now)"
  FAKE_MODE=hang-tree TURN_TIMEOUT_SECONDS=1 driver_env "$root" "synthetic deadline task" \
    >"$root/stdout" 2>"$root/stderr"
  rc=$?
  end="$(monotonic_now)"
  read -r child child_nonce < <(grep -v 'bystander' "$root/child-pids" | tail -n 1)
  elapsed_fast="$(python3 - "$start" "$end" <<'PY'
import sys
print("yes" if float(sys.argv[2]) - float(sys.argv[1]) < 4.0 else "no")
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

test_halt_is_durable_and_blocks_resume() {
  local root rc before after
  root="$(mktemp -d "$TEST_TMP/halt.XXXXXX")"
  prepare_fixture "$root"

  FAKE_MODE=halt driver_env "$root" "synthetic halt task" >"$root/stdout.1" 2>"$root/stderr.1"
  rc=$?
  if [[ "$rc" -ne 4 || "$(control_field "$root" phase)" != "halted" ]]; then
    fail "validator HALT was not durably latched before exit"
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
}

test_signal_drains_group_and_requires_recovery() {
  local root driver child child_nonce phase bystander
  root="$(mktemp -d "$TEST_TMP/signal.XXXXXX")"
  prepare_fixture "$root"
  start_bystander "$root"
  bystander="$BYSTANDER_IDENTITY"

  (
    FAKE_MODE=signal TURN_TIMEOUT_SECONDS=10 driver_env "$root" "synthetic signal task" \
      >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  register_controller "$root" "$driver"
  if ! wait_for_file "$root/child-pids"; then
    fail "signal test did not start descendant"
    kill -KILL "$driver" 2>/dev/null || true
    return
  fi
  read -r child child_nonce < <(grep -v 'bystander' "$root/child-pids" | tail -n 1)
  kill -TERM "$driver" 2>/dev/null || true
  wait "$driver" 2>/dev/null || true
  /bin/sleep 0.1

  if owned_process_alive "$child" "$child_nonce"; then
    fail "signal left descendant $child alive"
  fi
  phase="$(control_field "$root" phase)"
  if [[ "$phase" != "recovery_required" && "$phase" != "halted" ]]; then
    fail "signal persisted phase=$phase, want recovery_required or halted"
  fi
  assert_bystander_alive "$bystander" "signal teardown"
}

test_fence_movement_fails_closed() {
  local root driver child child_nonce rc=0
  root="$(mktemp -d "$TEST_TMP/fence.XXXXXX")"
  prepare_fixture "$root"

  (
    FAKE_MODE=signal TURN_TIMEOUT_SECONDS=10 driver_env "$root" "synthetic fence task" \
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
}

test_turn_limit_survives_pause_resume() {
  local root first_count rc
  root="$(mktemp -d "$TEST_TMP/turn-limit.XXXXXX")"
  prepare_fixture "$root"

  FAKE_MODE=normal MAX_TURNS=1 driver_env "$root" "synthetic capped task" \
    >"$root/stdout.1" 2>"$root/stderr.1"
  rc=$?
  first_count="$(event_count "$root" orchestrator)"
  if [[ "$rc" -ne 3 || "$first_count" -ne 1 || "$(control_field "$root" phase)" != "paused" ]] || \
     [[ "$(control_field "$root" turn_ordinal)" != "1" ]]; then
    fail "turn limit did not persist paused ordinal 1"
    return
  fi

  FAKE_MODE=normal MAX_TURNS=1 driver_env "$root" --resume >"$root/stdout.2" 2>"$root/stderr.2"
  rc=$?
  if [[ "$rc" -ne 3 || "$(event_count "$root" orchestrator)" -ne "$first_count" ]] || \
     [[ "$(control_field "$root" turn_ordinal)" != "1" ]]; then
    fail "resume reset or exceeded persisted turn cap"
  fi
}

test_concurrent_controllers_have_one_winner
test_hard_deadline_drains_process_tree
test_leader_exit_with_live_descendant_halts
test_halt_is_durable_and_blocks_resume
test_signal_drains_group_and_requires_recovery
test_fence_movement_fails_closed
test_turn_limit_survives_pause_resume

if (( failures > 0 )); then
  printf 'pi-shepherd-supervision: %d failure(s)\n' "$failures" >&2
  exit 1
fi

printf 'pi-shepherd-supervision: ok\n'
