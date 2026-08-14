## Summary

Implements dynamic typed PostgreSQL source catalog discovery for the configured
database/schema. The runtime derives base relations, ordered columns,
nullability, ordered primary keys, and supported native/logical type metadata
from the live PostgreSQL catalog rather than a hard-coded connector schema.

## Linked Issue

Refs #3976
Refs #3972

## Stacked PR

- Parent issue: #4097
- Parent branch: `integration/4015-mvp-flat-r1`
- PR base branch: `integration/4015-mvp-flat-r1`
- Sub-issue: #3976
- Child PR: #4065
- Integration synchronization: safely merged base head
  `fbd06e7d7c5c0632182e98cbb3a223ba25b19883` in `0df3d5d4d`; no force push
  or discarded work

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
- Changed-path compliance: passed — only PostgreSQL source-catalog discovery,
  its focused tests, connector-description generated artifacts, and this GSD
  evidence changed. The inherited PostgreSQL surface audit found only
  test-only fixture schemas and test setup DDL; no live hard-coded table or
  column shape remains in #3976's discovery path.
- Foundation issue/PR path: #3974 / #4034 typed database foundation
- Shared runtime/tooling or unrelated connector changes: none planned
- no-mistakes foundation split status: not applicable; stop if a shared-foundation path is required

## Verification

RED is committed in `db7e06d36`; catalog GREEN is `24d0055f5`; live-read
RED/GREEN is recorded in `traces/live-reads-{red,green}.txt`. Focused PostgreSQL,
database-foundation, engine, and CLI tests, the PostgreSQL race suite, `go vet
./...`, build, lint, docs, smoke, contract, connector-generation/boundary, and
release workflow gates are green. The opt-in real Docker proof through Colima
discovered the seeded catalog, returned full IDs `1,2,3,4,5`, and returned only
`3,4,5` after cursor `10`; its exact command and output are posted on #3976.
The historical CDC integration test remains an intentional fail-closed skip,
not claimed coverage. No-mistakes is pending the final evidence commit.

## Automated Review

- Primary route: pending until the child is non-draft and locally green
- Fallback route: Copilot only if Claude is unavailable and coverage blocks progress
- PR base/default branch: `integration/4015-mvp-flat-r1` / `main`
- Latest reviewed commit: pending
- Reviewed range: pending
- Coverage route: pending
- Coverage status: pending
- Disposition summary: pending
- Follow-up review status: pending
