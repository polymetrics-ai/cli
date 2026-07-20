#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/pi-shepherd-verdict-guard.XXXXXX")"
trap 'rm -rf "$tmp_dir"' EXIT

AUTO_LOOP_STATE_DIR="$tmp_dir/state" \
SHEPHERD_VERDICT_GUARD_SELF_TEST=1 \
"$repo_root/scripts/pi-shepherd-loop.sh"

fake_orchestrator="$tmp_dir/fake-orchestrator.sh"
fake_validator="$tmp_dir/fake-validator.sh"

cat > "$fake_orchestrator" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
mkdir -p "$AUTO_LOOP_STATE_DIR"
printf '{"stage":"review","terminal":"human_gate"}\n' > "$AUTO_LOOP_STATE_DIR/RUN.json"
EOF

cat > "$fake_validator" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
mkdir -p "$AUTO_LOOP_STATE_DIR"
printf '{"verdict":"PROCEED","step_score":1,"reason":"must be ignored"}\n' \
  > "$AUTO_LOOP_STATE_DIR/VALIDATOR-VERDICT.json"
exit 9
EOF

chmod +x "$fake_orchestrator" "$fake_validator"

set +e
AUTO_LOOP_STATE_DIR="$tmp_dir/failed-validator-state" \
PI_BIN="$fake_orchestrator" \
VALIDATOR_BIN="$fake_validator" \
MAX_ITERATIONS=1 \
MAX_NO_VERDICT=1 \
COOLDOWN_SECONDS=0 \
WATCHDOG_POLL_SECONDS=0.05 \
"$repo_root/scripts/pi-shepherd-loop.sh" "failed-validator regression" \
  > "$tmp_dir/failed-validator.log" 2>&1
rc=$?
set -e

if (( rc == 0 )); then
  echo "test failed: nonzero validator authorized terminal success" >&2
  cat "$tmp_dir/failed-validator.log" >&2
  exit 1
fi
if [[ -e "$tmp_dir/failed-validator-state/VALIDATOR-VERDICT.json" ]]; then
  echo "test failed: verdict from nonzero validator was retained" >&2
  exit 1
fi
grep -q 'validator returned non-zero' "$tmp_dir/failed-validator.log"
grep -q 'verdict=NONE' "$tmp_dir/failed-validator.log"

echo "main-loop test passed: failed validator cannot authorize terminal success"
