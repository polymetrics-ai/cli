# Requirements

This file is the explicit capability and coverage contract for the project.

## Active

### R003 — Representative Asana list commands execute through the bounded generic stream runner with pagination and deterministic JSON output, without connector-specific duplicate transport.
- Class: primary-user-loop
- Status: active
- Description: Representative Asana list commands execute through the bounded generic stream runner with pagination and deterministic JSON output, without connector-specific duplicate transport.
- Why it matters: Users need real CLI reads, while maintainers need one bounded, reusable execution path.
- Source: spec
- Primary owning slice: M001/S03
- Supporting slices: M001/S01, M001/S02
- Validation: mapped
- Notes: Issue #383; fixture and httptest proof only.

### R004 — Every one of the 250 baseline Asana REST operations retains committed source evidence and has exactly one safe primary classification or an explicit exclusion.
- Class: compliance/security
- Status: active
- Description: Every one of the 250 baseline Asana REST operations retains committed source evidence and has exactly one safe primary classification or an explicit exclusion.
- Why it matters: Exhaustive, mutually exclusive accounting prevents silent coverage gaps and accidental executable permission.
- Source: spec
- Primary owning slice: M001/S04
- Supporting slices: M001/S05, M001/S06, M001/S07
- Validation: mapped
- Notes: Issue #384; classification does not itself grant runtime execution.

### R005 — Only fixed allow-listed Asana direct-read operations execute, with bounded JSON output, declared redaction, cancellation, and no caller-controlled method or arbitrary URL.
- Class: core-capability
- Status: active
- Description: Only fixed allow-listed Asana direct-read operations execute, with bounded JSON output, declared redaction, cancellation, and no caller-controlled method or arbitrary URL.
- Why it matters: Users need targeted reads that do not fit streams without receiving a generic API escape hatch.
- Source: spec
- Primary owning slice: M001/S05
- Supporting slices: M001/S04, M001/S02
- Validation: mapped
- Notes: Issue #385; concrete allow-list is derived from the classified committed ledger.

### R006 — Fixed Asana attachment download operations enforce byte limits, safe-root confinement, traversal and symlink defenses, context cancellation, and an explicit no-clobber or approved-overwrite policy.
- Class: compliance/security
- Status: active
- Description: Fixed Asana attachment download operations enforce byte limits, safe-root confinement, traversal and symlink defenses, context cancellation, and an explicit no-clobber or approved-overwrite policy.
- Why it matters: Binary downloads cross a filesystem trust boundary and must not permit unbounded data or path escape.
- Source: spec
- Primary owning slice: M001/S06
- Supporting slices: M001/S04, M001/S02
- Validation: mapped
- Notes: Issue #386; fixture and httptest proof only, with no live attachment access.

## Validated

### R001 — Asana exposes a validated declarative CLI command vocabulary that references all 12 existing streams and existing typed write actions without unresolved or raw-API mappings.
- Class: core-capability
- Status: validated
- Description: Asana exposes a validated declarative CLI command vocabulary that references all 12 existing streams and existing typed write actions without unresolved or raw-API mappings.
- Why it matters: Every downstream help and runtime command needs one stable, machine-validated source of truth.
- Source: spec
- Primary owning slice: M001/S01
- Supporting slices: M001/S02, M001/S03, M001/S07
- Validation: S01 evidence: gsd_exec 18679e3c-c484-4051-942c-1ff429fab5b9 confirms 250 unique API identities, 25 unique implemented CLI leaves, exactly 12 unique stream references and 13 unique action references, with no raw transport fields; gsd_exec 5d2e058e-750d-4842-a8b5-54b5c040d37f confirms credential-free production bundle validation with 547 connectors and zero findings; gsd_exec 2fc99a43-c1b6-4064-a517-68c94d0ca712 passes duplicate-path, missing-target, deterministic-ordering, and raw-API negative tests.
- Notes: Validated by M001 S01 contract proof. Runtime help/docs and command execution are intentionally owned by downstream slices.

### R002 — Users can discover Asana namespace, topic, and leaf commands without credentials through text help, JSON help, generated manual artifacts, docs/cli, connector docs, and website documentation that agree.
- Class: launchability
- Status: validated
- Description: Users can discover Asana namespace, topic, and leaf commands without credentials through text help, JSON help, generated manual artifacts, docs/cli, connector docs, and website documentation that agree.
- Why it matters: A CLI surface is not usable or supportable when its runtime help and published documentation drift.
- Source: spec
- Primary owning slice: M001/S02
- Supporting slices: M001/S01, M001/S03, M001/S05, M001/S06, M001/S07
- Validation: S02 evidence: gsd_exec f77cb004-d9c7-4bd3-a002-78d8fcee244a proves built-CLI credential-free namespace, topic, and leaf text aliases, JSON manual equality, invalid-prefix rejection, and no .polymetrics state; gsd_exec 9fe9a1c0-e528-4ec5-a994-8416920f4867 proves exactly four targeted docs are byte-identical and repeat-generation idempotent in external scratch; gsd_exec 2f09b230-ac3f-48df-9c43-8cd57dea9a5a proves focused CLI/connector, website unit/type, catalog, docs validation, and whitespace gates; gsd_exec d4988b68-c49f-4dd4-836b-2060633df340 passes the exact repository-wide make verify aggregate, including format, tidy, vet, full Go tests, build, docs validation, smoke, lint, and connector validation.
- Notes: Validated without live Asana credentials, reverse-ETL execution, or project-state creation. Full-corpus documentation generation was not run; targeted generation stayed under external TMPDIR per K001.

## Deferred

## Out of Scope

## Traceability

| ID | Class | Status | Primary owner | Supporting | Proof |
|---|---|---|---|---|---|
| R001 | core-capability | validated | M001/S01 | M001/S02, M001/S03, M001/S07 | S01 evidence: gsd_exec 18679e3c-c484-4051-942c-1ff429fab5b9 confirms 250 unique API identities, 25 unique implemented CLI leaves, exactly 12 unique stream references and 13 unique action references, with no raw transport fields; gsd_exec 5d2e058e-750d-4842-a8b5-54b5c040d37f confirms credential-free production bundle validation with 547 connectors and zero findings; gsd_exec 2fc99a43-c1b6-4064-a517-68c94d0ca712 passes duplicate-path, missing-target, deterministic-ordering, and raw-API negative tests. |
| R002 | launchability | validated | M001/S02 | M001/S01, M001/S03, M001/S05, M001/S06, M001/S07 | S02 evidence: gsd_exec f77cb004-d9c7-4bd3-a002-78d8fcee244a proves built-CLI credential-free namespace, topic, and leaf text aliases, JSON manual equality, invalid-prefix rejection, and no .polymetrics state; gsd_exec 9fe9a1c0-e528-4ec5-a994-8416920f4867 proves exactly four targeted docs are byte-identical and repeat-generation idempotent in external scratch; gsd_exec 2f09b230-ac3f-48df-9c43-8cd57dea9a5a proves focused CLI/connector, website unit/type, catalog, docs validation, and whitespace gates; gsd_exec d4988b68-c49f-4dd4-836b-2060633df340 passes the exact repository-wide make verify aggregate, including format, tidy, vet, full Go tests, build, docs validation, smoke, lint, and connector validation. |
| R003 | primary-user-loop | active | M001/S03 | M001/S01, M001/S02 | mapped |
| R004 | compliance/security | active | M001/S04 | M001/S05, M001/S06, M001/S07 | mapped |
| R005 | core-capability | active | M001/S05 | M001/S04, M001/S02 | mapped |
| R006 | compliance/security | active | M001/S06 | M001/S04, M001/S02 | mapped |

## Coverage Summary

- Active requirements: 4
- Mapped to slices: 4
- Validated: 2 (R001, R002)
- Unmapped active requirements: 0
