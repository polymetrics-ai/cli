# Canonical PM Local Review Disposition Template

Use this template for every fresh-context `local_codex` finding in the canonical PM route. Keep the
record in the issue phase artifact and summarize it truthfully in the PR body when applicable.

```markdown
## Finding <ID>

- severity: <critical | high | medium | low>
- finding_disposition_values: [accepted, accepted_with_modification, declined, duplicate, deferred, needs_human]
- disposition: <one finding_disposition_values value>
- exact_base_sha:
- exact_head_sha:
- candidate_lineage:
- local_codex.packet_id:
- local_codex.reviewer_identity:
- local_codex.raw_response_path_or_hash:
- local_codex.synthesis_artifact:
- local_codex.finding_artifact:

### Action

<What changed, what will change, or why no code change is being made.>

### Reason

<Why this disposition follows issue scope, project rules, source evidence, and safety gates.>

### Evidence

<Tests, source URLs, issue scope, commits, or file references.>

### Follow-up

<Focused follow-up issue, human approval gate, or "None".>

### Gate state

- local_codex.status: <pending | findings_correction_required | clean | comments_addressed | blocked>
- local_codex.disposition_artifact:
- shepherd.status: <pending | proceed | retry | revert | halt | blocked>
- shepherd.exact_head_sha:
- shepherd.verdict: <PROCEED | RETRY | REVERT | HALT | absent while pending/blocked>
- human.gate: <none | required with reason>
```

## Rules

- Validate every disposition table data row independent of finding-ID prefix. Identifiers such as
  `SEC1`, `UX-22`, `F1`, `N4-R`, and future nonempty ids use the same exact disposition enum.
- Verify the finding against the exact range and compiled packet before accepting or declining it.
- Do not silently dismiss a finding; record evidence and the smallest safe action.
- Create or reference a focused follow-up issue for valid out-of-scope work.
- Accepted changes create a new `exact_head_sha`; rerun affected verification, deterministic
  compilation, fresh packet review, and one synthesis before independent Shepherd validation.
- Do not invent a Shepherd verdict while it is pending or blocked.
- If the correction budget is exceeded or the disposition is `needs_human`, persist the blocker and
  stop for a human decision.
- Human merge and readiness authority is never delegated by this template.
