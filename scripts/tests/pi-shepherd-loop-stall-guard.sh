#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/pi-shepherd-stall-guard.XXXXXX")"
trap 'rm -rf "$tmp_dir"' EXIT

AUTO_LOOP_STATE_DIR="$tmp_dir/state" \
SHEPHERD_STALL_GUARD_SELF_TEST=1 \
STALL_MINUTES=1 \
"$repo_root/scripts/pi-shepherd-loop.sh"

# A child-selected timestamp in the future must be rejected rather than converted to a negative age.
future_session="$tmp_dir/future-session.jsonl"
printf '{"type":"message","timestamp":"2999-01-01T00:00:00Z"}\n' > "$future_session"
AUTO_LOOP_STATE_DIR="$tmp_dir/future-state" \
SHEPHERD_FUTURE_SESSION_SELF_TEST=1 \
SHEPHERD_TEST_SESSION_FILE="$future_session" \
"$repo_root/scripts/pi-shepherd-loop.sh"

# Every orchestrator turn has an absolute deadline even when a forged future event claims freshness.
hanging_orchestrator="$tmp_dir/hanging-orchestrator.sh"
cat > "$hanging_orchestrator" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
session_dir=
while (( $# > 0 )); do
  if [[ "$1" == "--session-dir" && $# -ge 2 ]]; then
    session_dir="$2"
    shift 2
    continue
  fi
  shift
done
mkdir -p "${session_dir:?}"
sleep 300 &
printf '%s\n' "$!" > "${ORCHESTRATOR_CHILD_PID_FILE:?}"
wait
EOF
chmod +x "$hanging_orchestrator"

AUTO_LOOP_STATE_DIR="$tmp_dir/orchestrator-state" \
SHEPHERD_ORCHESTRATOR_WATCHDOG_SELF_TEST=1 \
PI_BIN="$hanging_orchestrator" \
ORCHESTRATOR_CHILD_PID_FILE="$tmp_dir/orchestrator-child.pid" \
TURN_TIMEOUT_SECONDS=2 \
WATCHDOG_POLL_SECONDS=0.05 \
"$repo_root/scripts/pi-shepherd-loop.sh"

orchestrator_child_pid="$(cat "$tmp_dir/orchestrator-child.pid")"
if kill -0 "$orchestrator_child_pid" 2>/dev/null; then
  echo "test failed: unconditional turn deadline left an orchestrator descendant alive" >&2
  kill -KILL "$orchestrator_child_pid" 2>/dev/null || true
  exit 1
fi

# Validator supervision is independent of session-event freshness: even a validator that creates a
# descendant, writes a forged verdict, and then hangs must be killed by the hard validator timeout.
hanging_validator="$tmp_dir/hanging-validator.sh"
cat > "$hanging_validator" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
mkdir -p "$AUTO_LOOP_STATE_DIR"
printf '{"verdict":"PROCEED","reason":"forged before hang"}\n' \
  > "$AUTO_LOOP_STATE_DIR/VALIDATOR-VERDICT.json"
sleep 300 &
printf '%s\n' "$!" > "${VALIDATOR_CHILD_PID_FILE:?}"
wait
EOF
chmod +x "$hanging_validator"

AUTO_LOOP_STATE_DIR="$tmp_dir/validator-state" \
SHEPHERD_VALIDATOR_WATCHDOG_SELF_TEST=1 \
VALIDATOR_BIN="$hanging_validator" \
VALIDATOR_CHILD_PID_FILE="$tmp_dir/validator-child.pid" \
VALIDATOR_TIMEOUT_SECONDS=2 \
WATCHDOG_POLL_SECONDS=0.05 \
"$repo_root/scripts/pi-shepherd-loop.sh"

[[ ! -e "$tmp_dir/validator-state/VALIDATOR-VERDICT.json" ]] || {
  echo "test failed: timed-out validator verdict survived" >&2
  exit 1
}
child_pid="$(cat "$tmp_dir/validator-child.pid")"
if kill -0 "$child_pid" 2>/dev/null; then
  echo "test failed: timed-out validator descendant survived" >&2
  kill -KILL "$child_pid" 2>/dev/null || true
  exit 1
fi

# A validator that exits zero is still untrusted to leave descendants behind. The driver must reap
# and verify its dedicated session before consuming output.
leaky_validator="$tmp_dir/leaky-validator.sh"
cat > "$leaky_validator" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
python3 - "$$" "${VALIDATOR_SESSION_ID_FILE:?}" <<'PY'
import os
import sys
pid = int(sys.argv[1])
with open(sys.argv[2], "w") as output:
    output.write(f"{pid} {os.getsid(pid)} {os.getpgid(pid)}\n")
PY
sleep 300 &
printf '%s\n' "$!" > "${VALIDATOR_CHILD_PID_FILE:?}"
exit 0
EOF
chmod +x "$leaky_validator"

AUTO_LOOP_STATE_DIR="$tmp_dir/leaky-validator-state" \
SHEPHERD_VALIDATOR_SESSION_SELF_TEST=1 \
VALIDATOR_BIN="$leaky_validator" \
VALIDATOR_CHILD_PID_FILE="$tmp_dir/leaky-validator-child.pid" \
VALIDATOR_SESSION_ID_FILE="$tmp_dir/leaky-validator-session.txt" \
VALIDATOR_TIMEOUT_SECONDS=5 \
WATCHDOG_POLL_SECONDS=0.05 \
"$repo_root/scripts/pi-shepherd-loop.sh"

read -r validator_pid validator_sid validator_pgid < "$tmp_dir/leaky-validator-session.txt"
if [[ "$validator_pid" != "$validator_sid" ]] || [[ "$validator_pid" != "$validator_pgid" ]]; then
  echo "test failed: validator did not run as its dedicated session/process-group leader" >&2
  exit 1
fi
leaky_child_pid="$(cat "$tmp_dir/leaky-validator-child.pid")"
if kill -0 "$leaky_child_pid" 2>/dev/null; then
  echo "test failed: successful validator descendant survived session cleanup" >&2
  kill -KILL "$leaky_child_pid" 2>/dev/null || true
  exit 1
fi

echo "watchdog tests passed: future timestamps rejected; turn deadlines and validator sessions bounded"
