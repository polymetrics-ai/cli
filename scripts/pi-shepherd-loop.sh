#!/usr/bin/env bash
# Codex-only autonomous Pi orchestration driver WITH the Shepherd validator layer.
#
# Best-of-both-worlds merge of scripts/pi-auto-loop.sh (pi-native orchestrator with the
# `subagent` tool and the .pi/agents/pm-* roster) and scripts/claude-auto-loop.sh (Shepherd
# supervision: after exact-head review is clean, an independent validator judges the accumulated
# trajectory with checkpoints, RETRY corrections, REVERT-to-checkpoint, and HALT).
#
# Each iteration:
#   1. ORCHESTRATOR turn  — `pi -p` runs the loop prompt (default /pm-auto-loop; use
#        LOOP_CMD=/pm-connector-loop for connector runs). It RECONCILES from durable state,
#        advances exactly ONE stage, and dispatches implementation via the pi `subagent` tool
#        (or the detached-worker recipe for long EXECUTE stages).
#   2. ELIGIBILITY/GATE  — authenticate clean synthesis plus exact base/head/tree/lineage/hash.
#        Before clean synthesis, defer Shepherd and continue preparatory turns. Once eligible, an
#        independent Shepherd re-derives ground truth, scores the trajectory, writes a bound
#        VALIDATOR-VERDICT.json, and appends one bound VALIDATION.jsonl entry under a watchdog.
#   3. ACT on the verdict — PROCEED (checkpoint + continue) / RETRY (replay stage with the
#        correction) / REVERT (restore last checkpoint, write REVERT-CLEANUP.json, replay) /
#        HALT (stop for a human).
#
# All progress is durable (git + GitHub + RUN.json/ORCHESTRATION-STATE.json), so a run killed
# anywhere resumes by reconciling: scripts/pi-shepherd-loop.sh --resume
#
# Model policy: Codex on the ChatGPT plan via the `openai-codex/*` provider ONLY. Never route
# any role through OpenRouter or another pay-per-token gateway.
#
# Requires the `subagent` tool package once per machine:  pi install npm:pi-sub-agent
# (project agents in .pi/agents/ are auto-discovered when running with --approve).
#
# Usage:
#   scripts/pi-shepherd-loop.sh "Add full CLI parity for the <name> connector"
#   scripts/pi-shepherd-loop.sh --resume
#
# Config (env; defaults shown):
#   PI_BIN=pi
#   ORCH_MODEL=openai-codex/gpt-5.6-sol       # orchestrator model
#   ORCH_THINKING=xhigh                       # orchestrator reasoning effort
#   PI_TOOLS=read,bash,edit,write,grep,find,ls,subagent
#   VALIDATOR_BIN=pi                          # Shepherd CLI (cross-model judging is a feature)
#   VALIDATOR_ARGS="--model openai-codex/gpt-5.6-sol --thinking xhigh --tools read,bash,edit,write,grep,find,ls --approve"
#   MAX_ITERATIONS=200                        # hard backstop on orchestrator turns
#   MAX_REVERTS=6                             # total revert budget per run before HALT
#   MAX_NO_VERDICT=3                          # consecutive no-verdict turns before HALT
#   MAX_MINUTES=0                             # wall-clock cap (0 = none)
#   COOLDOWN_SECONDS=5
#   PI_EXTRA_FLAGS=""                         # extra flags for every orchestrator invocation
#   LOOP_CMD=/pm-auto-loop                    # /pm-connector-loop for connector runs
#   AUTO_LOOP_STATE_DIR=.planning/auto-loop   # set to isolate separate Shepherd runs
#   STALL_MINUTES=90                          # allow full verify/race stages to report progress
#   WATCHDOG_POLL_SECONDS=60                  # process/session liveness poll interval
#   TURN_TIMEOUT_SECONDS=1800                 # unconditional orchestrator turn deadline
#   VALIDATOR_TIMEOUT_SECONDS=900             # hard validator wall-clock/process-tree timeout
#   NODE_BIN_DIR=~/.nvm/versions/node/v24.13.1/bin  # prepended when present so pi uses current Node
#   SEARXNG_BASE=                             # research via the audited searxng connector (pm)
set -euo pipefail
umask 077

NODE_BIN_DIR="${NODE_BIN_DIR:-$HOME/.nvm/versions/node/v24.13.1/bin}"
if [[ -x "$NODE_BIN_DIR/node" ]]; then
  PATH="$NODE_BIN_DIR:$PATH"; export PATH
fi

PI_BIN="${PI_BIN:-pi}"
ORCH_MODEL="${ORCH_MODEL:-openai-codex/gpt-5.6-sol}"
ORCH_THINKING="${ORCH_THINKING:-xhigh}"
PI_TOOLS="${PI_TOOLS:-read,bash,edit,write,grep,find,ls,subagent}"
VALIDATOR_BIN="${VALIDATOR_BIN:-pi}"
VALIDATOR_ARGS="${VALIDATOR_ARGS:---model openai-codex/gpt-5.6-sol --thinking xhigh --tools read,bash,edit,write,grep,find,ls --approve}"
MAX_ITERATIONS="${MAX_ITERATIONS:-200}"
MAX_REVERTS="${MAX_REVERTS:-6}"
MAX_NO_VERDICT="${MAX_NO_VERDICT:-3}"
MAX_MINUTES="${MAX_MINUTES:-0}"
COOLDOWN_SECONDS="${COOLDOWN_SECONDS:-5}"
PI_EXTRA_FLAGS="${PI_EXTRA_FLAGS:-}"
LOOP_CMD="${LOOP_CMD:-/pm-auto-loop}"
# Research: default SEARXNG_BASE from the shell's SEARXNG_URL (name mismatch guard) and export.
SEARXNG_BASE="${SEARXNG_BASE:-${SEARXNG_URL:-}}"; export SEARXNG_BASE
STALL_MINUTES="${STALL_MINUTES:-90}"
STALL_KILL_LIVE_CHILDREN="${STALL_KILL_LIVE_CHILDREN:-1}"
WATCHDOG_POLL_SECONDS="${WATCHDOG_POLL_SECONDS:-60}"
TURN_TIMEOUT_SECONDS="${TURN_TIMEOUT_SECONDS:-1800}"
VALIDATOR_TIMEOUT_SECONDS="${VALIDATOR_TIMEOUT_SECONDS:-900}"
for timeout_name in TURN_TIMEOUT_SECONDS VALIDATOR_TIMEOUT_SECONDS; do
  timeout_value="${!timeout_name}"
  if [[ ! "$timeout_value" =~ ^[1-9][0-9]*$ ]]; then
    echo "FATAL: $timeout_name must be a positive integer" >&2
    exit 2
  fi
done

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATE_DIR="${AUTO_LOOP_STATE_DIR:-$REPO_ROOT/.planning/auto-loop}"
[[ "$STATE_DIR" = /* ]] || STATE_DIR="$REPO_ROOT/$STATE_DIR"
# Subagent observability: locally-patched pi-sub-agent records child sessions here (opt-in).
PI_SUBAGENT_SESSION_DIR="${PI_SUBAGENT_SESSION_DIR:-$STATE_DIR/sessions}"; export PI_SUBAGENT_SESSION_DIR
CKPT_DIR="$STATE_DIR/checkpoints"
RUN_JSON="$STATE_DIR/RUN.json"
ORCHESTRATION_STATE_JSON="$STATE_DIR/ORCHESTRATION-STATE.json"
REQUEST_JSON="$STATE_DIR/SHEPHERD-REQUEST.json"
VALIDATION_JSONL="$STATE_DIR/VALIDATION.jsonl"
VERDICT_JSON="$STATE_DIR/VALIDATOR-VERDICT.json"
PROMPT_FILE="$STATE_DIR/PROMPT.txt"
LOG_FILE="$STATE_DIR/driver.log"
VAL_PROMPT="$REPO_ROOT/.agents/agentic-delivery/prompts/shepherd-validator-prompt.md"

# Checkpoint authority must never be rooted through a symlink or a shared/preexisting unsafe
# directory. Files below it are created separately with openat-style no-follow/exclusive writes.
python3 - "$STATE_DIR" "$CKPT_DIR" <<'PY'
import os
import pathlib
import stat
import sys

state = pathlib.Path(sys.argv[1])
checkpoint = pathlib.Path(sys.argv[2])
try:
    state.mkdir(parents=True, mode=0o700, exist_ok=True)
    state_stat = os.lstat(state)
    if stat.S_ISLNK(state_stat.st_mode) or not stat.S_ISDIR(state_stat.st_mode):
        raise OSError("state root is not a real directory")
    if state_stat.st_uid != os.geteuid():
        raise OSError("state root is not driver-owned")
    try:
        checkpoint_stat = os.lstat(checkpoint)
    except FileNotFoundError:
        os.mkdir(checkpoint, 0o700)
        checkpoint_stat = os.lstat(checkpoint)
    if stat.S_ISLNK(checkpoint_stat.st_mode) or not stat.S_ISDIR(checkpoint_stat.st_mode):
        raise OSError("checkpoint root is not a real directory")
    if checkpoint_stat.st_uid != os.geteuid() or stat.S_IMODE(checkpoint_stat.st_mode) & 0o077:
        raise OSError("checkpoint root is not private and driver-owned")
    flags = os.O_RDONLY | getattr(os, "O_DIRECTORY", 0) | getattr(os, "O_NOFOLLOW", 0)
    fd = os.open(checkpoint, flags)
    os.close(fd)
except OSError as exc:
    print(f"FATAL: unsafe checkpoint root: {exc}", file=sys.stderr)
    raise SystemExit(2)
PY

log() { printf '%s %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*" | tee -a "$LOG_FILE" >&2; }

state_root_instruction() {
  cat <<EOF
SHEPHERD STATE ROOT FOR THIS RUN:
- Use $STATE_DIR as the run ledger root.
- Read/write RUN.json, ORCHESTRATION-STATE.json, VALIDATION.jsonl, VALIDATOR-VERDICT.json, checkpoints, tasks, trace, sessions, and RESEARCH under $STATE_DIR.
- If any prompt, workflow, or reference says .planning/auto-loop, interpret that as $STATE_DIR for this run.
- Do not read or write $REPO_ROOT/.planning/auto-loop for this run unless explicitly using it as historical evidence and clearly labeling it historical.
EOF
}

json_field() { # $1=file $2=key
  [[ -f "$1" ]] || { echo ""; return 0; }
  python3 - "$1" "$2" <<'PY' 2>/dev/null || echo ""
import json,sys
try:
    d=json.load(open(sys.argv[1])); v=d.get(sys.argv[2])
    if isinstance(v,dict): v=v.get("type","")
    print("" if v is None else v)
except Exception: print("")
PY
}

validated_verdict=""
VALIDATOR_TURN_DIR=""
VERDICT_CANDIDATE=""
VALIDATION_CANDIDATE=""
VERDICT_SNAPSHOT=""
VALIDATION_RECORD_SNAPSHOT=""
LEDGER_BASE_SNAPSHOT=""
LEDGER_BASE_PRESENT=0
LEDGER_BASE_HASH=""
LEDGER_FINALIZED=1
ACTIVE_VALIDATOR_PID=""
ACTIVE_VALIDATOR_SID=""

secure_remove_validator_paths() {
  python3 - "$VERDICT_JSON" "$VERDICT_CANDIDATE" "$VALIDATION_CANDIDATE" <<'PY'
import os
import stat
import sys

for raw in sys.argv[1:]:
    if not raw:
        continue
    try:
        current = os.lstat(raw)
    except FileNotFoundError:
        continue
    if stat.S_ISDIR(current.st_mode):
        print(f"SHEPHERD OUTPUT BLOCK: refusing to remove directory output {raw}", file=sys.stderr)
        raise SystemExit(1)
    os.unlink(raw)
PY
}

prepare_validator_turn() {
  local turn_meta ledger_meta
  validated_verdict=""
  LEDGER_FINALIZED=1
  LEDGER_BASE_SNAPSHOT=""
  VERDICT_SNAPSHOT=""
  VALIDATION_RECORD_SNAPSHOT=""
  secure_remove_validator_paths || return 1
  turn_meta="$(python3 - "$STATE_DIR" <<'PY'
import os
import pathlib
import stat
import sys
import tempfile

state = pathlib.Path(sys.argv[1])
root = state / ".validator-turns"
try:
    try:
        current = os.lstat(root)
    except FileNotFoundError:
        os.mkdir(root, 0o700)
        current = os.lstat(root)
    if stat.S_ISLNK(current.st_mode) or not stat.S_ISDIR(current.st_mode):
        raise OSError("validator turn root is not a real directory")
    if current.st_uid != os.geteuid() or stat.S_IMODE(current.st_mode) & 0o077:
        raise OSError("validator turn root is not private and driver-owned")
    print(tempfile.mkdtemp(prefix="turn-", dir=root))
except OSError as exc:
    print(f"SHEPHERD OUTPUT BLOCK: {exc}", file=sys.stderr)
    raise SystemExit(1)
PY
)" || return 1
  VALIDATOR_TURN_DIR="$turn_meta"
  VERDICT_CANDIDATE="$VALIDATOR_TURN_DIR/VALIDATOR-VERDICT.json"
  VALIDATION_CANDIDATE="$VALIDATOR_TURN_DIR/VALIDATION.jsonl"
  VERDICT_SNAPSHOT="$VALIDATOR_TURN_DIR/VERDICT.snapshot"
  VALIDATION_RECORD_SNAPSHOT="$VALIDATOR_TURN_DIR/VALIDATION.snapshot"
  LEDGER_BASE_SNAPSHOT="$VALIDATOR_TURN_DIR/LEDGER.before"
  ledger_meta="$(python3 - "$VALIDATION_JSONL" "$LEDGER_BASE_SNAPSHOT" "$VALIDATOR_TURN_DIR/LEDGER.absent" <<'PY'
import hashlib
import os
import stat
import sys

source, snapshot, absent = sys.argv[1:]
flags = os.O_RDONLY | getattr(os, "O_NOFOLLOW", 0)
create = os.O_WRONLY | os.O_CREAT | os.O_EXCL | getattr(os, "O_NOFOLLOW", 0)
present = 1
source_fd = None
try:
    try:
        source_fd = os.open(source, flags)
    except FileNotFoundError:
        present = 0
    if source_fd is not None:
        source_stat = os.fstat(source_fd)
        if not stat.S_ISREG(source_stat.st_mode) or source_stat.st_nlink != 1:
            raise OSError("validation ledger is not a single-link regular file")
        if source_stat.st_uid != os.geteuid():
            raise OSError("validation ledger is not driver-owned")
    snapshot_fd = os.open(snapshot, create, 0o600)
    digest = hashlib.sha256()
    with os.fdopen(snapshot_fd, "wb", closefd=True) as output:
        if source_fd is not None:
            with os.fdopen(source_fd, "rb", closefd=True) as source_file:
                source_fd = None
                while True:
                    chunk = source_file.read(1024 * 1024)
                    if not chunk:
                        break
                    digest.update(chunk)
                    output.write(chunk)
        output.flush()
        os.fsync(output.fileno())
        os.fchmod(output.fileno(), 0o400)
    if not present:
        marker = os.open(absent, create, 0o400)
        os.close(marker)
    print(f"{digest.hexdigest()} {present}")
except OSError as exc:
    if source_fd is not None:
        os.close(source_fd)
    print(f"SHEPHERD LEDGER BLOCK: {exc}", file=sys.stderr)
    raise SystemExit(1)
PY
)" || return 1
  read -r LEDGER_BASE_HASH LEDGER_BASE_PRESENT <<< "$ledger_meta"
  if [[ "$LEDGER_BASE_HASH" =~ ^[0-9a-f]{64}$ && "$LEDGER_BASE_PRESENT" =~ ^[01]$ ]]; then
    LEDGER_FINALIZED=0
    return 0
  fi
  return 1
}

terminal_is_authoritative() {
  # Terminal state written by a rejected, stale, dirty, or unvalidated turn is never success.
  [[ "$validated_verdict" == "PROCEED" ]]
}

if [[ "${SHEPHERD_VERDICT_GUARD_SELF_TEST:-}" == "1" ]]; then
  printf '{"verdict":"PROCEED"}\n' > "$VERDICT_JSON"
  prepare_validator_turn
  [[ ! -e "$VERDICT_JSON" ]] || { echo "self-test failed: stale validator verdict survived" >&2; exit 1; }
  validated_verdict=PROCEED
  terminal_is_authoritative || { echo "self-test failed: bound PROCEED terminal rejected" >&2; exit 1; }
  for rejected in RETRY REVERT HALT ""; do
    validated_verdict="$rejected"
    if terminal_is_authoritative; then
      echo "self-test failed: terminal accepted after ${rejected:-missing} verdict" >&2
      exit 1
    fi
  done
  echo "self-test passed: stale verdict removed; terminal requires fresh bound PROCEED"
  exit 0
fi

prepare_shepherd_request() { # $1=request output; returns 10 when review is not clean yet
  local output="${1:-$REQUEST_JSON}"
  python3 - "$REPO_ROOT" "$RUN_JSON" "$ORCHESTRATION_STATE_JSON" "$output" <<'PY'
import hashlib
import json
import os
import pathlib
import re
import subprocess
import sys
import tempfile

root = pathlib.Path(sys.argv[1]).resolve()
run_path = pathlib.Path(sys.argv[2])
state_path = pathlib.Path(sys.argv[3])
output = pathlib.Path(sys.argv[4])
hex40 = re.compile(r"^[0-9a-f]{40}$")
hex64 = re.compile(r"^[0-9a-f]{64}$")


def block(reason, code=1):
    print(f"SHEPHERD IDENTITY BLOCK: {reason}", file=sys.stderr)
    raise SystemExit(code)


def read_json(path, label):
    try:
        value = json.loads(path.read_text())
    except (OSError, json.JSONDecodeError, UnicodeError):
        block(f"{label} is missing or malformed")
    if not isinstance(value, dict):
        block(f"{label} must be an object")
    return value


def git(*args):
    result = subprocess.run(
        ["git", "-C", str(root), *args],
        stdout=subprocess.PIPE,
        stderr=subprocess.DEVNULL,
        text=True,
        check=False,
    )
    if result.returncode:
        block(f"git identity check failed: {args[0]}")
    return result.stdout.strip()


# Preparatory turns do not invoke Shepherd. Enforce exact cleanliness only once canonical review
# state claims a clean synthesis, so an in-progress implementation turn can still be reconciled.
if not state_path.is_file():
    block("canonical review state is not ready", 10)
state = read_json(state_path, "canonical review state")
run = read_json(run_path, "RUN.json") if run_path.is_file() else {}
if state.get("schema_version") != "canonical_v2":
    block("canonical review state schema is not exact current v2")

# Canonical ledgers may store review coverage at the run level or on the active subissue. Select one
# unambiguous clean review; never guess among multiple candidate lineages.
state_subissue = None
run_subissue = None
review = state.get("automated_review")
if not isinstance(review, dict) or review.get("status") != "clean":
    clean_subissues = [
        item
        for item in state.get("subissues", [])
        if isinstance(item, dict)
        and isinstance(item.get("automated_review"), dict)
        and item["automated_review"].get("status") == "clean"
    ]
    run_subissues = {
        item.get("number"): item
        for item in run.get("subissues", [])
        if isinstance(item, dict) and item.get("number") is not None
    }
    active = [
        item
        for item in clean_subissues
        if isinstance(run_subissues.get(item.get("number")), dict)
        and run_subissues[item.get("number")].get("stage")
        in {"review_pending", "sub_pr_reviewed", "integrating"}
    ]
    if len(active) == 1:
        state_subissue = active[0]
    elif not active and len(clean_subissues) == 1:
        state_subissue = clean_subissues[0]
    elif not clean_subissues:
        block("local-Codex review synthesis is not clean", 10)
    else:
        block("multiple clean candidate reviews make Shepherd identity ambiguous")
    review = state_subissue["automated_review"]
    run_subissue = run_subissues.get(state_subissue.get("number"))
if review.get("primary_route") != "local_codex":
    block("canonical local-Codex review route is not authenticated")

# Exact review evidence is meaningful only for a clean candidate. Do not print dirty path names.
head = git("rev-parse", "HEAD")
tree = git("rev-parse", "HEAD^{tree}")
if git("status", "--porcelain=v1", "--untracked-files=all"):
    block("candidate worktree is dirty")
if not hex40.fullmatch(head) or not hex40.fullmatch(tree):
    block("current HEAD/tree is malformed")

base = review.get("exact_base_sha")
expected_head = review.get("exact_head_sha")
expected_tree = review.get("exact_head_tree")
for label, value in (
    ("exact_base_sha", base),
    ("exact_head_sha", expected_head),
    ("exact_head_tree", expected_tree),
):
    if not isinstance(value, str) or not hex40.fullmatch(value):
        block(f"{label} is missing or malformed")
if expected_head != head or expected_tree != tree:
    block("review identity is stale against current HEAD/tree")

# The exact base must be a real commit and the merge base for this candidate.
git("cat-file", "-e", f"{base}^{{commit}}")
git("cat-file", "-e", f"{expected_head}^{{commit}}")
if git("merge-base", base, expected_head) != base:
    block("exact_base_sha is not the candidate merge base")

lineage_record = None
for owner in (state_subissue, run_subissue, state, run):
    if isinstance(owner, dict) and owner.get("candidate_lineage") is not None:
        lineage_record = owner["candidate_lineage"]
        break
if isinstance(lineage_record, dict):
    lineage = lineage_record.get("id")
    lineage_base = lineage_record.get("exact_base_sha")
    replacements = lineage_record.get("replacement_heads")
    if lineage_base is not None and lineage_base != base:
        block("candidate lineage base disagrees with exact_base_sha")
    if replacements is not None and (
        not isinstance(replacements, list) or expected_head not in replacements
    ):
        block("candidate lineage does not contain exact_head_sha")
else:
    lineage = lineage_record
if not isinstance(lineage, str) or not lineage or len(lineage) > 512 or any(ord(c) < 32 for c in lineage):
    block("candidate_lineage is missing or malformed")
run_lineage = (
    run_subissue.get("candidate_lineage")
    if isinstance(run_subissue, dict)
    else run.get("candidate_lineage")
)
if isinstance(run_lineage, dict):
    run_lineage = run_lineage.get("id")
if run_lineage is not None and run_lineage != lineage:
    block("RUN.json candidate_lineage disagrees with canonical state")

compiler = review.get("review_compiler")
if not isinstance(compiler, dict) or compiler.get("status") != "ready":
    block("review compiler is not ready/authenticated")
for field, expected in (
    ("exact_base_sha", base),
    ("exact_head_sha", expected_head),
    ("exact_head_tree", expected_tree),
):
    if compiler.get(field) != expected:
        block(f"review compiler {field} mismatch")
auth = compiler.get("authentication")
if not isinstance(auth, dict) or auth.get("algorithm") != "sha256":
    block("review compiler authentication is missing")
for field in ("coverage_sha256", "packets_sha256", "semantic_manifest_sha256"):
    if not isinstance(auth.get(field), str) or not hex64.fullmatch(auth[field]):
        block(f"review compiler authentication {field} is malformed")

local_codex = review.get("local_codex")
if not isinstance(local_codex, dict):
    block("local-Codex synthesis record is missing")
if local_codex.get("status") != "clean" or local_codex.get("fresh_context") is not True:
    block("local-Codex synthesis is not fresh and clean")
for field, expected in (
    ("exact_base_sha", base),
    ("exact_head_sha", expected_head),
    ("exact_head_tree", expected_tree),
):
    if local_codex.get(field) != expected:
        block(f"local-Codex {field} mismatch")

raw_synthesis_path = local_codex.get("synthesis_artifact")
if not isinstance(raw_synthesis_path, str) or not raw_synthesis_path:
    block("synthesis artifact path is missing")
synthesis_path = pathlib.Path(raw_synthesis_path)
if not synthesis_path.is_absolute():
    synthesis_path = root / synthesis_path
if synthesis_path.is_symlink():
    block("synthesis artifact must not be a symlink")
try:
    synthesis_path = synthesis_path.resolve(strict=True)
    stat = synthesis_path.stat()
except OSError:
    block("synthesis artifact is missing")
if not synthesis_path.is_file() or stat.st_size > 50 * 1024 * 1024:
    block("synthesis artifact is not a bounded regular file")
try:
    synthesis_bytes = synthesis_path.read_bytes()
    synthesis = json.loads(synthesis_bytes)
except (OSError, json.JSONDecodeError, UnicodeError):
    block("synthesis artifact is malformed")
if not isinstance(synthesis, dict):
    block("synthesis artifact must be an object")
if synthesis.get("schema_version") != "polymetrics.ai/pm-review-synthesis/v4":
    block("synthesis schema is not exact current v4")
if synthesis.get("owner") != "parent_orchestrator" or synthesis.get("status") != "clean":
    block("synthesis is not PM-owned and clean")
if synthesis.get("findings") != [] or synthesis.get("blockers") != []:
    block("clean synthesis contains findings or blockers")
if synthesis.get("human_merge_authority") is not True:
    block("synthesis does not preserve human merge authority")
if not isinstance(synthesis.get("shepherd"), dict) or synthesis["shepherd"].get("status") != "pending":
    block("synthesis does not place Shepherd downstream")
for field, expected in (
    ("exact_base_sha", base),
    ("exact_head_sha", expected_head),
    ("exact_head_tree", expected_tree),
):
    if synthesis.get(field) != expected:
        block(f"synthesis {field} mismatch")

# Recheck after every state/artifact read to reject concurrent drift.
if git("rev-parse", "HEAD") != head or git("rev-parse", "HEAD^{tree}") != tree:
    block("candidate identity changed during authentication")
if git("status", "--porcelain=v1", "--untracked-files=all"):
    block("candidate became dirty during authentication")

def digest_file(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

request = {
    "schema_version": "polymetrics.ai/shepherd-request/v1",
    "exact_base_sha": base,
    "exact_head_sha": expected_head,
    "exact_head_tree": expected_tree,
    "candidate_lineage": lineage,
    "synthesis_sha256": hashlib.sha256(synthesis_bytes).hexdigest(),
    "synthesis_artifact": str(synthesis_path),
    "run_stage": run_subissue.get("stage") if isinstance(run_subissue, dict) else run.get("stage"),
    "run_state_sha256": digest_file(run_path) if run_path.is_file() else None,
    "orchestration_state_sha256": digest_file(state_path),
}
output.parent.mkdir(parents=True, exist_ok=True)
fd, temporary = tempfile.mkstemp(prefix=f".{output.name}.", dir=output.parent)
try:
    with os.fdopen(fd, "w") as fh:
        json.dump(request, fh, sort_keys=True, indent=2)
        fh.write("\n")
        fh.flush()
        os.fsync(fh.fileno())
    os.chmod(temporary, 0o600)
    os.replace(temporary, output)
finally:
    try:
        os.unlink(temporary)
    except FileNotFoundError:
        pass
PY
}

revalidate_shepherd_request() {
  local post="$STATE_DIR/.SHEPHERD-REQUEST.post.$$"
  if ! prepare_shepherd_request "$post"; then
    rm -f "$post"
    return 1
  fi
  if ! cmp -s "$REQUEST_JSON" "$post"; then
    rm -f "$post"
    echo "SHEPHERD IDENTITY BLOCK: request identity changed during validator turn" >&2
    return 1
  fi
  rm -f "$post"
}

capture_validator_outputs() {
  python3 - "$VERDICT_CANDIDATE" "$VERDICT_SNAPSHOT" \
    "$VALIDATION_CANDIDATE" "$VALIDATION_RECORD_SNAPSHOT" <<'PY'
import os
import stat
import sys

pairs = ((sys.argv[1], sys.argv[2], 10 * 1024 * 1024), (sys.argv[3], sys.argv[4], 10 * 1024 * 1024))
read_flags = os.O_RDONLY | getattr(os, "O_NOFOLLOW", 0)
create_flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL | getattr(os, "O_NOFOLLOW", 0)
for source, snapshot, maximum in pairs:
    source_fd = None
    try:
        before = os.lstat(source)
        if not stat.S_ISREG(before.st_mode) or before.st_nlink != 1:
            raise OSError("candidate is not a single-link regular file")
        if before.st_uid != os.geteuid() or before.st_size > maximum:
            raise OSError("candidate ownership or size is unsafe")
        source_fd = os.open(source, read_flags)
        opened = os.fstat(source_fd)
        if (opened.st_dev, opened.st_ino) != (before.st_dev, before.st_ino):
            raise OSError("candidate pathname changed before snapshot")
        snapshot_fd = os.open(snapshot, create_flags, 0o600)
        with os.fdopen(snapshot_fd, "wb", closefd=True) as output:
            while True:
                chunk = os.read(source_fd, 1024 * 1024)
                if not chunk:
                    break
                output.write(chunk)
            output.flush()
            os.fsync(output.fileno())
            os.fchmod(output.fileno(), 0o400)
        after = os.fstat(source_fd)
        if (after.st_dev, after.st_ino, after.st_size, after.st_mtime_ns) != (
            opened.st_dev, opened.st_ino, opened.st_size, opened.st_mtime_ns
        ):
            raise OSError("candidate changed while being snapshotted")
        os.close(source_fd)
        source_fd = None
        current = os.lstat(source)
        if (current.st_dev, current.st_ino) != (opened.st_dev, opened.st_ino):
            raise OSError("candidate pathname changed after snapshot")
        os.unlink(source)
    except OSError as exc:
        if source_fd is not None:
            os.close(source_fd)
        print(f"SHEPHERD OUTPUT BLOCK: {source}: {exc}", file=sys.stderr)
        raise SystemExit(1)
PY
}

verify_bound_validator_output() {
  python3 - "$REQUEST_JSON" "$VERDICT_SNAPSHOT" "$VALIDATION_RECORD_SNAPSHOT" "$CKPT_DIR" <<'PY'
import json
import os
import pathlib
import stat
import sys

request_path, verdict_path, validation_path, checkpoint_root = map(pathlib.Path, sys.argv[1:])
identity_fields = (
    "exact_base_sha",
    "exact_head_sha",
    "exact_head_tree",
    "candidate_lineage",
    "synthesis_sha256",
)


def fail(reason):
    print(f"SHEPHERD VERDICT BLOCK: {reason}", file=sys.stderr)
    raise SystemExit(1)


def read_secure(path, label, maximum=10 * 1024 * 1024):
    flags = os.O_RDONLY | getattr(os, "O_NOFOLLOW", 0)
    try:
        fd = os.open(path, flags)
        current = os.fstat(fd)
        if not stat.S_ISREG(current.st_mode) or current.st_nlink != 1:
            raise OSError("not a single-link regular file")
        if current.st_uid != os.geteuid() or current.st_size > maximum:
            raise OSError("unsafe ownership or size")
        with os.fdopen(fd, "rb", closefd=True) as source:
            return source.read()
    except OSError:
        fail(f"{label} is missing or unsafe")


def load(path, label):
    try:
        value = json.loads(read_secure(path, label))
    except (json.JSONDecodeError, UnicodeError):
        fail(f"{label} is malformed")
    if not isinstance(value, dict):
        fail(f"{label} must be an object")
    return value


request = load(request_path, "Shepherd request")
verdict = load(verdict_path, "validator verdict snapshot")
if verdict.get("schema_version") != "polymetrics.ai/shepherd-verdict/v1":
    fail("verdict schema mismatch")
for field in identity_fields:
    if verdict.get(field) != request.get(field):
        fail(f"verdict {field} mismatch")
value = verdict.get("verdict")
if value not in {"PROCEED", "RETRY", "REVERT", "HALT"}:
    fail("verdict value is invalid")
if not isinstance(verdict.get("reason"), str) or not verdict["reason"].strip():
    fail("verdict lacks cited reason")
if value in {"RETRY", "REVERT"} and (
    not isinstance(verdict.get("correction"), str) or not verdict["correction"].strip()
):
    fail("corrective verdict lacks correction")
if value == "REVERT":
    last_good = read_secure(checkpoint_root / "LAST_GOOD", "last-good checkpoint", 1024).decode().strip()
    if last_good != verdict.get("revert_to_checkpoint"):
        fail("REVERT checkpoint does not match driver last-good checkpoint")
    if last_good != "initial" and not last_good.isdigit():
        fail("REVERT checkpoint name is unsafe")
    checkpoint = checkpoint_root / last_good
    try:
        checkpoint_stat = os.lstat(checkpoint)
    except OSError:
        fail("REVERT checkpoint is missing")
    if stat.S_ISLNK(checkpoint_stat.st_mode) or not stat.S_ISDIR(checkpoint_stat.st_mode):
        fail("REVERT checkpoint is unsafe")
    read_secure(checkpoint / "HEAD.sha", "checkpoint HEAD identity", 1024)
elif verdict.get("revert_to_checkpoint") is not None:
    fail("non-REVERT verdict names a rollback checkpoint")

raw_entry = read_secure(validation_path, "validation record snapshot")
if not raw_entry.endswith(b"\n") or len(raw_entry.splitlines()) != 1 or not raw_entry.strip():
    fail("validator did not propose exactly one validation entry")
try:
    entry = json.loads(raw_entry)
except (json.JSONDecodeError, UnicodeError):
    fail("validation entry is malformed")
if not isinstance(entry, dict) or entry.get("schema_version") != "polymetrics.ai/shepherd-validation/v1":
    fail("validation entry schema mismatch")
for field in identity_fields:
    if entry.get(field) != request.get(field):
        fail(f"validation {field} mismatch")
if entry.get("verdict") != value:
    fail("validation/verdict decision mismatch")
PY
}

finalize_validation_ledger() { # optional $1=verified record snapshot; absent means restore only
  local record="${1:-}" finalize_rc=0
  python3 - "$LEDGER_BASE_SNAPSHOT" "$LEDGER_BASE_PRESENT" "$LEDGER_BASE_HASH" \
    "$VALIDATION_JSONL" "$VALIDATION_JSONL.lock" "${record:--}" <<'PY'
import fcntl
import hashlib
import json
import os
import secrets
import stat
import sys

base_path, base_present_raw, base_hash, ledger, lock_path, record_path = sys.argv[1:]
base_present = base_present_raw == "1"
no_follow = getattr(os, "O_NOFOLLOW", 0)


def fail(reason):
    print(f"SHEPHERD LEDGER BLOCK: {reason}", file=sys.stderr)
    raise SystemExit(1)


def read_single_link(path, label, maximum=100 * 1024 * 1024):
    try:
        fd = os.open(path, os.O_RDONLY | no_follow)
        current = os.fstat(fd)
        if not stat.S_ISREG(current.st_mode) or current.st_nlink != 1:
            raise OSError("not a single-link regular file")
        if current.st_uid != os.geteuid() or current.st_size > maximum:
            raise OSError("unsafe ownership or size")
        with os.fdopen(fd, "rb", closefd=True) as source:
            return source.read(), current
    except OSError as exc:
        fail(f"{label} is unsafe: {exc}")


def write_temporary(data):
    parent = os.path.dirname(ledger)
    for _ in range(20):
        temporary = os.path.join(parent, f".VALIDATION.restore.{secrets.token_hex(12)}")
        try:
            fd = os.open(temporary, os.O_WRONLY | os.O_CREAT | os.O_EXCL | no_follow, 0o600)
            with os.fdopen(fd, "wb", closefd=True) as output:
                output.write(data)
                output.flush()
                os.fsync(output.fileno())
            return temporary
        except FileExistsError:
            continue
    fail("could not allocate exclusive ledger temporary")


def restore_base():
    try:
        current = os.lstat(ledger)
    except FileNotFoundError:
        current = None
    if current is not None:
        if stat.S_ISDIR(current.st_mode):
            quarantine = os.path.join(
                os.path.dirname(ledger), f".VALIDATION.quarantine.{secrets.token_hex(12)}"
            )
            os.rename(ledger, quarantine)
        else:
            os.unlink(ledger)
    if base_present:
        temporary = write_temporary(base)
        os.replace(temporary, ledger)
    directory = os.open(os.path.dirname(ledger), os.O_RDONLY | getattr(os, "O_DIRECTORY", 0))
    try:
        os.fsync(directory)
    finally:
        os.close(directory)


base, _ = read_single_link(base_path, "driver ledger snapshot")
if hashlib.sha256(base).hexdigest() != base_hash:
    fail("driver ledger snapshot hash changed")
try:
    lock_fd = os.open(lock_path, os.O_RDWR | os.O_CREAT | no_follow, 0o600)
    lock_stat = os.fstat(lock_fd)
    if not stat.S_ISREG(lock_stat.st_mode) or lock_stat.st_nlink != 1:
        raise OSError("lock is not a single-link regular file")
    if lock_stat.st_uid != os.geteuid() or stat.S_IMODE(lock_stat.st_mode) & 0o077:
        raise OSError("lock is not private and driver-owned")
except OSError as exc:
    # The validator session is already terminated. Restore the exact pre-run bytes even when it
    # attacked the lock pathname, then reject the turn rather than leaving truncated history.
    restore_base()
    fail(f"cannot acquire safe ledger lock: {exc}")

with os.fdopen(lock_fd, "r+b", closefd=True) as lock_file:
    fcntl.flock(lock_file.fileno(), fcntl.LOCK_EX)
    try:
        try:
            unsafe = os.lstat(ledger)
        except FileNotFoundError:
            current = b""
            current_stat = None
            current_present = False
        else:
            current_present = True
            try:
                current, current_stat = read_single_link(ledger, "validation ledger")
            except SystemExit:
                # A symlink/hardlink can be unlinked without touching its target, then the exact
                # driver snapshot is restored. Directories remain a hard failure in restore_base.
                if stat.S_ISDIR(unsafe.st_mode):
                    raise
                current = None
                current_stat = None
        expected_matches = current is not None and current_present == base_present and current == base
        if not expected_matches:
            restore_base()
            if record_path != "-":
                fail("validator changed pre-run ledger bytes; exact snapshot restored")
            raise SystemExit(0)
        if record_path == "-":
            raise SystemExit(0)
        if base and not base.endswith(b"\n"):
            fail("pre-run validation ledger is not appendable JSONL")
        raw_record, _ = read_single_link(record_path, "verified validation record", 10 * 1024 * 1024)
        try:
            record = json.loads(raw_record)
        except (json.JSONDecodeError, UnicodeError):
            fail("verified validation record became malformed")
        appended = json.dumps(record, sort_keys=True, separators=(",", ":")).encode() + b"\n"
        if base_present:
            fd = os.open(ledger, os.O_RDWR | os.O_APPEND | no_follow)
            opened = os.fstat(fd)
            if (
                not stat.S_ISREG(opened.st_mode)
                or opened.st_nlink != 1
                or opened.st_uid != os.geteuid()
                or current_stat is None
                or (opened.st_dev, opened.st_ino, opened.st_size, opened.st_mtime_ns)
                != (current_stat.st_dev, current_stat.st_ino, current_stat.st_size, current_stat.st_mtime_ns)
            ):
                os.close(fd)
                fail("validation ledger changed before driver append")
        else:
            fd = os.open(ledger, os.O_RDWR | os.O_CREAT | os.O_EXCL | no_follow, 0o600)
        with os.fdopen(fd, "r+b", closefd=True) as output:
            output.write(appended)
            output.flush()
            os.fsync(output.fileno())
            output.seek(0)
            final = output.read()
        if final != base + appended:
            restore_base()
            fail("validation ledger changed during driver append; exact snapshot restored")
    finally:
        fcntl.flock(lock_file.fileno(), fcntl.LOCK_UN)
PY
  finalize_rc=$?
  if (( finalize_rc == 0 )); then
    LEDGER_FINALIZED=1
  fi
  return "$finalize_rc"
}

latest_session_file() { # $1=exact launched session root $2=minimum epoch
  local root="$1" minimum_epoch="${2:-0}"
  python3 - "$root" "$minimum_epoch" <<'PY'
import datetime as dt
import json
import os
import stat
import sys

root = sys.argv[1]
minimum = float(sys.argv[2])
best = None


def parse_ts(value):
    if isinstance(value, (int, float)):
        return float(value) / (1000 if value > 10_000_000_000 else 1)
    if not isinstance(value, str) or not value:
        return None
    try:
        return dt.datetime.fromisoformat(value.replace("Z", "+00:00")).timestamp()
    except ValueError:
        return None


def last_event_epoch(path):
    last = None
    flags = os.O_RDONLY | getattr(os, "O_NOFOLLOW", 0)
    try:
        fd = os.open(path, flags)
        current = os.fstat(fd)
        if not stat.S_ISREG(current.st_mode) or current.st_nlink != 1:
            return None
        with os.fdopen(fd, errors="replace") as source:
            for line in source:
                if line.strip():
                    last = line
        if last is None:
            return None
        obj = json.loads(last)
    except Exception:
        return None
    candidates = [
        obj.get("timestamp"),
        obj.get("message", {}).get("timestamp") if isinstance(obj.get("message"), dict) else None,
    ]
    for candidate in candidates:
        parsed = parse_ts(candidate)
        if parsed is not None:
            return parsed
    return None


if os.path.isdir(root) and not os.path.islink(root):
    for current_root, subdirs, names in os.walk(root, followlinks=False):
        subdirs[:] = [name for name in subdirs if name not in {".git", "node_modules", "vendor"}]
        for name in names:
            if not name.endswith(".jsonl"):
                continue
            path = os.path.join(current_root, name)
            try:
                event = last_event_epoch(path)
                item = (event if event is not None else os.lstat(path).st_mtime, path)
            except OSError:
                continue
            if item[0] < minimum:
                continue
            if best is None or item > best:
                best = item
if best:
    print(best[1])
PY
}

session_age_seconds() { # $1=session-file; FUTURE is an explicit fail-closed result
  local sess="$1"
  python3 - "$sess" <<'PY' 2>/dev/null || true
import datetime as dt
import json
import os
import stat
import sys
import time

path = sys.argv[1]
now = int(time.time())


def parse_ts(value):
    if isinstance(value, (int, float)):
        return int(value / (1000 if value > 10_000_000_000 else 1))
    if not isinstance(value, str) or not value:
        return None
    try:
        return int(dt.datetime.fromisoformat(value.replace("Z", "+00:00")).timestamp())
    except ValueError:
        return None


try:
    fd = os.open(path, os.O_RDONLY | getattr(os, "O_NOFOLLOW", 0))
    current = os.fstat(fd)
    if not stat.S_ISREG(current.st_mode) or current.st_nlink != 1:
        raise OSError("unsafe session file")
    last = None
    with os.fdopen(fd, errors="replace") as source:
        for line in source:
            if line.strip():
                last = line
    event = None
    if last:
        obj = json.loads(last)
        for candidate in (
            obj.get("timestamp"),
            obj.get("message", {}).get("timestamp") if isinstance(obj.get("message"), dict) else None,
        ):
            event = parse_ts(candidate)
            if event is not None:
                break
    observed = event if event is not None else int(current.st_mtime)
    if observed > now + 5:
        print("FUTURE")
    else:
        print(max(0, now - observed))
except Exception:
    pass
PY
}

live_child_count() { # $1=pid
  { pgrep -P "$1" 2>/dev/null || true; } | wc -l | tr -d ' '
}

kill_process_tree() { # $1=TERM|KILL $2=root-pid
  local sig="$1" root="$2" child
  for child in $(pgrep -P "$root" 2>/dev/null || true); do
    kill_process_tree "$sig" "$child"
  done
  kill "-$sig" "$root" 2>/dev/null || true
}

process_matches_session() { # $1=pid $2=session-id
  python3 - "$1" "$2" <<'PY'
import os
import sys
pid, expected = map(int, sys.argv[1:])
try:
    raise SystemExit(0 if os.getsid(pid) == expected else 1)
except ProcessLookupError:
    raise SystemExit(2)
PY
}

session_live_pids() { # $1=session-id
  python3 - "$1" <<'PY'
import os
import subprocess
import sys

session = int(sys.argv[1])
rows = subprocess.run(
    ["ps", "-axo", "pid=,stat="], capture_output=True, text=True, check=False
).stdout.splitlines()
for row in rows:
    fields = row.split(None, 1)
    if not fields or not fields[0].isdigit():
        continue
    pid = int(fields[0])
    state = fields[1] if len(fields) > 1 else ""
    if state.startswith("Z"):
        continue
    try:
        if os.getsid(pid) == session:
            print(pid)
    except (ProcessLookupError, PermissionError):
        continue
PY
}

signal_process_session() { # $1=session-id $2=TERM|KILL
  local sid="$1" sig="$2" member
  while read -r member; do
    [[ "$member" =~ ^[1-9][0-9]*$ ]] || continue
    kill "-$sig" "$member" 2>/dev/null || true
  done < <(session_live_pids "$sid")
}

terminate_validator_session() { # $1=exact validator session-id
  local sid="$1" attempt live
  [[ "$sid" =~ ^[1-9][0-9]*$ ]] || return 1
  signal_process_session "$sid" TERM
  for (( attempt=0; attempt<20; attempt++ )); do
    live="$(session_live_pids "$sid")"
    [[ -z "$live" ]] && break
    sleep 0.05
  done
  for (( attempt=0; attempt<40; attempt++ )); do
    live="$(session_live_pids "$sid")"
    [[ -z "$live" ]] && return 0
    signal_process_session "$sid" KILL
    sleep 0.05
  done
  log "VALIDATOR WATCHDOG: session $sid still has live members after KILL"
  return 1
}

cleanup_active_validator() {
  local pid="$ACTIVE_VALIDATOR_PID" sid="$ACTIVE_VALIDATOR_SID" cleanup_rc=0
  [[ -n "$pid" ]] || return 0
  if [[ "$sid" =~ ^[1-9][0-9]*$ ]]; then
    terminate_validator_session "$sid" || cleanup_rc=1
  else
    # Before the trusted launcher's handshake, untrusted validator code has not been exec'd. Kill
    # the launcher tree and verify its exact PID instead of trusting unconfirmed session metadata.
    kill_process_tree TERM "$pid"
    sleep 0.1
    kill_process_tree KILL "$pid"
  fi
  wait "$pid" 2>/dev/null || true
  if kill -0 "$pid" 2>/dev/null; then
    cleanup_rc=1
  fi
  if (( cleanup_rc == 0 )); then
    ACTIVE_VALIDATOR_PID=""
    ACTIVE_VALIDATOR_SID=""
  fi
  return "$cleanup_rc"
}

# shellcheck disable=SC2329 # invoked by the EXIT trap
on_driver_exit() {
  local rc=$?
  set +e
  cleanup_active_validator
  if (( LEDGER_FINALIZED == 0 )) && [[ -n "$LEDGER_BASE_SNAPSHOT" ]]; then
    finalize_validation_ledger || true
  fi
  exit "$rc"
}
trap on_driver_exit EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

stale_session_requires_kill() { # $1=session-file $2=pid
  local sess="$1" pid="$2" age children
  age="$(session_age_seconds "$sess")"
  [[ -n "$age" ]] || return 1
  if [[ "$age" != "FUTURE" ]]; then
    [[ "$age" =~ ^[0-9]+$ ]] || return 1
    (( age > STALL_MINUTES * 60 )) || return 1
  fi
  children="$(live_child_count "$pid")"
  if (( children > 0 )) && [[ "$STALL_KILL_LIVE_CHILDREN" != "1" ]]; then
    return 1
  fi
  return 0
}

kill_stale_turn() { # $1=session-file $2=pid
  local sess="$1" pid="$2" age children
  age="$(session_age_seconds "$sess")"
  children="$(live_child_count "$pid")"
  log "STALL GUARD: unsafe session age ${age:-unknown}; live_children=${children:-0}; killing turn pid $pid"
  kill_process_tree TERM "$pid"; sleep 5; kill_process_tree KILL "$pid"
  wait "$pid" 2>/dev/null || true
}

run_validator_with_watchdog() {
  local rc=0 pid sess request started now handshake_pid handshake_sid handshake_pgid prompt fifo identity_rc
  local -a validator_argv=()
  prepare_validator_turn || return 65
  request="$(cat "$REQUEST_JSON" 2>/dev/null || printf '{}')"
  started="$(date +%s)"
  read -r -a validator_argv <<< "$VALIDATOR_ARGS"
  prompt="$(state_root_instruction)

SHEPHERD EXACT IDENTITY REQUEST (driver-owned; echo every field verbatim):
$request

$(cat "$VAL_PROMPT")

DRIVER OUTPUT OVERRIDE (final and driver-enforced):
- Read trajectory state and canonical validation history from $STATE_DIR.
- Write the one proposed verdict only to $VERDICT_CANDIDATE.
- Write exactly one new proposed validation JSON line only to $VALIDATION_CANDIDATE.
- Do not write or truncate canonical $VERDICT_JSON or $VALIDATION_JSONL. The driver consumes each
  candidate once, verifies immutable snapshots, and appends the accepted record under its lock."
  mkdir "$VALIDATOR_TURN_DIR/sessions"
  fifo="$VALIDATOR_TURN_DIR/.session-handshake"
  mkfifo -m 600 "$fifo"
  exec 9<> "$fifo"
  rm -f "$fifo"
  SHEPHERD_STATE_ROOT="$STATE_DIR" \
  SHEPHERD_VALIDATOR_OUTPUT_DIR="$VALIDATOR_TURN_DIR" \
  AUTO_LOOP_STATE_DIR="$VALIDATOR_TURN_DIR" \
    python3 -c '
import os, sys
fd = int(sys.argv[1])
command = sys.argv[2:]
os.setsid()
os.write(fd, f"{os.getpid()} {os.getsid(0)} {os.getpgrp()}\n".encode())
os.close(fd)
os.execvp(command[0], command)
' 9 "$VALIDATOR_BIN" -p "${validator_argv[@]}" --session-dir "$VALIDATOR_TURN_DIR/sessions" \
      "$prompt" >>"$LOG_FILE" 2>&1 & pid=$!
  ACTIVE_VALIDATOR_PID="$pid"
  if ! IFS=' ' read -r -t 5 handshake_pid handshake_sid handshake_pgid <&9; then
    exec 9>&-
    log "VALIDATOR WATCHDOG: validator failed dedicated-session handshake"
    cleanup_active_validator
    finalize_validation_ledger || true
    secure_remove_validator_paths || true
    return 125
  fi
  exec 9>&-
  if [[ "$handshake_pid" != "$pid" || "$handshake_sid" != "$pid" || "$handshake_pgid" != "$pid" ]]; then
    log "VALIDATOR WATCHDOG: validator did not enter its exact dedicated session/process group"
    ACTIVE_VALIDATOR_SID="$pid"
    cleanup_active_validator || true
    finalize_validation_ledger || true
    secure_remove_validator_paths || true
    return 125
  fi
  ACTIVE_VALIDATOR_SID="$handshake_sid"
  while kill -0 "$pid" 2>/dev/null; do
    sleep "$WATCHDOG_POLL_SECONDS"
    kill -0 "$pid" 2>/dev/null || break
    now="$(date +%s)"
    identity_rc=0
    process_matches_session "$pid" "$handshake_sid" || identity_rc=$?
    (( identity_rc == 2 )) && break
    if (( identity_rc != 0 )); then
      log "VALIDATOR WATCHDOG: launched validator PID escaped its bound session"
      cleanup_active_validator || true
      finalize_validation_ledger || true
      secure_remove_validator_paths || true
      return 125
    fi
    if (( now - started >= VALIDATOR_TIMEOUT_SECONDS )); then
      log "VALIDATOR WATCHDOG: hard timeout ${VALIDATOR_TIMEOUT_SECONDS}s; killing validator session $handshake_sid"
      if ! cleanup_active_validator; then
        finalize_validation_ledger || true
        secure_remove_validator_paths || true
        return 125
      fi
      finalize_validation_ledger || true
      secure_remove_validator_paths || true
      return 124
    fi
    sess="$(latest_session_file "$VALIDATOR_TURN_DIR/sessions" "$started")"
    if [[ -n "$sess" ]] && stale_session_requires_kill "$sess" "$pid"; then
      log "VALIDATOR WATCHDOG: stale or future-dated bound session; killing validator session $handshake_sid"
      if ! cleanup_active_validator; then
        finalize_validation_ledger || true
        secure_remove_validator_paths || true
        return 125
      fi
      finalize_validation_ledger || true
      secure_remove_validator_paths || true
      return 124
    fi
  done
  wait "$pid" 2>/dev/null || rc=$?
  if ! cleanup_active_validator; then
    rc=125
  fi
  if (( rc != 0 )); then
    finalize_validation_ledger || true
    secure_remove_validator_paths || true
  fi
  return "$rc"
}

if [[ "${SHEPHERD_FUTURE_SESSION_SELF_TEST:-}" == "1" ]]; then
  test_sess="${SHEPHERD_TEST_SESSION_FILE:?}"
  if ! stale_session_requires_kill "$test_sess" "$$"; then
    echo "self-test failed: future session timestamp was accepted" >&2
    exit 1
  fi
  echo "self-test passed: future session timestamp is rejected"
  exit 0
fi

if [[ "${SHEPHERD_STALL_GUARD_SELF_TEST:-}" == "1" ]]; then
  mkdir -p "$STATE_DIR/sessions"
  test_sess="$STATE_DIR/sessions/stale-live-child-test.jsonl"
  printf '{"type":"message","timestamp":"2000-01-01T00:00:00Z","message":{"role":"assistant","content":"stale"}}\n' > "$test_sess"
  touch "$test_sess"
  (trap 'exit 0' TERM; sleep 300 & wait) & test_pid=$!
  sleep 1
  if ! stale_session_requires_kill "$test_sess" "$test_pid"; then
    echo "self-test failed: stale live-child turn was not killable" >&2
    kill_process_tree KILL "$test_pid"
    exit 1
  fi
  kill_stale_turn "$test_sess" "$test_pid"
  wait "$test_pid" 2>/dev/null || true
  sleep 1
  if kill -0 "$test_pid" 2>/dev/null; then
    echo "self-test failed: stale live-child turn survived kill" >&2
    kill_process_tree KILL "$test_pid"
    exit 1
  fi
  echo "self-test passed: stale live-child turn is killed"
  exit 0
fi

if [[ "${SHEPHERD_VALIDATOR_WATCHDOG_SELF_TEST:-}" == "1" ]]; then
  validator_rc=0
  run_validator_with_watchdog || validator_rc=$?
  if (( validator_rc != 124 )); then
    echo "self-test failed: hanging validator returned $validator_rc, want watchdog status 124" >&2
    exit 1
  fi
  [[ ! -e "$VERDICT_JSON" ]] || {
    echo "self-test failed: timed-out validator verdict survived" >&2
    exit 1
  }
  echo "self-test passed: validator watchdog killed the process tree and discarded its verdict"
  exit 0
fi

if [[ "${SHEPHERD_VALIDATOR_SESSION_SELF_TEST:-}" == "1" ]]; then
  validator_rc=0
  run_validator_with_watchdog || validator_rc=$?
  if (( validator_rc != 0 )); then
    echo "self-test failed: zero-exit validator returned $validator_rc" >&2
    exit 1
  fi
  echo "self-test passed: zero-exit validator session descendants were terminated"
  exit 0
fi

# --- preflight: the subagent tool must be available (vendored extension OR installed package) ---
# We vendor pi-sub-agent under .pi/extensions/ (records child sessions via PI_SUBAGENT_SESSION_DIR),
# loaded through .pi/settings.json. Accept either the vendored extension or the npm package; fail
# only if neither is present (subagent tool silently absent → .pi/agents/* cannot be spawned).
if [[ ! -f "$REPO_ROOT/.pi/extensions/pi-sub-agent/index.ts" ]] \
   && ! "$PI_BIN" list 2>/dev/null | grep -q "pi-sub-agent"; then
  echo "FATAL: the pi 'subagent' tool is unavailable — no vendored .pi/extensions/pi-sub-agent and" >&2
  echo "no installed package, so .pi/agents/* cannot be spawned. Restore the vendored extension or" >&2
  echo "run:  $PI_BIN install npm:pi-sub-agent" >&2
  exit 2
fi

# --- resolve the problem prompt --------------------------------------------------------------
if [[ "${SHEPHERD_ORCHESTRATOR_WATCHDOG_SELF_TEST:-}" == "1" ]]; then
  PROBLEM="orchestrator watchdog self-test"
elif [[ "${1:-}" == "--resume" ]]; then
  [[ -f "$PROMPT_FILE" ]] || { echo "No run to resume (missing $PROMPT_FILE)." >&2; exit 2; }
  PROBLEM="$(cat "$PROMPT_FILE")"; log "RESUME: $PROBLEM"
elif [[ -n "${1:-}" ]]; then
  PROBLEM="$*"; printf '%s' "$PROBLEM" > "$PROMPT_FILE"; log "START: $PROBLEM"
else
  echo "Usage: scripts/pi-shepherd-loop.sh \"<problem prompt>\" | --resume" >&2; exit 2
fi

new_orchestrator_session_root() {
  python3 - "$STATE_DIR/sessions" <<'PY'
import os
import pathlib
import stat
import sys
import tempfile

root = pathlib.Path(sys.argv[1])
try:
    try:
        current = os.lstat(root)
    except FileNotFoundError:
        os.mkdir(root, 0o700)
        current = os.lstat(root)
    if stat.S_ISLNK(current.st_mode) or not stat.S_ISDIR(current.st_mode):
        raise OSError("session root is not a real directory")
    if current.st_uid != os.geteuid() or stat.S_IMODE(current.st_mode) & 0o077:
        raise OSError("session root is not private and driver-owned")
    print(tempfile.mkdtemp(prefix="orchestrator.", dir=root))
except OSError as exc:
    print(f"TURN WATCHDOG: unsafe session root: {exc}", file=sys.stderr)
    raise SystemExit(1)
PY
}

run_orchestrator() { # $1=turn-message — exact session root + unconditional turn deadline
  local msg="$1" rc=0 pid sess started now session_root identity_rc
  local -a extra_argv=()
  session_root="$(new_orchestrator_session_root)" || return 125
  read -r -a extra_argv <<< "$PI_EXTRA_FLAGS"
  started="$(date +%s)"
  python3 -c '
import os, sys
command = sys.argv[1:]
os.setsid()
os.execvp(command[0], command)
' "$PI_BIN" -p --model "$ORCH_MODEL" --thinking "$ORCH_THINKING" --tools "$PI_TOOLS" \
    --approve --session-dir "$session_root" "${extra_argv[@]}" "$msg" >>"$LOG_FILE" 2>&1 & pid=$!
  while kill -0 "$pid" 2>/dev/null; do
    sleep "$WATCHDOG_POLL_SECONDS"
    kill -0 "$pid" 2>/dev/null || break
    now="$(date +%s)"
    identity_rc=0
    process_matches_session "$pid" "$pid" || identity_rc=$?
    (( identity_rc == 2 )) && break
    if (( identity_rc != 0 )); then
      log "TURN WATCHDOG: launched orchestrator PID escaped its exact session"
      terminate_validator_session "$pid" || true
      wait "$pid" 2>/dev/null || true
      return 125
    fi
    if (( now - started >= TURN_TIMEOUT_SECONDS )); then
      log "TURN WATCHDOG: unconditional ${TURN_TIMEOUT_SECONDS}s deadline; killing turn session $pid"
      terminate_validator_session "$pid" || true
      wait "$pid" 2>/dev/null || true
      return 124
    fi
    sess="$(latest_session_file "$session_root" "$started")"
    if [[ -n "$sess" ]] && stale_session_requires_kill "$sess" "$pid"; then
      log "TURN WATCHDOG: stale or future-dated event in exact launched session; killing $pid"
      terminate_validator_session "$pid" || true
      wait "$pid" 2>/dev/null || true
      return 124
    fi
  done
  wait "$pid" 2>/dev/null || rc=$?
  return "$rc"
}

if [[ "${SHEPHERD_ORCHESTRATOR_WATCHDOG_SELF_TEST:-}" == "1" ]]; then
  orchestrator_rc=0
  run_orchestrator "$PROBLEM" || orchestrator_rc=$?
  if (( orchestrator_rc != 124 )); then
    echo "self-test failed: hanging orchestrator returned $orchestrator_rc, want deadline status 124" >&2
    exit 1
  fi
  echo "self-test passed: unconditional orchestrator turn deadline killed exact session"
  exit 0
fi

checkpoint() { # $1=turn — exclusive private snapshot of RUN presence + verified HEAD.
  local turn="$1" sha
  [[ "$turn" == "initial" || "$turn" =~ ^[0-9]+$ ]] || return 1
  sha="$(git -C "$REPO_ROOT" rev-parse HEAD 2>/dev/null)" || return 1
  [[ "$sha" =~ ^[0-9a-f]{40}$ ]] || return 1
  git -C "$REPO_ROOT" cat-file -e "$sha^{commit}" 2>/dev/null || return 1
  python3 - "$CKPT_DIR" "$RUN_JSON" "$turn" "$sha" <<'PY'
import os
import secrets
import stat
import sys

root, run_path, turn, head = sys.argv[1:]
no_follow = getattr(os, "O_NOFOLLOW", 0)
directory = getattr(os, "O_DIRECTORY", 0)
root_fd = directory_fd = source_fd = None
created = False


def fail(reason):
    print(f"CHECKPOINT BLOCK: {reason}", file=sys.stderr)
    raise SystemExit(1)


def write_exclusive(parent_fd, name, data):
    fd = os.open(
        name,
        os.O_WRONLY | os.O_CREAT | os.O_EXCL | no_follow,
        0o600,
        dir_fd=parent_fd,
    )
    with os.fdopen(fd, "wb", closefd=True) as output:
        output.write(data)
        output.flush()
        os.fsync(output.fileno())


try:
    root_fd = os.open(root, os.O_RDONLY | directory | no_follow)
    root_stat = os.fstat(root_fd)
    if root_stat.st_uid != os.geteuid() or stat.S_IMODE(root_stat.st_mode) & 0o077:
        fail("checkpoint root is not private and driver-owned")
    try:
        os.stat(turn, dir_fd=root_fd, follow_symlinks=False)
    except FileNotFoundError:
        pass
    else:
        fail("checkpoint destination already exists")
    os.mkdir(turn, 0o700, dir_fd=root_fd)
    created = True
    directory_fd = os.open(turn, os.O_RDONLY | directory | no_follow, dir_fd=root_fd)
    try:
        source_fd = os.open(run_path, os.O_RDONLY | no_follow)
    except FileNotFoundError:
        write_exclusive(directory_fd, "RUN.absent", b"")
    else:
        source_stat = os.fstat(source_fd)
        if not stat.S_ISREG(source_stat.st_mode) or source_stat.st_nlink != 1:
            fail("RUN.json is not a single-link regular file")
        if source_stat.st_uid != os.geteuid():
            fail("RUN.json is not driver-owned")
        with os.fdopen(source_fd, "rb", closefd=True) as source:
            source_fd = None
            write_exclusive(directory_fd, "RUN.json", source.read())
    write_exclusive(directory_fd, "HEAD.sha", (head + "\n").encode())
    os.fsync(directory_fd)

    try:
        marker = os.stat("LAST_GOOD", dir_fd=root_fd, follow_symlinks=False)
    except FileNotFoundError:
        marker = None
    if marker is not None and (
        not stat.S_ISREG(marker.st_mode)
        or marker.st_nlink != 1
        or marker.st_uid != os.geteuid()
        or stat.S_IMODE(marker.st_mode) & 0o077
    ):
        fail("preexisting LAST_GOOD is unsafe")
    temporary = f".LAST_GOOD.{secrets.token_hex(12)}"
    write_exclusive(root_fd, temporary, (turn + "\n").encode())
    os.replace(temporary, "LAST_GOOD", src_dir_fd=root_fd, dst_dir_fd=root_fd)
    os.fsync(root_fd)
    created = False
except OSError as exc:
    fail(str(exc))
finally:
    if source_fd is not None:
        os.close(source_fd)
    if directory_fd is not None:
        os.close(directory_fd)
    if root_fd is not None:
        if created:
            # Partial exclusive checkpoints remain a fail-closed marker if cleanup cannot complete.
            for name in ("RUN.json", "RUN.absent", "HEAD.sha"):
                try:
                    os.unlink(f"{turn}/{name}", dir_fd=root_fd)
                except OSError:
                    pass
            try:
                os.rmdir(turn, dir_fd=root_fd)
            except OSError:
                pass
        os.close(root_fd)
PY
}

verified_checkpoint_marker() { # 0=safe marker, 3=missing marker, other=unsafe
  python3 - "$CKPT_DIR" <<'PY'
import os
import re
import stat
import sys

root = sys.argv[1]
no_follow = getattr(os, "O_NOFOLLOW", 0)
directory = getattr(os, "O_DIRECTORY", 0)
try:
    root_fd = os.open(root, os.O_RDONLY | directory | no_follow)
    try:
        fd = os.open("LAST_GOOD", os.O_RDONLY | no_follow, dir_fd=root_fd)
    except FileNotFoundError:
        raise SystemExit(3)
    marker_stat = os.fstat(fd)
    if (
        not stat.S_ISREG(marker_stat.st_mode)
        or marker_stat.st_nlink != 1
        or marker_stat.st_uid != os.geteuid()
        or stat.S_IMODE(marker_stat.st_mode) & 0o077
    ):
        raise OSError("unsafe LAST_GOOD")
    with os.fdopen(fd) as source:
        turn = source.read().strip()
    if turn != "initial" and not turn.isdigit():
        raise OSError("unsafe checkpoint name")
    checkpoint_stat = os.stat(turn, dir_fd=root_fd, follow_symlinks=False)
    if (
        not stat.S_ISDIR(checkpoint_stat.st_mode)
        or checkpoint_stat.st_uid != os.geteuid()
        or stat.S_IMODE(checkpoint_stat.st_mode) & 0o077
    ):
        raise OSError("unsafe checkpoint directory")
    checkpoint_fd = os.open(turn, os.O_RDONLY | directory | no_follow, dir_fd=root_fd)
    head_fd = os.open("HEAD.sha", os.O_RDONLY | no_follow, dir_fd=checkpoint_fd)
    head_stat = os.fstat(head_fd)
    if (
        not stat.S_ISREG(head_stat.st_mode)
        or head_stat.st_nlink != 1
        or head_stat.st_uid != os.geteuid()
        or stat.S_IMODE(head_stat.st_mode) & 0o077
    ):
        raise OSError("unsafe checkpoint HEAD")
    with os.fdopen(head_fd) as source:
        head = source.read().strip()
    if not re.fullmatch(r"[0-9a-f]{40}", head):
        raise OSError("malformed checkpoint HEAD")
    run_present = False
    absent_present = False
    for name in ("RUN.json", "RUN.absent"):
        try:
            item = os.stat(name, dir_fd=checkpoint_fd, follow_symlinks=False)
        except FileNotFoundError:
            continue
        if (
            not stat.S_ISREG(item.st_mode)
            or item.st_nlink != 1
            or item.st_uid != os.geteuid()
            or stat.S_IMODE(item.st_mode) & 0o077
        ):
            raise OSError("unsafe checkpoint RUN state")
        if name == "RUN.json":
            run_present = True
        else:
            absent_present = True
    if run_present == absent_present:
        raise OSError("ambiguous checkpoint RUN state")
finally:
    for candidate in (locals().get("checkpoint_fd"), locals().get("root_fd")):
        if isinstance(candidate, int):
            os.close(candidate)
PY
}

ensure_initial_checkpoint() {
  local marker_rc=0
  verified_checkpoint_marker || marker_rc=$?
  case "$marker_rc" in
    0) return 0 ;;
    3)
      checkpoint initial || {
        log "HALT: could not create verified initial checkpoint"
        return 1
      }
      log "initial checkpoint created before first turn"
      ;;
    *)
      log "HALT: preexisting checkpoint authority is unsafe"
      return 1
      ;;
  esac
}

next_checkpoint_number() {
  python3 - "$CKPT_DIR" <<'PY'
import os
import stat
import sys

root = sys.argv[1]
root_fd = os.open(
    root,
    os.O_RDONLY | getattr(os, "O_DIRECTORY", 0) | getattr(os, "O_NOFOLLOW", 0),
)
maximum = 0
try:
    for name in os.listdir(root_fd):
        if not name.isdigit():
            continue
        item = os.stat(name, dir_fd=root_fd, follow_symlinks=False)
        if (
            not stat.S_ISDIR(item.st_mode)
            or item.st_uid != os.geteuid()
            or stat.S_IMODE(item.st_mode) & 0o077
        ):
            raise OSError(f"unsafe preexisting checkpoint destination: {name}")
        maximum = max(maximum, int(name))
    print(maximum + 1)
finally:
    os.close(root_fd)
PY
}

restore_checkpoint() { # Restore verified ledger presence; never rewrite git history or merge here.
  local identity last good_sha cur_sha
  verified_checkpoint_marker || {
    log "HALT: missing or unsafe rollback checkpoint"
    return 1
  }
  identity="$(python3 - "$CKPT_DIR" <<'PY'
import os
import stat
import sys
root = sys.argv[1]
no_follow = getattr(os, "O_NOFOLLOW", 0)
directory = getattr(os, "O_DIRECTORY", 0)
root_fd = os.open(root, os.O_RDONLY | directory | no_follow)
marker_fd = os.open("LAST_GOOD", os.O_RDONLY | no_follow, dir_fd=root_fd)
with os.fdopen(marker_fd) as source:
    turn = source.read().strip()
checkpoint_fd = os.open(turn, os.O_RDONLY | directory | no_follow, dir_fd=root_fd)
head_fd = os.open("HEAD.sha", os.O_RDONLY | no_follow, dir_fd=checkpoint_fd)
with os.fdopen(head_fd) as source:
    head = source.read().strip()
print(turn, head)
os.close(checkpoint_fd)
os.close(root_fd)
PY
)" || return 1
  read -r last good_sha <<< "$identity"
  cur_sha="$(git -C "$REPO_ROOT" rev-parse HEAD 2>/dev/null || true)"
  [[ "$good_sha" =~ ^[0-9a-f]{40}$ && "$cur_sha" =~ ^[0-9a-f]{40}$ ]] || {
    log "HALT: rollback checkpoint has invalid git identity"
    return 1
  }
  git -C "$REPO_ROOT" cat-file -e "$good_sha^{commit}" 2>/dev/null || {
    log "HALT: rollback checkpoint commit is unavailable"
    return 1
  }
  python3 - "$CKPT_DIR" "$last" "$RUN_JSON" <<'PY' || {
import os
import secrets
import stat
import sys

root, turn, destination = sys.argv[1:]
no_follow = getattr(os, "O_NOFOLLOW", 0)
directory = getattr(os, "O_DIRECTORY", 0)
root_fd = os.open(root, os.O_RDONLY | directory | no_follow)
checkpoint_fd = os.open(turn, os.O_RDONLY | directory | no_follow, dir_fd=root_fd)


def safe_item(name):
    try:
        current = os.stat(name, dir_fd=checkpoint_fd, follow_symlinks=False)
    except FileNotFoundError:
        return None
    if not stat.S_ISREG(current.st_mode) or current.st_nlink != 1 or current.st_uid != os.geteuid():
        raise OSError(f"unsafe checkpoint {name}")
    return current


run_item = safe_item("RUN.json")
absent_item = safe_item("RUN.absent")
if (run_item is None) == (absent_item is None):
    raise OSError("ambiguous checkpoint RUN state")
try:
    destination_item = os.lstat(destination)
except FileNotFoundError:
    destination_item = None
if destination_item is not None:
    if stat.S_ISDIR(destination_item.st_mode):
        raise OSError("RUN destination is a directory")
    os.unlink(destination)
if run_item is not None:
    source_fd = os.open("RUN.json", os.O_RDONLY | no_follow, dir_fd=checkpoint_fd)
    with os.fdopen(source_fd, "rb", closefd=True) as source:
        data = source.read()
    parent = os.path.dirname(destination)
    for _ in range(20):
        temporary = os.path.join(parent, f".RUN.restore.{secrets.token_hex(12)}")
        try:
            output_fd = os.open(
                temporary,
                os.O_WRONLY | os.O_CREAT | os.O_EXCL | no_follow,
                0o600,
            )
            break
        except FileExistsError:
            continue
    else:
        raise OSError("could not allocate RUN restore temporary")
    with os.fdopen(output_fd, "wb", closefd=True) as output:
        output.write(data)
        output.flush()
        os.fsync(output.fileno())
    os.replace(temporary, destination)
os.close(checkpoint_fd)
os.close(root_fd)
PY
    log "HALT: rollback checkpoint RUN state is unsafe"
    return 1
  }
  python3 - "$STATE_DIR/REVERT-CLEANUP.json" "$good_sha" "$cur_sha" "$last" <<'PY'
import json
import os
import sys
import tempfile

path, good, current, checkpoint = sys.argv[1:]
record = {
    "good_fork_sha": good,
    "diverged_head_sha": current,
    "checkpoint": checkpoint,
    "instruction": (
        "REVERT: reset local-only commits after good_fork_sha, or revert-forward pushed commits "
        "per orchestrator gates; never force-push or merge. Then replay the stage."
    ),
}
fd, temporary = tempfile.mkstemp(prefix=".REVERT-CLEANUP.", dir=os.path.dirname(path))
try:
    with os.fdopen(fd, "w") as fh:
        json.dump(record, fh, indent=2)
        fh.write("\n")
    os.replace(temporary, path)
finally:
    try:
        os.unlink(temporary)
    except FileNotFoundError:
        pass
PY
  log "reverted ledger to checkpoint $last (fork ${good_sha:0:8}); orchestrator cleanup required for ${cur_sha:0:8}"
}

START_EPOCH="$(date +%s)"; reverts=0; no_verdict=0; correction=""
ensure_initial_checkpoint || exit 4
next_checkpoint="$(next_checkpoint_number)" || {
  log "HALT: preexisting checkpoint destinations are unsafe"
  exit 4
}
for (( i=1; i<=MAX_ITERATIONS; i++ )); do
  if (( MAX_MINUTES > 0 )) && (( ( $(date +%s) - START_EPOCH ) / 60 >= MAX_MINUTES )); then
    log "STOP: wall-clock cap ${MAX_MINUTES}m (resumable via --resume)"; exit 3
  fi

  log "── turn $i: ORCHESTRATOR ──${correction:+ (with correction)}"
  turn_msg="$LOOP_CMD $PROBLEM

$(state_root_instruction)"
  if [[ -n "$correction" ]]; then
    turn_msg="$turn_msg

VALIDATOR CORRECTION (apply first): $correction"
  fi
  run_orchestrator "$turn_msg" || log "turn $i: orchestrator returned non-zero (validator will assess)"

  validated_verdict=""
  shepherd_ready=0
  identity_rc=0
  if prepare_shepherd_request; then
    shepherd_ready=1
  else
    identity_rc=$?
    if (( identity_rc != 10 )); then
      log "HALT: stale, dirty, or unauthenticated Shepherd identity"
      rm -f "$REQUEST_JSON" "$VERDICT_JSON"
      exit 4
    fi
    rm -f "$REQUEST_JSON" "$VERDICT_JSON"
    log "turn $i: Shepherd deferred until authenticated clean exact-head synthesis"
  fi

  if (( shepherd_ready == 1 )); then
    log "── turn $i: VALIDATOR (exact clean synthesis) ──"
    validator_rc=0
    run_validator_with_watchdog || validator_rc=$?
    if (( validator_rc == 0 )); then
      if ! capture_validator_outputs \
        || ! revalidate_shepherd_request \
        || ! verify_bound_validator_output \
        || ! finalize_validation_ledger "$VALIDATION_RECORD_SNAPSHOT"; then
        validator_rc=65
        log "turn $i: validator output failed exact bound validation; discarding its verdict"
        finalize_validation_ledger || true
        secure_remove_validator_paths || true
      fi
    fi
    if (( validator_rc == 125 )); then
      log "HALT: validator session descendants could not be terminated and verified"
      exit 4
    fi
    if (( validator_rc != 0 )); then
      log "turn $i: validator returned non-zero ($validator_rc); discarding its verdict"
      validated_verdict=""
      VERDICT_SNAPSHOT=""
    fi

    verdict="$(json_field "$VERDICT_SNAPSHOT" verdict)"
    score="$(json_field "$VERDICT_SNAPSHOT" step_score)"
    reason="$(json_field "$VERDICT_SNAPSHOT" reason)"
    correction=""
    log "turn $i: verdict=${verdict:-NONE} step_score=${score:-?} — ${reason:-}"

    case "$verdict" in
      PROCEED)
        no_verdict=0
        validated_verdict=PROCEED
        checkpoint "$next_checkpoint" || { log "HALT: failed to record bound PROCEED checkpoint"; exit 4; }
        next_checkpoint=$((next_checkpoint + 1))
        ;;
      RETRY)
        no_verdict=0
        validated_verdict=RETRY
        correction="$(json_field "$VERDICT_SNAPSHOT" correction)"
        log "turn $i: RETRY — $correction"
        ;;
      REVERT)
        no_verdict=0
        validated_verdict=REVERT
        reverts=$((reverts+1))
        if (( reverts > MAX_REVERTS )); then
          log "HALT: MAX_REVERTS=$MAX_REVERTS exceeded"
          exit 4
        fi
        restore_checkpoint || exit 4
        correction="$(json_field "$VERDICT_SNAPSHOT" correction)"
        log "turn $i: REVERT #$reverts — $correction"
        ;;
      HALT)
        validated_verdict=HALT
        log "HALT: validator hard-stop — ${reason:-}"
        exit 4
        ;;
      *)
        no_verdict=$((no_verdict+1))
        if (( no_verdict >= MAX_NO_VERDICT )); then
          log "HALT: no valid bound VALIDATOR-VERDICT.json for $no_verdict consecutive turns. Check validator auth/tools and watchdog logs, then --resume."
          exit 4
        fi
        log "turn $i: no bound verdict ($no_verdict/$MAX_NO_VERDICT); retrying"
        correction="Emit exactly one bound Shepherd v1 validation entry and verdict with cited evidence."
        ;;
    esac
  fi

  AUTO_LOOP_STATE_DIR="$STATE_DIR" "$REPO_ROOT/scripts/loop-trace.sh" distill >/dev/null 2>&1 && log "turn $i: trace digest written (see $STATE_DIR/trace/INDEX.md)" || true

  terminal="$(json_field "$RUN_JSON" terminal)"; stage="$(json_field "$RUN_JSON" stage)"
  log "turn $i: stage=${stage:-?} terminal=${terminal:-none}"
  case "$terminal" in
    blocked)
      log "STOP: blocked (see ORCHESTRATION-STATE.json / VALIDATION.jsonl)."
      exit 4
      ;;
    budget)
      log "STOP: budget ceiling; re-run --resume."
      exit 3
      ;;
    human_gate)
      if ! terminal_is_authoritative; then
        log "HALT: human gate cannot bypass authenticated clean synthesis and fresh bound Shepherd PROCEED"
        exit 4
      fi
      gate_class="$(bash "$REPO_ROOT/scripts/pm-terminal-classifier.sh" "$RUN_JSON" 2>/dev/null || printf 'blocked_human_decision')"
      if [[ "$gate_class" == "human_ready" ]]; then
        log "DONE: human-ready gate reached; this driver never marks ready or merges."
        exit 0
      fi
      log "STOP: blocked human decision (see RUN.json and ORCHESTRATION-STATE.json for the gate kind)."
      exit 4
      ;;
    done)
      if ! terminal_is_authoritative; then
        log "HALT: done cannot bypass authenticated clean synthesis and fresh bound Shepherd PROCEED"
        exit 4
      fi
      log "DONE: all sub-issues complete and verified; no merge performed by driver."
      exit 0
      ;;
  esac

  sleep "$COOLDOWN_SECONDS"
done
log "STOP: MAX_ITERATIONS=$MAX_ITERATIONS without terminal (resumable via --resume)"; exit 3
