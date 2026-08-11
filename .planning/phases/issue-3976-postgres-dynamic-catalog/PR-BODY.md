## Summary

Implements dynamic typed PostgreSQL source catalog discovery for the configured
database/schema. The runtime derives base relations, ordered columns,
nullability, ordered primary keys, and supported native/logical type metadata
from the live PostgreSQL catalog rather than a hard-coded connector schema.

## Linked Issue

Refs #3976
Refs #3972

## Stacked PR

- Parent issue: #3972
- Parent branch: `feat/3972-postgres-parity`
- PR base branch: `feat/3972-postgres-parity`
- Sub-issue: #3976

## Parent Orchestration

- Orchestrator: canonical inline delivery worker
- State ledger: `.planning/phases/issue-3976-postgres-dynamic-catalog/RUN-STATE.json`
- Worker handoff: this phase directory
- Merge owner: human/captain for parent PR
- Integration state: held until corrected #4058 is green and merged

## Connector Implementation Scope

- Applies: yes
- Target connector scope: native PostgreSQL source catalog adapter only
- Connector-owned paths: `internal/connectors/native/postgres/**` and focused live-test evidence
- Ownership guard evidence: #3976 owns source catalog/type/fingerprint discovery; #3980/#3982/#3983/#3987 remain untouched
- Changed-path compliance: pending implementation
- Foundation issue/PR path: #3974 / #4034 typed database foundation
- Shared runtime/tooling or unrelated connector changes: none planned
- no-mistakes foundation split status: not applicable; stop if a shared-foundation path is required

## Verification

Pending implementation. The completed record will include RED/GREEN/REFACTOR
evidence, focused and live PostgreSQL oracle tests, race/vet/build/repository
gates, and no-mistakes result.

## Automated Review

- Primary route: pending until the child is non-draft and locally green
- Fallback route: Copilot only if Claude is unavailable and coverage blocks progress
- PR base/default branch: `feat/3972-postgres-parity` / `main`
- Latest reviewed commit: pending
- Reviewed range: pending
- Coverage route: pending
- Coverage status: pending
- Disposition summary: pending
- Follow-up review status: pending
