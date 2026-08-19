# Issue #3866: Shared transport family conformance

## Task Delivery Header

- Issue: Refs #3866 — test(sync): add cross-transport any-to-any conformance matrix.
- Base branch: `integration/4015-mvp-flat-r1` at `a7a4c98e41af0ea32aa33db802a2a2afe0f6b8d0`.
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: A pull request open against `integration/4015-mvp-flat-r1`, with the stated local gates green and its API-reported base verified.
- Working branch: `fm/cli-3866-any-to-any-conformance-r1`.
- Task: Add deterministic, fake/contract-fixture conformance coverage for API/database source and destination **family half-paths**, every applicable declared sync mode, and the fixed coordination/durability invariants. Prove a named path can fail after schema compilation. Do not modify production registration, certification matrices/flags, providers, databases, or protocol executors.
- Verification: Focused matrix and owning-package tests; `internal/cli` and `cmd/connectorgen` consumer tests; vet/build; each non-test `make verify` gate; generated docs twice for byte stability; the connector boundary; an intentional schema-valid fault proving one named case turns red then passes once restored; PR-base API read-back.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| API source → warehouse family path | fake | A deterministic fake API source produces exact fixture records and the fake warehouse records its sealed staged workset. A provider call would prove a connector route, which is expressly out of scope. |
| Database source → warehouse family path | fake | A deterministic fake database source produces exact fixture records and the fake warehouse records its sealed staged workset. A database call would be live certification, which is expressly out of scope. |
| Warehouse → API destination family path | fake | A reopened sealed workset is supplied to a fake API destination, whose received records and acknowledgement are asserted. A provider call would prove a connector route, which is expressly out of scope. |
| Warehouse → database destination family path | fake | A reopened sealed workset is supplied to a fake database destination, whose received records and acknowledgement are asserted. A database call would be live certification, which is expressly out of scope. |
| Canonical modes and durability/coordination invariants | fake | Contract fixtures assert declared mode admission, verified-auth fencing, rate park/resume, acknowledgement-before-checkpoint, cancellation, conflict, and restart results without a far-side protocol. Their purpose is deterministic shared-layer coverage. |
| Matrix failure sensitivity | fake | A scratch-only schema-valid replacement of one selected family binding makes that path's named test fail while independent cases remain valid; restoration returns the test to green. |

## Scope decisions

- This is a transport-family matrix, not a connector-route matrix. PR #4195's `TestWarehouseMediatedFourPathConformance` already proves four distinct registered GitHub/PostgreSQL route contracts and sealed-workset ordering. This phase must neither duplicate nor re-label that proof.
- Branch evidence supersedes stale issue status: #3864 is closed; base history contains #4131 (the #4035 correction), #4138 (verified-auth cohort fencing), #4152 (durable rate parking/resume), and #4197 (#4072's final residual). The missing GitHub issue closures do not represent missing base behavior.
- `incremental_dedupe_history` is executable on this base. The matrix records executable behavior; it must not revive stale refusal language.
- The no-role-spawn contract and this harness's no-delegation rule require inline/manual GSD execution. This is a fallback for execution mechanics only, not for TDD, verification, or review.

## Required skills

`golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `golang-database`, `golang-code-style`, and `golang-lint`.
