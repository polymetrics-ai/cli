# Canonical PM Worker Handoff Template

Current and forward PM-orchestrated workers use this template when returning control to the parent
orchestrator. Historical records keep their original shape.

````markdown
## PM Worker Handoff

- schema_version: `canonical_v2`
- sub_issue:
- parent_issue:
- worker_agent:
- branch:
- sub_pr:
- parent_pr:
- base_branch:
- worker_directory:
- exact_base_sha:
- exact_head_sha:
- candidate_lineage:

## Scope Delivered

- <change summary>

## Files Changed

- `<path>`: <reason>

## GSD / TDD / Skill Evidence

- GSD registry discovery: <commands and result>
- Lifecycle owner: <available registered GSD command or canonical PM fallback>
- Required skills loaded: <skills or not applicable>
- RED evidence: <command/result or docs-only exemption>
- GREEN evidence: <command/result>
- REFACTOR evidence: <command/result or not applicable>

## CLI Help / Docs / Website Parity

- Applies: <yes | no>
- Runtime help checked: <commands/result or not applicable>
- Bare namespace behavior checked: <command/result or not applicable>
- `docs/cli/**`: <updated | not applicable with reason>
- `website/**`: <updated | not applicable with reason>
- Generated help/manual artifacts: <updated | not applicable with reason>

## Verification

```bash
<exact commands>
```

Result: <pass | fail | blocked>

## Canonical Review Gates

- local_codex.status: <pending | findings_correction_required | clean | blocked>
- local_codex.exact_base_sha:
- local_codex.exact_head_sha:
- local_codex.findings_artifact:
- local_codex.disposition_artifact:
- shepherd.status: <pending | proceed | retry | revert | halt | blocked>
- shepherd.exact_head_sha:
- shepherd.verdict: <PROCEED | RETRY | REVERT | HALT | absent while pending/blocked>
- shepherd.evidence_artifact:
- correction_budget.max_correction_rounds:
- correction_budget.candidate_lineage:
- correction_budget.rounds_by_range:

## Integration Recommendation

- recommended_state: <provisional_parent_integration | blocked | human_gate>
- reason:
- human_gates:
- follow_up_issues:
````

## Rules

- Do not include secrets or credential values.
- Bind verification and both review gates to `exact_head_sha`; any head change invalidates them.
- A pending or blocked Shepherd record has no invented verdict.
- Only the parent orchestrator may integrate a sub-PR into the parent branch.
- Integration to the default branch, parent readiness, and final release remain human-only.
- Name blockers explicitly instead of weakening verification or review requirements.
