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

# Captain correction corpus is separately frozen before impact-graph/lab treatment. The detector
# sees only correction-inputs.json and one case id; the oracle is read after each subprocess exits.
if ! python3 - "$fixture_root" <<'PY'
import hashlib
import json
import sys
from pathlib import Path

root = Path(sys.argv[1])
manifest = json.loads((root / "correction-corpus-manifest.json").read_text())
for name, expected in manifest["files"].items():
    path = root / name
    digest = hashlib.sha256(path.read_bytes()).hexdigest()
    if digest != expected["sha256"] or path.stat().st_size != expected["bytes"]:
        raise SystemExit(f"frozen captain correction corpus drift: {name}")
PY
then
  fail "frozen captain correction corpus hash/size mismatch"
fi

if ! python3 - "$repo_root" "$fixture_root" <<'PY'
import json
import subprocess
import sys
from pathlib import Path

repo = Path(sys.argv[1])
fixtures = Path(sys.argv[2])
inputs_path = fixtures / "correction-inputs.json"
cases = json.loads(inputs_path.read_text())["cases"]
oracle = json.loads((fixtures / "correction-oracle.json").read_text())["cases"]
script = repo / "scripts" / "pm-review-system.py"
errors = []
for case in cases:
    case_id = case["case_id"]
    proc = subprocess.run(
        [sys.executable, str(script), "detect", "--mode", "treatment", "--input", str(inputs_path), "--case-id", case_id],
        cwd=repo,
        check=False,
        capture_output=True,
        text=True,
    )
    if proc.returncode != 0:
        errors.append(f"{case_id}: detector exited {proc.returncode}: {proc.stderr.strip()}")
        continue
    observation = json.loads(proc.stdout)
    expected = oracle[case_id]
    findings = observation.get("findings", [])
    if expected["expected"] == "finding" and not findings:
        errors.append(f"{case_id}: missed captain correction mutation ({case['source_identity']})")
    elif expected["expected"] == "clean" and findings:
        errors.append(f"{case_id}: false positive on captain correction clean control")
if errors:
    print("\n".join(errors), file=sys.stderr)
    raise SystemExit(1)
PY
then
  fail "captain impact-graph/lab detector did not satisfy the pre-frozen correction oracle"
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

for mode in baseline treatment; do
  python3 "$repo_root/scripts/pm-review-system.py" observe --mode "$mode" \
    --input "$fixture_root/correction-inputs.json" >"$test_tmp/correction-$mode-observations.json" || \
    fail "captain correction $mode observation failed"
  python3 "$repo_root/scripts/pm-review-system.py" score \
    --observations "$test_tmp/correction-$mode-observations.json" \
    --oracle "$fixture_root/correction-oracle.json" >"$test_tmp/correction-$mode-score.json" || \
    fail "captain correction $mode scoring failed"
done
if ! python3 - "$repo_root" "$test_tmp" <<'PY'
import json,sys
from pathlib import Path
repo=Path(sys.argv[1]); tmp=Path(sys.argv[2])
report=json.loads((repo/'.planning/phases/397-pm-first-round-review-system-r1/CORRECTION-MEASUREMENT.json').read_text())
for mode in ('baseline','treatment'):
    score=json.loads((tmp/f'correction-{mode}-score.json').read_text()); expected=report[mode]
    counts=score['counts']
    assert counts['true_positive']==expected['true_positive']
    assert counts['false_positive']==expected['false_positive']
    assert counts['false_negative']==expected['false_negative']
    assert counts['true_negative']==expected['true_negative']
    assert score['first_round_recall']==expected['first_round_recall']
    assert score['first_round_precision']==expected['first_round_precision']
    assert score['escaped_defects']==expected['escaped_defects']
assert report['reporting_scope']['actual_model_packet_metrics']['status']=='unavailable'
assert report['reporting_scope']['prospective_observation']['status']=='unavailable'
PY
then
  fail "captain correction measurement does not reproduce from frozen inputs and separate oracle"
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

# Captain impact-graph integration fixture: index the declared universe before traversal from all
# changed files and canonical roots. This real Git/Go fixture is separate from the detector corpus.
impact_repo="$test_tmp/impact-repo"
mkdir -p \
  "$impact_repo/.agents/templates" "$impact_repo/.agents/schema" "$impact_repo/.agents/generated" \
  "$impact_repo/.agents/cycle" "$impact_repo/.pi/prompts" "$impact_repo/.planning" \
  "$impact_repo/scripts" "$impact_repo/internal/lib" "$impact_repo/cmd/app"
git -C "$impact_repo" init -q
cat >"$impact_repo/go.mod" <<'EOF'
module example.test/impact

go 1.24
EOF
cat >"$impact_repo/.pi/prompts/canonical.md" <<'EOF'
# Canonical root
EOF
cat >"$impact_repo/.pi/prompts/upstream.md" <<'EOF'
Required template: `.agents/templates/leaf.md`
EOF
cat >"$impact_repo/.agents/templates/leaf.md" <<'EOF'
leaf v1
EOF
cat >"$impact_repo/.pi/prompts/script-user.md" <<'EOF'
Run `python3 scripts/tool.py` for the review input.
EOF
cat >"$impact_repo/scripts/tool.py" <<'EOF'
from helper import value
print(value)
EOF
cat >"$impact_repo/scripts/helper.py" <<'EOF'
value = "helper"
EOF
cat >"$impact_repo/.agents/schema/state.yaml" <<'EOF'
schema_version: v1
EOF
cat >"$impact_repo/.pi/prompts/state-writer.md" <<'EOF'
Writes `.agents/schema/state.yaml`.
EOF
cat >"$impact_repo/scripts/state_reader.py" <<'EOF'
from pathlib import Path
STATE = Path('.agents/schema/state.yaml')
EOF
cat >"$impact_repo/.planning/state-mirror.json" <<'EOF'
{"schema_version":"v1"}
EOF
cat >"$impact_repo/scripts/generate.py" <<'EOF'
print("fixture generator")
EOF
cat >"$impact_repo/.agents/generated/data.json" <<'EOF'
{"version":1}
EOF
cat >"$impact_repo/.pi/prompts/generated-user.md" <<'EOF'
Consumes `.agents/generated/data.json`.
EOF
cat >"$impact_repo/internal/lib/lib.go" <<'EOF'
package lib
func Value() string { return "v1" }
EOF
cat >"$impact_repo/internal/lib/lib_test.go" <<'EOF'
package lib
import "testing"
func TestValue(t *testing.T) { if Value() == "" { t.Fatal("empty") } }
EOF
cat >"$impact_repo/internal/lib/impl_darwin.go" <<'EOF'
package lib
const Platform = "darwin"
EOF
cat >"$impact_repo/internal/lib/impl_linux.go" <<'EOF'
package lib
const Platform = "linux"
EOF
cat >"$impact_repo/cmd/app/main.go" <<'EOF'
package main
import "example.test/impact/internal/lib"
func main() { _ = lib.Value() }
EOF
cat >"$impact_repo/.agents/cycle/a.md" <<'EOF'
Requires `.agents/cycle/b.md`.
EOF
cat >"$impact_repo/.agents/cycle/b.md" <<'EOF'
Requires `.agents/cycle/a.md`.
EOF
cat >"$impact_repo/.agents/unrelated.md" <<'EOF'
unrelated control
EOF
cat >"$impact_repo/behavior.txt" <<'EOF'
reject
EOF
cat >"$impact_repo/scripts/check_behavior.py" <<'EOF'
from pathlib import Path
raise SystemExit(0 if Path("behavior.txt").read_text().strip() == "reject" else 23)
EOF
cat >"$impact_repo/scripts/check_env.py" <<'EOF'
import os
raise SystemExit(99 if any("SECRET_SENTINEL" in key or "do-not-copy" in value for key, value in os.environ.items()) else 0)
EOF
cat >"$impact_repo/scripts/sleep.py" <<'EOF'
import time
time.sleep(5)
EOF
cat >"$impact_repo/scripts/output.py" <<'EOF'
print("x" * 200000)
EOF
cat >"$impact_repo/scripts/disk.py" <<'EOF'
from pathlib import Path
for index in range(200): Path(f"disk-{index}.bin").write_bytes(b"x" * 65536)
EOF
cat >"$impact_repo/scripts/processes.py" <<'EOF'
import subprocess, sys, time
children = [subprocess.Popen([sys.executable, "-c", "import time; time.sleep(5)"]) for _ in range(12)]
time.sleep(5)
EOF
outside_target="$test_tmp/lab-outside-sentinel"
python3 - "$outside_target" "$impact_repo/scripts" <<'PY'
import pathlib,sys
outside=repr(sys.argv[1]); scripts=pathlib.Path(sys.argv[2])
(scripts/'outside_write.py').write_text(f"from pathlib import Path\nPath({outside}).write_text('escaped')\n")
(scripts/'outside_read.py').write_text(f"from pathlib import Path\nprint(Path({outside}).read_text())\n")
PY
cat >"$impact_repo/scripts/network.py" <<'EOF'
import socket
socket.create_connection(('127.0.0.1', 9), timeout=0.2)
EOF
cat >"$impact_repo/scripts/check_isolation.py" <<'EOF'
from pathlib import Path
try:
    list(Path.cwd().parents[1].iterdir())
except PermissionError:
    raise SystemExit(0)
raise SystemExit(77)
EOF
ln -s "$outside_target" "$impact_repo/escape-link"
cat >"$impact_repo/.agents/review-config.json" <<'JSON'
{
  "schema_version":"polymetrics.ai/pm-review-system/v2",
  "owner":"parent_orchestrator",
  "canonical_roots":[".pi/prompts/canonical.md"],
  "reference_prefixes":[".agents/",".pi/",".planning/","scripts/","cmd/","internal/"],
  "ignored_reference_prefixes":[],
  "prohibited_active_targets":[],
  "impact_graph":{
    "index_prefixes":[".agents/",".pi/",".planning/","scripts/","cmd/","internal/"],
    "max_index_files":200,"max_index_bytes":2000000,"max_nodes":300,"max_edges":1000,
    "max_traversal_states":1000,"max_depth":12,"max_impact_files":200,"max_impact_edges":1000,
    "go_command_timeout_seconds":20,"packet_max_impact_files":10,"packet_max_impact_edges":40
  },
  "configured_relationships":[
    {"source":"scripts/generate.py","target":".agents/generated/data.json","relation":"generates","certainty":"active"},
    {"source":".pi/prompts/generated-user.md","target":".agents/generated/data.json","relation":"generated_consumer","certainty":"active"},
    {"source":".agents/schema/state.yaml","target":".planning/state-mirror.json","relation":"temporal_state_mirror","certainty":"active"}
  ],
  "authorities":[{
    "id":"fixture_state","authoritative_path":".agents/schema/state.yaml",
    "writers":[".pi/prompts/state-writer.md"],"readers":["scripts/state_reader.py"],
    "mirrors":[".planning/state-mirror.json"],"invariants":["fixture_authority_complete"]
  }],
  "thresholds":{
    "combined_max_files":20,"combined_max_non_generated_lines":600,"combined_max_domains":1,
    "mandatory_split_files":25,"mandatory_split_non_generated_lines":800,"mandatory_split_domains":2,
    "packet_max_changed_files":20,"packet_max_context_files":10,"packet_target_tokens":30000,
    "estimated_tokens_per_changed_line":4,"estimated_tokens_per_context_file":200
  },
  "domain_rules":[
    {"domain":"implementation_test","patterns":["scripts/**","cmd/**","internal/**"]},
    {"domain":"architecture_reference","patterns":[".agents/**",".pi/**"]},
    {"domain":"authority_workflow_state","patterns":[".planning/**"]}
  ],
  "packet_invariants":{
    "architecture_reference":["impact_complete"],"authority_workflow_state":["impact_complete"],
    "implementation_test":["impact_complete"],"impact_graph":["impact_complete"],"combined":["impact_complete"]
  },
  "allowed_changed_paths":["**"],"forbidden_changed_paths":[]
}
JSON
python3 - "$impact_repo/.agents/review-config.json" "$impact_repo/.agents/bound-config.json" "$impact_repo/.agents/legacy-config.json" <<'PY'
import json,sys
normal=json.load(open(sys.argv[1]))
bound=json.loads(json.dumps(normal)); bound["impact_graph"]["max_impact_files"]=3
open(sys.argv[2],"w").write(json.dumps(bound,indent=2)+"\n")
legacy=json.loads(json.dumps(normal)); legacy["schema_version"]="polymetrics.ai/pm-review-system/v1"
open(sys.argv[3],"w").write(json.dumps(legacy,indent=2)+"\n")
PY
git -C "$impact_repo" add .
git -C "$impact_repo" -c user.name='PM Review Test' -c user.email='pm-review-test@example.invalid' commit -qm base
impact_base="$(git -C "$impact_repo" rev-parse HEAD)"
printf 'leaf v2\n' >"$impact_repo/.agents/templates/leaf.md"
printf 'schema_version: v2\n' >"$impact_repo/.agents/schema/state.yaml"
printf '{"version":2}\n' >"$impact_repo/.agents/generated/data.json"
printf 'from helper import value\nprint(value.upper())\n' >"$impact_repo/scripts/tool.py"
printf 'package lib\nfunc Value() string { return "v2" }\n' >"$impact_repo/internal/lib/lib.go"
cat >"$impact_repo/.agents/cycle/a.md" <<'EOF'
Requires `.agents/cycle/b.md`.
Changed cycle seed.
EOF
git -C "$impact_repo" add .
git -C "$impact_repo" -c user.name='PM Review Test' -c user.email='pm-review-test@example.invalid' commit -qm head
impact_head="$(git -C "$impact_repo" rev-parse HEAD)"
impact_manifest="$test_tmp/impact-manifest.json"
python3 "$repo_root/scripts/pm-review-system.py" compile --repo-root "$impact_repo" \
  --config .agents/review-config.json --base "$impact_base" --head "$impact_head" >"$impact_manifest"
impact_status=$?
if [[ $impact_status -ne 0 ]] || ! python3 - "$impact_manifest" <<'PY'
import json,sys
value=json.load(open(sys.argv[1]))
assert value["schema_version"] == "polymetrics.ai/pm-review-compile/v2"
graph=value["impact_graph"]
files=set(value["coverage_manifest"]["impact_files"])
required={
 ".agents/templates/leaf.md", ".pi/prompts/upstream.md", "scripts/tool.py", "scripts/helper.py",
 ".pi/prompts/script-user.md", ".agents/schema/state.yaml", ".pi/prompts/state-writer.md",
 "scripts/state_reader.py", ".planning/state-mirror.json", "scripts/generate.py",
 ".agents/generated/data.json", ".pi/prompts/generated-user.md", "internal/lib/lib.go",
 "internal/lib/lib_test.go", "internal/lib/impl_darwin.go", "internal/lib/impl_linux.go",
 "cmd/app/main.go", ".agents/cycle/a.md", ".agents/cycle/b.md"
}
assert required <= files, sorted(required-files)
assert ".agents/unrelated.md" not in files
assert not graph["bounds"]["hit"]
assert {"active", "unknown"} <= {edge["certainty"] for edge in graph["edges"]}
relations={edge["relation"] for edge in graph["edges"]}
assert {"required_reference","python_import","authority_writes","authority_reads","authority_mirror","generates","go_imports","go_test","platform_variant"} <= relations
packet_files={path for packet in value["packets"] for path in packet["impact_files"]}
packet_edges={edge for packet in value["packets"] for edge in packet["impact_edge_ids"]}
assert packet_files == files
assert packet_edges == set(value["coverage_manifest"]["impact_edge_ids"])
PY
then
  fail "changed-file-seeded bidirectional typed impact graph or exact packet coverage is absent"
fi
for config_case in bound-config legacy-config; do
  python3 "$repo_root/scripts/pm-review-system.py" compile --repo-root "$impact_repo" \
    --config ".agents/$config_case.json" --base "$impact_base" --head "$impact_head" \
    >"$test_tmp/$config_case-result.json"
  config_status=$?
  if [[ $config_status -eq 0 ]] || ! python3 - "$test_tmp/$config_case-result.json" <<'PY'
import json,sys
value=json.load(open(sys.argv[1])); assert value["status"] == "blocked"; assert value["findings"]
PY
  then
    fail "$config_case did not fail closed for graph bound or incompatible schema migration"
  fi
done

# Counterfactual lab integration is authored before the runner. Each experiment must create and
# destroy its own exact-head copy; the candidate is never the writable experiment root.
lab_script="$repo_root/scripts/pm-review-lab.py"
if [[ ! -f "$lab_script" ]]; then
  fail "counterfactual lab runner is absent; safety/identity/cleanup RED is expected"
elif ! python3 - "$lab_script" "$impact_repo" "$impact_base" "$impact_head" "$test_tmp" "$outside_target" <<'PY'
import json, os, pathlib, subprocess, sys
from concurrent.futures import ThreadPoolExecutor

script, repo, base, head, tmp, outside_arg = sys.argv[1:]
root=pathlib.Path(tmp); labs=root/'labs'; labs.mkdir()

def request(name, **overrides):
    value={
      "schema_version":"polymetrics.ai/pm-review-lab-request/v1","hypothesis_id":name,
      "claim":"temporary accept behavior must make the targeted check fail",
      "alternative":"the behavior change is irrelevant to the targeted check",
      "impact_edges_examined":["edge-fixture-1"],
      "temporary_change":"replace reject with accept in the disposable copy",
      "changes":[{"path":"behavior.txt","find":"reject\n","replace":"accept\n"}],
      "command":["python3","scripts/check_behavior.py"],
      "expected_discriminator":{"exit_code":23}
    }
    value.update(overrides); path=root/f"{name}.json"; path.write_text(json.dumps(value)); return path

def run(path, extra_env=None):
    env=os.environ.copy(); env.update(extra_env or {})
    proc=subprocess.run([sys.executable,script,"run","--repo-root",repo,"--base",base,"--head",head,"--packet-id",path.stem,"--request",str(path),"--temp-root",str(labs)],text=True,capture_output=True,env=env)
    try: value=json.loads(proc.stdout)
    except Exception as exc: raise AssertionError((path.name,proc.returncode,proc.stdout,proc.stderr)) from exc
    return proc.returncode,value

probe=subprocess.run([sys.executable,script,"probe"],text=True,capture_output=True)
probe_value=json.loads(probe.stdout)
if probe_value["status"]=="blocked":
    code,safe=run(request('safe'))
    assert code!=0 and safe["status"]=="blocked" and safe["candidate_unchanged"] and safe["lab_cleanup_verified"]
    assert "policy-only" in json.dumps(safe)
    raise SystemExit(0)
code,safe=run(request('safe'))
assert code==0 and safe["status"]=="evidence" and safe["candidate_unchanged"] and safe["lab_cleanup_verified"]
assert safe["experiment"]["discriminator_matched"] and safe["experiment"]["exit_code"]==23, safe
assert safe["experiment"]["temporary_diff"]["sha256"] and safe["experiment"]["duration_ms"] is not None
assert not any(labs.iterdir())

def blocked(name, **changes):
    code,value=run(request(name,**changes)); assert code!=0 and value["status"]=="blocked"; return value
blocked('candidate-write', change_scope='candidate')
blocked('outside-write', changes=[{"path":"../outside","find":"x","replace":"y"}])
blocked('symlink-write', changes=[{"path":"escape-link","find":"x","replace":"y"}])
outside=pathlib.Path(outside_arg); outside.write_text('sentinel')
blocked('nested-outside-write',changes=[],command=['python3','scripts/outside_write.py'],expected_discriminator={'exit_code':0})
assert outside.read_text()=='sentinel'
blocked('nested-outside-read',changes=[],command=['python3','scripts/outside_read.py'],expected_discriminator={'exit_code':0})
blocked('nested-network',changes=[],command=['python3','scripts/network.py'],expected_discriminator={'exit_code':0})
outside.unlink()
for name,command in {
  'network':['curl','https://example.invalid'], 'commit':['git','commit','-am','x'],
  'push':['git','push','origin','HEAD'], 'install':['python3','-m','pip','install','x'],
  'generic-shell':['bash','-c','true'], 'live-connector':['pm','etl','run','--connection','live'],
  'deploy':['kubectl','apply','-f','x']
}.items(): blocked(name,changes=[],command=command,expected_discriminator={"exit_code":0})
code,envsafe=run(request('env-safe',changes=[],command=['python3','scripts/check_env.py'],expected_discriminator={'exit_code':0}),{'PM_REVIEW_SECRET_SENTINEL':'do-not-copy'})
assert code==0 and envsafe['status']=='evidence' and 'do-not-copy' not in json.dumps(envsafe)
for name,script_name,limit in (
 ('timeout','sleep.py',{'timeout_seconds':0.2}),('output','output.py',{'max_output_bytes':4096}),
 ('disk','disk.py',{'max_disk_bytes':131072}),('processes','processes.py',{'max_processes':4})
): blocked(name,changes=[],command=['python3',f'scripts/{script_name}'],expected_discriminator={'exit_code':0},limits=limit)
blocked('cleanup-failure',changes=[],extra='ignored') if False else None
code,cleanup=run(request('cleanup-failure',changes=[]),{'PM_REVIEW_LAB_TEST_FORCE_CLEANUP_FAILURE':'1'})
assert code!=0 and cleanup['status']=='blocked'
code,drift=run(request('identity-drift',changes=[]),{'PM_REVIEW_LAB_TEST_FORCE_IDENTITY_DRIFT':'1'})
assert code!=0 and drift['status']=='blocked'
with ThreadPoolExecutor(max_workers=2) as pool:
    rows=list(pool.map(lambda p: run(p),[
      request('parallel-a'),
      request('parallel-b',changes=[],command=['python3','scripts/check_isolation.py'],expected_discriminator={"exit_code":0})
    ]))
assert all(code==0 and value['status']=='evidence' for code,value in rows)
assert not any(labs.iterdir())
assert subprocess.check_output(['git','-C',repo,'rev-parse','HEAD'],text=True).strip()==head
assert not subprocess.check_output(['git','-C',repo,'status','--porcelain'],text=True)
PY
then
  fail "counterfactual lab did not enforce exact-head isolation, denial, bounds, evidence, concurrency, and cleanup"
fi

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
assert document["schema_version"] == "polymetrics.ai/pm-review-compile/v2"
assert document["status"] == "ready"
assert document["selection"] in {"combined", "split"}
assert document["content_policy"].startswith("paths and metadata only")
changed = set(document["changed_files"])
assigned = {path for packet in document["packets"] for path in packet["changed_files"]}
assert changed == assigned
closure = set(document["coverage_manifest"]["closure_files"])
covered_closure = {path for packet in document["packets"] for path in packet["closure_files"]}
assert closure == covered_closure
impact = set(document["coverage_manifest"]["impact_files"])
covered_impact = {path for packet in document["packets"] for path in packet["impact_files"]}
assert impact == covered_impact
impact_edges = set(document["coverage_manifest"]["impact_edge_ids"])
covered_edges = {edge for packet in document["packets"] for edge in packet["impact_edge_ids"]}
assert impact_edges == covered_edges
assert document["impact_graph"]["seed_files"]
assert not document["impact_graph"]["bounds"]["hit"]
for packet in document["packets"]:
    assert packet["exact_base_sha"] == document["exact_base_sha"]
    assert packet["exact_head_sha"] == document["exact_head_sha"]
    assert packet["exact_head_tree"] == document["exact_head_tree"]
    assert len(packet["changed_files"]) <= 20
    assert len(packet["closure_files"]) <= 10
    assert len(packet["authority_files"]) <= 10
    assert len(packet["impact_files"]) <= 10
    assert len(packet["impact_edge_ids"]) <= 40
    assert set(packet["impact_edge_ids"]) == {edge["id"] for edge in packet["impact_edges"]}
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
        "schema_version": "polymetrics.ai/pm-review-packet-response/v2",
        "packet_id": packet["packet_id"],
        "exact_base_sha": manifest["exact_base_sha"],
        "exact_head_sha": manifest["exact_head_sha"],
        "exact_head_tree": manifest["exact_head_tree"],
        "status": "clean",
        "reviewed_files": packet["changed_files"],
        "closure_files": packet["closure_files"],
        "authority_files": packet["authority_files"],
        "impact_files": packet.get("impact_files", []),
        "impact_edge_ids": packet.get("impact_edge_ids", []),
        "invariants": [
            {"id": item, "status": "pass", "evidence_paths": []}
            for item in packet["invariants"]
        ],
        "unreviewed_files": [],
        "review_behaviors": {
            "impact_model_built_first": True,
            "directions_traced": ["upstream", "downstream", "lateral", "temporal"],
            "history_inspected": {"status": "not_needed", "reason": "fixture evidence is unambiguous"},
            "sibling_paths_compared": {"status": "not_needed", "reason": "fixture has no divergent sibling"},
            "hypotheses": [],
            "disconfirming_evidence": "fixture contract and exact manifest agree"
        },
        "experiments": [],
        "no_experiment_reason": "static fixture evidence is decisive",
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
cp "$first_response" "$first_response.clean"
python3 - "$first_response" <<'PY'
import json,sys
path=sys.argv[1]; value=json.load(open(path)); value["schema_version"]="polymetrics.ai/pm-review-packet-response/v1"; open(path,"w").write(json.dumps(value))
PY
python3 "$repo_root/scripts/pm-review-system.py" synthesize --repo-root "$repo_root" \
  --manifest "$manifest_relative" --responses-dir "$responses_relative" >"$synthesis_output"
v1_status=$?
if [[ $v1_status -eq 0 ]] || ! python3 - "$synthesis_output" <<'PY'
import json,sys
value=json.load(open(sys.argv[1])); assert value["status"]=="blocked"; assert any("migration" in item["claim"] for item in value["blockers"])
PY
then
  fail "incompatible v1 packet response did not require explicit migration"
fi
mv "$first_response.clean" "$first_response"
python3 - "$first_response" <<'PY'
import json,sys
path=sys.argv[1]; value=json.load(open(path)); value["no_experiment_reason"]=None
value["experiments"]=[{"hypothesis_id":"H-inconclusive","claim":"suspected defect","alternative":"safe behavior","impact_edges_examined":[],"temporary_change":"fixture change","command":["python3","fixture.py"],"expected_discriminator":"different exits","observed":"same exit","supports":"inconclusive","candidate_unchanged":True,"lab_cleanup_verified":True}]
open(path,"w").write(json.dumps(value))
PY
python3 "$repo_root/scripts/pm-review-system.py" synthesize --repo-root "$repo_root" \
  --manifest "$manifest_relative" --responses-dir "$responses_relative" >"$synthesis_output"
inconclusive_status=$?
if [[ $inconclusive_status -eq 0 ]] || ! python3 - "$synthesis_output" <<'PY'
import json,sys
value=json.load(open(sys.argv[1])); assert value["status"]=="blocked"; assert any(item["category"]=="hypothesis_evidence" for item in value["blockers"])
PY
then
  fail "inconclusive counterfactual experiment was allowed to prove clean"
fi
cp "$first_response.clean" "$first_response" 2>/dev/null || true
if [[ ! -f "$first_response.clean" ]]; then
  python3 - "$compile_output" "$first_response" <<'PY'
# Restore is regenerated from the manifest because the clean backup was moved above.
import json,sys
manifest=json.load(open(sys.argv[1])); packet=next(p for p in manifest['packets'] if p['packet_id'] in sys.argv[2])
value={"schema_version":"polymetrics.ai/pm-review-packet-response/v2","packet_id":packet["packet_id"],"exact_base_sha":manifest["exact_base_sha"],"exact_head_sha":manifest["exact_head_sha"],"exact_head_tree":manifest["exact_head_tree"],"status":"clean","reviewed_files":packet["changed_files"],"closure_files":packet["closure_files"],"authority_files":packet["authority_files"],"impact_files":packet.get("impact_files",[]),"impact_edge_ids":packet.get("impact_edge_ids",[]),"invariants":[{"id":x,"status":"pass","evidence_paths":[]} for x in packet["invariants"]],"unreviewed_files":[],"review_behaviors":{"impact_model_built_first":True,"directions_traced":["upstream","downstream","lateral","temporal"],"history_inspected":{"status":"not_needed","reason":"fixture"},"sibling_paths_compared":{"status":"not_needed","reason":"fixture"},"hypotheses":[],"disconfirming_evidence":"fixture"},"experiments":[],"no_experiment_reason":"static fixture evidence is decisive","findings":[],"context":{"input_tokens":None,"output_tokens":None,"cost":None,"overflow":False,"truncated":False},"wall_clock_ms":None}
open(sys.argv[2],"w").write(json.dumps(value))
PY
fi
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

printf 'pm review system ok: semantic gates, bidirectional impact, disposable labs, bounded exact-head packets, one PM synthesis, measured fixtures\n'
