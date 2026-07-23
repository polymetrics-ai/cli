#!/usr/bin/env bash
set -uo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fixture_root="$repo_root/scripts/tests/fixtures/pm-review-system"
test_tmp="$(mktemp -d "$repo_root/.pm-review-test.XXXXXX")"
trap 'rm -rf "$test_tmp"' EXIT
failures=0

fail() {
  printf 'PM review-system semantic failure: %s\n' "$1" >&2
  failures=$((failures + 1))
}

# Frozen-corpus integrity is checked before any detector runs.
if ! python3 - "$fixture_root" <<'PY'
import hashlib
import json
import sys
from pathlib import Path

root = Path(sys.argv[1])
manifest = json.loads((root / "corpus-manifest.json").read_text())
for name, expected in manifest["files"].items():
    path = root / name
    digest = hashlib.sha256(path.read_bytes()).hexdigest()
    if digest != expected["sha256"] or path.stat().st_size != expected["bytes"]:
        raise SystemExit(f"frozen corpus drift: {name}")
PY
then
  fail "frozen corpus hash/size mismatch"
fi

# Accepted PR #495 finding 1: an explicit current-schema noncanonical gate kind must not be ready.
classification="$({
  bash "$repo_root/scripts/pm-terminal-classifier.sh" \
    "$repo_root/scripts/tests/fixtures/pm-orchestrator-review-state/noncanonical-final-parent-readiness.json"
} 2>&1)"
if [[ "$classification" != "blocked_human_decision" ]]; then
  fail "canonical final_parent_readiness classified as '$classification', want blocked_human_decision"
fi

# Accepted PR #495 finding 2: arbitrary valid finding IDs must not bypass disposition validation.
disposition_output="$({
  PM_DISPOSITION_ARTIFACT="scripts/tests/fixtures/pm-review-system/invalid-disposition-arbitrary-id.md" \
    bash "$repo_root/scripts/tests/pm-orchestrator-contract.sh"
} 2>&1)"
disposition_status=$?
if [[ $disposition_status -eq 0 ]]; then
  fail "SEC1 with noncanonical disposition was accepted by the current row parser"
elif [[ "$disposition_output" != *"noncanonical"* ]]; then
  fail "SEC1 rejection did not report the intended noncanonical-disposition reason: $disposition_output"
fi

# Detector receives no oracle argument. Comparison occurs only after each subprocess exits.
if ! python3 - "$repo_root" "$fixture_root" <<'PY'
import json
import subprocess
import sys
from pathlib import Path

repo = Path(sys.argv[1])
fixtures = Path(sys.argv[2])
inputs_path = fixtures / "inputs.json"
inputs = json.loads(inputs_path.read_text())
oracle = json.loads((fixtures / "oracle.json").read_text())["cases"]
script = repo / "scripts" / "pm-review-system.py"
errors = []

for case in inputs["cases"]:
    case_id = case["case_id"]
    proc = subprocess.run(
        [
            sys.executable,
            str(script),
            "detect",
            "--mode",
            "treatment",
            "--input",
            str(inputs_path),
            "--case-id",
            case_id,
        ],
        cwd=repo,
        check=False,
        capture_output=True,
        text=True,
    )
    if proc.returncode != 0:
        errors.append(f"{case_id}: detector exited {proc.returncode}: {proc.stderr.strip()}")
        continue
    try:
        observation = json.loads(proc.stdout)
    except json.JSONDecodeError as exc:
        errors.append(f"{case_id}: detector output is not one JSON envelope: {exc}")
        continue
    expected = oracle[case_id]
    findings = observation.get("findings", [])
    if expected["expected"] == "finding" and not findings:
        errors.append(
            f"{case_id}: missed semantic {expected['category']} mutation ({case['source_identity']})"
        )
    elif expected["expected"] == "clean" and findings:
        errors.append(f"{case_id}: false positive on clean {expected['category']} control")
    elif expected["expected"] == "decision" and observation.get("decision") != expected["decision"]:
        errors.append(
            f"{case_id}: threshold decision {observation.get('decision')!r}, "
            f"want {expected['decision']!r}"
        )

if errors:
    print("\n".join(errors), file=sys.stderr)
    raise SystemExit(1)
PY
then
  fail "treatment detector did not satisfy the frozen mutation/clean-control oracle"
fi

# Additional transition RED/GREEN cases exercise one-way legacy migration and append-only head
# history without modifying the frozen scoring corpus.
cat >"$test_tmp/transition-inputs.json" <<'JSON'
{
  "cases": [
    {
      "case_id": "transition-bad",
      "suite": "transition_contract",
      "family": "lineage_events",
      "input": {
        "exact_base_sha": "1212121212121212121212121212121212121212",
        "candidate_lineage": "stable-transition",
        "prior_head_history": ["3434343434343434343434343434343434343434", "5656565656565656565656565656565656565656"],
        "head_history": ["5656565656565656565656565656565656565656"],
        "events": [
          {"event": "migrate_legacy", "rounds": 2},
          {"event": "write_legacy", "rounds": 2}
        ],
        "correction_budget": {"max_correction_rounds": 5, "rounds_by_range": {"1212121212121212121212121212121212121212...stable-transition": 2}}
      }
    },
    {
      "case_id": "transition-clean",
      "suite": "transition_contract",
      "family": "lineage_events",
      "input": {
        "exact_base_sha": "7878787878787878787878787878787878787878",
        "candidate_lineage": "stable-transition",
        "prior_head_history": ["9090909090909090909090909090909090909090"],
        "head_history": ["9090909090909090909090909090909090909090", "abababababababababababababababababababab"],
        "events": [
          {"event": "migrate_legacy", "rounds": 2},
          {"event": "replace_head", "rounds": 2}
        ],
        "correction_budget": {"max_correction_rounds": 5, "rounds_by_range": {"7878787878787878787878787878787878787878...stable-transition": 2}}
      }
    }
  ]
}
JSON
for transition_id in transition-bad transition-clean; do
  python3 "$repo_root/scripts/pm-review-system.py" detect --mode treatment \
    --input "$test_tmp/transition-inputs.json" --case-id "$transition_id" \
    >"$test_tmp/$transition_id.json"
done
if ! python3 - "$test_tmp" <<'PY'
import json,pathlib,sys
root=pathlib.Path(sys.argv[1])
bad=json.loads((root/'transition-bad.json').read_text())
clean=json.loads((root/'transition-clean.json').read_text())
assert {item['category'] for item in bad['findings']} == {'lineage_events'}
assert not clean['findings']
PY
then
  fail "one-way legacy migration or append-only head-history transition is not enforced"
fi

# Reproduce the committed machine-readable baseline/treatment measurements. Timing is captured but
# not compared exactly; model token/cost and prospective evidence remain explicitly unavailable.
for mode in baseline treatment; do
  if ! python3 "$repo_root/scripts/pm-review-system.py" observe --mode "$mode" \
    --input "$fixture_root/inputs.json" >"$test_tmp/$mode-observations.json"
  then
    fail "$mode observation run failed"
    continue
  fi
  if ! python3 "$repo_root/scripts/pm-review-system.py" score \
    --observations "$test_tmp/$mode-observations.json" \
    --oracle "$fixture_root/oracle.json" >"$test_tmp/$mode-score.json"
  then
    fail "$mode scoring run failed"
  fi
done
if ! python3 - "$repo_root" "$test_tmp" <<'PY'
import json
import sys
from pathlib import Path

repo = Path(sys.argv[1])
tmp = Path(sys.argv[2])
report = json.loads((repo / ".planning/phases/397-pm-first-round-review-system-r1/MEASUREMENT.json").read_text())
for mode in ("baseline", "treatment"):
    score = json.loads((tmp / f"{mode}-score.json").read_text())
    expected = report[mode]
    counts = score["counts"]
    assert counts["true_positive"] == expected["true_positive"]
    assert counts["false_positive"] == expected["false_positive"]
    assert counts["false_negative"] == expected["false_negative"]
    assert counts["true_negative"] == expected["true_negative"]
    assert score["first_round_recall"] == expected["first_round_recall"]
    assert score["first_round_precision"] == expected["first_round_precision"]
    assert score["escaped_defects"] == expected["escaped_defects"]
    assert score["threshold_decisions"]["accuracy"] == expected["threshold_accuracy"]
    assert score["tokens"]["status"] == "unavailable"
    assert score["cost"]["status"] == "unavailable"
assert report["prospective_observation"]["status"] == "unavailable"
PY
then
  fail "committed measurement report does not reproduce from frozen inputs and separate oracle"
fi

# Preserve terminal-classifier command behavior as current enum handling is hardened.
usage_stdout="$test_tmp/usage.stdout"
usage_stderr="$test_tmp/usage.stderr"
malformed="$test_tmp/malformed.json"
printf '{not-json\n' >"$malformed"
bash "$repo_root/scripts/pm-terminal-classifier.sh" >"$usage_stdout" 2>"$usage_stderr"
usage_status=$?
if [[ $usage_status -ne 2 || -s "$usage_stdout" ]] || ! grep -Fq 'usage: scripts/pm-terminal-classifier.sh <RUN.json>' "$usage_stderr"; then
  fail "terminal classifier usage stdout/stderr/exit contract changed"
fi
bash "$repo_root/scripts/pm-terminal-classifier.sh" "$malformed" >"$usage_stdout" 2>"$usage_stderr"
malformed_status=$?
if [[ $malformed_status -eq 0 || -s "$usage_stdout" ]] || ! grep -Fq 'cannot classify PM terminal state:' "$usage_stderr"; then
  fail "terminal classifier malformed-JSON stdout/stderr/exit contract changed"
fi
rm -f "$usage_stdout" "$usage_stderr" "$malformed"

# Compile the exact committed range. Output must be one non-TTY JSON envelope containing paths and
# metadata only, with complete packet assignment and no environment-value leakage.
head_sha="$(git -C "$repo_root" rev-parse HEAD)"
compile_output="$test_tmp/compile.json"
PM_REVIEW_SECRET_SENTINEL='do-not-copy-this-environment-value' \
  python3 "$repo_root/scripts/pm-review-system.py" compile \
    --repo-root "$repo_root" \
    --base 0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20 \
    --head "$head_sha" >"$compile_output"
compile_status=$?
if [[ $compile_status -ne 0 ]]; then
  fail "exact-head compiler blocked the allowlisted current range"
elif grep -Fq 'do-not-copy-this-environment-value' "$compile_output"; then
  fail "compiler leaked an environment value"
elif ! python3 - "$compile_output" <<'PY'
import json
import sys

document = json.load(open(sys.argv[1]))
assert document["schema_version"] == "polymetrics.ai/pm-review-compile/v1"
assert document["status"] == "ready"
assert document["selection"] in {"combined", "split"}
assert document["content_policy"].startswith("paths and metadata only")
changed = set(document["changed_files"])
assigned = {path for packet in document["packets"] for path in packet["changed_files"]}
assert changed == assigned
closure = set(document["coverage_manifest"]["closure_files"])
covered_closure = {path for packet in document["packets"] for path in packet["closure_files"]}
assert closure == covered_closure
for packet in document["packets"]:
    assert packet["exact_base_sha"] == document["exact_base_sha"]
    assert packet["exact_head_sha"] == document["exact_head_sha"]
    assert len(packet["changed_files"]) <= 20
    assert len(packet["closure_files"]) <= 10
    assert len(packet["authority_files"]) <= 10
    assert not packet["context"]["overflow"]
    assert not packet["context"]["truncated"]
PY
then
  fail "compiled JSON envelope or packet coverage is invalid"
fi

# Malformed identities and escaping/symlinked config paths stop without reading broad files.
unsafe_dir="$test_tmp/unsafe"
mkdir -p "$unsafe_dir"
ln -s /etc/passwd "$unsafe_dir/escape.json"
unsafe_relative="$(python3 - "$repo_root" "$unsafe_dir/escape.json" <<'PY'
import os,sys
print(os.path.relpath(sys.argv[2], sys.argv[1]))
PY
)"
for unsafe_args in \
  "--base=-bad --config=.agents/agentic-delivery/contracts/pm-review-system.json" \
  "--base=0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20 --config=../../outside.json" \
  "--base=0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20 --config=$unsafe_relative"
do
  # shellcheck disable=SC2086 # fixed test-only arguments contain no user-controlled values.
  python3 "$repo_root/scripts/pm-review-system.py" compile --repo-root "$repo_root" \
    $unsafe_args --head "$head_sha" >"$unsafe_dir/result.json" 2>"$unsafe_dir/error.log"
  unsafe_status=$?
  if [[ $unsafe_status -eq 0 ]] || ! python3 - "$unsafe_dir/result.json" <<'PY'
import json,sys
assert json.load(open(sys.argv[1]))["status"] == "blocked"
PY
  then
    fail "unsafe identity/path case did not stop with a blocked JSON envelope: $unsafe_args"
  fi
done

# Clean packet responses synthesize once under PM ownership. Stale/incomplete/overflow responses
# block and cannot silently return clean.
responses_dir="$unsafe_dir/responses"
mkdir -p "$responses_dir"
cp "$compile_output" "$unsafe_dir/manifest.json"
python3 - "$compile_output" "$responses_dir" <<'PY'
import json
import pathlib
import sys

manifest = json.load(open(sys.argv[1]))
root = pathlib.Path(sys.argv[2])
for packet in manifest["packets"]:
    response = {
        "schema_version": "polymetrics.ai/pm-review-packet-response/v1",
        "packet_id": packet["packet_id"],
        "exact_base_sha": manifest["exact_base_sha"],
        "exact_head_sha": manifest["exact_head_sha"],
        "status": "clean",
        "reviewed_files": packet["changed_files"],
        "closure_files": packet["closure_files"],
        "authority_files": packet["authority_files"],
        "invariants": [
            {"id": item, "status": "pass", "evidence_paths": []}
            for item in packet["invariants"]
        ],
        "unreviewed_files": [],
        "findings": [],
        "context": {"input_tokens": None, "output_tokens": None, "cost": None, "overflow": False, "truncated": False},
        "wall_clock_ms": None,
    }
    (root / f"{packet['packet_id']}.json").write_text(json.dumps(response))
PY
manifest_relative="$(python3 - "$repo_root" "$unsafe_dir/manifest.json" <<'PY'
import os,sys
print(os.path.relpath(sys.argv[2], sys.argv[1]))
PY
)"
responses_relative="$(python3 - "$repo_root" "$responses_dir" <<'PY'
import os,sys
print(os.path.relpath(sys.argv[2], sys.argv[1]))
PY
)"
synthesis_output="$unsafe_dir/synthesis.json"
python3 "$repo_root/scripts/pm-review-system.py" synthesize --repo-root "$repo_root" \
  --manifest "$manifest_relative" --responses-dir "$responses_relative" >"$synthesis_output"
synthesis_status=$?
if [[ $synthesis_status -ne 0 ]] || ! python3 - "$synthesis_output" <<'PY'
import json,sys
value=json.load(open(sys.argv[1]))
assert value["status"] == "clean"
assert value["owner"] == "parent_orchestrator"
assert value["shepherd"]["status"] == "pending"
assert value["human_merge_authority"] is True
PY
then
  fail "complete clean packet responses did not produce one PM-owned clean synthesis"
fi
first_response="$(python3 - "$responses_dir" <<'PY'
import json,pathlib,sys
for path in sorted(pathlib.Path(sys.argv[1]).glob('*.json')):
    if json.loads(path.read_text()).get('reviewed_files'):
        print(path)
        break
PY
)"
python3 - "$first_response" <<'PY'
import json,sys
path=sys.argv[1]
value=json.load(open(path))
value["exact_head_sha"]="ffffffffffffffffffffffffffffffffffffffff"
value["context"]["overflow"]=True
value["reviewed_files"]=[]
open(path,"w").write(json.dumps(value))
PY
python3 "$repo_root/scripts/pm-review-system.py" synthesize --repo-root "$repo_root" \
  --manifest "$manifest_relative" --responses-dir "$responses_relative" >"$synthesis_output"
stale_status=$?
if [[ $stale_status -eq 0 ]] || ! python3 - "$synthesis_output" <<'PY'
import json,sys
value=json.load(open(sys.argv[1]))
assert value["status"] == "blocked"
categories={item["category"] for item in value["blockers"]}
assert {"stale_evidence", "packet_coverage", "packet_overflow"}.issubset(categories)
PY
then
  fail "stale/incomplete/overflow packet response did not block synthesis"
fi
rm -rf "$unsafe_dir" "$compile_output"

if [[ $failures -ne 0 ]]; then
  printf 'PM review-system contract: %d semantic group(s) failed\n' "$failures" >&2
  exit 1
fi

printf 'pm review system ok: semantic gates, bounded exact-head packets, one PM synthesis, measured fixtures\n'
