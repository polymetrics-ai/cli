#!/usr/bin/env bash
set -uo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fixture_root="$repo_root/scripts/tests/fixtures/pm-review-system"
test_parent="$repo_root/pm-review-tests.tmp"
mkdir -p "$test_parent"
test_tmp="$(mktemp -d "$test_parent/run.XXXXXX")"
cleanup() {
  chmod -R u+w "$test_tmp" 2>/dev/null || true
  if [[ "${PM_REVIEW_TEST_KEEP_TMP:-0}" != "1" ]]; then
    rm -rf "$test_tmp"
    rmdir "$test_parent" 2>/dev/null || true
  else
    printf 'PM review-system kept test root: %s\n' "$test_tmp" >&2
  fi
}
trap cleanup EXIT
failures=0

fail() {
  printf 'PM review-system semantic failure: %s\n' "$1" >&2
  failures=$((failures + 1))
}

write_trust_bundle() {
  local candidate_root="$1" base_sha="$2" head_sha="$3" config_path="$4" scope_path="$5" output_path="$6"
  python3 - "$candidate_root" "$base_sha" "$head_sha" "$config_path" "$scope_path" \
    "$repo_root/scripts/pm-review-system.py" "$output_path" <<'PY'
import hashlib,json,subprocess,sys
from pathlib import Path
root,base,head,config_path,scope_path,compiler,output=sys.argv[1:]
def git(*args):
    return subprocess.check_output(["git","-C",root,*args])
def digest(value):
    return hashlib.sha256(value).hexdigest()
def canonical(value):
    return hashlib.sha256(json.dumps(value,sort_keys=True,separators=(",",":"),ensure_ascii=True).encode()).hexdigest()
config_bytes=git("show",f"{head}:{config_path}")
scope_bytes=git("show",f"{head}:{scope_path}")
config=json.loads(config_bytes)
scope=json.loads(scope_bytes)
prompt=[]
for relative in config.get("prompt_contract_files",[]):
    payload=git("show",f"{head}:{relative}")
    prompt.append({"path":relative,"bytes":len(payload),"sha256":digest(payload)})
bundle={
 "schema_version":"polymetrics.ai/pm-review-trust-bundle/v1",
 "exact_base_sha":base,"exact_head_sha":head,
 "exact_head_tree":git("rev-parse",f"{head}^{{tree}}").decode().strip(),
 "candidate_lineage":scope["candidate_lineage"],"review_round":scope["review_round"],
 "compiler_sha256":digest(Path(compiler).read_bytes()),"config_sha256":digest(config_bytes),
 "scope_sha256":digest(scope_bytes),"prompt_contract_sha256":canonical(prompt),
 "approval_reference":"captain-approved deterministic test trust binding",
}
Path(output).write_text(json.dumps(bundle,sort_keys=True)+"\n")
PY
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
explicit_null_classification="$({
  bash "$repo_root/scripts/pm-terminal-classifier.sh" \
    "$repo_root/scripts/tests/fixtures/pm-orchestrator-review-state/explicit-null-schema-human-gate.json"
} 2>&1)"
if [[ "$explicit_null_classification" != "blocked_human_decision" ]]; then
  fail "explicit null schema alias classified as '$explicit_null_classification', want blocked_human_decision"
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
Requires `.agents/mixed/a.md`.
Requires `AGENTS.md`.
EOF
cat >"$impact_repo/AGENTS.md" <<'EOF'
# Governing fixture instructions
EOF
mkdir -p "$impact_repo/.agents/mixed"
cat >"$impact_repo/.agents/mixed/a.md" <<'EOF'
Requires `.agents/mixed/b.md`.
EOF
cat >"$impact_repo/.agents/mixed/b.md" <<'EOF'
Run `python3 scripts/mixed.py`.
EOF
cat >"$impact_repo/scripts/mixed.py" <<'EOF'
print("mixed relation target")
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
cat >"$impact_repo/scripts/escape_process.py" <<'EOF'
import os, subprocess, sys
subprocess.Popen([sys.executable, "-c", "import time; time.sleep(5)"], start_new_session=True)
EOF
cat >"$impact_repo/scripts/read_system.py" <<'EOF'
from pathlib import Path
print(Path("/etc/hosts").read_text())
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
sock = socket.socket()
sock.bind(('127.0.0.1', 0))
raise SystemExit(77)
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
  "schema_version":"polymetrics.ai/pm-review-system/v4",
  "owner":"parent_orchestrator",
  "canonical_roots":[".pi/prompts/canonical.md"],
  "explicit_reference_files":["AGENTS.md"],
  "reference_prefixes":[".agents/",".pi/",".planning/","scripts/","cmd/","internal/"],
  "ignored_reference_prefixes":[],
  "prohibited_active_targets":[],
  "impact_graph":{
    "index_prefixes":[".agents/",".pi/",".planning/","scripts/","cmd/","internal/"],
    "max_index_files":200,"max_index_bytes":2000000,"max_nodes":300,"max_edges":1000,
    "max_traversal_states":1000,"max_depth":12,"max_impact_files":200,"max_impact_edges":1000,
    "go_command_timeout_seconds":20,"go_max_output_bytes":8388608,"go_max_packages":2000,
    "packet_max_impact_files":10,"packet_max_impact_edges":40,"max_packets":64,
    "edge_context_max_bytes_per_file":4096,"packet_max_file_slice_bytes":4096,
    "default_relation_policy":{"upstream":1,"downstream":1,"lateral":1,"temporal":1},
    "relation_policy":{
      "required_reference":{"upstream":4,"downstream":4,"lateral":0,"temporal":0},
      "script_invokes":{"upstream":1,"downstream":1,"lateral":0,"temporal":0}
    }
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
    "response_reserve_tokens":6000,"context_window_tokens":100000
  },
  "domain_rules":[
    {"domain":"implementation_test","patterns":["scripts/**","cmd/**","internal/**"]},
    {"domain":"architecture_reference","patterns":[".agents/**",".pi/**"]},
    {"domain":"authority_workflow_state","patterns":[".planning/**"]}
  ],
  "packet_invariants":{
    "architecture_reference":["impact_complete"],"authority_workflow_state":["impact_complete"],
    "implementation_test":["impact_complete"],"impact_graph":["impact_complete"],"combined":["impact_complete"]
  }
}
JSON
cat >"$impact_repo/.agents/review-scope.json" <<'JSON'
{
  "schema_version":"polymetrics.ai/pm-review-scope/v1",
  "issue":397,
  "review_round":3,
  "candidate_lineage":"fixture-impact",
  "allowed_changed_paths":["**"],
  "forbidden_changed_paths":[]
}
JSON
python3 - "$impact_repo/.agents/review-config.json" "$impact_repo/.agents/bound-config.json" "$impact_repo/.agents/legacy-config.json" "$impact_repo/.agents/index-bound-config.json" <<'PY'
import json,sys
normal=json.load(open(sys.argv[1]))
bound=json.loads(json.dumps(normal)); bound["impact_graph"]["max_impact_files"]=3
open(sys.argv[2],"w").write(json.dumps(bound,indent=2)+"\n")
legacy=json.loads(json.dumps(normal)); legacy["schema_version"]="polymetrics.ai/pm-review-system/v3"
open(sys.argv[3],"w").write(json.dumps(legacy,indent=2)+"\n")
index_bound=json.loads(json.dumps(normal)); index_bound["impact_graph"]["max_index_files"]=1
open(sys.argv[4],"w").write(json.dumps(index_bound,indent=2)+"\n")
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
impact_trust="$test_tmp/impact-trust.json"
write_trust_bundle "$impact_repo" "$impact_base" "$impact_head" .agents/review-config.json .agents/review-scope.json "$impact_trust"
impact_manifest="$test_tmp/impact-manifest.json"
python3 "$repo_root/scripts/pm-review-system.py" compile --repo-root "$impact_repo" \
  --config .agents/review-config.json --scope .agents/review-scope.json --base "$impact_base" --head "$impact_head" \
  --trust-bundle "$impact_trust" >"$impact_manifest"
impact_status=$?
if [[ $impact_status -ne 0 ]] || ! python3 - "$impact_manifest" <<'PY'
import json,sys
value=json.load(open(sys.argv[1]))
assert value["schema_version"] == "polymetrics.ai/pm-review-compile/v4"
graph=value["impact_graph"]
files=set(value["coverage_manifest"]["impact_files"])
required={
 "AGENTS.md", ".agents/templates/leaf.md", ".pi/prompts/upstream.md", "scripts/tool.py", "scripts/helper.py",
 ".agents/mixed/a.md", ".agents/mixed/b.md", "scripts/mixed.py",
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
assert all(edge["revision"] in {"base","head"} and edge["source_blob_sha256"] is not None or edge["parser"] in {"config","go_list"} for edge in graph["edges"])
relations={edge["relation"] for edge in graph["edges"]}
assert {"required_reference","python_import","authority_writes","authority_reads","authority_mirror","generates","go_imports","go_test","platform_variant"} <= relations
packet_files={path for packet in value["packets"] for path in packet["impact_files"]}
packet_slice_files={item["path"] for packet in value["packets"] for field in ("impact_file_slices","edge_context_slices") for item in packet.get(field,[])}
packet_edges={edge for packet in value["packets"] for edge in packet["impact_edge_ids"]}
assert packet_files == files
assert packet_slice_files == files
assert packet_edges == set(value["coverage_manifest"]["impact_edge_ids"])
for packet in value["packets"]:
    endpoint_files={endpoint for edge in packet.get("impact_edges",[]) for endpoint in (edge["source"],edge["target"]) if not endpoint.startswith("go-package:")}
    assert endpoint_files <= set(packet.get("edge_context_files",[])), (packet["packet_id"], sorted(endpoint_files-set(packet.get("edge_context_files",[]))))
    assert packet["context"].get("estimation") == "complete_rendered_prompt_one_token_per_byte"
    assert packet["context"].get("bytes_per_token_upper_bound") == 1
    assert packet["context"].get("response_reserve_tokens") > 0
PY
then
  fail "changed-file-seeded bidirectional typed impact graph, policy-state traversal, or coherent exact packet coverage is absent"
fi

# Unsafe Markdown and active planning/test targets must not disappear from the graph merely because
# the parser rejects a path or the source lives under .planning/scripts/tests.
impact_branch="$(git -C "$impact_repo" symbolic-ref --short HEAD)"
git -C "$impact_repo" checkout -qb unsafe-impact "$impact_head"
cat >"$impact_repo/.planning/unsafe.md" <<'EOF'
Required unsafe target: [escape](../../outside.py)
Required missing target: `scripts/required-missing.py`
EOF
git -C "$impact_repo" add .planning/unsafe.md
git -C "$impact_repo" -c user.name='PM Review Test' -c user.email='pm-review-test@example.invalid' commit -qm unsafe-impact
unsafe_impact_head="$(git -C "$impact_repo" rev-parse HEAD)"
unsafe_impact_trust="$test_tmp/unsafe-impact-trust.json"
write_trust_bundle "$impact_repo" "$impact_head" "$unsafe_impact_head" .agents/review-config.json .agents/review-scope.json "$unsafe_impact_trust"
python3 "$repo_root/scripts/pm-review-system.py" compile --repo-root "$impact_repo" \
  --config .agents/review-config.json --scope .agents/review-scope.json --base "$impact_head" --head "$unsafe_impact_head" \
  --trust-bundle "$unsafe_impact_trust" >"$test_tmp/unsafe-impact.json"
unsafe_impact_status=$?
if [[ $unsafe_impact_status -eq 0 ]] || ! python3 - "$test_tmp/unsafe-impact.json" <<'PY'
import json,sys
value=json.load(open(sys.argv[1])); claims="\n".join(item.get("claim","") for item in value.get("findings",[]))
assert value["status"]=="blocked"
assert "escapes" in claims or "unsafe" in claims or "repository-relative" in claims
assert "required-missing.py" in claims
PY
then
  fail "unsafe Markdown or active missing planning target did not fail closed"
fi
git -C "$impact_repo" checkout -q "$impact_branch"

# Non-head filesystem reads, diverged comparison bases, and operational index limits must block.
python3 "$repo_root/scripts/pm-review-system.py" compile --repo-root "$impact_repo" \
  --config .agents/review-config.json --scope .agents/review-scope.json --base "$impact_base" --head "$impact_base" --allow-non-head \
  --trust-bundle "$impact_trust" \
  >"$test_tmp/non-head.json" 2>"$test_tmp/non-head.stderr"
non_head_status=$?
if [[ $non_head_status -eq 0 ]]; then
  fail "allow-non-head compiled worktree content under a different advertised head"
fi
git -C "$impact_repo" checkout -qb divergent-base "$impact_base"
printf 'divergent\n' >"$impact_repo/.agents/divergent.md"
git -C "$impact_repo" add .agents/divergent.md
git -C "$impact_repo" -c user.name='PM Review Test' -c user.email='pm-review-test@example.invalid' commit -qm divergent
divergent_base="$(git -C "$impact_repo" rev-parse HEAD)"
git -C "$impact_repo" checkout -q "$impact_branch"
python3 "$repo_root/scripts/pm-review-system.py" compile --repo-root "$impact_repo" \
  --config .agents/review-config.json --scope .agents/review-scope.json --base "$divergent_base" --head "$impact_head" \
  --trust-bundle "$impact_trust" >"$test_tmp/divergent-base.json"
divergent_status=$?
if [[ $divergent_status -eq 0 ]]; then
  fail "divergent supplied base was silently replaced by merge-base semantics"
fi
index_bound_trust="$test_tmp/index-bound-trust.json"
write_trust_bundle "$impact_repo" "$impact_base" "$impact_head" .agents/index-bound-config.json .agents/review-scope.json "$index_bound_trust"
python3 "$repo_root/scripts/pm-review-system.py" compile --repo-root "$impact_repo" \
  --config .agents/index-bound-config.json --scope .agents/review-scope.json --base "$impact_base" --head "$impact_head" \
  --trust-bundle "$index_bound_trust" >"$test_tmp/index-bound.json"
index_bound_status=$?
if [[ $index_bound_status -eq 0 ]] || ! python3 - "$test_tmp/index-bound.json" <<'PY'
import json,sys
value=json.load(open(sys.argv[1])); assert value["status"]=="blocked"; assert value["impact_graph"]["universe"]["index_file_count"] <= 1
PY
then
  fail "max_index_files did not prevent broad index materialization"
fi
# External-module Go indexing must use a scrubbed pre-populated read-only cache without network, and
# deleted Go seeds must retain base package/test/importer context.
module_cache="$test_tmp/go-mod-cache"
module_proxy="$test_tmp/go-module-proxy"
mkdir -p "$module_cache" "$module_proxy/example.test/external/@v"
cat >"$module_proxy/example.test/external/@v/v1.0.0.mod" <<'EOF'
module example.test/external

go 1.24
EOF
printf '{"Version":"v1.0.0","Time":"2026-01-01T00:00:00Z"}\n' >"$module_proxy/example.test/external/@v/v1.0.0.info"
printf 'v1.0.0\n' >"$module_proxy/example.test/external/@v/list"
python3 - "$module_proxy/example.test/external/@v/v1.0.0.zip" <<'PY'
import sys,zipfile
with zipfile.ZipFile(sys.argv[1],'w') as archive:
 archive.writestr('example.test/external@v1.0.0/go.mod','module example.test/external\n\ngo 1.24\n')
 archive.writestr('example.test/external@v1.0.0/external.go','package external\nconst Name = "external"\n')
PY
GOMODCACHE="$module_cache" GOPROXY="file://$module_proxy" GOSUMDB=off GOENV=off \
  go mod download example.test/external@v1.0.0
git -C "$impact_repo" checkout -qb external-go "$impact_head"
cat >>"$impact_repo/go.mod" <<'EOF'

require example.test/external v1.0.0
EOF
python3 - "$impact_repo/internal/lib/lib.go" <<'PY'
from pathlib import Path
import sys
p=Path(sys.argv[1]); text=p.read_text(); p.write_text(text.replace('package lib\n','package lib\nimport "example.test/external"\n').replace('return "v2"','return external.Name'))
PY
git -C "$impact_repo" add go.mod internal/lib/lib.go
git -C "$impact_repo" -c user.name='PM Review Test' -c user.email='pm-review-test@example.invalid' commit -qm external-go
external_head="$(git -C "$impact_repo" rev-parse HEAD)"
external_trust="$test_tmp/external-trust.json"
write_trust_bundle "$impact_repo" "$impact_head" "$external_head" .agents/review-config.json .agents/review-scope.json "$external_trust"
GOMODCACHE="$module_cache" python3 "$repo_root/scripts/pm-review-system.py" compile --repo-root "$impact_repo" \
  --config .agents/review-config.json --scope .agents/review-scope.json --base "$impact_head" --head "$external_head" \
  --trust-bundle "$external_trust" >"$test_tmp/external-go.json"
external_status=$?
if [[ $external_status -ne 0 ]] || ! python3 - "$test_tmp/external-go.json" <<'PY'
import json,sys
value=json.load(open(sys.argv[1])); assert value["impact_graph"]["go_context"]["status"]=="complete"
PY
then
  fail "offline external-module Go impact indexing did not use the pre-populated cache"
fi
git -C "$impact_repo" checkout -q "$impact_branch"
git -C "$impact_repo" checkout -qb deleted-go "$impact_head"
git -C "$impact_repo" rm -q internal/lib/lib.go
git -C "$impact_repo" -c user.name='PM Review Test' -c user.email='pm-review-test@example.invalid' commit -qm deleted-go
deleted_head="$(git -C "$impact_repo" rev-parse HEAD)"
deleted_trust="$test_tmp/deleted-trust.json"
write_trust_bundle "$impact_repo" "$impact_head" "$deleted_head" .agents/review-config.json .agents/review-scope.json "$deleted_trust"
python3 "$repo_root/scripts/pm-review-system.py" compile --repo-root "$impact_repo" \
  --config .agents/review-config.json --scope .agents/review-scope.json --base "$impact_head" --head "$deleted_head" \
  --trust-bundle "$deleted_trust" >"$test_tmp/deleted-go.json"
deleted_status=$?
if [[ $deleted_status -ne 0 ]] || ! python3 - "$test_tmp/deleted-go.json" <<'PY'
import json,sys
value=json.load(open(sys.argv[1])); files=set(value["coverage_manifest"]["impact_files"])
assert {"internal/lib/lib.go","internal/lib/lib_test.go","cmd/app/main.go"} <= files
PY
then
  fail "deleted Go seed lost base package tests or reverse importer context"
fi
git -C "$impact_repo" checkout -q "$impact_branch"

for config_case in bound-config legacy-config; do
  case_trust="$test_tmp/$config_case-trust.json"
  write_trust_bundle "$impact_repo" "$impact_base" "$impact_head" ".agents/$config_case.json" .agents/review-scope.json "$case_trust"
  python3 "$repo_root/scripts/pm-review-system.py" compile --repo-root "$impact_repo" \
    --config ".agents/$config_case.json" --scope .agents/review-scope.json --base "$impact_base" --head "$impact_head" \
    --trust-bundle "$case_trust" >"$test_tmp/$config_case-result.json"
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
      "schema_version":"polymetrics.ai/pm-review-lab-request/v2","hypothesis_id":name,
      "claim":"temporary accept behavior must make the targeted check fail",
      "alternative":"the behavior change is irrelevant to the targeted check",
      "impact_edges_examined":["edge-fixture-1"],
      "temporary_change":"replace reject with accept in the disposable copy",
      "changes":[{"path":"behavior.txt","find":"reject\n","replace":"accept\n"}],
      "command":["python3","scripts/check_behavior.py"],
      "expected_discriminator":{"exit_code":23}
    }
    value.update(overrides); path=root/f"{name}.json"; path.write_text(json.dumps(value)); return path

def run(path, extra_env=None, approvals=None):
    env=os.environ.copy(); env.update(extra_env or {})
    argv=[sys.executable,script,"run","--repo-root",repo,"--base",base,"--head",head,"--packet-id",path.stem,"--request",str(path),"--temp-root",str(labs)]
    for approval in approvals or []: argv.extend(["--approval",str(approval)])
    proc=subprocess.run(argv,text=True,capture_output=True,env=env)
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

# Explicit human-like modes: read, disposable writes/dummy SQLite, compiler cache, local service,
# counterfactual diff, and trusted denial evidence. All effects remain inside each disposable lab.
def explicit(name,mode,command,expected={"exit_code":0},**overrides):
    roots=["repo","home","tmp","xdg_config","xdg_cache","go_cache","go_mod_cache","services","credentials"]
    value={"mode":mode,"allowed_read_roots":roots,
           "allowed_write_roots":[] if mode=="read_inspect" else ["repo","home","tmp","xdg_config","xdg_cache","go_cache","go_mod_cache","services"],
           "changes":[],"command":command,"expected_discriminator":expected}
    value.update(overrides); return request(name,**value)
code,read_evidence=run(explicit("read-inspect","read_inspect",["python3","scripts/check_behavior.py"]))
assert code==0 and read_evidence["status"]=="evidence" and read_evidence["experiment"]["mode"]=="read_inspect"
db_code="import os,sqlite3; p=os.path.join(os.environ['HOME'],'dummy.db'); c=sqlite3.connect(p); c.execute('create table t(v)'); c.execute('insert into t values (7)'); c.commit(); assert c.execute('select v from t').fetchone()[0]==7"
code,db_evidence=run(explicit("dummy-db","disposable_write_test",["python3","-c",db_code]))
assert code==0 and db_evidence["status"]=="evidence" and any("dummy.db" in path for path in db_evidence["experiment"]["actual_effects"]["lab_owned_roots"]["changed_paths"])
code,cache_evidence=run(explicit("compile-cache","disposable_write_test",["go","test","./internal/lib"]))
assert code==0 and cache_evidence["status"]=="evidence", cache_evidence
server=("import socket; s=socket.socket(); s.setsockopt(socket.SOL_SOCKET,socket.SO_REUSEADDR,1); "
        "s.bind(('127.0.0.1',{service.http.port})); s.listen(); "
        "exec(\"while True:\\n c,_=s.accept(); c.recv(4096); "
        "c.sendall(b'HTTP/1.0 200 OK\\\\r\\\\nContent-Length: 2\\\\r\\\\n\\\\r\\\\nok'); c.close()\")")
service={"id":"http","transport":"tcp","host":"127.0.0.1","port":0,
         "command":["python3","-c",server],"readiness":"connect"}
client="import urllib.request; assert urllib.request.urlopen('http://{service.http.endpoint}/',timeout=2).status==200"
code,service_evidence=run(explicit("local-service","local_service_simulation",["python3","-c",client],services=[service]))
assert code==0 and service_evidence["status"]=="evidence" and service_evidence["experiment"]["services"][0]["local_only"], service_evidence
code,caught=run(explicit("caught-denial","read_inspect",["python3","scripts/check_isolation.py"]))
assert code!=0 and caught["status"]=="blocked" and caught["experiment"]["observed"]["sandbox_denial_observed"]
assert not any(labs.iterdir())

def blocked(name, **changes):
    code,value=run(request(name,**changes)); assert code!=0 and value["status"]=="blocked"; return value
blocked('candidate-write', change_scope='candidate')
blocked('outside-write', changes=[{"path":"../outside","find":"x","replace":"y"}])
blocked('symlink-write', changes=[{"path":"escape-link","find":"x","replace":"y"}])
blocked('git-admin-change', changes=[{"path":".git/config","find":"repositoryformatversion = 0","replace":"repositoryformatversion = 1"}])
outside=pathlib.Path(outside_arg); outside.write_text('sentinel')
blocked('nested-outside-write',changes=[],command=['python3','scripts/outside_write.py'],expected_discriminator={'exit_code':0})
assert outside.read_text()=='sentinel'
blocked('nested-outside-read',changes=[],command=['python3','scripts/outside_read.py'],expected_discriminator={'exit_code':0})
blocked('nonlab-system-read',changes=[],command=['python3','scripts/read_system.py'],expected_discriminator={'exit_code':0})
blocked('escaped-process-group',changes=[],command=['python3','scripts/escape_process.py'],expected_discriminator={'exit_code':0})
blocked('nested-network',changes=[],command=['python3','scripts/network.py'],expected_discriminator={'exit_code':77})
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
      request('parallel-b',changes=[],command=['python3','scripts/check_env.py'],expected_discriminator={"exit_code":0})
    ]))
assert all(code==0 and value['status']=='evidence' for code,value in rows)
assert not any(labs.iterdir())
assert subprocess.check_output(['git','-C',repo,'rev-parse','HEAD'],text=True).strip()==head
assert not subprocess.check_output(['git','-C',repo,'status','--porcelain'],text=True)
PY
then
  fail "counterfactual lab did not enforce exact-head isolation, denial, bounds, evidence, concurrency, and cleanup"
fi

# Current route siblings and durable authority/state evidence must agree before compilation. These
# are narrow parity checks and intentionally exclude every explicit PR #493-owned path.
if ! python3 - "$repo_root" <<'PY'
import json, pathlib, sys
root=pathlib.Path(sys.argv[1])
route_paths=[
 '.agents/agentic-delivery/contracts/parent-issue-roadmap-template.md',
 '.agents/agentic-delivery/contracts/issue-prompt-template.md',
 '.agents/agentic-delivery/agents/implementation/issue-first-implementation-agent.agent.yaml',
 '.agents/connector-migration/rollout-checklist.md',
 '.agents/connector-migration/validation-gates.md',
 '.planning/traces/cli-architecture-v2-pi-prompts.md',
]
for relative in route_paths:
 text=(root/relative).read_text().lower()
 assert 'claude-review-loop.md' not in text, relative
 assert 'copilot_backup' not in text and 'copilot backup' not in text, relative
state=json.loads((root/'.planning/phases/397-cli-architecture-v2-orchestration/RUN-STATE.json').read_text())
assert state['currentHead']=='0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20'
phase=json.loads((root/'.planning/phases/397-pm-first-round-review-system-r1/RUN-STATE.json').read_text())
assert phase['schemaVersion']=='pm-review-system-phase/v2'
assert (root/'.agents/agentic-delivery/schemas/pm-review-system-phase-state.schema.json').is_file()
workflow=(root/'.agents/agentic-delivery/workflows/local-codex-review-loop.md').read_text().lower()
assert 'remote head equal' not in workflow and 'gh-axi' not in workflow
PY
then
  fail "current PM route, post-PR-495 authority, phase schema, or no-network reviewer identity is inconsistent"
fi

# Compile the exact committed range. Output must be one non-TTY JSON envelope containing paths and
# metadata only, with complete packet assignment and no environment-value leakage.
system_repo="$test_tmp/system-repo"
git clone --no-hardlinks --quiet "$repo_root" "$system_repo"
git -C "$repo_root" diff HEAD --binary -- \
  scripts/pm-review-system.py scripts/pm-review-lab.py scripts/tests/pm-review-system.sh \
  .agents/agentic-delivery/contracts/pm-review-system.json \
  .agents/agentic-delivery/contracts/pm-review-packet-template.md \
  .agents/agentic-delivery/prompts/local-codex-review-prompt.md \
  .agents/agentic-delivery/schemas/orchestration-state.schema.yaml \
  .agents/agentic-delivery/schemas/pm-review-system-phase-state.schema.json \
  .agents/agentic-delivery/workflows/local-codex-review-loop.md \
  .planning/phases/397-pm-first-round-review-system-r1/REVIEW-SCOPE.json \
  .pi/agents/pm-reviewer.md .pi/prompts/pm-review-loop.md | git -C "$system_repo" apply
# Materialize the frozen dedup manifest from a test-owned literal. Its hash-bound R1/R2 sources
# already come from the isolated clone's immutable Git objects; no mutable outer fixture is read.
cat >"$system_repo/scripts/tests/fixtures/pm-review-system/dedup-corpus-manifest.json" <<'JSON'
{
  "schema_version": "polymetrics.ai/pm-review-dedup-corpus-manifest/v1",
  "measurement_policy": "R1 and R2 are retrospective development only because R2 labels were inspected during tuning; freeze deterministic-partial-keys/v1 before prospective R3",
  "raw_response_limitation": "Hash-verified raw response objects are unavailable in Git. Candidate evaluation uses only committed table fields and makes no semantic/causal acceptance claim.",
  "sources": {
    "R1": {
      "path": ".planning/phases/397-pm-first-round-review-system-r1/REVIEW-R1-DISPOSITION.md",
      "bytes": 23593,
      "sha256": "9a745178167768606c64fd3faa13603b17f92568e9de6377708baffc44133134",
      "role": "retrospective_development"
    },
    "R2": {
      "path": ".planning/phases/397-pm-first-round-review-system-r1/REVIEW-R2-DISPOSITION.md",
      "bytes": 34633,
      "sha256": "caf455bd48949df5e2d0829be930674bd3f0af0d947e89dfa158a0dd13b2fa80",
      "role": "retrospective_development_labels_inspected"
    }
  },
  "next_untouched_measurement": "fresh prospective exact-head R3"
}
JSON
# Test the immutable compiler against an exact commit containing the current uncommitted treatment;
# the user's worktree remains uncommitted and untouched.
git -C "$system_repo" add .
git -C "$system_repo" -c user.name='PM Review Test' -c user.email='pm-review-test@example.invalid' commit -qm system-green
head_sha="$(git -C "$system_repo" rev-parse HEAD)"
system_trust="$test_tmp/system-trust.json"
write_trust_bundle "$system_repo" 0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20 "$head_sha" \
  .agents/agentic-delivery/contracts/pm-review-system.json \
  .planning/phases/397-pm-first-round-review-system-r1/REVIEW-SCOPE.json "$system_trust"
export PM_REVIEW_TRUST_BUNDLE="$system_trust"
compile_output="$test_tmp/compile-manifest.tmp.json"
PM_REVIEW_SECRET_SENTINEL='do-not-copy-this-environment-value' \
  python3 "$repo_root/scripts/pm-review-system.py" compile \
    --repo-root "$system_repo" \
    --scope .planning/phases/397-pm-first-round-review-system-r1/REVIEW-SCOPE.json \
    --base 0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20 \
    --head "$head_sha" --trust-bundle "$system_trust" >"$compile_output"
compile_status=$?
if [[ $compile_status -ne 0 ]]; then
  fail "exact-head compiler blocked the allowlisted current range"
elif grep -Fq 'do-not-copy-this-environment-value' "$compile_output"; then
  fail "compiler leaked an environment value"
elif ! python3 - "$compile_output" <<'PY'
import json
import sys

document = json.load(open(sys.argv[1]))
assert document["schema_version"] == "polymetrics.ai/pm-review-compile/v4"
assert document["status"] == "ready"
assert document["selection"] in {"combined", "split"}
assert document["content_policy"].startswith("paths, exact revision/blob/slice metadata")
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
assert document["source_mode"] == "detached_exact_commit_snapshot"
assert document["config"]["sha256"] and document["scope"]["sha256"]
assert set(document["authentication"]) == {"algorithm","coverage_sha256","packets_sha256","semantic_manifest_sha256"}
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
    assert {item["path"] for item in packet.get("impact_file_slices", [])} <= set(packet["impact_files"])
    assert {item["path"] for item in packet.get("changed_file_slices", [])} <= set(packet["changed_files"])
    assert {item["path"] for item in packet.get("edge_context_slices", [])} <= set(packet["edge_context_files"])
    all_slices=packet.get("changed_file_slices",[])+packet.get("context_file_slices",[])+packet.get("impact_file_slices",[])+packet.get("edge_context_slices",[])
    assert all(item["revision"] in {document["exact_base_sha"],document["exact_head_sha"]} and item["blob_sha256"] and item["sha256"] for item in all_slices)
    assert packet["context"]["bytes_per_token_upper_bound"] == 1
    assert packet["context"]["response_reserve_tokens"] > 0
    assert packet["context"]["total_context_upper_bound"] == packet["context"]["rendered_prompt_bytes"] + packet["context"]["response_reserve_tokens"]
    assert not packet["context"]["overflow"]
    assert not packet["context"]["truncated"]
PY
then
  fail "compiled JSON envelope or packet coverage is invalid"
fi

# Malformed identities and escaping/symlinked config paths stop without reading broad files.
unsafe_dir="$system_repo/pm-review-evidence.tmp"
mkdir -p "$unsafe_dir"
cp "$compile_output" "$unsafe_dir/render-manifest.json"
render_packet_id="$(python3 - "$compile_output" <<'PY'
import json,sys
print(json.load(open(sys.argv[1]))["packets"][0]["packet_id"])
PY
)"
render_expected_bytes="$(python3 - "$compile_output" "$render_packet_id" <<'PY'
import json,sys
value=json.load(open(sys.argv[1]))
print(next(packet for packet in value["packets"] if packet["packet_id"] == sys.argv[2])["context"]["rendered_prompt_bytes"])
PY
)"
render_output="$test_tmp/rendered-packet.prompt"
python3 "$repo_root/scripts/pm-review-system.py" render --repo-root "$system_repo" \
  --manifest pm-review-evidence.tmp/render-manifest.json --packet-id "$render_packet_id" \
  --trust-bundle "$system_trust" >"$render_output"
render_status=$?
if [[ $render_status -ne 0 || ! -s "$render_output" ]] ||
  [[ "$(wc -c <"$render_output" | tr -d ' ')" != "$render_expected_bytes" ]] ||
  ! grep -Fq 'EXACT SLICE PAYLOADS (canonical descriptor order)' "$render_output"
then
  fail "authenticated renderer did not emit the exactly accounted bounded packet prompt"
fi
cp "$unsafe_dir/render-manifest.json" "$unsafe_dir/tampered-render-manifest.json"
python3 - "$unsafe_dir/tampered-render-manifest.json" <<'PY'
import json,sys
path=sys.argv[1]
value=json.load(open(path))
value["packets"][0]["context"]["rendered_prompt_bytes"] += 1
def digest(item):
    return hashlib.sha256(json.dumps(item,sort_keys=True,separators=(",",":"),ensure_ascii=True).encode()).hexdigest()
import hashlib
unsigned={key:item for key,item in value.items() if key != "authentication"}
value["authentication"]={
    "algorithm":"sha256-canonical-json-v1",
    "coverage_sha256":digest(value["coverage_manifest"]),
    "packets_sha256":digest(value["packets"]),
    "semantic_manifest_sha256":digest(unsigned),
}
open(path,"w").write(json.dumps(value)+"\n")
PY
if python3 "$repo_root/scripts/pm-review-system.py" render --repo-root "$system_repo" \
  --manifest pm-review-evidence.tmp/tampered-render-manifest.json --packet-id "$render_packet_id" \
  --trust-bundle "$system_trust" >"$test_tmp/tampered-render.prompt" 2>"$test_tmp/tampered-render.stderr"
then
  fail "renderer accepted a manifest whose packet authentication was tampered"
fi
ln -s /etc/passwd "$unsafe_dir/escape.json"
unsafe_relative="$(python3 - "$system_repo" "$unsafe_dir/escape.json" <<'PY'
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
  python3 "$repo_root/scripts/pm-review-system.py" compile --repo-root "$system_repo" \
    --scope .planning/phases/397-pm-first-round-review-system-r1/REVIEW-SCOPE.json \
    $unsafe_args --head "$head_sha" --trust-bundle "$system_trust" >"$unsafe_dir/result.json" 2>"$unsafe_dir/error.log"
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
        "schema_version": "polymetrics.ai/pm-review-packet-response/v4",
        "packet_id": packet["packet_id"],
        "exact_base_sha": manifest["exact_base_sha"],
        "exact_head_sha": manifest["exact_head_sha"],
        "exact_head_tree": manifest["exact_head_tree"],
        "status": "clean",
        "reviewed_files": packet["changed_files"],
        "changed_file_slices": packet.get("changed_file_slices", []),
        "closure_files": packet["closure_files"],
        "authority_files": packet["authority_files"],
        "context_file_slices": packet.get("context_file_slices", []),
        "impact_files": packet.get("impact_files", []),
        "impact_edge_ids": packet.get("impact_edge_ids", []),
        "impact_file_slices": packet.get("impact_file_slices", []),
        "edge_context_files": packet.get("edge_context_files", []),
        "edge_context_slices": packet.get("edge_context_slices", []),
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
            "hypotheses": [{"id":"H1","claim":"fixture coverage is complete","strongest_alternative":"an assigned item is omitted","falsifier":"any exact assignment differs","evidence_paths":["scripts/tests/pm-review-system.sh"]}],
            "disconfirming_evidence": "exact assigned sets and fixture contract agree"
        },
        "experiments": [],
        "no_experiment_reason": "static fixture evidence is decisive",
        "findings": [],
        "residual_risk": [],
        "context": {"input_tokens": None, "output_tokens": None, "cost": None, "overflow": False, "truncated": False},
        "wall_clock_ms": None,
    }
    (root / f"{packet['packet_id']}.json").write_text(json.dumps(response))
PY
manifest_relative="$(python3 - "$system_repo" "$unsafe_dir/manifest.json" <<'PY'
import os,sys
print(os.path.relpath(sys.argv[2], sys.argv[1]))
PY
)"
responses_relative="$(python3 - "$system_repo" "$responses_dir" <<'PY'
import os,sys
print(os.path.relpath(sys.argv[2], sys.argv[1]))
PY
)"
synthesis_output="$unsafe_dir/synthesis.json"
python3 "$repo_root/scripts/pm-review-system.py" synthesize --repo-root "$system_repo" \
  --manifest "$manifest_relative" --responses-dir "$responses_relative" >"$synthesis_output"
synthesis_status=$?
if [[ $synthesis_status -ne 0 ]] || ! python3 - "$synthesis_output" <<'PY'
import json,sys
value=json.load(open(sys.argv[1]))
assert value["status"] == "clean"
assert value["owner"] == "parent_orchestrator"
assert value["shepherd"]["status"] == "pending"
assert value["human_merge_authority"] is True
assert value["dedup"]["schema_version"] == "polymetrics.ai/pm-review-dedup/v1"
assert value["dedup"]["raw_finding_count"] == 0
assert value["dedup"]["disclosure"]["retained_observation_count"] == 0
assert value["telemetry"]["claim"].startswith("only validated")
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
clean_backup="$unsafe_dir/clean-response.backup"
cp "$first_response" "$clean_backup"

# Status/content and invariant semantics are distinct from missing coverage. Malformed status blocks;
# an explicitly failed invariant paired with a finding enters the correction loop, not a blocker.
python3 - "$first_response" <<'PY'
import json,sys
path=sys.argv[1]; value=json.load(open(path)); value["status"]="findings"; value["findings"]=[]
open(path,"w").write(json.dumps(value))
PY
python3 "$repo_root/scripts/pm-review-system.py" synthesize --repo-root "$system_repo" \
  --manifest "$manifest_relative" --responses-dir "$responses_relative" >"$synthesis_output"
status_mismatch=$?
if [[ $status_mismatch -eq 0 ]] || ! python3 - "$synthesis_output" <<'PY'
import json,sys
value=json.load(open(sys.argv[1])); assert value["status"]=="blocked"; assert any("status" in item["claim"] for item in value["blockers"])
PY
then
  fail "findings status with empty findings did not block"
fi
cp "$clean_backup" "$first_response"
python3 - "$first_response" <<'PY'
import json,sys
path=sys.argv[1]; value=json.load(open(path)); value["status"]="findings"; value["invariants"][0]["status"]="fail"
value["findings"]=[{"severity":"high","category":"fixture_invariant","path":"scripts/pm-review-system.py","line":"1","evidence":"fixture failed invariant","impact":"fixture impact","smallest_safe_correction":"fixture correction"}]
open(path,"w").write(json.dumps(value))
PY
python3 "$repo_root/scripts/pm-review-system.py" synthesize --repo-root "$system_repo" \
  --manifest "$manifest_relative" --responses-dir "$responses_relative" >"$synthesis_output"
failed_invariant_status=$?
if [[ $failed_invariant_status -eq 0 ]] || ! python3 - "$synthesis_output" <<'PY'
import json,sys
value=json.load(open(sys.argv[1])); assert value["status"]=="findings_correction_required"; assert value["findings"] and not value["blockers"]
assert value["dedup"]["raw_finding_count"] == len(value["findings"]) == 1
assert len(value["dedup"]["observations"]) == 1
PY
then
  fail "failed invariant paired with a finding did not synthesize findings_correction_required"
fi
cp "$clean_backup" "$first_response"
python3 - "$first_response" <<'PY'
import json,sys
path=sys.argv[1]; value=json.load(open(path)); value["impact_edge_ids"].append("edge-unassigned")
open(path,"w").write(json.dumps(value))
PY
python3 "$repo_root/scripts/pm-review-system.py" synthesize --repo-root "$system_repo" \
  --manifest "$manifest_relative" --responses-dir "$responses_relative" >"$synthesis_output"
extra_coverage_status=$?
if [[ $extra_coverage_status -eq 0 ]] || ! python3 - "$synthesis_output" <<'PY'
import json,sys
value=json.load(open(sys.argv[1])); assert value["status"]=="blocked"; assert any(item["category"]=="packet_coverage" for item in value["blockers"])
PY
then
  fail "unassigned extra packet coverage synthesized clean"
fi
cp "$clean_backup" "$first_response"

# A blocked compile manifest and mutually matching stale fake identities can never synthesize clean.
cp "$unsafe_dir/manifest.json" "$unsafe_dir/blocked-manifest.json"
python3 - "$unsafe_dir/blocked-manifest.json" <<'PY'
import json,sys
path=sys.argv[1]; value=json.load(open(path)); value["status"]="blocked"; value["findings"]=[{"category":"fixture","claim":"compile blocked"}]
open(path,"w").write(json.dumps(value))
PY
blocked_manifest_relative="$(python3 - "$system_repo" "$unsafe_dir/blocked-manifest.json" <<'PY'
import os,sys
print(os.path.relpath(sys.argv[2],sys.argv[1]))
PY
)"
python3 "$repo_root/scripts/pm-review-system.py" synthesize --repo-root "$system_repo" \
  --manifest "$blocked_manifest_relative" --responses-dir "$responses_relative" >"$synthesis_output"
blocked_manifest_status=$?
if [[ $blocked_manifest_status -eq 0 ]] || ! python3 - "$synthesis_output" <<'PY'
import json,sys
value=json.load(open(sys.argv[1])); assert value["status"]=="blocked"; assert any("manifest" in item["claim"] or "compile" in item["claim"] for item in value["blockers"])
PY
then
  fail "blocked compile manifest synthesized clean"
fi
cp "$unsafe_dir/manifest.json" "$unsafe_dir/stale-manifest.json"
cp -R "$responses_dir" "$unsafe_dir/stale-responses"
python3 - "$unsafe_dir/stale-manifest.json" "$unsafe_dir/stale-responses" <<'PY'
import json,pathlib,sys
manifest_path=pathlib.Path(sys.argv[1]); manifest=json.loads(manifest_path.read_text()); fake="f"*40; manifest["exact_head_sha"]=fake; manifest_path.write_text(json.dumps(manifest))
for path in pathlib.Path(sys.argv[2]).glob("*.json"):
 value=json.loads(path.read_text()); value["exact_head_sha"]=fake; path.write_text(json.dumps(value))
PY
stale_manifest_relative="$(python3 - "$system_repo" "$unsafe_dir/stale-manifest.json" <<'PY'
import os,sys
print(os.path.relpath(sys.argv[2],sys.argv[1]))
PY
)"
stale_responses_relative="$(python3 - "$system_repo" "$unsafe_dir/stale-responses" <<'PY'
import os,sys
print(os.path.relpath(sys.argv[2],sys.argv[1]))
PY
)"
python3 "$repo_root/scripts/pm-review-system.py" synthesize --repo-root "$system_repo" \
  --manifest "$stale_manifest_relative" --responses-dir "$stale_responses_relative" >"$synthesis_output"
stale_manifest_status=$?
if [[ $stale_manifest_status -eq 0 ]] || ! python3 - "$synthesis_output" <<'PY'
import json,sys
value=json.load(open(sys.argv[1])); assert value["status"]=="blocked"; assert value["blockers"]
assert {item["category"] for item in value["blockers"]} & {"stale_evidence","synthesis_input","compile_manifest"}
PY
then
  fail "mutually matching stale manifest/responses ignored current candidate identity"
fi

python3 - "$first_response" <<'PY'
import json,sys
path=sys.argv[1]; value=json.load(open(path)); value["schema_version"]="polymetrics.ai/pm-review-packet-response/v2"; open(path,"w").write(json.dumps(value))
PY
python3 "$repo_root/scripts/pm-review-system.py" synthesize --repo-root "$system_repo" \
  --manifest "$manifest_relative" --responses-dir "$responses_relative" >"$synthesis_output"
v1_status=$?
if [[ $v1_status -eq 0 ]] || ! python3 - "$synthesis_output" <<'PY'
import json,sys
value=json.load(open(sys.argv[1])); assert value["status"]=="blocked"; assert any("migration" in item["claim"] for item in value["blockers"])
PY
then
  fail "incompatible v1 packet response did not require explicit migration"
fi
mv "$clean_backup" "$first_response"
python3 - "$first_response" <<'PY'
import json,sys
path=sys.argv[1]; value=json.load(open(path)); value["no_experiment_reason"]=None
value["experiments"]=[{"hypothesis_id":"H-inconclusive","claim":"suspected defect","alternative":"safe behavior","impact_edges_examined":[],"temporary_change":"fixture change","command":["python3","fixture.py"],"expected_discriminator":"different exits","observed":"same exit","supports":"inconclusive","candidate_unchanged":True,"lab_cleanup_verified":True}]
open(path,"w").write(json.dumps(value))
PY
python3 "$repo_root/scripts/pm-review-system.py" synthesize --repo-root "$system_repo" \
  --manifest "$manifest_relative" --responses-dir "$responses_relative" >"$synthesis_output"
inconclusive_status=$?
if [[ $inconclusive_status -eq 0 ]] || ! python3 - "$synthesis_output" <<'PY'
import json,sys
value=json.load(open(sys.argv[1])); assert value["status"]=="blocked"; assert any(item["category"]=="hypothesis_evidence" for item in value["blockers"])
PY
then
  fail "inconclusive counterfactual experiment was allowed to prove clean"
fi
cp "$clean_backup" "$first_response" 2>/dev/null || true
if [[ ! -f "$clean_backup" ]]; then
  python3 - "$compile_output" "$first_response" <<'PY'
# Restore is regenerated from the manifest because the clean backup was moved above.
import json,sys
manifest=json.load(open(sys.argv[1])); packet=next(p for p in manifest['packets'] if p['packet_id'] in sys.argv[2])
value={"schema_version":"polymetrics.ai/pm-review-packet-response/v4","packet_id":packet["packet_id"],"exact_base_sha":manifest["exact_base_sha"],"exact_head_sha":manifest["exact_head_sha"],"exact_head_tree":manifest["exact_head_tree"],"status":"clean","reviewed_files":packet["changed_files"],"changed_file_slices":packet.get("changed_file_slices",[]),"closure_files":packet["closure_files"],"authority_files":packet["authority_files"],"context_file_slices":packet.get("context_file_slices",[]),"impact_files":packet.get("impact_files",[]),"impact_edge_ids":packet.get("impact_edge_ids",[]),"impact_file_slices":packet.get("impact_file_slices",[]),"edge_context_files":packet.get("edge_context_files",[]),"edge_context_slices":packet.get("edge_context_slices",[]),"invariants":[{"id":x,"status":"pass","evidence_paths":[]} for x in packet["invariants"]],"unreviewed_files":[],"review_behaviors":{"impact_model_built_first":True,"directions_traced":["upstream","downstream","lateral","temporal"],"history_inspected":{"status":"not_needed","reason":"fixture"},"sibling_paths_compared":{"status":"not_needed","reason":"fixture"},"hypotheses":[{"id":"H1","claim":"fixture coverage is complete","strongest_alternative":"an assigned item is omitted","falsifier":"any exact assignment differs","evidence_paths":["scripts/tests/pm-review-system.sh"]}],"disconfirming_evidence":"exact assigned sets agree"},"experiments":[],"no_experiment_reason":"static fixture evidence is decisive","findings":[],"residual_risk":[],"context":{"input_tokens":None,"output_tokens":None,"cost":None,"overflow":False,"truncated":False},"wall_clock_ms":None}
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
python3 "$repo_root/scripts/pm-review-system.py" synthesize --repo-root "$system_repo" \
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
# Round-2 recurrence RED: these assertions use independent adversarial oracles rather than the
# compiler's own labels. They intentionally cover the systemic causes behind the 141 exact-head
# findings: authenticated manifest semantics, exact response/lab binding, rendered-prompt bounds,
# structural graph parsing, nested phase state, bounded scoring, and endpoint validation.
if ! python3 - "$system_repo" "$unsafe_dir/manifest.json" "$system_repo/pm-review-r2.tmp" <<'PY'
import argparse
import contextlib
import hashlib
import importlib.util
import io
import json
import math
import os
import pathlib
import shutil
import subprocess
import sys
import time

repo = pathlib.Path(sys.argv[1])
manifest_path = pathlib.Path(sys.argv[2])
tmp = pathlib.Path(sys.argv[3]) / "r2-red"
tmp.mkdir(parents=True)
spec = importlib.util.spec_from_file_location("pm_review_system", repo / "scripts/pm-review-system.py")
pm = importlib.util.module_from_spec(spec)
spec.loader.exec_module(pm)
manifest = json.loads(manifest_path.read_text())
errors = []

def clean_response(packet):
    return {
        "schema_version": pm.PACKET_RESPONSE_SCHEMA,
        "packet_id": packet["packet_id"],
        "exact_base_sha": manifest["exact_base_sha"],
        "exact_head_sha": manifest["exact_head_sha"],
        "exact_head_tree": manifest["exact_head_tree"],
        "status": "clean",
        "reviewed_files": packet["changed_files"],
        "changed_file_slices": packet.get("changed_file_slices", []),
        "closure_files": packet["closure_files"],
        "authority_files": packet["authority_files"],
        "context_file_slices": packet.get("context_file_slices", []),
        "impact_files": packet.get("impact_files", []),
        "impact_edge_ids": packet.get("impact_edge_ids", []),
        "impact_file_slices": packet.get("impact_file_slices", []),
        "edge_context_files": packet.get("edge_context_files", []),
        "edge_context_slices": packet.get("edge_context_slices", []),
        "invariants": [{"id": item, "status": "pass", "evidence_paths": []} for item in packet["invariants"]],
        "unreviewed_files": [],
        "review_behaviors": {
            "impact_model_built_first": True,
            "directions_traced": ["upstream", "downstream", "lateral", "temporal"],
            "history_inspected": {"status": "not_needed", "reason": "fixture is exact"},
            "sibling_paths_compared": {"status": "not_needed", "reason": "fixture has no sibling"},
            "hypotheses": [{"id":"H1","claim":"complete","strongest_alternative":"omitted","falsifier":"set mismatch","evidence_paths":["scripts/tests/pm-review-system.sh"]}],
            "disconfirming_evidence": "exact fixture assignments agree",
        },
        "experiments": [],
        "no_experiment_reason": "static fixture evidence is decisive",
        "findings": [],
        "residual_risk": [],
        "context": {"input_tokens": None, "output_tokens": None, "cost": None, "overflow": False, "truncated": False},
        "wall_clock_ms": None,
    }

def write_responses(directory):
    directory.mkdir()
    for packet in manifest["packets"]:
        (directory / f"{packet['packet_id']}.json").write_text(json.dumps(clean_response(packet)))

def synthesize(candidate_manifest, responses):
    rel_manifest = os.path.relpath(candidate_manifest, repo)
    rel_responses = os.path.relpath(responses, repo)
    proc = subprocess.run(
        [sys.executable, str(repo / "scripts/pm-review-system.py"), "synthesize", "--repo-root", str(repo),
         "--manifest", rel_manifest, "--responses-dir", rel_responses],
        cwd=repo, capture_output=True, text=True,
    )
    try:
        value = json.loads(proc.stdout)
    except json.JSONDecodeError:
        value = None
    return proc, value

# Same-head manifest packet removal must be detected independently of matching responses.
responses = tmp / "responses"
write_responses(responses)
tampered = json.loads(json.dumps(manifest))
removed = tampered["packets"].pop()
def canonical_digest(item):
    return hashlib.sha256(json.dumps(item,sort_keys=True,separators=(",",":"),ensure_ascii=True).encode()).hexdigest()
unsigned = {key:item for key,item in tampered.items() if key != "authentication"}
tampered["authentication"] = {
    "algorithm":"sha256-canonical-json-v1",
    "coverage_sha256":canonical_digest(tampered["coverage_manifest"]),
    "packets_sha256":canonical_digest(tampered["packets"]),
    "semantic_manifest_sha256":canonical_digest(unsigned),
}
(tampered_path := tmp / "tampered-manifest.json").write_text(json.dumps(tampered))
(responses / f"{removed['packet_id']}.json").unlink()
proc, value = synthesize(tampered_path, responses)
if proc.returncode == 0 or not isinstance(value, dict) or value.get("status") != "blocked":
    errors.append("same-head packet removal was not rejected by independent manifest/coverage validation")

# Non-object response input must return a blocked synthesis envelope, never a traceback.
shutil.rmtree(responses)
write_responses(responses)
first_packet = manifest["packets"][0]
(responses / f"{first_packet['packet_id']}.json").write_text("[]\n")
proc, value = synthesize(manifest_path, responses)
if not isinstance(value, dict) or value.get("status") != "blocked":
    errors.append("non-object response did not return the canonical blocked synthesis envelope")

# Claimed telemetry is null or non-negative and correctly typed; malformed values cannot be clean.
shutil.rmtree(responses)
write_responses(responses)
first_packet = manifest["packets"][0]
telemetry_path = responses / f"{first_packet['packet_id']}.json"
telemetry = json.loads(telemetry_path.read_text())
telemetry["context"]["input_tokens"] = -1
telemetry_path.write_text(json.dumps(telemetry))
proc, value = synthesize(manifest_path, responses)
if proc.returncode == 0 or not isinstance(value, dict) or value.get("status") != "blocked":
    errors.append("negative packet telemetry synthesized clean")

# Identifier-only hypotheses do not satisfy a falsifiable claim/alternative/discriminator contract.
shutil.rmtree(responses)
write_responses(responses)
first_response_path = responses / f"{first_packet['packet_id']}.json"
first_response = json.loads(first_response_path.read_text())
first_response["review_behaviors"]["hypotheses"] = ["H1"]
first_response_path.write_text(json.dumps(first_response))
proc, value = synthesize(manifest_path, responses)
if proc.returncode == 0 or not isinstance(value, dict) or value.get("status") != "blocked":
    errors.append("identifier-only hypothesis synthesized clean")

# A valid hashed lab artifact cannot be relabeled as another response experiment.
packet = first_packet
evidence_rel = os.path.relpath(tmp / "lab-evidence.json", repo)
experiment = {
    "hypothesis_id": "H1", "claim": "original claim", "alternative": "original alternative",
    "impact_edges_examined": packet.get("impact_edge_ids", [])[:1], "temporary_change": "original change",
    "command": ["python3", "scripts/check.py"], "expected_discriminator": {"exit_code": 1},
    "observed": {"exit_code": 1, "limit_hit": None, "process_residue_detected": False, "processes_remaining": 0, "sandbox_denial_observed": False},
}
evidence = {
    "schema_version": pm.LAB_EVIDENCE_SCHEMA, "status": "evidence", "packet_id": packet["packet_id"],
    "exact_base_sha": manifest["exact_base_sha"], "exact_head_sha": manifest["exact_head_sha"],
    "exact_head_tree": manifest["exact_head_tree"], "candidate_unchanged": True,
    "lab_cleanup_verified": True, "experiment": experiment,
    "final_state": {"resource_bounds_satisfied": True, "candidate_unchanged": True, "cleanup_verified": True},
}
evidence_path = repo / evidence_rel
evidence_path.write_text(json.dumps(evidence))
response = clean_response(packet)
response["experiments"] = [{**experiment, "claim": "relabeled claim", "supports": "claim",
    "candidate_unchanged": True, "lab_cleanup_verified": True,
    "lab_evidence_path": evidence_rel,
    "lab_evidence_sha256": hashlib.sha256(evidence_path.read_bytes()).hexdigest()}]
response["no_experiment_reason"] = None
if not pm.validate_experiments(repo, manifest, packet, response):
    errors.append("response experiment was not bound field-for-field to its hashed lab artifact")
response["experiments"][0]["claim"] = experiment["claim"]
evidence["experiment"]["observed"]["sandbox_denial_observed"] = True
evidence_path.write_text(json.dumps(evidence))
response["experiments"][0]["observed"]["sandbox_denial_observed"] = True
response["experiments"][0]["lab_evidence_sha256"] = hashlib.sha256(evidence_path.read_bytes()).hexdigest()
if not pm.validate_experiments(repo, manifest, packet, response):
    errors.append("sandbox denial was allowed to support clean experiment evidence")

# Packet accounting must charge real exact slices under the 30K/64 hard caps; a 100 KiB line is
# split without average-token assumptions, silent truncation, or an artificial empty-slice blocker.
config = json.loads((repo / ".agents/agentic-delivery/contracts/pm-review-system.json").read_text())
raw = b"x" * 100_000
blob_digest = hashlib.sha256(raw).hexdigest()
huge_slices=[]
for start in range(0,len(raw),4096):
    payload=raw[start:start+4096]
    huge_slices.append({
        "path":"scripts/huge.py","revision":manifest["exact_head_sha"],"revision_kind":"diff",
        "blob_sha256":blob_digest,"start_line":1,"end_line":1,"start_byte":start,
        "end_byte":start+len(payload),"bytes":len(payload),"sha256":hashlib.sha256(payload).hexdigest(),
    })
selection, packets, problems = pm.build_packets(
    manifest["exact_base_sha"], manifest["exact_head_sha"], manifest["exact_head_tree"],
    ["scripts/huge.py"], {"scripts/huge.py": 1}, {"scripts/huge.py": "implementation_test"},
    [], [], [], [], {"scripts/huge.py": 100_000}, {"scripts/huge.py": huge_slices}, config,
    changed_content_sizes={"scripts/huge.py":100_000}, changed_content_slices={"scripts/huge.py":huge_slices},
)
if problems or not packets or len(packets) > 64 or any(p["context"]["rendered_prompt_bytes"] > 30_000 for p in packets):
    errors.append("real 100 KiB one-line slices did not split within exact hard packet bounds")
if sum(item["bytes"] for packet in packets for item in packet["changed_file_slices"]) != len(raw):
    errors.append("real 100 KiB changed slices were not assigned exactly once")

for field,value in (("packet_target_tokens",30001),):
    bad=json.loads(json.dumps(config)); bad["thresholds"][field]=value
    try: pm.validate_config(bad)
    except pm.ReviewSystemError: pass
    else: errors.append("v4 accepted packet_target_tokens above the hard 30000 maximum")
bad=json.loads(json.dumps(config)); bad["impact_graph"]["max_packets"]=65
try: pm.validate_config(bad)
except pm.ReviewSystemError: pass
else: errors.append("v4 accepted max_packets above the hard 64 maximum")
bad=json.loads(json.dumps(config)); bad["impact_graph"]["max_nodes"]=1.5
try: pm.validate_config(bad)
except pm.ReviewSystemError: pass
else: errors.append("v4 accepted a fractional count bound")

# Every bounded chunk of a long provenance line must be assigned; choosing the first chunk would
# silently omit a reference located later on the same line.
def bound_slice(path,start,payload):
    return {"path":path,"revision":manifest["exact_head_sha"],"revision_kind":"head",
            "blob_sha256":hashlib.sha256(b"x"*512 if path=="scripts/source.py" else b"y").hexdigest(),
            "start_line":1,"end_line":1,"start_byte":start,"end_byte":start+len(payload),
            "bytes":len(payload),"sha256":hashlib.sha256(payload).hexdigest()}
source_slices=[bound_slice("scripts/source.py",0,b"x"*256),bound_slice("scripts/source.py",256,b"x"*256)]
target_slices=[bound_slice("scripts/target.py",0,b"y")]
edge={"id":"edge-long-line","source":"scripts/source.py","target":"scripts/target.py",
      "relation":"required_reference","direction":"forward","parser":"python_ast",
      "reason":"long_line_path","certainty":"active","line":1,"revision":"head",
      "source_blob_sha256":hashlib.sha256(b"x"*512).hexdigest(),"traversal_directions":["downstream"],"minimum_depth":1}
_, edge_packets, edge_problems = pm.build_packets(
    manifest["exact_base_sha"],manifest["exact_head_sha"],manifest["exact_head_tree"],
    [],{}, {}, [],[],["scripts/source.py","scripts/target.py"],[edge],
    {"scripts/source.py":512,"scripts/target.py":1},
    {"scripts/source.py":source_slices,"scripts/target.py":target_slices},config,
    edge_blob_slices={"scripts/source.py":source_slices,"scripts/target.py":target_slices},
)
assigned_source={item["start_byte"] for packet in edge_packets for item in packet.get("edge_context_slices",[]) if item["path"]=="scripts/source.py"}
if edge_problems or assigned_source != {0,256}:
    errors.append("long-line edge provenance did not include every bounded source-line chunk")

# Format-aware relation/certainty parsing must not depend on same-line filename keywords.
universe = {".agents/required.md", ".agents/next.md", "scripts/body.py", "scripts/helper.py"}
yaml_edges, _ = pm.typed_file_edges(
    ".agents/worker.yaml", "inputs:\n  required:\n    - .agents/required.md\n", universe
)
if not yaml_edges or yaml_edges[0]["relation"] != "required_reference":
    errors.append("YAML inputs.required lost structural required-reference semantics")
md_edges, _ = pm.typed_file_edges(
    ".agents/start.md", "Read .agents/required.md and .agents/next.md before work.\n", universe
)
if len(md_edges) != 2 or {edge["relation"] for edge in md_edges} != {"required_reference"}:
    errors.append("imperative multi-path read clause did not type both edges identically")
json_edges, _ = pm.typed_file_edges(
    ".agents/worker.json", '{"inputs":{"required":[".agents/required.md"]}}', universe
)
if not json_edges or json_edges[0]["relation"] != "required_reference":
    errors.append("JSON inputs.required lost structural required-reference semantics")
relative_md_edges, _ = pm.typed_file_edges(
    ".agents/sub/start.md", "## Required inputs\n- [required](../required.md)\n", universe
)
if not any(edge["target"] == ".agents/required.md" and edge["relation"] == "required_reference" for edge in relative_md_edges):
    errors.append("Markdown heading/list relative reference was not normalized and typed")
if pm.line_certainty("This unconditional contract reads .agents/required.md") != "active":
    errors.append("unconditional active reference was misclassified as unknown")
heredoc_edges, _ = pm.typed_file_edges(
    "scripts/check.sh", "python3 - <<'PY'\nfrom pathlib import Path\nPath('.agents/required.md').read_text()\nPY\n", universe
)
if not heredoc_edges or all(edge["certainty"] == "inactive" for edge in heredoc_edges):
    errors.append("executed Python heredoc dependency was classified wholly inactive")
python_edges, _ = pm.typed_file_edges("scripts/body.py", "from . import helper\n", universe)
if not any(edge["relation"] == "python_import" and edge["target"] == "scripts/helper.py" for edge in python_edges):
    errors.append("relative Python ImportFrom dependency was omitted")

# Both impact-edge endpoints and every seed are mandatory graph nodes.
invalid_graph = {"nodes": ["target"], "edges": [{"source": "missing", "target": "target", "certainty": "active"}],
                 "seeds": ["target"], "reported_impact": ["missing", "target"], "reported_status": "complete"}
if not pm.impact_contract_findings(invalid_graph):
    errors.append("impact detector accepted a missing source endpoint")
invalid_seed = {"nodes":["target"],"edges":[],"seeds":["absent"],"reported_impact":[],"reported_status":"complete"}
if not any("seed is absent" in item["claim"] for item in pm.impact_contract_findings(invalid_seed)):
    errors.append("impact detector accepted a seed outside the materialized node set")

# Dedicated phase-state validation must enforce nested correction constraints, not only markers.
phase_root = tmp / "phase-root"
phase_root.mkdir()
(phase_root / "schema.json").write_text(json.dumps({
    "required": ["schemaVersion", "correctionBudget"],
    "properties": {"schemaVersion": {"const": "pm-review-system-phase/v2"},
                   "correctionBudget": {"type": "object"}},
}))
(phase_root / "state.json").write_text(json.dumps({"schemaVersion": "pm-review-system-phase/v2", "correctionBudget": {}}))
_, _, phase_findings = pm.authority_inventory(phase_root, {"authorities": [], "configured_relationships": [{
    "source": "schema.json", "target": "state.json", "relation": "temporal_phase_state_instance",
    "certainty": "active", "validator": "pm_review_phase_v2",
}]})
if not phase_findings:
    errors.append("empty nested correction budget passed dedicated phase-state validation")
(source_mirror := phase_root / "source.json").write_text(json.dumps({"status":"clean","nested":{"head":"a"*40}}))
(target_mirror := phase_root / "target.json").write_text(json.dumps({"review":{"status":"blocked","head":"a"*40}}))
_, _, mirror_findings = pm.authority_inventory(phase_root, {"authorities": [], "configured_relationships": [{
    "source":"source.json", "target":"target.json", "relation":"authority_mirror", "certainty":"active",
    "validator":"authority_mirror_v1", "field_mappings":[{"source":"/status","target":"/review/status"},{"source":"/nested/head","target":"/review/head"}],
}]})
if not mirror_findings:
    errors.append("configured authority mirror disagreement was accepted")
bad_config = json.loads((repo / ".agents/agentic-delivery/contracts/pm-review-system.json").read_text())
bad_config["configured_relationships"][0]["certainty"] = "maybe"
try:
    pm.validate_config(bad_config)
except pm.ReviewSystemError:
    pass
else:
    errors.append("configured endpoint/certainty enum validation failed open")

# Scoring must reject a partial observed set instead of reporting perfect metrics.
observed_path = tmp / "partial-observed.json"
oracle_path = tmp / "oracle.json"
observed_path.write_text(json.dumps({"mode": "treatment", "observations": [{"case_id": "a", "suite": "x", "findings": [{}]}]}))
oracle_path.write_text(json.dumps({"cases": {"a": {"expected": "finding"}, "b": {"expected": "finding"}}}))
score = subprocess.run([sys.executable, str(repo / "scripts/pm-review-system.py"), "score",
                        "--observations", str(observed_path), "--oracle", str(oracle_path)],
                       cwd=repo, capture_output=True, text=True)
if score.returncode == 0:
    errors.append("partial observation set produced a measurement instead of blocking")
duplicate_observed = tmp / "duplicate-observed.json"
corpus = tmp / "duplicate-input.json"
corpus.write_text(json.dumps({"cases":[{"case_id":"a"},{"case_id":"a"}]}))
duplicate_observed.write_text(json.dumps({"mode":"treatment","input":str(corpus),"input_sha256":hashlib.sha256(corpus.read_bytes()).hexdigest(),"case_set_sha256":"0"*64,"observations":[{"case_id":"a","suite":"x","findings":[]},{"case_id":"a","suite":"x","findings":[]}]}))
duplicate = subprocess.run([sys.executable, str(repo / "scripts/pm-review-system.py"), "score", "--observations", str(duplicate_observed), "--oracle", str(oracle_path)], cwd=repo, capture_output=True, text=True)
if duplicate.returncode == 0:
    errors.append("duplicate measurement case ids bypassed exact corpus/oracle bijection")

# Dedup v1 preserves immutable observations and uses labels only after deterministic candidate
# generation. R1/R2 are retrospective development because R2 labels were inspected during tuning;
# prospective R3 is the next untouched measurement and no retrospective acceptance claim is made.
dedup_manifest_path = repo / "scripts/tests/fixtures/pm-review-system/dedup-corpus-manifest.json"
dedup_manifest = json.loads(dedup_manifest_path.read_text())
if dedup_manifest.get("next_untouched_measurement") != "fresh prospective exact-head R3":
    errors.append("dedup corpus policy does not reserve prospective R3")
historical_metrics = {}
for round_name, source in dedup_manifest.get("sources", {}).items():
    source_path = repo / source["path"]
    payload = source_path.read_bytes()
    if len(payload) != source["bytes"] or hashlib.sha256(payload).hexdigest() != source["sha256"]:
        errors.append(f"dedup retrospective source drifted: {round_name}")
        continue
    rows=[]
    for line in payload.decode().splitlines():
        if not line.startswith(f"| {round_name}-F"):
            continue
        cells=[item.strip().strip("`") for item in line.strip("|").split("|")]
        if len(cells) >= 8:
            rows.append(cells[:8])
    table_observations=[]; labels={}; paths={}
    for alias,packet_id,severity,category,location,root_group,disposition,response_digest in rows:
        path, _, line_value = location.partition(":")
        raw={"severity":severity,"category":category,"path":path,"line":line_value,
             "evidence":category.replace("_"," "),"impact":category.replace("_"," "),
             "smallest_safe_correction":category.replace("_"," ")}
        assignment={"invariants":["historical_table_limited"],"changed_file_slices":[],
                    "context_file_slices":[],"impact_file_slices":[],"edge_context_slices":[]}
        table_observations.append({"observation_id":alias,"features":pm.observation_features(assignment,raw)})
        labels[alias]=root_group; paths[alias]=path
    generated=[pm.generate_candidate_pairs(table_observations) for _ in range(10)]
    digests={pm.sha256_json(item) for item in generated}
    if len(digests) != 1 or any(item["model_tokens"] != 0 for item in generated):
        errors.append(f"dedup candidate generation is nondeterministic or used model tokens: {round_name}")
    pairs={tuple(item["observation_ids"]) for item in generated[0]["pairs"]}
    same=[]; cross=[]
    for index,left in enumerate(table_observations):
        for right in table_observations[index+1:]:
            left_id,right_id=left["observation_id"],right["observation_id"]
            if labels[left_id] == labels[right_id]:
                pair=tuple(sorted((left_id,right_id))); same.append(pair)
                if paths[left_id] != paths[right_id]: cross.append(pair)
    historical_metrics[round_name]={
        "candidate_recall":sum(pair in pairs for pair in same)/len(same) if same else 1.0,
        "cross_file_recall":sum(pair in pairs for pair in cross)/len(cross) if cross else 1.0,
        "candidate_pair_fraction":generated[0]["candidate_pair_fraction"],
        "claim":"retrospective development diagnostic only; labels inspected; no acceptance claim",
    }
if set(historical_metrics) != {"R1","R2"} or not all("no acceptance claim" in item["claim"] for item in historical_metrics.values()):
    errors.append("dedup retrospective metrics were mislabeled as held-out acceptance")

# Fixed transparent all-pairs generation remains bounded at 500 observations on this reference run.
perf_observations=[]
for index in range(500):
    raw={"category":f"unique_{index}","path":f"scripts/f{index}.py","line":"1",
         "evidence":f"mechanism_{index}","impact":f"impact_{index}",
         "smallest_safe_correction":f"replace_{index}"}
    assignment={"invariants":[f"inv_{index}"],"changed_file_slices":[],"context_file_slices":[],
                "impact_file_slices":[],"edge_context_slices":[]}
    perf_observations.append({"observation_id":f"PERF-{index:03d}","features":pm.observation_features(assignment,raw)})
latencies=[]
for _ in range(3):
    started=time.perf_counter(); pm.generate_candidate_pairs(perf_observations); latencies.append(time.perf_counter()-started)
if sorted(latencies)[-1] > 2.0:
    errors.append(f"dedup deterministic 500-observation p95 proxy exceeded 2s: {latencies}")

# One explicit causal decision may consolidate presentation, while a positive chain cannot bridge
# an ambiguous pair; recurrence/reassignment remain separate exact-head occurrences with audit IDs.
synthetic_packet={"packet_id":"implementation_test-99","invariants":["same_contract"],
                  "changed_file_slices":[],"context_file_slices":[],"impact_file_slices":[],"edge_context_slices":[]}
synthetic_findings=[]
for ordinal,(path,category,correction) in enumerate((
    ("scripts/a.py","same_mechanism","fix shared validator"),
    ("scripts/b.py","same_mechanism","fix shared validator"),
    ("scripts/a.py","distinct_trigger","fix separate parser"),
    ("scripts/moved.py","recurring_mechanism","fix prior validator"),
    ("scripts/reassigned.py","reassigned_mechanism","fix prior owner"),
),1):
    raw={"severity":"high","category":category,"path":path,"line":str(ordinal),
         "evidence":category,"impact":"candidate behavior fails","smallest_safe_correction":correction}
    synthetic_findings.append({"packet":synthetic_packet,"finding":raw,
                               "response_sha256":hashlib.sha256(f"response-{ordinal}".encode()).hexdigest(),
                               "raw_ordinal":ordinal})
synthetic_run, synthetic_observations = pm.build_observations(manifest, synthetic_findings)
ids=[item["observation_id"] for item in synthetic_observations]
def causal(ids_value, same=True):
    return {"same_invariant":same,"same_mechanism":same,"single_treatment":same,
            "contradictions":[] if same else ["independent trigger"],
            "counterfactual_all_cease":same,"evidence_observation_ids":sorted(ids_value)}
def event(sequence,previous,decision,ids_value,**extra):
    value={"sequence":sequence,"previous_event_sha256":previous,"decision":decision,
           "observation_ids":sorted(ids_value),"causal_test":causal(ids_value,decision in {"same","recurrence","reassign"}),**extra}
    value["event_sha256"]=pm.sha256_json(value); return value
events=[]; previous=None
for decision,ids_value,extra in (
    ("same",ids[:2],{}),("same",ids[1:3],{}),("ambiguous",[ids[0],ids[2]],{}),
    ("recurrence",[ids[3]],{"root_id":"RC-HISTORICAL-1"}),
    ("reassign",[ids[4]],{"root_id":"RC-HISTORICAL-2","retired_root_id":"RC-RETIRED-1"}),
):
    item=event(len(events)+1,previous,decision,ids_value,**extra);events.append(item);previous=item["event_sha256"]
decisions_path=tmp/"dedup-decisions.json"
decisions_path.write_text(json.dumps({"schema_version":pm.DEDUP_DECISIONS_SCHEMA,
                                      "run_id":synthetic_run["run_id"],"lineage_id":synthetic_run["lineage_id"],"events":events}))
history_path=tmp/"dedup-history.json"
history_path.write_text(json.dumps({"schema_version":pm.DEDUP_HISTORY_SCHEMA,
 "lineage_id":synthetic_run["lineage_id"],"roots":[
  {"root_id":"RC-HISTORICAL-1","friendly_aliases":["R1-X"],"occurrence_ids":["OCC-OLD-1"],"retired_root_ids":[]},
  {"root_id":"RC-HISTORICAL-2","friendly_aliases":["R1-Y"],"occurrence_ids":["OCC-OLD-2"],"retired_root_ids":[]}] }))
dedup=pm.synthesize_dedup(repo,manifest,synthetic_findings,
                           os.path.relpath(decisions_path,repo),os.path.relpath(history_path,repo))
if dedup["raw_finding_count"] != 5 or dedup["disclosure"]["retained_observation_count"] != 5:
    errors.append("dedup did not preserve complete flat observation disclosure")
if sorted(len(item["observation_ids"]) for item in dedup["occurrences"]) != [1,1,1,2]:
    errors.append("dedup allowed transitive candidate chaining or lost an explicit same pair")
if not any(item["recurrence_of"] == "RC-HISTORICAL-1" for item in dedup["occurrences"]):
    errors.append("later-head recurrence was treated as a same-run duplicate")
if not any("RC-RETIRED-1" in item["retired_root_ids"] for item in dedup["roots"]):
    errors.append("dedup reassignment did not retain the retired root id")
tampered=json.loads(decisions_path.read_text());tampered["events"][0]["causal_test"]["same_mechanism"]=False
tampered_path=tmp/"dedup-tampered.json";tampered_path.write_text(json.dumps(tampered))
try:
    pm.load_dedup_decisions(repo,os.path.relpath(tampered_path,repo),synthetic_run,synthetic_observations)
except pm.ReviewSystemError:
    pass
else:
    errors.append("dedup accepted a tampered append-only causal decision")

if errors:
    print("\n".join(errors), file=sys.stderr)
    raise SystemExit(1)
PY
then
  fail "round-2 systemic recurrence RED assertions did not fail closed"
fi

rm -rf "$unsafe_dir" "$compile_output"

if [[ $failures -ne 0 ]]; then
  printf 'PM review-system contract: %d semantic group(s) failed\n' "$failures" >&2
  exit 1
fi

printf 'pm review system ok: semantic gates, bidirectional impact, disposable labs, bounded exact-head packets, one PM synthesis, measured fixtures\n'
