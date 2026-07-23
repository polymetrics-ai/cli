#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

python3 - "$repo_root" <<'PY'
from __future__ import annotations

import json
import pathlib
import re
import sys

root = pathlib.Path(sys.argv[1])
errors: list[str] = []


def read(relative: str) -> str:
    path = root / relative
    if not path.is_file():
        errors.append(f"missing canonical file: {relative}")
        return ""
    return path.read_text()


def require(relative: str, *needles: str) -> None:
    text = read(relative)
    for needle in needles:
        if needle not in text:
            errors.append(f"{relative}: missing required route marker {needle!r}")


def forbid(relative: str, patterns: dict[str, str]) -> None:
    text = read(relative)
    for label, pattern in patterns.items():
        if re.search(pattern, text, flags=re.IGNORECASE):
            errors.append(f"{relative}: forward PM route still contains {label}")


local_review = ".agents/agentic-delivery/workflows/local-codex-review-loop.md"
local_prompt = ".agents/agentic-delivery/prompts/local-codex-review-prompt.md"
require(
    local_review,
    "fresh-context",
    "exact base",
    "exact head",
    "disposition",
    "re-review",
    "Shepherd",
)
require(local_prompt, "exact base", "exact head", "read-only", "disposition")

forward_paths = [
    ".agents/agentic-delivery/contracts/parent-orchestrator-contract.md",
    ".agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md",
    ".agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md",
    ".agents/agentic-delivery/workflows/pi-active-orchestration-loop.md",
    ".agents/agentic-delivery/workflows/pi-autonomous-orchestration-loop.md",
    ".agents/agentic-delivery/workflows/codex-active-orchestration-loop.md",
    ".agents/agentic-delivery/references/gsd-pi-adapter.md",
    ".pi/prompts/pm-orchestrate.md",
    ".pi/prompts/pm-auto-loop.md",
    ".pi/prompts/pm-gsd-loop.md",
    ".pi/prompts/pm-review-loop.md",
    ".agents/agentic-delivery/agents/coordination/parent-issue-orchestrator.agent.yaml",
]
legacy_review_patterns = {
    "Claude review requirement": r"claude-review-loop|@claude\s+review|claude_(?:auto|manual|required)",
    "Copilot review route": r"copilot(?:_backup|\s+(?:backup\s+)?review)",
    "legacy PM disposition agent": r"pm-claude-review-disposition",
}
unavailable_command_patterns = {
    "unavailable programming-loop command": (
        r"/gsd-programming-loop|scripts/gsd\s+prompt\s+programming-loop|"
        r"\bgsd-programming-loop\b"
    ),
}
for relative in forward_paths:
    require(relative, "local-codex-review-loop.md", "shepherd-validator.md")
    forbid(relative, legacy_review_patterns)
    forbid(relative, unavailable_command_patterns)


def markdown_dependencies(seeds: list[str]) -> set[str]:
    pending = list(seeds)
    seen: set[str] = set()
    reference = re.compile(r"(?:\.agents|\.pi)/[A-Za-z0-9_./-]+\.md")
    while pending:
        relative = pending.pop()
        if relative in seen:
            continue
        seen.add(relative)
        for dependency in reference.findall(read(relative)):
            if dependency not in seen and (root / dependency).is_file():
                pending.append(dependency)
    return seen


pm_entry_paths = forward_paths + [
    ".pi/prompts/pm-review-loop.md",
    ".pi/agents/pm-gsd-worker.md",
    ".pi/agents/pm-issue-worker.md",
    ".agents/agentic-delivery/agents/coordination/parent-issue-orchestrator.agent.yaml",
]
pm_dependencies = markdown_dependencies(pm_entry_paths)
for deprecated_template in (
    ".agents/agentic-delivery/contracts/worker-handoff-template.md",
    ".agents/agentic-delivery/contracts/code-review-disposition-template.md",
):
    for relative in sorted(pm_dependencies):
        if deprecated_template in read(relative):
            errors.append(
                f"{relative}: current PM dependency graph reaches bot-era template {deprecated_template}"
            )
for canonical_template in (
    ".agents/agentic-delivery/contracts/pm-worker-handoff-template.md",
    ".agents/agentic-delivery/contracts/pm-code-review-disposition-template.md",
):
    if canonical_template not in pm_dependencies:
        errors.append(f"PM dependency graph does not reach canonical template: {canonical_template}")
    require(
        canonical_template,
        "exact_base_sha",
        "exact_head_sha",
        "local_codex",
        "shepherd",
        "human",
    )
    forbid(canonical_template, legacy_review_patterns)

require(
    ".agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md",
    "Canonical PM fallback",
    "/pm-orchestrate",
    "local-codex-review-loop.md",
    "shepherd-validator.md",
)
require(
    ".agents/agentic-delivery/references/gsd-pi-adapter.md",
    "programming-loop",
    "/pm-orchestrate",
    "local-codex-review-loop.md",
    "shepherd-validator.md",
)
require(
    ".agents/agentic-delivery/schemas/orchestration-state.schema.yaml",
    "canonical_v2",
    "local_codex",
    "shepherd",
    "verdict_required_when",
    "verdict_must_be_absent_when",
    "legacy_review_route",
    "legacy_local_codex_v1",
    "correction_budget",
    "max_correction_rounds",
    "rounds_by_range",
)
require(
    ".agents/agentic-delivery/contracts/parent-orchestrator-contract.md",
    "max_correction_rounds",
    "rounds_by_range",
)
require(
    ".agents/agentic-delivery/workflows/local-codex-review-loop.md",
    "max_correction_rounds",
    "rounds_by_range",
)
require(
    ".pi/prompts/pm-orchestrate.md",
    "max_correction_rounds: 4",
    "rounds_by_range",
)
require(
    ".planning/phases/397-pm-orchestrator-extension/PLAN.md",
    "Subsequent PR #493 migration gate",
    "not_spawned_dependency_blocked",
)
require(
    ".planning/traces/cli-architecture-v2-orchestration-state.yaml",
    "post_wave1_routing_migration",
    "pull_request: 493",
    "blocks_worker_start: true",
    "not_spawned_dependency_blocked",
)
require(
    ".planning/phases/397-cli-architecture-v2-orchestration/RUN-STATE.json",
    '"postWave1RoutingMigrationGate"',
    '"pullRequest": 493',
    '"blocksWorkerStart": true',
    "not_spawned_dependency_blocked",
)
require(
    ".planning/phases/397-cli-architecture-v2-orchestration/SUMMARY.md",
    "PR #493 canonical PM migration gate",
    "before another CLI Architecture v2 implementation worker starts",
    "not_spawned_dependency_blocked",
)
trace = read(".planning/traces/cli-architecture-v2-orchestration-state.yaml")
subissues = trace.split("\nsubissues:\n", 1)
if len(subissues) != 2:
    errors.append("authoritative trace: missing subissues section")
else:
    issue_408 = re.search(r"(?ms)^  - number: 408\n(?P<body>.*?)(?=^  - number:|\Z)", subissues[1])
    if issue_408 is None:
        errors.append("authoritative trace: missing scoped #408 subissue")
    else:
        body = issue_408.group("body")
        for marker in ("parent_sync_and_pr_493_migration_pending", "Wave 1", "PR #493"):
            if marker not in body:
                errors.append(f"authoritative trace #408 subissue: missing {marker!r}")

run_state = json.loads(read(".planning/phases/397-cli-architecture-v2-orchestration/RUN-STATE.json"))
main_sync = run_state.get("mainSync", {})
if main_sync.get("taskBranch") != "chore/cli-architecture-v2-wave1-parent-sync-r1":
    errors.append("#397 run state: current mainSync.taskBranch is not PR #495's chore/... branch")
if main_sync.get("historicalTaskBranch") != "fm/cli-architecture-v2-wave1-parent-sync-r1":
    errors.append("#397 run state: superseded fm/... branch is not preserved as historicalTaskBranch")
require(
    ".agents/agentic-delivery/workflows/pi-autonomous-orchestration-loop.md",
    '"correction_budget"',
    '"rounds_by_range"',
    "candidate lineage",
    "read-only legacy input",
)
require(
    ".pi/prompts/pm-auto-loop.md",
    "correction_budget.max_correction_rounds",
    "correction_budget.rounds_by_range",
    "candidate lineage",
    "read-only legacy input",
)
for relative in (
    ".agents/agentic-delivery/workflows/pi-autonomous-orchestration-loop.md",
    ".pi/prompts/pm-auto-loop.md",
):
    forbid(relative, {"active per-subissue correction counter": r'"correction_rounds"\s*:'})

local_statuses = ("pending", "findings_correction_required", "clean", "comments_addressed", "blocked")
canonical_contract_paths = (
    ".agents/agentic-delivery/schemas/orchestration-state.schema.yaml",
    ".agents/agentic-delivery/workflows/local-codex-review-loop.md",
    ".agents/agentic-delivery/prompts/local-codex-review-prompt.md",
    ".agents/agentic-delivery/contracts/parent-orchestrator-contract.md",
    ".agents/agentic-delivery/contracts/pm-worker-handoff-template.md",
    ".agents/agentic-delivery/contracts/pm-code-review-disposition-template.md",
)
for relative in canonical_contract_paths:
    require(relative, *local_statuses)

expected_dispositions = [
    "accepted",
    "accepted_with_modification",
    "declined",
    "duplicate",
    "deferred",
    "needs_human",
]
for relative in canonical_contract_paths:
    match = re.search(r"finding_disposition_values:\s*\[([^\]]+)\]", read(relative))
    if match is None:
        errors.append(f"{relative}: missing exact finding_disposition_values enum")
        continue
    actual = [value.strip().strip("`") for value in match.group(1).split(",")]
    if actual != expected_dispositions:
        errors.append(f"{relative}: disposition enum drift: {actual!r}")

for relative in (
    ".agents/agentic-delivery/workflows/pi-autonomous-orchestration-loop.md",
    ".pi/prompts/pm-auto-loop.md",
):
    require(relative, "schema_version", "human_gate", "human_gate_kind", "correction_cap_exceeded")

require(
    "scripts/pm-terminal-classifier.sh",
    "correction_cap_exceeded",
    "blocked_human_decision",
    "human_ready",
)
for relative in ("scripts/pi-auto-loop.sh", "scripts/pi-shepherd-loop.sh"):
    require(relative, "pm-terminal-classifier.sh", "blocked human decision")

reviewer = read(".pi/agents/pm-reviewer.md")
if not re.search(r"^tools:.*\bbash\b", reviewer, flags=re.MULTILINE):
    errors.append(".pi/agents/pm-reviewer.md: exact-head reviewer lacks read-only bash/git access")
for marker in ("fresh-context", "exact base", "exact head", "Do not modify"):
    if marker not in reviewer:
        errors.append(f".pi/agents/pm-reviewer.md: missing {marker!r}")

legacy_agent = read(".pi/agents/pm-claude-review-disposition.md")
for marker in ("Deprecated", "local-codex-review-loop.md", "pm-reviewer"):
    if marker not in legacy_agent:
        errors.append(f".pi/agents/pm-claude-review-disposition.md: missing migration marker {marker!r}")

for relative in (
    ".agents/agentic-delivery/workflows/automated-review-routing-loop.md",
    ".agents/agentic-delivery/workflows/claude-review-loop.md",
):
    require(relative, "Legacy", "local-codex-review-loop.md", "not part of the canonical PM")

model_test = read("scripts/tests/pi-model-routing.sh")
if "scripts/tests/pm-orchestrator-contract.sh" not in model_test:
    errors.append("scripts/tests/pi-model-routing.sh: focused PM contract is not in make verify path")

fixture_root = root / "scripts" / "tests" / "fixtures" / "pm-orchestrator-review-state"
fixtures = {
    "pending": fixture_root / "pending.json",
    "clean": fixture_root / "clean.json",
    "blocked": fixture_root / "blocked.json",
    "cap_lineage": fixture_root / "correction-cap-lineage.json",
    "canonical_missing_kind": fixture_root / "canonical-missing-human-gate-kind.json",
    "parent_ready": fixture_root / "parent-ready.json",
    "historical": fixture_root / "historical-local-codex.json",
}
for name, path in fixtures.items():
    if not path.is_file():
        errors.append(f"missing review-state fixture: {path.relative_to(root)}")
        continue
    record = json.loads(path.read_text())
    review = record.get("automated_review", {})
    if record.get("schema_version") == "canonical_v2":
        budget = record.get("correction_budget", {})
        if budget.get("max_correction_rounds", 0) < 1 or not isinstance(budget.get("rounds_by_range"), dict):
            errors.append(f"{name}: invalid canonical correction budget")
        shepherd = review.get("shepherd", {})
        status = shepherd.get("status")
        verdict = shepherd.get("verdict")
        if status in {"pending", "blocked"} and verdict is not None:
            errors.append(f"{name}: {status} Shepherd record invents a verdict")
        if status in {"proceed", "retry", "revert", "halt"} and not verdict:
            errors.append(f"{name}: completed Shepherd record lacks verdict")
    elif review.get("status") != "complete_no_unresolved_findings" or "head_sha" not in review:
        errors.append(f"{name}: historical local Codex fixture is not recognized read-only evidence")

cap_fixture = fixtures["cap_lineage"]
if cap_fixture.is_file():
    record = json.loads(cap_fixture.read_text())
    budget = record.get("correction_budget", {})
    rounds = budget.get("rounds_by_range", {})
    lineage = record.get("candidate_lineage", {})
    if len(lineage.get("replacement_heads", [])) < 2:
        errors.append("cap_lineage: replacement head history is missing")
    if not rounds or max(rounds.values()) <= budget.get("max_correction_rounds", 0):
        errors.append("cap_lineage: correction cap is not exceeded on the stable lineage")
    if record.get("terminal") != "human_gate" or record.get("automated_review", {}).get("status") != "blocked":
        errors.append("cap_lineage: cap exceed does not persist a blocked human gate")
    if record.get("human_gate_kind") != "correction_cap_exceeded":
        errors.append("cap_lineage: human gate kind does not identify correction-cap exceed")

if errors:
    raise SystemExit("PM orchestrator contract violations:\n- " + "\n- ".join(errors))

print(
    "pm orchestrator contract ok: one canonical owner; PR #493 routing migration gate; "
    "unavailable-command fallback; exact-head local Codex review; independent Shepherd; "
    "no Claude/Copilot PM coverage"
)
PY

classification="$(bash "$repo_root/scripts/pm-terminal-classifier.sh" \
    "$repo_root/scripts/tests/fixtures/pm-orchestrator-review-state/correction-cap-lineage.json")"
if [[ "$classification" != "blocked_human_decision" ]]; then
  printf 'PM orchestrator contract violation: cap classifier returned %s\n' "$classification" >&2
  exit 1
fi
classification="$(bash "$repo_root/scripts/pm-terminal-classifier.sh" \
    "$repo_root/scripts/tests/fixtures/pm-orchestrator-review-state/canonical-missing-human-gate-kind.json")"
if [[ "$classification" != "blocked_human_decision" ]]; then
  printf 'PM orchestrator contract violation: canonical missing-kind classifier returned %s\n' "$classification" >&2
  exit 1
fi
classification="$(bash "$repo_root/scripts/pm-terminal-classifier.sh" \
    "$repo_root/scripts/tests/fixtures/pm-orchestrator-review-state/parent-ready.json")"
if [[ "$classification" != "human_ready" ]]; then
  printf 'PM orchestrator contract violation: parent-ready classifier returned %s\n' "$classification" >&2
  exit 1
fi
