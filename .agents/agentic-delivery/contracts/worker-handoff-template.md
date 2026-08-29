# Durable Wave Checkpoint Template

The canonical worker uses this compatibility template to persist one wave checkpoint in an issue,
PR, or GSD artifact. It is durable resume state for the same worker or a successor, not a handoff to
a spawned parent-orchestrator role.

````markdown
## Wave Checkpoint

Sub-issue:
Parent issue:
Canonical worker:
Branch:
Sub-PR:
Parent PR:
Base branch:
Worktree:
Head SHA:

## Scope Delivered

- <change summary>

## Files Changed

- `<path>`: <reason>

## Connector Implementation Scope

- Applies: <yes | no>
- Named connector cohort: <one or more slugs, or not applicable>
- Immutable source-lock ledger / ownership-path matrix: <path, or not applicable>
- Changed-path compliance: <pass | fail | blocked; list any out-of-scope path>
- Foundation Atlas disposition: <reuse | constrained extension | actual gap with captain approval | not applicable>
- Shared runtime/tooling scope: <named foundation paths and affected connectors, or none>
- Unrelated connector changes: <none | out-of-scope blocker>
- Cohort guard status: <not run | pass | blocker>

## GSD / TDD / Skill Evidence

- GSD mode: <pi / scripts/gsd / manual fallback>
- GSD command: </gsd ... or scripts/gsd prompt ...>
- GSD adapter source: `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- Required skills source: `.agents/agentic-delivery/references/required-skills-routing.md`
- Required Go skills loaded: <golang-how-to, golang-cli, golang-testing, ... or not applicable>
- Required design skills loaded: <frontend-design, web-design-guidelines, vercel-react-best-practices, ... or not applicable>
- Red test evidence: <command/result or docs-only exemption>
- Green implementation evidence: <command/result>
- Refactor evidence: <command/result or not applicable>

## CLI Help / Docs / Website Parity

- Applies: <yes | no>
- Runtime help checked: <pm help topic / pm namespace / pm command --help / not applicable>
- Bare namespace behavior checked: <command/result or not applicable>
- `docs/cli/**` updated: <yes | no | not applicable>
- `website/**` updated: <yes | no | not applicable>
- Generated help/manual artifacts updated: <yes | no | not applicable>
- Parity exemptions:

## Verification

```bash
<command>
```

Result: <pass | fail | blocked>

## Automated Review

- Primary route: <claude_auto | claude_auto_incremental | claude_manual_fallback | copilot_backup | human>
- Fallback route: <copilot_backup | human | none>
- Coverage route: <sub_pr | parent_pr_fallback | copilot_backup | blocked>
- Coverage status: <pending | clean | comments_addressed | skipped | blocked>
- Review URL:
- Disposition summary:
- Unresolved findings:

## Integration Recommendation

- Recommendation: <integrate | hold | block>
- Reason:
- Human gates:
- Follow-up issues:
````

## Rules

- Do not include secrets or credential values.
- Compact handoff prose is allowed, but exact commands, code, test output, security warnings,
  destructive-action warnings, ordered safety gates, and approval gates must remain exact and
  unambiguous.
- Do not claim Claude approval from a skipped-review status.
- Do not claim GitHub Copilot review as approval.
- Use `hold` for a technical parent-branch landing awaiting fallback coverage; do not record the
  canonical `integrate_sub_pr` state or unblock dependent work until coverage is complete.
- Name blockers explicitly instead of weakening verification.
