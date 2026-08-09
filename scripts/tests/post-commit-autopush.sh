#!/usr/bin/env sh
# Exercise the post-commit auto-push hook against local bare repositories only.
# In particular, do not turn this into a static grep test: the non-fast-forward
# case below proves Git rejects the exact attempted push and that the hook
# swallows it without a force option.
set -eu

ROOT_DIR="$(CDPATH='' cd "$(dirname "$0")/../.." && pwd)"
HOOK="$ROOT_DIR/.githooks/post-commit"
TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/pm-autopush.XXXXXX")"
CURRENT_HOOK_LOG=""

cleanup() {
  rm -rf "$TMP_DIR"
}

trap cleanup 0
trap 'exit 1' 1 2 15

fail() {
  printf 'post-commit autopush test failed: %s\n' "$1" >&2
  exit 1
}

[ -x "$HOOK" ] || fail "post-commit hook is missing or not executable"

export PM_TEST_HOOK_SOURCE="$HOOK"

new_repo() {
  case_dir=$1
  TEST_REMOTE="$case_dir/remote.git"
  TEST_REPO="$case_dir/repo"

  git init --bare -q -b main "$TEST_REMOTE"
  git clone -q "$TEST_REMOTE" "$TEST_REPO" 2>/dev/null
  git -C "$TEST_REPO" config user.name "Auto Push Test"
  git -C "$TEST_REPO" config user.email "autopush-test@example.invalid"
  printf 'base\n' > "$TEST_REPO/base"
  git -C "$TEST_REPO" add base
  git -C "$TEST_REPO" commit -qm "test: base"
  git -C "$TEST_REPO" push -q origin main
  git -C "$TEST_REPO" remote set-head origin -a >/dev/null 2>&1 || :
}

install_hook() {
  hook_repo=$1
  hook_dir="$hook_repo/test-hooks"

  mkdir -p "$hook_dir"
  # shellcheck disable=SC2016 # The wrapper must defer these variables until Git invokes the hook.
  printf '%s\n' \
    '#!/bin/sh' \
    'if [ -n "${PM_TEST_HOOK_LOG:-}" ]; then' \
    '  printf "%s\n" post-commit >> "$PM_TEST_HOOK_LOG"' \
    'fi' \
    'exec "$PM_TEST_HOOK_SOURCE"' > "$hook_dir/post-commit"
  chmod +x "$hook_dir/post-commit"
  git -C "$hook_repo" config core.hooksPath "$hook_dir"
}

record_change() {
  change_repo=$1
  change_text=$2

  printf '%s\n' "$change_text" >> "$change_repo/changes"
  git -C "$change_repo" add changes
}

run_hooked() {
  PM_TEST_HOOK_LOG="$CURRENT_HOOK_LOG" "$@"
}

state_file() {
  state_common_dir="$(git -C "$1" rev-parse --path-format=absolute --git-common-dir)"
  GIT_DIR="$state_common_dir" git rev-parse --git-path "pm-autopush/$2.last-push"
}

status_file() {
  status_common_dir="$(git -C "$1" rev-parse --path-format=absolute --git-common-dir)"
  GIT_DIR="$status_common_dir" git rev-parse --git-path "pm-autopush/$2.log"
}

install_receive_delay() {
  delay_remote=$1

  printf '%s\n' '#!/bin/sh' 'sleep 3' > "$delay_remote/hooks/pre-receive"
  chmod +x "$delay_remote/hooks/pre-receive"
}

wait_for_ref() {
  wait_remote=$1
  wait_ref=$2
  wait_expected=$3
  wait_attempt=0

  while [ "$wait_attempt" -lt 10 ]; do
    wait_actual="$(git --git-dir="$wait_remote" rev-parse --verify -q "$wait_ref" 2>/dev/null || :)"
    [ "$wait_actual" = "$wait_expected" ] && return 0
    sleep 1
    wait_attempt=$((wait_attempt + 1))
  done

  fail "timed out waiting for $wait_ref to reach $wait_expected"
}

wait_for_text() {
  wait_file=$1
  wait_text=$2
  wait_attempt=0

  while [ "$wait_attempt" -lt 10 ]; do
    if [ -f "$wait_file" ] && grep -Fq "$wait_text" "$wait_file"; then
      return 0
    fi
    sleep 1
    wait_attempt=$((wait_attempt + 1))
  done

  fail "timed out waiting for log text: $wait_text"
}

assert_hook_invoked() {
  [ -s "$1" ] || fail "Git did not invoke the post-commit hook"
}

test_default_branch_refusal() {
  case_dir="$TMP_DIR/default"
  mkdir -p "$case_dir"
  new_repo "$case_dir"
  install_hook "$TEST_REPO"
  CURRENT_HOOK_LOG="$case_dir/hook.log"
  : > "$CURRENT_HOOK_LOG"

  record_change "$TEST_REPO" default
  if ! run_hooked git -C "$TEST_REPO" commit -qm "test: default branch refusal"; then
    fail "default-branch commit should succeed"
  fi
  assert_hook_invoked "$CURRENT_HOOK_LOG"
  sleep 1
  local_sha="$(git -C "$TEST_REPO" rev-parse HEAD)"
  remote_sha="$(git --git-dir="$TEST_REMOTE" rev-parse --verify refs/heads/main)"
  [ "$local_sha" != "$remote_sha" ] || fail "default branch was pushed"
}

test_detached_head_refusal() {
  case_dir="$TMP_DIR/detached"
  mkdir -p "$case_dir"
  new_repo "$case_dir"
  install_hook "$TEST_REPO"
  CURRENT_HOOK_LOG="$case_dir/hook.log"
  : > "$CURRENT_HOOK_LOG"

  git -C "$TEST_REPO" checkout -q --detach
  record_change "$TEST_REPO" detached
  if ! run_hooked git -C "$TEST_REPO" commit -qm "test: detached head refusal"; then
    fail "detached-head commit should succeed"
  fi
  assert_hook_invoked "$CURRENT_HOOK_LOG"
  sleep 1
  local_sha="$(git -C "$TEST_REPO" rev-parse HEAD)"
  remote_sha="$(git --git-dir="$TEST_REMOTE" rev-parse --verify refs/heads/main)"
  [ "$local_sha" != "$remote_sha" ] || fail "detached HEAD was pushed"
}

test_opt_out() {
  case_dir="$TMP_DIR/opt-out"
  mkdir -p "$case_dir"
  new_repo "$case_dir"
  git -C "$TEST_REPO" checkout -qb feature
  install_hook "$TEST_REPO"
  CURRENT_HOOK_LOG="$case_dir/hook.log"
  : > "$CURRENT_HOOK_LOG"

  record_change "$TEST_REPO" opt-out
  if ! PM_NO_AUTOPUSH=1 \
    PM_TEST_HOOK_LOG="$CURRENT_HOOK_LOG" \
      git -C "$TEST_REPO" commit -qm "test: opt out"; then
    fail "opt-out commit should succeed"
  fi
  assert_hook_invoked "$CURRENT_HOOK_LOG"
  sleep 1
  if git --git-dir="$TEST_REMOTE" rev-parse --verify -q refs/heads/feature >/dev/null; then
    fail "opt-out commit reached the remote"
  fi
}

test_rate_limit_and_detached_push() {
  case_dir="$TMP_DIR/rate-limit"
  linked_dir="$case_dir/linked"
  mkdir -p "$case_dir"
  new_repo "$case_dir"
  git -C "$TEST_REPO" checkout -qb feature
  install_hook "$TEST_REPO"
  CURRENT_HOOK_LOG="$case_dir/hook.log"
  : > "$CURRENT_HOOK_LOG"

  git -C "$TEST_REPO" worktree add -q --detach "$linked_dir" feature
  state_from_primary="$(state_file "$TEST_REPO" feature)"
  state_from_linked="$(state_file "$linked_dir" feature)"
  [ "$state_from_primary" = "$state_from_linked" ] || fail "timestamp path is not shared by linked worktrees"

  install_receive_delay "$TEST_REMOTE"
  record_change "$TEST_REPO" first
  started="$(date +%s)"
  if ! run_hooked git -C "$TEST_REPO" commit -qm "test: first auto push"; then
    fail "feature commit should succeed"
  fi
  assert_hook_invoked "$CURRENT_HOOK_LOG"
  finished="$(date +%s)"
  [ $((finished - started)) -lt 2 ] || fail "detached push delayed the commit"

  first_sha="$(git -C "$TEST_REPO" rev-parse HEAD)"
  wait_for_ref "$TEST_REMOTE" refs/heads/feature "$first_sha"

  record_change "$TEST_REPO" second
  if ! run_hooked git -C "$TEST_REPO" commit -qm "test: rate limited commit"; then
    fail "rate-limited commit should succeed"
  fi
  second_sha="$(git -C "$TEST_REPO" rev-parse HEAD)"
  sleep 4
  remote_sha="$(git --git-dir="$TEST_REMOTE" rev-parse --verify refs/heads/feature)"
  [ "$remote_sha" = "$first_sha" ] || fail "inside-window commit reached the remote"

  expired=$(( $(date +%s) - 601 ))
  printf '%s\n' "$expired" > "$state_from_primary"
  record_change "$TEST_REPO" third
  if ! run_hooked git -C "$TEST_REPO" commit -qm "test: expired rate window"; then
    fail "post-window commit should succeed"
  fi
  third_sha="$(git -C "$TEST_REPO" rev-parse HEAD)"
  wait_for_ref "$TEST_REMOTE" refs/heads/feature "$third_sha"
  git -C "$TEST_REPO" merge-base --is-ancestor "$second_sha" "$third_sha" || fail "catch-up push omitted the rate-limited commit"
}

test_real_multicommit_rebase_refusal() {
  case_dir="$TMP_DIR/rebase"
  mkdir -p "$case_dir"
  new_repo "$case_dir"
  git -C "$TEST_REPO" checkout -qb feature
  record_change "$TEST_REPO" feature-one
  git -C "$TEST_REPO" commit -qm "test: feature one"
  record_change "$TEST_REPO" feature-two
  git -C "$TEST_REPO" commit -qm "test: feature two"
  old_tip="$(git -C "$TEST_REPO" rev-parse HEAD)"

  git -C "$TEST_REPO" checkout -q main
  printf 'main-advance\n' > "$TEST_REPO/main-advance"
  git -C "$TEST_REPO" add main-advance
  git -C "$TEST_REPO" commit -qm "test: main advance"
  git -C "$TEST_REPO" push -q origin main
  git -C "$TEST_REPO" checkout -q feature

  install_hook "$TEST_REPO"
  CURRENT_HOOK_LOG="$case_dir/hook.log"
  : > "$CURRENT_HOOK_LOG"

  if ! run_hooked git -C "$TEST_REPO" rebase main; then
    fail "real multi-commit rebase should succeed"
  fi
  new_tip="$(git -C "$TEST_REPO" rev-parse HEAD)"
  replayed_count="$(git -C "$TEST_REPO" rev-list --count main..feature)"
  hook_count="$(wc -l < "$CURRENT_HOOK_LOG" | tr -d ' ')"
  [ "$old_tip" != "$new_tip" ] || fail "rebase did not rewrite the feature commits"
  [ "$replayed_count" -eq 2 ] || fail "rebase did not replay two commits"
  [ "$hook_count" -ge 2 ] || fail "post-commit hook was not invoked for each replay"
  sleep 1
  if git --git-dir="$TEST_REMOTE" rev-parse --verify -q refs/heads/feature >/dev/null; then
    fail "rebase replay pushed the feature branch"
  fi
}

test_rejected_push_is_swallowed_without_force() {
  case_dir="$TMP_DIR/rejected"
  upstream_repo="$case_dir/upstream"
  mkdir -p "$case_dir"
  new_repo "$case_dir"
  git -C "$TEST_REPO" checkout -qb feature
  git -C "$TEST_REPO" push -q -u origin feature
  install_hook "$TEST_REPO"

  git clone -q "$TEST_REMOTE" "$upstream_repo"
  git -C "$upstream_repo" config user.name "Upstream Test"
  git -C "$upstream_repo" config user.email "upstream-test@example.invalid"
  git -C "$upstream_repo" checkout -qb feature origin/feature
  record_change "$upstream_repo" remote-wins
  git -C "$upstream_repo" commit -qm "test: remote advance"
  git -C "$upstream_repo" push -q origin feature
  upstream_sha="$(git -C "$upstream_repo" rev-parse HEAD)"

  CURRENT_HOOK_LOG="$case_dir/hook.log"
  : > "$CURRENT_HOOK_LOG"
  record_change "$TEST_REPO" local-loses
  if ! run_hooked git -C "$TEST_REPO" commit -qm "test: rejected push"; then
    fail "commit must remain successful when its asynchronous push is rejected"
  fi
  local_sha="$(git -C "$TEST_REPO" rev-parse HEAD)"
  rejection_log="$(status_file "$TEST_REPO" feature)"
  wait_for_text "$rejection_log" "pm autopush: push failed for feature (ignored)"
  remote_sha="$(git --git-dir="$TEST_REMOTE" rev-parse --verify refs/heads/feature)"

  [ "$remote_sha" = "$upstream_sha" ] || fail "rejected push changed the diverged remote"
  [ "$remote_sha" != "$local_sha" ] || fail "diverged local commit unexpectedly reached the remote"
  rejection_lines="$(wc -l < "$rejection_log" | tr -d ' ')"
  [ "$rejection_lines" -eq 1 ] || fail "rejected push did not write exactly one short log line"
}

test_default_branch_refusal
test_detached_head_refusal
test_opt_out
test_rate_limit_and_detached_push
test_real_multicommit_rebase_refusal
test_rejected_push_is_swallowed_without_force

printf 'post-commit autopush: all executable refusal and safety checks passed\n'
