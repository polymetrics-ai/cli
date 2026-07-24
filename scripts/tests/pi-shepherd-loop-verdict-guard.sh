#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/pi-shepherd-verdict-guard.XXXXXX")"
trap '[[ "${KEEP_SHEPHERD_TEST_TMP:-0}" == "1" ]] || rm -rf "$tmp_dir"' EXIT

AUTO_LOOP_STATE_DIR="$tmp_dir/state" \
SHEPHERD_VERDICT_GUARD_SELF_TEST=1 \
"$repo_root/scripts/pi-shepherd-loop.sh"

workflow="$repo_root/.agents/agentic-delivery/workflows/shepherd-validator.md"
validator_prompt="$repo_root/.agents/agentic-delivery/prompts/shepherd-validator-prompt.md"
python3 - "$workflow" "$validator_prompt" <<'PY'
import pathlib
import sys

policies = (
    "dependency",
    "auth-scope",
    "destructive/admin",
    "production deploy",
    "credentialed connector",
    "reverse-ETL",
    "generic shell/HTTP/SQL write",
    "quality-gate reduction",
    "parent-readiness mutation",
    "merge to `main`",
)
for path in map(pathlib.Path, sys.argv[1:]):
    normalized = " ".join(path.read_text().split())
    for policy in policies:
        assert policy in normalized, f"{path}: Shepherd contract omits human gate: {policy}"
assert "Never merge." in pathlib.Path(sys.argv[2]).read_text()
PY

# The old auto-loop entry point must not interpret terminal state independently. It must exec the
# canonical Shepherd driver so the driver's terminal status reaches the caller unchanged.
auto_repo="$tmp_dir/auto-repo"
mkdir -p "$auto_repo/scripts"
cp "$repo_root/scripts/pi-auto-loop.sh" "$auto_repo/scripts/pi-auto-loop.sh"
fake_shepherd="$auto_repo/scripts/fake-shepherd.sh"
cat > "$fake_shepherd" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" > "${AUTO_DELEGATION_ARGS:?}"
exit 37
EOF
chmod +x "$fake_shepherd"
set +e
AUTO_DELEGATION_ARGS="$tmp_dir/auto-args" \
SHEPHERD_DRIVER="$fake_shepherd" \
PI_BIN=/usr/bin/false \
MAX_ITERATIONS=1 \
COOLDOWN_SECONDS=0 \
"$auto_repo/scripts/pi-auto-loop.sh" "delegation regression" \
  > "$tmp_dir/auto.log" 2>&1
rc=$?
set -e
if (( rc != 37 )); then
  echo "test failed: pi-auto-loop did not exec-propagate Shepherd status (got $rc, want 37)" >&2
  cat "$tmp_dir/auto.log" >&2
  exit 1
fi
[[ "$(cat "$tmp_dir/auto-args")" == "delegation regression" ]] || {
  echo "test failed: pi-auto-loop changed Shepherd arguments" >&2
  exit 1
}

# Build a minimal clean repository so exact HEAD/tree and dirty-worktree checks exercise the real
# main loop without depending on this development worktree's unrelated edits.
driver_repo="$tmp_dir/driver-repo"
mkdir -p "$driver_repo/scripts" \
  "$driver_repo/.agents/agentic-delivery/prompts" \
  "$driver_repo/.pi/extensions/pi-sub-agent"
printf 'base\n' > "$driver_repo/README.md"
git -C "$driver_repo" init -q
git -C "$driver_repo" config user.email test@example.invalid
git -C "$driver_repo" config user.name 'Shepherd Test'
git -C "$driver_repo" add README.md
git -C "$driver_repo" commit -qm base
test_base="$(git -C "$driver_repo" rev-parse HEAD)"
cp "$repo_root/scripts/pi-shepherd-loop.sh" "$driver_repo/scripts/pi-shepherd-loop.sh"
cp "$repo_root/scripts/pm-terminal-classifier.sh" "$driver_repo/scripts/pm-terminal-classifier.sh"
cp "$repo_root/.agents/agentic-delivery/prompts/shepherd-validator-prompt.md" \
  "$driver_repo/.agents/agentic-delivery/prompts/shepherd-validator-prompt.md"
printf '// test extension marker\n' > "$driver_repo/.pi/extensions/pi-sub-agent/index.ts"
git -C "$driver_repo" add .
git -C "$driver_repo" commit -qm driver
test_head="$(git -C "$driver_repo" rev-parse HEAD)"
test_tree="$(git -C "$driver_repo" rev-parse 'HEAD^{tree}')"
test_lineage="$test_base...r2-n3-test"

fake_orchestrator="$tmp_dir/fake-orchestrator.sh"
fake_validator="$tmp_dir/fake-validator.sh"

cat > "$fake_orchestrator" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
mkdir -p "$AUTO_LOOP_STATE_DIR"
reported_head="${REPORTED_HEAD:-$TEST_HEAD}"
reported_tree="${REPORTED_TREE:-$TEST_TREE}"
review_status="${REVIEW_STATUS:-clean}"
python3 - "$AUTO_LOOP_STATE_DIR" "$TEST_BASE" "$reported_head" "$reported_tree" "$TEST_LINEAGE" "$review_status" <<'PY'
import json
import pathlib
import sys

state_dir = pathlib.Path(sys.argv[1])
base, head, tree, lineage, review_status = sys.argv[2:]
synthesis = {
    "schema_version": "polymetrics.ai/pm-review-synthesis/v4",
    "owner": "parent_orchestrator",
    "exact_base_sha": base,
    "exact_head_sha": head,
    "exact_head_tree": tree,
    "status": "clean",
    "packet_count": 1,
    "response_count": 1,
    "findings": [],
    "blockers": [],
    "shepherd": {"status": "pending", "rule": "run independently only after clean synthesis"},
    "human_merge_authority": True,
}
synthesis_path = state_dir / "synthesis.json"
synthesis_path.write_text(json.dumps(synthesis) + "\n")
digest = "a" * 64
state = {
    "schema_version": "canonical_v2",
    "candidate_lineage": {
        "id": lineage,
        "exact_base_sha": base,
        "replacement_heads": [head],
    },
    "automated_review": {
        "primary_route": "local_codex",
        "status": review_status,
        "exact_base_sha": base,
        "exact_head_sha": head,
        "exact_head_tree": tree,
        "review_compiler": {
            "status": "ready",
            "exact_base_sha": base,
            "exact_head_sha": head,
            "exact_head_tree": tree,
            "authentication": {
                "algorithm": "sha256",
                "coverage_sha256": digest,
                "packets_sha256": digest,
                "semantic_manifest_sha256": digest,
            },
        },
        "local_codex": {
            "reviewer_identity": "fresh-test-reviewer",
            "fresh_context": True,
            "exact_base_sha": base,
            "exact_head_sha": head,
            "exact_head_tree": tree,
            "status": "clean",
            "synthesis_artifact": str(synthesis_path),
            "findings_artifact": "none",
            "lab_evidence_artifacts": [],
        },
        "shepherd": {
            "status": "pending",
            "exact_head_sha": head,
            "evidence_artifact": "pending",
        },
    },
}
(state_dir / "ORCHESTRATION-STATE.json").write_text(json.dumps(state) + "\n")
(state_dir / "RUN.json").write_text(json.dumps({
    "schema_version": "canonical_v2",
    "stage": "REVIEW",
    "terminal": "done",
    "candidate_lineage": lineage,
}) + "\n")
PY
EOF

cat > "$fake_validator" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
output_dir="${SHEPHERD_VALIDATOR_OUTPUT_DIR:-$AUTO_LOOP_STATE_DIR}"
source_dir="${SHEPHERD_STATE_ROOT:-$AUTO_LOOP_STATE_DIR}"
mkdir -p "$output_dir"
printf 'invoked\n' >> "${VALIDATOR_MARKER:?}"
python3 - "$output_dir" "$source_dir" "${VALIDATOR_IDENTITY_MODE:-match}" "${TEST_VERDICT:-PROCEED}" "${LEDGER_ATTACK_MODE:-none}" <<'PY'
import json
import pathlib
import sys

output_dir = pathlib.Path(sys.argv[1])
source_dir = pathlib.Path(sys.argv[2])
mode, verdict, ledger_attack = sys.argv[3:]
request = json.loads((source_dir / "SHEPHERD-REQUEST.json").read_text())
bound = {key: request[key] for key in (
    "exact_base_sha", "exact_head_sha", "exact_head_tree", "candidate_lineage", "synthesis_sha256"
)}
mismatch_fields = {
    "mismatch-base": ("exact_base_sha", "b" * 40),
    "mismatch-head": ("exact_head_sha", "f" * 40),
    "mismatch-tree": ("exact_head_tree", "c" * 40),
    "mismatch-lineage": ("candidate_lineage", "wrong-lineage"),
    "mismatch-synthesis": ("synthesis_sha256", "d" * 64),
}
if mode in mismatch_fields:
    field, replacement = mismatch_fields[mode]
    bound[field] = replacement
correction = "replay the rejected first transition" if verdict == "REVERT" else None
entry = {
    "schema_version": "polymetrics.ai/shepherd-validation/v1",
    "turn": 1,
    "stage_from": "REVIEW",
    "action": "validated exact clean synthesis",
    "stage_to": "REVIEW",
    **bound,
    "checks": {
        "correct_stage": 5,
        "artifact_valid": 5,
        "gates_respected": 5,
        "real_progress": 5,
        "no_hallucination": 5,
        "no_conflict": 5,
    },
    "step_score": 5,
    "trajectory_geomean": 5,
    "verdict": verdict,
    "reason": "test exact identity",
    "correction": correction,
}
if ledger_attack == "truncate":
    (source_dir / "VALIDATION.jsonl").write_text("")
with (output_dir / "VALIDATION.jsonl").open("a") as fh:
    fh.write(json.dumps(entry) + "\n")
verdict_record = {
    "schema_version": "polymetrics.ai/shepherd-verdict/v1",
    **bound,
    "verdict": verdict,
    "step_score": 5,
    "trajectory_geomean": 5,
    "reason": "test exact identity",
    "correction": correction,
    "revert_to_checkpoint": "initial" if verdict == "REVERT" else None,
}
(output_dir / "VALIDATOR-VERDICT.json").write_text(json.dumps(verdict_record) + "\n")
PY
EOF

chmod +x "$fake_orchestrator" "$fake_validator"

run_driver() { # $1=state-dir; remaining args are optional env assignments before the command.
  local state_dir="$1"
  shift
  mkdir -p "$state_dir"
  env \
    AUTO_LOOP_STATE_DIR="$state_dir" \
    PI_BIN="$fake_orchestrator" \
    VALIDATOR_BIN="$fake_validator" \
    VALIDATOR_MARKER="$state_dir/validator-invoked" \
    TEST_BASE="$test_base" \
    TEST_HEAD="$test_head" \
    TEST_TREE="$test_tree" \
    TEST_LINEAGE="$test_lineage" \
    MAX_ITERATIONS=1 \
    MAX_NO_VERDICT=1 \
    COOLDOWN_SECONDS=0 \
    WATCHDOG_POLL_SECONDS=0.05 \
    VALIDATOR_TIMEOUT_SECONDS=3 \
    "$@" \
    "$driver_repo/scripts/pi-shepherd-loop.sh" "exact identity regression"
}

# A matching, clean, authenticated synthesis and verdict may authorize terminal success.
matching_state="$tmp_dir/matching-state"
run_driver "$matching_state" > "$tmp_dir/matching.log" 2>&1
[[ -f "$matching_state/checkpoints/initial/HEAD.sha" ]] || {
  echo "test failed: initial pre-turn checkpoint was not created" >&2
  exit 1
}
grep -q 'DONE: all sub-issues complete and verified' "$tmp_dir/matching.log"
python3 - "$matching_state/SHEPHERD-REQUEST.json" "$matching_state/synthesis.json" <<'PY'
import hashlib
import json
import sys
request = json.load(open(sys.argv[1]))
assert request["synthesis_sha256"] == hashlib.sha256(open(sys.argv[2], "rb").read()).hexdigest()
assert set((
    "exact_base_sha", "exact_head_sha", "exact_head_tree", "candidate_lineage", "synthesis_sha256"
)).issubset(request)
PY
python3 - "$matching_state/VALIDATION.jsonl" <<'PY'
import json
import pathlib
import sys
raw = pathlib.Path(sys.argv[1]).read_bytes()
assert raw.endswith(b"\n")
lines = raw.splitlines()
assert len(lines) == 1, lines
assert json.loads(lines[0])["schema_version"] == "polymetrics.ai/shepherd-validation/v1"
PY

# Shepherd must not run, and a terminal state must not succeed, before authenticated clean review.
set +e
run_driver "$tmp_dir/pending-synthesis-state" REVIEW_STATUS=pending \
  > "$tmp_dir/pending-synthesis.log" 2>&1
rc=$?
set -e
if (( rc == 0 )); then
  echo "test failed: terminal succeeded before clean synthesis" >&2
  cat "$tmp_dir/pending-synthesis.log" >&2
  exit 1
fi
[[ ! -e "$tmp_dir/pending-synthesis-state/validator-invoked" ]] || {
  echo "test failed: Shepherd ran before clean synthesis" >&2
  exit 1
}
grep -q 'Shepherd deferred until authenticated clean exact-head synthesis' \
  "$tmp_dir/pending-synthesis.log"

# Every field in the exact identity tuple is mandatory; a fresh misbound verdict must not authorize
# the model-written terminal state.
for mismatch in base head tree lineage synthesis; do
  mismatch_state="$tmp_dir/mismatch-$mismatch-state"
  set +e
  run_driver "$mismatch_state" VALIDATOR_IDENTITY_MODE="mismatch-$mismatch" \
    > "$tmp_dir/mismatch-$mismatch.log" 2>&1
  rc=$?
  set -e
  if (( rc == 0 )); then
    echo "test failed: mismatched Shepherd $mismatch authorized terminal success" >&2
    cat "$tmp_dir/mismatch-$mismatch.log" >&2
    exit 1
  fi
  grep -Eq 'identity|bound validation' "$tmp_dir/mismatch-$mismatch.log"
done

# A synthesis consistently bound to a stale head is still stale against the current repository.
set +e
run_driver "$tmp_dir/stale-state" REPORTED_HEAD="$(printf 'e%.0s' {1..40})" \
  > "$tmp_dir/stale.log" 2>&1
rc=$?
set -e
if (( rc == 0 )); then
  echo "test failed: stale exact-head synthesis authorized terminal success" >&2
  cat "$tmp_dir/stale.log" >&2
  exit 1
fi
[[ ! -e "$tmp_dir/stale-state/validator-invoked" ]] || {
  echo "test failed: validator ran against stale synthesis identity" >&2
  exit 1
}

# A dirty candidate invalidates the exact-head synthesis before the validator is invoked.
printf 'dirty\n' >> "$driver_repo/README.md"
set +e
run_driver "$tmp_dir/dirty-state" > "$tmp_dir/dirty.log" 2>&1
rc=$?
set -e
git -C "$driver_repo" checkout -q -- README.md
if (( rc == 0 )); then
  echo "test failed: dirty candidate authorized terminal success" >&2
  cat "$tmp_dir/dirty.log" >&2
  exit 1
fi
[[ ! -e "$tmp_dir/dirty-state/validator-invoked" ]] || {
  echo "test failed: validator ran against dirty candidate identity" >&2
  exit 1
}

# A first-turn REVERT restores the verified initial checkpoint rather than continuing from the
# rejected transition.
set +e
run_driver "$tmp_dir/revert-state" TEST_VERDICT=REVERT > "$tmp_dir/revert.log" 2>&1
rc=$?
set -e
if (( rc == 0 )); then
  echo "test failed: REVERT unexpectedly reported terminal success" >&2
  exit 1
fi
python3 - "$tmp_dir/revert-state/REVERT-CLEANUP.json" <<'PY'
import json
import sys
record = json.load(open(sys.argv[1]))
assert record["checkpoint"] == "initial", record
assert record["good_fork_sha"], record
PY
if grep -q 'no checkpoint to revert to' "$tmp_dir/revert.log"; then
  echo "test failed: first-turn REVERT had no checkpoint" >&2
  exit 1
fi

# A validator process failure cannot retain a forged PROCEED or authorize terminal success.
failed_validator="$tmp_dir/failed-validator.sh"
cat > "$failed_validator" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
mkdir -p "$AUTO_LOOP_STATE_DIR"
printf '{"verdict":"PROCEED","step_score":5,"reason":"must be ignored"}\n' \
  > "$AUTO_LOOP_STATE_DIR/VALIDATOR-VERDICT.json"
exit 9
EOF
chmod +x "$failed_validator"
set +e
AUTO_LOOP_STATE_DIR="$tmp_dir/failed-validator-state" \
PI_BIN="$fake_orchestrator" \
VALIDATOR_BIN="$failed_validator" \
VALIDATOR_MARKER="$tmp_dir/unused-validator-marker" \
TEST_BASE="$test_base" \
TEST_HEAD="$test_head" \
TEST_TREE="$test_tree" \
TEST_LINEAGE="$test_lineage" \
MAX_ITERATIONS=1 \
MAX_NO_VERDICT=1 \
COOLDOWN_SECONDS=0 \
WATCHDOG_POLL_SECONDS=0.05 \
VALIDATOR_TIMEOUT_SECONDS=3 \
"$driver_repo/scripts/pi-shepherd-loop.sh" "failed-validator regression" \
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

# The validator verdict pathname is adversarial. Swap it immediately after bound verification; the
# driver must continue from its one-time immutable snapshot rather than reread the swapped path.
real_python3="$(command -v python3)"
python_wrapper_dir="$tmp_dir/python-wrapper"
mkdir -p "$python_wrapper_dir"
cat > "$python_wrapper_dir/python3" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [[ "${2:-}" == *SHEPHERD-REQUEST.json ]] && [[ "${3:-}" == *VERDICT* ]]; then
  "${REAL_PYTHON3:?}" "$@"
  rc=$?
  if (( rc == 0 )); then
    validator_path="$(dirname "$3")/VALIDATOR-VERDICT.json"
    replacement="$validator_path.replacement"
    printf '{"verdict":"HALT","reason":"swapped after verification"}\n' > "$replacement"
    mv -f "$replacement" "$validator_path"
  fi
  exit "$rc"
fi
exec "${REAL_PYTHON3:?}" "$@"
EOF
chmod +x "$python_wrapper_dir/python3"
swap_state="$tmp_dir/verdict-swap-state"
PATH="$python_wrapper_dir:$PATH" \
REAL_PYTHON3="$real_python3" \
run_driver "$swap_state" > "$tmp_dir/verdict-swap.log" 2>&1

grep -q 'DONE: all sub-issues complete and verified' "$tmp_dir/verdict-swap.log"
swapped_candidate="$(find "$swap_state/.validator-turns" -name VALIDATOR-VERDICT.json -type f -print -quit)"
if [[ -z "$swapped_candidate" ]] || ! grep -q 'swapped after verification' "$swapped_candidate"; then
  echo "test failed: adversarial verdict pathname swap did not execute" >&2
  exit 1
fi
# shellcheck disable=SC2016 # intentional literal source-code regression guard
if grep -Fq 'json_field "$VERDICT_JSON"' "$repo_root/scripts/pi-shepherd-loop.sh"; then
  echo "test failed: driver still rereads the validator-controlled verdict pathname" >&2
  exit 1
fi

# The driver, not the validator, appends the accepted record while preserving every historical byte.
append_state="$tmp_dir/ledger-append-state"
mkdir -p "$append_state"
printf '{"historical":"byte-exact prefix"}\n' > "$append_state/VALIDATION.jsonl"
cp "$append_state/VALIDATION.jsonl" "$tmp_dir/append-before"
run_driver "$append_state" > "$tmp_dir/ledger-append.log" 2>&1
python3 - "$tmp_dir/append-before" "$append_state/VALIDATION.jsonl" <<'PY'
import json
import pathlib
import sys
before = pathlib.Path(sys.argv[1]).read_bytes()
after = pathlib.Path(sys.argv[2]).read_bytes()
assert after[:len(before)] == before
lines = after.splitlines()
assert len(lines) == 2, lines
assert json.loads(lines[-1])["schema_version"] == "polymetrics.ai/shepherd-validation/v1"
PY

# A validator may not rewrite/truncate history and replace it with one superficially valid record.
# Failure must restore the ledger's exact pre-run bytes and reject the turn.
truncated_state="$tmp_dir/ledger-truncation-state"
mkdir -p "$truncated_state"
printf '{"historical":"exact bytes must survive"}\n' > "$truncated_state/VALIDATION.jsonl"
cp "$truncated_state/VALIDATION.jsonl" "$tmp_dir/ledger-before"
set +e
run_driver "$truncated_state" LEDGER_ATTACK_MODE=truncate \
  > "$tmp_dir/ledger-truncation.log" 2>&1
rc=$?
set -e
if (( rc == 0 )); then
  echo "test failed: validator ledger truncation authorized terminal success" >&2
  exit 1
fi
cmp -s "$tmp_dir/ledger-before" "$truncated_state/VALIDATION.jsonl" || {
  echo "test failed: exact pre-run validation ledger bytes were not restored" >&2
  exit 1
}

# Checkpoint roots are driver authority. A symlink must fail closed without writing through it.
checkpoint_state="$tmp_dir/checkpoint-symlink-state"
checkpoint_victim="$tmp_dir/checkpoint-victim"
mkdir -p "$checkpoint_state" "$checkpoint_victim"
ln -s "$checkpoint_victim" "$checkpoint_state/checkpoints"
set +e
run_driver "$checkpoint_state" > "$tmp_dir/checkpoint-symlink.log" 2>&1
rc=$?
set -e
if (( rc == 0 )); then
  echo "test failed: symlink checkpoint root was accepted" >&2
  exit 1
fi
if find "$checkpoint_victim" -mindepth 1 -print -quit | grep -q .; then
  echo "test failed: checkpoint write followed the symlink root" >&2
  exit 1
fi

# A hardlinked checkpoint authority file is equally unsafe: replacing or rewriting it could mutate
# an object outside the private checkpoint root.
hardlink_state="$tmp_dir/checkpoint-hardlink-state"
mkdir -p "$hardlink_state/checkpoints"
chmod 700 "$hardlink_state/checkpoints"
printf 'initial\n' > "$tmp_dir/hardlink-marker-source"
ln "$tmp_dir/hardlink-marker-source" "$hardlink_state/checkpoints/LAST_GOOD"
set +e
run_driver "$hardlink_state" > "$tmp_dir/checkpoint-hardlink.log" 2>&1
rc=$?
set -e
if (( rc == 0 )); then
  echo "test failed: hardlinked checkpoint authority was accepted" >&2
  exit 1
fi
[[ "$(cat "$tmp_dir/hardlink-marker-source")" == "initial" ]] || {
  echo "test failed: hardlinked checkpoint source was modified" >&2
  exit 1
}

echo "main-loop tests passed: exact identity, immutable verdict/ledger, and safe checkpoints enforced"
