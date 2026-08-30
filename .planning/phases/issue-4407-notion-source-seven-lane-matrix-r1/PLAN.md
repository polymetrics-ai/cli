# Plan

## Task Delivery Header

- Issue: Refs #4407 — Notion source-to-seven-lane matrix.
- Base branch: `feat/4407-notion-track-a-matrix-r1` at frozen review target `46e6d645668d520d3e144c9a36d0ef80204d4b42`.
- Merges into: `feat/4407-notion-track-a-matrix-r1` → the existing Batch R1 parent branch → `main` only after captain approval.
- Delivery: A scoped repair branch is committed, pushed, and independently re-reviewed; no PR, parent integration, or merge is performed by this repair task.
- Working branch: `fix/4407-notion-semantic-post-direct-read-r1`.
- Task: Repair the Notion Track A direct-read mapping so the four retained, source-documented bounded semantic POST reads are `source_candidate` / `mapped_unproven`, while mutation POSTs remain not-applicable unless their retained source contract proves read semantics and a bounded response. Preserve `query-meeting-notes` as both a direct-read and restricted ETL candidate.
- Verification: JSON parsing; focused red/green/edge reconciliation test; package and race tests; vet; agent-contract check; diff check; independent re-review request.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Four semantic POST reads remain mapped direct reads | live | The connector-local matrix test reads the retained lock and asserts the four source IDs have source-backed `direct_read` cells; a GET-only matrix fails. |
| Mutation POSTs do not become direct reads by method alone | live | An in-memory `post-page` mutation promotion is rejected by the reconciliation validator. |
| Semantic POST reads require retained source evidence | live | Removing the retained `200 application/json` response fact for `post-search` causes the reconciliation validator to fail. |
| Meeting-notes retains both applicable lanes | live | The test asserts `query-meeting-notes` is mapped-unproven in direct-read and ETL and retains the ETL continuation restriction. |

GSD lifecycle evidence: `scripts/gsd doctor`; `scripts/gsd sources` for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; and all five generated prompts were reviewed. The repair was executed inline because this isolated Firstmate repair lane has no compatible Pi execution role; that fallback preserves the existing issue plan, red/green ledger, verification checklist, and independent review gate.

## 1. Preserve frozen source input

Materialize only the Notion operation lock and crosswalk missing from `origin/main` by Git archive from the fixed Batch R1 snapshot. Verify each destination blob hash and byte count against the source Git objects.

## 2. Build explicit seven-lane matrix

Create one source-operation row per retained lock ID, retaining method, path, operation ID, citation, scope parameters, media, pagination, and callback/event facts. Each row has all seven cells with an allowed disposition. Keep source-only restrictions visible; use `mapped_unproven`, never `implemented`, because this track has no runtime proof.

## 3. Add reconciliation test

Validate lock denominator, exact source IDs, every lane cell, source backlink, crosswalk-only boundary records, manual semantic-read/mutation/ETL/binary cohorts, lane-count summary, and no promotion to `implemented`. Include deliberate red cases for hidden rows, a missing cell, a missing backlink, an invalid disposition, and a dropped boundary identity.

## 4. Verification and review

Run source JSON validity, the focused Notion package test and vet, source/projection check modes that do not change files, `agentcontractgen check`, staged whitespace checks, and scoped manual review. Record the two expected mapping-control restrictions separately: v2 embedded `source_operation` payload parsing and the absent canonical descriptor. Neither is a runtime-foundation claim.

## Skills applied

`connector-lane-build-order`, `go-engineering`, `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-safety`, `golang-security`, `golang-design-patterns`, and `golang-structs-interfaces`.

## CLI help/manual/website parity

Not applicable: this repair changes no `pm` command, command metadata, flag, help text, generated manual, or website page. It changes only the Track A source-mapping artifact and its connector-local reconciliation test; no runtime surface is claimed as executable.
