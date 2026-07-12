#!/usr/bin/env bash
set -u

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
failures=0

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  failures=$((failures + 1))
}

make_stub() {
  local path="$1"
  local marker="$2"
  printf '#!/usr/bin/env bash\nprintf invoked > %q\nexit 99\n' "$marker" >"$path"
  chmod +x "$path"
}

prepare_repo() {
  local root="$1"
  mkdir -p "$root/scripts" "$root/.agents/agentic-delivery/prompts"
  cp "$REPO_ROOT/scripts/claude-auto-loop.sh" "$root/scripts/claude-auto-loop.sh"
  cp "$REPO_ROOT/scripts/pi-auto-loop.sh" "$root/scripts/pi-auto-loop.sh"
  if [[ -f "$REPO_ROOT/scripts/auto-loop-safety.sh" ]]; then
    cp "$REPO_ROOT/scripts/auto-loop-safety.sh" "$root/scripts/auto-loop-safety.sh"
  fi
  printf 'synthetic orchestrator prompt\n' >"$root/.agents/agentic-delivery/prompts/claude-orchestrator.md"
  printf 'synthetic validator prompt\n' >"$root/.agents/agentic-delivery/prompts/shepherd-validator-prompt.md"
}

assert_driver_denied() {
  local driver="$1"
  local mode="$2"
  local root
  local marker
  local rc
  root="$(mktemp -d)"
  marker="$root/process-invoked"
  prepare_repo "$root"
  make_stub "$root/fake-runtime" "$marker"

  if [[ "$mode" == "resume" ]]; then
    mkdir -p "$root/.planning/auto-loop"
    printf 'synthetic resume prompt' >"$root/.planning/auto-loop/PROMPT.txt"
  fi

  if [[ "$driver" == "claude" ]]; then
    CLAUDE_BIN="$root/fake-runtime" MAX_ITERATIONS=1 MAX_NO_VERDICT=1 COOLDOWN_SECONDS=0 \
      "$root/scripts/claude-auto-loop.sh" "$([[ "$mode" == "resume" ]] && printf '%s' --resume || printf '%s' 'synthetic task')" \
      >"$root/stdout" 2>"$root/stderr"
    rc=$?
  else
    PI_BIN="$root/fake-runtime" MAX_ITERATIONS=1 CONTINUE_SESSION=0 COOLDOWN_SECONDS=0 \
      "$root/scripts/pi-auto-loop.sh" "$([[ "$mode" == "resume" ]] && printf '%s' --resume || printf '%s' 'synthetic task')" \
      >"$root/stdout" 2>"$root/stderr"
    rc=$?
  fi

  if [[ "$rc" -ne 78 ]]; then
    fail "$driver $mode exit=$rc, want 78"
  fi
  if [[ -e "$marker" ]]; then
    fail "$driver $mode invoked a child process before the safety denial"
  fi
  if [[ "$mode" == "run" && -e "$root/.planning/auto-loop" ]]; then
    fail "$driver $mode persisted loop state before the safety denial"
  fi
  if ! grep -q 'AUTO_LOOP_DISABLED_PHASE_0' "$root/stderr"; then
    fail "$driver $mode did not emit the stable safety code"
  fi

  rm -rf "$root"
}

assert_help_available() {
  local driver="$1"
  local root
  local marker
  local rc
  root="$(mktemp -d)"
  marker="$root/process-invoked"
  prepare_repo "$root"
  make_stub "$root/fake-runtime" "$marker"

  if [[ "$driver" == "claude" ]]; then
    CLAUDE_BIN="$root/fake-runtime" MAX_ITERATIONS=1 MAX_NO_VERDICT=1 COOLDOWN_SECONDS=0 \
      "$root/scripts/claude-auto-loop.sh" --help >"$root/stdout" 2>"$root/stderr"
    rc=$?
  else
    PI_BIN="$root/fake-runtime" MAX_ITERATIONS=1 CONTINUE_SESSION=0 COOLDOWN_SECONDS=0 \
      "$root/scripts/pi-auto-loop.sh" --help >"$root/stdout" 2>"$root/stderr"
    rc=$?
  fi

  if [[ "$rc" -ne 0 ]]; then
    fail "$driver help exit=$rc, want 0"
  fi
  if [[ -e "$marker" || -e "$root/.planning/auto-loop" ]]; then
    fail "$driver help caused a process launch or state persistence"
  fi
  if ! grep -qi 'usage:' "$root/stdout"; then
    fail "$driver help did not print usage to stdout"
  fi

  rm -rf "$root"
}

for driver in claude pi; do
  assert_driver_denied "$driver" run
  assert_driver_denied "$driver" resume
  assert_help_available "$driver"
done

if [[ ! -x "$REPO_ROOT/scripts/auto-loop-safety.sh" ]]; then
  fail "scripts/auto-loop-safety.sh is missing or not executable"
else
  status_output="$($REPO_ROOT/scripts/auto-loop-safety.sh status --json)"
  if ! grep -q '"state":"closed"' <<<"$status_output"; then
    fail "safety status is not closed"
  fi

  expected="$($REPO_ROOT/scripts/auto-loop-safety.sh entrypoints)"
  actual="$(grep -l 'AUTO_LOOP_RUN_ENTRYPOINT' "$REPO_ROOT"/scripts/*auto-loop*.sh | sed "s|$REPO_ROOT/||" | sort)"
  if [[ "$actual" != "$expected" ]]; then
    fail "tracked entrypoint inventory differs from marked wrappers"
  fi
fi

if ! grep -q '^agent-loop-test:' "$REPO_ROOT/Makefile"; then
  fail "Makefile is missing agent-loop-test"
fi
if ! grep -E '^verify:.*agent-loop-test' "$REPO_ROOT/Makefile" >/dev/null; then
  fail "make verify does not include agent-loop-test"
fi

if (( failures > 0 )); then
  printf 'auto-loop-control: %d failure(s)\n' "$failures" >&2
  exit 1
fi

printf 'auto-loop-control: ok\n'
