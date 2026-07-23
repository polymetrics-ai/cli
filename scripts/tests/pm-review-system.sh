#!/usr/bin/env bash
set -uo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fixture_root="$repo_root/scripts/tests/fixtures/pm-review-system"
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

if [[ $failures -ne 0 ]]; then
  printf 'PM review-system RED: %d semantic group(s) failing as expected before treatment\n' "$failures" >&2
  exit 1
fi

printf 'pm review system ok: semantic gates, bounded exact-head packets, one PM synthesis, measured fixtures\n'
