# Worker Handoff Template

Workers use this template when returning control to a parent issue orchestrator.

````markdown
## Worker Handoff

Sub-issue:
Parent issue:
Worker agent:
Branch:
Sub-PR:
Parent PR:
Base branch:
Head SHA:

## Scope Delivered

- <change summary>

## Files Changed

- `<path>`: <reason>

## GSD / TDD Evidence

- GSD mode: <scripted | manual fallback>
- Red test evidence: <command/result or docs-only exemption>
- Green implementation evidence: <command/result>
- Refactor evidence: <command/result or not applicable>

## Verification

```bash
<command>
```

Result: <pass | fail | blocked>

## Automated Review

- Primary route: <coderabbit_auto | coderabbit_auto_incremental | coderabbit_manual_fallback | copilot_backup | human>
- Fallback route: <copilot_backup | human | none>
- Review route: <sub_pr | parent_pr_fallback | copilot_backup | blocked>
- Review status: <pending | clean | comments_addressed | skipped | blocked>
- Review URL:
- Disposition summary:
- Unresolved findings:

## Merge Recommendation

- Recommended state: <ready_for_merge | provisional_parent_integration | blocked>
- Reason:
- Human gates:
- Follow-up issues:
````

## Rules

- Do not include secrets or credential values.
- Do not claim CodeRabbit approval from a skipped-review status.
- Do not claim GitHub Copilot review as approval.
- Use `provisional_parent_integration` when parent PR fallback coverage is still pending.
- Name blockers explicitly instead of weakening verification.
