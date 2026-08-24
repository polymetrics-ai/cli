# Context — Zoom full definition mapping continuation

## Task Delivery Header

- Issue: Refs #4265 — Zoom five-class parity foundation cohort.
- Base branch: `main` (the temporary #4304 stack has landed).
- Merges into: `fm/cli-zoom-full-definition-mapping-r1` → `main`.
- Delivery: PR #4285 remains draft to `main`, with committed Zoom-local definitions, generated artifacts, and a complete seven-surface readiness ledger. The retained-artifact foundation is landed: 34 exact current OpenAPI captures are checked in, while the one irrecoverable historical `accounts` capture is an explicit blocking source gap. Credentialed proof remains pending. This lane never merges the PR.
- Working branch: `fm/cli-zoom-full-definition-mapping-r1`.
- Task: Reconcile the 1,937-operation source inventory and 206 typed action candidates, retaining user-reachable CLI actions without declaring a destination whose provider delete semantics cannot be represented safely.
- Verification: connector-local red/green tests; `connectorgen validate`, certification artifact generation/checks, `surface-sync --check`, `connector-boundary`, focused CLI timeout reproduction, and `make verify`.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| The source declaration selects only the executable ETL adapter | local | The Zoom bundle test loads `sync_transport.json`, asserts the exact source executor and source delivery facts, and refuses a destination declaration until tombstone-aware delete semantics exist. |
| Delete actions remain guarded and user-reachable | local | The eligibility test cross-checks eight direct CLI DELETE commands and source-traces why they are not ordinary replay destinations; no provider mutation is attempted. |
| Every source/ledger operation is auditable across all seven surfaces | local | The committed readiness ledger and its test assert one record per 1,937 provider source operation and every 1,913 ledger operation with source, parity, binding, transport, implementation, certification, and exact gap fields. |
| Certification is not overstated | live | Matrix and readiness assertions require fixture and accepted live proof before a row can claim certification; no new live evidence is written. |

## Locked decisions

- #4304 has landed through `main`; PR #4285 stays draft to `main` and is never merged by this lane.
- Definitions, tests, fixtures, generated artifacts, and readiness evidence remain under `internal/connectors/defs/zoom/`. No shared engine hook is authorized for delete semantics.
- All eight `user_id`-only source-key overlaps are provider DELETE actions. `internal/app/issue_label_warehouse_transport.go:944` correctly rejects DELETE as ordinary `full_append`; no destination is declared, no substitute action is invented, and all eight actions remain independently direct-CLI-reachable.
- The v3 source lock now owns 35 documents: 34 direct first-party API Hub OpenAPI documents (1,871 operation identities) and one zero-operation `accounts` unavailable document. `sources/artifacts/` contains all 34 exact checked-in raw captures and `zoom-retained-artifacts.json` records their provenance. `zoom-source-repin-report.json` records every old/new byte count and digest under Firstmate inbox 002 authorization.
- `accounts` is not re-pinned: its historic Next-data identity was 805,789 bytes / `d8d650…a98319a`, the source URL returned HTTP 404 with an 8,329-byte error body, and the historic blob search found no verified copy. Its 66 historic crosswalk identities remain visible; no error body, redirect, login wall, replacement descriptor, or fabricated provenance is accepted.
- A destructive action may be declared and command-reachable, but live execution remains limited to lane-created resources with plan → preview → approval → execute. No destructive Zoom call is part of this continuation.
- Certification remains fixture-plus-live-proof only. Existing evidence is retained as observed/live proof where applicable; no uncertified cell is promoted.
- The GSD lifecycle is executed inline because the canonical single-worker contract forbids role spawning. `no-mistakes` is not run because the direct-PR brief expressly excludes it.

## Required skills

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
`golang-concurrency`, and `golang-documentation`.
