#!/usr/bin/env bash
set -u

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
failures=0
TEST_TMP="$(mktemp -d)"
trap 'rm -rf "$TEST_TMP"' EXIT

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  failures=$((failures + 1))
}

discover_loop_entrypoints() {
  local file
  {
    for file in "$REPO_ROOT"/scripts/*auto-loop*.sh; do
      [[ -e "$file" ]] || continue
      case "${file#"$REPO_ROOT/"}" in
        scripts/auto-loop-safety.sh) continue ;;
      esac
      printf '%s\n' "${file#"$REPO_ROOT/"}"
    done
    for file in "$REPO_ROOT"/scripts/*.sh; do
      if grep -Eq '(\.planning/auto-loop|/pm-auto-loop|claude-orchestrator)' "$file" && \
         grep -Eq -- '(--resume|RESUME)' "$file"; then
        printf '%s\n' "${file#"$REPO_ROOT/"}"
      fi
    done
  } | sort -u
}

tree_snapshot() {
  local root="$1"
  local path
  if [[ ! -e "$root" ]]; then
    printf '<absent>\n'
    return
  fi
  find "$root" -mindepth 1 -print | sort
  while IFS= read -r path; do
    if stat -f '%Sp %z %m %N' "$path" >/dev/null 2>&1; then
      stat -f '%Sp %z %m %N' "$path"
    else
      stat -c '%A %s %Y %n' "$path"
    fi
  done < <(find "$root" -mindepth 1 -print | sort)
}

prepare_repo() {
  local root="$1"
  mkdir -p "$root/scripts" "$root/.agents/agentic-delivery/prompts"
  cp "$REPO_ROOT/scripts/claude-auto-loop.sh" "$root/scripts/claude-auto-loop.sh"
  cp "$REPO_ROOT/scripts/pi-auto-loop.sh" "$root/scripts/pi-auto-loop.sh"
  cp "$REPO_ROOT/scripts/pi-shepherd-loop.sh" "$root/scripts/pi-shepherd-loop.sh"
  if [[ -f "$REPO_ROOT/scripts/auto-loop-safety.sh" ]]; then
    cp "$REPO_ROOT/scripts/auto-loop-safety.sh" "$root/scripts/auto-loop-safety.sh"
  fi
  printf 'synthetic orchestrator fixture\n' >"$root/.agents/agentic-delivery/prompts/claude-orchestrator.md"
  printf 'synthetic validator fixture\n' >"$root/.agents/agentic-delivery/prompts/shepherd-validator-prompt.md"
}

prepare_tool_sandbox() {
  local root="$1"
  local tool
  local stub="$root/stub-tool"
  mkdir -p "$root/bin" "$root/home" "$root/config/git" "$root/config/gh"
  printf '#!/usr/bin/env bash\nprintf "%s\\n" "${0##*/}" >>"$TOOL_LOG"\nexit 99\n' '%s' >"$stub"
  chmod +x "$stub"
  for tool in claude pi git gh curl wget ssh scp rsync nc openssl python3 podman docker make go cat mkdir tee date sleep dirname; do
    ln -s "$stub" "$root/bin/$tool"
  done
  ln -s /bin/bash "$root/bin/bash"
}

run_driver() {
  local driver="$1"
  local mode="$2"
  local root
  local argument
  local -a clean_env
  local before
  local after
  local rc
  root="$(mktemp -d "$TEST_TMP/driver.XXXXXX")"
  prepare_repo "$root"
  prepare_tool_sandbox "$root"
  : >"$root/tool-log"

  if [[ "$mode" == "resume" ]]; then
    mkdir -p "$root/.planning/auto-loop"
    printf 'synthetic resume fixture' >"$root/.planning/auto-loop/PROMPT.txt"
    chmod 000 "$root/.planning/auto-loop/PROMPT.txt"
    chmod 555 "$root/.planning/auto-loop" "$root/.planning"
    argument="--resume"
  elif [[ "$mode" == "help" ]]; then
    mkdir -p "$root/.planning"
    chmod 555 "$root/.planning"
    argument="--help"
  elif [[ "$mode" == "flag-enable" ]]; then
    mkdir -p "$root/.planning"
    chmod 555 "$root/.planning"
    argument="--enable"
  elif [[ "$mode" == "flag-force" ]]; then
    mkdir -p "$root/.planning"
    chmod 555 "$root/.planning"
    argument="--force"
  else
    mkdir -p "$root/.planning"
    chmod 555 "$root/.planning"
    argument="synthetic task"
  fi
  before="$(tree_snapshot "$root/.planning")"

  clean_env=(
    "PATH=$root/bin"
    "HOME=$root/home"
    "XDG_CONFIG_HOME=$root/config"
    "GIT_CONFIG_GLOBAL=$root/config/git/config"
    "GH_CONFIG_DIR=$root/config/gh"
    "TOOL_LOG=$root/tool-log"
  )
  if [[ "$mode" == "env-enable" ]]; then
    clean_env+=(
      "POLYMETRICS_AGENT_LOOP_ENABLE=1"
      "AUTO_LOOP_ENABLE=1"
      "ENABLE_AUTO_LOOP=1"
    )
  fi

  if [[ "$driver" == "claude" ]]; then
    /usr/bin/env -i "${clean_env[@]}" \
      CLAUDE_BIN="$root/bin/claude" MAX_ITERATIONS=1 MAX_NO_VERDICT=1 COOLDOWN_SECONDS=0 \
      "$root/scripts/claude-auto-loop.sh" "$argument" >"$root/stdout" 2>"$root/stderr"
    rc=$?
  elif [[ "$driver" == "pi" ]]; then
    /usr/bin/env -i "${clean_env[@]}" \
      PI_BIN="$root/bin/pi" MAX_ITERATIONS=1 CONTINUE_SESSION=0 COOLDOWN_SECONDS=0 \
      "$root/scripts/pi-auto-loop.sh" "$argument" >"$root/stdout" 2>"$root/stderr"
    rc=$?
  else
    /usr/bin/env -i "${clean_env[@]}" \
      PI_BIN="$root/bin/pi" VALIDATOR_BIN="$root/bin/pi" MAX_ITERATIONS=1 \
      MAX_NO_VERDICT=1 COOLDOWN_SECONDS=0 \
      "$root/scripts/pi-shepherd-loop.sh" "$argument" >"$root/stdout" 2>"$root/stderr"
    rc=$?
  fi
  after="$(tree_snapshot "$root/.planning")"

  if [[ "$mode" == "help" ]]; then
    if [[ "$rc" -ne 0 ]]; then
      fail "$driver help exit=$rc, want 0"
    fi
    if ! grep -qi 'usage:' "$root/stdout"; then
      fail "$driver help did not print usage to stdout"
    fi
  else
    if [[ "$rc" -ne 78 ]]; then
      fail "$driver $mode exit=$rc, want 78"
    fi
    if ! grep -q 'AUTO_LOOP_DISABLED_PHASE_0' "$root/stderr"; then
      fail "$driver $mode did not emit the stable safety code"
    fi
  fi

  if [[ -s "$root/tool-log" ]]; then
    fail "$driver $mode invoked a process before the guard: $(tr '\n' ',' <"$root/tool-log")"
  fi
  if [[ "$before" != "$after" ]]; then
    fail "$driver $mode read or persisted loop state before the guard"
  fi

  if [[ -d "$root/.planning/auto-loop" ]]; then
    chmod 700 "$root/.planning/auto-loop"
  fi
  if [[ -e "$root/.planning/auto-loop/PROMPT.txt" ]]; then
    chmod 600 "$root/.planning/auto-loop/PROMPT.txt"
  fi
  if [[ -d "$root/.planning" ]]; then
    chmod 700 "$root/.planning"
  fi
  rm -rf "$root"
}

candidate_entrypoints="$(discover_loop_entrypoints)"
if [[ -z "$candidate_entrypoints" ]]; then
  fail "independent discovery found no autonomous-loop entrypoints"
fi

if [[ ! -x "$REPO_ROOT/scripts/auto-loop-safety.sh" ]]; then
  fail "scripts/auto-loop-safety.sh is missing or not executable"
else
  registered_entrypoints="$($REPO_ROOT/scripts/auto-loop-safety.sh entrypoints)"
  if [[ "$candidate_entrypoints" != "$registered_entrypoints" ]]; then
    fail "independently discovered entrypoints differ from the canonical inventory"
  fi

  while IFS= read -r entrypoint; do
    [[ -n "$entrypoint" ]] || continue
    if ! grep -Fxq "$entrypoint" <<<"$registered_entrypoints"; then
      fail "$entrypoint is not registered"
    fi
    if ! grep -q 'auto-loop-safety.sh' "$REPO_ROOT/$entrypoint"; then
      fail "$entrypoint does not source the safety policy"
    fi
    if ! grep -q "auto_loop_safety_guard_driver \"$entrypoint\"" "$REPO_ROOT/$entrypoint"; then
      fail "$entrypoint does not guard its own canonical identity"
    fi
  done <<<"$candidate_entrypoints"

  status_output="$($REPO_ROOT/scripts/auto-loop-safety.sh status --json)"
  if ! grep -q '"state":"closed"' <<<"$status_output"; then
    fail "shell safety status is not closed"
  fi

  "$REPO_ROOT/scripts/auto-loop-safety.sh" guard-driver scripts/untracked-auto-loop.sh --json \
    >"$TEST_TMP/helper-stdout" 2>"$TEST_TMP/helper-stderr"
  untracked_rc=$?
  if [[ "$untracked_rc" -ne 64 ]] || ! grep -q 'AUTO_LOOP_ENTRYPOINT_UNTRACKED' "$TEST_TMP/helper-stderr"; then
    fail "untracked shell guard did not fail closed with exit 64"
  fi

  for forbidden in enable open run resume; do
    "$REPO_ROOT/scripts/auto-loop-safety.sh" "$forbidden" >/dev/null 2>&1
    forbidden_rc=$?
    if [[ "$forbidden_rc" -ne 64 ]]; then
      fail "shell safety command $forbidden exit=$forbidden_rc, want 64"
    fi
  done

  go_status="$(cd "$REPO_ROOT" && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/loopctl safety status --json 2>/dev/null)"
  if [[ "$go_status" != "$status_output" ]]; then
    fail "Go and shell safety status outputs differ"
  fi
  go_entrypoints="$(cd "$REPO_ROOT" && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/loopctl safety entrypoints 2>/dev/null)"
  if [[ "$go_entrypoints" != "$registered_entrypoints" ]]; then
    fail "Go and shell entrypoint inventories differ"
  fi
fi

for driver in claude pi shepherd; do
  run_driver "$driver" run
  run_driver "$driver" resume
  run_driver "$driver" help
  run_driver "$driver" env-enable
  run_driver "$driver" flag-enable
  run_driver "$driver" flag-force
done

if ! grep -Fq 'ORCH_MODEL="${ORCH_MODEL:-openai-codex/gpt-5.5}"' \
  "$REPO_ROOT/scripts/pi-shepherd-loop.sh"; then
  fail "Shepherd changed the orchestrator model; only the validator may move to GPT-5.6"
fi
if ! grep -Fq 'VALIDATOR_MODEL="openai-codex/gpt-5.6-sol"' \
  "$REPO_ROOT/scripts/pi-shepherd-loop.sh" || \
  ! grep -Fq 'VALIDATOR_ARGS="--model $VALIDATOR_MODEL --thinking high ' \
  "$REPO_ROOT/scripts/pi-shepherd-loop.sh"; then
  fail "Shepherd validator default is not GPT-5.6 Sol with high reasoning"
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
