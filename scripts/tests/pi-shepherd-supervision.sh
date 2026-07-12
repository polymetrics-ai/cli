#!/usr/bin/env bash
set -u

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TEST_TMP="$(mktemp -d)"
failures=0

cleanup() {
  local pid_file pid
  while IFS= read -r pid_file; do
    [[ -f "$pid_file" ]] || continue
    while IFS= read -r pid; do
      [[ "$pid" =~ ^[0-9]+$ ]] || continue
      kill -KILL "$pid" 2>/dev/null || true
    done <"$pid_file"
  done < <(find "$TEST_TMP" -name child-pids -type f 2>/dev/null)
  rm -rf "$TEST_TMP"
}
trap cleanup EXIT

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  failures=$((failures + 1))
}

copy_launcher_without_phase0_guard() {
  local destination="$1"
  # The production fuse has no enable route. Tests exercise the exact post-guard launcher body by
  # removing only the guard block from a temporary copy; no production flag or environment bypass
  # exists. PI_BIN is the first line after the block in the canonical script.
  awk '
    /^# AUTO_LOOP_RUN_ENTRYPOINT:/ { skipping=1 }
    skipping && /^PI_BIN=/ { skipping=0 }
    !skipping { print }
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
    "$root/.planning/auto-loop"

  copy_launcher_without_phase0_guard "$root/scripts/pi-shepherd-loop.sh"
  : >"$root/.pi/extensions/pi-sub-agent/index.ts"
  printf 'synthetic validator prompt\n' >"$root/.agents/agentic-delivery/prompts/shepherd-validator-prompt.md"
  printf '{"stage":"PLAN","terminal":null}\n' >"$root/.planning/auto-loop/RUN.json"
  : >"$root/events"
  : >"$root/child-pids"

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
printf '%s %s %s\n' "$role" "$$" "$(date +%s%N 2>/dev/null || date +%s)" >>"$EVENT_LOG"

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
    /bin/sleep 0.4
    ;;
  hang-tree)
    (
      trap '' TERM
      /bin/sleep 3
    ) &
    child=$!
    printf '%s\n' "$child" >>"$CHILD_PID_FILE"
    trap '' TERM
    wait "$child"
    ;;
  orphan)
    (
      trap '' TERM
      /bin/sleep 3
    ) &
    child=$!
    printf '%s\n' "$child" >>"$CHILD_PID_FILE"
    exit 0
    ;;
  signal)
    (
      trap '' TERM
      /bin/sleep 5
    ) &
    child=$!
    printf '%s\n' "$child" >>"$CHILD_PID_FILE"
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
  env \
    PATH="$root/bin:$PATH" \
    PI_BIN="$root/bin/pi" \
    VALIDATOR_BIN="$root/bin/pi" \
    EVENT_LOG="$root/events" \
    CHILD_PID_FILE="$root/child-pids" \
    TEST_REPO="$root" \
    MAX_ITERATIONS="${MAX_ITERATIONS:-1}" \
    MAX_TURNS="${MAX_TURNS:-20}" \
    TURN_TIMEOUT_SECONDS="${TURN_TIMEOUT_SECONDS:-2}" \
    TERM_GRACE_SECONDS="${TERM_GRACE_SECONDS:-1}" \
    CONTROL_HEARTBEAT_SECONDS="${CONTROL_HEARTBEAT_SECONDS:-1}" \
    COOLDOWN_SECONDS=0 \
    STALL_MINUTES=1 \
    "$root/scripts/pi-shepherd-loop.sh" "$@"
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
  done

  for pid in "${pids[@]}"; do
    wait "$pid"
    rc=$?
    case "$rc" in
      3|4) winners=$((winners + 1)) ;;
      75) held=$((held + 1)) ;;
      *) other=$((other + 1)) ;;
    esac
  done

  if [[ "$(event_count "$root" orchestrator)" -ne 1 || "$winners" -ne 1 || "$held" -ne 31 || "$other" -ne 0 ]]; then
    fail "32 concurrent controllers: orchestrators=$(event_count "$root" orchestrator) winners=$winners held=$held other=$other; want 1/1/31/0"
  fi
}

test_hard_deadline_drains_process_tree() {
  local root child rc
  root="$(mktemp -d "$TEST_TMP/deadline.XXXXXX")"
  prepare_fixture "$root"

  FAKE_MODE=hang-tree TURN_TIMEOUT_SECONDS=1 driver_env "$root" "synthetic deadline task" \
    >"$root/stdout" 2>"$root/stderr"
  rc=$?
  child="$(tail -n 1 "$root/child-pids" 2>/dev/null || true)"

  if [[ "$rc" -ne 4 ]]; then
    fail "hard deadline exit=$rc, want 4"
  fi
  if [[ "$(event_count "$root" validator)" -ne 0 ]]; then
    fail "validator started after hard deadline"
  fi
  if [[ "$child" =~ ^[0-9]+$ ]] && kill -0 "$child" 2>/dev/null; then
    fail "hard deadline left descendant $child alive"
  fi
  if [[ "$(control_field "$root" phase)" != "halted" ]] || \
     [[ "$(control_field "$root" children_quiescent)" != "true" ]]; then
    fail "hard deadline did not persist halted/quiescent control state"
  fi
}

test_leader_exit_with_live_descendant_halts() {
  local root child rc
  root="$(mktemp -d "$TEST_TMP/orphan.XXXXXX")"
  prepare_fixture "$root"

  FAKE_MODE=orphan driver_env "$root" "synthetic orphan task" >"$root/stdout" 2>"$root/stderr"
  rc=$?
  child="$(tail -n 1 "$root/child-pids" 2>/dev/null || true)"

  if [[ "$rc" -ne 4 || "$(event_count "$root" validator)" -ne 0 ]]; then
    fail "leader-exit orphan was accepted: exit=$rc validators=$(event_count "$root" validator)"
  fi
  if [[ "$child" =~ ^[0-9]+$ ]] && kill -0 "$child" 2>/dev/null; then
    fail "leader-exit orphan left descendant $child alive"
  fi
  if [[ "$(control_field "$root" halt.code)" != "ORCHESTRATOR_ORPHAN" ]]; then
    fail "leader-exit orphan did not persist ORCHESTRATOR_ORPHAN"
  fi
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
  chmod 000 "$root/.planning/auto-loop/PROMPT.txt"

  FAKE_MODE=normal driver_env "$root" --resume >"$root/stdout.2" 2>"$root/stderr.2"
  rc=$?
  chmod 600 "$root/.planning/auto-loop/PROMPT.txt"
  after="$(shasum -a 256 "$root/.planning/auto-loop/CONTROL.json" | awk '{print $1}')"

  if [[ "$rc" -ne 4 || -s "$root/events" || "$before" != "$after" ]] || \
     ! grep -q 'HALT_LATCHED' "$root/stderr.2"; then
    fail "halted resume was not rejected before prompt/provider access"
  fi
}

test_signal_drains_group_and_requires_recovery() {
  local root driver child phase
  root="$(mktemp -d "$TEST_TMP/signal.XXXXXX")"
  prepare_fixture "$root"

  (
    FAKE_MODE=signal TURN_TIMEOUT_SECONDS=10 driver_env "$root" "synthetic signal task" \
      >"$root/stdout" 2>"$root/stderr"
  ) &
  driver=$!
  if ! wait_for_file "$root/child-pids"; then
    fail "signal test did not start descendant"
    kill -KILL "$driver" 2>/dev/null || true
    return
  fi
  child="$(tail -n 1 "$root/child-pids")"
  kill -TERM "$driver" 2>/dev/null || true
  wait "$driver" 2>/dev/null || true
  /bin/sleep 0.1

  if [[ "$child" =~ ^[0-9]+$ ]] && kill -0 "$child" 2>/dev/null; then
    fail "signal left descendant $child alive"
  fi
  phase="$(control_field "$root" phase)"
  if [[ "$phase" != "recovery_required" && "$phase" != "halted" ]]; then
    fail "signal persisted phase=$phase, want recovery_required or halted"
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
  if [[ "$rc" -ne 3 || "$(control_field "$root" phase)" != "paused" ]] || \
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
test_turn_limit_survives_pause_resume

if (( failures > 0 )); then
  printf 'pi-shepherd-supervision: %d failure(s)\n' "$failures" >&2
  exit 1
fi

printf 'pi-shepherd-supervision: ok\n'
