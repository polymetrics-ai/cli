# PLAN — Issue 3981: managed-target ownership and provisioning

## GSD setup and fallback

- Passed `scripts/gsd doctor` and `go run ./cmd/agentcontractgen check`.
- Resolved sources for `discuss-phase`, `plan-phase`, `execute-phase`,
  `verify-work`, and `code-review`; generated the `discuss-phase --auto` and
  `plan-phase --tdd --skip-research` prompts.
- #3981 is an issue foundation rather than a numbered `.planning/ROADMAP.md`
  phase. The single-worker contract and unavailable compatible isolated GSD
  runtime require the permitted inline/manual fallback. This plan, ledger, and
  verification checklist preserve the lifecycle without weakening TDD or review.
- Required skills loaded: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, and
  `golang-database`. `golang-database`'s requested pattern scan was completed
  read-only; no database code or SQL is in scope.

## Goal

Make a managed target structurally owned by workspace/source connector/source
connection/immutable stream, with one asserted namespace owner and one control
row per relation. An existing connection can create and reassert multiple stable
stream relations; a new connection cannot adopt its predecessor's target.

## TDD slices

1. **Red — second stream:** Change the old collision test into a two-stream
   requirement: shared owner/namespace, different immutable stream IDs and
   relations, first stream created then second stream succeeds. Retain this
   failing output before production changes.
2. **Green — ownership model:** Add opaque target-database/namespace-native
   identity and namespace-owner record/state to the connector-neutral model.
   Split observation control state, make namespace-owned + relation/control
   absent creatable, retain all other exact/refusal outcomes, and serialize
   namespace creation across streams. Reassert after every create.
3. **Red/green — immutable identity:** Add `StreamID` to persisted stream state;
   allocate once on creation and migrate missing legacy values exactly once with
   collision protection. Derive/compare managed relations by stream ID, never
   mutable table/map-key/display text.
4. **Regression:** Cover every required table-driven state, target namespace
   reuse/isolation, rename stability, cancellation, cross-provisioner concurrency,
   typed-plan mutation authority, no credential/display leakage, and driver fake
   only execution.
5. **Verification/review:** Run focused/race/cancellation tests, scoped build and
   quality gates. Generate and execute `verify-work` and `code-review` prompts
   inline; use gap plans only if verification finds a real behavior gap.

## Guardrails

- Do not change `internal/warehouse.ArtifactRef` semantics.
- Do not put target database identity, mode, mapping, schema, keys, cursor, or
  ordering into physical names.
- Do not add dependencies, DDL, PostgreSQL terms, driver branches, generic SQL,
  automatic adoption, or automatic schema migration.
- Driver lock scope must prevent competing streams from racing namespace creation;
  cancellation is propagated and reassertion uses the existing no-cancel
  observation rule only to classify a possibly committed mutation.

## Checkpoints

1. Commit planning artifacts.
2. Commit the preserved red test output plus failing expectation.
3. Commit green implementation and focused tests.
4. Commit review/gap fixes only after their focused green proof.
