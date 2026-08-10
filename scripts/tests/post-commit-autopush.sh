#!/usr/bin/env sh
# Exercise the post-commit auto-push hook against local bare repositories only.
# In particular, do not turn this into a static grep test: the non-fast-forward
# case below proves Git rejects the exact attempted push and that the hook
# swallows it without a force option.
set -eu

ROOT_DIR="$(CDPATH='' cd "$(dirname "$0")/../.." && pwd)"
HOOK="$ROOT_DIR/.githooks/post-commit"
PREPARE_HOOK="$ROOT_DIR/.githooks/prepare-commit-msg"
TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/pm-autopush.XXXXXX")"
CURRENT_HOOK_LOG=""
CURRENT_OPERATION_LOG=""

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
[ -x "$PREPARE_HOOK" ] || fail "prepare-commit-msg hook is missing or not executable"

export PM_TEST_HOOK_SOURCE="$HOOK"
export PM_TEST_PREPARE_HOOK_SOURCE="$PREPARE_HOOK"

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
    'if [ -n "${PM_TEST_OPERATION_LOG:-}" ] && [ -f "$(git rev-parse --path-format=absolute --git-dir)/pm-autopush-operation" ]; then' \
    '  printf "%s\n" operation-snapshot >> "$PM_TEST_OPERATION_LOG"' \
    'fi' \
    'exec "$PM_TEST_HOOK_SOURCE"' > "$hook_dir/post-commit"
  printf '%s\n' \
    '#!/bin/sh' \
    "exec \"\$PM_TEST_PREPARE_HOOK_SOURCE\" \"\$@\"" > "$hook_dir/prepare-commit-msg"
  chmod +x "$hook_dir/post-commit" "$hook_dir/prepare-commit-msg"
  git -C "$hook_repo" config core.hooksPath "$hook_dir"
}

record_change() {
  change_repo=$1
  change_text=$2

  printf '%s\n' "$change_text" >> "$change_repo/changes"
  git -C "$change_repo" add changes
}

run_hooked() {
  PM_TEST_HOOK_LOG="$CURRENT_HOOK_LOG" \
    PM_TEST_OPERATION_LOG="$CURRENT_OPERATION_LOG" \
    "$@"
}

state_file() {
  state_common_dir="$(git -C "$1" rev-parse --path-format=absolute --git-common-dir)"
  GIT_DIR="$state_common_dir" git rev-parse --git-path "pm-autopush/$2.last-push"
}

status_file() {
  status_common_dir="$(git -C "$1" rev-parse --path-format=absolute --git-common-dir)"
  GIT_DIR="$status_common_dir" git rev-parse --git-path "pm-autopush/$2.log"
}

operation_file() {
  operation_git_dir="$(git -C "$1" rev-parse --path-format=absolute --git-dir)"
  printf '%s/pm-autopush-operation\n' "$operation_git_dir"
}

install_receive_delay() {
  delay_remote=$1

  printf '%s\n' '#!/bin/sh' 'sleep 3' > "$delay_remote/hooks/pre-receive"
  chmod +x "$delay_remote/hooks/pre-receive"
}

install_receive_count_delay() {
  count_remote=$1

  printf '%s\n' \
    '#!/bin/sh' \
    'printf "%s\n" receive >> pm-autopush-receives' \
    'sleep 3' > "$count_remote/hooks/pre-receive"
  chmod +x "$count_remote/hooks/pre-receive"
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

assert_manual_operation_refusal() {
  operation_repo=$1
  operation_remote=$2
  operation_branch=$3
  operation_state="$(state_file "$operation_repo" "$operation_branch")"
  operation_snapshot="$(operation_file "$operation_repo")"

  [ ! -f "$operation_state" ] || fail "manual operation recorded an auto-push attempt"
  [ ! -f "$operation_snapshot" ] || fail "manual operation snapshot was not consumed"
  if git --git-dir="$operation_remote" rev-parse --verify -q "refs/heads/$operation_branch" >/dev/null; then
    fail "manual operation reached the remote"
  fi
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

test_stale_remote_default_refusal() {
  case_dir="$TMP_DIR/stale-default"
  mkdir -p "$case_dir"
  new_repo "$case_dir"
  git -C "$TEST_REPO" checkout -qb trunk
  git -C "$TEST_REPO" push -q -u origin trunk
  git -C "$TEST_REPO" checkout -qb release main
  git -C "$TEST_REPO" push -q -u origin release
  git --git-dir="$TEST_REMOTE" symbolic-ref HEAD refs/heads/trunk
  git -C "$TEST_REPO" remote set-head origin -a >/dev/null 2>&1 || fail "could not set local remote default"
  git --git-dir="$TEST_REMOTE" symbolic-ref HEAD refs/heads/release
  install_hook "$TEST_REPO"
  CURRENT_HOOK_LOG="$case_dir/hook.log"
  : > "$CURRENT_HOOK_LOG"

  before_sha="$(git --git-dir="$TEST_REMOTE" rev-parse refs/heads/release)"
  record_change "$TEST_REPO" stale-default
  if ! run_hooked git -C "$TEST_REPO" commit -qm "test: stale remote default refusal"; then
    fail "stale-default commit should succeed"
  fi
  assert_hook_invoked "$CURRENT_HOOK_LOG"
  default_log="$(status_file "$TEST_REPO" release)"
  wait_for_text "$default_log" "pm autopush: default branch skipped for release (ignored)"
  after_sha="$(git --git-dir="$TEST_REMOTE" rev-parse refs/heads/release)"
  [ "$before_sha" = "$after_sha" ] || fail "stale remote default branch was pushed"
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

test_concurrent_linked_worktree_lease() {
  case_dir="$TMP_DIR/concurrent-lease"
  linked_dir="$case_dir/linked"
  wrapper_dir="$case_dir/wrappers"
  mkdir -p "$case_dir" "$wrapper_dir"
  new_repo "$case_dir"
  git -C "$TEST_REPO" checkout -qb feature
  git -C "$TEST_REPO" worktree add -q --force "$linked_dir" feature

  state_from_primary="$(state_file "$TEST_REPO" feature)"
  state_from_linked="$(state_file "$linked_dir" feature)"
  [ "$state_from_primary" = "$state_from_linked" ] || fail "concurrent linked worktrees do not share state"

  real_nohup="$(command -v nohup)"
  schedule_log="$case_dir/schedules.log"
  printf '%s\n' \
    '#!/bin/sh' \
    "printf \"%s\\n\" scheduled >> \"\$PM_AUTOPUSH_SCHEDULE_LOG\"" \
    "exec \"\$PM_AUTOPUSH_REAL_NOHUP\" \"\$@\"" > "$wrapper_dir/nohup"
  chmod +x "$wrapper_dir/nohup"
  install_receive_count_delay "$TEST_REMOTE"

  (
    cd "$TEST_REPO"
    PATH="$wrapper_dir:$PATH" \
      PM_AUTOPUSH_SCHEDULE_LOG="$schedule_log" \
      PM_AUTOPUSH_REAL_NOHUP="$real_nohup" \
      "$HOOK"
  ) &
  primary_pid=$!
  (
    cd "$linked_dir"
    PATH="$wrapper_dir:$PATH" \
      PM_AUTOPUSH_SCHEDULE_LOG="$schedule_log" \
      PM_AUTOPUSH_REAL_NOHUP="$real_nohup" \
      "$HOOK"
  ) &
  linked_pid=$!
  if ! wait "$primary_pid"; then
    fail "primary concurrent hook should succeed"
  fi
  if ! wait "$linked_pid"; then
    fail "linked concurrent hook should succeed"
  fi

  expected_sha="$(git -C "$TEST_REPO" rev-parse HEAD)"
  wait_for_ref "$TEST_REMOTE" refs/heads/feature "$expected_sha"
  sleep 4
  [ -f "$schedule_log" ] || fail "concurrent hooks did not schedule a push"
  schedule_count="$(wc -l < "$schedule_log" | tr -d ' ')"
  [ "$schedule_count" -eq 1 ] || fail "concurrent hooks scheduled more than one push"
  receive_log="$TEST_REMOTE/pm-autopush-receives"
  [ -f "$receive_log" ] || fail "concurrent hooks did not reach the receiver"
  receive_count="$(wc -l < "$receive_log" | tr -d ' ')"
  [ "$receive_count" -eq 1 ] || fail "concurrent hooks reached the receiver more than once"
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

test_manual_merge_completion_refusal() {
  case_dir="$TMP_DIR/manual-merge"
  mkdir -p "$case_dir"
  new_repo "$case_dir"
  git -C "$TEST_REPO" checkout -qb feature
  git -C "$TEST_REPO" checkout -qb merge-source
  record_change "$TEST_REPO" merge-source
  git -C "$TEST_REPO" commit -qm "test: merge source"
  git -C "$TEST_REPO" checkout -q feature
  install_hook "$TEST_REPO"
  CURRENT_HOOK_LOG="$case_dir/hook.log"
  CURRENT_OPERATION_LOG="$case_dir/operation.log"
  : > "$CURRENT_HOOK_LOG"
  : > "$CURRENT_OPERATION_LOG"

  if ! run_hooked git -C "$TEST_REPO" merge --no-ff --no-commit merge-source; then
    fail "manual merge setup should succeed"
  fi
  merge_head="$(git -C "$TEST_REPO" rev-parse --git-path MERGE_HEAD)"
  [ -f "$merge_head" ] || fail "manual merge did not retain MERGE_HEAD"
  if ! run_hooked git -C "$TEST_REPO" commit -qm "test: manual merge completion"; then
    fail "manual merge completion should succeed"
  fi
  assert_hook_invoked "$CURRENT_HOOK_LOG"
  [ -s "$CURRENT_OPERATION_LOG" ] || fail "manual merge did not capture an operation snapshot"
  assert_manual_operation_refusal "$TEST_REPO" "$TEST_REMOTE" feature
}

test_manual_cherry_pick_completion_refusal() {
  case_dir="$TMP_DIR/manual-cherry-pick"
  mkdir -p "$case_dir"
  new_repo "$case_dir"
  git -C "$TEST_REPO" checkout -qb feature
  git -C "$TEST_REPO" checkout -qb cherry-source
  printf 'cherry source\n' > "$TEST_REPO/base"
  git -C "$TEST_REPO" add base
  git -C "$TEST_REPO" commit -qm "test: cherry source"
  cherry_source="$(git -C "$TEST_REPO" rev-parse HEAD)"
  git -C "$TEST_REPO" checkout -q feature
  printf 'feature change\n' > "$TEST_REPO/base"
  git -C "$TEST_REPO" add base
  git -C "$TEST_REPO" commit -qm "test: feature conflict"
  install_hook "$TEST_REPO"
  CURRENT_HOOK_LOG="$case_dir/hook.log"
  CURRENT_OPERATION_LOG="$case_dir/operation.log"
  : > "$CURRENT_HOOK_LOG"
  : > "$CURRENT_OPERATION_LOG"

  if run_hooked git -C "$TEST_REPO" cherry-pick "$cherry_source"; then
    fail "cherry-pick should stop for manual conflict resolution"
  fi
  cherry_head="$(git -C "$TEST_REPO" rev-parse --git-path CHERRY_PICK_HEAD)"
  [ -f "$cherry_head" ] || fail "manual cherry-pick did not retain CHERRY_PICK_HEAD"
  printf 'resolved cherry-pick\n' > "$TEST_REPO/base"
  git -C "$TEST_REPO" add base
  if ! GIT_EDITOR=: run_hooked git -C "$TEST_REPO" cherry-pick --continue; then
    fail "manual cherry-pick completion should succeed"
  fi
  assert_hook_invoked "$CURRENT_HOOK_LOG"
  [ -s "$CURRENT_OPERATION_LOG" ] || fail "manual cherry-pick did not capture an operation snapshot"
  assert_manual_operation_refusal "$TEST_REPO" "$TEST_REMOTE" feature
}

test_manual_revert_completion_refusal() {
  case_dir="$TMP_DIR/manual-revert"
  mkdir -p "$case_dir"
  new_repo "$case_dir"
  git -C "$TEST_REPO" checkout -qb feature
  printf 'revert target\n' > "$TEST_REPO/base"
  git -C "$TEST_REPO" add base
  git -C "$TEST_REPO" commit -qm "test: revert target"
  revert_target="$(git -C "$TEST_REPO" rev-parse HEAD)"
  printf 'later change\n' > "$TEST_REPO/base"
  git -C "$TEST_REPO" add base
  git -C "$TEST_REPO" commit -qm "test: conflicting later change"
  install_hook "$TEST_REPO"
  CURRENT_HOOK_LOG="$case_dir/hook.log"
  CURRENT_OPERATION_LOG="$case_dir/operation.log"
  : > "$CURRENT_HOOK_LOG"
  : > "$CURRENT_OPERATION_LOG"

  if run_hooked git -C "$TEST_REPO" revert "$revert_target"; then
    fail "revert should stop for manual conflict resolution"
  fi
  revert_head="$(git -C "$TEST_REPO" rev-parse --git-path REVERT_HEAD)"
  [ -f "$revert_head" ] || fail "manual revert did not retain REVERT_HEAD"
  printf 'resolved revert\n' > "$TEST_REPO/base"
  git -C "$TEST_REPO" add base
  if ! GIT_EDITOR=: run_hooked git -C "$TEST_REPO" revert --continue; then
    fail "manual revert completion should succeed"
  fi
  assert_hook_invoked "$CURRENT_HOOK_LOG"
  [ -s "$CURRENT_OPERATION_LOG" ] || fail "manual revert did not capture an operation snapshot"
  assert_manual_operation_refusal "$TEST_REPO" "$TEST_REMOTE" feature
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
test_stale_remote_default_refusal
test_detached_head_refusal
test_opt_out
test_rate_limit_and_detached_push
test_concurrent_linked_worktree_lease
test_real_multicommit_rebase_refusal
test_manual_merge_completion_refusal
test_manual_cherry_pick_completion_refusal
test_manual_revert_completion_refusal
test_rejected_push_is_swallowed_without_force

printf 'post-commit autopush: all executable refusal and safety checks passed\n'
