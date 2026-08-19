# Connector certification foundation: G1, G2, G6

## Task Delivery Header

- Issue: Refs #4260 — dedicated G1/G2/G6 certification-foundation tracking
  issue (incremental delivery to the #4015 integration branch).
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with local gates complete and GitHub checks green; read the API-reported base after opening.
- Working branch: `fm/cli-certification-foundation-g1g2g6-r1`
- Task: Make generated connector parity classification exhaustive and trustworthy, preserve generated write action (including delete), atomically publish immutable proof batches, repair the live evidence order, regenerate the GitHub sweep, and document the architecture.
- Verification: Focused red/green generator tests; sweep/candidate/matrix byte-drift checks; concurrent-reader and ordering regressions; Go package tests, vet, build, repository gates, and GitHub PR checks.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every sweep row has one exact parity projection. | live | Unit tests assert API, CLI intent/reference, CDC/changefeed, and transport rows each emit a legal kind/class; invalid/mismatched rows return an error. |
| API and managed database under-reporting are eliminated. | live | Zoom/GitLab `capability:read` projects as `rest_read`; PostgreSQL's declared managed destination projects as `reverse_etl` despite generic direct-write metadata. The classifier applies the same rule to any future MySQL managed-destination descriptor; this base SHA declares no such MySQL descriptor. |
| Delete can be independently selected. | live | A declared `delete` write row emits `direct_write` plus `write_action=delete`. |
| Published accepted evidence is complete and immutable. | live | A controlled write failure publishes zero records; a concurrent matrix reader never observes malformed JSON; duplicate deterministic paths refuse replacement. |
| Live evidence survives its validation/order path. | live | A regression test asserts script-equivalent draft/import/scoped-generation/check order retains evidence; the old publish/check/delete sequence would remove it. |
| Documentation guides later work. | live | Design describes generated cells, mutable resource keys, atomic publication/fan-in; AGENTS.md points to that section. |

## Scope and exclusions

- In scope: G1, G2, G6; generated sweep shards as required; GitHub proof; architecture docs and concise AGENTS pointer.
- Excluded: G3–G5, G7–G14, `certificationallowlist.go`, global `certification-matrix --all` refresh, and any new certification authority. A bounded credentialed GitHub direct-read proof is required by the added live-proof requirement.

## TDD execution slices

1. **Classification.** Red: table tests make direct API reads, CLI operation/write/stream combinations, CDC/changefeed, managed DB destination, delete action, invalid N/A, and mismatches fail. Green: a single generated classifier produces only legal exact kind/class/action projections. Refactor: remove parallel ad-hoc classification paths.
2. **Exhaustive sweep.** Red: snapshot/byte-drift test identifies missing projection fields and an unprojectable row. Green: every generated sweep row has `operation_kind`, `op_class`, and applicable `write_action`; regenerate all sweep artifacts including GitHub. Refactor: surface-sync verifies projection consistency.
3. **Evidence transaction.** Red: tests expose partial publication after validation/write failure, reader-visible partial JSON, replacement, and script deletion after global drift. Green: staged no-replace atomic publication, batch prevalidation, reader-safe handoff, and draft/import/scoped-generation script ordering. Refactor: retain error context and clean staging files.
4. **Documentation and verification.** Update architecture and AGENTS pointer; run focused then repository gates; record exact results and PR base/read-back.

## Required skills and lifecycle

- Skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-database`, `golang-documentation`.
- GSD: `scripts/gsd doctor`; `scripts/gsd sources` for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; `go run ./cmd/agentcontractgen check`; generated prompts run inline under the documented no-spawn fallback.

## CLI help/manual/website parity

Not applicable: this changes a generator and generated connector metadata, not an end-user CLI command, flag, help topic, output contract, manual page, or website page. The architecture documentation and existing generator command tests are in scope.

## Commit checkpoints

1. Plan + red test evidence.
2. Classification/sweep green slice and regenerated artifacts.
3. Atomic proof/import/script green slice.
4. Documentation, verification, review fixes, PR.
